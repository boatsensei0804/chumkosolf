package handler

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chumkosoft/backend/internal/httputil"
	"github.com/chumkosoft/backend/internal/service"
)

// ================= Attendance (เช็คชื่อเข้าเรียน) =================

// AttendanceService contract ที่ handler ใช้
type AttendanceService interface {
	ListRoster(ctx context.Context, classID, dateStr string) ([]service.AttendanceRosterDTO, error)
	Save(ctx context.Context, classID, dateStr string, marks []service.AttendanceMarkInput) error
}

type AttendanceHandler struct {
	svc      AttendanceService
	validate *validator.Validate
}

func NewAttendanceHandler(svc AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{svc: svc, validate: validator.New()}
}

type attendanceMarkBody struct {
	StudentID string `json:"student_id" validate:"required,uuid"`
	Status    string `json:"status" validate:"required,oneof=present absent late sick_leave personal_leave"`
	Note      string `json:"note" validate:"max=500"`
}

type saveAttendanceBody struct {
	Date    string               `json:"date" validate:"required,datetime=2006-01-02"`
	Records []attendanceMarkBody `json:"records" validate:"required,dive"`
}

// List — GET /classes/:id/attendance?date=YYYY-MM-DD
func (h *AttendanceHandler) List(c *fiber.Ctx) error {
	date := c.Query("date")
	if date == "" {
		return httputil.Error(c, fiber.StatusBadRequest, "INVALID_DATE", "กรุณาระบุวันที่ (date=YYYY-MM-DD)")
	}
	items, err := h.svc.ListRoster(c.UserContext(), c.Params("id"), date)
	if err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, items)
}

// Save — POST /classes/:id/attendance
func (h *AttendanceHandler) Save(c *fiber.Ctx) error {
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
	if err := h.svc.Save(c.UserContext(), c.Params("id"), body.Date, marks); err != nil {
		return respondServiceError(c, err)
	}
	return httputil.OK(c, fiber.Map{"message": "บันทึกการเช็คชื่อแล้ว", "saved": len(marks)})
}
