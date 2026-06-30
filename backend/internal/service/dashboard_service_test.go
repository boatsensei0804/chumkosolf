package service

import (
	"context"
	"testing"
	"time"

	"github.com/chumkosoft/backend/internal/domain"
)

type fakeDashboardRepo struct {
	classCount   int
	studentCount int
	counts       []domain.StatusCount
	slots        []domain.TeacherCheckinSlot
}

func (r *fakeDashboardRepo) AdvisorSummary(_ context.Context, _, _, _ string) (int, int, error) {
	return r.classCount, r.studentCount, nil
}
func (r *fakeDashboardRepo) TodayAttendanceCounts(_ context.Context, _, _, _ string, _ time.Time) ([]domain.StatusCount, error) {
	return r.counts, nil
}
func (r *fakeDashboardRepo) TeacherSlots(_ context.Context, _, _, _ string) ([]domain.TeacherCheckinSlot, error) {
	return r.slots, nil
}

func TestDashboard_Summary(t *testing.T) {
	repo := &fakeDashboardRepo{
		classCount: 1, studentCount: 6,
		counts: []domain.StatusCount{
			{Status: domain.AttendancePresent, Count: 4},
			{Status: domain.AttendanceAbsent, Count: 1},
			{Status: "", Count: 1}, // ยังไม่เช็ค
		},
		slots: []domain.TeacherCheckinSlot{{DayOfWeek: 1, PeriodNo: 1, SubjectCode: "ค21101", GradeLevel: "ม.1", RoomName: "1/2"}},
	}
	svc := NewDashboardService(repo)
	svc.now = func() time.Time { return time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC) } // จันทร์

	d, err := svc.Summary(wAdmin("school-A", "sem-1"))
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if !d.IsAdvisor || d.AdviseeCount != 6 {
		t.Errorf("advisor=%v advisee=%d ควร true/6", d.IsAdvisor, d.AdviseeCount)
	}
	if d.Attendance.Present != 4 || d.Attendance.Absent != 1 || d.Attendance.Unchecked != 1 || d.Attendance.Total != 6 {
		t.Errorf("attendance = %+v", d.Attendance)
	}
	if d.TodayWeekday != 1 || len(d.Slots) != 1 || d.Slots[0].ClassLabel != "ม.1 1/2" {
		t.Errorf("weekday=%d slots=%+v", d.TodayWeekday, d.Slots)
	}
}

func TestDashboard_NoSemester(t *testing.T) {
	svc := NewDashboardService(&fakeDashboardRepo{classCount: 1, studentCount: 6})
	// ไม่มี semester ใน context → คืนค่าว่าง ไม่ error
	d, err := svc.Summary(adminCtx("school-A"))
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if d.IsAdvisor || d.AdviseeCount != 0 {
		t.Errorf("ไม่มีเทอม ควรว่าง ได้ advisor=%v advisee=%d", d.IsAdvisor, d.AdviseeCount)
	}
}
