package handler

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/httputil"
)

// WorkGroupService contract ที่ handler ใช้
type WorkGroupService interface {
	ListGroups(ctx context.Context) ([]domain.WorkGroup, error)
	ListForPersonnel(ctx context.Context, personnelID string) ([]domain.WorkGroupMembership, error)
	Assign(ctx context.Context, personnelID, workGroupID string, isGroupAdmin bool) error
	Unassign(ctx context.Context, personnelID, workGroupID string) error
}

type WorkGroupHandler struct {
	svc      WorkGroupService
	validate *validator.Validate
}

func NewWorkGroupHandler(svc WorkGroupService) *WorkGroupHandler {
	return &WorkGroupHandler{svc: svc, validate: validator.New()}
}

type assignWorkGroupBody struct {
	WorkGroupID  string `json:"work_group_id" validate:"required,uuid"`
	IsGroupAdmin bool   `json:"is_group_admin"`
}

// ListGroups — GET /work-groups
func (h *WorkGroupHandler) ListGroups(c *fiber.Ctx) error {
	groups, err := h.svc.ListGroups(c.UserContext())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, groups)
}

// ListForPersonnel — GET /personnel/:id/work-groups
func (h *WorkGroupHandler) ListForPersonnel(c *fiber.Ctx) error {
	items, err := h.svc.ListForPersonnel(c.UserContext(), c.Params("id"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

// Assign — POST /personnel/:id/work-groups
func (h *WorkGroupHandler) Assign(c *fiber.Ctx) error {
	var body assignWorkGroupBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณาเลือกกลุ่มงานให้ถูกต้อง")
	}
	if err := h.svc.Assign(c.UserContext(), c.Params("id"), body.WorkGroupID, body.IsGroupAdmin); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "มอบหมายกลุ่มงานแล้ว"})
}

// Unassign — DELETE /personnel/:id/work-groups/:wgId
func (h *WorkGroupHandler) Unassign(c *fiber.Ctx) error {
	if err := h.svc.Unassign(c.UserContext(), c.Params("id"), c.Params("wgId")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ถอดออกจากกลุ่มงานแล้ว"})
}
