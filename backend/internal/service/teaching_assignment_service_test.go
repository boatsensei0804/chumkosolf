package service

import (
	"context"
	"errors"
	"testing"

	"github.com/chumkosoft/backend/internal/domain"
)

// --- fake teaching assignment repo ---

type fakeTaRepo struct {
	items map[string]*domain.TeachingAssignment
	seq   int
}

func newFakeTaRepo() *fakeTaRepo {
	return &fakeTaRepo{items: map[string]*domain.TeachingAssignment{}}
}

func (r *fakeTaRepo) ListBySemester(_ context.Context, schoolID, semesterID string) ([]domain.TeachingAssignment, error) {
	var out []domain.TeachingAssignment
	for _, a := range r.items {
		if a.SchoolID == schoolID && a.SemesterID == semesterID {
			out = append(out, *a)
		}
	}
	return out, nil
}

func (r *fakeTaRepo) GetByID(_ context.Context, schoolID, id string) (*domain.TeachingAssignment, error) {
	a, ok := r.items[id]
	if !ok || a.SchoolID != schoolID {
		return nil, nil
	}
	cp := *a
	return &cp, nil
}

func (r *fakeTaRepo) Create(_ context.Context, schoolID, semesterID string, na domain.NewTeachingAssignment, _ domain.AuditEntry) (string, error) {
	r.seq++
	id := "ta" + string(rune('0'+r.seq))
	r.items[id] = &domain.TeachingAssignment{
		ID: id, SchoolID: schoolID, SemesterID: semesterID,
		PersonnelID: na.PersonnelID, SubjectID: na.SubjectID, ClassID: na.ClassID,
	}
	return id, nil
}

func (r *fakeTaRepo) SoftDelete(_ context.Context, schoolID, id string, _ domain.AuditEntry) (bool, error) {
	a, ok := r.items[id]
	if !ok || a.SchoolID != schoolID {
		return false, nil
	}
	delete(r.items, id)
	return true, nil
}

const taSchool = "school-A"

func newTaSvc() (*TeachingAssignmentService, *fakeChecker) {
	repo := newFakeTaRepo()
	personnel := guardWith(taSchool, "person-1") // *fakePersonnelRepo: มี personnel "person-1"
	subjects := newFakeSubjectRepo()
	subjects.items["subj1"] = &domain.Subject{ID: "subj1", SchoolID: taSchool, SubjectCode: "ค21101", Name: "คณิต"}
	classes := &fakeClassExists{byID: map[string]*domain.Class{
		"c1": {ID: "c1", SchoolID: taSchool, SemesterID: "sem-1"},
	}}
	checker := &fakeChecker{groups: map[string]bool{}}
	return NewTeachingAssignmentService(repo, personnel, subjects, classes, checker), checker
}

func validTaInput() TeachingAssignmentInput {
	return TeachingAssignmentInput{PersonnelID: "person-1", SubjectID: "subj1", ClassID: "c1"}
}

// --- tests ---

func TestTeaching_CreateSuccess(t *testing.T) {
	svc, _ := newTaSvc()
	id, err := svc.Create(wAdmin(taSchool, "sem-1"), validTaInput())
	if err != nil || id == "" {
		t.Fatalf("create: id=%q err=%v", id, err)
	}
}

func TestTeaching_PersonnelNotFound(t *testing.T) {
	svc, _ := newTaSvc()
	in := validTaInput()
	in.PersonnelID = "ghost"
	_, err := svc.Create(wAdmin(taSchool, "sem-1"), in)
	if !errors.Is(err, domain.ErrPersonnelNotFound) {
		t.Errorf("err = %v, want ErrPersonnelNotFound", err)
	}
}

func TestTeaching_SubjectNotFound(t *testing.T) {
	svc, _ := newTaSvc()
	in := validTaInput()
	in.SubjectID = "ghost"
	_, err := svc.Create(wAdmin(taSchool, "sem-1"), in)
	if !errors.Is(err, domain.ErrSubjectNotFound) {
		t.Errorf("err = %v, want ErrSubjectNotFound", err)
	}
}

func TestTeaching_ClassNotFound(t *testing.T) {
	svc, _ := newTaSvc()
	in := validTaInput()
	in.ClassID = "ghost"
	_, err := svc.Create(wAdmin(taSchool, "sem-1"), in)
	if !errors.Is(err, domain.ErrClassNotFound) {
		t.Errorf("err = %v, want ErrClassNotFound", err)
	}
}

func TestTeaching_NoSemester(t *testing.T) {
	svc, _ := newTaSvc()
	_, err := svc.Create(adminCtx(taSchool), validTaInput())
	if !errors.Is(err, domain.ErrNoActiveSemester) {
		t.Errorf("err = %v, want ErrNoActiveSemester", err)
	}
}

func TestTeaching_ForbiddenForNonMember(t *testing.T) {
	svc, _ := newTaSvc()
	_, err := svc.Create(wMember(taSchool, "u9", "sem-1"), validTaInput())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestTeaching_CrossSchoolPersonnelNotFound(t *testing.T) {
	svc, _ := newTaSvc()
	// scope school-B แต่ personnel/subject/class อยู่ school-A → ไม่พบ (isolation)
	_, err := svc.Create(wAdmin("school-B", "sem-1"), validTaInput())
	if !errors.Is(err, domain.ErrPersonnelNotFound) {
		t.Errorf("err = %v, want ErrPersonnelNotFound", err)
	}
}

func TestTeaching_DeleteNotFound(t *testing.T) {
	svc, _ := newTaSvc()
	if err := svc.Delete(wAdmin(taSchool, "sem-1"), "missing"); !errors.Is(err, domain.ErrTeachingAssignmentNotFound) {
		t.Errorf("err = %v, want ErrTeachingAssignmentNotFound", err)
	}
}
