package handler

import (
	"context"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumkosoft/backend/internal/httputil"
	"github.com/chumkosoft/backend/internal/service"
)

type GuardianService interface {
	List(ctx context.Context, page, pageSize int) ([]service.GuardianListItem, int, error)
	Get(ctx context.Context, id string) (*service.GuardianDetail, error)
	Create(ctx context.Context, in service.CreateGuardianInput) (string, error)
	Update(ctx context.Context, id string, in service.UpdateGuardianInput) error
	Delete(ctx context.Context, id string) error
}

type GuardianHandler struct {
	svc      GuardianService
	validate *validator.Validate
}

func NewGuardianHandler(svc GuardianService) *GuardianHandler {
	return &GuardianHandler{svc: svc, validate: validator.New()}
}

type createGuardianBody struct {
	NationalID string      `json:"national_id" validate:"required,len=13,numeric"`
	Prefix     string      `json:"prefix" validate:"omitempty,max=50"`
	FirstName  string      `json:"first_name" validate:"required,max=150"`
	LastName   string      `json:"last_name" validate:"required,max=150"`
	BirthDate  string      `json:"birth_date" validate:"omitempty,datetime=2006-01-02"`
	Phone      string      `json:"phone" validate:"omitempty,max=20"`
	Address    addressBody `json:"address"`
}

type updateGuardianBody struct {
	NationalID string      `json:"national_id" validate:"omitempty,len=13,numeric"`
	Prefix     string      `json:"prefix" validate:"omitempty,max=50"`
	FirstName  string      `json:"first_name" validate:"required,max=150"`
	LastName   string      `json:"last_name" validate:"required,max=150"`
	BirthDate  string      `json:"birth_date" validate:"omitempty,datetime=2006-01-02"`
	Phone      string      `json:"phone" validate:"omitempty,max=20"`
	Address    addressBody `json:"address"`
}

func (h *GuardianHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	items, total, err := h.svc.List(c.UserContext(), page, pageSize)
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OKWithMeta(c, items, &httputil.Meta{Page: page, Total: total})
}

func (h *GuardianHandler) Get(c *fiber.Ctx) error {
	detail, err := h.svc.Get(c.UserContext(), c.Params("id"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, detail)
}

func (h *GuardianHandler) Create(c *fiber.Ctx) error {
	var body createGuardianBody
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
	id, err := h.svc.Create(c.UserContext(), service.CreateGuardianInput{
		NationalID: body.NationalID, Prefix: body.Prefix, FirstName: body.FirstName, LastName: body.LastName,
		BirthDate: birth, Phone: body.Phone, Address: body.Address.toDTO(),
	})
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.Created(c, fiber.Map{"id": id})
}

func (h *GuardianHandler) Update(c *fiber.Ctx) error {
	var body updateGuardianBody
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
	if err := h.svc.Update(c.UserContext(), c.Params("id"), service.CreateGuardianInput{
		NationalID: body.NationalID, Prefix: body.Prefix, FirstName: body.FirstName, LastName: body.LastName,
		BirthDate: birth, Phone: body.Phone, Address: body.Address.toDTO(),
	}); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "บันทึกข้อมูลผู้ปกครองแล้ว"})
}

func (h *GuardianHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.UserContext(), c.Params("id")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ลบข้อมูลผู้ปกครองแล้ว"})
}
