package service

import (
	"context"
	"errors"
	"testing"

	"github.com/chumko-platform/backend/internal/domain"
)

// --- fake work group repo ---

type fakeWGRepo struct {
	groups     []domain.WorkGroup
	exists     map[string]bool // schoolID|wgID
	groupAdmin map[string]bool // schoolID|userID|wgID
	assigned   map[string]bool // schoolID|userID|wgID
}

func newFakeWGRepo() *fakeWGRepo {
	return &fakeWGRepo{
		exists:     map[string]bool{},
		groupAdmin: map[string]bool{},
		assigned:   map[string]bool{},
	}
}

func (r *fakeWGRepo) ListBySchool(_ context.Context, _ string) ([]domain.WorkGroup, error) {
	return r.groups, nil
}
func (r *fakeWGRepo) ExistsInSchool(_ context.Context, schoolID, wgID string) (bool, error) {
	return r.exists[schoolID+"|"+wgID], nil
}
func (r *fakeWGRepo) ListForUser(_ context.Context, _, _ string) ([]domain.WorkGroupMembership, error) {
	return nil, nil
}
func (r *fakeWGRepo) IsGroupAdmin(_ context.Context, schoolID, userID, wgID string) (bool, error) {
	return r.groupAdmin[schoolID+"|"+userID+"|"+wgID], nil
}
func (r *fakeWGRepo) Assign(_ context.Context, schoolID, userID, wgID string, _ bool, _ domain.AuditEntry) error {
	r.assigned[schoolID+"|"+userID+"|"+wgID] = true
	return nil
}
func (r *fakeWGRepo) Unassign(_ context.Context, schoolID, userID, wgID string, _ domain.AuditEntry) (bool, error) {
	key := schoolID + "|" + userID + "|" + wgID
	if !r.assigned[key] {
		return false, nil
	}
	delete(r.assigned, key)
	return true, nil
}

const wgID = "wg-1"

func setupWG(t *testing.T) (*WorkGroupService, *fakeWGRepo) {
	t.Helper()
	repo := newFakeWGRepo()
	repo.exists[subSchool+"|"+wgID] = true
	guard := guardWith(subSchool, subPersonnel) // personnel ใน subSchool, UserID = "" (พอสำหรับ fake)
	return NewWorkGroupService(repo, guard), repo
}

func TestWG_AssignBySchoolAdmin(t *testing.T) {
	svc, repo := setupWG(t)
	if err := svc.Assign(adminCtx(subSchool), subPersonnel, wgID, true); err != nil {
		t.Fatalf("assign: %v", err)
	}
	if !repo.assigned[subSchool+"||"+wgID] {
		t.Error("ควรถูกมอบหมายเข้ากลุ่มงาน")
	}
}

func TestWG_AssignForbiddenForNonAdminNonGroupAdmin(t *testing.T) {
	svc, _ := setupWG(t)
	err := svc.Assign(memberCtx(subSchool, "u9"), subPersonnel, wgID, false)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestWG_AssignAllowedForGroupAdmin(t *testing.T) {
	svc, repo := setupWG(t)
	repo.groupAdmin[subSchool+"|u9|"+wgID] = true // u9 เป็นหัวหน้ากลุ่มนี้
	if err := svc.Assign(memberCtx(subSchool, "u9"), subPersonnel, wgID, false); err != nil {
		t.Errorf("หัวหน้ากลุ่มควรมอบหมายเข้ากลุ่มตัวเองได้: %v", err)
	}
}

func TestWG_AssignWorkGroupNotFound(t *testing.T) {
	svc, _ := setupWG(t)
	if err := svc.Assign(adminCtx(subSchool), subPersonnel, "wg-unknown", false); !errors.Is(err, domain.ErrWorkGroupNotFound) {
		t.Errorf("err = %v, want ErrWorkGroupNotFound", err)
	}
}

func TestWG_AssignPersonnelNotFoundCrossSchool(t *testing.T) {
	svc, repo := setupWG(t)
	repo.exists["school-B|"+wgID] = true
	// personnel อยู่ subSchool แต่ scope school-B → ไม่พบ
	if err := svc.Assign(adminCtx("school-B"), subPersonnel, wgID, false); !errors.Is(err, domain.ErrPersonnelNotFound) {
		t.Errorf("err = %v, want ErrPersonnelNotFound", err)
	}
}

func TestWG_UnassignNotFound(t *testing.T) {
	svc, _ := setupWG(t)
	if err := svc.Unassign(adminCtx(subSchool), subPersonnel, wgID); !errors.Is(err, domain.ErrWorkGroupAssignmentNotFound) {
		t.Errorf("err = %v, want ErrWorkGroupAssignmentNotFound", err)
	}
}

func TestWG_AssignThenUnassign(t *testing.T) {
	svc, _ := setupWG(t)
	ctx := adminCtx(subSchool)
	if err := svc.Assign(ctx, subPersonnel, wgID, false); err != nil {
		t.Fatalf("assign: %v", err)
	}
	if err := svc.Unassign(ctx, subPersonnel, wgID); err != nil {
		t.Errorf("unassign: %v", err)
	}
}

func TestWG_ListGroups(t *testing.T) {
	svc, repo := setupWG(t)
	repo.groups = []domain.WorkGroup{{ID: "wg-1", Code: "personnel", Name: "กลุ่มงานบุคคล"}}
	got, err := svc.ListGroups(adminCtx(subSchool))
	if err != nil || len(got) != 1 {
		t.Fatalf("list groups: got %d err %v", len(got), err)
	}
}
