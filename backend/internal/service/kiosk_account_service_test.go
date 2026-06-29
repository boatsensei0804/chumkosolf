package service

import (
	"context"
	"errors"
	"testing"

	"github.com/chumko-platform/backend/internal/domain"
)

type fakeKioskRepo struct {
	created  bool
	list     []domain.UserBrief
	deleteOK bool
}

func (r *fakeKioskRepo) CreateKiosk(_ context.Context, _, _, _ string) (string, error) {
	r.created = true
	return "k1", nil
}
func (r *fakeKioskRepo) ListKiosk(_ context.Context, _ string) ([]domain.UserBrief, error) {
	return r.list, nil
}
func (r *fakeKioskRepo) DeleteKiosk(_ context.Context, _, _ string) (bool, error) {
	return r.deleteOK, nil
}

func TestKioskAccount_CreateForbiddenForNonAdmin(t *testing.T) {
	svc := NewKioskAccountService(&fakeKioskRepo{})
	if _, err := svc.Create(memberCtx("school-A", "u9"), CreateKioskInput{Username: "gate1", Password: "secret1"}); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestKioskAccount_CreateValidation(t *testing.T) {
	svc := NewKioskAccountService(&fakeKioskRepo{})
	if _, err := svc.Create(adminCtx("school-A"), CreateKioskInput{Username: "ab", Password: "secret1"}); err == nil {
		t.Error("ชื่อสั้นเกินต้อง error")
	}
	if _, err := svc.Create(adminCtx("school-A"), CreateKioskInput{Username: "gate1", Password: "123"}); err == nil {
		t.Error("รหัสสั้นเกินต้อง error")
	}
}

func TestKioskAccount_CreateSuccess(t *testing.T) {
	repo := &fakeKioskRepo{}
	svc := NewKioskAccountService(repo)
	if _, err := svc.Create(adminCtx("school-A"), CreateKioskInput{Username: "gate1", Password: "secret1"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if !repo.created {
		t.Error("ควรเรียก CreateKiosk")
	}
}

func TestKioskAccount_DeleteNotFound(t *testing.T) {
	svc := NewKioskAccountService(&fakeKioskRepo{deleteOK: false})
	if err := svc.Delete(adminCtx("school-A"), "missing"); !errors.Is(err, errKioskNotFound) {
		t.Errorf("err = %v, want errKioskNotFound", err)
	}
}
