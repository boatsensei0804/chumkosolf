package handler

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumkosoft/backend/internal/httputil"
	"github.com/chumkosoft/backend/internal/service"
)

// SchoolService contract ที่ handler ใช้
type SchoolService interface {
	Get(ctx context.Context) (*service.SchoolDTO, error)
	Update(ctx context.Context, in service.UpdateSchoolInput) error
}

type SchoolHandler struct {
	svc      SchoolService
	validate *validator.Validate
}

func NewSchoolHandler(svc SchoolService) *SchoolHandler {
	return &SchoolHandler{svc: svc, validate: validator.New()}
}

type schoolAddressBody struct {
	HouseNo     string `json:"house_no" validate:"max=50"`
	Moo         string `json:"moo" validate:"max=50"`
	Road        string `json:"road" validate:"max=100"`
	Subdistrict string `json:"subdistrict" validate:"max=100"`
	District    string `json:"district" validate:"max=100"`
	Province    string `json:"province" validate:"max=100"`
	PostalCode  string `json:"postal_code" validate:"max=10"`
}

type updateSchoolBody struct {
	Name                string            `json:"name" validate:"required,max=255"`
	Address             schoolAddressBody `json:"address"`
	Phone               string            `json:"phone" validate:"max=20"`
	Email               string            `json:"email" validate:"omitempty,email,max=150"`
	Website             string            `json:"website" validate:"max=255"`
	DirectorName          string            `json:"director_name" validate:"max=150"`
	AttendanceLateAfter   string            `json:"attendance_late_after" validate:"omitempty,len=5"`
	AttendanceLatePenalty int               `json:"attendance_late_penalty" validate:"min=0,max=100"`
}

// Get — GET /school
func (h *SchoolHandler) Get(c *fiber.Ctx) error {
	res, err := h.svc.Get(c.UserContext())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, res)
}

// Update — PUT /school (school admin)
func (h *SchoolHandler) Update(c *fiber.Ctx) error {
	var body updateSchoolBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกชื่อโรงเรียน")
	}
	if err := h.svc.Update(c.UserContext(), service.UpdateSchoolInput{
		Name: body.Name,
		Address: service.AddressDTO{
			HouseNo: body.Address.HouseNo, Moo: body.Address.Moo, Road: body.Address.Road, Subdistrict: body.Address.Subdistrict,
			District: body.Address.District, Province: body.Address.Province, PostalCode: body.Address.PostalCode,
		},
		Phone: body.Phone, Email: body.Email, Website: body.Website, DirectorName: body.DirectorName,
		AttendanceLateAfter: body.AttendanceLateAfter, AttendanceLatePenalty: body.AttendanceLatePenalty,
	}); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "บันทึกข้อมูลโรงเรียนแล้ว"})
}
