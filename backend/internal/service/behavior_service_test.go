package service

import (
	"context"
	"errors"
	"testing"

	"github.com/chumko-platform/backend/internal/domain"
)

// --- fakes ---

type fakeBehaviorRepo struct {
	items map[string]*domain.BehaviorRecord
	seq   int
}

func newFakeBehaviorRepo() *fakeBehaviorRepo {
	return &fakeBehaviorRepo{items: map[string]*domain.BehaviorRecord{}}
}

func (r *fakeBehaviorRepo) ListByStudent(_ context.Context, schoolID, semesterID, studentID string) ([]domain.BehaviorRecord, error) {
	var out []domain.BehaviorRecord
	for _, b := range r.items {
		if b.SchoolID == schoolID && b.SemesterID == semesterID && b.StudentID == studentID {
			out = append(out, *b)
		}
	}
	return out, nil
}

func (r *fakeBehaviorRepo) Create(_ context.Context, schoolID, semesterID, studentID string, nr domain.NewBehaviorRecord, _ string, _ domain.AuditEntry) (string, error) {
	r.seq++
	id := "b" + string(rune('0'+r.seq))
	r.items[id] = &domain.BehaviorRecord{
		ID: id, SchoolID: schoolID, SemesterID: semesterID, StudentID: studentID,
		Points: nr.Points, Reason: nr.Reason, OccurredAt: nr.OccurredAt,
	}
	return id, nil
}

func (r *fakeBehaviorRepo) SoftDelete(_ context.Context, schoolID, studentID, id string, _ domain.AuditEntry) (bool, error) {
	b, ok := r.items[id]
	if !ok || b.SchoolID != schoolID || b.StudentID != studentID {
		return false, nil
	}
	delete(r.items, id)
	return true, nil
}

type fakeBehaviorStudentRepo struct {
	byID map[string]*domain.Student
}

func (r *fakeBehaviorStudentRepo) GetByID(_ context.Context, schoolID, id string) (*domain.Student, error) {
	st, ok := r.byID[id]
	if !ok || st.SchoolID != schoolID {
		return nil, nil
	}
	return st, nil
}

const behSchool = "school-A"
const behStudent = "stu-1"

func newBehaviorSvc() (*BehaviorService, *fakeBehaviorRepo, *fakeChecker) {
	repo := newFakeBehaviorRepo()
	students := &fakeBehaviorStudentRepo{byID: map[string]*domain.Student{
		behStudent: {ID: behStudent, SchoolID: behSchool},
	}}
	checker := &fakeChecker{groups: map[string]bool{}}
	return NewBehaviorService(repo, students, checker), repo, checker
}

// --- tests ---

func TestBehavior_SummaryStartsAtDefault(t *testing.T) {
	svc, _, _ := newBehaviorSvc()
	sum, err := svc.Summary(wAdmin(behSchool, "sem-1"), behStudent)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if sum.CurrentScore != domain.DefaultBehaviorScore {
		t.Errorf("current = %d ควรเท่าตั้งต้น %d", sum.CurrentScore, domain.DefaultBehaviorScore)
	}
}

func TestBehavior_CurrentScoreIsStartPlusSum(t *testing.T) {
	svc, _, _ := newBehaviorSvc()
	ctx := wAdmin(behSchool, "sem-1")
	if _, err := svc.Create(ctx, behStudent, BehaviorInput{Points: -10, Reason: "มาสาย"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.Create(ctx, behStudent, BehaviorInput{Points: 5, Reason: "ช่วยงาน"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	sum, err := svc.Summary(ctx, behStudent)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	want := domain.DefaultBehaviorScore - 10 + 5
	if sum.CurrentScore != want {
		t.Errorf("current = %d ควรเป็น %d", sum.CurrentScore, want)
	}
	if len(sum.Records) != 2 {
		t.Errorf("records = %d ควรเป็น 2", len(sum.Records))
	}
}

func TestBehavior_CreateZeroPoints(t *testing.T) {
	svc, _, _ := newBehaviorSvc()
	_, err := svc.Create(wAdmin(behSchool, "sem-1"), behStudent, BehaviorInput{Points: 0, Reason: "x"})
	if !errors.Is(err, domain.ErrInvalidPoints) {
		t.Errorf("err = %v, want ErrInvalidPoints", err)
	}
}

func TestBehavior_CreateRequiresReason(t *testing.T) {
	svc, _, _ := newBehaviorSvc()
	_, err := svc.Create(wAdmin(behSchool, "sem-1"), behStudent, BehaviorInput{Points: -5, Reason: "  "})
	if !errors.Is(err, domain.ErrReasonRequired) {
		t.Errorf("err = %v, want ErrReasonRequired", err)
	}
}

func TestBehavior_ForbiddenForNonMember(t *testing.T) {
	svc, _, _ := newBehaviorSvc()
	_, err := svc.Create(wMember(behSchool, "u9", "sem-1"), behStudent, BehaviorInput{Points: -5, Reason: "มาสาย"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestBehavior_GeneralAffairsMemberCanCreate(t *testing.T) {
	svc, _, checker := newBehaviorSvc()
	checker.groups[behSchool+"|u9|general_affairs"] = true
	id, err := svc.Create(wMember(behSchool, "u9", "sem-1"), behStudent, BehaviorInput{Points: -5, Reason: "มาสาย"})
	if err != nil || id == "" {
		t.Fatalf("create by group member: id=%q err=%v", id, err)
	}
}

func TestBehavior_CrossSchoolStudentNotFound(t *testing.T) {
	svc, _, _ := newBehaviorSvc()
	// นักเรียนอยู่ school-A แต่เรียกด้วย scope school-B → ไม่พบ (isolation)
	_, err := svc.Create(wAdmin("school-B", "sem-1"), behStudent, BehaviorInput{Points: -5, Reason: "x"})
	if !errors.Is(err, domain.ErrStudentNotFound) {
		t.Errorf("err = %v, want ErrStudentNotFound", err)
	}
}

func TestBehavior_SummaryIsolatedBySemester(t *testing.T) {
	svc, _, _ := newBehaviorSvc()
	if _, err := svc.Create(wAdmin(behSchool, "sem-1"), behStudent, BehaviorInput{Points: -10, Reason: "x"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	// เทอม 2 ต้องไม่เห็นรายการของเทอม 1 → คะแนนกลับไปตั้งต้น
	sum, err := svc.Summary(wAdmin(behSchool, "sem-2"), behStudent)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if sum.CurrentScore != domain.DefaultBehaviorScore || len(sum.Records) != 0 {
		t.Errorf("เทอม 2 current=%d records=%d ควรเป็นตั้งต้น/ว่าง (ข้อมูลปนข้ามเทอม)", sum.CurrentScore, len(sum.Records))
	}
}

func TestBehavior_DeleteNotFound(t *testing.T) {
	svc, _, _ := newBehaviorSvc()
	if err := svc.Delete(wAdmin(behSchool, "sem-1"), behStudent, "missing"); !errors.Is(err, domain.ErrBehaviorNotFound) {
		t.Errorf("err = %v, want ErrBehaviorNotFound", err)
	}
}
