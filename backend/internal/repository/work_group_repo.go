package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// WorkGroupRepository เข้าถึงตาราง work_groups + user_work_groups (scope school_id)
type WorkGroupRepository struct {
	db *pgxpool.Pool
}

func NewWorkGroupRepository(db *pgxpool.Pool) *WorkGroupRepository {
	return &WorkGroupRepository{db: db}
}

// ListBySchool คืนกลุ่มงานทั้งหมดของโรงเรียน
func (r *WorkGroupRepository) ListBySchool(ctx context.Context, schoolID string) ([]domain.WorkGroup, error) {
	const q = `SELECT id, code, name FROM work_groups WHERE school_id = $1 AND deleted_at IS NULL ORDER BY code`
	rows, err := r.db.Query(ctx, q, schoolID)
	if err != nil {
		return nil, fmt.Errorf("repository: list work groups: %w", err)
	}
	defer rows.Close()

	var out []domain.WorkGroup
	for rows.Next() {
		var wg domain.WorkGroup
		if err := rows.Scan(&wg.ID, &wg.Code, &wg.Name); err != nil {
			return nil, fmt.Errorf("repository: scan work group: %w", err)
		}
		out = append(out, wg)
	}
	return out, rows.Err()
}

// ExistsInSchool ตรวจว่ากลุ่มงานนี้อยู่ในโรงเรียนจริง
func (r *WorkGroupRepository) ExistsInSchool(ctx context.Context, schoolID, workGroupID string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM work_groups WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL)`
	var exists bool
	if err := r.db.QueryRow(ctx, q, schoolID, workGroupID).Scan(&exists); err != nil {
		return false, fmt.Errorf("repository: work group exists: %w", err)
	}
	return exists, nil
}

// ListForUser คืนกลุ่มงานที่ user สังกัด (พร้อมสถานะหัวหน้ากลุ่ม)
func (r *WorkGroupRepository) ListForUser(ctx context.Context, schoolID, userID string) ([]domain.WorkGroupMembership, error) {
	const q = `
		SELECT wg.id, wg.code, wg.name, uwg.is_group_admin
		FROM user_work_groups uwg
		JOIN work_groups wg ON wg.id = uwg.work_group_id
		WHERE uwg.school_id = $1 AND uwg.user_id = $2
		  AND uwg.deleted_at IS NULL AND wg.deleted_at IS NULL
		ORDER BY wg.code`
	rows, err := r.db.Query(ctx, q, schoolID, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: list user work groups: %w", err)
	}
	defer rows.Close()

	var out []domain.WorkGroupMembership
	for rows.Next() {
		var m domain.WorkGroupMembership
		if err := rows.Scan(&m.WorkGroupID, &m.Code, &m.Name, &m.IsGroupAdmin); err != nil {
			return nil, fmt.Errorf("repository: scan membership: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// IsGroupAdmin ตรวจว่า user เป็นหัวหน้ากลุ่มงานนี้ไหม
func (r *WorkGroupRepository) IsGroupAdmin(ctx context.Context, schoolID, userID, workGroupID string) (bool, error) {
	const q = `
		SELECT EXISTS(
			SELECT 1 FROM user_work_groups
			WHERE school_id = $1 AND user_id = $2 AND work_group_id = $3
			  AND is_group_admin = TRUE AND deleted_at IS NULL
		)`
	var ok bool
	if err := r.db.QueryRow(ctx, q, schoolID, userID, workGroupID).Scan(&ok); err != nil {
		return false, fmt.Errorf("repository: is group admin: %w", err)
	}
	return ok, nil
}

// Assign มอบหมาย user เข้ากลุ่มงาน (upsert: ถ้าเคยถอดไว้จะ activate ใหม่ + อัปเดต is_group_admin) + audit
func (r *WorkGroupRepository) Assign(ctx context.Context, schoolID, userID, workGroupID string, isGroupAdmin bool, audit domain.AuditEntry) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO user_work_groups (school_id, user_id, work_group_id, is_group_admin)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, work_group_id) DO UPDATE
			SET is_group_admin = EXCLUDED.is_group_admin, deleted_at = NULL, updated_at = now()`
	if _, err := tx.Exec(ctx, q, schoolID, userID, workGroupID, isGroupAdmin); err != nil {
		return fmt.Errorf("repository: assign work group: %w", err)
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repository: commit assign: %w", err)
	}
	return nil
}

// Unassign ถอด user ออกจากกลุ่มงาน (soft delete) + audit; คืน found=false ถ้าไม่ได้สังกัด
func (r *WorkGroupRepository) Unassign(ctx context.Context, schoolID, userID, workGroupID string, audit domain.AuditEntry) (bool, error) {
	return softDeleteWithAudit(ctx, r.db,
		`UPDATE user_work_groups SET deleted_at = now(), updated_at = now()
		 WHERE school_id = $1 AND user_id = $2 AND work_group_id = $3 AND deleted_at IS NULL`,
		schoolID, userID, workGroupID, audit)
}
