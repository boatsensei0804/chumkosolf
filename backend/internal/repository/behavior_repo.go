package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// BehaviorRepository — คะแนนความประพฤติ (ประวัติหัก/เพิ่ม, รายเทอม)
type BehaviorRepository struct {
	db *pgxpool.Pool
}

func NewBehaviorRepository(db *pgxpool.Pool) *BehaviorRepository {
	return &BehaviorRepository{db: db}
}

// ListByStudent คืนประวัติคะแนนของนักเรียนในเทอม (ใหม่สุดก่อน)
func (r *BehaviorRepository) ListByStudent(ctx context.Context, schoolID, semesterID, studentID string) ([]domain.BehaviorRecord, error) {
	const q = `
		SELECT id, points, reason, occurred_at, created_at, updated_at
		FROM behavior_records
		WHERE school_id = $1 AND semester_id = $2 AND student_id = $3 AND deleted_at IS NULL
		ORDER BY COALESCE(occurred_at, created_at::date) DESC, created_at DESC`
	rows, err := r.db.Query(ctx, q, schoolID, semesterID, studentID)
	if err != nil {
		return nil, fmt.Errorf("repository: list behavior records: %w", err)
	}
	defer rows.Close()

	var out []domain.BehaviorRecord
	for rows.Next() {
		var b domain.BehaviorRecord
		b.SchoolID = schoolID
		b.SemesterID = semesterID
		b.StudentID = studentID
		if err := rows.Scan(&b.ID, &b.Points, &b.Reason, &b.OccurredAt, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan behavior record: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// Create เพิ่มรายการคะแนน (เซ็ต school_id + semester_id จาก context) + audit
func (r *BehaviorRepository) Create(ctx context.Context, schoolID, semesterID, studentID string, nr domain.NewBehaviorRecord, recordedBy string, audit domain.AuditEntry) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO behavior_records (school_id, semester_id, student_id, points, reason, recorded_by, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	var id string
	if err := tx.QueryRow(ctx, q, schoolID, semesterID, studentID, nr.Points, nr.Reason, recordedBy, nr.OccurredAt).Scan(&id); err != nil {
		return "", fmt.Errorf("repository: insert behavior record: %w", err)
	}
	audit.TargetID = id
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("repository: commit behavior record: %w", err)
	}
	return id, nil
}

// SoftDelete ลบรายการคะแนน (scope school_id + student_id) + audit
func (r *BehaviorRepository) SoftDelete(ctx context.Context, schoolID, studentID, id string, audit domain.AuditEntry) (bool, error) {
	return softDeleteWithAudit(ctx, r.db,
		`UPDATE behavior_records SET deleted_at = now(), updated_at = now()
		 WHERE id = $3 AND student_id = $2 AND school_id = $1 AND deleted_at IS NULL`,
		schoolID, studentID, id, audit)
}
