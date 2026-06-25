package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// PersonnelRepository เข้าถึงตาราง personnel/users/audit_logs — ทุก query scope ด้วย school_id
type PersonnelRepository struct {
	db *pgxpool.Pool
}

// NewPersonnelRepository สร้าง repository
func NewPersonnelRepository(db *pgxpool.Pool) *PersonnelRepository {
	return &PersonnelRepository{db: db}
}

// SELECT มาตรฐานสำหรับ personnel + ข้อมูล user (COALESCE คอลัมน์ที่ nullable เป็น '')
const personnelSelectCols = `
	p.id, p.user_id, p.national_id_encrypted, p.national_id_hash,
	p.civil_servant_id_encrypted, COALESCE(p.civil_servant_id_hash, ''),
	COALESCE(p.prefix, ''), p.first_name, p.last_name, p.birth_date,
	COALESCE(p.phone, ''), COALESCE(p.email, ''),
	COALESCE(p.house_no, ''), COALESCE(p.moo, ''), COALESCE(p.road, ''),
	COALESCE(p.subdistrict, ''), COALESCE(p.district, ''), COALESCE(p.province, ''),
	COALESCE(p.postal_code, ''), COALESCE(p.photo_path, ''),
	u.username, u.role, u.is_active, p.created_at, p.updated_at`

