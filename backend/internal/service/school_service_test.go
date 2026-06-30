package service

import (
	"context"
	"errors"
	"testing"

	"github.com/chumkosoft/backend/internal/domain"
)

type fakeSchoolRepo struct {
	byID map[string]*domain.School
}

func (r *fakeSchoolRepo) Get(_ context.Context, schoolID string) (*domain.School, error) {
	return r.byID[schoolID], nil
}
func (r *fakeSchoolRepo) Update(_ context.Context, schoolID string, us domain.UpdateSchool, _ domain.AuditEntry) (bool, error) {
	sc, ok := r.byID[schoolID]
	if !ok {
		return false, nil
	}
	sc.Name = us.Name
	sc.Address = us.Address
	sc.Phone = us.Phone
	return true, nil
}

func newSchoolSvc() (*SchoolService, *fakeSchoolRepo) {
	repo := &fakeSchoolRepo{byID: map[string]*domain.School{
		"school-A": {ID: "school-A", Name: "โรงเรียนชุมโค", Code: "CHUMKO"},
	}}
	return NewSchoolService(repo), repo
}

func TestSchool_GetByAnyUser(t *testing.T) {
	svc, _ := newSchoolSvc()
	sc, err := svc.Get(memberCtx("school-A", "u9"))
	if err != nil || sc.Code != "CHUMKO" {
		t.Fatalf("get school: %+v err=%v", sc, err)
	}
}

func TestSchool_UpdateByAdmin(t *testing.T) {
	svc, repo := newSchoolSvc()
	if err := svc.Update(adminCtx("school-A"), UpdateSchoolInput{Name: "โรงเรียนใหม่", Phone: "077000000"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if repo.byID["school-A"].Name != "โรงเรียนใหม่" {
		t.Error("ชื่อโรงเรียนไม่ถูกอัปเดต")
	}
}

func TestSchool_UpdateForbiddenForNonAdmin(t *testing.T) {
	svc, _ := newSchoolSvc()
	if err := svc.Update(memberCtx("school-A", "u9"), UpdateSchoolInput{Name: "x"}); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestSchool_UpdateRequiresName(t *testing.T) {
	svc, _ := newSchoolSvc()
	if err := svc.Update(adminCtx("school-A"), UpdateSchoolInput{Name: "  "}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
}
