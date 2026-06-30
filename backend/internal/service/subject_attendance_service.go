package service

import (
	"context"
	"strings"
	"time"

	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/tenant"
)

// SubjectAttendanceRepository contract ของชั้น DB
type SubjectAttendanceRepository interface {
	SlotContext(ctx context.Context, schoolID, slotID string) (classID, teacherUserID string, found bool, err error)
	RosterBySlotDate(ctx context.Context, schoolID, slotID, classID string, date time.Time) ([]domain.AttendanceRosterEntry, error)
	BulkUpsert(ctx context.Context, schoolID, semesterID, slotID string, date time.Time, marks []domain.AttendanceMark, checkedBy string, audit domain.AuditEntry) error
	TeacherSlots(ctx context.Context, schoolID, semesterID, userID string) ([]domain.TeacherCheckinSlot, error)
	CheckedSlotDates(ctx context.Context, schoolID, semesterID, userID string) ([]domain.SlotDate, error)
	SemesterRange(ctx context.Context, schoolID, semesterID string) (start, end *time.Time, err error)
}

// CheckinSlotDTO คาบในกริดเช็คชื่อ + สถานะว่าเช็คของวันนั้นแล้วหรือยัง
type CheckinSlotDTO struct {
	SlotID      string `json:"slot_id"`
	DayOfWeek   int    `json:"day_of_week"`
	PeriodNo    int    `json:"period_no"`
	SubjectCode string `json:"subject_code"`
	SubjectName string `json:"subject_name"`
	ClassLabel  string `json:"class_label"`
	Date        string `json:"date"` // วันที่ของคาบนี้ในสัปดาห์ที่เลือก
	Checked     bool   `json:"checked"`
}

// CheckinWeekDTO คือสัปดาห์ของภาคเรียน (ใช้ทำ dropdown เลือกสัปดาห์)
type CheckinWeekDTO struct {
	Index int    `json:"index"` // สัปดาห์ที่ 1..N
	Start string `json:"start"` // จันทร์ของสัปดาห์
	End   string `json:"end"`   // อาทิตย์ของสัปดาห์
}

// CheckinOverviewDTO สรุปภาพรวมการเช็คชื่อรายวิชาของครู
type CheckinOverviewDTO struct {
	WeekStart         string           `json:"week_start"`
	Slots             []CheckinSlotDTO `json:"slots"`
	UncheckedThisWeek int              `json:"unchecked_this_week"`
	TotalThisWeek     int              `json:"total_this_week"`
	HasWeekStats      bool             `json:"has_week_stats"`
	IncompleteWeeks   int              `json:"incomplete_weeks"`
	TotalWeeks        int              `json:"total_weeks"`
	// รายการสัปดาห์ของเทอม (จากวันเปิด-ปิดภาคเรียน) + สัปดาห์ปัจจุบันที่เลือก
	Weeks            []CheckinWeekDTO `json:"weeks"`
	CurrentWeekIndex int              `json:"current_week_index"`
}

const isoDate = "2006-01-02"