func scanPersonnel(row pgx.Row) (*domain.Personnel, error) {
	var p domain.Personnel
	err := row.Scan(
		&p.ID, &p.UserID, &p.NationalIDEnc, &p.NationalIDHash,
		&p.CivilServantIDEnc, &p.CivilServantIDHash,
		&p.Profile.Prefix, &p.Profile.FirstName, &p.Profile.LastName, &p.Profile.BirthDate,
		&p.Profile.Phone, &p.Profile.Email,
		&p.Profile.Address.HouseNo, &p.Profile.Address.Moo, &p.Profile.Address.Road,
		&p.Profile.Address.Subdistrict, &p.Profile.Address.District, &p.Profile.Address.Province,
		&p.Profile.Address.PostalCode, &p.PhotoPath,
		&p.Username, &p.Role, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// List คืนบุคลากรของโรงเรียน (เรียงใหม่สุดก่อน) + จำนวนรวม
func (r *PersonnelRepository) List(ctx context.Context, schoolID string, limit, offset int) ([]domain.Personnel, int, error) {
	const countQ = `SELECT count(*) FROM personnel WHERE school_id = $1 AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRow(ctx, countQ, schoolID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("repository: count personnel: %w", err)
	}

	q := `
		SELECT ` + personnelSelectCols + `
		FROM personnel p JOIN users u ON u.id = p.user_id
		WHERE p.school_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, schoolID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: list personnel: %w", err)
	}
	defer rows.Close()

	var out []domain.Personnel
	for rows.Next() {
		p, err := scanPersonnel(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("repository: scan personnel: %w", err)
		}
		out = append(out, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("repository: list personnel rows: %w", err)
	}
	return out, total, nil
}

// GetByID คืนบุคลากร 1 รายภายในโรงเรียน; คืน (nil, nil) ถ้าไม่พบ
func (r *PersonnelRepository) GetByID(ctx context.Context, schoolID, id string) (*domain.Personnel, error) {
	q := `
		SELECT ` + personnelSelectCols + `
		FROM personnel p JOIN users u ON u.id = p.user_id
		WHERE p.school_id = $1 AND p.id = $2 AND p.deleted_at IS NULL`
	p, err := scanPersonnel(r.db.QueryRow(ctx, q, schoolID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get personnel: %w", err)
	}
	return p, nil
}

// Create สร้าง user + personnel + audit ใน transaction เดียว; คืน id ของ personnel ใหม่
func (r *PersonnelRepository) Create(ctx context.Context, schoolID string, np domain.NewPersonnel, audit domain.AuditEntry) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const userQ = `
		INSERT INTO users (school_id, username, password_hash, role, is_school_admin, is_active)
		VALUES ($1, $2, $3, $4, $5, TRUE)
		RETURNING id`
	var userID string
	err = tx.QueryRow(ctx, userQ, schoolID, np.Username, np.PasswordHash, np.Role, np.IsSchoolAdmin).Scan(&userID)
	if err != nil {
		return "", mapUniqueViolation(err)
	}

	const pQ = `
		INSERT INTO personnel (
			school_id, user_id, national_id_encrypted, national_id_hash,
			civil_servant_id_encrypted, civil_servant_id_hash,
			prefix, first_name, last_name, birth_date, phone, email,
			house_no, moo, road, subdistrict, district, province, postal_code, photo_path)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
		RETURNING id`
	var personnelID string
	err = tx.QueryRow(ctx, pQ,
		schoolID, userID, np.NationalIDEnc, np.NationalIDHash,
		nilIfEmptyBytes(np.CivilServantIDEnc), nilIfEmpty(np.CivilServantIDHash),
		nilIfEmpty(np.Profile.Prefix), np.Profile.FirstName, np.Profile.LastName, np.Profile.BirthDate,
		nilIfEmpty(np.Profile.Phone), nilIfEmpty(np.Profile.Email),
		nilIfEmpty(np.Profile.Address.HouseNo), nilIfEmpty(np.Profile.Address.Moo), nilIfEmpty(np.Profile.Address.Road),
		nilIfEmpty(np.Profile.Address.Subdistrict), nilIfEmpty(np.Profile.Address.District), nilIfEmpty(np.Profile.Address.Province),
		nilIfEmpty(np.Profile.Address.PostalCode), nilIfEmpty(np.PhotoPath),
	).Scan(&personnelID)
	if err != nil {
		return "", mapUniqueViolation(err)
	}

	audit.TargetID = personnelID
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("repository: commit create personnel: %w", err)
	}
	return personnelID, nil
}

// Update แก้โปรไฟล์ (+ เลขบัตรถ้าระบุ) + audit ใน tx; คืน found=false ถ้าไม่พบ
func (r *PersonnelRepository) Update(ctx context.Context, schoolID, id string, up domain.UpdatePersonnel, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const baseQ = `
		UPDATE personnel SET
			prefix = $3, first_name = $4, last_name = $5, birth_date = $6,
			phone = $7, email = $8,
			house_no = $9, moo = $10, road = $11, subdistrict = $12,
			district = $13, province = $14, postal_code = $15,
			updated_at = now()
		WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`
	tag, err := tx.Exec(ctx, baseQ,
		schoolID, id,
		nilIfEmpty(up.Profile.Prefix), up.Profile.FirstName, up.Profile.LastName, up.Profile.BirthDate,
		nilIfEmpty(up.Profile.Phone), nilIfEmpty(up.Profile.Email),
		nilIfEmpty(up.Profile.Address.HouseNo), nilIfEmpty(up.Profile.Address.Moo), nilIfEmpty(up.Profile.Address.Road),
		nilIfEmpty(up.Profile.Address.Subdistrict), nilIfEmpty(up.Profile.Address.District),
		nilIfEmpty(up.Profile.Address.Province), nilIfEmpty(up.Profile.Address.PostalCode),
	)
	if err != nil {
		return false, fmt.Errorf("repository: update personnel: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil // ไม่พบ (หรือคนละโรงเรียน) — ไม่ commit
	}

	if up.ChangeNationalID {
		const q = `UPDATE personnel SET national_id_encrypted = $3, national_id_hash = $4, updated_at = now()
			WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`
		if _, err := tx.Exec(ctx, q, schoolID, id, up.NationalIDEnc, up.NationalIDHash); err != nil {
			return false, mapUniqueViolation(err)
		}
	}
	if up.ChangeCivilID {
		const q = `UPDATE personnel SET civil_servant_id_encrypted = $3, civil_servant_id_hash = $4, updated_at = now()
			WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`
		if _, err := tx.Exec(ctx, q, schoolID, id,
			nilIfEmptyBytes(up.CivilServantIDEnc), nilIfEmpty(up.CivilServantIDHash)); err != nil {
			return false, fmt.Errorf("repository: update civil id: %w", err)
		}
	}

	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit update personnel: %w", err)
	}
	return true, nil
}

// SoftDelete ตั้ง deleted_at ของ personnel + ปิดใช้งาน user + audit ใน tx; คืน found=false ถ้าไม่พบ
func (r *PersonnelRepository) SoftDelete(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const delQ = `
		UPDATE personnel SET deleted_at = now(), updated_at = now()
		WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL
		RETURNING user_id`
	var userID string
	err = tx.QueryRow(ctx, delQ, schoolID, id).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("repository: soft delete personnel: %w", err)
	}

	const deactivateQ = `
		UPDATE users SET is_active = FALSE, updated_at = now()
		WHERE id = $2 AND school_id = $1`
	if _, err := tx.Exec(ctx, deactivateQ, schoolID, userID); err != nil {
		return false, fmt.Errorf("repository: deactivate user: %w", err)
	}

	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit delete personnel: %w", err)
	}
	return true, nil
}

