package service

import (
	"context"
	"errors"
	"testing"

	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/tenant"
)

type fakeDirClassRepo struct{ classes []domain.Class }

func (r *fakeDirClassRepo) ListBySemester(_ context.Context, _, _ string) ([]domain.Class, error) {
	return r.classes, nil
}

type fakeDirEnrollRepo struct {
	roster []domain.ClassEnrollment
	search []domain.StudentClassBrief
}

func (r *fakeDirEnrollRepo) ListByClass(_ context.Context, _, _ string) ([]domain.ClassEnrollment, error) {
	return r.roster, nil
}
func (r *fakeDirEnrollRepo) SearchByName(_ context.Context, _, _, _ string) ([]domain.StudentClassBrief, error) {
	return r.search, nil
}

func dirTeacherCtx(schoolID string) context.Context {
	return tenant.WithIdentity(context.Background(), tenant.Identity{
		UserID: "u1", SchoolID: schoolID, Role: "teacher", SemesterID: "sem-1",
	})
}
func dirStudentCtx(schoolID string) context.Context {
	return tenant.WithIdentity(context.Background(), tenant.Identity{
		UserID: "stu", SchoolID: schoolID, Role: "student", SemesterID: "sem-1",
	})
}

func TestDirectory_StudentForbidden(t *testing.T) {
	svc := NewDirectoryService(&fakeDirClassRepo{}, &fakeDirEnrollRepo{})
	if _, err := svc.Classes(dirStudentCtx("school-A")); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("Classes err = %v, want ErrForbidden", err)
	}
	if _, err := svc.SearchStudents(dirStudentCtx("school-A"), "ก"); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("SearchStudents err = %v, want ErrForbidden", err)
	}
}

func TestDirectory_ClassesForTeacher(t *testing.T) {
	repo := &fakeDirClassRepo{classes: []domain.Class{{ID: "c1", GradeLevel: "ม.1", RoomName: "1", StudentCount: 30}}}
	svc := NewDirectoryService(repo, &fakeDirEnrollRepo{})
	out, err := svc.Classes(dirTeacherCtx("school-A"))
	if err != nil {
		t.Fatalf("classes: %v", err)
	}
	if len(out) != 1 || out[0].StudentCount != 30 {
		t.Errorf("classes = %+v", out)
	}
}

func TestDirectory_SearchEmptyTermReturnsEmpty(t *testing.T) {
	svc := NewDirectoryService(&fakeDirClassRepo{}, &fakeDirEnrollRepo{search: []domain.StudentClassBrief{{StudentID: "s1"}}})
	out, err := svc.SearchStudents(dirTeacherCtx("school-A"), "   ")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("คำค้นว่าง → ต้องคืนว่าง, ได้ %d", len(out))
	}
}

func TestDirectory_SearchReturnsClassLabel(t *testing.T) {
	repo := &fakeDirEnrollRepo{search: []domain.StudentClassBrief{
		{StudentID: "s1", StudentCode: "S001", FirstName: "เด็กชาย", LastName: "ก", GradeLevel: "ม.2", RoomName: "3"},
	}}
	svc := NewDirectoryService(&fakeDirClassRepo{}, repo)
	out, err := svc.SearchStudents(dirTeacherCtx("school-A"), "เด็ก")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(out) != 1 || out[0].ClassLabel != "ม.2 3" {
		t.Errorf("search = %+v", out)
	}
}
