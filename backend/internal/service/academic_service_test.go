package service

import (
	"context"
	"errors"
	"testing"

	"github.com/chumko-platform/backend/internal/domain"
)

type fakeAcademicRepo struct {
	years     map[string]*domain.AcademicYear
	semesters map[string]*domain.Semester
	seq       int
}

func newFakeAcademicRepo() *fakeAcademicRepo {
	return &fakeAcademicRepo{years: map[string]*domain.AcademicYear{}, semesters: map[string]*domain.Semester{}}
}

func (r *fakeAcademicRepo) ListYears(_ context.Context, schoolID string) ([]domain.AcademicYear, error) {
	var out []domain.AcademicYear
	for _, y := range r.years {
		if y.SchoolID == schoolID {
			out = append(out, *y)
		}
	}
	return out, nil
}

func (r *fakeAcademicRepo) CreateYear(_ context.Context, schoolID string, year int, _ domain.AuditEntry) (string, error) {
	for _, y := range r.years {
		if y.SchoolID == schoolID && y.Year == year {
			return "", domain.ErrDuplicateYear
		}
	}
	r.seq++
	id := "y" + string(rune('0'+r.seq))
	r.years[id] = &domain.AcademicYear{ID: id, SchoolID: schoolID, Year: year}
	return id, nil
}

func (r *fakeAcademicRepo) SetCurrentYear(_ context.Context, schoolID, id string, _ domain.AuditEntry) (bool, error) {
	y, ok := r.years[id]
	if !ok || y.SchoolID != schoolID {
		return false, nil
	}
	for _, o := range r.years {
		if o.SchoolID == schoolID {
			o.IsCurrent = false
		}
	}
	y.IsCurrent = true
	return true, nil
}

func (r *fakeAcademicRepo) ListSemesters(_ context.Context, schoolID string) ([]domain.Semester, error) {
	var out []domain.Semester
	for _, s := range r.semesters {
		if s.SchoolID == schoolID {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (r *fakeAcademicRepo) CreateSemester(_ context.Context, schoolID string, ns domain.NewSemester, _ domain.AuditEntry) (string, error) {
	y, ok := r.years[ns.AcademicYearID]
	if !ok || y.SchoolID != schoolID {
		return "", domain.ErrYearNotFound
	}
	r.seq++
	id := "sm" + string(rune('0'+r.seq))
	r.semesters[id] = &domain.Semester{ID: id, SchoolID: schoolID, AcademicYearID: ns.AcademicYearID, Year: y.Year, Term: ns.Term}
	return id, nil
}

func (r *fakeAcademicRepo) SetCurrentSemester(_ context.Context, schoolID, id string, _ domain.AuditEntry) (bool, error) {
	s, ok := r.semesters[id]
	if !ok || s.SchoolID != schoolID {
		return false, nil
	}
	for _, o := range r.semesters {
		if o.SchoolID == schoolID {
			o.IsCurrent = false
		}
	}
	s.IsCurrent = true
	return true, nil
}

const acadSchool = "school-A"

func TestAcademic_CreateYearByAdmin(t *testing.T) {
	svc := NewAcademicService(newFakeAcademicRepo())
	id, err := svc.CreateYear(adminCtx(acadSchool), 2568)
	if err != nil || id == "" {
		t.Fatalf("create year: id=%q err=%v", id, err)
	}
}

func TestAcademic_CreateYearForbiddenForNonAdmin(t *testing.T) {
	svc := NewAcademicService(newFakeAcademicRepo())
	_, err := svc.CreateYear(memberCtx(acadSchool, "u9"), 2568)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestAcademic_CreateYearInvalid(t *testing.T) {
	svc := NewAcademicService(newFakeAcademicRepo())
	_, err := svc.CreateYear(adminCtx(acadSchool), 1990)
	if !errors.Is(err, domain.ErrInvalidYear) {
		t.Errorf("err = %v, want ErrInvalidYear", err)
	}
}

func TestAcademic_SetCurrentYearNotFound(t *testing.T) {
	svc := NewAcademicService(newFakeAcademicRepo())
	if err := svc.SetCurrentYear(adminCtx(acadSchool), "missing"); !errors.Is(err, domain.ErrYearNotFound) {
		t.Errorf("err = %v, want ErrYearNotFound", err)
	}
}

func TestAcademic_CreateSemesterInvalidTerm(t *testing.T) {
	repo := newFakeAcademicRepo()
	repo.years["y1"] = &domain.AcademicYear{ID: "y1", SchoolID: acadSchool, Year: 2568}
	svc := NewAcademicService(repo)
	_, err := svc.CreateSemester(adminCtx(acadSchool), NewSemesterInput{AcademicYearID: "y1", Term: 3})
	if !errors.Is(err, domain.ErrInvalidTerm) {
		t.Errorf("err = %v, want ErrInvalidTerm", err)
	}
}

func TestAcademic_CreateSemesterForbiddenForNonAdmin(t *testing.T) {
	svc := NewAcademicService(newFakeAcademicRepo())
	_, err := svc.CreateSemester(memberCtx(acadSchool, "u9"), NewSemesterInput{AcademicYearID: "y1", Term: 1})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestAcademic_CreateSemesterSuccess(t *testing.T) {
	repo := newFakeAcademicRepo()
	repo.years["y1"] = &domain.AcademicYear{ID: "y1", SchoolID: acadSchool, Year: 2568}
	svc := NewAcademicService(repo)
	id, err := svc.CreateSemester(adminCtx(acadSchool), NewSemesterInput{AcademicYearID: "y1", Term: 1})
	if err != nil || id == "" {
		t.Fatalf("create semester: id=%q err=%v", id, err)
	}
}

func TestAcademic_SetCurrentSemesterNotFound(t *testing.T) {
	svc := NewAcademicService(newFakeAcademicRepo())
	if err := svc.SetCurrentSemester(adminCtx(acadSchool), "missing"); !errors.Is(err, domain.ErrSemesterNotFound) {
		t.Errorf("err = %v, want ErrSemesterNotFound", err)
	}
}

func TestAcademic_ListSemestersAllowsNonAdmin(t *testing.T) {
	repo := newFakeAcademicRepo()
	repo.semesters["sm1"] = &domain.Semester{ID: "sm1", SchoolID: acadSchool, Year: 2568, Term: 1}
	svc := NewAcademicService(repo)
	// list ไม่ต้องเป็น admin (ใช้ในตัวสลับเทอม)
	list, err := svc.ListSemesters(memberCtx(acadSchool, "u9"))
	if err != nil {
		t.Fatalf("list semesters: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("list = %d ควรเป็น 1", len(list))
	}
}
