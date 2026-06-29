package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// ClassRepository เข้าถึงตาราง classes — scope ด้วย school_id + semester_id (ข้อมูลรายเทอม)
type ClassRepository struct {
	db *pgxpool.Pool
}

func NewClassRepository(db *pgxpool.Pool) *ClassRepository {
	return &ClassRepository{db: db}
}

// ListBySemester คืนห้องเรียนของเทอม พร้อมจำนวนนักเรียน/ครูที่ปรึกษา
func (r *ClassRepository) ListBySemester(ctx context.Context, schoolID, semesterID string) ([]domain.Class, error) {
	const q = `
		SELECT c.id, c.grade_level, c.room_name, c.created_at, c.updated_at,
			(SELECT count(*) FROM student_enrollments se WHERE se.class_id = c.id AND se.deleted_at IS NULL),
			(SELECT count(*) FROM class_advisors ca WHERE ca.class_id = c.id AND ca.deleted_at IS NULL)
		FROM classes c
		WHERE c.school_id = $1 AND c.semester_id = $2 AND c.deleted_at IS NULL
		ORDER BY c.grade_level, c.room_name`
	rows, err := r.db.Query(ctx, q, schoolID, semesterID)
	if err != nil {
		return nil, fmt.Errorf("repository: list classes: %w", err)
	}
	defer rows.Close()

	var out []domain.Class
	for rows.Next() {
		var c domain.Class
		c.SchoolID = schoolID
		c.SemesterID = semesterID
		if err := rows.Scan(&c.ID, &c.GradeLevel, &c.RoomName, &c.CreatedAt, &c.UpdatedAt, &c.StudentCount, &c.AdvisorCount); err != nil {
			return nil, fmt.Errorf("repository: scan class: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetByID คืนห้องเรียน (scope school_id)
func (r *ClassRepository) GetByID(ctx context.Context, schoolID, id string) (*domain.Class, error) {
	const q = `
		SELECT id, semester_id, grade_level, room_name, created_at, updated_at
		FROM classes WHERE school_id = $1 AND id = $2 AND deleted_at IS NULL`
	var c domain.Class
	c.SchoolID = schoolID
	err := r.db.QueryRow(ctx, q, schoolID, id).Scan(&c.ID, &c.SemesterID, &c.GradeLevel, &c.RoomName, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get class: %w", err)
	}
	return &c, nil
}

func (r *ClassRepository) Create(ctx context.Context, schoolID, semesterID string, nc domain.NewClass, audit domain.AuditEntry) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO classes (school_id, semester_id, grade_level, room_name)
		VALUES ($1, $2, $3, $4) RETURNING id`
	var id string
	if err := tx.QueryRow(ctx, q, schoolID, semesterID, nc.GradeLevel, nc.RoomName).Scan(&id); err != nil {
		return "", mapUniqueViolation(err)
	}
	audit.TargetID = id
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("repository: commit create class: %w", err)
	}
	return id, nil
}

func (r *ClassRepository) Update(ctx context.Context, schoolID, id string, uc domain.UpdateClass, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`UPDATE classes SET grade_level = $3, room_name = $4, updated_at = now()
		 WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`,
		schoolID, id, uc.GradeLevel, uc.RoomName)
	if err != nil {
		return false, mapUniqueViolation(err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit update class: %w", err)
	}
	return true, nil
}

func (r *ClassRepository) SoftDelete(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`UPDATE classes SET deleted_at = now(), updated_at = now()
		 WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`, schoolID, id)
	if err != nil {
		return false, fmt.Errorf("repository: soft delete class: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit delete class: %w", err)
	}
	return true, nil
}
