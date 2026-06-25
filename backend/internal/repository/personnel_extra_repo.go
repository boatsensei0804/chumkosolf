package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// --- Admin positions ---

// AdminPositionRepository เข้าถึงตาราง admin_positions (scope school_id + personnel_id)
type AdminPositionRepository struct {
	db *pgxpool.Pool
}

func NewAdminPositionRepository(db *pgxpool.Pool) *AdminPositionRepository {
	return &AdminPositionRepository{db: db}
}

// ListByPersonnel คืนตำแหน่งบริหารของบุคลากร
func (r *AdminPositionRepository) ListByPersonnel(ctx context.Context, schoolID, personnelID string) ([]domain.AdminPosition, error) {
	const q = `
		SELECT id, personnel_id, position, is_active, appointed_at, created_at, updated_at
		FROM admin_positions
		WHERE school_id = $1 AND personnel_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, schoolID, personnelID)
	if err != nil {
		return nil, fmt.Errorf("repository: list admin positions: %w", err)
	}
	defer rows.Close()

	var out []domain.AdminPosition
	for rows.Next() {
		var p domain.AdminPosition
		p.SchoolID = schoolID
		if err := rows.Scan(&p.ID, &p.PersonnelID, &p.Position, &p.IsActive, &p.AppointedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan admin position: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Create เพิ่มตำแหน่งบริหาร + audit; ผอ. ที่ active ซ้ำ → ErrDuplicateDirector
func (r *AdminPositionRepository) Create(ctx context.Context, schoolID, personnelID string, np domain.NewAdminPosition, audit domain.AuditEntry) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO admin_positions (school_id, personnel_id, position, is_active, appointed_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	var id string
	err = tx.QueryRow(ctx, q, schoolID, personnelID, np.Position, np.IsActive, np.AppointedAt).Scan(&id)
	if err != nil {
		if isUniqueViolation(err, "active_director") {
			return "", domain.ErrDuplicateDirector
		}
		return "", fmt.Errorf("repository: insert admin position: %w", err)
	}

	audit.TargetID = id
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("repository: commit admin position: %w", err)
	}
	return id, nil
}

// SoftDelete ลบตำแหน่งบริหาร (scope school_id + personnel_id) + audit
func (r *AdminPositionRepository) SoftDelete(ctx context.Context, schoolID, personnelID, id string, audit domain.AuditEntry) (bool, error) {
	return softDeleteWithAudit(ctx, r.db,
		`UPDATE admin_positions SET deleted_at = now(), updated_at = now()
		 WHERE id = $3 AND personnel_id = $2 AND school_id = $1 AND deleted_at IS NULL`,
		schoolID, personnelID, id, audit)
}

// --- Academic standings ---

// AcademicStandingRepository เข้าถึงตาราง academic_standings (scope school_id + personnel_id)
type AcademicStandingRepository struct {
	db *pgxpool.Pool
}

func NewAcademicStandingRepository(db *pgxpool.Pool) *AcademicStandingRepository {
	return &AcademicStandingRepository{db: db}
}

// ListByPersonnel คืนประวัติวิทยฐานะ (ใหม่สุดก่อน)
func (r *AcademicStandingRepository) ListByPersonnel(ctx context.Context, schoolID, personnelID string) ([]domain.AcademicStanding, error) {
	const q = `
		SELECT id, personnel_id, standing, effective_date, is_current, created_at, updated_at
		FROM academic_standings
		WHERE school_id = $1 AND personnel_id = $2 AND deleted_at IS NULL
		ORDER BY is_current DESC, effective_date DESC NULLS LAST, created_at DESC`
	rows, err := r.db.Query(ctx, q, schoolID, personnelID)
	if err != nil {
		return nil, fmt.Errorf("repository: list standings: %w", err)
	}
	defer rows.Close()

	var out []domain.AcademicStanding
	for rows.Next() {
		var s domain.AcademicStanding
		s.SchoolID = schoolID
		if err := rows.Scan(&s.ID, &s.PersonnelID, &s.Standing, &s.EffectiveDate, &s.IsCurrent, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan standing: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// Create เพิ่มวิทยฐานะ + audit; ถ้า is_current=true จะ unset อันอื่นก่อน (กันชน unique index)
func (r *AcademicStandingRepository) Create(ctx context.Context, schoolID, personnelID string, ns domain.NewAcademicStanding, audit domain.AuditEntry) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if ns.IsCurrent {
		if err := unsetCurrentStandings(ctx, tx, schoolID, personnelID, ""); err != nil {
			return "", err
		}
	}

	const q = `
		INSERT INTO academic_standings (school_id, personnel_id, standing, effective_date, is_current)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	var id string
	if err := tx.QueryRow(ctx, q, schoolID, personnelID, ns.Standing, ns.EffectiveDate, ns.IsCurrent).Scan(&id); err != nil {
		return "", fmt.Errorf("repository: insert standing: %w", err)
	}

	audit.TargetID = id
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("repository: commit standing: %w", err)
	}
	return id, nil
}

