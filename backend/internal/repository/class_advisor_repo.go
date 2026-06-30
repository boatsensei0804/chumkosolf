package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumkosoft/backend/internal/domain"
)

// ClassAdvisorRepository — ครูที่ปรึกษาของห้อง (M:N, รายเทอม)
type ClassAdvisorRepository struct {
	db *pgxpool.Pool
}

func NewClassAdvisorRepository(db *pgxpool.Pool) *ClassAdvisorRepository {
	return &ClassAdvisorRepository{db: db}
}

func (r *ClassAdvisorRepository) ListByClass(ctx context.Context, schoolID, classID string) ([]domain.ClassAdvisor, error) {
	const q = `
		SELECT ca.id, ca.personnel_id, COALESCE(p.prefix, ''), p.first_name, p.last_name
		FROM class_advisors ca
		JOIN personnel p ON p.id = ca.personnel_id
		WHERE ca.school_id = $1 AND ca.class_id = $2 AND ca.deleted_at IS NULL AND p.deleted_at IS NULL
		ORDER BY p.first_name`
	rows, err := r.db.Query(ctx, q, schoolID, classID)
	if err != nil {
		return nil, fmt.Errorf("repository: list advisors: %w", err)
	}
	defer rows.Close()

	var out []domain.ClassAdvisor
	for rows.Next() {
		var a domain.ClassAdvisor
		if err := rows.Scan(&a.ID, &a.PersonnelID, &a.Prefix, &a.FirstName, &a.LastName); err != nil {
			return nil, fmt.Errorf("repository: scan advisor: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Add มอบหมายครูที่ปรึกษา (upsert: คืนสภาพถ้าเคยถอด) + audit
func (r *ClassAdvisorRepository) Add(ctx context.Context, schoolID, semesterID, classID, personnelID string, audit domain.AuditEntry) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO class_advisors (school_id, semester_id, class_id, personnel_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (class_id, personnel_id) DO UPDATE SET deleted_at = NULL, updated_at = now()`
	if _, err := tx.Exec(ctx, q, schoolID, semesterID, classID, personnelID); err != nil {
		return fmt.Errorf("repository: add advisor: %w", err)
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repository: commit add advisor: %w", err)
	}
	return nil
}

// Remove ถอดครูที่ปรึกษา (soft delete ด้วย advisor id) + audit
func (r *ClassAdvisorRepository) Remove(ctx context.Context, schoolID, classID, advisorID string, audit domain.AuditEntry) (bool, error) {
	return softDeleteWithAudit(ctx, r.db,
		`UPDATE class_advisors SET deleted_at = now(), updated_at = now()
		 WHERE id = $3 AND class_id = $2 AND school_id = $1 AND deleted_at IS NULL`,
		schoolID, classID, advisorID, audit)
}
