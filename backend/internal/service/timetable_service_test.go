package service

import (
	"context"
	"errors"
	"testing"

	"github.com/chumkosoft/backend/internal/domain"
)

// --- fake timetable repo ---

type fakeTimetableRepo struct {
	settings      map[string]*domain.TimetableSettings
	periods       map[string][]domain.PeriodDefinition
	slots         map[string]*domain.TimetableSlot
	teachers      []domain.TeacherBrief
	busy          []domain.TeacherPeriod
	seq           int
	forceConflict bool
}

func newFakeTimetableRepo() *fakeTimetableRepo {
	return &fakeTimetableRepo{
		settings: map[string]*domain.TimetableSettings{},
		periods:  map[string][]domain.PeriodDefinition{},
		slots:    map[string]*domain.TimetableSlot{},
	}
}

func ttKey(schoolID, semesterID string) string { return schoolID + "|" + semesterID }

func (r *fakeTimetableRepo) GetSettings(_ context.Context, schoolID, semesterID string) (*domain.TimetableSettings, error) {
	return r.settings[ttKey(schoolID, semesterID)], nil
}

func (r *fakeTimetableRepo) ListPeriods(_ context.Context, schoolID, semesterID string) ([]domain.PeriodDefinition, error) {
	return r.periods[ttKey(schoolID, semesterID)], nil
}

func (r *fakeTimetableRepo) SaveConfig(_ context.Context, schoolID, semesterID string, days, periods int, defs []domain.NewPeriodDefinition, _ domain.AuditEntry) error {
	r.settings[ttKey(schoolID, semesterID)] = &domain.TimetableSettings{SchoolID: schoolID, SemesterID: semesterID, DaysPerWeek: days, PeriodsPerDay: periods}
	var out []domain.PeriodDefinition
	for _, d := range defs {
		out = append(out, domain.PeriodDefinition{PeriodNo: d.PeriodNo, Label: d.Label, IsBreak: d.IsBreak})
	}
	r.periods[ttKey(schoolID, semesterID)] = out
	return nil
}