// Update แก้วิทยฐานะ; ถ้า is_current=true จะ unset อันอื่นก่อน
func (r *AcademicStandingRepository) Update(ctx context.Context, schoolID, personnelID, id string, us domain.UpdateAcademicStanding, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if us.IsCurrent {
		if err := unsetCurrentStandings(ctx, tx, schoolID, personnelID, id); err != nil {
			return false, err
		}
	}

	const q = `
		UPDATE academic_standings
		SET standing = $4, effective_date = $5, is_current = $6, updated_at = now()
		WHERE id = $3 AND personnel_id = $2 AND school_id = $1 AND deleted_at IS NULL`
	tag, err := tx.Exec(ctx, q, schoolID, personnelID, id, us.Standing, us.EffectiveDate, us.IsCurrent)
	if err != nil {
		return false, fmt.Errorf("repository: update standing: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}

	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit update standing: %w", err)
	}
	return true, nil
}

// SoftDelete ลบวิทยฐานะ + audit
func (r *AcademicStandingRepository) SoftDelete(ctx context.Context, schoolID, personnelID, id string, audit domain.AuditEntry) (bool, error) {
	return softDeleteWithAudit(ctx, r.db,
		`UPDATE academic_standings SET deleted_at = now(), updated_at = now()
		 WHERE id = $3 AND personnel_id = $2 AND school_id = $1 AND deleted_at IS NULL`,
		schoolID, personnelID, id, audit)
}

// --- shared helpers ---

// unsetCurrentStandings ตั้ง is_current=false ของวิทยฐานะอื่นทั้งหมด (ยกเว้น excludeID)
// excludeID เป็น "" ได้ (ตอน create) — ใช้ cast ::uuid + IS NULL กัน error เทียบ uuid กับ ''
func unsetCurrentStandings(ctx context.Context, tx pgx.Tx, schoolID, personnelID, excludeID string) error {
	const q = `
		UPDATE academic_standings SET is_current = FALSE, updated_at = now()
		WHERE school_id = $1 AND personnel_id = $2 AND is_current = TRUE AND deleted_at IS NULL
		  AND ($3::uuid IS NULL OR id <> $3::uuid)`
	if _, err := tx.Exec(ctx, q, schoolID, personnelID, nilIfEmpty(excludeID)); err != nil {
		return fmt.Errorf("repository: unset current standings: %w", err)
	}
	return nil
}

// softDeleteWithAudit รัน soft-delete query (3 args: school, personnel, id) + audit ใน tx
func softDeleteWithAudit(ctx context.Context, db *pgxpool.Pool, query, schoolID, personnelID, id string, audit domain.AuditEntry) (bool, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, query, schoolID, personnelID, id)
	if err != nil {
		return false, fmt.Errorf("repository: soft delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit soft delete: %w", err)
	}
	return true, nil
}

// isUniqueViolation ตรวจว่าเป็น unique violation (23505) ที่ constraint ชื่อมี substr
func isUniqueViolation(err error, constraintSubstr string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, constraintSubstr)
}
