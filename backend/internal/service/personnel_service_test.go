package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/chumko-platform/backend/internal/crypto"
	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/tenant"
)

const testEncKey = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

// --- fake repo ---

type fakePersonnelRepo struct {
	byID   map[string]*domain.Personnel
	hashes map[string]bool // key: schoolID + "|" + hash
	groups map[string]bool // key: schoolID + "|" + userID + "|" + code
	audits []domain.AuditEntry
	seq    int
}

func newFakePersonnelRepo() *fakePersonnelRepo {
	return &fakePersonnelRepo{
		byID:   map[string]*domain.Personnel{},
		hashes: map[string]bool{},
		groups: map[string]bool{},
	}
}

func (r *fakePersonnelRepo) Create(_ context.Context, schoolID string, np domain.NewPersonnel, audit domain.AuditEntry) (string, error) {
	hk := schoolID + "|" + np.NationalIDHash
	if r.hashes[hk] {
		return "", domain.ErrDuplicateNationalID
	}
	r.seq++
	id := "p" + string(rune('0'+r.seq))
	r.byID[id] = &domain.Personnel{
		ID: id, SchoolID: schoolID, UserID: "u" + id,
		NationalIDEnc: np.NationalIDEnc, NationalIDHash: np.NationalIDHash,
		CivilServantIDEnc: np.CivilServantIDEnc,
		Profile:           np.Profile, Username: np.Username, Role: np.Role, IsActive: true,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	r.hashes[hk] = true
	audit.TargetID = id
	r.audits = append(r.audits, audit)
	return id, nil
}

func (r *fakePersonnelRepo) List(_ context.Context, schoolID string, limit, offset int) ([]domain.Personnel, int, error) {
	var all []domain.Personnel
	for _, p := range r.byID {
		if p.SchoolID == schoolID {
			all = append(all, *p)
		}
	}
	return all, len(all), nil
}

func (r *fakePersonnelRepo) GetByID(_ context.Context, schoolID, id string) (*domain.Personnel, error) {
	p, ok := r.byID[id]
	if !ok || p.SchoolID != schoolID {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (r *fakePersonnelRepo) Update(_ context.Context, schoolID, id string, up domain.UpdatePersonnel, audit domain.AuditEntry) (bool, error) {
	p, ok := r.byID[id]
	if !ok || p.SchoolID != schoolID {
		return false, nil
	}
	p.Profile = up.Profile
	if up.ChangeNationalID {
		p.NationalIDEnc = up.NationalIDEnc
		p.NationalIDHash = up.NationalIDHash
	}
	r.audits = append(r.audits, audit)
	return true, nil
}

func (r *fakePersonnelRepo) SoftDelete(_ context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error) {
	p, ok := r.byID[id]
	if !ok || p.SchoolID != schoolID {
		return false, nil
	}
	delete(r.byID, id)
	r.audits = append(r.audits, audit)
	return true, nil
}

func (r *fakePersonnelRepo) InsertAudit(_ context.Context, audit domain.AuditEntry) error {
	r.audits = append(r.audits, audit)
	return nil
}

func (r *fakePersonnelRepo) IsUserInWorkGroup(_ context.Context, schoolID, userID, code string) (bool, error) {
	return r.groups[schoolID+"|"+userID+"|"+code], nil
}

// --- helpers ---

func newPersonnelService(t *testing.T) (*PersonnelService, *fakePersonnelRepo) {
	t.Helper()
	cipher, err := crypto.NewCipher(testEncKey)
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	repo := newFakePersonnelRepo()
	return NewPersonnelService(repo, cipher), repo
}

func adminCtx(schoolID string) context.Context {
	return tenant.WithIdentity(context.Background(), tenant.Identity{
		UserID: "admin", SchoolID: schoolID, Role: "super_admin", IsSchoolAdmin: true,
	})
}

func memberCtx(schoolID, userID string) context.Context {
	return tenant.WithIdentity(context.Background(), tenant.Identity{
		UserID: userID, SchoolID: schoolID, Role: "teacher", IsSchoolAdmin: false,
	})
}

func validCreateInput() CreatePersonnelInput {
	return CreatePersonnelInput{
		Username: "kru.somchai", Password: "password123", Role: "teacher",
		NationalID: "1234567890123", FirstName: "สมชาย", LastName: "ใจดี",
	}
}

// --- tests ---

func TestPersonnelCreate_SuccessAndEncryption(t *testing.T) {
	svc, repo := newPersonnelService(t)

	id, err := svc.Create(adminCtx("school-A"), validCreateInput())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	stored := repo.byID[id]
	if stored == nil {
		t.Fatal("personnel ไม่ถูกเก็บ")
	}
	// เลขบัตรต้องถูกเข้ารหัส — ciphertext ห้ามมี plaintext
	if strings.Contains(string(stored.NationalIDEnc), "1234567890123") {
		t.Error("เลขบัตรประชาชนต้องไม่เก็บเป็น plaintext")
	}
	// ต้องมี audit create
	if len(repo.audits) == 0 || repo.audits[len(repo.audits)-1].Action != domain.AuditCreate {
		t.Error("ต้องบันทึก audit create")
	}
}

func TestPersonnelCreate_ForbiddenForNonMember(t *testing.T) {
	svc, _ := newPersonnelService(t)

	_, err := svc.Create(memberCtx("school-A", "teacher-x"), validCreateInput())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestPersonnelCreate_AllowedForPersonnelGroupMember(t *testing.T) {
	svc, repo := newPersonnelService(t)
	repo.groups["school-A|teacher-x|personnel"] = true

	if _, err := svc.Create(memberCtx("school-A", "teacher-x"), validCreateInput()); err != nil {
		t.Errorf("สมาชิกกลุ่มบุคคลควรสร้างได้: %v", err)
	}
}

func TestPersonnelCreate_InvalidRole(t *testing.T) {
	svc, _ := newPersonnelService(t)
	in := validCreateInput()
	in.Role = "student"

	_, err := svc.Create(adminCtx("school-A"), in)
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != "INVALID_ROLE" {
		t.Errorf("err = %v, want INVALID_ROLE", err)
	}
}

func TestPersonnelCreate_InvalidNationalID(t *testing.T) {
	svc, _ := newPersonnelService(t)
	in := validCreateInput()
	in.NationalID = "123"

	_, err := svc.Create(adminCtx("school-A"), in)
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != "INVALID_NATIONAL_ID" {
		t.Errorf("err = %v, want INVALID_NATIONAL_ID", err)
	}
}

func TestPersonnelCreate_DuplicateNationalID(t *testing.T) {
	svc, _ := newPersonnelService(t)
	ctx := adminCtx("school-A")
	if _, err := svc.Create(ctx, validCreateInput()); err != nil {
		t.Fatalf("first create: %v", err)
	}

	in := validCreateInput()
	in.Username = "kru.other"
	_, err := svc.Create(ctx, in)
	if !errors.Is(err, domain.ErrDuplicateNationalID) {
		t.Errorf("err = %v, want ErrDuplicateNationalID", err)
	}
}

func TestPersonnelGet_MasksNationalID(t *testing.T) {
	svc, _ := newPersonnelService(t)
	ctx := adminCtx("school-A")
	id, _ := svc.Create(ctx, validCreateInput())

	detail, err := svc.Get(ctx, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if detail.NationalIDMasked != crypto.Mask("1234567890123") {
		t.Errorf("national_id_masked = %q, want masked", detail.NationalIDMasked)
	}
	if strings.Contains(detail.NationalIDMasked, "234567890") {
		t.Error("ต้องไม่เปิดเผยเลขบัตรเต็มใน response")
	}
}

func TestPersonnelGet_NotFound(t *testing.T) {
	svc, _ := newPersonnelService(t)

	_, err := svc.Get(adminCtx("school-A"), "missing")
	if !errors.Is(err, domain.ErrPersonnelNotFound) {
		t.Errorf("err = %v, want ErrPersonnelNotFound", err)
	}
}

// TestPersonnel_SchoolIsolation: ข้อมูลของโรงเรียน A เข้าถึงด้วย scope โรงเรียน B ไม่ได้
func TestPersonnel_SchoolIsolation(t *testing.T) {
	svc, _ := newPersonnelService(t)
	id, _ := svc.Create(adminCtx("school-A"), validCreateInput())

	// อ่านด้วย scope โรงเรียน B → ต้องไม่พบ
	if _, err := svc.Get(adminCtx("school-B"), id); !errors.Is(err, domain.ErrPersonnelNotFound) {
		t.Errorf("get cross-school err = %v, want ErrPersonnelNotFound", err)
	}
	// ลบด้วย scope โรงเรียน B → ต้องไม่พบ (ลบไม่ได้)
	if err := svc.Delete(adminCtx("school-B"), id); !errors.Is(err, domain.ErrPersonnelNotFound) {
		t.Errorf("delete cross-school err = %v, want ErrPersonnelNotFound", err)
	}
}

func TestPersonnelUpdate_NotFound(t *testing.T) {
	svc, _ := newPersonnelService(t)
	in := UpdatePersonnelInput{FirstName: "ก", LastName: "ข"}

	if err := svc.Update(adminCtx("school-A"), "missing", in); !errors.Is(err, domain.ErrPersonnelNotFound) {
		t.Errorf("err = %v, want ErrPersonnelNotFound", err)
	}
}

func TestPersonnelDelete_Success(t *testing.T) {
	svc, repo := newPersonnelService(t)
	ctx := adminCtx("school-A")
	id, _ := svc.Create(ctx, validCreateInput())

	if err := svc.Delete(ctx, id); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, ok := repo.byID[id]; ok {
		t.Error("personnel ควรถูกลบแล้ว")
	}
}
