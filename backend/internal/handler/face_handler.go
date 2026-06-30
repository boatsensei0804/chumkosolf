package handler

import (
	"context"
	"io"

	"github.com/gofiber/fiber/v2"

	"github.com/chumkosoft/backend/internal/httputil"
	"github.com/chumkosoft/backend/internal/service"
)

// FaceService contract ที่ handler ใช้ (สแกนหน้าเข้าเรียน)
type FaceService interface {
	Reindex(ctx context.Context) (service.ReindexResult, error)
	RecognizeAndMark(ctx context.Context, frames [][]byte) (service.RecognizeResult, error)
}

type FaceHandler struct {
	svc FaceService
}

func NewFaceHandler(svc FaceService) *FaceHandler {
	return &FaceHandler{svc: svc}
}

// Reindex — POST /face/reindex (สร้างฐานใบหน้าใหม่จากรูปนักเรียนทั้งหมด)
func (h *FaceHandler) Reindex(c *fiber.Ctx) error {
	res, err := h.svc.Reindex(c.UserContext())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, res)
}

// Recognize — POST /face/recognize (multipart: file = หลายเฟรมจากกล้อง สำหรับตรวจ liveness) แล้วบันทึกเช็คชื่อ
func (h *FaceHandler) Recognize(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	headers := form.File["file"]
	if len(headers) == 0 {
		return httputil.Error(c, fiber.StatusBadRequest, "FILE_REQUIRED", "ไม่พบภาพจากกล้อง")
	}

	frames := make([][]byte, 0, len(headers))
	for _, fh := range headers {
		f, err := fh.Open()
		if err != nil {
			return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "เปิดภาพไม่สำเร็จ")
		}
		buf, err := io.ReadAll(f)
		_ = f.Close()
		if err != nil {
			return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "อ่านภาพไม่สำเร็จ")
		}
		frames = append(frames, buf)
	}

	res, err := h.svc.RecognizeAndMark(c.UserContext(), frames)
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, res)
}
