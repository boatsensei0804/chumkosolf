package service

import (
	"context"
	"strings"
	"time"

	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/tenant"
)

// generalAffairsWorkGroupCode = กลุ่มงานบริหารทั่วไป (เจ้าของฟีเจอร์เช็คชื่อ/ความประพฤติ)
const generalAffairsWorkGroupCode = "general_affairs"

// parseISODate แปลง "YYYY-MM-DD" → time.Time
func parseISODate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

// AttendanceRepository contract ของชั้น DB
type AttendanceRepository interface {
	RosterByClassDate(ctx context.Context, schoolID, classID string, date time.Time) ([]domain.AttendanceRosterEntry, error)
	BulkUpsert(ctx context.Context, schoolID, semesterID, classID string, date time.Time, marks []domain.AttendanceMark, checkedBy string, audit domain.AuditEntry) error
	IsClassAdvisorUser(ctx context.Context, schoolID, classID, userID string) (bool, error)
}

// attendanceClassRepo ใช้ยืนยันว่าห้องมีอยู่จริงในโรงเรียน (scope)
type attendanceClassRepo interface {
	GetByID(ctx context.Context, schoolID, id string) (*domain.Class, error)
}

// AttendanceRosterDTO รายชื่อนักเรียน + สถานะเช็คชื่อของวันนั้น (status ว่าง = ยังไม่เช็ค)
type AttendanceRosterDTO struct {
	StudentID   string `json:"student_id"`
	StudentNo   *int   `json:"student_no"`
	StudentCode string `json:"student_code"`
	Prefix      string `json:"prefix"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Status      string `json:"status"`
	Note        string `json:"note"`
	// DailyStatus = สถานะเช็คชื่อเข้าเรียนรายวัน (โชว์ "มาสาย" ในเช็คชื่อรายวิชา); ว่างในหน้าเช็คชื่อรายวันเอง
	DailyStatus string `json:"daily_status"`
}

// AttendanceMarkInput ผลเช็คชื่อของนักเรียน 1 คน
type AttendanceMarkInput struct {
	StudentID string
	Status    string
	Note      string
}

// AttendanceService จัดการเช็คชื่อเข้าเรียนรายวัน (รายเทอม)
type AttendanceService struct {
	repo    AttendanceRepository
	classes attendanceClassRepo
	checker WorkGroupChecker
}

func NewAttendanceService(repo AttendanceRepository, classes attendanceClassRepo, checker WorkGroupChecker) *AttendanceService {
	return &AttendanceService{repo: repo, classes: classes, checker: checker}
}

// authorizeClass: ห้องต้องมีจริง (scope) และผู้ใช้ต้องเป็น admin / กลุ่มบริหารทั่วไป / ครูที่ปรึกษาห้องนี้
func (s *AttendanceService) authorizeClass(ctx context.Context, classID string) error {
	schoolID := tenant.SchoolIDFromContext(ctx)
	c, err := s.classes.GetByID(ctx, schoolID, classID)
	if err != nil {
		return err
	}
	if c == nil {
		return domain.ErrClassNotFound
	}
	if tenant.IsSchoolAdminFromContext(ctx) {
		return nil
	}
	userID := tenant.UserIDFromContext(ctx)
	inGroup, err := s.checker.IsUserInWorkGroup(ctx, schoolID, userID, generalAffairsWorkGroupCode)
	if err != nil {
		return err
	}
	if inGroup {
		return nil
	}
	isAdvisor, err := s.repo.IsClassAdvisorUser(ctx, schoolID, classID, userID)
	if err != nil {
		return err
	}
	if isAdvisor {
		return nil
	}
	return domain.ErrForbidden
}

// ListRoster คืนรายชื่อนักเรียนในห้อง + สถานะเช็คชื่อของวันที่ระบุ
func (s *AttendanceService) ListRoster(ctx context.Context, classID, dateStr string) ([]AttendanceRosterDTO, error) {
	if err := s.authorizeClass(ctx, classID); err != nil {
		return nil, err
	}
	date, err := parseISODate(dateStr)
	if err != nil {
		return nil, domain.ErrInvalidDate
	}
	rows, err := s.repo.RosterByClassDate(ctx, tenant.SchoolIDFromContext(ctx), classID, date)
	if err != nil {
		return nil, err
	}
	out := make([]AttendanceRosterDTO, 0, len(rows))
	for i := range rows {
		out = append(out, AttendanceRosterDTO{
			StudentID: rows[i].StudentID, StudentNo: rows[i].StudentNo, StudentCode: rows[i].StudentCode,
			Prefix: rows[i].Prefix, FirstName: rows[i].FirstName, LastName: rows[i].LastName,
			Status: rows[i].Status, Note: rows[i].Note,
		})
	}
	return out, nil
}

// Save บันทึกผลเช็คชื่อทั้งห้องของวันนั้น (เฉพาะนักเรียนที่อยู่ในห้องจริง)
func (s *AttendanceService) Save(ctx context.Context, classID, dateStr string, marks []AttendanceMarkInput) error {
	if err := s.authorizeClass(ctx, classID); err != nil {
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
	// โหลดรายชื่อในห้องเพื่อยืนยันว่าทุก mark เป็นนักเรียนในห้องจริง (กันยัด student ข้ามห้อง/โรงเรียน)
	roster, err := s.repo.RosterByClassDate(ctx, schoolID, classID, date)
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

	audit := auditFor(ctx, domain.AuditUpdate, "attendance", classID, map[string]any{"date": dateStr, "count": len(domainMarks)})
	return s.repo.BulkUpsert(ctx, schoolID, sem, classID, date, domainMarks, tenant.UserIDFromContext(ctx), audit)
}
