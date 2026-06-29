package handler

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumko-platform/backend/internal/httputil"
	"github.com/chumko-platform/backend/internal/service"
)

// ================= Subject attendance (เช็คชื่อรายวิชา/รายคาบ) =================

type SubjectAttendanceService interface {
	ListRoster(ctx context.Context, slotID, dateStr string) ([]service.AttendanceRosterDTO, error)
	Save(ctx context.Context, slotID, dateStr string, marks []service.AttendanceMarkInput) error
	CheckinOverview(ctx context.Context, dateStr string) (*service.CheckinOverviewDTO, error)
}

type SubjectAttendanceHandler struct {
	svc      SubjectAttendanceService
	validate *validator.Validate
}

func NewSubjectAttendanceHandler(svc SubjectAttendanceService) *SubjectAttendanceHandler {
	return &SubjectAttendanceHandler{svc: svc, validate: validator.New()}
}

// List — GET /timetable/slots/:slotId/attendance?date=YYYY-MM-DD
func (h *SubjectAttendanceHandler) List(c *fiber.Ctx) error {
	date := c.Query("date")
	if date == "" {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_DATE", "กรุณาระบุวันที่ (date=YYYY-MM-DD)")
	}
	items, err := h.svc.ListRoster(c.UserContext(), c.Params("slotId"), date)
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

// Overview — GET /timetable/my-checkin?date=YYYY-MM-DD
func (h *SubjectAttendanceHandler) Overview(c *fiber.Ctx) error {
	date := c.Query("date")
	if date == "" {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_DATE", "กรุณาระบุวันที่ (date=YYYY-MM-DD)")
	}
	res, err := h.svc.CheckinOverview(c.UserContext(), date)
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, res)
}

// Save — POST /timetable/slots/:slotId/attendance
func (h *SubjectAttendanceHandler) Save(c *fiber.Ctx) error {
	var body saveAttendanceBody
	if err := c.BodyParser(&body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "รูปแบบข้อมูลไม่ถูกต้อง")
	}
	if err := h.validate.Struct(body); err != nil {
		return httputil.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "ข้อมูลการเช็คชื่อไม่ถูกต้อง")
	}
	marks := make([]service.AttendanceMarkInput, 0, len(body.Records))
	for _, r := range body.Records {
		marks = append(marks, service.AttendanceMarkInput{StudentID: r.StudentID, Status: r.Status, Note: r.Note})
	}
	if err := h.svc.Save(c.UserContext(), c.Params("slotId"), body.Date, marks); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "บันทึกการเช็คชื่อรายวิชาแล้ว", "saved": len(marks)})
}
