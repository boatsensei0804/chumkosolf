package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumkosoft/backend/internal/domain"
)

// StudentGuardianRepository จัดการความเชื่อมโยงนักเรียน↔ผู้ปกครอง (M:N) — scope school_id
type StudentGuardianRepository struct {
	db *pgxpool.Pool
}

func NewStudentGuardianRepository(db *pgxpool.Pool) *StudentGuardianRepository {
	return &StudentGuardianRepository{db: db}
}

// ListByStudent คืนผู้ปกครองของนักเรียน (join ข้อมูลผู้ปกครองสำหรับแสดงผล; เลขบัตรเป็น ciphertext)
func (r *StudentGuardianRepository) ListByStudent(ctx context.Context, schoolID, studentID string) ([]domain.StudentGuardian, error) {
	const q = `
		SELECT sg.id, sg.guardian_id, sg.relationship, sg.is_primary,
			COALESCE(g.prefix, ''), g.first_name, g.last_name, COALESCE(g.phone, ''), g.national_id_encrypted
		FROM student_guardians sg
		JOIN guardians g ON g.id = sg.guardian_id
		WHERE sg.school_id = $1 AND sg.student_id = $2
		  AND sg.deleted_at IS NULL AND g.deleted_at IS NULL
		ORDER BY sg.is_primary DESC, g.first_name`
	rows, err := r.db.Query(ctx, q, schoolID, studentID)
	if err != nil {
		return nil, fmt.Errorf("repository: list student guardians: %w", err)
	}
	defer rows.Close()

	var out []domain.StudentGuardian
	for rows.Next() {
		var sg domain.StudentGuardian
		if err := rows.Scan(&sg.ID, &sg.GuardianID, &sg.Relationship, &sg.IsPrimary,
			&sg.Prefix, &sg.FirstName, &sg.LastName, &sg.Phone, &sg.NationalIDEnc); err != nil {
			return nil, fmt.Errorf("repository: scan student guardian: %w", err)
		}
		out = append(out, sg)
	}
	return out, rows.Err()
}

// Link เชื่อมผู้ปกครองเข้านักเรียน (upsert: เชื่อมซ้ำ = อัปเดต relationship/primary, คืนสภาพถ้าเคยถอด)
// ถ้า is_primary=true จะ unset ผู้ปกครองหลักคนอื่นก่อน (กันชน partial unique)
func (r *StudentGuardianRepository) Link(ctx context.Context, schoolID, studentID string, nsg domain.NewStudentGuardian, audit domain.AuditEntry) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if nsg.IsPrimary {
		if err := unsetPrimaryGuardians(ctx, tx, schoolID, studentID, nsg.GuardianID); err != nil {
			return err
		}
	}

	const q = `
		INSERT INTO student_guardians (school_id, student_id, guardian_id, relationship, is_primary)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (student_id, guardian_id) DO UPDATE
			SET relationship = EXCLUDED.relationship, is_primary = EXCLUDED.is_primary,
			    deleted_at = NULL, updated_at = now()`
	if _, err := tx.Exec(ctx, q, schoolID, studentID, nsg.GuardianID, nsg.Relationship, nsg.IsPrimary); err != nil {
		return fmt.Errorf("repository: link guardian: %w", err)
	}

	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repository: commit link guardian: %w", err)
	}
	return nil
}

// Unlink ถอดผู้ปกครองออกจากนักเรียน (soft delete ด้วย link id) + audit
func (r *StudentGuardianRepository) Unlink(ctx context.Context, schoolID, studentID, linkID string, audit domain.AuditEntry) (bool, error) {
	return softDeleteWithAudit(ctx, r.db,
		`UPDATE student_guardians SET deleted_at = now(), updated_at = now()
		 WHERE id = $3 AND student_id = $2 AND school_id = $1 AND deleted_at IS NULL`,
		schoolID, studentID, linkID, audit)
}

// unsetPrimaryGuardians ตั้ง is_primary=false ของผู้ปกครองคนอื่น (ยกเว้น guardian ที่กำลังตั้ง)
func unsetPrimaryGuardians(ctx context.Context, tx pgx.Tx, schoolID, studentID, exceptGuardianID string) error {
	const q = `
		UPDATE student_guardians SET is_primary = FALSE, updated_at = now()
		WHERE school_id = $1 AND student_id = $2 AND is_primary = TRUE AND deleted_at IS NULL
		  AND ($3::uuid IS NULL OR guardian_id <> $3::uuid)`
	if _, err := tx.Exec(ctx, q, schoolID, studentID, nilIfEmpty(exceptGuardianID)); err != nil {
		return fmt.Errorf("repository: unset primary guardians: %w", err)
	}
	return nil
}
