package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/chumko-platform/backend/internal/crypto"
	"github.com/chumko-platform/backend/internal/domain"
)

// --- fakes ---

type fakeWGChecker struct{ groups map[string]bool } // schoolID|userID|code

func (f *fakeWGChecker) IsUserInWorkGroup(_ context.Context, schoolID, userID, code string) (bool, error) {
	return f.groups[schoolID+"|"+userID+"|"+code], nil
}

type fakeStudentRepo struct {
	byID   map[string]*domain.Student
	hashes map[string]bool // schoolID|hash
	codes  map[string]bool // schoolID|code
	seq    int
}

func newFakeStudentRepo() *fakeStudentRepo {
	return &fakeStudentRepo{byID: map[string]*domain.Student{}, hashes: map[string]bool{}, codes: map[string]bool{}}
}

func (r *fakeStudentRepo) List(_ context.Context, schoolID string, _, _ int, _ string) ([]domain.Student, int, error) {
	var out []domain.Student
	for _, s := range r.byID {
		if s.SchoolID == schoolID {
			out = append(out, *s)
		}
	}
	return out, len(out), nil
}
func (r *fakeStudentRepo) GetByID(_ context.Context, schoolID, id string) (*domain.Student, error) {
	s, ok := r.byID[id]
	if !ok || s.SchoolID != schoolID {
		return nil, nil
	}
	cp := *s
	return &cp, nil
}
func (r *fakeStudentRepo) CurrentClass(_ context.Context, _, _, _ string) (string, string, string, error) {
	return "", "", "", nil
}
func (r *fakeStudentRepo) Create(_ context.Context, schoolID string, ns domain.NewStudent, _ domain.AuditEntry) (string, error) {
	if r.hashes[schoolID+"|"+ns.NationalIDHash] {
		return "", domain.ErrDuplicateNationalID
	}
	if r.codes[schoolID+"|"+ns.StudentCode] {
		return "", domain.ErrDuplicateStudentCode
	}
	r.seq++
	id := "st" + string(rune('0'+r.seq))
	r.byID[id] = &domain.Student{ID: id, SchoolID: schoolID, NationalIDEnc: ns.NationalIDEnc, NationalIDHash: ns.NationalIDHash, StudentCode: ns.StudentCode, Status: ns.Status, Profile: ns.Profile}
	r.hashes[schoolID+"|"+ns.NationalIDHash] = true
	r.codes[schoolID+"|"+ns.StudentCode] = true
	return id, nil
}
func (r *fakeStudentRepo) Update(_ context.Context, schoolID, id string, _ domain.UpdateStudent, _ domain.AuditEntry) (bool, error) {
	s, ok := r.byID[id]
	return ok && s.SchoolID == schoolID, nil
}
func (r *fakeStudentRepo) SoftDelete(_ context.Context, schoolID, id string, _ domain.AuditEntry) (bool, error) {
	s, ok := r.byID[id]
	if !ok || s.SchoolID != schoolID {
		return false, nil
	}
	delete(r.byID, id)
	return true, nil
}

type fakeGuardianRepo struct {
	byID map[string]*domain.Guardian
	seq  int
}

func newFakeGuardianRepo() *fakeGuardianRepo { return &fakeGuardianRepo{byID: map[string]*domain.Guardian{}} }

func (r *fakeGuardianRepo) List(_ context.Context, schoolID string, _, _ int) ([]domain.Guardian, int, error) {
	var out []domain.Guardian
	for _, g := range r.byID {
		if g.SchoolID == schoolID {
			out = append(out, *g)
		}
	}
	return out, len(out), nil
}
func (r *fakeGuardianRepo) GetByID(_ context.Context, schoolID, id string) (*domain.Guardian, error) {
	g, ok := r.byID[id]
	if !ok || g.SchoolID != schoolID {
		return nil, nil
	}
	cp := *g
	return &cp, nil
}
func (r *fakeGuardianRepo) Create(_ context.Context, schoolID string, ng domain.NewGuardian, _ domain.AuditEntry) (string, error) {
	r.seq++
	id := "g" + string(rune('0'+r.seq))
	r.byID[id] = &domain.Guardian{ID: id, SchoolID: schoolID, NationalIDEnc: ng.NationalIDEnc, NationalIDHash: ng.NationalIDHash, Profile: ng.Profile}
	return id, nil
}
func (r *fakeGuardianRepo) Upsert(_ context.Context, schoolID string, ng domain.NewGuardian) (string, error) {
	for id, g := range r.byID { // dedup by national id hash (พี่น้องใช้ผู้ปกครองเดิม)
		if g.SchoolID == schoolID && g.NationalIDHash == ng.NationalIDHash {
			return id, nil
		}
	}
	r.seq++
	id := "g" + string(rune('0'+r.seq))
	r.byID[id] = &domain.Guardian{ID: id, SchoolID: schoolID, NationalIDEnc: ng.NationalIDEnc, NationalIDHash: ng.NationalIDHash, Profile: ng.Profile}
	return id, nil
}
func (r *fakeGuardianRepo) Update(_ context.Context, schoolID, id string, _ domain.UpdateGuardian, _ domain.AuditEntry) (bool, error) {
	g, ok := r.byID[id]
	return ok && g.SchoolID == schoolID, nil
}
func (r *fakeGuardianRepo) SoftDelete(_ context.Context, schoolID, id string, _ domain.AuditEntry) (bool, error) {
	g, ok := r.byID[id]
	if !ok || g.SchoolID != schoolID {
		return false, nil
	}
	delete(r.byID, id)
	return true, nil
}

