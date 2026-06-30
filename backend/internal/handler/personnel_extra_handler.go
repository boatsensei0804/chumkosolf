package handler

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumkosoft/backend/internal/httputil"
	"github.com/chumkosoft/backend/internal/service"
)

// ================= Admin positions =================

// AdminPositionService contract ที่ handler ใช้
type AdminPositionService interface {
	List(ctx context.Context, personnelID string) ([]service.AdminPositionDTO, error)
	Create(ctx context.Context, personnelID string, in service.CreateAdminPositionInput) (string, error)
	Delete(ctx context.Context, personnelID, id string) error
}

type AdminPositionHandler struct {
	svc      AdminPositionService
	validate *validator.Validate
}

func NewAdminPositionHandler(svc AdminPositionService) *AdminPositionHandler {
	return &AdminPositionHandler{svc: svc, validate: validator.New()}
}

type createPositionBody struct {
	Position    string `json:"position" validate:"required,oneof=director deputy_director"`
	IsActive    *bool  `json:"is_active"`
	AppointedAt string `json:"appointed_at" validate:"omitempty,datetime=2006-01-02"`
}

// List — GET /personnel/:id/positions
func (h *AdminPositionHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.List(c.UserContext(), c.Params("id"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

// Create — POST /personnel/:id/positions
func (h *AdminPositionHandler) Create(c *fiber.Ctx) error {
	var body createPositionBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณาเลือกตำแหน่งให้ถูกต้อง")
	}
	appointedAt, err := parseBirthDate(body.AppointedAt)
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบวันที่แต่งตั้งไม่ถูกต้อง")
	}
	// is_active default = true (ตำแหน่งที่เพิ่งแต่งตั้งถือว่ากำลังดำรงตำแหน่ง)
	isActive := true
	if body.IsActive != nil {
		isActive = *body.IsActive
	}

	id, err := h.svc.Create(c.UserContext(), c.Params("id"), service.CreateAdminPositionInput{
		Position:    body.Position,
		IsActive:    isActive,
		AppointedAt: appointedAt,
	})
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.Created(c, fiber.Map{"id": id})
}

// Delete — DELETE /personnel/:id/positions/:posId
func (h *AdminPositionHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.UserContext(), c.Params("id"), c.Params("posId")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ลบตำแหน่งบริหารแล้ว"})
}

// ================= Academic standings =================

// AcademicStandingService contract ที่ handler ใช้
type AcademicStandingService interface {
	List(ctx context.Context, personnelID string) ([]service.AcademicStandingDTO, error)
	Create(ctx context.Context, personnelID string, in service.StandingInput) (string, error)
	Update(ctx context.Context, personnelID, id string, in service.StandingInput) error
	Delete(ctx context.Context, personnelID, id string) error
}

type AcademicStandingHandler struct {
	svc      AcademicStandingService
	validate *validator.Validate
}

func NewAcademicStandingHandler(svc AcademicStandingService) *AcademicStandingHandler {
	return &AcademicStandingHandler{svc: svc, validate: validator.New()}
}

type standingBody struct {
	Standing      string `json:"standing" validate:"required,max=100"`
	EffectiveDate string `json:"effective_date" validate:"omitempty,datetime=2006-01-02"`
	IsCurrent     bool   `json:"is_current"`
}

func (b standingBody) toInput() (service.StandingInput, error) {
	eff, err := parseBirthDate(b.EffectiveDate)
	if err != nil {
		return service.StandingInput{}, err
	}
	return service.StandingInput{Standing: b.Standing, EffectiveDate: eff, IsCurrent: b.IsCurrent}, nil
}

// List — GET /personnel/:id/standings
func (h *AcademicStandingHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.List(c.UserContext(), c.Params("id"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

// Create — POST /personnel/:id/standings
func (h *AcademicStandingHandler) Create(c *fiber.Ctx) error {
	var body standingBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกชื่อวิทยฐานะ")
	}
	in, err := body.toInput()
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบวันที่ไม่ถูกต้อง")
	}

	id, err := h.svc.Create(c.UserContext(), c.Params("id"), in)
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.Created(c, fiber.Map{"id": id})
}

// Update — PUT /personnel/:id/standings/:sid
func (h *AcademicStandingHandler) Update(c *fiber.Ctx) error {
	var body standingBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกชื่อวิทยฐานะ")
	}
	in, err := body.toInput()
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบวันที่ไม่ถูกต้อง")
	}

	if err := h.svc.Update(c.UserContext(), c.Params("id"), c.Params("sid"), in); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "บันทึกวิทยฐานะแล้ว"})
}

// Delete — DELETE /personnel/:id/standings/:sid
func (h *AcademicStandingHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.UserContext(), c.Params("id"), c.Params("sid")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ลบวิทยฐานะแล้ว"})
}
