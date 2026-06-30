package service

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/tenant"
)

// KioskAccountRepo — จัดการบัญชี role kiosk
type KioskAccountRepo interface {
	CreateKiosk(ctx context.Context, schoolID, username, passwordHash string) (string, error)
	ListKiosk(ctx context.Context, schoolID string) ([]domain.UserBrief, error)
	DeleteKiosk(ctx context.Context, schoolID, userID string) (bool, error)
}

// KioskAccountService — บัญชีเครื่องสแกนหน้า; จัดการได้เฉพาะ school admin
type KioskAccountService struct {
	repo KioskAccountRepo
}

func NewKioskAccountService(repo KioskAccountRepo) *KioskAccountService {
	return &KioskAccountService{repo: repo}
}

type KioskAccountDTO struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

type CreateKioskInput struct {
	Username string
	Password string
}

var errKioskNotFound = &domain.Error{Status: 404, Code: "KIOSK_NOT_FOUND", Message: "ไม่พบบัญชีสแกนหน้านี้"}

func (s *KioskAccountService) requireAdmin(ctx context.Context) error {
	if tenant.IsSchoolAdminFromContext(ctx) {
		return nil
	}
	return domain.ErrForbidden
}

func (s *KioskAccountService) List(ctx context.Context) ([]KioskAccountDTO, error) {
	if err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListKiosk(ctx, tenant.SchoolIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	out := make([]KioskAccountDTO, 0, len(rows))
	for i := range rows {
		out = append(out, KioskAccountDTO{ID: rows[i].ID, Username: rows[i].Username, IsActive: rows[i].IsActive, CreatedAt: rows[i].CreatedAt})
	}
	return out, nil
}

func (s *KioskAccountService) Create(ctx context.Context, in CreateKioskInput) (string, error) {
	if err := s.requireAdmin(ctx); err != nil {
		return "", err
	}
	username := strings.TrimSpace(in.Username)
	if len(username) < 3 {
		return "", &domain.Error{Status: 400, Code: "VALIDATION_ERROR", Message: "ชื่อผู้ใช้อย่างน้อย 3 ตัวอักษร"}
	}
	if len(in.Password) < 6 {
		return "", &domain.Error{Status: 400, Code: "VALIDATION_ERROR", Message: "รหัสผ่านอย่างน้อย 6 ตัวอักษร"}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return s.repo.CreateKiosk(ctx, tenant.SchoolIDFromContext(ctx), username, string(hash))
}

func (s *KioskAccountService) Delete(ctx context.Context, userID string) error {
	if err := s.requireAdmin(ctx); err != nil {
		return err
	}
	found, err := s.repo.DeleteKiosk(ctx, tenant.SchoolIDFromContext(ctx), userID)
	if err != nil {
		return err
	}
	if !found {
		return errKioskNotFound
	}
	return nil
}
