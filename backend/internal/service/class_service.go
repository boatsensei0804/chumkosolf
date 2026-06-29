package service

import (
	"context"
	"strings"
	"time"

	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/tenant"
)

// semesterOrErr ดึง semester ปัจจุบันจาก context (ข้อมูลรายเทอมต้องมีเสมอ)
func semesterOrErr(ctx context.Context) (string, error) {
	s := tenant.SemesterIDFromContext(ctx)
	if s == "" {
		return "", domain.ErrNoActiveSemester
	}
	return s, nil
}

// ================= Classes =================

type ClassRepository interface {
	ListBySemester(ctx context.Context, schoolID, semesterID string) ([]domain.Class, error)
	GetByID(ctx context.Context, schoolID, id string) (*domain.Class, error)
	Create(ctx context.Context, schoolID, semesterID string, nc domain.NewClass, audit domain.AuditEntry) (string, error)
	Update(ctx context.Context, schoolID, id string, uc domain.UpdateClass, audit domain.AuditEntry) (bool, error)
	SoftDelete(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error)
}

type ClassListItem struct {
	ID           string `json:"id"`
	GradeLevel   string `json:"grade_level"`
	RoomName     string `json:"room_name"`
	StudentCount int    `json:"student_count"`
	AdvisorCount int    `json:"advisor_count"`
}

type ClassDetail struct {
	ID         string `json:"id"`
	GradeLevel string `json:"grade_level"`
	RoomName   string `json:"room_name"`
	SemesterID string `json:"semester_id"`
	CreatedAt  string `json:"created_at"`
}

type ClassInput struct {
	GradeLevel string
	RoomName   string
}

type ClassService struct {
	guard academicGuard
	repo  ClassRepository
}

func NewClassService(repo ClassRepository, checker WorkGroupChecker) *ClassService {
	return &ClassService{guard: academicGuard{checker: checker}, repo: repo}
}

func (s *ClassService) List(ctx context.Context) ([]ClassListItem, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return nil, err
	}
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListBySemester(ctx, tenant.SchoolIDFromContext(ctx), sem)
	if err != nil {
		return nil, err
	}
	out := make([]ClassListItem, 0, len(rows))
	for i := range rows {
		out = append(out, ClassListItem{
			ID: rows[i].ID, GradeLevel: rows[i].GradeLevel, RoomName: rows[i].RoomName,
			StudentCount: rows[i].StudentCount, AdvisorCount: rows[i].AdvisorCount,
		})
	}
	return out, nil
}

func (s *ClassService) Get(ctx context.Context, id string) (*ClassDetail, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return nil, err
	}
	c, err := s.repo.GetByID(ctx, tenant.SchoolIDFromContext(ctx), id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, domain.ErrClassNotFound
	}
	return &ClassDetail{ID: c.ID, GradeLevel: c.GradeLevel, RoomName: c.RoomName, SemesterID: c.SemesterID, CreatedAt: c.CreatedAt.Format(time.RFC3339)}, nil
}

func (s *ClassService) Create(ctx context.Context, in ClassInput) (string, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return "", err
	}
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(in.GradeLevel) == "" || strings.TrimSpace(in.RoomName) == "" {
		return "", domain.ErrValidation
	}
	audit := auditFor(ctx, domain.AuditCreate, "class", "", map[string]any{"grade_level": in.GradeLevel, "room_name": in.RoomName})
	return s.repo.Create(ctx, tenant.SchoolIDFromContext(ctx), sem, domain.NewClass{
		GradeLevel: strings.TrimSpace(in.GradeLevel), RoomName: strings.TrimSpace(in.RoomName),
	}, audit)
}

func (s *ClassService) Update(ctx context.Context, id string, in ClassInput) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(in.GradeLevel) == "" || strings.TrimSpace(in.RoomName) == "" {
		return domain.ErrValidation
	}
	audit := auditFor(ctx, domain.AuditUpdate, "class", id, nil)
	found, err := s.repo.Update(ctx, tenant.SchoolIDFromContext(ctx), id, domain.UpdateClass{
		GradeLevel: strings.TrimSpace(in.GradeLevel), RoomName: strings.TrimSpace(in.RoomName),
	}, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrClassNotFound
	}
	return nil
}

