package service

import (
	"context"
	"strings"

	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/tenant"
)

// SubjectRepository contract ของชั้น DB
type SubjectRepository interface {
	List(ctx context.Context, schoolID string) ([]domain.Subject, error)
	GetByID(ctx context.Context, schoolID, id string) (*domain.Subject, error)
	Create(ctx context.Context, schoolID string, ns domain.NewSubject, audit domain.AuditEntry) (string, error)
	Update(ctx context.Context, schoolID, id string, us domain.UpdateSubject, audit domain.AuditEntry) (bool, error)
	SoftDelete(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error)
}

// SubjectDTO ข้อมูลรายวิชาสำหรับ response
type SubjectDTO struct {
	ID          string   `json:"id"`
	SubjectCode string   `json:"subject_code"`
	Name        string   `json:"name"`
	Credit      *float64 `json:"credit"`
}

// SubjectInput ข้อมูลสร้าง/แก้รายวิชา
type SubjectInput struct {
	SubjectCode string
	Name        string
	Credit      *float64
}

// SubjectService จัดการรายวิชา (กลุ่มวิชาการ + admin)
type SubjectService struct {
	guard academicGuard
	repo  SubjectRepository
}

func NewSubjectService(repo SubjectRepository, checker WorkGroupChecker) *SubjectService {
	return &SubjectService{guard: academicGuard{checker: checker}, repo: repo}
}

func subjectToDTO(s domain.Subject) SubjectDTO {
	return SubjectDTO{ID: s.ID, SubjectCode: s.SubjectCode, Name: s.Name, Credit: s.Credit}
}

func (s *SubjectService) List(ctx context.Context) ([]SubjectDTO, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return nil, err
	}
	rows, err := s.repo.List(ctx, tenant.SchoolIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	out := make([]SubjectDTO, 0, len(rows))
	for i := range rows {
		out = append(out, subjectToDTO(rows[i]))
	}
	return out, nil
}

func (s *SubjectService) Create(ctx context.Context, in SubjectInput) (string, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return "", err
	}
	code := strings.TrimSpace(in.SubjectCode)
	name := strings.TrimSpace(in.Name)
	if code == "" || name == "" {
		return "", domain.ErrValidation
	}
	audit := auditFor(ctx, domain.AuditCreate, "subject", "", map[string]any{"subject_code": code})
	return s.repo.Create(ctx, tenant.SchoolIDFromContext(ctx), domain.NewSubject{
		SubjectCode: code, Name: name, Credit: in.Credit,
	}, audit)
}

func (s *SubjectService) Update(ctx context.Context, id string, in SubjectInput) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	code := strings.TrimSpace(in.SubjectCode)
	name := strings.TrimSpace(in.Name)
	if code == "" || name == "" {
		return domain.ErrValidation
	}
	audit := auditFor(ctx, domain.AuditUpdate, "subject", id, nil)
	found, err := s.repo.Update(ctx, tenant.SchoolIDFromContext(ctx), id, domain.UpdateSubject{
		SubjectCode: code, Name: name, Credit: in.Credit,
	}, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrSubjectNotFound
	}
	return nil
}

func (s *SubjectService) Delete(ctx context.Context, id string) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditDelete, "subject", id, nil)
	found, err := s.repo.SoftDelete(ctx, tenant.SchoolIDFromContext(ctx), id, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrSubjectNotFound
	}
	return nil
}
