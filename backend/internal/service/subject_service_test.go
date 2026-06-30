package service

import (
	"context"
	"errors"
	"testing"

	"github.com/chumkosoft/backend/internal/domain"
)

// --- fake subject repo ---

type fakeSubjectRepo struct {
	items map[string]*domain.Subject
	seq   int
}

func newFakeSubjectRepo() *fakeSubjectRepo {
	return &fakeSubjectRepo{items: map[string]*domain.Subject{}}
}

func (r *fakeSubjectRepo) List(_ context.Context, schoolID string) ([]domain.Subject, error) {
	var out []domain.Subject
	for _, s := range r.items {
		if s.SchoolID == schoolID {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (r *fakeSubjectRepo) GetByID(_ context.Context, schoolID, id string) (*domain.Subject, error) {
	s, ok := r.items[id]
	if !ok || s.SchoolID != schoolID {
		return nil, nil
	}
	cp := *s
	return &cp, nil
}

func (r *fakeSubjectRepo) Create(_ context.Context, schoolID string, ns domain.NewSubject, _ domain.AuditEntry) (string, error) {
	r.seq++
	id := "subj" + string(rune('0'+r.seq))
	r.items[id] = &domain.Subject{ID: id, SchoolID: schoolID, SubjectCode: ns.SubjectCode, Name: ns.Name, Credit: ns.Credit}
	return id, nil
}

func (r *fakeSubjectRepo) Update(_ context.Context, schoolID, id string, us domain.UpdateSubject, _ domain.AuditEntry) (bool, error) {
	s, ok := r.items[id]
	if !ok || s.SchoolID != schoolID {
		return false, nil
	}
	s.SubjectCode = us.SubjectCode
	s.Name = us.Name
	s.Credit = us.Credit
	return true, nil
}

func (r *fakeSubjectRepo) SoftDelete(_ context.Context, schoolID, id string, _ domain.AuditEntry) (bool, error) {
	s, ok := r.items[id]
	if !ok || s.SchoolID != schoolID {
		return false, nil
	}
	delete(r.items, id)
	return true, nil
}

const subjSchool = "school-A"

// --- subject tests ---

func TestSubject_CreateSuccess(t *testing.T) {
	svc := NewSubjectService(newFakeSubjectRepo(), &fakeChecker{groups: map[string]bool{}})
	id, err := svc.Create(adminCtx(subjSchool), SubjectInput{SubjectCode: "ค21101", Name: "คณิตศาสตร์"})
	if err != nil || id == "" {
		t.Fatalf("create: id=%q err=%v", id, err)
	}
}

func TestSubject_CreateRequiresCodeAndName(t *testing.T) {
	svc := NewSubjectService(newFakeSubjectRepo(), &fakeChecker{groups: map[string]bool{}})
	_, err := svc.Create(adminCtx(subjSchool), SubjectInput{SubjectCode: "  ", Name: "x"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
}

func TestSubject_ForbiddenForNonMember(t *testing.T) {
	svc := NewSubjectService(newFakeSubjectRepo(), &fakeChecker{groups: map[string]bool{}})
	_, err := svc.Create(memberCtx(subjSchool, "u9"), SubjectInput{SubjectCode: "ค21101", Name: "คณิต"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestSubject_AcademicMemberCanCreate(t *testing.T) {
	checker := &fakeChecker{groups: map[string]bool{subjSchool + "|u9|academic": true}}
	svc := NewSubjectService(newFakeSubjectRepo(), checker)
	id, err := svc.Create(memberCtx(subjSchool, "u9"), SubjectInput{SubjectCode: "ว21101", Name: "วิทยาศาสตร์"})
	if err != nil || id == "" {
		t.Fatalf("create by academic member: id=%q err=%v", id, err)
	}
}

func TestSubject_UpdateNotFound(t *testing.T) {
	svc := NewSubjectService(newFakeSubjectRepo(), &fakeChecker{groups: map[string]bool{}})
	err := svc.Update(adminCtx(subjSchool), "missing", SubjectInput{SubjectCode: "ค", Name: "ค"})
	if !errors.Is(err, domain.ErrSubjectNotFound) {
		t.Errorf("err = %v, want ErrSubjectNotFound", err)
	}
}

func TestSubject_DeleteNotFound(t *testing.T) {
	svc := NewSubjectService(newFakeSubjectRepo(), &fakeChecker{groups: map[string]bool{}})
	if err := svc.Delete(adminCtx(subjSchool), "missing"); !errors.Is(err, domain.ErrSubjectNotFound) {
		t.Errorf("err = %v, want ErrSubjectNotFound", err)
	}
}
