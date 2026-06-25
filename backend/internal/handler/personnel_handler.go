package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumko-platform/backend/internal/httputil"
	"github.com/chumko-platform/backend/internal/service"
)

// PersonnelService คือ contract ของ service ที่ handler ใช้ (interface เพื่อ test ง่าย)
type PersonnelService interface {
	List(ctx context.Context, page, pageSize int) ([]service.PersonnelListItem, int, error)
	Get(ctx context.Context, id string) (*service.PersonnelDetail, error)
	Create(ctx context.Context, in service.CreatePersonnelInput) (string, error)
	Update(ctx context.Context, id string, in service.UpdatePersonnelInput) error
	Delete(ctx context.Context, id string) error
}

// PersonnelHandler จัดการ endpoint กลุ่ม /personnel
type PersonnelHandler struct {
	svc      PersonnelService
	validate *validator.Validate
}

// NewPersonnelHandler สร้าง handler
func NewPersonnelHandler(svc PersonnelService) *PersonnelHandler {
	return &PersonnelHandler{svc: svc, validate: validator.New()}
}

type addressBody struct {
	HouseNo     string `json:"house_no"`
	Moo         string `json:"moo"`
	Road        string `json:"road"`
	Subdistrict string `json:"subdistrict"`
	District    string `json:"district"`
	Province    string `json:"province"`
	PostalCode  string `json:"postal_code" validate:"omitempty,max=10"`
}

func (a addressBody) toDTO() service.AddressDTO {
	return service.AddressDTO{
		HouseNo:     a.HouseNo,
		Moo:         a.Moo,
		Road:        a.Road,
		Subdistrict: a.Subdistrict,
		District:    a.District,
		Province:    a.Province,
		PostalCode:  a.PostalCode,
	}
}

type createPersonnelBody struct {
	Username       string      `json:"username" validate:"required,min=3,max=100"`
	Password       string      `json:"password" validate:"required,min=8"`
	Role           string      `json:"role" validate:"required,oneof=teacher executive"`
	NationalID     string      `json:"national_id" validate:"required,len=13,numeric"`
	CivilServantID string      `json:"civil_servant_id" validate:"omitempty,max=50"`
	Prefix         string      `json:"prefix" validate:"omitempty,max=50"`
	FirstName      string      `json:"first_name" validate:"required,max=150"`
	LastName       string      `json:"last_name" validate:"required,max=150"`
	BirthDate      string      `json:"birth_date" validate:"omitempty,datetime=2006-01-02"`
	Phone          string      `json:"phone" validate:"omitempty,max=20"`
	Email          string      `json:"email" validate:"omitempty,email"`
	Address        addressBody `json:"address"`
}

type updatePersonnelBody struct {
	NationalID     string      `json:"national_id" validate:"omitempty,len=13,numeric"`
	CivilServantID string      `json:"civil_servant_id" validate:"omitempty,max=50"`
	Prefix         string      `json:"prefix" validate:"omitempty,max=50"`
	FirstName      string      `json:"first_name" validate:"required,max=150"`
	LastName       string      `json:"last_name" validate:"required,max=150"`
	BirthDate      string      `json:"birth_date" validate:"omitempty,datetime=2006-01-02"`
	Phone          string      `json:"phone" validate:"omitempty,max=20"`
	Email          string      `json:"email" validate:"omitempty,email"`
	Address        addressBody `json:"address"`
}

// List godoc — GET /api/v1/personnel?page=&page_size=
func (h *PersonnelHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	items, total, err := h.svc.List(c.UserContext(), page, pageSize)
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OKWithMeta(c, items, &httputil.Meta{Page: page, Total: total})
}

// Get godoc — GET /api/v1/personnel/:id
func (h *PersonnelHandler) Get(c *fiber.Ctx) error {
	detail, err := h.svc.Get(c.UserContext(), c.Params("id"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, detail)
}

// Create godoc — POST /api/v1/personnel
func (h *PersonnelHandler) Create(c *fiber.Ctx) error {
	var body createPersonnelBody
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

	id, err := h.svc.Create(c.UserContext(), service.CreatePersonnelInput{
		Username:       body.Username,
		Password:       body.Password,
		Role:           body.Role,
		NationalID:     body.NationalID,
		CivilServantID: body.CivilServantID,
		Prefix:         body.Prefix,
		FirstName:      body.FirstName,
		LastName:       body.LastName,
		BirthDate:      birth,
		Phone:          body.Phone,
		Email:          body.Email,
		Address:        body.Address.toDTO(),
	})
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.Created(c, fiber.Map{"id": id})
}

// Update godoc — PUT /api/v1/personnel/:id
func (h *PersonnelHandler) Update(c *fiber.Ctx) error {
	var body updatePersonnelBody
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

	if err := h.svc.Update(c.UserContext(), c.Params("id"), service.UpdatePersonnelInput{
		NationalID:     body.NationalID,
		CivilServantID: body.CivilServantID,
		Prefix:         body.Prefix,
		FirstName:      body.FirstName,
		LastName:       body.LastName,
		BirthDate:      birth,
		Phone:          body.Phone,
		Email:          body.Email,
		Address:        body.Address.toDTO(),
	}); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "บันทึกข้อมูลบุคลากรแล้ว"})
}

// Delete godoc — DELETE /api/v1/personnel/:id
func (h *PersonnelHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.UserContext(), c.Params("id")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ลบข้อมูลบุคลากรแล้ว"})
}

// parseBirthDate แปลง "YYYY-MM-DD" → *time.Time (ว่าง = nil)
func parseBirthDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
