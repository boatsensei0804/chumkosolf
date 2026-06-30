package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chumkosoft/backend/internal/domain"
)

type slotCtx struct {
	classID       string
	teacherUserID string
}

type fakeSubjectAttendanceRepo struct {
	slots        map[string]slotCtx
	roster       map[string][]domain.AttendanceRosterEntry // key: classID
	saved        []domain.AttendanceMark
	saveCalls    int
	teacherSlots []domain.TeacherCheckinSlot
	checkedDates []domain.SlotDate
	semStart     *time.Time
	semEnd       *time.Time
}

func newFakeSubjectAttendanceRepo() *fakeSubjectAttendanceRepo {
	return &fakeSubjectAttendanceRepo{slots: map[string]slotCtx{}, roster: map[string][]domain.AttendanceRosterEntry{}}
}

func (r *fakeSubjectAttendanceRepo) SlotContext(_ context.Context, _, slotID string) (string, string, bool, error) {
	sc, ok := r.slots[slotID]
	if !ok {
		return "", "", false, nil
	}
	return sc.classID, sc.teacherUserID, true, nil
}

func (r *fakeSubjectAttendanceRepo) RosterBySlotDate(_ context.Context, _, _, classID string, _ time.Time) ([]domain.AttendanceRosterEntry, error) {
	return r.roster[classID], nil
}

func (r *fakeSubjectAttendanceRepo) BulkUpsert(_ context.Context, _, _, _ string, _ time.Time, marks []domain.AttendanceMark, _ string, _ domain.AuditEntry) error {
	r.saved = append(r.saved, marks...)
	r.saveCalls++
	return nil
}

func (r *fakeSubjectAttendanceRepo) TeacherSlots(_ context.Context, _, _, _ string) ([]domain.TeacherCheckinSlot, error) {
	return r.teacherSlots, nil
}
func (r *fakeSubjectAttendanceRepo) CheckedSlotDates(_ context.Context, _, _, _ string) ([]domain.SlotDate, error) {
	return r.checkedDates, nil
}
func (r *fakeSubjectAttendanceRepo) SemesterRange(_ context.Context, _, _ string) (*time.Time, *time.Time, error) {
	return r.semStart, r.semEnd, nil
}

const saSchool = "school-A"
const saSlot = "slot1"
const saTeacherUser = "teacher-u"

func newSubjAttSvc() (*SubjectAttendanceService, *fakeSubjectAttendanceRepo, *fakeChecker) {
	repo := newFakeSubjectAttendanceRepo()
	repo.slots[saSlot] = slotCtx{classID: "c1", teacherUserID: saTeacherUser}
	repo.roster["c1"] = []domain.AttendanceRosterEntry{
		{StudentID: "s1"}, {StudentID: "s2"},
	}
	checker := &fakeChecker{groups: map[string]bool{}}
	return NewSubjectAttendanceService(repo, checker), repo, checker
}

func TestSubjectAttendance_AdminListsRoster(t *testing.T) {
	svc, _, _ := newSubjAttSvc()
	roster, err := svc.ListRoster(wAdmin(saSchool, "sem-1"), saSlot, "2026-06-26")
	if err != nil {
		t.Fatalf("list roster: %v", err)
	}
	if len(roster) != 2 {
		t.Errorf("roster = %d ควรเป็น 2", len(roster))
	}
}

func TestSubjectAttendance_SlotNotFound(t *testing.T) {
	svc, _, _ := newSubjAttSvc()
	_, err := svc.ListRoster(wAdmin(saSchool, "sem-1"), "missing", "2026-06-26")
	if !errors.Is(err, domain.ErrTimetableSlotNotFound) {
		t.Errorf("err = %v, want ErrTimetableSlotNotFound", err)
	}
}