// mondayOf คืนวันจันทร์ของสัปดาห์ที่มี t (day_of_week 1=จันทร์)
func mondayOf(t time.Time) time.Time {
	wd := int(t.Weekday()) // อาทิตย์=0..เสาร์=6
	if wd == 0 {
		wd = 7
	}
	t = t.AddDate(0, 0, -(wd - 1))
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// SubjectAttendanceService จัดการเช็คชื่อรายวิชา (รายคาบ, รายเทอม)
// สิทธิ์: admin / กลุ่มวิชาการ / ครูประจำวิชาของคาบนั้น
type SubjectAttendanceService struct {
	repo    SubjectAttendanceRepository
	checker WorkGroupChecker
}

func NewSubjectAttendanceService(repo SubjectAttendanceRepository, checker WorkGroupChecker) *SubjectAttendanceService {
	return &SubjectAttendanceService{repo: repo, checker: checker}
}

// authorizeSlot ตรวจสิทธิ์ของคาบ แล้วคืน class_id ของคาบ (ใช้หา roster)
func (s *SubjectAttendanceService) authorizeSlot(ctx context.Context, slotID string) (string, error) {
	schoolID := tenant.SchoolIDFromContext(ctx)
	classID, teacherUserID, found, err := s.repo.SlotContext(ctx, schoolID, slotID)
	if err != nil {
		return "", err
	}
	if !found {
		return "", domain.ErrTimetableSlotNotFound
	}
	if tenant.IsSchoolAdminFromContext(ctx) {
		return classID, nil
	}
	userID := tenant.UserIDFromContext(ctx)
	if userID == teacherUserID {
		return classID, nil
	}
	inGroup, err := s.checker.IsUserInWorkGroup(ctx, schoolID, userID, academicWorkGroupCode)
	if err != nil {
		return "", err
	}
	if inGroup {
		return classID, nil
	}
	return "", domain.ErrForbidden
}

// CheckinOverview คืนกริดคาบที่ครูสอน + สถานะเช็คของสัปดาห์ที่เลือก + จำนวนสัปดาห์ที่เช็คไม่ครบทั้งเทอม
// (ผู้ล็อกอินดูของตัวเอง — แสดงเฉพาะคาบที่ user นี้สอน)
func (s *SubjectAttendanceService) CheckinOverview(ctx context.Context, dateStr string) (*CheckinOverviewDTO, error) {
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return nil, err
	}
	date, err := parseISODate(dateStr)
	if err != nil {
		return nil, domain.ErrInvalidDate
	}
	schoolID := tenant.SchoolIDFromContext(ctx)
	userID := tenant.UserIDFromContext(ctx)

	slots, err := s.repo.TeacherSlots(ctx, schoolID, sem, userID)
	if err != nil {
		return nil, err
	}
	checkedRows, err := s.repo.CheckedSlotDates(ctx, schoolID, sem, userID)
	if err != nil {
		return nil, err
	}
	// set ของ "slotID|YYYY-MM-DD" ที่เช็คแล้ว
	checked := make(map[string]bool, len(checkedRows))
	for _, cr := range checkedRows {
		checked[cr.SlotID+"|"+cr.Date.UTC().Format(isoDate)] = true
	}

	week := mondayOf(date)
	out := &CheckinOverviewDTO{WeekStart: week.Format(isoDate), Slots: make([]CheckinSlotDTO, 0, len(slots))}
	for i := range slots {
		slotDate := week.AddDate(0, 0, slots[i].DayOfWeek-1)
		isChecked := checked[slots[i].SlotID+"|"+slotDate.Format(isoDate)]
		out.TotalThisWeek++
		if !isChecked {
			out.UncheckedThisWeek++
		}
		out.Slots = append(out.Slots, CheckinSlotDTO{
			SlotID: slots[i].SlotID, DayOfWeek: slots[i].DayOfWeek, PeriodNo: slots[i].PeriodNo,
			SubjectCode: slots[i].SubjectCode, SubjectName: slots[i].SubjectName,
			ClassLabel: strings.TrimSpace(slots[i].GradeLevel + " " + slots[i].RoomName),
			Date:       slotDate.Format(isoDate), Checked: isChecked,
		})
	}

	// รายการสัปดาห์ของเทอม + จำนวนสัปดาห์ที่เช็คไม่ครบ (ต้องมีวันเริ่ม-จบเทอม)
	start, end, err := s.repo.SemesterRange(ctx, schoolID, sem)
	if err != nil {
		return nil, err
	}
	if start != nil && end != nil {
		startKey := start.UTC().Format(isoDate)
		endKey := end.UTC().Format(isoDate)
		hasSlots := len(slots) > 0
		out.HasWeekStats = hasSlots
		idx := 0
		for w := mondayOf(*start); !w.After(*end); w = w.AddDate(0, 0, 7) {
			idx++
			out.Weeks = append(out.Weeks, CheckinWeekDTO{
				Index: idx, Start: w.Format(isoDate), End: w.AddDate(0, 0, 6).Format(isoDate),
			})
			if w.Equal(week) {
				out.CurrentWeekIndex = idx
			}
			// สถิติเช็คไม่ครบ (เฉพาะเมื่อครูมีคาบสอน)
			if hasSlots {
				expected, complete := false, true
				for i := range slots {
					d := w.AddDate(0, 0, slots[i].DayOfWeek-1).Format(isoDate)
					if d < startKey || d > endKey {
						continue // คาบนี้อยู่นอกช่วงเทอมในสัปดาห์นี้
					}
					expected = true
					if !checked[slots[i].SlotID+"|"+d] {
						complete = false
					}
				}
				if expected {
					out.TotalWeeks++
					if !complete {
						out.IncompleteWeeks++
					}
				}
			}
		}
	}
	return out, nil
}

