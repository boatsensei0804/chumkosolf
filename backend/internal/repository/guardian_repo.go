package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumkosoft/backend/internal/domain"
)

// GuardianRepository เข้าถึงตาราง guardians — scope ด้วย school_id (PDPA)
type GuardianRepository struct {
	db *pgxpool.Pool
}

func NewGuardianRepository(db *pgxpool.Pool) *GuardianRepository {
	return &GuardianRepository{db: db}
}

const guardianSelectCols = `
	g.id, g.national_id_encrypted, g.national_id_hash,
	COALESCE(g.prefix, ''), g.first_name, g.last_name, g.birth_date, COALESCE(g.phone, ''),
	COALESCE(g.house_no, ''), COALESCE(g.moo, ''), COALESCE(g.road, ''),
	COALESCE(g.subdistrict, ''), COALESCE(g.district, ''), COALESCE(g.province, ''),
	COALESCE(g.postal_code, ''), g.created_at, g.updated_at`

func scanGuardian(row pgx.Row) (*domain.Guardian, error) {
	var g domain.Guardian
	err := row.Scan(
		&g.ID, &g.NationalIDEnc, &g.NationalIDHash,
		&g.Profile.Prefix, &g.Profile.FirstName, &g.Profile.LastName, &g.Profile.BirthDate, &g.Profile.Phone,
		&g.Profile.Address.HouseNo, &g.Profile.Address.Moo, &g.Profile.Address.Road,
		&g.Profile.Address.Subdistrict, &g.Profile.Address.District, &g.Profile.Address.Province,
		&g.Profile.Address.PostalCode, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GuardianRepository) List(ctx context.Context, schoolID string, limit, offset int) ([]domain.Guardian, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM guardians WHERE school_id = $1 AND deleted_at IS NULL`, schoolID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("repository: count guardians: %w", err)
	}
	q := `SELECT ` + guardianSelectCols + `
		FROM guardians g WHERE g.school_id = $1 AND g.deleted_at IS NULL
		ORDER BY g.created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, schoolID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: list guardians: %w", err)
	}
	defer rows.Close()

	var out []domain.Guardian
	for rows.Next() {
		g, err := scanGuardian(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("repository: scan guardian: %w", err)
		}
		out = append(out, *g)
	}
	return out, total, rows.Err()
}

func (r *GuardianRepository) GetByID(ctx context.Context, schoolID, id string) (*domain.Guardian, error) {
	q := `SELECT ` + guardianSelectCols + ` FROM guardians g WHERE g.school_id = $1 AND g.id = $2 AND g.deleted_at IS NULL`
	g, err := scanGuardian(r.db.QueryRow(ctx, q, schoolID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get guardian: %w", err)
	}
	return g, nil
}

func (r *GuardianRepository) Create(ctx context.Context, schoolID string, ng domain.NewGuardian, audit domain.AuditEntry) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO guardians (school_id, national_id_encrypted, national_id_hash,
			prefix, first_name, last_name, birth_date, phone,
			house_no, moo, road, subdistrict, district, province, postal_code)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id`
	var id string
	err = tx.QueryRow(ctx, q,
		schoolID, ng.NationalIDEnc, ng.NationalIDHash,
		nilIfEmpty(ng.Profile.Prefix), ng.Profile.FirstName, ng.Profile.LastName, ng.Profile.BirthDate, nilIfEmpty(ng.Profile.Phone),
		nilIfEmpty(ng.Profile.Address.HouseNo), nilIfEmpty(ng.Profile.Address.Moo), nilIfEmpty(ng.Profile.Address.Road),
		nilIfEmpty(ng.Profile.Address.Subdistrict), nilIfEmpty(ng.Profile.Address.District), nilIfEmpty(ng.Profile.Address.Province),
		nilIfEmpty(ng.Profile.Address.PostalCode),
	).Scan(&id)
	if err != nil {
		return "", mapUniqueViolation(err)
	}

	audit.TargetID = id
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("repository: commit create guardian: %w", err)
	}
	return id, nil
}

// Upsert สร้างผู้ปกครอง หรือคืน id เดิมถ้าเลขบัตรซ้ำ (รองรับพี่น้องใช้ผู้ปกครองคนเดียวกัน)
// ไม่ทับข้อมูลโปรไฟล์เดิม (เก็บของเดิมไว้) — แค่คืนสภาพถ้าเคยถูกลบ
func (r *GuardianRepository) Upsert(ctx context.Context, schoolID string, ng domain.NewGuardian) (string, error) {
	const q = `
		INSERT INTO guardians (school_id, national_id_encrypted, national_id_hash,
			prefix, first_name, last_name, birth_date, phone,
			house_no, moo, road, subdistrict, district, province, postal_code)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (school_id, national_id_hash) DO UPDATE SET deleted_at = NULL, updated_at = now()
		RETURNING id`
	var id string
	err := r.db.QueryRow(ctx, q,
		schoolID, ng.NationalIDEnc, ng.NationalIDHash,
		nilIfEmpty(ng.Profile.Prefix), ng.Profile.FirstName, ng.Profile.LastName, ng.Profile.BirthDate, nilIfEmpty(ng.Profile.Phone),
		nilIfEmpty(ng.Profile.Address.HouseNo), nilIfEmpty(ng.Profile.Address.Moo), nilIfEmpty(ng.Profile.Address.Road),
		nilIfEmpty(ng.Profile.Address.Subdistrict), nilIfEmpty(ng.Profile.Address.District), nilIfEmpty(ng.Profile.Address.Province),
		nilIfEmpty(ng.Profile.Address.PostalCode),
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("repository: upsert guardian: %w", err)
	}
	return id, nil
}

func (r *GuardianRepository) Update(ctx context.Context, schoolID, id string, ug domain.UpdateGuardian, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const baseQ = `
		UPDATE guardians SET prefix = $3, first_name = $4, last_name = $5, birth_date = $6, phone = $7,
			house_no = $8, moo = $9, road = $10, subdistrict = $11, district = $12, province = $13,
			postal_code = $14, updated_at = now()
		WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`
	tag, err := tx.Exec(ctx, baseQ,
		schoolID, id,
		nilIfEmpty(ug.Profile.Prefix), ug.Profile.FirstName, ug.Profile.LastName, ug.Profile.BirthDate, nilIfEmpty(ug.Profile.Phone),
		nilIfEmpty(ug.Profile.Address.HouseNo), nilIfEmpty(ug.Profile.Address.Moo), nilIfEmpty(ug.Profile.Address.Road),
		nilIfEmpty(ug.Profile.Address.Subdistrict), nilIfEmpty(ug.Profile.Address.District), nilIfEmpty(ug.Profile.Address.Province),
		nilIfEmpty(ug.Profile.Address.PostalCode),
	)
	if err != nil {
		return false, fmt.Errorf("repository: update guardian: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}

	if ug.ChangeNationalID {
		if _, err := tx.Exec(ctx,
			`UPDATE guardians SET national_id_encrypted = $3, national_id_hash = $4, updated_at = now()
			 WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`,
			schoolID, id, ug.NationalIDEnc, ug.NationalIDHash); err != nil {
			return false, mapUniqueViolation(err)
		}
	}

	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit update guardian: %w", err)
	}
	return true, nil
}

func (r *GuardianRepository) SoftDelete(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`UPDATE guardians SET deleted_at = now(), updated_at = now()
		 WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`, schoolID, id)
	if err != nil {
		return false, fmt.Errorf("repository: soft delete guardian: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit delete guardian: %w", err)
	}
	return true, nil
}
