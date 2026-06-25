package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// StudentRepository เข้าถึงตาราง students — ทุก query scope ด้วย school_id (PDPA: ข้อมูลเด็ก)
type StudentRepository struct {
	db *pgxpool.Pool
}

func NewStudentRepository(db *pgxpool.Pool) *StudentRepository {
	return &StudentRepository{db: db}
}

const studentSelectCols = `
	s.id, s.national_id_encrypted, s.national_id_hash, s.student_code,
	COALESCE(s.prefix, ''), s.first_name, s.last_name, s.birth_date, COALESCE(s.phone, ''),
	COALESCE(s.house_no, ''), COALESCE(s.moo, ''), COALESCE(s.road, ''),
	COALESCE(s.subdistrict, ''), COALESCE(s.district, ''), COALESCE(s.province, ''),
	COALESCE(s.postal_code, ''), COALESCE(s.photo_path, ''),
	s.created_at, s.updated_at`

func scanStudent(row pgx.Row) (*domain.Student, error) {
	var s domain.Student
	err := row.Scan(
		&s.ID, &s.NationalIDEnc, &s.NationalIDHash, &s.StudentCode,
		&s.Profile.Prefix, &s.Profile.FirstName, &s.Profile.LastName, &s.Profile.BirthDate, &s.Profile.Phone,
		&s.Profile.Address.HouseNo, &s.Profile.Address.Moo, &s.Profile.Address.Road,
		&s.Profile.Address.Subdistrict, &s.Profile.Address.District, &s.Profile.Address.Province,
		&s.Profile.Address.PostalCode, &s.PhotoPath, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepository) List(ctx context.Context, schoolID string, limit, offset int) ([]domain.Student, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM students WHERE school_id = $1 AND deleted_at IS NULL`, schoolID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("repository: count students: %w", err)
	}
	q := `SELECT ` + studentSelectCols + `
		FROM students s
		WHERE s.school_id = $1 AND s.deleted_at IS NULL
		ORDER BY s.created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, schoolID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: list students: %w", err)
	}
	defer rows.Close()

	var out []domain.Student
	for rows.Next() {
		s, err := scanStudent(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("repository: scan student: %w", err)
		}
		out = append(out, *s)
	}
	return out, total, rows.Err()
}

func (r *StudentRepository) GetByID(ctx context.Context, schoolID, id string) (*domain.Student, error) {
	q := `SELECT ` + studentSelectCols + ` FROM students s WHERE s.school_id = $1 AND s.id = $2 AND s.deleted_at IS NULL`
	s, err := scanStudent(r.db.QueryRow(ctx, q, schoolID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get student: %w", err)
	}
	return s, nil
}

func (r *StudentRepository) Create(ctx context.Context, schoolID string, ns domain.NewStudent, audit domain.AuditEntry) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO students (school_id, national_id_encrypted, national_id_hash, student_code,
			prefix, first_name, last_name, birth_date, phone,
			house_no, moo, road, subdistrict, district, province, postal_code)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING id`
	var id string
	err = tx.QueryRow(ctx, q,
		schoolID, ns.NationalIDEnc, ns.NationalIDHash, ns.StudentCode,
		nilIfEmpty(ns.Profile.Prefix), ns.Profile.FirstName, ns.Profile.LastName, ns.Profile.BirthDate, nilIfEmpty(ns.Profile.Phone),
		nilIfEmpty(ns.Profile.Address.HouseNo), nilIfEmpty(ns.Profile.Address.Moo), nilIfEmpty(ns.Profile.Address.Road),
		nilIfEmpty(ns.Profile.Address.Subdistrict), nilIfEmpty(ns.Profile.Address.District), nilIfEmpty(ns.Profile.Address.Province),
		nilIfEmpty(ns.Profile.Address.PostalCode),
	).Scan(&id)
	if err != nil {
		return "", mapUniqueViolation(err)
	}

	audit.TargetID = id
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("repository: commit create student: %w", err)
	}
	return id, nil
}

func (r *StudentRepository) Update(ctx context.Context, schoolID, id string, us domain.UpdateStudent, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const baseQ = `
		UPDATE students SET student_code = $3, prefix = $4, first_name = $5, last_name = $6,
			birth_date = $7, phone = $8, house_no = $9, moo = $10, road = $11,
			subdistrict = $12, district = $13, province = $14, postal_code = $15, updated_at = now()
		WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`
	tag, err := tx.Exec(ctx, baseQ,
		schoolID, id, us.StudentCode,
		nilIfEmpty(us.Profile.Prefix), us.Profile.FirstName, us.Profile.LastName, us.Profile.BirthDate, nilIfEmpty(us.Profile.Phone),
		nilIfEmpty(us.Profile.Address.HouseNo), nilIfEmpty(us.Profile.Address.Moo), nilIfEmpty(us.Profile.Address.Road),
		nilIfEmpty(us.Profile.Address.Subdistrict), nilIfEmpty(us.Profile.Address.District), nilIfEmpty(us.Profile.Address.Province),
		nilIfEmpty(us.Profile.Address.PostalCode),
	)
	if err != nil {
		return false, mapUniqueViolation(err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}

	if us.ChangeNationalID {
		if _, err := tx.Exec(ctx,
			`UPDATE students SET national_id_encrypted = $3, national_id_hash = $4, updated_at = now()
			 WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`,
			schoolID, id, us.NationalIDEnc, us.NationalIDHash); err != nil {
			return false, mapUniqueViolation(err)
		}
	}

	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit update student: %w", err)
	}
	return true, nil
}

func (r *StudentRepository) SoftDelete(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`UPDATE students SET deleted_at = now(), updated_at = now()
		 WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`, schoolID, id)
	if err != nil {
		return false, fmt.Errorf("repository: soft delete student: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit delete student: %w", err)
	}
	return true, nil
}
