package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// TermRepository — อ่านปีการศึกษา/ภาคเรียนปัจจุบันของโรงเรียน
type TermRepository struct {
	db *pgxpool.Pool
}

func NewTermRepository(db *pgxpool.Pool) *TermRepository {
	return &TermRepository{db: db}
}

// SemesterInSchool ตรวจว่า semester นี้อยู่ในโรงเรียนที่ระบุ (ใช้ validate ตอนสลับเทอม)
func (r *TermRepository) SemesterInSchool(ctx context.Context, schoolID, semesterID string) (bool, error) {
	const q = `SELECT EXISTS(
		SELECT 1 FROM semesters WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL)`
	var ok bool
	if err := r.db.QueryRow(ctx, q, schoolID, semesterID).Scan(&ok); err != nil {
		return false, fmt.Errorf("repository: semester in school: %w", err)
	}
	return ok, nil
}

// GetCurrent คืนปี+เทอมปัจจุบัน (nil ถ้ายังไม่ได้กำหนด)
func (r *TermRepository) GetCurrent(ctx context.Context, schoolID string) (*domain.CurrentTerm, error) {
	const q = `
		SELECT s.id, ay.year, s.term
		FROM semesters s
		JOIN academic_years ay ON ay.id = s.academic_year_id
		WHERE s.school_id = $1 AND s.is_current = TRUE AND s.deleted_at IS NULL
		LIMIT 1`
	var t domain.CurrentTerm
	err := r.db.QueryRow(ctx, q, schoolID).Scan(&t.SemesterID, &t.AcademicYear, &t.Term)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: current term: %w", err)
	}
	return &t, nil
}