// ListRoster คืนรายชื่อนักเรียนในคาบ + สถานะของวันที่ระบุ
func (s *SubjectAttendanceService) ListRoster(ctx context.Context, slotID, dateStr string) ([]AttendanceRosterDTO, error) {
	classID, err := s.authorizeSlot(ctx, slotID)
	if err != nil {
		return nil, err
	}
	date, err := parseISODate(dateStr)
	if err != nil {
		return nil, domain.ErrInvalidDate
	}
	rows, err := s.repo.RosterBySlotDate(ctx, tenant.SchoolIDFromContext(ctx), slotID, classID, date)
	if err != nil {
		return nil, err
	}
	out := make([]AttendanceRosterDTO, 0, len(rows))
	for i := range rows {
		out = append(out, AttendanceRosterDTO{
			StudentID: rows[i].StudentID, StudentNo: rows[i].StudentNo, StudentCode: rows[i].StudentCode,
			Prefix: rows[i].Prefix, FirstName: rows[i].FirstName, LastName: rows[i].LastName,
			Status: rows[i].Status, Note: rows[i].Note, DailyStatus: rows[i].DailyStatus,
		})
	}
	return out, nil
}

// Save บันทึกผลเช็คชื่อรายวิชาทั้งคาบของวันนั้น (เฉพาะนักเรียนในห้องของคาบ)
func (s *SubjectAttendanceService) Save(ctx context.Context, slotID, dateStr string, marks []AttendanceMarkInput) error {
	classID, err := s.authorizeSlot(ctx, slotID)
	if err != nil {
		return err
	}
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return err
	}
	date, err := parseISODate(dateStr)
	if err != nil {
		return domain.ErrInvalidDate
	}

	schoolID := tenant.SchoolIDFromContext(ctx)
	roster, err := s.repo.RosterBySlotDate(ctx, schoolID, slotID, classID, date)
	if err != nil {
		return err
	}
	enrolled := make(map[string]bool, len(roster))
	for i := range roster {
		enrolled[roster[i].StudentID] = true
	}

	domainMarks := make([]domain.AttendanceMark, 0, len(marks))
	for _, m := range marks {
		if !domain.ValidAttendanceStatus(m.Status) {
			return domain.ErrInvalidAttendanceStatus
		}
		if !enrolled[m.StudentID] {
			return domain.ErrStudentNotInClass
		}
		domainMarks = append(domainMarks, domain.AttendanceMark{
			StudentID: m.StudentID, Status: m.Status, Note: strings.TrimSpace(m.Note),
		})
	}
	if len(domainMarks) == 0 {
		return nil
	}

	audit := auditFor(ctx, domain.AuditUpdate, "subject_attendance", slotID, map[string]any{"date": dateStr, "count": len(domainMarks)})
	return s.repo.BulkUpsert(ctx, schoolID, sem, slotID, date, domainMarks, tenant.UserIDFromContext(ctx), audit)
}
