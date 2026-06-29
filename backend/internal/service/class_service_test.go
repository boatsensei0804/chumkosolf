package service

import (
	"context"
	"errors"
	"testing"

	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/tenant"
)

// --- fakes ---

type fakeClassRepo struct {
	byID   map[string]*domain.Class
	combos map[string]bool // semester|grade|room
	seq    int
}

func newFakeClassRepo() *fakeClassRepo {
	return &fakeClassRepo{byID: map[string]*domain.Class{}, combos: map[string]bool{}}
}
func (r *fakeClassRepo) ListBySemester(_ context.Context, schoolID, sem string) ([]domain.Class, error) {
	var out []domain.Class
	for _, c := range r.byID {
		if c.SchoolID == schoolID && c.SemesterID == sem {
			out = append(out, *c)
		}
	}
	return out, nil
}
func (r *fakeClassRepo) GetByID(_ context.Context, schoolID, id string) (*domain.Class, error) {
	c, ok := r.byID[id]
	if !ok || c.SchoolID != schoolID {
		return nil, nil
	}
	cp := *c
	return &cp, nil
}
func (r *fakeClassRepo) Create(_ context.Context, schoolID, sem string, nc domain.NewClass, _ domain.AuditEntry) (string, error) {
	key := sem + "|" + nc.GradeLevel + "|" + nc.RoomName
	if r.combos[key] {
		return "", domain.ErrDuplicateClass
	}
	r.seq++
	id := "c" + string(rune('0'+r.seq))
	r.byID[id] = &domain.Class{ID: id, SchoolID: schoolID, SemesterID: sem, GradeLevel: nc.GradeLevel, RoomName: nc.RoomName}
	r.combos[key] = true
	return id, nil
}
func (r *fakeClassRepo) Update(_ context.Context, schoolID, id string, _ domain.UpdateClass, _ domain.AuditEntry) (bool, error) {
	c, ok := r.byID[id]
	return ok && c.SchoolID == schoolID, nil
}
func (r *fakeClassRepo) SoftDelete(_ context.Context, schoolID, id string, _ domain.AuditEntry) (bool, error) {
	c, ok := r.byID[id]
	if !ok || c.SchoolID != schoolID {
		return false, nil
	}
	delete(r.byID, id)
	return true, nil
}

type fakeAdvisorRepo struct{ added map[string]bool }

func (r *fakeAdvisorRepo) ListByClass(_ context.Context, _, _ string) ([]domain.ClassAdvisor, error) {
	return nil, nil
}
func (r *fakeAdvisorRepo) Add(_ context.Context, _, _, classID, personnelID string, _ domain.AuditEntry) error {
	r.added[classID+"|"+personnelID] = true
	return nil
}
func (r *fakeAdvisorRepo) Remove(_ context.Context, _, _, advisorID string, _ domain.AuditEntry) (bool, error) {
	return advisorID == "exists", nil
}

type fakeEnrollRepo struct{ classOfStudent map[string]string }

func (r *fakeEnrollRepo) ListByClass(_ context.Context, _, _ string) ([]domain.ClassEnrollment, error) {
	return nil, nil
}
func (r *fakeEnrollRepo) Enroll(_ context.Context, _, _, classID string, ne domain.NewEnrollment, _ domain.AuditEntry) error {
	r.classOfStudent[ne.StudentID] = classID // 1 ห้องต่อนักเรียน (เทอมนี้)
	return nil
}
func (r *fakeEnrollRepo) Remove(_ context.Context, _, _, enrollmentID string, _ domain.AuditEntry) (bool, error) {
	return enrollmentID == "exists", nil
}

// --- helpers ---

func acadAdminCtx(schoolID string) context.Context {
	return tenant.WithIdentity(context.Background(), tenant.Identity{
		UserID: "admin", SchoolID: schoolID, Role: "super_admin", IsSchoolAdmin: true, SemesterID: "sem-1",
	})
}
func acadMemberCtx(schoolID, userID string) context.Context {
	return tenant.WithIdentity(context.Background(), tenant.Identity{
		UserID: userID, SchoolID: schoolID, Role: "teacher", SemesterID: "sem-1",
	})
}

