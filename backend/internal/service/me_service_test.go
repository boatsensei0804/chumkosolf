package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chumko-platform/backend/internal/crypto"
	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/tenant"
)

// --- fakes ---

type fakeMePersonnelRepo struct {
	byUser  map[string]*domain.Personnel // key: schoolID|userID
	updated bool
}

func (r *fakeMePersonnelRepo) GetByUserID(_ context.Context, schoolID, userID string) (*domain.Personnel, error) {
	p, ok := r.byUser[schoolID+"|"+userID]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (r *fakeMePersonnelRepo) Update(_ context.Context, schoolID, id string, _ domain.UpdatePersonnel, _ domain.AuditEntry) (bool, error) {
	for _, p := range r.byUser {
		if p.SchoolID == schoolID && p.ID == id {
			r.updated = true
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeMePersonnelRepo) InsertAudit(_ context.Context, _ domain.AuditEntry) error { return nil }

type fakeAdviseeRepo struct {
	rows      []domain.Advisee
	owns      map[string]bool // key: schoolID|semesterID|userID|studentID
	gotSchool string
	gotUser   string
}

func (r *fakeAdviseeRepo) Advisees(_ context.Context, schoolID, _, userID string, _ time.Time) ([]domain.Advisee, error) {
	r.gotSchool = schoolID
	r.gotUser = userID
	return r.rows, nil
}

func (r *fakeAdviseeRepo) IsAdvisee(_ context.Context, schoolID, semesterID, userID, studentID string) (bool, error) {
	return r.owns[schoolID+"|"+semesterID+"|"+userID+"|"+studentID], nil
}

type fakeMeStudentRepo struct {
	byID    map[string]*domain.Student
	updated *domain.UpdateStudent
}

func (r *fakeMeStudentRepo) GetByID(_ context.Context, schoolID, id string) (*domain.Student, error) {
	s, ok := r.byID[id]
	if !ok || s.SchoolID != schoolID {
		return nil, nil
	}
	cp := *s
	return &cp, nil
}

func (r *fakeMeStudentRepo) Update(_ context.Context, schoolID, id string, us domain.UpdateStudent, _ domain.AuditEntry) (bool, error) {
	s, ok := r.byID[id]
	if !ok || s.SchoolID != schoolID {
		return false, nil
	}
	cp := us
	r.updated = &cp
	return true, nil
}

// --- helpers ---

func meSemCtx(schoolID, userID, sem string) context.Context {
	return tenant.WithIdentity(context.Background(), tenant.Identity{
		UserID: userID, SchoolID: schoolID, SemesterID: sem, Role: "teacher",
	})
}

func personnelWithNID(t *testing.T, cipher Cipher, schoolID, userID, id, nid string) *domain.Personnel {
	t.Helper()
	enc, err := cipher.Encrypt(nid)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	return &domain.Personnel{
		ID: id, SchoolID: schoolID, UserID: userID, NationalIDEnc: enc,
		Username: "kru.somchai", Role: "teacher", IsActive: true,
		Profile: domain.PersonnelProfile{Prefix: "นาย", FirstName: "สมชาย", LastName: "ใจดี"},
	}
}

func meSvcWithProfile(t *testing.T, cipher Cipher) (*MeService, *fakeMePersonnelRepo) {
	t.Helper()
	repo := &fakeMePersonnelRepo{byUser: map[string]*domain.Personnel{
		"school-A|u1": personnelWithNID(t, cipher, "school-A", "u1", "p1", "1234567890123"),
	}}
	return NewMeService(repo, &fakeAdviseeRepo{}, &fakeMeStudentRepo{}, cipher), repo
}

// --- profile tests ---

func TestMeProfile_SuccessMask(t *testing.T) {
	cipher := testCipher(t)
	svc, _ := meSvcWithProfile(t, cipher)

	d, err := svc.Profile(meSemCtx("school-A", "u1", ""))
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	if d.NationalIDMasked != crypto.Mask("1234567890123") {
		t.Errorf("national_id ต้อง mask: %q", d.NationalIDMasked)
	}
	if d.FirstName != "สมชาย" {
		t.Errorf("first_name = %q", d.FirstName)
	}
}

func TestMeProfile_NotFound(t *testing.T) {
	svc := NewMeService(&fakeMePersonnelRepo{byUser: map[string]*domain.Personnel{}}, &fakeAdviseeRepo{}, &fakeMeStudentRepo{}, testCipher(t))
	_, err := svc.Profile(meSemCtx("school-A", "ghost", ""))
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != "PROFILE_NOT_FOUND" {
		t.Errorf("err = %v, want PROFILE_NOT_FOUND", err)
	}
}

func TestMeProfile_Isolation(t *testing.T) {
	svc, _ := meSvcWithProfile(t, testCipher(t))
	// user เดียวกันแต่คนละโรงเรียน → ไม่พบ (scope ด้วย school_id)
	if _, err := svc.Profile(meSemCtx("school-B", "u1", "")); err == nil {
		t.Error("ข้ามโรงเรียนต้องไม่พบโปรไฟล์")
	}
}

func TestMeUpdateProfile_Validation(t *testing.T) {
	svc, _ := meSvcWithProfile(t, testCipher(t))
	err := svc.UpdateProfile(meSemCtx("school-A", "u1", ""), UpdateMyProfileInput{FirstName: "", LastName: "ใจดี"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
}

func TestMeUpdateProfile_Success(t *testing.T) {
	svc, repo := meSvcWithProfile(t, testCipher(t))
	if err := svc.UpdateProfile(meSemCtx("school-A", "u1", ""), UpdateMyProfileInput{FirstName: "สมชาย", LastName: "ใจเย็น", Phone: "0810000000"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if !repo.updated {
		t.Error("ควรเรียก repo.Update")
	}
}

func TestMeUpdateProfile_NotFound(t *testing.T) {
	svc := NewMeService(&fakeMePersonnelRepo{byUser: map[string]*domain.Personnel{}}, &fakeAdviseeRepo{}, &fakeMeStudentRepo{}, testCipher(t))
	err := svc.UpdateProfile(meSemCtx("school-A", "ghost", ""), UpdateMyProfileInput{FirstName: "x", LastName: "y"})
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != "PROFILE_NOT_FOUND" {
		t.Errorf("err = %v, want PROFILE_NOT_FOUND", err)
	}
}

// --- advisees (list) tests ---

func TestMeAdvisees_NoSemester(t *testing.T) {
	adv := &fakeAdviseeRepo{rows: []domain.Advisee{{StudentID: "s1"}}}
	svc := NewMeService(&fakeMePersonnelRepo{byUser: map[string]*domain.Personnel{}}, adv, &fakeMeStudentRepo{}, testCipher(t))
	out, err := svc.Advisees(meSemCtx("school-A", "u1", "")) // ไม่มีเทอม
	if err != nil {
		t.Fatalf("advisees: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("ไม่มีเทอม → ต้องคืนว่าง, ได้ %d", len(out))
	}
}

func TestMeAdvisees_ListMaskAndLabel(t *testing.T) {
	cipher := testCipher(t)
	enc, _ := cipher.Encrypt("1234567890123")
	adv := &fakeAdviseeRepo{rows: []domain.Advisee{
		{StudentID: "s1", StudentCode: "S001", FirstName: "เด็กชาย", LastName: "ก", NationalIDEnc: enc, GradeLevel: "ม.1", RoomName: "1", TodayStatus: "present"},
	}}
	svc := NewMeService(&fakeMePersonnelRepo{byUser: map[string]*domain.Personnel{}}, adv, &fakeMeStudentRepo{}, cipher)

	out, err := svc.Advisees(meSemCtx("school-A", "u1", "sem-1"))
	if err != nil {
		t.Fatalf("advisees: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("len = %d, want 1", len(out))
	}
	if out[0].NationalIDMasked != crypto.Mask("1234567890123") {
		t.Errorf("ต้อง mask เลขบัตร: %q", out[0].NationalIDMasked)
	}
	if out[0].ClassLabel != "ม.1 1" {
		t.Errorf("class_label = %q", out[0].ClassLabel)
	}
	if adv.gotSchool != "school-A" || adv.gotUser != "u1" {
		t.Errorf("scope ผิด: school=%q user=%q", adv.gotSchool, adv.gotUser)
	}
}

// --- advisee edit tests ---

func studentFixture(schoolID, id string) *domain.Student {
	return &domain.Student{
		ID: id, SchoolID: schoolID, StudentCode: "S001", Status: domain.StudentStatusStudying,
		Profile: domain.PersonProfile{Prefix: "เด็กชาย", FirstName: "ก", LastName: "ข", Phone: "0800000000"},
	}
}

func meSvcWithAdvisee(t *testing.T, owns bool) (*MeService, *fakeMeStudentRepo) {
	t.Helper()
	adv := &fakeAdviseeRepo{owns: map[string]bool{}}
	if owns {
		adv.owns["school-A|sem-1|u1|s1"] = true
	}
	students := &fakeMeStudentRepo{byID: map[string]*domain.Student{"s1": studentFixture("school-A", "s1")}}
	return NewMeService(&fakeMePersonnelRepo{byUser: map[string]*domain.Personnel{}}, adv, students, testCipher(t)), students
}

func TestMeUpdateAdvisee_ForbiddenWhenNotOwner(t *testing.T) {
	svc, students := meSvcWithAdvisee(t, false)
	err := svc.UpdateAdvisee(meSemCtx("school-A", "u1", "sem-1"), "s1", UpdateAdviseeInput{FirstName: "ก", LastName: "ข"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
	if students.updated != nil {
		t.Error("ต้องไม่แก้ข้อมูลเมื่อไม่ใช่ที่ปรึกษา")
	}
}

func TestMeUpdateAdvisee_NoSemesterForbidden(t *testing.T) {
	svc, _ := meSvcWithAdvisee(t, true)
	if err := svc.UpdateAdvisee(meSemCtx("school-A", "u1", ""), "s1", UpdateAdviseeInput{FirstName: "ก", LastName: "ข"}); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden (ไม่มีเทอม)", err)
	}
}

func TestMeUpdateAdvisee_SuccessPreservesRegistrarFields(t *testing.T) {
	svc, students := meSvcWithAdvisee(t, true)
	err := svc.UpdateAdvisee(meSemCtx("school-A", "u1", "sem-1"), "s1", UpdateAdviseeInput{
		Prefix: "เด็กชาย", FirstName: "ก", LastName: "ใหม่", Phone: "0899999999",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if students.updated == nil {
		t.Fatal("ควรเรียก students.Update")
	}
	// รหัสนักเรียน/สถานะต้องคงเดิม (ครูที่ปรึกษาแก้ไม่ได้)
	if students.updated.StudentCode != "S001" || students.updated.Status != domain.StudentStatusStudying {
		t.Errorf("registrar fields เปลี่ยน: code=%q status=%q", students.updated.StudentCode, students.updated.Status)
	}
	if students.updated.ChangeNationalID {
		t.Error("ต้องไม่แตะเลขบัตรประชาชน")
	}
	if students.updated.Profile.LastName != "ใหม่" {
		t.Errorf("นามสกุลไม่อัปเดต: %q", students.updated.Profile.LastName)
	}
}

func TestMeAdviseeDetail_ForbiddenWhenNotOwner(t *testing.T) {
	svc, _ := meSvcWithAdvisee(t, false)
	if _, err := svc.AdviseeDetail(meSemCtx("school-A", "u1", "sem-1"), "s1"); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestMeAdviseeDetail_Success(t *testing.T) {
	svc, _ := meSvcWithAdvisee(t, true)
	d, err := svc.AdviseeDetail(meSemCtx("school-A", "u1", "sem-1"), "s1")
	if err != nil {
		t.Fatalf("detail: %v", err)
	}
	if d.StudentCode != "S001" || d.FirstName != "ก" {
		t.Errorf("detail ผิด: %+v", d)
	}
}
