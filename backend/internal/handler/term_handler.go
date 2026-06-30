package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/chumkosoft/backend/internal/httputil"
	"github.com/chumkosoft/backend/internal/service"
)

// TermService contract ที่ handler ใช้
type TermService interface {
	Current(ctx context.Context) (*service.CurrentTermDTO, error)
}

type TermHandler struct {
	svc TermService
}

func NewTermHandler(svc TermService) *TermHandler {
	return &TermHandler{svc: svc}
}

// Current — GET /current-term
func (h *TermHandler) Current(c *fiber.Ctx) error {
	res, err := h.svc.Current(c.UserContext())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, res)
}
