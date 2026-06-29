package handler

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumko-platform/backend/internal/httputil"
	"github.com/chumko-platform/backend/internal/service"
)

// ================= Behavior records (คะแนนความประพฤติ) =================

// BehaviorService contract ที่ handler ใช้
type BehaviorService interface {
	Summary(ctx context.Context, studentID string) (*service.BehaviorSummaryDTO, error)
	Create(ctx context.Context, studentID string, in service.BehaviorInput) (string, error)
	Delete(ctx context.Context, studentID, id string) error
}

type BehaviorHandler struct {
	svc      BehaviorService
	validate *validator.Validate
}

func NewBehaviorHandler(svc BehaviorService) *BehaviorHandler {
	return &BehaviorHandler{svc: svc, validate: validator.New()}
}

type createBehaviorBody struct {
	Points     int    `json:"points" validate:"required,ne=0"`
	Reason     string `json:"reason" validate:"required,max=500"`
	OccurredAt string `json:"occurred_at" validate:"omitempty,datetime=2006-01-02"`
}

// Summary — GET /students/:id/behavior
func (h *BehaviorHandler) Summary(c *fiber.Ctx) error {
	res, err := h.svc.Summary(c.UserContext(), c.Params("id"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, res)
}

// Create — POST /students/:id/behavior
func (h *BehaviorHandler) Create(c *fiber.Ctx) error {
	var body createBehaviorBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณาระบุคะแนน (ไม่เป็นศูนย์) และเหตุผล")
	}
	occurredAt, err := parseBirthDate(body.OccurredAt)
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบวันที่ไม่ถูกต้อง")
	}
	id, err := h.svc.Create(c.UserContext(), c.Params("id"), service.BehaviorInput{
		Points:     body.Points,
		Reason:     body.Reason,
		OccurredAt: occurredAt,
	})
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.Created(c, fiber.Map{"id": id})
}

// Delete — DELETE /students/:id/behavior/:recordId
func (h *BehaviorHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.UserContext(), c.Params("id"), c.Params("recordId")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ลบรายการคะแนนแล้ว"})
}
