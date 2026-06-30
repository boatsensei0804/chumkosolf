package service

import (
	"context"
	"strings"
	"time"

	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/tenant"
)

// DashboardRepository contract ของชั้น DB
type DashboardRepository interface {
	AdvisorSummary(ctx context.Context, schoolID, semesterID, userID string) (classCount, studentCount int, err error)
	TodayAttendanceCounts(ctx context.Context, schoolID, semesterID, userID string, date time.Time) ([]domain.StatusCount, error)
	TeacherSlots(ctx context.Context, schoolID, semesterID, userID string) ([]domain.TeacherCheckinSlot, error)
}

// DashboardAttendanceDTO สรุปการเช็คชื่อนักเรียนที่ปรึกษาของวันนี้
type DashboardAttendanceDTO struct {
	Present       int `json:"present"`
	Late          int `json:"late"`
	Absent        int `json:"absent"`
	SickLeave     int `json:"sick_leave"`
	PersonalLeave int `json:"personal_leave"`
	Unchecked     int `json:"unchecked"`
	Total         int `json:"total"`
}

// DashboardSlotDTO คาบสอนในตารางสอนหน้าแรก
type DashboardSlotDTO struct {
	DayOfWeek   int    `json:"day_of_week"`
	PeriodNo    int    `json:"period_no"`
	SubjectCode string `json:"subject_code"`
	SubjectName string `json:"subject_name"`
	ClassLabel  string `json:"class_label"`
}

// DashboardDTO ข้อมูลสรุปหน้าแรก
type DashboardDTO struct {
	IsAdvisor    bool                   `json:"is_advisor"`
	AdviseeCount int                    `json:"advisee_count"`
	Today        string                 `json:"today"`
	TodayWeekday int                    `json:"today_weekday"` // 1=จันทร์..7=อาทิตย์
	Attendance   DashboardAttendanceDTO `json:"attendance"`
	Slots        []DashboardSlotDTO     `json:"slots"`
}

// DashboardService สร้างข้อมูลสรุปหน้าแรกของผู้ใช้
type DashboardService struct {
	repo DashboardRepository
	now  func() time.Time
}

func NewDashboardService(repo DashboardRepository) *DashboardService {
	return &DashboardService{repo: repo, now: time.Now}
}

func (s *DashboardService) Summary(ctx context.Context) (*DashboardDTO, error) {
	today := s.now().UTC()
	wd := int(today.Weekday()) // อาทิตย์=0
	if wd == 0 {
		wd = 7
	}
	out := &DashboardDTO{Today: today.Format(isoDate), TodayWeekday: wd, Slots: []DashboardSlotDTO{}}

	sem := tenant.SemesterIDFromContext(ctx)
	if sem == "" {
		return out, nil // ยังไม่กำหนดเทอม → คืนค่าว่าง (หน้าแรกยังแสดงได้)
	}
	schoolID := tenant.SchoolIDFromContext(ctx)
	userID := tenant.UserIDFromContext(ctx)

	classCount, studentCount, err := s.repo.AdvisorSummary(ctx, schoolID, sem, userID)
	if err != nil {
		return nil, err
	}
	out.IsAdvisor = classCount > 0
	out.AdviseeCount = studentCount

	counts, err := s.repo.TodayAttendanceCounts(ctx, schoolID, sem, userID, today)
	if err != nil {
		return nil, err
	}
	for _, c := range counts {
		switch c.Status {
		case domain.AttendancePresent:
			out.Attendance.Present = c.Count
		case domain.AttendanceLate:
			out.Attendance.Late = c.Count
		case domain.AttendanceAbsent:
			out.Attendance.Absent = c.Count
		case domain.AttendanceSickLeave:
			out.Attendance.SickLeave = c.Count
		case domain.AttendancePersonalLeave:
			out.Attendance.PersonalLeave = c.Count
		default:
			out.Attendance.Unchecked += c.Count
		}
		out.Attendance.Total += c.Count
	}

	slots, err := s.repo.TeacherSlots(ctx, schoolID, sem, userID)
	if err != nil {
		return nil, err
	}
	for i := range slots {
		out.Slots = append(out.Slots, DashboardSlotDTO{
			DayOfWeek: slots[i].DayOfWeek, PeriodNo: slots[i].PeriodNo,
			SubjectCode: slots[i].SubjectCode, SubjectName: slots[i].SubjectName,
			ClassLabel: strings.TrimSpace(slots[i].GradeLevel + " " + slots[i].RoomName),
		})
	}
	return out, nil
}
