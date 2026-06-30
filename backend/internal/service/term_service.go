package service

import (
	"context"

	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/tenant"
)

// TermRepository contract ของชั้น DB
type TermRepository interface {
	GetCurrent(ctx context.Context, schoolID string) (*domain.CurrentTerm, error)
}

// CurrentTermDTO ปี/เทอมปัจจุบันสำหรับ response
type CurrentTermDTO struct {
	HasCurrent   bool   `json:"has_current"`
	SemesterID   string `json:"semester_id"`
	AcademicYear int    `json:"academic_year"`
	Term         int    `json:"term"`
}

// TermService ให้ข้อมูลปี/เทอมปัจจุบัน (ผู้ล็อกอินทุกคนเรียกได้ — scope ตามโรงเรียน)
type TermService struct {
	repo TermRepository
}

func NewTermService(repo TermRepository) *TermService {
	return &TermService{repo: repo}
}

func (s *TermService) Current(ctx context.Context) (*CurrentTermDTO, error) {
	t, err := s.repo.GetCurrent(ctx, tenant.SchoolIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	if t == nil {
		return &CurrentTermDTO{HasCurrent: false}, nil
	}
	return &CurrentTermDTO{
		HasCurrent:   true,
		SemesterID:   t.SemesterID,
		AcademicYear: t.AcademicYear,
		Term:         t.Term,
	}, nil
}
