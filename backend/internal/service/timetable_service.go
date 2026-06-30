package service

import (
	"context"
	"strings"

	"github.com/chumkosoft/backend/internal/domain"
	"github.com/chumkosoft/backend/internal/tenant"
)

const (
	defaultDaysPerWeek   = 5
	defaultPeriodsPerDay = 8
)

// TimetableRepository contract ของชั้น DB
type TimetableRepository interface {
	GetSettings(ctx context.Context, schoolID, semesterID string) (*domain.TimetableSettings, error)
	ListPeriods(ctx context.Context, schoolID, semesterID string) ([]domain.PeriodDefinition, error)
	SaveConfig(ctx context.Context, schoolID, semesterID string, daysPerWeek, periodsPerDay int, periods []domain.NewPeriodDefinition, audit domain.AuditEntry) error
	ListSlotsByClass(ctx context.Context, schoolID, classID string) ([]domain.TimetableSlot, error)
	ListTeachers(ctx context.Context, schoolID string) ([]domain.TeacherBrief, error)
	BusyTeacherPeriods(ctx context.Context, schoolID, semesterID string, dayOfWeek int) ([]domain.TeacherPeriod, error)
	TeacherSlotConflict(ctx context.Context, schoolID, semesterID, personnelID string, dayOfWeek, periodNo int, excludeClassID string) (bool, error)
	UpsertSlot(ctx context.Context, schoolID, semesterID, classID string, ns domain.NewTimetableSlot, audit domain.AuditEntry) (string, error)
	DeleteSlot(ctx context.Context, schoolID, classID, slotID string, audit domain.AuditEntry) (bool, error)
}

type ttTeachingRepo interface {
	GetByID(ctx context.Context, schoolID, id string) (*domain.TeachingAssignment, error)
}
type ttClassRepo interface {
	GetByID(ctx context.Context, schoolID, id string) (*domain.Class, error)
}

