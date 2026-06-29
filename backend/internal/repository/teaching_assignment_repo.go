package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// TeachingAssignmentRepository — มอบหมายการสอน (รายเทอม, scope school_id + semester_id)
type TeachingAssignmentRepository struct {
	db *pgxpool.Pool
}

func NewTeachingAssignmentRepository(db *pgxpool.Pool) *TeachingAssignmentRepository {
	return &TeachingAssignmentRepository{db: db}
}

func (r *TeachingAssignmentRepository) ListBySemester(ctx context.Context, schoolID, semesterID string) ([]domain.TeachingAssignment, error) {
	const q = `
		SELECT ta.id, ta.personnel_id, ta.subject_id, ta.class_id,
			COALESCE(p.prefix, ''), p.first_name, p.last_name,
			s.subject_code, s.name,
			c.grade_level, c.room_name,
			ta.created_at, ta.updated_at
		FROM teaching_assignments ta
		JOIN personnel p ON p.id = ta.personnel_id
		JOIN subjects s ON s.id = ta.subject_id
		JOIN classes c ON c.id = ta.class_id
		WHERE ta.school_id = $1 AND ta.semester_id = $2 AND ta.deleted_at IS NULL
		ORDER BY c.grade_level, c.room_name, s.subject_code`
	rows, err := r.db.Query(ctx, q, schoolID, semesterID)
	if err != nil {
		return nil, fmt.Errorf("repository: list teaching assignments: %w", err)
	}
	defer rows.Close()

	var out []domain.TeachingAssignment
	for rows.Next() {
		var a domain.TeachingAssignment
		a.SchoolID = schoolID
		a.SemesterID = semesterID
		if err := rows.Scan(&a.ID, &a.PersonnelID, &a.SubjectID, &a.ClassID,
			&a.TeacherPrefix, &a.TeacherFirstName, &a.TeacherLastName,
			&a.SubjectCode, &a.SubjectName, &a.GradeLevel, &a.RoomName,
			&a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan teaching assignment: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// GetByID คืนการมอบหมาย (scope school_id) — ใช้ยืนยันการมีอยู่/เจ้าของ
func (r *TeachingAssignmentRepository) GetByID(ctx context.Context, schoolID, id string) (*domain.TeachingAssignment, error) {
	const q = `
		SELECT id, semester_id, personnel_id, subject_id, class_id, created_at, updated_at
		FROM teaching_assignments WHERE school_id = $1 AND id = $2 AND deleted_at IS NULL`
	var a domain.TeachingAssignment
	a.SchoolID = schoolID
	err := r.db.QueryRow(ctx, q, schoolID, id).Scan(&a.ID, &a.SemesterID, &a.PersonnelID, &a.SubjectID, &a.ClassID, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get teaching assignment: %w", err)
	}
	return &a, nil
}

// Create มอบหมายการสอน (upsert: คืนสภาพถ้าเคยถอด) + audit
func (r *TeachingAssignmentRepository) Create(ctx context.Context, schoolID, semesterID string, na domain.NewTeachingAssignment, audit domain.AuditEntry) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO teaching_assignments (school_id, semester_id, personnel_id, subject_id, class_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (semester_id, personnel_id, subject_id, class_id)
			DO UPDATE SET deleted_at = NULL, updated_at = now()
		RETURNING id`
	var id string
	if err := tx.QueryRow(ctx, q, schoolID, semesterID, na.PersonnelID, na.SubjectID, na.ClassID).Scan(&id); err != nil {
		return "", fmt.Errorf("repository: insert teaching assignment: %w", err)
	}
	audit.TargetID = id
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("repository: commit teaching assignment: %w", err)
	}
	return id, nil
}

func (r *TeachingAssignmentRepository) SoftDelete(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`UPDATE teaching_assignments SET deleted_at = now(), updated_at = now()
		 WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`, schoolID, id)
	if err != nil {
		return false, fmt.Errorf("repository: soft delete teaching assignment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit delete teaching assignment: %w", err)
	}
	return true, nil
}
