package handler

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumko-platform/backend/internal/httputil"
	"github.com/chumko-platform/backend/internal/service"
)

// ================= Timetable (ตั้งค่าคาบ + ตารางสอน) =================

type TimetableService interface {
	GetConfig(ctx context.Context) (*service.TimetableConfigDTO, error)
	SaveConfig(ctx context.Context, in service.ConfigInput) error
	ListSlots(ctx context.Context, classID string) ([]service.TimetableSlotDTO, error)
	SetSlot(ctx context.Context, classID string, in service.SlotInput) (string, error)
	ClearSlot(ctx context.Context, classID, slotID string) error
	FreeTeachers(ctx context.Context, day int) (*service.FreeTeachersDTO, error)
}

type TimetableHandler struct {
	svc      TimetableService
	validate *validator.Validate
}

func NewTimetableHandler(svc TimetableService) *TimetableHandler {
	return &TimetableHandler{svc: svc, validate: validator.New()}
}

type periodBody struct {
	PeriodNo  int    `json:"period_no" validate:"required,min=1,max=20"`
	Label     string `json:"label" validate:"max=50"`
	StartTime string `json:"start_time" validate:"omitempty,len=5"`
	EndTime   string `json:"end_time" validate:"omitempty,len=5"`
	IsBreak   bool   `json:"is_break"`
}

type configBody struct {
	DaysPerWeek   int          `json:"days_per_week" validate:"required,min=1,max=7"`
	PeriodsPerDay int          `json:"periods_per_day" validate:"required,min=1,max=20"`
	Periods       []periodBody `json:"periods" validate:"dive"`
}

type slotBody struct {
	DayOfWeek            int    `json:"day_of_week" validate:"required,min=1,max=7"`
	PeriodNo             int    `json:"period_no" validate:"required,min=1,max=20"`
	TeachingAssignmentID string `json:"teaching_assignment_id" validate:"required,uuid"`
}

// GetConfig — GET /timetable/config
func (h *TimetableHandler) GetConfig(c *fiber.Ctx) error {
	cfg, err := h.svc.GetConfig(c.UserContext())
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, cfg)
}

// FreeTeachers — GET /timetable/free-teachers?day=N (1=จันทร์..7=อาทิตย์; ไม่ระบุ = วันนี้)
func (h *TimetableHandler) FreeTeachers(c *fiber.Ctx) error {
	day := c.QueryInt("day", 0)
	if day < 1 || day > 7 {
		wd := int(time.Now().UTC().Weekday()) // อาทิตย์=0
		if wd == 0 {
			wd = 7
		}
		day = wd
	}
	res, err := h.svc.FreeTeachers(c.UserContext(), day)
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, res)
}

// SaveConfig — PUT /timetable/config
func (h *TimetableHandler) SaveConfig(c *fiber.Ctx) error {
	var body configBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณากรอกจำนวนวัน/คาบและข้อมูลคาบให้ถูกต้อง")
	}
	periods := make([]service.PeriodInput, 0, len(body.Periods))
	for _, p := range body.Periods {
		periods = append(periods, service.PeriodInput{
			PeriodNo: p.PeriodNo, Label: p.Label, StartTime: p.StartTime, EndTime: p.EndTime, IsBreak: p.IsBreak,
		})
	}
	if err := h.svc.SaveConfig(c.UserContext(), service.ConfigInput{
		DaysPerWeek: body.DaysPerWeek, PeriodsPerDay: body.PeriodsPerDay, Periods: periods,
	}); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "บันทึกการตั้งค่าตารางสอนแล้ว"})
}

// ListSlots — GET /timetable/classes/:classId
func (h *TimetableHandler) ListSlots(c *fiber.Ctx) error {
	items, err := h.svc.ListSlots(c.UserContext(), c.Params("classId"))
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

// SetSlot — POST /timetable/classes/:classId/slots
func (h *TimetableHandler) SetSlot(c *fiber.Ctx) error {
	var body slotBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "กรุณาเลือกวัน คาบ และวิชาที่สอน")
	}
	id, err := h.svc.SetSlot(c.UserContext(), c.Params("classId"), service.SlotInput{
		DayOfWeek: body.DayOfWeek, PeriodNo: body.PeriodNo, TeachingAssignmentID: body.TeachingAssignmentID,
	})
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"id": id, "message": "บันทึกช่องตารางแล้ว"})
}

// ClearSlot — DELETE /timetable/classes/:classId/slots/:slotId
func (h *TimetableHandler) ClearSlot(c *fiber.Ctx) error {
	if err := h.svc.ClearSlot(c.UserContext(), c.Params("classId"), c.Params("slotId")); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "ล้างช่องตารางแล้ว"})
}
