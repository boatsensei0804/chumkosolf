package handler

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumko-platform/backend/internal/httputil"
	"github.com/chumko-platform/backend/internal/service"
)

// KioskAccountService contract ที่ handler ใช้ (จัดการบัญชีเครื่องสแกนหน้า — school admin)
type KioskAccountService interface {
	List(ctx context.Context) ([]service.KioskAccountDTO, error)
	Create(ctx context.Context, in service.CreateKioskInput) (string, error)
	Delete(ctx context.Context, userID string) error
}

type KioskAccountHandler struct {
	svc      KioskAccountService
	validate *validator.Validate
}

func NewKioskAccountHandler(svc KioskAccountService) *KioskAccountHandler {
	return &KioskAccountHandler{svc: svc, validate: validator.New()}
}

type createKioskBody struct {
	Username string `json:"username" validate:"required,min=3,max=100"`
	Password string `json:"password" validate:"required,min=6"`
}

func (h *KioskAccountHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.List(c.UserContext())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

func (h *KioskAccountHandler) Create(c *fiber.Ctx) error {
	var body createKioskBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกชื่อผู้ใช้ (≥3) และรหัสผ่าน (≥6)")
	}
	id, err := h.svc.Create(c.UserContext(), service.CreateKioskInput{Username: body.Username, Password: body.Password})
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.Created(c, fiber.Map{"id": id})
}

func (h *KioskAccountHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.UserContext(), c.Params("id")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ลบบัญชีแล้ว"})
}