func (r *fakeTimetableRepo) ListSlotsByClass(_ context.Context, schoolID, classID string) ([]domain.TimetableSlot, error) {
	var out []domain.TimetableSlot
	for _, s := range r.slots {
		if s.SchoolID == schoolID && s.ClassID == classID {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (r *fakeTimetableRepo) ListTeachers(_ context.Context, _ string) ([]domain.TeacherBrief, error) {
	return r.teachers, nil
}

func (r *fakeTimetableRepo) BusyTeacherPeriods(_ context.Context, _, _ string, _ int) ([]domain.TeacherPeriod, error) {
	return r.busy, nil
}

func (r *fakeTimetableRepo) TeacherSlotConflict(_ context.Context, _, _, _ string, _, _ int, _ string) (bool, error) {
	return r.forceConflict, nil
}

func (r *fakeTimetableRepo) UpsertSlot(_ context.Context, schoolID, semesterID, classID string, ns domain.NewTimetableSlot, _ domain.AuditEntry) (string, error) {
	r.seq++
	id := "slot" + string(rune('0'+r.seq))
	r.slots[id] = &domain.TimetableSlot{
		ID: id, SchoolID: schoolID, SemesterID: semesterID, ClassID: classID,
		DayOfWeek: ns.DayOfWeek, PeriodNo: ns.PeriodNo, TeachingAssignmentID: ns.TeachingAssignmentID,
	}
	return id, nil
}

func (r *fakeTimetableRepo) DeleteSlot(_ context.Context, schoolID, classID, slotID string, _ domain.AuditEntry) (bool, error) {
	s, ok := r.slots[slotID]
	if !ok || s.SchoolID != schoolID || s.ClassID != classID {
		return false, nil
	}
	delete(r.slots, slotID)
	return true, nil
}

const ttSchool = "school-A"
const ttClass = "c1"

func newTimetableSvc() (*TimetableService, *fakeTimetableRepo, *fakeTaRepo, *fakeChecker) {
	repo := newFakeTimetableRepo()
	taFake := newFakeTaRepo()
	taFake.items["ta1"] = &domain.TeachingAssignment{ID: "ta1", SchoolID: ttSchool, SemesterID: "sem-1", ClassID: ttClass}
	classes := &fakeClassExists{byID: map[string]*domain.Class{
		ttClass: {ID: ttClass, SchoolID: ttSchool, SemesterID: "sem-1"},
	}}
	checker := &fakeChecker{groups: map[string]bool{}}
	return NewTimetableService(repo, taFake, classes, checker), repo, taFake, checker
}

// --- config tests ---

func TestTimetable_GetConfigDefault(t *testing.T) {
	svc, _, _, _ := newTimetableSvc()
	cfg, err := svc.GetConfig(wAdmin(ttSchool, "sem-1"))
	if err != nil {
		t.Fatalf("get config: %v", err)
	}
	if cfg.DaysPerWeek != defaultDaysPerWeek || cfg.PeriodsPerDay != defaultPeriodsPerDay {
		t.Errorf("default = %d/%d ควรเป็น %d/%d", cfg.DaysPerWeek, cfg.PeriodsPerDay, defaultDaysPerWeek, defaultPeriodsPerDay)
	}
}

func TestTimetable_SaveConfigInvalidDays(t *testing.T) {
	svc, _, _, _ := newTimetableSvc()
	err := svc.SaveConfig(wAdmin(ttSchool, "sem-1"), ConfigInput{DaysPerWeek: 0, PeriodsPerDay: 8})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
}

func TestTimetable_SaveConfigDuplicatePeriod(t *testing.T) {
	svc, _, _, _ := newTimetableSvc()
	err := svc.SaveConfig(wAdmin(ttSchool, "sem-1"), ConfigInput{
		DaysPerWeek: 5, PeriodsPerDay: 8,
		Periods: []PeriodInput{{PeriodNo: 1}, {PeriodNo: 1}},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
}

func TestTimetable_SaveConfigSuccess(t *testing.T) {
	svc, repo, _, _ := newTimetableSvc()
	err := svc.SaveConfig(wAdmin(ttSchool, "sem-1"), ConfigInput{
		DaysPerWeek: 5, PeriodsPerDay: 2,
		Periods: []PeriodInput{{PeriodNo: 1, Label: "คาบ 1"}, {PeriodNo: 2, Label: "คาบ 2"}},
	})
	if err != nil {
		t.Fatalf("save config: %v", err)
	}
	cfg, _ := svc.GetConfig(wAdmin(ttSchool, "sem-1"))
	if cfg.PeriodsPerDay != 2 || len(cfg.Periods) != 2 {
		t.Errorf("saved cfg = %+v ควร periods_per_day=2 และ 2 คาบ", cfg)
	}
	_ = repo
}

func TestTimetable_GetConfigOpenToAnyUser(t *testing.T) {
	svc, _, _, _ := newTimetableSvc()
	// ครูทั่วไป (ไม่ใช่กลุ่มวิชาการ) ต้องดู config ได้
	if _, err := svc.GetConfig(wMember(ttSchool, "u9", "sem-1")); err != nil {
		t.Errorf("GetConfig ควรเปิดให้ทุก auth, ได้ err=%v", err)
	}
}

func TestTimetable_SaveConfigForbiddenForNonMember(t *testing.T) {
	svc, _, _, _ := newTimetableSvc()
	// แก้ไขยังจำกัดเฉพาะกลุ่มวิชาการ
	err := svc.SaveConfig(wMember(ttSchool, "u9", "sem-1"), ConfigInput{DaysPerWeek: 5, PeriodsPerDay: 6})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

// --- slot tests ---

func TestTimetable_SetSlotSuccess(t *testing.T) {
	svc, _, _, _ := newTimetableSvc()
	id, err := svc.SetSlot(wAdmin(ttSchool, "sem-1"), ttClass, SlotInput{DayOfWeek: 1, PeriodNo: 1, TeachingAssignmentID: "ta1"})
	if err != nil || id == "" {
		t.Fatalf("set slot: id=%q err=%v", id, err)
	}
}

func TestTimetable_SetSlotInvalidDay(t *testing.T) {
	svc, _, _, _ := newTimetableSvc()
	_, err := svc.SetSlot(wAdmin(ttSchool, "sem-1"), ttClass, SlotInput{DayOfWeek: 9, PeriodNo: 1, TeachingAssignmentID: "ta1"})
	if !errors.Is(err, domain.ErrInvalidTimetableSlot) {
		t.Errorf("err = %v, want ErrInvalidTimetableSlot", err)
	}
}

func TestTimetable_SetSlotAssignmentNotFound(t *testing.T) {
	svc, _, _, _ := newTimetableSvc()
	_, err := svc.SetSlot(wAdmin(ttSchool, "sem-1"), ttClass, SlotInput{DayOfWeek: 1, PeriodNo: 1, TeachingAssignmentID: "ghost"})
	if !errors.Is(err, domain.ErrTeachingAssignmentNotFound) {
		t.Errorf("err = %v, want ErrTeachingAssignmentNotFound", err)
	}
}

func TestTimetable_SetSlotAssignmentWrongClass(t *testing.T) {
	svc, _, taFake, _ := newTimetableSvc()
	// ta2 เป็นของห้องอื่น
	taFake.items["ta2"] = &domain.TeachingAssignment{ID: "ta2", SchoolID: ttSchool, ClassID: "other-class"}
	_, err := svc.SetSlot(wAdmin(ttSchool, "sem-1"), ttClass, SlotInput{DayOfWeek: 1, PeriodNo: 1, TeachingAssignmentID: "ta2"})
	if !errors.Is(err, domain.ErrInvalidTimetableSlot) {
		t.Errorf("err = %v, want ErrInvalidTimetableSlot (เอาวิชาห้องอื่นมาใส่)", err)
	}
}

func TestTimetable_SetSlotTeacherConflict(t *testing.T) {
	svc, repo, _, _ := newTimetableSvc()
	repo.forceConflict = true // ครูติดสอนห้องอื่นในคาบนี้
	_, err := svc.SetSlot(wAdmin(ttSchool, "sem-1"), ttClass, SlotInput{DayOfWeek: 1, PeriodNo: 1, TeachingAssignmentID: "ta1"})
	if !errors.Is(err, domain.ErrTeacherTimeConflict) {
		t.Errorf("err = %v, want ErrTeacherTimeConflict", err)
	}
}

func TestTimetable_ClearSlotNotFound(t *testing.T) {
	svc, _, _, _ := newTimetableSvc()
	if err := svc.ClearSlot(wAdmin(ttSchool, "sem-1"), ttClass, "missing"); !errors.Is(err, domain.ErrTimetableSlotNotFound) {
		t.Errorf("err = %v, want ErrTimetableSlotNotFound", err)
	}
}

func TestTimetable_FreeTeachers(t *testing.T) {
	svc, repo, _, _ := newTimetableSvc()
	repo.periods[ttKey(ttSchool, "sem-1")] = []domain.PeriodDefinition{
		{PeriodNo: 1, IsBreak: false},
		{PeriodNo: 2, IsBreak: true}, // คาบพัก → ต้องข้าม
		{PeriodNo: 3, IsBreak: false},
	}
	repo.teachers = []domain.TeacherBrief{
		{ID: "t1", FirstName: "ครูเอ"}, {ID: "t2", FirstName: "ครูบี"}, {ID: "t3", FirstName: "ครูซี"},
	}
	repo.busy = []domain.TeacherPeriod{{PersonnelID: "t1", PeriodNo: 1}, {PersonnelID: "t2", PeriodNo: 3}}

	res, err := svc.FreeTeachers(wAdmin(ttSchool, "sem-1"), 1)
	if err != nil {
		t.Fatalf("free teachers: %v", err)
	}
	if len(res.Periods) != 2 {
		t.Fatalf("ควรมี 2 คาบ (ข้ามคาบพัก), ได้ %d", len(res.Periods))
	}
	// คาบ 1: t1 ติดสอน → ว่าง t2,t3
	free1 := map[string]bool{}
	for _, ft := range res.Periods[0].FreeTeachers {
		free1[ft.ID] = true
	}
	if free1["t1"] || !free1["t2"] || !free1["t3"] {
		t.Errorf("คาบ 1 ครูว่างผิด: %+v", res.Periods[0].FreeTeachers)
	}
}
