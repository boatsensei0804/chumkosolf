package handler

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumko-platform/backend/internal/httputil"
	"github.com/chumko-platform/backend/internal/service"
)

// MeService คือ contract ที่ handler ใช้ (self-service ของผู้ใช้เอง)
type MeService interface {
	Profile(ctx context.Context) (*service.PersonnelDetail, error)
	UpdateProfile(ctx context.Context, in service.UpdateMyProfileInput) error
	Advisees(ctx context.Context) ([]service.AdviseeDTO, error)
	AdviseeDetail(ctx context.Context, studentID string) (*service.StudentDetail, error)
	UpdateAdvisee(ctx context.Context, studentID string, in service.UpdateAdviseeInput) error
}

// MeHandler จัดการ endpoint กลุ่ม /me (ข้อมูลของผู้ใช้เอง)
type MeHandler struct {
	svc      MeService
	validate *validator.Validate
}

func NewMeHandler(svc MeService) *MeHandler {
	return &MeHandler{svc: svc, validate: validator.New()}
}

type updateMyProfileBody struct {
	Prefix    string      `json:"prefix" validate:"omitempty,max=50"`
	FirstName string      `json:"first_name" validate:"required,max=150"`
	LastName  string      `json:"last_name" validate:"required,max=150"`
	BirthDate string      `json:"birth_date" validate:"omitempty,datetime=2006-01-02"`
	Phone     string      `json:"phone" validate:"omitempty,max=20"`
	Email     string      `json:"email" validate:"omitempty,email"`
	Address   addressBody `json:"address"`
}

// Profile — GET /me/profile
func (h *MeHandler) Profile(c *fiber.Ctx) error {
	detail, err := h.svc.Profile(c.UserContext())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, detail)
}

// UpdateProfile — PUT /me/profile
func (h *MeHandler) UpdateProfile(c *fiber.Ctx) error {
	var body updateMyProfileBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกข้อมูลให้ครบและถูกต้อง")
	}

	birth, err := parseBirthDate(body.BirthDate)
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบวันเกิดไม่ถูกต้อง (YYYY-MM-DD)")
	}

	if err := h.svc.UpdateProfile(c.UserContext(), service.UpdateMyProfileInput{
		Prefix:    body.Prefix,
		FirstName: body.FirstName,
		LastName:  body.LastName,
		BirthDate: birth,
		Phone:     body.Phone,
		Email:     body.Email,
		Address:   body.Address.toDTO(),
	}); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "บันทึกข้อมูลของคุณแล้ว"})
}

// Advisees — GET /me/advisees
func (h *MeHandler) Advisees(c *fiber.Ctx) error {
	items, err := h.svc.Advisees(c.UserContext())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

type updateAdviseeBody struct {
	Prefix    string      `json:"prefix" validate:"omitempty,max=50"`
	FirstName string      `json:"first_name" validate:"required,max=150"`
	LastName  string      `json:"last_name" validate:"required,max=150"`
	BirthDate string      `json:"birth_date" validate:"omitempty,datetime=2006-01-02"`
	Phone     string      `json:"phone" validate:"omitempty,max=20"`
	Address   addressBody `json:"address"`
}

// AdviseeDetail — GET /me/advisees/:studentId
func (h *MeHandler) AdviseeDetail(c *fiber.Ctx) error {
	detail, err := h.svc.AdviseeDetail(c.UserContext(), c.Params("studentId"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, detail)
}

// UpdateAdvisee — PUT /me/advisees/:studentId
func (h *MeHandler) UpdateAdvisee(c *fiber.Ctx) error {
	var body updateAdviseeBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกข้อมูลให้ครบและถูกต้อง")
	}

	birth, err := parseBirthDate(body.BirthDate)
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบวันเกิดไม่ถูกต้อง (YYYY-MM-DD)")
	}

	if err := h.svc.UpdateAdvisee(c.UserContext(), c.Params("studentId"), service.UpdateAdviseeInput{
		Prefix:    body.Prefix,
		FirstName: body.FirstName,
		LastName:  body.LastName,
		BirthDate: birth,
		Phone:     body.Phone,
		Address:   body.Address.toDTO(),
	}); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "บันทึกข้อมูลนักเรียนแล้ว"})
}
