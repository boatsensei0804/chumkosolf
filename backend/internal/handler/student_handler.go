package handler

import (
	"context"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumko-platform/backend/internal/httputil"
	"github.com/chumko-platform/backend/internal/service"
)

// ================= Students =================

type StudentService interface {
	List(ctx context.Context, page, pageSize int) ([]service.StudentListItem, int, error)
	Get(ctx context.Context, id string) (*service.StudentDetail, error)
	Create(ctx context.Context, in service.CreateStudentInput) (string, error)
	Update(ctx context.Context, id string, in service.UpdateStudentInput) error
	Delete(ctx context.Context, id string) error
}

type StudentHandler struct {
	svc      StudentService
	validate *validator.Validate
}

func NewStudentHandler(svc StudentService) *StudentHandler {
	return &StudentHandler{svc: svc, validate: validator.New()}
}

type createStudentBody struct {
	NationalID  string      `json:"national_id" validate:"required,len=13,numeric"`
	StudentCode string      `json:"student_code" validate:"required,max=50"`
	Prefix      string      `json:"prefix" validate:"omitempty,max=50"`
	FirstName   string      `json:"first_name" validate:"required,max=150"`
	LastName    string      `json:"last_name" validate:"required,max=150"`
	BirthDate   string      `json:"birth_date" validate:"omitempty,datetime=2006-01-02"`
	Phone       string      `json:"phone" validate:"omitempty,max=20"`
	Address     addressBody `json:"address"`
}

type updateStudentBody struct {
	NationalID  string      `json:"national_id" validate:"omitempty,len=13,numeric"`
	StudentCode string      `json:"student_code" validate:"required,max=50"`
	Prefix      string      `json:"prefix" validate:"omitempty,max=50"`
	FirstName   string      `json:"first_name" validate:"required,max=150"`
	LastName    string      `json:"last_name" validate:"required,max=150"`
	BirthDate   string      `json:"birth_date" validate:"omitempty,datetime=2006-01-02"`
	Phone       string      `json:"phone" validate:"omitempty,max=20"`
	Address     addressBody `json:"address"`
}

func (h *StudentHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	items, total, err := h.svc.List(c.UserContext(), page, pageSize)
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OKWithMeta(c, items, &httputil.Meta{Page: page, Total: total})
}

func (h *StudentHandler) Get(c *fiber.Ctx) error {
	detail, err := h.svc.Get(c.UserContext(), c.Params("id"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, detail)
}

func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var body createStudentBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกข้อมูลให้ครบและถูกต้อง")
	}
	birth, err := parseBirthDate(body.BirthDate)
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบวันเกิดไม่ถูกต้อง (YYYY-MM-DD)")
	}
	id, err := h.svc.Create(c.UserContext(), service.CreateStudentInput{
		NationalID: body.NationalID, StudentCode: body.StudentCode, Prefix: body.Prefix,
		FirstName: body.FirstName, LastName: body.LastName, BirthDate: birth, Phone: body.Phone,
		Address: body.Address.toDTO(),
	})
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.Created(c, fiber.Map{"id": id})
}

func (h *StudentHandler) Update(c *fiber.Ctx) error {
	var body updateStudentBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกข้อมูลให้ครบและถูกต้อง")
	}
	birth, err := parseBirthDate(body.BirthDate)
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบวันเกิดไม่ถูกต้อง (YYYY-MM-DD)")
	}
	if err := h.svc.Update(c.UserContext(), c.Params("id"), service.UpdateStudentInput{
		NationalID: body.NationalID, StudentCode: body.StudentCode, Prefix: body.Prefix,
		FirstName: body.FirstName, LastName: body.LastName, BirthDate: birth, Phone: body.Phone,
		Address: body.Address.toDTO(),
	}); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "บันทึกข้อมูลนักเรียนแล้ว"})
}

func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.UserContext(), c.Params("id")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ลบข้อมูลนักเรียนแล้ว"})
}

// ================= Student ↔ Guardian links =================

type StudentGuardianService interface {
	List(ctx context.Context, studentID string) ([]service.StudentGuardianDTO, error)
	Link(ctx context.Context, studentID string, in service.LinkGuardianInput) error
	Unlink(ctx context.Context, studentID, linkID string) error
}

type StudentGuardianHandler struct {
	svc      StudentGuardianService
	validate *validator.Validate
}

func NewStudentGuardianHandler(svc StudentGuardianService) *StudentGuardianHandler {
	return &StudentGuardianHandler{svc: svc, validate: validator.New()}
}

type linkGuardianBody struct {
	NationalID   string      `json:"national_id" validate:"required,len=13,numeric"`
	Prefix       string      `json:"prefix" validate:"omitempty,max=50"`
	FirstName    string      `json:"first_name" validate:"required,max=150"`
	LastName     string      `json:"last_name" validate:"required,max=150"`
	BirthDate    string      `json:"birth_date" validate:"omitempty,datetime=2006-01-02"`
	Phone        string      `json:"phone" validate:"omitempty,max=20"`
	Address      addressBody `json:"address"`
	Relationship string      `json:"relationship" validate:"required,oneof=father mother other"`
	IsPrimary    bool        `json:"is_primary"`
}

func (h *StudentGuardianHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.List(c.UserContext(), c.Params("id"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

func (h *StudentGuardianHandler) Link(c *fiber.Ctx) error {
	var body linkGuardianBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกข้อมูลผู้ปกครองและความสัมพันธ์ให้ครบ")
	}
	birth, err := parseBirthDate(body.BirthDate)
	if err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบวันเกิดไม่ถูกต้อง (YYYY-MM-DD)")
	}
	if err := h.svc.Link(c.UserContext(), c.Params("id"), service.LinkGuardianInput{
		NationalID: body.NationalID, Prefix: body.Prefix, FirstName: body.FirstName, LastName: body.LastName,
		BirthDate: birth, Phone: body.Phone, Address: body.Address.toDTO(),
		Relationship: body.Relationship, IsPrimary: body.IsPrimary,
	}); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "เพิ่มผู้ปกครองแล้ว"})
}

func (h *StudentGuardianHandler) Unlink(c *fiber.Ctx) error {
	if err := h.svc.Unlink(c.UserContext(), c.Params("id"), c.Params("linkId")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ถอดผู้ปกครองแล้ว"})
}
