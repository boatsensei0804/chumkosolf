package handler

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumko-platform/backend/internal/httputil"
	"github.com/chumko-platform/backend/internal/service"
)

// ================= Academic years / semesters (จัดการเทอม) =================

type AcademicService interface {
	ListYears(ctx context.Context) ([]service.AcademicYearDTO, error)
	CreateYear(ctx context.Context, year int) (string, error)
	SetCurrentYear(ctx context.Context, id string) error
	ListSemesters(ctx context.Context) ([]service.SemesterDTO, error)
	CreateSemester(ctx context.Context, in service.NewSemesterInput) (string, error)
	SetCurrentSemester(ctx context.Context, id string) error
}

type AcademicHandler struct {
	svc      AcademicService
	validate *validator.Validate
}

func NewAcademicHandler(svc AcademicService) *AcademicHandler {
	return &AcademicHandler{svc: svc, validate: validator.New()}
}

type createYearBody struct {
	Year int `json:"year" validate:"required,min=2400,max=2700"`
}

type createSemesterBody struct {
	AcademicYearID string `json:"academic_year_id" validate:"required,uuid"`
	Term           int    `json:"term" validate:"required,oneof=1 2"`
	StartDate      string `json:"start_date" validate:"omitempty,datetime=2006-01-02"`
	EndDate        string `json:"end_date" validate:"omitempty,datetime=2006-01-02"`
}

// --- ปีการศึกษา ---

func (h *AcademicHandler) ListYears(c *fiber.Ctx) error {
	items, err := h.svc.ListYears(c.UserContext())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

func (h *AcademicHandler) CreateYear(c *fiber.Ctx) error {
	var body createYearBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกปีการศึกษา (พ.ศ.)")
	}
	id, err := h.svc.CreateYear(c.UserContext(), body.Year)
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.Created(c, fiber.Map{"id": id})
}

func (h *AcademicHandler) SetCurrentYear(c *fiber.Ctx) error {
	if err := h.svc.SetCurrentYear(c.UserContext(), c.Params("id")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ตั้งปีการศึกษาปัจจุบันแล้ว"})
}

// --- ภาคเรียน ---

func (h *AcademicHandler) ListSemesters(c *fiber.Ctx) error {
	items, err := h.svc.ListSemesters(c.UserContext())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

func (h *AcademicHandler) CreateSemester(c *fiber.Ctx) error {
	var body createSemesterBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณาเลือกปีการศึกษาและภาคเรียน (1/2)")
	}
	start, err := parseBirthDate(body.StartDate)
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบวันที่เริ่มไม่ถูกต้อง")
	}
	end, err := parseBirthDate(body.EndDate)
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบวันที่จบไม่ถูกต้อง")
	}
	id, err := h.svc.CreateSemester(c.UserContext(), service.NewSemesterInput{
		AcademicYearID: body.AcademicYearID, Term: body.Term, StartDate: start, EndDate: end,
	})
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.Created(c, fiber.Map{"id": id})
}

func (h *AcademicHandler) SetCurrentSemester(c *fiber.Ctx) error {
	if err := h.svc.SetCurrentSemester(c.UserContext(), c.Params("id")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ตั้งภาคเรียนปัจจุบันแล้ว"})
}
