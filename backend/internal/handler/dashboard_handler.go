package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/chumkosoft/backend/internal/httputil"
	"github.com/chumkosoft/backend/internal/service"
)

// DashboardService contract ที่ handler ใช้
type DashboardService interface {
	Summary(ctx context.Context) (*service.DashboardDTO, error)
}

type DashboardHandler struct {
	svc DashboardService
}

func NewDashboardHandler(svc DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// Summary — GET /dashboard
func (h *DashboardHandler) Summary(c *fiber.Ctx) error {
	res, err := h.svc.Summary(c.UserContext())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, res)
}
