package service

import (
	"context"
	"time"

	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/tenant"
)

// AcademicRepository contract ของชั้น DB
type AcademicRepository interface {
	ListYears(ctx context.Context, schoolID string) ([]domain.AcademicYear, error)
	CreateYear(ctx context.Context, schoolID string, year int, audit domain.AuditEntry) (string, error)
	SetCurrentYear(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error)
	ListSemesters(ctx context.Context, schoolID string) ([]domain.Semester, error)
	CreateSemester(ctx context.Context, schoolID string, ns domain.NewSemester, audit domain.AuditEntry) (string, error)
	SetCurrentSemester(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error)
}

// DTOs
type AcademicYearDTO struct {
	ID        string `json:"id"`
	Year      int    `json:"year"`
	IsCurrent bool   `json:"is_current"`
}

type SemesterDTO struct {
	ID             string `json:"id"`
	AcademicYearID string `json:"academic_year_id"`
	Year           int    `json:"year"`
	Term           int    `json:"term"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	IsCurrent      bool   `json:"is_current"`
}

// Inputs
type NewSemesterInput struct {
	AcademicYearID string
	Term           int
	StartDate      *time.Time
	EndDate        *time.Time
}

// AcademicService จัดการปีการศึกษา/ภาคเรียน — อ่านได้ทุกคน, แก้ไขเฉพาะ school admin
type AcademicService struct {
	repo AcademicRepository
}

func NewAcademicService(repo AcademicRepository) *AcademicService {
	return &AcademicService{repo: repo}
}

func requireSchoolAdmin(ctx context.Context) error {
	if !tenant.IsSchoolAdminFromContext(ctx) {
		return domain.ErrForbidden
	}
	return nil
}

// --- ปีการศึกษา ---

func (s *AcademicService) ListYears(ctx context.Context) ([]AcademicYearDTO, error) {
	rows, err := s.repo.ListYears(ctx, tenant.SchoolIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	out := make([]AcademicYearDTO, 0, len(rows))
	for i := range rows {
		out = append(out, AcademicYearDTO{ID: rows[i].ID, Year: rows[i].Year, IsCurrent: rows[i].IsCurrent})
	}
	return out, nil
}

func (s *AcademicService) CreateYear(ctx context.Context, year int) (string, error) {
	if err := requireSchoolAdmin(ctx); err != nil {
		return "", err
	}
	if year < 2400 || year > 2700 {
		return "", domain.ErrInvalidYear
	}
	audit := auditFor(ctx, domain.AuditCreate, "academic_year", "", map[string]any{"year": year})
	return s.repo.CreateYear(ctx, tenant.SchoolIDFromContext(ctx), year, audit)
}

func (s *AcademicService) SetCurrentYear(ctx context.Context, id string) error {
	if err := requireSchoolAdmin(ctx); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditUpdate, "academic_year", id, map[string]any{"action": "set_current"})
	found, err := s.repo.SetCurrentYear(ctx, tenant.SchoolIDFromContext(ctx), id, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrYearNotFound
	}
	return nil
}

// --- ภาคเรียน ---

func (s *AcademicService) ListSemesters(ctx context.Context) ([]SemesterDTO, error) {
	rows, err := s.repo.ListSemesters(ctx, tenant.SchoolIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	out := make([]SemesterDTO, 0, len(rows))
	for i := range rows {
		out = append(out, SemesterDTO{
			ID: rows[i].ID, AcademicYearID: rows[i].AcademicYearID, Year: rows[i].Year, Term: rows[i].Term,
			StartDate: dateStr(rows[i].StartDate), EndDate: dateStr(rows[i].EndDate), IsCurrent: rows[i].IsCurrent,
		})
	}
	return out, nil
}

func (s *AcademicService) CreateSemester(ctx context.Context, in NewSemesterInput) (string, error) {
	if err := requireSchoolAdmin(ctx); err != nil {
		return "", err
	}
	if in.Term != 1 && in.Term != 2 {
		return "", domain.ErrInvalidTerm
	}
	if in.AcademicYearID == "" {
		return "", domain.ErrValidation
	}
	audit := auditFor(ctx, domain.AuditCreate, "semester", "", map[string]any{"academic_year_id": in.AcademicYearID, "term": in.Term})
	return s.repo.CreateSemester(ctx, tenant.SchoolIDFromContext(ctx), domain.NewSemester{
		AcademicYearID: in.AcademicYearID, Term: in.Term, StartDate: in.StartDate, EndDate: in.EndDate,
	}, audit)
}

func (s *AcademicService) SetCurrentSemester(ctx context.Context, id string) error {
	if err := requireSchoolAdmin(ctx); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditUpdate, "semester", id, map[string]any{"action": "set_current"})
	found, err := s.repo.SetCurrentSemester(ctx, tenant.SchoolIDFromContext(ctx), id, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrSemesterNotFound
	}
	return nil
}