// DTOs
type PeriodDTO struct {
	PeriodNo  int    `json:"period_no"`
	Label     string `json:"label"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	IsBreak   bool   `json:"is_break"`
}

type TimetableConfigDTO struct {
	DaysPerWeek   int         `json:"days_per_week"`
	PeriodsPerDay int         `json:"periods_per_day"`
	Periods       []PeriodDTO `json:"periods"`
}

type TimetableSlotDTO struct {
	ID                   string `json:"id"`
	DayOfWeek            int    `json:"day_of_week"`
	PeriodNo             int    `json:"period_no"`
	TeachingAssignmentID string `json:"teaching_assignment_id"`
	SubjectCode          string `json:"subject_code"`
	SubjectName          string `json:"subject_name"`
	TeacherName          string `json:"teacher_name"`
}

// DTOs สำหรับ "ครูว่างวันนี้"
type TeacherBriefDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type FreePeriodDTO struct {
	PeriodNo     int               `json:"period_no"`
	Label        string            `json:"label"`
	FreeTeachers []TeacherBriefDTO `json:"free_teachers"`
}
type FreeTeachersDTO struct {
	Day     int             `json:"day"`
	Periods []FreePeriodDTO `json:"periods"`
}

// Inputs
type ConfigInput struct {
	DaysPerWeek   int
	PeriodsPerDay int
	Periods       []PeriodInput
}
type PeriodInput struct {
	PeriodNo  int
	Label     string
	StartTime string
	EndTime   string
	IsBreak   bool
}
type SlotInput struct {
	DayOfWeek            int
	PeriodNo             int
	TeachingAssignmentID string
}

// TimetableService จัดการตั้งค่าคาบ + ตารางสอน (กลุ่มวิชาการ + admin, รายเทอม)
type TimetableService struct {
	guard    academicGuard
	repo     TimetableRepository
	teaching ttTeachingRepo
	classes  ttClassRepo
}

func NewTimetableService(repo TimetableRepository, teaching ttTeachingRepo, classes ttClassRepo, checker WorkGroupChecker) *TimetableService {
	return &TimetableService{guard: academicGuard{checker: checker}, repo: repo, teaching: teaching, classes: classes}
}

// GetConfig คืนค่าตั้ง + นิยามคาบของเทอม (ใช้ค่า default ถ้ายังไม่เคยตั้ง)
// GetConfig — ดูได้ทุกผู้ใช้ที่ล็อกอิน (ครูดูตารางได้) การแก้ไขเท่านั้นที่จำกัดกลุ่มวิชาการ
func (s *TimetableService) GetConfig(ctx context.Context) (*TimetableConfigDTO, error) {
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return nil, err
	}
	schoolID := tenant.SchoolIDFromContext(ctx)
	settings, err := s.repo.GetSettings(ctx, schoolID, sem)
	if err != nil {
		return nil, err
	}
	periods, err := s.repo.ListPeriods(ctx, schoolID, sem)
	if err != nil {
		return nil, err
	}
	cfg := &TimetableConfigDTO{DaysPerWeek: defaultDaysPerWeek, PeriodsPerDay: defaultPeriodsPerDay, Periods: []PeriodDTO{}}
	if settings != nil {
		cfg.DaysPerWeek = settings.DaysPerWeek
		cfg.PeriodsPerDay = settings.PeriodsPerDay
	}
	for i := range periods {
		cfg.Periods = append(cfg.Periods, PeriodDTO{
			PeriodNo: periods[i].PeriodNo, Label: periods[i].Label,
			StartTime: periods[i].StartTime, EndTime: periods[i].EndTime, IsBreak: periods[i].IsBreak,
		})
	}
	return cfg, nil
}

// FreeTeachers คืนรายชื่อครูที่ "ว่าง" (ไม่ติดสอน) ในแต่ละคาบของวันที่ระบุ
// ดูได้ทุกผู้ใช้ที่ล็อกอิน (อ่านอย่างเดียว) — ข้ามคาบพัก
func (s *TimetableService) FreeTeachers(ctx context.Context, day int) (*FreeTeachersDTO, error) {
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return nil, err
	}
	schoolID := tenant.SchoolIDFromContext(ctx)

	periods, err := s.repo.ListPeriods(ctx, schoolID, sem)
	if err != nil {
		return nil, err
	}
	teachers, err := s.repo.ListTeachers(ctx, schoolID)
	if err != nil {
		return nil, err
	}
	busy, err := s.repo.BusyTeacherPeriods(ctx, schoolID, sem, day)
	if err != nil {
		return nil, err
	}

	busyByPeriod := make(map[int]map[string]bool)
	for _, bp := range busy {
		if busyByPeriod[bp.PeriodNo] == nil {
			busyByPeriod[bp.PeriodNo] = make(map[string]bool)
		}
		busyByPeriod[bp.PeriodNo][bp.PersonnelID] = true
	}

	out := &FreeTeachersDTO{Day: day, Periods: []FreePeriodDTO{}}
	for i := range periods {
		p := &periods[i]
		if p.IsBreak {
			continue
		}
		free := make([]TeacherBriefDTO, 0)
		for j := range teachers {
			t := &teachers[j]
			if busyByPeriod[p.PeriodNo][t.ID] {
				continue
			}
			free = append(free, TeacherBriefDTO{
				ID:   t.ID,
				Name: strings.TrimSpace(t.Prefix + t.FirstName + " " + t.LastName),
			})
		}
		out.Periods = append(out.Periods, FreePeriodDTO{PeriodNo: p.PeriodNo, Label: p.Label, FreeTeachers: free})
	}
	return out, nil
}

// SaveConfig บันทึกค่าตั้ง + นิยามคาบทั้งชุด
func (s *TimetableService) SaveConfig(ctx context.Context, in ConfigInput) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return err
	}
	if in.DaysPerWeek < 1 || in.DaysPerWeek > 7 || in.PeriodsPerDay < 1 || in.PeriodsPerDay > 20 {
		return domain.ErrValidation
	}
	defs := make([]domain.NewPeriodDefinition, 0, len(in.Periods))
	seen := make(map[int]bool, len(in.Periods))
	for _, p := range in.Periods {
		if p.PeriodNo < 1 || p.PeriodNo > 20 || seen[p.PeriodNo] {
			return domain.ErrValidation
		}
		seen[p.PeriodNo] = true
		defs = append(defs, domain.NewPeriodDefinition{
			PeriodNo: p.PeriodNo, Label: strings.TrimSpace(p.Label),
			StartTime: strings.TrimSpace(p.StartTime), EndTime: strings.TrimSpace(p.EndTime), IsBreak: p.IsBreak,
		})
	}
	audit := auditFor(ctx, domain.AuditUpdate, "timetable_settings", "", map[string]any{"days": in.DaysPerWeek, "periods": in.PeriodsPerDay})
	return s.repo.SaveConfig(ctx, tenant.SchoolIDFromContext(ctx), sem, in.DaysPerWeek, in.PeriodsPerDay, defs, audit)
}

func (s *TimetableService) ensureClass(ctx context.Context, classID string) error {
	c, err := s.classes.GetByID(ctx, tenant.SchoolIDFromContext(ctx), classID)
	if err != nil {
		return err
	}
	if c == nil {
		return domain.ErrClassNotFound
	}
	return nil
}

// ListSlots คืนช่องตารางสอนของห้อง — ดูได้ทุกผู้ใช้ที่ล็อกอิน (ครูดูตารางห้องได้)
func (s *TimetableService) ListSlots(ctx context.Context, classID string) ([]TimetableSlotDTO, error) {
	if err := s.ensureClass(ctx, classID); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListSlotsByClass(ctx, tenant.SchoolIDFromContext(ctx), classID)
	if err != nil {
		return nil, err
	}
	out := make([]TimetableSlotDTO, 0, len(rows))
	for i := range rows {
		out = append(out, TimetableSlotDTO{
			ID: rows[i].ID, DayOfWeek: rows[i].DayOfWeek, PeriodNo: rows[i].PeriodNo,
			TeachingAssignmentID: rows[i].TeachingAssignmentID,
			SubjectCode:          rows[i].SubjectCode, SubjectName: rows[i].SubjectName,
			TeacherName: strings.TrimSpace(rows[i].TeacherPrefix + rows[i].TeacherFirstName + " " + rows[i].TeacherLastName),
		})
	}
	return out, nil
}

// SetSlot ตั้งค่าช่องตาราง (ห้อง×วัน×คาบ) → มอบหมายการสอน (ต้องเป็นของห้องเดียวกัน)
func (s *TimetableService) SetSlot(ctx context.Context, classID string, in SlotInput) (string, error) {
	if err := s.guard.authorize(ctx); err != nil {
		return "", err
	}
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return "", err
	}
	if in.DayOfWeek < 1 || in.DayOfWeek > 7 || in.PeriodNo < 1 || in.PeriodNo > 20 {
		return "", domain.ErrInvalidTimetableSlot
	}
	if err := s.ensureClass(ctx, classID); err != nil {
		return "", err
	}
	ta, err := s.teaching.GetByID(ctx, tenant.SchoolIDFromContext(ctx), in.TeachingAssignmentID)
	if err != nil {
		return "", err
	}
	if ta == nil {
		return "", domain.ErrTeachingAssignmentNotFound
	}
	// การมอบหมายต้องเป็นของห้องนี้ (กันเอาวิชาห้องอื่นมาใส่)
	if ta.ClassID != classID {
		return "", domain.ErrInvalidTimetableSlot
	}
	// กันครูสอน 2 ห้องพร้อมกัน (ชนวัน+คาบในห้องอื่น)
	conflict, err := s.repo.TeacherSlotConflict(ctx, tenant.SchoolIDFromContext(ctx), sem, ta.PersonnelID, in.DayOfWeek, in.PeriodNo, classID)
	if err != nil {
		return "", err
	}
	if conflict {
		return "", domain.ErrTeacherTimeConflict
	}
	audit := auditFor(ctx, domain.AuditUpdate, "timetable_slot", "", map[string]any{
		"class_id": classID, "day": in.DayOfWeek, "period": in.PeriodNo,
	})
	return s.repo.UpsertSlot(ctx, tenant.SchoolIDFromContext(ctx), sem, classID, domain.NewTimetableSlot{
		DayOfWeek: in.DayOfWeek, PeriodNo: in.PeriodNo, TeachingAssignmentID: in.TeachingAssignmentID,
	}, audit)
}

// ClearSlot ล้างช่องตาราง
func (s *TimetableService) ClearSlot(ctx context.Context, classID, slotID string) error {
	if err := s.guard.authorize(ctx); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditDelete, "timetable_slot", slotID, map[string]any{"class_id": classID})
	found, err := s.repo.DeleteSlot(ctx, tenant.SchoolIDFromContext(ctx), classID, slotID, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrTimetableSlotNotFound
	}
	return nil
}
