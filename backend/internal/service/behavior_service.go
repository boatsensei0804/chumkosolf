package service

import (
	"context"
	"strings"
	"time"

	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/tenant"
)

// BehaviorRepository contract ของชั้น DB
type BehaviorRepository interface {
	ListByStudent(ctx context.Context, schoolID, semesterID, studentID string) ([]domain.BehaviorRecord, error)
	Create(ctx context.Context, schoolID, semesterID, studentID string, nr domain.NewBehaviorRecord, recordedBy string, audit domain.AuditEntry) (string, error)
	SoftDelete(ctx context.Context, schoolID, studentID, id string, audit domain.AuditEntry) (bool, error)
}

// behaviorStudentRepo ใช้ยืนยันว่ามีนักเรียนจริงในโรงเรียน (scope)
type behaviorStudentRepo interface {
	GetByID(ctx context.Context, schoolID, id string) (*domain.Student, error)
}

// BehaviorRecordDTO รายการคะแนนสำหรับ response
type BehaviorRecordDTO struct {
	ID         string `json:"id"`
	Points     int    `json:"points"`
	Reason     string `json:"reason"`
	OccurredAt string `json:"occurred_at"`
	CreatedAt  string `json:"created_at"`
}

// BehaviorSummaryDTO สรุปคะแนนความประพฤติ + ประวัติ
type BehaviorSummaryDTO struct {
	StartingScore int                 `json:"starting_score"`
	CurrentScore  int                 `json:"current_score"`
	Records       []BehaviorRecordDTO `json:"records"`
}

// BehaviorInput ข้อมูลเพิ่มรายการคะแนน
type BehaviorInput struct {
	Points     int
	Reason     string
	OccurredAt *time.Time
}

// BehaviorService จัดการคะแนนความประพฤติ (กลุ่มบริหารทั่วไป + admin)
type BehaviorService struct {
	repo     BehaviorRepository
	students behaviorStudentRepo
	checker  WorkGroupChecker
}

func NewBehaviorService(repo BehaviorRepository, students behaviorStudentRepo, checker WorkGroupChecker) *BehaviorService {
	return &BehaviorService{repo: repo, students: students, checker: checker}
}

// authorizeStudent: นักเรียนต้องมีจริง (scope) และผู้ใช้ต้องเป็น admin หรือกลุ่มบริหารทั่วไป
func (s *BehaviorService) authorizeStudent(ctx context.Context, studentID string) error {
	schoolID := tenant.SchoolIDFromContext(ctx)
	if !tenant.IsSchoolAdminFromContext(ctx) {
		ok, err := s.checker.IsUserInWorkGroup(ctx, schoolID, tenant.UserIDFromContext(ctx), generalAffairsWorkGroupCode)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrForbidden
		}
	}
	st, err := s.students.GetByID(ctx, schoolID, studentID)
	if err != nil {
		return err
	}
	if st == nil {
		return domain.ErrStudentNotFound
	}
	return nil
}

// Summary คืนคะแนนปัจจุบัน (ตั้งต้น + SUM) + ประวัติของเทอมปัจจุบัน
func (s *BehaviorService) Summary(ctx context.Context, studentID string) (*BehaviorSummaryDTO, error) {
	if err := s.authorizeStudent(ctx, studentID); err != nil {
		return nil, err
	}
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListByStudent(ctx, tenant.SchoolIDFromContext(ctx), sem, studentID)
	if err != nil {
		return nil, err
	}
	sum := 0
	records := make([]BehaviorRecordDTO, 0, len(rows))
	for i := range rows {
		sum += rows[i].Points
		records = append(records, BehaviorRecordDTO{
			ID:         rows[i].ID,
			Points:     rows[i].Points,
			Reason:     rows[i].Reason,
			OccurredAt: dateStr(rows[i].OccurredAt),
			CreatedAt:  rows[i].CreatedAt.Format(time.RFC3339),
		})
	}
	return &BehaviorSummaryDTO{
		StartingScore: domain.DefaultBehaviorScore,
		CurrentScore:  domain.DefaultBehaviorScore + sum,
		Records:       records,
	}, nil
}

// Create เพิ่มรายการหัก/เพิ่มคะแนน
func (s *BehaviorService) Create(ctx context.Context, studentID string, in BehaviorInput) (string, error) {
	if err := s.authorizeStudent(ctx, studentID); err != nil {
		return "", err
	}
	sem, err := semesterOrErr(ctx)
	if err != nil {
		return "", err
	}
	if in.Points == 0 {
		return "", domain.ErrInvalidPoints
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		return "", domain.ErrReasonRequired
	}
	audit := auditFor(ctx, domain.AuditCreate, "behavior_record", "", map[string]any{"student_id": studentID, "points": in.Points})
	return s.repo.Create(ctx, tenant.SchoolIDFromContext(ctx), sem, studentID, domain.NewBehaviorRecord{
		Points:     in.Points,
		Reason:     reason,
		OccurredAt: in.OccurredAt,
	}, tenant.UserIDFromContext(ctx), audit)
}

// Delete ลบรายการคะแนน
func (s *BehaviorService) Delete(ctx context.Context, studentID, id string) error {
	if err := s.authorizeStudent(ctx, studentID); err != nil {
		return err
	}
	audit := auditFor(ctx, domain.AuditDelete, "behavior_record", id, map[string]any{"student_id": studentID})
	found, err := s.repo.SoftDelete(ctx, tenant.SchoolIDFromContext(ctx), studentID, id, audit)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrBehaviorNotFound
	}
	return nil
}
