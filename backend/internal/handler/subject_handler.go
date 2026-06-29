package handler

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumko-platform/backend/internal/httputil"
	"github.com/chumko-platform/backend/internal/service"
)

// ================= Subjects (รายวิชา) =================

type SubjectService interface {
	List(ctx context.Context) ([]service.SubjectDTO, error)
	Create(ctx context.Context, in service.SubjectInput) (string, error)
	Update(ctx context.Context, id string, in service.SubjectInput) error
	Delete(ctx context.Context, id string) error
}

type SubjectHandler struct {
	svc      SubjectService
	validate *validator.Validate
}

func NewSubjectHandler(svc SubjectService) *SubjectHandler {
	return &SubjectHandler{svc: svc, validate: validator.New()}
}

type subjectBody struct {
	SubjectCode string   `json:"subject_code" validate:"required,max=50"`
	Name        string   `json:"name" validate:"required,max=255"`
	Credit      *float64 `json:"credit" validate:"omitempty,min=0,max=99"`
}

func (b subjectBody) toInput() service.SubjectInput {
	return service.SubjectInput{SubjectCode: b.SubjectCode, Name: b.Name, Credit: b.Credit}
}

func (h *SubjectHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.List(c.UserContext())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

func (h *SubjectHandler) Create(c *fiber.Ctx) error {
	var body subjectBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกรหัสวิชาและชื่อวิชา")
	}
	id, err := h.svc.Create(c.UserContext(), body.toInput())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.Created(c, fiber.Map{"id": id})
}

func (h *SubjectHandler) Update(c *fiber.Ctx) error {
	var body subjectBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกรหัสวิชาและชื่อวิชา")
	}
	if err := h.svc.Update(c.UserContext(), c.Params("id"), body.toInput()); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "บันทึกรายวิชาแล้ว"})
}

func (h *SubjectHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.UserContext(), c.Params("id")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ลบรายวิชาแล้ว"})
}

// ================= Teaching assignments (มอบหมายการสอน) =================

type TeachingAssignmentService interface {
	List(ctx context.Context) ([]service.TeachingAssignmentDTO, error)
	Create(ctx context.Context, in service.TeachingAssignmentInput) (string, error)
	Delete(ctx context.Context, id string) error
}

type TeachingAssignmentHandler struct {
	svc      TeachingAssignmentService
	validate *validator.Validate
}

func NewTeachingAssignmentHandler(svc TeachingAssignmentService) *TeachingAssignmentHandler {
	return &TeachingAssignmentHandler{svc: svc, validate: validator.New()}
}

type teachingAssignmentBody struct {
	PersonnelID string `json:"personnel_id" validate:"required,uuid"`
	SubjectID   string `json:"subject_id" validate:"required,uuid"`
	ClassID     string `json:"class_id" validate:"required,uuid"`
}

func (h *TeachingAssignmentHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.List(c.UserContext())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

func (h *TeachingAssignmentHandler) Create(c *fiber.Ctx) error {
	var body teachingAssignmentBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณาเลือกครู วิชา และห้อง")
	}
	id, err := h.svc.Create(c.UserContext(), service.TeachingAssignmentInput{
		PersonnelID: body.PersonnelID, SubjectID: body.SubjectID, ClassID: body.ClassID,
	})
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.Created(c, fiber.Map{"id": id})
}

func (h *TeachingAssignmentHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.UserContext(), c.Params("id")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ลบการมอบหมายแล้ว"})
}