// InsertAudit บันทึก audit นอก transaction (ใช้กับ event ประเภท view)
func (r *PersonnelRepository) InsertAudit(ctx context.Context, audit domain.AuditEntry) error {
	return insertAuditTx(ctx, r.db, audit)
}

// IsUserInWorkGroup ตรวจว่า user สังกัดกลุ่มงานรหัสที่ระบุไหม (scope school_id)
func (r *PersonnelRepository) IsUserInWorkGroup(ctx context.Context, schoolID, userID, groupCode string) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1 FROM user_work_groups uwg
			JOIN work_groups wg ON wg.id = uwg.work_group_id
			WHERE uwg.school_id = $1 AND uwg.user_id = $2 AND wg.code = $3
			  AND uwg.deleted_at IS NULL AND wg.deleted_at IS NULL
		)`
	var exists bool
	if err := r.db.QueryRow(ctx, q, schoolID, userID, groupCode).Scan(&exists); err != nil {
		return false, fmt.Errorf("repository: check work group: %w", err)
	}
	return exists, nil
}

// querier ครอบทั้ง pool และ tx เพื่อ reuse insertAudit ได้ทั้งใน/นอก tx
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func insertAuditTx(ctx context.Context, q querier, audit domain.AuditEntry) error {
	const auditQ = `
		INSERT INTO audit_logs (school_id, actor_user_id, action, target_type, target_id, detail, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	var detailJSON []byte
	if audit.Detail != nil {
		b, err := json.Marshal(audit.Detail)
		if err != nil {
			return fmt.Errorf("repository: marshal audit detail: %w", err)
		}
		detailJSON = b
	}
	_, err := q.Exec(ctx, auditQ,
		audit.SchoolID, audit.ActorUserID, audit.Action, audit.TargetType,
		nilIfEmpty(audit.TargetID), detailJSON, nilIfEmpty(audit.IPAddress))
	if err != nil {
		return fmt.Errorf("repository: insert audit: %w", err)
	}
	return nil
}

// mapUniqueViolation แปลง unique violation (23505) เป็น domain error ที่สื่อความหมาย
func mapUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch {
		case strings.Contains(pgErr.ConstraintName, "national_id_hash"):
			return domain.ErrDuplicateNationalID
		case strings.Contains(pgErr.ConstraintName, "username"):
			return domain.ErrDuplicateUsername
		case strings.Contains(pgErr.ConstraintName, "student_code"):
			return domain.ErrDuplicateStudentCode
		case strings.Contains(pgErr.ConstraintName, "student_id_guardian_id"):
			return domain.ErrDuplicateGuardianLink
		}
	}
	return fmt.Errorf("repository: write: %w", err)
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nilIfEmptyBytes(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}