func (s *ClassService) Delete(ctx context.Context, id string) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditDelete, "class", id, nil)
	found, err := s.repo.SoftDelete(ctx, tenant.SchoolIDFromContext(ctx), id, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrClassNotFound
	}
	return nil
}

// ================= Class advisors =================

type personnelExists interface {
	GetByID(ctx context.Context, schoolID, id string) (*domain.Personnel, error)
}

type ClassAdvisorRepository interface {
	ListByClass(ctx context.Context, schoolID, classID string) ([]domain.ClassAdvisor, error)
	Add(ctx context.Context, schoolID, semesterID, classID, personnelID string, audit domain.AuditEntry) error
	Remove(ctx context.Context, schoolID, classID, advisorID string, audit domain.AuditEntry) (bool, error)
}

type ClassAdvisorDTO struct {
	ID          string `json:"id"`
	PersonnelID string `json:"personnel_id"`
	Prefix      string `json:"prefix"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
}

type ClassAdvisorService struct {
	guard     academicGuard
	repo      ClassAdvisorRepository
	classes   ClassRepository
	personnel personnelExists
}

func NewClassAdvisorService(repo ClassAdvisorRepository, classes ClassRepository, personnel personnelExists, checker WorkGroupChecker) *ClassAdvisorService {
	return &ClassAdvisorService{guard: academicGuard{checker: checker}, repo: repo, classes: classes, personnel: personnel}
}

func (s *ClassAdvisorService) ensureClass(ctx context.Context, classID string) error {
	c, err := s.classes.GetByID(ctx, tenant.SchoolIDFromContext(ctx), classID)
	if err != nil {
		return err
	}
	if c == nil {
		return domain.ErrClassNotFound
	}
	return nil
}

func (s *ClassAdvisorService) List(ctx context.Context, classID string) ([]ClassAdvisorDTO, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return nil, err
	}
	if err := s.ensureClass(ctx, classID); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListByClass(ctx, tenant.SchoolIDFromContext(ctx), classID)
	if err != nil {
		return nil, err
	}
	out := make([]ClassAdvisorDTO, 0, len(rows))
	for i := range rows {
		out = append(out, ClassAdvisorDTO{ID: rows[i].ID, PersonnelID: rows[i].PersonnelID, Prefix: rows[i].Prefix, FirstName: rows[i].FirstName, LastName: rows[i].LastName})
	}
	return out, nil
}

func (s *ClassAdvisorService) Add(ctx context.Context, classID, personnelID string) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return err
	}
	if err := s.ensureClass(ctx, classID); err != nil {
		return err
	}
	p, err := s.personnel.GetByID(ctx, tenant.SchoolIDFromContext(ctx), personnelID)
	if err != nil {
		return err
	}
	if p == nil {
		return domain.ErrPersonnelNotFound
	}
	audit := auditFor(ctx, domain.AuditUpdate, "class_advisor", classID, map[string]any{"personnel_id": personnelID})
	return s.repo.Add(ctx, tenant.SchoolIDFromContext(ctx), sem, classID, personnelID, audit)
}

func (s *ClassAdvisorService) Remove(ctx context.Context, classID, advisorID string) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditDelete, "class_advisor", classID, map[string]any{"advisor_id": advisorID})
	found, err := s.repo.Remove(ctx, tenant.SchoolIDFromContext(ctx), classID, advisorID, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrAdvisorNotFound
	}
	return nil
}

// ================= Enrollments =================

type EnrollmentRepository interface {
	ListByClass(ctx context.Context, schoolID, classID string) ([]domain.ClassEnrollment, error)
	Enroll(ctx context.Context, schoolID, semesterID, classID string, ne domain.NewEnrollment, audit domain.AuditEntry) error
	Remove(ctx context.Context, schoolID, classID, enrollmentID string, audit domain.AuditEntry) (bool, error)
}

type EnrollmentDTO struct {
	ID          string `json:"id"`
	StudentID   string `json:"student_id"`
	StudentNo   *int   `json:"student_no"`
	StudentCode string `json:"student_code"`
	Prefix      string `json:"prefix"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
}

