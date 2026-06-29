package handler

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumko-platform/backend/internal/httputil"
	"github.com/chumko-platform/backend/internal/service"
)

// ================= Classes =================

type ClassService interface {
	List(ctx context.Context) ([]service.ClassListItem, error)
	Get(ctx context.Context, id string) (*service.ClassDetail, error)
	Create(ctx context.Context, in service.ClassInput) (string, error)
	Update(ctx context.Context, id string, in service.ClassInput) error
	Delete(ctx context.Context, id string) error
}

type ClassHandler struct {
	svc      ClassService
	validate *validator.Validate
}

func NewClassHandler(svc ClassService) *ClassHandler {
	return &ClassHandler{svc: svc, validate: validator.New()}
}

type classBody struct {
	GradeLevel string `json:"grade_level" validate:"required,max=50"`
	RoomName   string `json:"room_name" validate:"required,max=50"`
}

func (h *ClassHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.List(c.UserContext())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

func (h *ClassHandler) Get(c *fiber.Ctx) error {
	detail, err := h.svc.Get(c.UserContext(), c.Params("id"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, detail)
}

func (h *ClassHandler) Create(c *fiber.Ctx) error {
	var body classBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกระดับชั้นและห้อง")
	}
	id, err := h.svc.Create(c.UserContext(), service.ClassInput{GradeLevel: body.GradeLevel, RoomName: body.RoomName})
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.Created(c, fiber.Map{"id": id})
}

func (h *ClassHandler) Update(c *fiber.Ctx) error {
	var body classBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกระดับชั้นและห้อง")
	}
	if err := h.svc.Update(c.UserContext(), c.Params("id"), service.ClassInput{GradeLevel: body.GradeLevel, RoomName: body.RoomName}); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "บันทึกห้องเรียนแล้ว"})
}

func (h *ClassHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.UserContext(), c.Params("id")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ลบห้องเรียนแล้ว"})
}

// ================= Class advisors =================

type ClassAdvisorService interface {
	List(ctx context.Context, classID string) ([]service.ClassAdvisorDTO, error)
	Add(ctx context.Context, classID, personnelID string) error
	Remove(ctx context.Context, classID, advisorID string) error
}

type ClassAdvisorHandler struct {
	svc      ClassAdvisorService
	validate *validator.Validate
}

func NewClassAdvisorHandler(svc ClassAdvisorService) *ClassAdvisorHandler {
	return &ClassAdvisorHandler{svc: svc, validate: validator.New()}
}

type addAdvisorBody struct {
	PersonnelID string `json:"personnel_id" validate:"required,uuid"`
}

func (h *ClassAdvisorHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.List(c.UserContext(), c.Params("id"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

func (h *ClassAdvisorHandler) Add(c *fiber.Ctx) error {
	var body addAdvisorBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณาเลือกครูที่ปรึกษา")
	}
	if err := h.svc.Add(c.UserContext(), c.Params("id"), body.PersonnelID); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "เพิ่มครูที่ปรึกษาแล้ว"})
}

func (h *ClassAdvisorHandler) Remove(c *fiber.Ctx) error {
	if err := h.svc.Remove(c.UserContext(), c.Params("id"), c.Params("advisorId")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ถอดครูที่ปรึกษาแล้ว"})
}

// ================= Enrollments =================

type EnrollmentService interface {
	List(ctx context.Context, classID string) ([]service.EnrollmentDTO, error)
	Enroll(ctx context.Context, classID string, in service.EnrollInput) error
	EnrollMany(ctx context.Context, classID string, studentIDs []string) (int, error)
	Remove(ctx context.Context, classID, enrollmentID string) error
}

type EnrollmentHandler struct {
	svc      EnrollmentService
	validate *validator.Validate
}

func NewEnrollmentHandler(svc EnrollmentService) *EnrollmentHandler {
	return &EnrollmentHandler{svc: svc, validate: validator.New()}
}

type enrollBody struct {
	StudentID string `json:"student_id" validate:"required,uuid"`
	StudentNo *int   `json:"student_no" validate:"omitempty,min=1"`
}

type enrollManyBody struct {
	StudentIDs []string `json:"student_ids" validate:"required,min=1,max=200,dive,uuid"`
}

func (h *EnrollmentHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.List(c.UserContext(), c.Params("id"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

func (h *EnrollmentHandler) Enroll(c *fiber.Ctx) error {
	var body enrollBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณาเลือกนักเรียน")
	}
	if err := h.svc.Enroll(c.UserContext(), c.Params("id"), service.EnrollInput{StudentID: body.StudentID, StudentNo: body.StudentNo}); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "จัดนักเรียนเข้าห้องแล้ว"})
}

// EnrollMany — POST /classes/:id/students/bulk (จัดนักเรียนหลายคนพร้อมกัน)
func (h *EnrollmentHandler) EnrollMany(c *fiber.Ctx) error {
	var body enrollManyBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณาเลือกนักเรียนอย่างน้อย 1 คน")
	}
	count, err := h.svc.EnrollMany(c.UserContext(), c.Params("id"), body.StudentIDs)
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"enrolled": count, "message": "จัดนักเรียนเข้าห้องแล้ว"})
}

func (h *EnrollmentHandler) Remove(c *fiber.Ctx) error {
	if err := h.svc.Remove(c.UserContext(), c.Params("id"), c.Params("enrollmentId")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ถอนนักเรียนออกจากห้องแล้ว"})
}
