package service

import (
	"context"
	"strings"

	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/tenant"
)

// TeachingAssignmentRepository contract ของชั้น DB
type TeachingAssignmentRepository interface {
	ListBySemester(ctx context.Context, schoolID, semesterID string) ([]domain.TeachingAssignment, error)
	GetByID(ctx context.Context, schoolID, id string) (*domain.TeachingAssignment, error)
	Create(ctx context.Context, schoolID, semesterID string, na domain.NewTeachingAssignment, audit domain.AuditEntry) (string, error)
	SoftDelete(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error)
}

// dependency สำหรับยืนยันว่า ครู/วิชา/ห้อง มีอยู่จริงในโรงเรียน (scope)
type taPersonnelRepo interface {
	GetByID(ctx context.Context, schoolID, id string) (*domain.Personnel, error)
}
type taSubjectRepo interface {
	GetByID(ctx context.Context, schoolID, id string) (*domain.Subject, error)
}
type taClassRepo interface {
	GetByID(ctx context.Context, schoolID, id string) (*domain.Class, error)
}

// TeachingAssignmentDTO ข้อมูลมอบหมายการสอนสำหรับ response
type TeachingAssignmentDTO struct {
	ID          string `json:"id"`
	PersonnelID string `json:"personnel_id"`
	SubjectID   string `json:"subject_id"`
	ClassID     string `json:"class_id"`
	TeacherName string `json:"teacher_name"`
	SubjectCode string `json:"subject_code"`
	SubjectName string `json:"subject_name"`
	GradeLevel  string `json:"grade_level"`
	RoomName    string `json:"room_name"`
}

// TeachingAssignmentInput ข้อมูลสร้างการมอบหมาย
type TeachingAssignmentInput struct {
	PersonnelID string
	SubjectID   string
	ClassID     string
}

// TeachingAssignmentService จัดการมอบหมายการสอน (กลุ่มวิชาการ + admin, รายเทอม)
type TeachingAssignmentService struct {
	guard     academicGuard
	repo      TeachingAssignmentRepository
	personnel taPersonnelRepo
	subjects  taSubjectRepo
	classes   taClassRepo
}

func NewTeachingAssignmentService(repo TeachingAssignmentRepository, personnel taPersonnelRepo, subjects taSubjectRepo, classes taClassRepo, checker WorkGroupChecker) *TeachingAssignmentService {
	return &TeachingAssignmentService{guard: academicGuard{checker: checker}, repo: repo, personnel: personnel, subjects: subjects, classes: classes}
}

func (s *TeachingAssignmentService) List(ctx context.Context) ([]TeachingAssignmentDTO, error) {
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
	out := make([]TeachingAssignmentDTO, 0, len(rows))
	for i := range rows {
		out = append(out, TeachingAssignmentDTO{
			ID: rows[i].ID, PersonnelID: rows[i].PersonnelID, SubjectID: rows[i].SubjectID, ClassID: rows[i].ClassID,
			TeacherName: strings.TrimSpace(rows[i].TeacherPrefix + rows[i].TeacherFirstName + " " + rows[i].TeacherLastName),
			SubjectCode: rows[i].SubjectCode, SubjectName: rows[i].SubjectName,
			GradeLevel: rows[i].GradeLevel, RoomName: rows[i].RoomName,
		})
	}
	return out, nil
}

func (s *TeachingAssignmentService) Create(ctx context.Context, in TeachingAssignmentInput) (string, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return "", err
	}
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return "", err
	}
	schoolID := tenant.SchoolIDFromContext(ctx)

	p, err := s.personnel.GetByID(ctx, schoolID, in.PersonnelID)
	if err != nil {
		return "", err
	}
	if p == nil {
		return "", domain.ErrPersonnelNotFound
	}
	subj, err := s.subjects.GetByID(ctx, schoolID, in.SubjectID)
	if err != nil {
		return "", err
	}
	if subj == nil {
		return "", domain.ErrSubjectNotFound
	}
	c, err := s.classes.GetByID(ctx, schoolID, in.ClassID)
	if err != nil {
		return "", err
	}
	if c == nil {
		return "", domain.ErrClassNotFound
	}

	audit := auditFor(ctx, domain.AuditCreate, "teaching_assignment", "", map[string]any{
		"personnel_id": in.PersonnelID, "subject_id": in.SubjectID, "class_id": in.ClassID,
	})
	return s.repo.Create(ctx, schoolID, sem, domain.NewTeachingAssignment{
		PersonnelID: in.PersonnelID, SubjectID: in.SubjectID, ClassID: in.ClassID,
	}, audit)
}

func (s *TeachingAssignmentService) Delete(ctx context.Context, id string) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditDelete, "teaching_assignment", id, nil)
	found, err := s.repo.SoftDelete(ctx, tenant.SchoolIDFromContext(ctx), id, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrTeachingAssignmentNotFound
	}
	return nil
}
