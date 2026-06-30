package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumkosoft/backend/internal/domain"
)

// SubjectRepository — รายวิชา (ถาวร, scope school_id)
type SubjectRepository struct {
	db *pgxpool.Pool
}

func NewSubjectRepository(db *pgxpool.Pool) *SubjectRepository {
	return &SubjectRepository{db: db}
}

func (r *SubjectRepository) List(ctx context.Context, schoolID string) ([]domain.Subject, error) {
	const q = `
		SELECT id, subject_code, name, credit::float8, created_at, updated_at
		FROM subjects
		WHERE school_id = $1 AND deleted_at IS NULL
		ORDER BY subject_code`
	rows, err := r.db.Query(ctx, q, schoolID)
	if err != nil {
		return nil, fmt.Errorf("repository: list subjects: %w", err)
	}
	defer rows.Close()

	var out []domain.Subject
	for rows.Next() {
		var s domain.Subject
		s.SchoolID = schoolID
		if err := rows.Scan(&s.ID, &s.SubjectCode, &s.Name, &s.Credit, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan subject: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SubjectRepository) GetByID(ctx context.Context, schoolID, id string) (*domain.Subject, error) {
	const q = `
		SELECT id, subject_code, name, credit::float8, created_at, updated_at
		FROM subjects WHERE school_id = $1 AND id = $2 AND deleted_at IS NULL`
	var s domain.Subject
	s.SchoolID = schoolID
	err := r.db.QueryRow(ctx, q, schoolID, id).Scan(&s.ID, &s.SubjectCode, &s.Name, &s.Credit, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get subject: %w", err)
	}
	return &s, nil
}

func (r *SubjectRepository) Create(ctx context.Context, schoolID string, ns domain.NewSubject, audit domain.AuditEntry) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO subjects (school_id, subject_code, name, credit)
		VALUES ($1, $2, $3, $4) RETURNING id`
	var id string
	if err := tx.QueryRow(ctx, q, schoolID, ns.SubjectCode, ns.Name, ns.Credit).Scan(&id); err != nil {
		return "", mapUniqueViolation(err)
	}
	audit.TargetID = id
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("repository: commit create subject: %w", err)
	}
	return id, nil
}

func (r *SubjectRepository) Update(ctx context.Context, schoolID, id string, us domain.UpdateSubject, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		UPDATE subjects SET subject_code = $3, name = $4, credit = $5, updated_at = now()
		WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`
	tag, err := tx.Exec(ctx, q, schoolID, id, us.SubjectCode, us.Name, us.Credit)
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
		return false, fmt.Errorf("repository: commit update subject: %w", err)
	}
	return true, nil
}

func (r *SubjectRepository) SoftDelete(ctx context.Context, schoolID, id string, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`UPDATE subjects SET deleted_at = now(), updated_at = now()
		 WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`, schoolID, id)
	if err != nil {
		return false, fmt.Errorf("repository: soft delete subject: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit delete subject: %w", err)
	}
	return true, nil
}
