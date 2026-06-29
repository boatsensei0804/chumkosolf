package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// SchoolRepository — ข้อมูลโรงเรียน (scope school_id จาก context)
type SchoolRepository struct {
	db *pgxpool.Pool
}

func NewSchoolRepository(db *pgxpool.Pool) *SchoolRepository {
	return &SchoolRepository{db: db}
}

func (r *SchoolRepository) Get(ctx context.Context, schoolID string) (*domain.School, error) {
	const q = `
		SELECT id, name, code,
			COALESCE(house_no, ''), COALESCE(moo, ''), COALESCE(road, ''), COALESCE(subdistrict, ''),
			COALESCE(district, ''), COALESCE(province, ''), COALESCE(postal_code, ''),
			COALESCE(phone, ''), COALESCE(email, ''), COALESCE(website, ''), COALESCE(director_name, ''),
			is_active, COALESCE(attendance_late_after, '08:00'), COALESCE(attendance_late_penalty, 5)
		FROM schools WHERE id = $1 AND deleted_at IS NULL`
	var s domain.School
	err := r.db.QueryRow(ctx, q, schoolID).Scan(&s.ID, &s.Name, &s.Code,
		&s.Address.HouseNo, &s.Address.Moo, &s.Address.Road, &s.Address.Subdistrict,
		&s.Address.District, &s.Address.Province, &s.Address.PostalCode,
		&s.Phone, &s.Email, &s.Website, &s.DirectorName,
		&s.IsActive, &s.AttendanceLateAfter, &s.AttendanceLatePenalty)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get school: %w", err)
	}
	return &s, nil
}

func (r *SchoolRepository) Update(ctx context.Context, schoolID string, us domain.UpdateSchool, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`UPDATE schools SET name = $2,
			house_no = $3, moo = $4, road = $5, subdistrict = $6, district = $7, province = $8, postal_code = $9,
			phone = $10, email = $11, website = $12, director_name = $13, attendance_late_after = $14,
			attendance_late_penalty = $15, updated_at = now()
		 WHERE id = $1 AND deleted_at IS NULL`,
		schoolID, us.Name,
		nilIfEmpty(us.Address.HouseNo), nilIfEmpty(us.Address.Moo), nilIfEmpty(us.Address.Road), nilIfEmpty(us.Address.Subdistrict),
		nilIfEmpty(us.Address.District), nilIfEmpty(us.Address.Province), nilIfEmpty(us.Address.PostalCode),
		nilIfEmpty(us.Phone), nilIfEmpty(us.Email), nilIfEmpty(us.Website), nilIfEmpty(us.DirectorName), us.AttendanceLateAfter,
		us.AttendanceLatePenalty)
	if err != nil {
		return false, fmt.Errorf("repository: update school: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit update school: %w", err)
	}
	return true, nil
}
