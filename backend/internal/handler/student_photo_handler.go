package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/chumko-platform/backend/internal/httputil"
	"github.com/chumko-platform/backend/internal/service"
)

// StudentPhotoService contract ที่ handler ใช้ (รูปนักเรียนหลายรูป + เลือกรูปโปรไฟล์)
type StudentPhotoService interface {
	List(ctx context.Context, studentID string) ([]service.PhotoDTO, error)
	Upload(ctx context.Context, studentID string, in service.PhotoInput) (service.PhotoDTO, error)
	SetPrimary(ctx context.Context, studentID, photoID string) error
	Delete(ctx context.Context, studentID, photoID string) error
	Dataset(ctx context.Context, classID string) ([]service.FaceDatasetStudentDTO, error)
}

type StudentPhotoHandler struct {
	svc StudentPhotoService
}

func NewStudentPhotoHandler(svc StudentPhotoService) *StudentPhotoHandler {
	return &StudentPhotoHandler{svc: svc}
}

// List — GET /students/:id/photos
func (h *StudentPhotoHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.List(c.UserContext(), c.Params("id"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

// Upload — POST /students/:id/photos (multipart: file)
func (h *StudentPhotoHandler) Upload(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "FILE_REQUIRED", "กรุณาเลือกรูปที่จะอัปโหลด")
	}
	f, err := fileHeader.Open()
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "เปิดไฟล์ไม่สำเร็จ")
	}
	defer func() { _ = f.Close() }()

	dto, err := h.svc.Upload(c.UserContext(), c.Params("id"), service.PhotoInput{
		OriginalName: fileHeader.Filename,
		ContentType:  fileHeader.Header.Get("Content-Type"),
		Size:         fileHeader.Size,
		Reader:       f,
	})
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.Created(c, dto)
}

// SetPrimary — PUT /students/:id/photos/:photoId/primary
func (h *StudentPhotoHandler) SetPrimary(c *fiber.Ctx) error {
	if err := h.svc.SetPrimary(c.UserContext(), c.Params("id"), c.Params("photoId")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ตั้งเป็นรูปโปรไฟล์แล้ว"})
}

// Dataset — GET /face-dataset?class_id= (รูปนักเรียนเป็นชุดสำหรับระบบสแกนหน้า)
func (h *StudentPhotoHandler) Dataset(c *fiber.Ctx) error {
	items, err := h.svc.Dataset(c.UserContext(), c.Query("class_id"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

// Delete — DELETE /students/:id/photos/:photoId
func (h *StudentPhotoHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.UserContext(), c.Params("id"), c.Params("photoId")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ลบรูปแล้ว"})
}
