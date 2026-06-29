package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/chumko-platform/backend/internal/httputil"
	"github.com/chumko-platform/backend/internal/service"
)

// DirectoryService contract ที่ handler ใช้ (ดู/ค้นหาห้องเรียน-นักเรียน read-only)
type DirectoryService interface {
	Classes(ctx context.Context) ([]service.DirectoryClassDTO, error)
	ClassStudents(ctx context.Context, classID string) ([]service.DirectoryStudentDTO, error)
	SearchStudents(ctx context.Context, term string) ([]service.DirectoryStudentClassDTO, error)
}

type DirectoryHandler struct {
	svc DirectoryService
}

func NewDirectoryHandler(svc DirectoryService) *DirectoryHandler {
	return &DirectoryHandler{svc: svc}
}

// Classes — GET /directory/classes
func (h *DirectoryHandler) Classes(c *fiber.Ctx) error {
	items, err := h.svc.Classes(c.UserContext())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

// ClassStudents — GET /directory/classes/:classId/students
func (h *DirectoryHandler) ClassStudents(c *fiber.Ctx) error {
	items, err := h.svc.ClassStudents(c.UserContext(), c.Params("classId"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

// SearchStudents — GET /directory/students?q=...
func (h *DirectoryHandler) SearchStudents(c *fiber.Ctx) error {
	items, err := h.svc.SearchStudents(c.UserContext(), c.Query("q"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}