func TestSubjectAttendance_ForbiddenForOtherTeacher(t *testing.T) {
	svc, _, _ := newSubjAttSvc()
	// ครูคนอื่นที่ไม่ใช่เจ้าของคาบ และไม่ได้อยู่กลุ่มวิชาการ
	err := svc.Save(wMember(saSchool, "other-teacher", "sem-1"), saSlot, "2026-06-26",
		[]AttendanceMarkInput{{StudentID: "s1", Status: domain.AttendancePresent}})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestSubjectAttendance_AssignedTeacherCanSave(t *testing.T) {
	svc, repo, _ := newSubjAttSvc()
	err := svc.Save(wMember(saSchool, saTeacherUser, "sem-1"), saSlot, "2026-06-26",
		[]AttendanceMarkInput{{StudentID: "s1", Status: domain.AttendancePresent}})
	if err != nil {
		t.Fatalf("teacher save: %v", err)
	}
	if repo.saveCalls != 1 {
		t.Errorf("saveCalls = %d ควรเป็น 1", repo.saveCalls)
	}
}

func TestSubjectAttendance_AcademicMemberCanSave(t *testing.T) {
	svc, _, checker := newSubjAttSvc()
	checker.groups[saSchool+"|office-u|academic"] = true
	err := svc.Save(wMember(saSchool, "office-u", "sem-1"), saSlot, "2026-06-26",
		[]AttendanceMarkInput{{StudentID: "s2", Status: domain.AttendanceLate}})
	if err != nil {
		t.Fatalf("academic save: %v", err)
	}
}

func TestSubjectAttendance_SaveInvalidStatus(t *testing.T) {
	svc, _, _ := newSubjAttSvc()
	err := svc.Save(wAdmin(saSchool, "sem-1"), saSlot, "2026-06-26",
		[]AttendanceMarkInput{{StudentID: "s1", Status: "bad"}})
	if !errors.Is(err, domain.ErrInvalidAttendanceStatus) {
		t.Errorf("err = %v, want ErrInvalidAttendanceStatus", err)
	}
}

func TestSubjectAttendance_SaveStudentNotInClass(t *testing.T) {
	svc, _, _ := newSubjAttSvc()
	err := svc.Save(wAdmin(saSchool, "sem-1"), saSlot, "2026-06-26",
		[]AttendanceMarkInput{{StudentID: "s999", Status: domain.AttendancePresent}})
	if !errors.Is(err, domain.ErrStudentNotInClass) {
		t.Errorf("err = %v, want ErrStudentNotInClass", err)
	}
}

func TestSubjectAttendance_SaveNoSemester(t *testing.T) {
	svc, _, _ := newSubjAttSvc()
	// admin ctx ไม่มี semester (authorize ผ่านเพราะ admin แต่ semesterOrErr ล้ม)
	err := svc.Save(adminCtx(saSchool), saSlot, "2026-06-26",
		[]AttendanceMarkInput{{StudentID: "s1", Status: domain.AttendancePresent}})
	if !errors.Is(err, domain.ErrNoActiveSemester) {
		t.Errorf("err = %v, want ErrNoActiveSemester", err)
	}
}

func TestSubjectAttendance_OverviewUncheckedAndWeeks(t *testing.T) {
	svc, repo, _ := newSubjAttSvc()
	repo.teacherSlots = []domain.TeacherCheckinSlot{
		{SlotID: "slot1", DayOfWeek: 1, PeriodNo: 1, SubjectCode: "ค21101", GradeLevel: "ม.1", RoomName: "1/1"},
	}
	start := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC) // จันทร์
	end := time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)   // ครอบ 2 สัปดาห์
	repo.semStart, repo.semEnd = &start, &end

	ov, err := svc.CheckinOverview(wAdmin(saSchool, "sem-1"), "2026-05-20") // พุธ → จันทร์ของสัปดาห์ = 2026-05-18
	if err != nil {
		t.Fatalf("overview: %v", err)
	}
	if len(ov.Slots) != 1 || ov.Slots[0].Date != ov.WeekStart {
		t.Errorf("คาบวันจันทร์ ควรมีวันที่ = ต้นสัปดาห์ (%s) ได้ %s", ov.WeekStart, ov.Slots[0].Date)
	}
	if ov.TotalThisWeek != 1 || ov.UncheckedThisWeek != 1 || ov.Slots[0].Checked {
		t.Errorf("ยังไม่เช็ค ควร unchecked=1 checked=false, got total=%d unchecked=%d checked=%v", ov.TotalThisWeek, ov.UncheckedThisWeek, ov.Slots[0].Checked)
	}
	if !ov.HasWeekStats || ov.TotalWeeks != 2 || ov.IncompleteWeeks != 2 {
		t.Errorf("ทุกสัปดาห์ยังไม่ครบ ควร incomplete=2/total=2, got %d/%d (hasStats=%v)", ov.IncompleteWeeks, ov.TotalWeeks, ov.HasWeekStats)
	}
	// รายการสัปดาห์: 18-31 พ.ค. = 2 สัปดาห์, วันที่ 20 อยู่สัปดาห์ที่ 1
	if len(ov.Weeks) != 2 || ov.CurrentWeekIndex != 1 || ov.Weeks[0].Start != "2026-05-18" {
		t.Errorf("weeks=%+v currentIndex=%d (ควร 2 สัปดาห์, current=1)", ov.Weeks, ov.CurrentWeekIndex)
	}
}

func TestSubjectAttendance_OverviewCheckedReducesIncomplete(t *testing.T) {
	svc, repo, _ := newSubjAttSvc()
	repo.teacherSlots = []domain.TeacherCheckinSlot{
		{SlotID: "slot1", DayOfWeek: 1, PeriodNo: 1, SubjectCode: "ค21101"},
	}
	start := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)
	repo.semStart, repo.semEnd = &start, &end
	// เช็คแล้วเฉพาะสัปดาห์แรก (จันทร์ 18)
	repo.checkedDates = []domain.SlotDate{{SlotID: "slot1", Date: start}}

	ov, err := svc.CheckinOverview(wAdmin(saSchool, "sem-1"), "2026-05-20")
	if err != nil {
		t.Fatalf("overview: %v", err)
	}
	if ov.UncheckedThisWeek != 0 || !ov.Slots[0].Checked {
		t.Errorf("สัปดาห์นี้เช็คแล้ว ควร unchecked=0 checked=true, got %d / %v", ov.UncheckedThisWeek, ov.Slots[0].Checked)
	}
	if ov.IncompleteWeeks != 1 {
		t.Errorf("เหลือสัปดาห์เดียวที่ยังไม่ครบ ควร incomplete=1, got %d", ov.IncompleteWeeks)
	}
}

func TestSubjectAttendance_SaveInvalidDate(t *testing.T) {
	svc, _, _ := newSubjAttSvc()
	err := svc.Save(wAdmin(saSchool, "sem-1"), saSlot, "bad-date",
		[]AttendanceMarkInput{{StudentID: "s1", Status: domain.AttendancePresent}})
	if !errors.Is(err, domain.ErrInvalidDate) {
		t.Errorf("err = %v, want ErrInvalidDate", err)
	}
}
