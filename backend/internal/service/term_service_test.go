package service

import (
	"context"
	"testing"

	"github.com/chumko-platform/backend/internal/domain"
)

type fakeTermRepo struct {
	term map[string]*domain.CurrentTerm // key: schoolID
}

func (r *fakeTermRepo) GetCurrent(_ context.Context, schoolID string) (*domain.CurrentTerm, error) {
	return r.term[schoolID], nil
}

func TestTerm_CurrentExists(t *testing.T) {
	repo := &fakeTermRepo{term: map[string]*domain.CurrentTerm{
		"school-A": {SemesterID: "sem-1", AcademicYear: 2568, Term: 1},
	}}
	svc := NewTermService(repo)
	res, err := svc.Current(adminCtx("school-A"))
	if err != nil {
		t.Fatalf("current: %v", err)
	}
	if !res.HasCurrent || res.AcademicYear != 2568 || res.Term != 1 {
		t.Errorf("res = %+v ควร year=2568 term=1", res)
	}
}

func TestTerm_NoCurrent(t *testing.T) {
	svc := NewTermService(&fakeTermRepo{term: map[string]*domain.CurrentTerm{}})
	res, err := svc.Current(adminCtx("school-A"))
	if err != nil {
		t.Fatalf("current: %v", err)
	}
	if res.HasCurrent {
		t.Errorf("res.HasCurrent = true ควรเป็น false เมื่อยังไม่กำหนดเทอม")
	}
}