type EnrollInput struct {
	StudentID string
	StudentNo *int
}

type EnrollmentService struct {
	guard    academicGuard
	repo     EnrollmentRepository
	classes  ClassRepository
	students StudentRepository
}

func NewEnrollmentService(repo EnrollmentRepository, classes ClassRepository, students StudentRepository, checker WorkGroupChecker) *EnrollmentService {
	return &EnrollmentService{guard: academicGuard{checker: checker}, repo: repo, classes: classes, students: students}
}

func (s *EnrollmentService) ensureClass(ctx context.Context, classID string) error {
	c, err := s.classes.GetByID(ctx, tenant.SchoolIDFromContext(ctx), classID)
	if err != nil {
		return err
	}
	if c == nil {
		return domain.ErrClassNotFound
	}
	return nil
}

func (s *EnrollmentService) List(ctx context.Context, classID string) ([]EnrollmentDTO, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return nil, err
	}
	if err := s.ensureClass(ctx, classID); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListByClass(ctx, tenant.SchoolIDFromContext(ctx), classID)
	if err != nil {
		return nil, err
	}
	out := make([]EnrollmentDTO, 0, len(rows))
	for i := range rows {
		out = append(out, EnrollmentDTO{
			ID: rows[i].ID, StudentID: rows[i].StudentID, StudentNo: rows[i].StudentNo,
			StudentCode: rows[i].StudentCode, Prefix: rows[i].Prefix, FirstName: rows[i].FirstName, LastName: rows[i].LastName,
		})
	}
	return out, nil
}

func (s *EnrollmentService) Enroll(ctx context.Context, classID string, in EnrollInput) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return err
	}
	if err := s.ensureClass(ctx, classID); err != nil {
		return err
	}
	st, err := s.students.GetByID(ctx, tenant.SchoolIDFromContext(ctx), in.StudentID)
	if err != nil {
		return err
	}
	if st == nil {
		return domain.ErrStudentNotFound
	}
	audit := auditFor(ctx, domain.AuditUpdate, "student_enrollment", classID, map[string]any{"student_id": in.StudentID})
	return s.repo.Enroll(ctx, tenant.SchoolIDFromContext(ctx), sem, classID, domain.NewEnrollment{StudentID: in.StudentID, StudentNo: in.StudentNo}, audit)
}

// EnrollMany จัดนักเรียนหลายคนเข้าห้องในครั้งเดียว (ข้ามรหัสที่ไม่พบ) คืนจำนวนที่จัดเข้าสำเร็จ
func (s *EnrollmentService) EnrollMany(ctx context.Context, classID string, studentIDs []string) (int, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return 0, err
	}
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return 0, err
	}
	if err := s.ensureClass(ctx, classID); err != nil {
		return 0, err
	}
	schoolID := tenant.SchoolIDFromContext(ctx)
	seen := make(map[string]bool, len(studentIDs))
	count := 0
	for _, sid := range studentIDs {
		if sid == "" || seen[sid] {
			continue
		}
		seen[sid] = true
		st, err := s.students.GetByID(ctx, schoolID, sid)
		if err != nil {
			return count, err
		}
		if st == nil {
			continue // ข้ามรหัสที่ไม่พบ (ไม่ทำให้ทั้งชุดล้ม)
		}
		audit := auditFor(ctx, domain.AuditUpdate, "student_enrollment", classID, map[string]any{"student_id": sid, "bulk": true})
		if err := s.repo.Enroll(ctx, schoolID, sem, classID, domain.NewEnrollment{StudentID: sid}, audit); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *EnrollmentService) Remove(ctx context.Context, classID, enrollmentID string) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditDelete, "student_enrollment", classID, map[string]any{"enrollment_id": enrollmentID})
	found, err := s.repo.Remove(ctx, tenant.SchoolIDFromContext(ctx), classID, enrollmentID, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrEnrollmentNotFound
	}
	return nil
}
