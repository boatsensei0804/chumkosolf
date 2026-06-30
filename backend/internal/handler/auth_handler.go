// Package handler มี HTTP handler (Fiber) — แปลง request/response เท่านั้น ไม่มี business logic
package handler

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/httputil"
	"github.com/chumkosoft/backend/internal/service"
	"github.com/chumkosoft/backend/internal/tenant"
)

// AuthService คือ contract ของ service ที่ handler ใช้ (interface เพื่อ test ง่าย)
type AuthService interface {
	Login(ctx context.Context, schoolCode, username, password string) (*service.LoginResult, error)
	Refresh(ctx context.Context, refreshToken string) (*service.LoginResult, error)
	Logout(ctx context.Context, refreshToken string) error
	Me(ctx context.Context, schoolID, userID string) (*service.UserInfo, error)
}

// AuthHandler จัดการ endpoint กลุ่ม /auth
type AuthHandler struct {
	svc      AuthService
	validate *validator.Validate
}

// NewAuthHandler สร้าง handler
func NewAuthHandler(svc AuthService) *AuthHandler {
	return &AuthHandler{svc: svc, validate: validator.New()}
}

type loginRequest struct {
	SchoolCode string `json:"school_code" validate:"required"`
	Username   string `json:"username" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Login godoc — POST /api/v1/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(req); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกโรงเรียน ชื่อผู้ใช้ และรหัสผ่านให้ครบ")
	}

	result, err := h.svc.Login(c.UserContext(), req.SchoolCode, req.Username, req.Password)
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, result)
}

// Refresh godoc — POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req refreshRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(req); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "ต้องระบุ refresh token")
	}

	result, err := h.svc.Refresh(c.UserContext(), req.RefreshToken)
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, result)
}

// Logout godoc — POST /api/v1/auth/logout (เพิกถอน refresh token)
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req refreshRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.svc.Logout(c.UserContext(), req.RefreshToken); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ออกจากระบบเรียบร้อย"})
}

// Me godoc — GET /api/v1/auth/me (ต้องผ่าน RequireAuth)
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	ctx := c.UserContext()
	schoolID := tenant.SchoolIDFromContext(ctx)
	userID := tenant.UserIDFromContext(ctx)

	info, err := h.svc.Me(ctx, schoolID, userID)
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, info)
}

// respondServiceError แปลง business error → response มาตรฐาน (domain.Error พก status/code/ข้อความไทย)
func respondServiceError(c *fiber.Ctx, err error) error {
	var de *domain.Error
	if errors.As(err, &de) {
		return httputil.Error(c, de.Status, de.Code, de.Message)
	}
	return httputil.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "เกิดข้อผิดพลาดภายในระบบ")
}