type fakeSGRepo struct {
	links map[string]*domain.StudentGuardian // linkID -> link
	stOf  map[string]string                  // linkID -> studentID
	seq   int
}

func newFakeSGRepo() *fakeSGRepo {
	return &fakeSGRepo{links: map[string]*domain.StudentGuardian{}, stOf: map[string]string{}}
}
func (r *fakeSGRepo) ListByStudent(_ context.Context, _, studentID string) ([]domain.StudentGuardian, error) {
	var out []domain.StudentGuardian
	for id, l := range r.links {
		if r.stOf[id] == studentID {
			out = append(out, *l)
		}
	}
	return out, nil
}
func (r *fakeSGRepo) Link(_ context.Context, _, studentID string, nsg domain.NewStudentGuardian, _ domain.AuditEntry) error {
	if nsg.IsPrimary {
		for id, l := range r.links {
			if r.stOf[id] == studentID && l.GuardianID != nsg.GuardianID {
				l.IsPrimary = false
			}
		}
	}
	r.seq++
	id := "l" + string(rune('0'+r.seq))
	r.links[id] = &domain.StudentGuardian{ID: id, GuardianID: nsg.GuardianID, Relationship: nsg.Relationship, IsPrimary: nsg.IsPrimary}
	r.stOf[id] = studentID
	return nil
}
func (r *fakeSGRepo) Unlink(_ context.Context, _, studentID, linkID string, _ domain.AuditEntry) (bool, error) {
	if r.stOf[linkID] != studentID {
		return false, nil
	}
	delete(r.links, linkID)
	delete(r.stOf, linkID)
	return true, nil
}

// --- helpers ---

func testCipher(t *testing.T) Cipher {
	t.Helper()
	c, err := crypto.NewCipher(testEncKey)
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	return c
}

func academicChecker(schoolID, userID string) *fakeWGChecker {
	return &fakeWGChecker{groups: map[string]bool{schoolID + "|" + userID + "|academic": true}}
}

func validStudentInput() CreateStudentInput {
	return CreateStudentInput{NationalID: "1234567890123", StudentCode: "S001", FirstName: "เด็กชาย", LastName: "ใจดี"}
}

// --- student tests ---

func TestStudentCreate_SuccessEncryption(t *testing.T) {
	repo := newFakeStudentRepo()
	svc := NewStudentService(repo, &fakeWGChecker{}, testCipher(t))
	id, err := svc.Create(adminCtx("school-A"), validStudentInput())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if strings.Contains(string(repo.byID[id].NationalIDEnc), "1234567890123") {
		t.Error("เลขบัตรต้องไม่เก็บเป็น plaintext")
	}
	// ไม่ระบุ status → default = กำลังศึกษา
	if repo.byID[id].Status != domain.StudentStatusStudying {
		t.Errorf("status = %q ควร default เป็น studying", repo.byID[id].Status)
	}
}

func TestStudentCreate_InvalidStatus(t *testing.T) {
	svc := NewStudentService(newFakeStudentRepo(), &fakeWGChecker{}, testCipher(t))
	in := validStudentInput()
	in.Status = "graduated"
	_, err := svc.Create(adminCtx("school-A"), in)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation (status ไม่ถูกต้อง)", err)
	}
}

