package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// EnrollmentRepository — นักเรียนในห้อง (รายเทอม, นักเรียน 1 คน 1 ห้องต่อเทอม)
type EnrollmentRepository struct {
	db *pgxpool.Pool
}

func NewEnrollmentRepository(db *pgxpool.Pool) *EnrollmentRepository {
	return &EnrollmentRepository{db: db}
}

func (r *EnrollmentRepository) ListByClass(ctx context.Context, schoolID, classID string) ([]domain.ClassEnrollment, error) {
	const q = `
		SELECT se.id, se.student_id, se.student_no, s.student_code, COALESCE(s.prefix, ''), s.first_name, s.last_name
		FROM student_enrollments se
		JOIN students s ON s.id = se.student_id
		WHERE se.school_id = $1 AND se.class_id = $2 AND se.deleted_at IS NULL AND s.deleted_at IS NULL
		ORDER BY se.student_no NULLS LAST, s.first_name`
	rows, err := r.db.Query(ctx, q, schoolID, classID)
	if err != nil {
		return nil, fmt.Errorf("repository: list enrollments: %w", err)
	}
	defer rows.Close()

	var out []domain.ClassEnrollment
	for rows.Next() {
		var e domain.ClassEnrollment
		if err := rows.Scan(&e.ID, &e.StudentID, &e.StudentNo, &e.StudentCode, &e.Prefix, &e.FirstName, &e.LastName); err != nil {
			return nil, fmt.Errorf("repository: scan enrollment: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// SearchByName ค้นหานักเรียนในเทอม (ตามชื่อ/นามสกุล/รหัส) พร้อมห้องที่สังกัด — ข้อมูลพื้นฐานเท่านั้น
func (r *EnrollmentRepository) SearchByName(ctx context.Context, schoolID, semesterID, term string) ([]domain.StudentClassBrief, error) {
	const q = `
		SELECT s.id, s.student_code, COALESCE(s.prefix, ''), s.first_name, s.last_name,
			c.grade_level, c.room_name
		FROM student_enrollments se
		JOIN students s ON s.id = se.student_id AND s.deleted_at IS NULL
		JOIN classes c ON c.id = se.class_id AND c.deleted_at IS NULL
		WHERE se.school_id = $1 AND se.semester_id = $2 AND se.deleted_at IS NULL
			AND (s.first_name ILIKE $3 OR s.last_name ILIKE $3 OR s.student_code ILIKE $3)
		ORDER BY c.grade_level, c.room_name, s.first_name
		LIMIT 50`
	rows, err := r.db.Query(ctx, q, schoolID, semesterID, "%"+term+"%")
	if err != nil {
		return nil, fmt.Errorf("repository: search students: %w", err)
	}
	defer rows.Close()

	var out []domain.StudentClassBrief
	for rows.Next() {
		var b domain.StudentClassBrief
		if err := rows.Scan(&b.StudentID, &b.StudentCode, &b.Prefix, &b.FirstName, &b.LastName, &b.GradeLevel, &b.RoomName); err != nil {
			return nil, fmt.Errorf("repository: scan student brief: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// Enroll จัดนักเรียนเข้าห้อง (upsert ตาม student_id+semester_id → ย้ายห้องถ้าเคยอยู่ห้องอื่นในเทอมนี้) + audit
func (r *EnrollmentRepository) Enroll(ctx context.Context, schoolID, semesterID, classID string, ne domain.NewEnrollment, audit domain.AuditEntry) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO student_enrollments (school_id, semester_id, student_id, class_id, student_no)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (student_id, semester_id) DO UPDATE
			SET class_id = EXCLUDED.class_id, student_no = EXCLUDED.student_no,
			    deleted_at = NULL, updated_at = now()`
	if _, err := tx.Exec(ctx, q, schoolID, semesterID, ne.StudentID, classID, ne.StudentNo); err != nil {
		return fmt.Errorf("repository: enroll student: %w", err)
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repository: commit enroll: %w", err)
	}
	return nil
}

// Remove ถอนนักเรียนออกจากห้อง (soft delete ด้วย enrollment id) + audit
func (r *EnrollmentRepository) Remove(ctx context.Context, schoolID, classID, enrollmentID string, audit domain.AuditEntry) (bool, error) {
	return softDeleteWithAudit(ctx, r.db,
		`UPDATE student_enrollments SET deleted_at = now(), updated_at = now()
		 WHERE id = $3 AND class_id = $2 AND school_id = $1 AND deleted_at IS NULL`,
		schoolID, classID, enrollmentID, audit)
}
