package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// AcademicRepository — ปีการศึกษา + ภาคเรียน (scope school_id)
type AcademicRepository struct {
	db *pgxpool.Pool
}

func NewAcademicRepository(db *pgxpool.Pool) *AcademicRepository {
	return &AcademicRepository{db: db}
}

// --- ปีการศึกษา ---

func (r *AcademicRepository) ListYears(ctx context.Context, schoolID string) ([]domain.AcademicYear, error) {
	const q = `
		SELECT id, year, is_current, created_at, updated_at
		FROM academic_years
		WHERE school_id = $1 AND deleted_at IS NULL
		ORDER BY year DESC`
	rows, err := r.db.Query(ctx, q, schoolID)
	if err != nil {
		return nil, fmt.Errorf("repository: list years: %w", err)
	}
	defer rows.Close()

	var out []domain.AcademicYear
	for rows.Next() {
		var y domain.AcademicYear
		y.SchoolID = schoolID
		if err := rows.Scan(&y.ID, &y.Year, &y.IsCurrent, &y.CreatedAt, &y.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan year: %w", err)
		}
		out = append(out, y)
	}
	return out, rows.Err()
}

func (r *AcademicRepository) CreateYear(ctx context.Context, schoolID string, year int, audit domain.AuditEntry) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id string
	err = tx.QueryRow(ctx, `INSERT INTO academic_years (school_id, year) VALUES ($1, $2) RETURNING id`, schoolID, year).Scan(&id)
	if err != nil {
		if isUniqueViolation(err, "year") {
			return "", domain.ErrDuplicateYear
		}
		return "", fmt.Errorf("repository: insert year: %w", err)
	}
	audit.TargetID = id
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("repository: commit create year: %w", err)
	}
	return id, nil
}

// SetCurrentYear ตั้งปีการศึกษาปัจจุบัน (unset อันอื่นก่อนเพื่อกันชน partial unique index)
func (r *AcademicRepository) SetCurrentYear(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error) {
	return r.setCurrent(ctx, "academic_years", schoolID, id, audit)
}

// --- ภาคเรียน ---

func (r *AcademicRepository) ListSemesters(ctx context.Context, schoolID string) ([]domain.Semester, error) {
	const q = `
		SELECT s.id, s.academic_year_id, ay.year, s.term,
			s.start_date, s.end_date, s.is_current, s.created_at, s.updated_at
		FROM semesters s
		JOIN academic_years ay ON ay.id = s.academic_year_id
		WHERE s.school_id = $1 AND s.deleted_at IS NULL
		ORDER BY ay.year DESC, s.term`
	rows, err := r.db.Query(ctx, q, schoolID)
	if err != nil {
		return nil, fmt.Errorf("repository: list semesters: %w", err)
	}
	defer rows.Close()

	var out []domain.Semester
	for rows.Next() {
		var s domain.Semester
		s.SchoolID = schoolID
		if err := rows.Scan(&s.ID, &s.AcademicYearID, &s.Year, &s.Term,
			&s.StartDate, &s.EndDate, &s.IsCurrent, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan semester: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// yearExists ตรวจว่าปีการศึกษาอยู่ในโรงเรียน (ใช้ก่อนสร้างเทอม)
func (r *AcademicRepository) yearExists(ctx context.Context, q querierRow, schoolID, yearID string) (bool, error) {
	var ok bool
	err := q.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM academic_years WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL)`,
		schoolID, yearID).Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("repository: year exists: %w", err)
	}
	return ok, nil
}

func (r *AcademicRepository) CreateSemester(ctx context.Context, schoolID string, ns domain.NewSemester, audit domain.AuditEntry) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	exists, err := r.yearExists(ctx, tx, schoolID, ns.AcademicYearID)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", domain.ErrYearNotFound
	}

	var id string
	err = tx.QueryRow(ctx,
		`INSERT INTO semesters (school_id, academic_year_id, term, start_date, end_date)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		schoolID, ns.AcademicYearID, ns.Term, ns.StartDate, ns.EndDate).Scan(&id)
	if err != nil {
		if isUniqueViolation(err, "term") {
			return "", domain.ErrDuplicateSemester
		}
		return "", fmt.Errorf("repository: insert semester: %w", err)
	}
	audit.TargetID = id
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("repository: commit create semester: %w", err)
	}
	return id, nil
}

// SetCurrentSemester ตั้งภาคเรียนปัจจุบัน (unset อันอื่นก่อน)
func (r *AcademicRepository) SetCurrentSemester(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error) {
	return r.setCurrent(ctx, "semesters", schoolID, id, audit)
}

// --- shared ---

// querierRow ครอบ QueryRow ได้ทั้ง pool และ tx
type querierRow interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// setCurrent ตั้ง is_current ของแถว id (unset แถวอื่นของโรงเรียนก่อน) + audit — table = academic_years/semesters
func (r *AcademicRepository) setCurrent(ctx context.Context, table, schoolID, id string, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var exists bool
	if err := tx.QueryRow(ctx,
		fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL)`, table),
		schoolID, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("repository: set current exists: %w", err)
	}
	if !exists {
		return false, nil
	}

	if _, err := tx.Exec(ctx,
		fmt.Sprintf(`UPDATE %s SET is_current = FALSE, updated_at = now() WHERE school_id = $1 AND is_current = TRUE AND deleted_at IS NULL`, table),
		schoolID); err != nil {
		return false, fmt.Errorf("repository: unset current: %w", err)
	}
	if _, err := tx.Exec(ctx,
		fmt.Sprintf(`UPDATE %s SET is_current = TRUE, updated_at = now() WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`, table),
		schoolID, id); err != nil {
		return false, fmt.Errorf("repository: set current: %w", err)
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit set current: %w", err)
	}
	return true, nil
}
