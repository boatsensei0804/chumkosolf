package service

import (
	"context"
	"errors"
	"testing"

	"github.com/chumko-platform/backend/internal/domain"
)

// --- fake sub-resource repos ---

type fakeAdminPosRepo struct {
	items             map[string]*domain.AdminPosition
	hasActiveDirector bool
	seq               int
}

func newFakeAdminPosRepo() *fakeAdminPosRepo {
	return &fakeAdminPosRepo{items: map[string]*domain.AdminPosition{}}
}

func (r *fakeAdminPosRepo) ListByPersonnel(_ context.Context, schoolID, personnelID string) ([]domain.AdminPosition, error) {
	var out []domain.AdminPosition
	for _, p := range r.items {
		if p.SchoolID == schoolID && p.PersonnelID == personnelID {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (r *fakeAdminPosRepo) Create(_ context.Context, schoolID, personnelID string, np domain.NewAdminPosition, _ domain.AuditEntry) (string, error) {
	if np.Position == domain.PositionDirector && np.IsActive && r.hasActiveDirector {
		return "", domain.ErrDuplicateDirector
	}
	r.seq++
	id := "pos" + string(rune('0'+r.seq))
	r.items[id] = &domain.AdminPosition{ID: id, SchoolID: schoolID, PersonnelID: personnelID, Position: np.Position, IsActive: np.IsActive}
	if np.Position == domain.PositionDirector && np.IsActive {
		r.hasActiveDirector = true
	}
	return id, nil
}

func (r *fakeAdminPosRepo) SoftDelete(_ context.Context, schoolID, personnelID, id string, _ domain.AuditEntry) (bool, error) {
	p, ok := r.items[id]
	if !ok || p.SchoolID != schoolID || p.PersonnelID != personnelID {
		return false, nil
	}
	delete(r.items, id)
	return true, nil
}

type fakeStandingRepo struct {
	items map[string]*domain.AcademicStanding
	seq   int
}

func newFakeStandingRepo() *fakeStandingRepo {
	return &fakeStandingRepo{items: map[string]*domain.AcademicStanding{}}
}

func (r *fakeStandingRepo) unsetCurrent(personnelID, exceptID string) {
	for _, s := range r.items {
		if s.PersonnelID == personnelID && s.ID != exceptID {
			s.IsCurrent = false
		}
	}
}

func (r *fakeStandingRepo) ListByPersonnel(_ context.Context, schoolID, personnelID string) ([]domain.AcademicStanding, error) {
	var out []domain.AcademicStanding
	for _, s := range r.items {
		if s.SchoolID == schoolID && s.PersonnelID == personnelID {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (r *fakeStandingRepo) Create(_ context.Context, schoolID, personnelID string, ns domain.NewAcademicStanding, _ domain.AuditEntry) (string, error) {
	if ns.IsCurrent {
		r.unsetCurrent(personnelID, "")
	}
	r.seq++
	id := "st" + string(rune('0'+r.seq))
	r.items[id] = &domain.AcademicStanding{ID: id, SchoolID: schoolID, PersonnelID: personnelID, Standing: ns.Standing, IsCurrent: ns.IsCurrent}
	return id, nil
}

func (r *fakeStandingRepo) Update(_ context.Context, schoolID, personnelID, id string, us domain.UpdateAcademicStanding, _ domain.AuditEntry) (bool, error) {
	s, ok := r.items[id]
	if !ok || s.SchoolID != schoolID || s.PersonnelID != personnelID {
		return false, nil
	}
	if us.IsCurrent {
		r.unsetCurrent(personnelID, id)
	}
	s.Standing = us.Standing
	s.IsCurrent = us.IsCurrent
	return true, nil
}

func (r *fakeStandingRepo) SoftDelete(_ context.Context, schoolID, personnelID, id string, _ domain.AuditEntry) (bool, error) {
	s, ok := r.items[id]
	if !ok || s.SchoolID != schoolID || s.PersonnelID != personnelID {
		return false, nil
	}
	delete(r.items, id)
	return true, nil
}

// --- helpers ---

const subSchool = "school-A"
const subPersonnel = "person-1"

func guardWith(schoolID, personnelID string) *fakePersonnelRepo {
	r := newFakePersonnelRepo()
	r.byID[personnelID] = &domain.Personnel{ID: personnelID, SchoolID: schoolID}
	return r
}

// --- admin position tests ---

func TestAdminPosition_CreateSuccess(t *testing.T) {
	svc := NewAdminPositionService(newFakeAdminPosRepo(), guardWith(subSchool, subPersonnel))
	id, err := svc.Create(adminCtx(subSchool), subPersonnel, CreateAdminPositionInput{Position: "deputy_director", IsActive: true})
	if err != nil || id == "" {
		t.Fatalf("create: id=%q err=%v", id, err)
	}
}

func TestAdminPosition_DuplicateDirector(t *testing.T) {
	svc := NewAdminPositionService(newFakeAdminPosRepo(), guardWith(subSchool, subPersonnel))
	ctx := adminCtx(subSchool)
	if _, err := svc.Create(ctx, subPersonnel, CreateAdminPositionInput{Position: "director", IsActive: true}); err != nil {
		t.Fatalf("first director: %v", err)
	}
	_, err := svc.Create(ctx, subPersonnel, CreateAdminPositionInput{Position: "director", IsActive: true})
	if !errors.Is(err, domain.ErrDuplicateDirector) {
		t.Errorf("err = %v, want ErrDuplicateDirector", err)
	}
}

func TestAdminPosition_InvalidPosition(t *testing.T) {
	svc := NewAdminPositionService(newFakeAdminPosRepo(), guardWith(subSchool, subPersonnel))
	_, err := svc.Create(adminCtx(subSchool), subPersonnel, CreateAdminPositionInput{Position: "teacher"})
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != "INVALID_POSITION" {
		t.Errorf("err = %v, want INVALID_POSITION", err)
	}
}

func TestAdminPosition_ForbiddenForNonMember(t *testing.T) {
	svc := NewAdminPositionService(newFakeAdminPosRepo(), guardWith(subSchool, subPersonnel))
	_, err := svc.Create(memberCtx(subSchool, "u9"), subPersonnel, CreateAdminPositionInput{Position: "director", IsActive: true})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestAdminPosition_PersonnelNotFoundCrossSchool(t *testing.T) {
	svc := NewAdminPositionService(newFakeAdminPosRepo(), guardWith(subSchool, subPersonnel))
	// personnel อยู่ school-A แต่เรียกด้วย scope school-B → ไม่พบ
	_, err := svc.Create(adminCtx("school-B"), subPersonnel, CreateAdminPositionInput{Position: "director", IsActive: true})
	if !errors.Is(err, domain.ErrPersonnelNotFound) {
		t.Errorf("err = %v, want ErrPersonnelNotFound", err)
	}
}

func TestAdminPosition_DeleteNotFound(t *testing.T) {
	svc := NewAdminPositionService(newFakeAdminPosRepo(), guardWith(subSchool, subPersonnel))
	if err := svc.Delete(adminCtx(subSchool), subPersonnel, "missing"); !errors.Is(err, domain.ErrAdminPositionNotFound) {
		t.Errorf("err = %v, want ErrAdminPositionNotFound", err)
	}
}

// --- academic standing tests ---

func TestStanding_CreateRequiresName(t *testing.T) {
	svc := NewAcademicStandingService(newFakeStandingRepo(), guardWith(subSchool, subPersonnel))
	_, err := svc.Create(adminCtx(subSchool), subPersonnel, StandingInput{Standing: "  "})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
}

func TestStanding_OnlyOneCurrent(t *testing.T) {
	repo := newFakeStandingRepo()
	svc := NewAcademicStandingService(repo, guardWith(subSchool, subPersonnel))
	ctx := adminCtx(subSchool)

	first, _ := svc.Create(ctx, subPersonnel, StandingInput{Standing: "ครูผู้ช่วย", IsCurrent: true})
	svc.Create(ctx, subPersonnel, StandingInput{Standing: "ครู คศ.1", IsCurrent: true}) //nolint:errcheck

	// อันแรกต้องถูก unset current เหลือ current เดียว
	if repo.items[first].IsCurrent {
		t.Error("วิทยฐานะเดิมต้องไม่ใช่ current หลังตั้งอันใหม่")
	}
	currentCount := 0
	for _, s := range repo.items {
		if s.IsCurrent {
			currentCount++
		}
	}
	if currentCount != 1 {
		t.Errorf("current count = %d, want 1", currentCount)
	}
}

func TestStanding_UpdateNotFound(t *testing.T) {
	svc := NewAcademicStandingService(newFakeStandingRepo(), guardWith(subSchool, subPersonnel))
	err := svc.Update(adminCtx(subSchool), subPersonnel, "missing", StandingInput{Standing: "ชำนาญการ"})
	if !errors.Is(err, domain.ErrStandingNotFound) {
		t.Errorf("err = %v, want ErrStandingNotFound", err)
	}
}

func TestStanding_DeleteNotFound(t *testing.T) {
	svc := NewAcademicStandingService(newFakeStandingRepo(), guardWith(subSchool, subPersonnel))
	if err := svc.Delete(adminCtx(subSchool), subPersonnel, "missing"); !errors.Is(err, domain.ErrStandingNotFound) {
		t.Errorf("err = %v, want ErrStandingNotFound", err)
	}
}
