package handler

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumkosoft/backend/internal/httputil"
	"github.com/chumkosoft/backend/internal/service"
)

// ================= Personnel works (ผลงานครู) =================

// PersonnelWorkService contract ที่ handler ใช้
type PersonnelWorkService interface {
	List(ctx context.Context, personnelID string) ([]service.PersonnelWorkDTO, error)
	Create(ctx context.Context, personnelID string, in service.WorkInput) (string, error)
	Update(ctx context.Context, personnelID, workID string, in service.WorkInput) error
	Delete(ctx context.Context, personnelID, workID string) error

	ListFiles(ctx context.Context, personnelID, workID string) ([]service.WorkFileDTO, error)
	UploadFile(ctx context.Context, personnelID, workID string, in service.UploadFileInput) (string, error)
	DeleteFile(ctx context.Context, personnelID, workID, fileID string) error
}

type PersonnelWorkHandler struct {
	svc      PersonnelWorkService
	validate *validator.Validate
}

func NewPersonnelWorkHandler(svc PersonnelWorkService) *PersonnelWorkHandler {
	return &PersonnelWorkHandler{svc: svc, validate: validator.New()}
}

type workBody struct {
	Title       string `json:"title" validate:"required,max=255"`
	Description string `json:"description" validate:"max=5000"`
	WorkDate    string `json:"work_date" validate:"omitempty,datetime=2006-01-02"`
}

func (b workBody) toInput() (service.WorkInput, error) {
	d, err := parseBirthDate(b.WorkDate)
	if err != nil {
		return service.WorkInput{}, err
	}
	return service.WorkInput{Title: b.Title, Description: b.Description, WorkDate: d}, nil
}

// List — GET /personnel/:id/works
func (h *PersonnelWorkHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.List(c.UserContext(), c.Params("id"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

// Create — POST /personnel/:id/works
func (h *PersonnelWorkHandler) Create(c *fiber.Ctx) error {
	var body workBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกชื่อผลงาน")
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

// Update — PUT /personnel/:id/works/:workId
func (h *PersonnelWorkHandler) Update(c *fiber.Ctx) error {
	var body workBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกชื่อผลงาน")
	}
	in, err := body.toInput()
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบวันที่ไม่ถูกต้อง")
	}
	if err := h.svc.Update(c.UserContext(), c.Params("id"), c.Params("workId"), in); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "บันทึกผลงานแล้ว"})
}

// Delete — DELETE /personnel/:id/works/:workId
func (h *PersonnelWorkHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.UserContext(), c.Params("id"), c.Params("workId")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ลบผลงานแล้ว"})
}

// ListFiles — GET /personnel/:id/works/:workId/files
func (h *PersonnelWorkHandler) ListFiles(c *fiber.Ctx) error {
	items, err := h.svc.ListFiles(c.UserContext(), c.Params("id"), c.Params("workId"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

// UploadFile — POST /personnel/:id/works/:workId/files (multipart: file, file_type)
func (h *PersonnelWorkHandler) UploadFile(c *fiber.Ctx) error {
	fileType := c.FormValue("file_type")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "FILE_REQUIRED", "กรุณาเลือกไฟล์ที่จะอัปโหลด")
	}
	f, err := fileHeader.Open()
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "เปิดไฟล์ไม่สำเร็จ")
	}
	defer func() { _ = f.Close() }()

	id, err := h.svc.UploadFile(c.UserContext(), c.Params("id"), c.Params("workId"), service.UploadFileInput{
		FileType:     fileType,
		OriginalName: fileHeader.Filename,
		ContentType:  fileHeader.Header.Get("Content-Type"),
		Size:         fileHeader.Size,
		Reader:       f,
	})
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.Created(c, fiber.Map{"id": id})
}

// DeleteFile — DELETE /personnel/:id/works/:workId/files/:fileId
func (h *PersonnelWorkHandler) DeleteFile(c *fiber.Ctx) error {
	if err := h.svc.DeleteFile(c.UserContext(), c.Params("id"), c.Params("workId"), c.Params("fileId")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ลบไฟล์แนบแล้ว"})
}
