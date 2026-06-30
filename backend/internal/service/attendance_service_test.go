package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chumkosoft/backend/internal/domain"
)

// --- fakes ---

type fakeAttendanceRepo struct {
	roster    map[string][]domain.AttendanceRosterEntry // key: classID
	advisors  map[string]bool                           // key: schoolID|classID|userID
	saved     []domain.AttendanceMark
	saveCalls int
}

func newFakeAttendanceRepo() *fakeAttendanceRepo {
	return &fakeAttendanceRepo{roster: map[string][]domain.AttendanceRosterEntry{}, advisors: map[string]bool{}}
}

func (r *fakeAttendanceRepo) RosterByClassDate(_ context.Context, _, classID string, _ time.Time) ([]domain.AttendanceRosterEntry, error) {
	return r.roster[classID], nil
}

func (r *fakeAttendanceRepo) BulkUpsert(_ context.Context, _, _, _ string, _ time.Time, marks []domain.AttendanceMark, _ string, _ domain.AuditEntry) error {
	r.saved = append(r.saved, marks...)
	r.saveCalls++
	return nil
}

func (r *fakeAttendanceRepo) IsClassAdvisorUser(_ context.Context, schoolID, classID, userID string) (bool, error) {
	return r.advisors[schoolID+"|"+classID+"|"+userID], nil
}

type fakeClassExists struct {
	byID map[string]*domain.Class
}

func (r *fakeClassExists) GetByID(_ context.Context, schoolID, id string) (*domain.Class, error) {
	c, ok := r.byID[id]
	if !ok || c.SchoolID != schoolID {
		return nil, nil
	}
	return c, nil
}

type fakeChecker struct {
	groups map[string]bool // key: schoolID|userID|code
}

func (c *fakeChecker) IsUserInWorkGroup(_ context.Context, schoolID, userID, code string) (bool, error) {
	return c.groups[schoolID+"|"+userID+"|"+code], nil
}

const attSchool = "school-A"
const attClass = "c1"

func newAttendanceSvc() (*AttendanceService, *fakeAttendanceRepo, *fakeChecker) {
	repo := newFakeAttendanceRepo()
	repo.roster[attClass] = []domain.AttendanceRosterEntry{
		{StudentID: "s1", FirstName: "ก", LastName: "ข"},
		{StudentID: "s2", FirstName: "ค", LastName: "ง"},
	}
	classes := &fakeClassExists{byID: map[string]*domain.Class{
		attClass: {ID: attClass, SchoolID: attSchool, SemesterID: "sem-1"},
	}}
	checker := &fakeChecker{groups: map[string]bool{}}
	return NewAttendanceService(repo, classes, checker), repo, checker
}

func presentMark(studentID string) AttendanceMarkInput {
	return AttendanceMarkInput{StudentID: studentID, Status: domain.AttendancePresent}
}

// --- tests ---

func TestAttendance_AdminListsRoster(t *testing.T) {
	svc, _, _ := newAttendanceSvc()
	roster, err := svc.ListRoster(wAdmin(attSchool, "sem-1"), attClass, "2026-06-26")
	if err != nil {
		t.Fatalf("list roster: %v", err)
	}
	if len(roster) != 2 {
		t.Errorf("roster = %d รายการ ควรเป็น 2", len(roster))
	}
}

func TestAttendance_ClassNotFound(t *testing.T) {
	svc, _, _ := newAttendanceSvc()
	_, err := svc.ListRoster(wAdmin(attSchool, "sem-1"), "missing", "2026-06-26")
	if !errors.Is(err, domain.ErrClassNotFound) {
		t.Errorf("err = %v, want ErrClassNotFound", err)
	}
}

func TestAttendance_CrossSchoolClassNotFound(t *testing.T) {
	svc, _, _ := newAttendanceSvc()
	// ห้องอยู่ school-A แต่เรียกด้วย scope school-B → ไม่พบ (isolation)
	_, err := svc.ListRoster(wAdmin("school-B", "sem-1"), attClass, "2026-06-26")
	if !errors.Is(err, domain.ErrClassNotFound) {
		t.Errorf("err = %v, want ErrClassNotFound", err)
	}
}

func TestAttendance_ForbiddenForOutsider(t *testing.T) {
	svc, _, _ := newAttendanceSvc()
	// teacher ที่ไม่ได้อยู่กลุ่มบริหารทั่วไป และไม่ใช่ครูที่ปรึกษาห้องนี้
	err := svc.Save(wMember(attSchool, "u9", "sem-1"), attClass, "2026-06-26", []AttendanceMarkInput{presentMark("s1")})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestAttendance_GeneralAffairsMemberCanSave(t *testing.T) {
	svc, repo, checker := newAttendanceSvc()
	checker.groups[attSchool+"|u9|general_affairs"] = true
	err := svc.Save(wMember(attSchool, "u9", "sem-1"), attClass, "2026-06-26", []AttendanceMarkInput{presentMark("s1"), presentMark("s2")})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if repo.saveCalls != 1 || len(repo.saved) != 2 {
		t.Errorf("saveCalls=%d saved=%d ควร 1 และ 2", repo.saveCalls, len(repo.saved))
	}
}

func TestAttendance_AdvisorCanSave(t *testing.T) {
	svc, repo, _ := newAttendanceSvc()
	repo.advisors[attSchool+"|"+attClass+"|u9"] = true
	err := svc.Save(wMember(attSchool, "u9", "sem-1"), attClass, "2026-06-26", []AttendanceMarkInput{presentMark("s1")})
	if err != nil {
		t.Fatalf("advisor save: %v", err)
	}
}

func TestAttendance_SaveInvalidStatus(t *testing.T) {
	svc, _, _ := newAttendanceSvc()
	err := svc.Save(wAdmin(attSchool, "sem-1"), attClass, "2026-06-26",
		[]AttendanceMarkInput{{StudentID: "s1", Status: "skipped"}})
	if !errors.Is(err, domain.ErrInvalidAttendanceStatus) {
		t.Errorf("err = %v, want ErrInvalidAttendanceStatus", err)
	}
}

func TestAttendance_SaveStudentNotInClass(t *testing.T) {
	svc, _, _ := newAttendanceSvc()
	err := svc.Save(wAdmin(attSchool, "sem-1"), attClass, "2026-06-26",
		[]AttendanceMarkInput{presentMark("s999")})
	if !errors.Is(err, domain.ErrStudentNotInClass) {
		t.Errorf("err = %v, want ErrStudentNotInClass", err)
	}
}

func TestAttendance_SaveNoSemester(t *testing.T) {
	svc, _, _ := newAttendanceSvc()
	// adminCtx ไม่มี SemesterID
	err := svc.Save(adminCtx(attSchool), attClass, "2026-06-26", []AttendanceMarkInput{presentMark("s1")})
	if !errors.Is(err, domain.ErrNoActiveSemester) {
		t.Errorf("err = %v, want ErrNoActiveSemester", err)
	}
}

func TestAttendance_SaveInvalidDate(t *testing.T) {
	svc, _, _ := newAttendanceSvc()
	err := svc.Save(wAdmin(attSchool, "sem-1"), attClass, "26/06/2026", []AttendanceMarkInput{presentMark("s1")})
	if !errors.Is(err, domain.ErrInvalidDate) {
		t.Errorf("err = %v, want ErrInvalidDate", err)
	}
}