// --- class tests ---

func TestClassCreate_Success(t *testing.T) {
	svc := NewClassService(newFakeClassRepo(), &fakeWGChecker{})
	id, err := svc.Create(acadAdminCtx("school-A"), ClassInput{GradeLevel: "ม.1", RoomName: "1/1"})
	if err != nil || id == "" {
		t.Fatalf("create: id=%q err=%v", id, err)
	}
}

func TestClassCreate_NoSemester(t *testing.T) {
	svc := NewClassService(newFakeClassRepo(), &fakeWGChecker{})
	// adminCtx ไม่มี SemesterID
	_, err := svc.Create(adminCtx("school-A"), ClassInput{GradeLevel: "ม.1", RoomName: "1/1"})
	if !errors.Is(err, domain.ErrNoActiveSemester) {
		t.Errorf("err = %v, want ErrNoActiveSemester", err)
	}
}

func TestClassCreate_Forbidden(t *testing.T) {
	svc := NewClassService(newFakeClassRepo(), &fakeWGChecker{})
	_, err := svc.Create(acadMemberCtx("school-A", "u9"), ClassInput{GradeLevel: "ม.1", RoomName: "1/1"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestClassCreate_AllowedAcademicMember(t *testing.T) {
	svc := NewClassService(newFakeClassRepo(), academicChecker("school-A", "u9"))
	if _, err := svc.Create(acadMemberCtx("school-A", "u9"), ClassInput{GradeLevel: "ม.1", RoomName: "1/1"}); err != nil {
		t.Errorf("academic member ควรสร้างได้: %v", err)
	}
}

func TestClassCreate_Duplicate(t *testing.T) {
	repo := newFakeClassRepo()
	svc := NewClassService(repo, &fakeWGChecker{})
	ctx := acadAdminCtx("school-A")
	if _, err := svc.Create(ctx, ClassInput{GradeLevel: "ม.1", RoomName: "1/1"}); err != nil {
		t.Fatalf("first: %v", err)
	}
	if _, err := svc.Create(ctx, ClassInput{GradeLevel: "ม.1", RoomName: "1/1"}); !errors.Is(err, domain.ErrDuplicateClass) {
		t.Errorf("err = %v, want ErrDuplicateClass", err)
	}
}

func TestClassGet_NotFound(t *testing.T) {
	svc := NewClassService(newFakeClassRepo(), &fakeWGChecker{})
	if _, err := svc.Get(acadAdminCtx("school-A"), "missing"); !errors.Is(err, domain.ErrClassNotFound) {
		t.Errorf("err = %v, want ErrClassNotFound", err)
	}
}

// --- advisor tests ---

func TestAdvisorAdd_ClassAndPersonnelMustExist(t *testing.T) {
	classRepo := newFakeClassRepo()
	per := newFakePersonnelRepo()
	svc := NewClassAdvisorService(&fakeAdvisorRepo{added: map[string]bool{}}, classRepo, per, &fakeWGChecker{})
	ctx := acadAdminCtx("school-A")

	// class ไม่มี
	if err := svc.Add(ctx, "missing", "p1"); !errors.Is(err, domain.ErrClassNotFound) {
		t.Errorf("err = %v, want ErrClassNotFound", err)
	}
	// มี class แต่ personnel ไม่มี
	cid, _ := NewClassService(classRepo, &fakeWGChecker{}).Create(ctx, ClassInput{GradeLevel: "ม.1", RoomName: "1/1"})
	if err := svc.Add(ctx, cid, "missing"); !errors.Is(err, domain.ErrPersonnelNotFound) {
		t.Errorf("err = %v, want ErrPersonnelNotFound", err)
	}
	// ครบ → ผ่าน
	per.byID["p1"] = &domain.Personnel{ID: "p1", SchoolID: "school-A"}
	if err := svc.Add(ctx, cid, "p1"); err != nil {
		t.Errorf("add advisor: %v", err)
	}
}

// --- enrollment tests ---

func TestEnroll_MoveSingleClassPerSemester(t *testing.T) {
	classRepo := newFakeClassRepo()
	stu := newFakeStudentRepo()
	enr := &fakeEnrollRepo{classOfStudent: map[string]string{}}
	svc := NewEnrollmentService(enr, classRepo, stu, &fakeWGChecker{})
	ctx := acadAdminCtx("school-A")

	c1, _ := NewClassService(classRepo, &fakeWGChecker{}).Create(ctx, ClassInput{GradeLevel: "ม.1", RoomName: "1/1"})
	c2, _ := NewClassService(classRepo, &fakeWGChecker{}).Create(ctx, ClassInput{GradeLevel: "ม.1", RoomName: "1/2"})
	stu.byID["s1"] = &domain.Student{ID: "s1", SchoolID: "school-A"}

	if err := svc.Enroll(ctx, c1, EnrollInput{StudentID: "s1"}); err != nil {
		t.Fatalf("enroll c1: %v", err)
	}
	if err := svc.Enroll(ctx, c2, EnrollInput{StudentID: "s1"}); err != nil {
		t.Fatalf("enroll c2: %v", err)
	}
	if enr.classOfStudent["s1"] != c2 {
		t.Errorf("นักเรียนควรย้ายไปห้อง c2 (ห้องเดียวต่อเทอม), got %q", enr.classOfStudent["s1"])
	}
}

func TestEnrollMany_BulkAndSkipMissing(t *testing.T) {
	classRepo := newFakeClassRepo()
	stu := newFakeStudentRepo()
	enr := &fakeEnrollRepo{classOfStudent: map[string]string{}}
	svc := NewEnrollmentService(enr, classRepo, stu, &fakeWGChecker{})
	ctx := acadAdminCtx("school-A")

	cid, _ := NewClassService(classRepo, &fakeWGChecker{}).Create(ctx, ClassInput{GradeLevel: "ม.1", RoomName: "1/1"})
	stu.byID["s1"] = &domain.Student{ID: "s1", SchoolID: "school-A"}
	stu.byID["s2"] = &domain.Student{ID: "s2", SchoolID: "school-A"}

	// s1, s2 มีจริง; "ghost" ไม่มี (ต้องข้าม); s1 ซ้ำ (ต้องนับครั้งเดียว)
	count, err := svc.EnrollMany(ctx, cid, []string{"s1", "s2", "ghost", "s1"})
	if err != nil {
		t.Fatalf("enroll many: %v", err)
	}
	if count != 2 {
		t.Errorf("enrolled = %d, want 2 (ข้าม ghost + ตัดซ้ำ s1)", count)
	}
	if enr.classOfStudent["s1"] != cid || enr.classOfStudent["s2"] != cid {
		t.Errorf("นักเรียนควรเข้าห้อง %q ทั้งคู่: %+v", cid, enr.classOfStudent)
	}
}

func TestEnrollMany_ForbiddenForNonMember(t *testing.T) {
	classRepo := newFakeClassRepo()
	svc := NewEnrollmentService(&fakeEnrollRepo{classOfStudent: map[string]string{}}, classRepo, newFakeStudentRepo(), &fakeWGChecker{})
	cid, _ := NewClassService(classRepo, &fakeWGChecker{}).Create(acadAdminCtx("school-A"), ClassInput{GradeLevel: "ม.1", RoomName: "1/1"})
	if _, err := svc.EnrollMany(acadMemberCtx("school-A", "u9"), cid, []string{"s1"}); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestEnroll_StudentMustExist(t *testing.T) {
	classRepo := newFakeClassRepo()
	svc := NewEnrollmentService(&fakeEnrollRepo{classOfStudent: map[string]string{}}, classRepo, newFakeStudentRepo(), &fakeWGChecker{})
	ctx := acadAdminCtx("school-A")
	cid, _ := NewClassService(classRepo, &fakeWGChecker{}).Create(ctx, ClassInput{GradeLevel: "ม.1", RoomName: "1/1"})
	if err := svc.Enroll(ctx, cid, EnrollInput{StudentID: "missing"}); !errors.Is(err, domain.ErrStudentNotFound) {
		t.Errorf("err = %v, want ErrStudentNotFound", err)
	}
}