func TestStudentCreate_Forbidden(t *testing.T) {
	svc := NewStudentService(newFakeStudentRepo(), &fakeWGChecker{}, testCipher(t))
	_, err := svc.Create(memberCtx("school-A", "u9"), validStudentInput())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestStudentCreate_AllowedAcademicMember(t *testing.T) {
	svc := NewStudentService(newFakeStudentRepo(), academicChecker("school-A", "u9"), testCipher(t))
	if _, err := svc.Create(memberCtx("school-A", "u9"), validStudentInput()); err != nil {
		t.Errorf("สมาชิกกลุ่มวิชาการควรสร้างได้: %v", err)
	}
}

func TestStudentCreate_InvalidNationalID(t *testing.T) {
	svc := NewStudentService(newFakeStudentRepo(), &fakeWGChecker{}, testCipher(t))
	in := validStudentInput()
	in.NationalID = "123"
	_, err := svc.Create(adminCtx("school-A"), in)
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != "INVALID_NATIONAL_ID" {
		t.Errorf("err = %v, want INVALID_NATIONAL_ID", err)
	}
}

func TestStudentCreate_DuplicateCode(t *testing.T) {
	repo := newFakeStudentRepo()
	svc := NewStudentService(repo, &fakeWGChecker{}, testCipher(t))
	ctx := adminCtx("school-A")
	if _, err := svc.Create(ctx, validStudentInput()); err != nil {
		t.Fatalf("first: %v", err)
	}
	in := validStudentInput()
	in.NationalID = "9999999999999" // เลขบัตรต่าง แต่รหัสนักเรียนซ้ำ
	_, err := svc.Create(ctx, in)
	if !errors.Is(err, domain.ErrDuplicateStudentCode) {
		t.Errorf("err = %v, want ErrDuplicateStudentCode", err)
	}
}

func TestStudentGet_MaskAndIsolation(t *testing.T) {
	repo := newFakeStudentRepo()
	svc := NewStudentService(repo, &fakeWGChecker{}, testCipher(t))
	id, _ := svc.Create(adminCtx("school-A"), validStudentInput())

	d, err := svc.Get(adminCtx("school-A"), id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if d.NationalIDMasked != crypto.Mask("1234567890123") {
		t.Errorf("national_id ต้อง mask: %q", d.NationalIDMasked)
	}
	// isolation: โรงเรียน B เข้าไม่ได้
	if _, err := svc.Get(adminCtx("school-B"), id); !errors.Is(err, domain.ErrStudentNotFound) {
		t.Errorf("cross-school get err = %v, want ErrStudentNotFound", err)
	}
}

// --- guardian + link tests ---

func TestGuardianCreate_Mask(t *testing.T) {
	repo := newFakeGuardianRepo()
	svc := NewGuardianService(repo, &fakeWGChecker{}, testCipher(t))
	id, err := svc.Create(adminCtx("school-A"), CreateGuardianInput{NationalID: "1234567890123", FirstName: "นาง", LastName: "ผู้ปกครอง"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	d, _ := svc.Get(adminCtx("school-A"), id)
	if d.NationalIDMasked != crypto.Mask("1234567890123") {
		t.Errorf("mask: %q", d.NationalIDMasked)
	}
}

func newLinkSvc(t *testing.T) (*StudentGuardianService, *fakeStudentRepo, *fakeGuardianRepo, *fakeSGRepo) {
	t.Helper()
	sr := newFakeStudentRepo()
	gr := newFakeGuardianRepo()
	lr := newFakeSGRepo()
	svc := NewStudentGuardianService(lr, sr, gr, &fakeWGChecker{}, testCipher(t))
	return svc, sr, gr, lr
}

func father(nid, first string) LinkGuardianInput {
	return LinkGuardianInput{NationalID: nid, FirstName: first, LastName: "ผู้ปกครอง", Relationship: "father"}
}

func TestLink_InvalidRelationship(t *testing.T) {
	svc, _, _, _ := newLinkSvc(t)
	in := father("1111111111111", "พ่อ")
	in.Relationship = "uncle"
	if err := svc.Link(adminCtx("school-A"), "st1", in); !errors.Is(err, domain.ErrInvalidRelationship) {
		t.Errorf("err = %v, want ErrInvalidRelationship", err)
	}
}

func TestLink_InvalidNationalID(t *testing.T) {
	svc, _, _, _ := newLinkSvc(t)
	in := father("123", "พ่อ")
	_, _ = svc, in
	if err := svc.Link(adminCtx("school-A"), "st1", in); err == nil {
		t.Error("เลขบัตรผู้ปกครองไม่ถูกต้องควร error")
	}
}

func TestLink_StudentMustExist(t *testing.T) {
	svc, _, _, _ := newLinkSvc(t)
	if err := svc.Link(adminCtx("school-A"), "missing", father("1111111111111", "พ่อ")); !errors.Is(err, domain.ErrStudentNotFound) {
		t.Errorf("err = %v, want ErrStudentNotFound", err)
	}
}

// สร้างผู้ปกครอง inline แล้วเชื่อม — primary ได้คนเดียว
func TestLink_InlineCreateAndPrimarySingle(t *testing.T) {
	svc, sr, _, lr := newLinkSvc(t)
	ctx := adminCtx("school-A")
	sid, _ := NewStudentService(sr, &fakeWGChecker{}, testCipher(t)).Create(ctx, validStudentInput())

	in1 := father("1111111111111", "พ่อ")
	in1.IsPrimary = true
	in2 := LinkGuardianInput{NationalID: "2222222222222", FirstName: "แม่", LastName: "ผู้ปกครอง", Relationship: "mother", IsPrimary: true}
	if err := svc.Link(ctx, sid, in1); err != nil {
		t.Fatalf("link1: %v", err)
	}
	if err := svc.Link(ctx, sid, in2); err != nil {
		t.Fatalf("link2: %v", err)
	}
	primary := 0
	for _, l := range lr.links {
		if l.IsPrimary {
			primary++
		}
	}
	if primary != 1 {
		t.Errorf("primary count = %d, want 1", primary)
	}
}

func TestUnlink_NotFound(t *testing.T) {
	svc, sr, _, _ := newLinkSvc(t)
	ctx := adminCtx("school-A")
	sid, _ := NewStudentService(sr, &fakeWGChecker{}, testCipher(t)).Create(ctx, validStudentInput())
	if err := svc.Unlink(ctx, sid, "missing"); !errors.Is(err, domain.ErrGuardianLinkNotFound) {
		t.Errorf("err = %v, want ErrGuardianLinkNotFound", err)
	}
}
