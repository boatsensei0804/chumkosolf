package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// PersonnelWorkRepository เข้าถึงตาราง personnel_works + personnel_work_files
// scope: works = school_id + semester_id (รายเทอม) + personnel_id; files = school_id + personnel_work_id
type PersonnelWorkRepository struct {
	db *pgxpool.Pool
}

func NewPersonnelWorkRepository(db *pgxpool.Pool) *PersonnelWorkRepository {
	return &PersonnelWorkRepository{db: db}
}

// ListByPersonnel คืนผลงานของบุคลากรในเทอมปัจจุบัน พร้อมจำนวนไฟล์แนบ
func (r *PersonnelWorkRepository) ListByPersonnel(ctx context.Context, schoolID, semesterID, personnelID string) ([]domain.PersonnelWork, error) {
	const q = `
		SELECT w.id, w.title, COALESCE(w.description, ''), w.work_date, w.created_at, w.updated_at,
			(SELECT count(*) FROM personnel_work_files f WHERE f.personnel_work_id = w.id AND f.deleted_at IS NULL)
		FROM personnel_works w
		WHERE w.school_id = $1 AND w.semester_id = $2 AND w.personnel_id = $3 AND w.deleted_at IS NULL
		ORDER BY w.work_date DESC NULLS LAST, w.created_at DESC`
	rows, err := r.db.Query(ctx, q, schoolID, semesterID, personnelID)
	if err != nil {
		return nil, fmt.Errorf("repository: list personnel works: %w", err)
	}
	defer rows.Close()

	var out []domain.PersonnelWork
	for rows.Next() {
		var w domain.PersonnelWork
		w.SchoolID = schoolID
		w.SemesterID = semesterID
		w.PersonnelID = personnelID
		if err := rows.Scan(&w.ID, &w.Title, &w.Description, &w.WorkDate, &w.CreatedAt, &w.UpdatedAt, &w.FileCount); err != nil {
			return nil, fmt.Errorf("repository: scan personnel work: %w", err)
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// GetByID คืนผลงาน (scope school_id + personnel_id) — ใช้ยืนยันว่ามีอยู่จริงก่อนแตะไฟล์
func (r *PersonnelWorkRepository) GetByID(ctx context.Context, schoolID, personnelID, id string) (*domain.PersonnelWork, error) {
	const q = `
		SELECT id, semester_id, title, COALESCE(description, ''), work_date, created_at, updated_at
		FROM personnel_works
		WHERE school_id = $1 AND personnel_id = $2 AND id = $3 AND deleted_at IS NULL`
	var w domain.PersonnelWork
	w.SchoolID = schoolID
	w.PersonnelID = personnelID
	err := r.db.QueryRow(ctx, q, schoolID, personnelID, id).
		Scan(&w.ID, &w.SemesterID, &w.Title, &w.Description, &w.WorkDate, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get personnel work: %w", err)
	}
	return &w, nil
}

// Create เพิ่มผลงาน (เซ็ต school_id + semester_id จาก context) + audit
func (r *PersonnelWorkRepository) Create(ctx context.Context, schoolID, semesterID, personnelID string, nw domain.NewPersonnelWork, audit domain.AuditEntry) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO personnel_works (school_id, semester_id, personnel_id, title, description, work_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	var id string
	if err := tx.QueryRow(ctx, q, schoolID, semesterID, personnelID, nw.Title, nilIfEmpty(nw.Description), nw.WorkDate).Scan(&id); err != nil {
		return "", fmt.Errorf("repository: insert personnel work: %w", err)
	}
	audit.TargetID = id
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("repository: commit create personnel work: %w", err)
	}
	return id, nil
}

// Update แก้ไขผลงาน (scope school_id + personnel_id) + audit
func (r *PersonnelWorkRepository) Update(ctx context.Context, schoolID, personnelID, id string, uw domain.UpdatePersonnelWork, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		UPDATE personnel_works
		SET title = $4, description = $5, work_date = $6, updated_at = now()
		WHERE id = $3 AND personnel_id = $2 AND school_id = $1 AND deleted_at IS NULL`
	tag, err := tx.Exec(ctx, q, schoolID, personnelID, id, uw.Title, nilIfEmpty(uw.Description), uw.WorkDate)
	if err != nil {
		return false, fmt.Errorf("repository: update personnel work: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit update personnel work: %w", err)
	}
	return true, nil
}

// SoftDelete ลบผลงาน + ไฟล์แนบทั้งหมด (scope school_id + personnel_id) + audit
// คืน storage path ของไฟล์ที่ถูกลบ เพื่อให้ service ลบ object ออกจาก storage ต่อ
func (r *PersonnelWorkRepository) SoftDelete(ctx context.Context, schoolID, personnelID, id string, audit domain.AuditEntry) (paths []string, found bool, err error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`UPDATE personnel_works SET deleted_at = now(), updated_at = now()
		 WHERE id = $3 AND personnel_id = $2 AND school_id = $1 AND deleted_at IS NULL`,
		schoolID, personnelID, id)
	if err != nil {
		return nil, false, fmt.Errorf("repository: soft delete personnel work: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, false, nil
	}

	// soft-delete ไฟล์แนบของผลงานนี้ พร้อมเก็บ storage_path ที่ถูกลบ
	rows, err := tx.Query(ctx,
		`UPDATE personnel_work_files SET deleted_at = now(), updated_at = now()
		 WHERE personnel_work_id = $2 AND school_id = $1 AND deleted_at IS NULL
		 RETURNING storage_path`,
		schoolID, id)
	if err != nil {
		return nil, false, fmt.Errorf("repository: soft delete work files: %w", err)
	}
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			rows.Close()
			return nil, false, fmt.Errorf("repository: scan deleted file path: %w", err)
		}
		paths = append(paths, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("repository: iterate deleted file paths: %w", err)
	}

	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return nil, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, false, fmt.Errorf("repository: commit delete personnel work: %w", err)
	}
	return paths, true, nil
}

// --- ไฟล์แนบ ---

// ListFiles คืนไฟล์แนบของผลงาน (scope school_id + personnel_work_id)
func (r *PersonnelWorkRepository) ListFiles(ctx context.Context, schoolID, workID string) ([]domain.PersonnelWorkFile, error) {
	const q = `
		SELECT id, file_type, storage_path, COALESCE(original_name, ''), COALESCE(content_type, ''), COALESCE(size_bytes, 0), created_at
		FROM personnel_work_files
		WHERE school_id = $1 AND personnel_work_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, schoolID, workID)
	if err != nil {
		return nil, fmt.Errorf("repository: list work files: %w", err)
	}
	defer rows.Close()

	var out []domain.PersonnelWorkFile
	for rows.Next() {
		var f domain.PersonnelWorkFile
		f.SchoolID = schoolID
		f.PersonnelWorkID = workID
		if err := rows.Scan(&f.ID, &f.FileType, &f.StoragePath, &f.OriginalName, &f.ContentType, &f.SizeBytes, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan work file: %w", err)
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// AddFile บันทึก metadata ไฟล์แนบ (หลังอัปโหลดขึ้น storage แล้ว) + audit
func (r *PersonnelWorkRepository) AddFile(ctx context.Context, schoolID, workID string, nf domain.NewPersonnelWorkFile, audit domain.AuditEntry) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO personnel_work_files (school_id, personnel_work_id, file_type, storage_path, original_name, content_type, size_bytes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	var id string
	if err := tx.QueryRow(ctx, q, schoolID, workID, nf.FileType, nf.StoragePath, nilIfEmpty(nf.OriginalName), nilIfEmpty(nf.ContentType), nf.SizeBytes).Scan(&id); err != nil {
		return "", fmt.Errorf("repository: insert work file: %w", err)
	}
	audit.TargetID = id
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("repository: commit add work file: %w", err)
	}
	return id, nil
}

// GetFile คืนไฟล์แนบรายการเดียว (scope school_id + personnel_work_id) — ใช้สร้าง signed URL/ลบ
func (r *PersonnelWorkRepository) GetFile(ctx context.Context, schoolID, workID, id string) (*domain.PersonnelWorkFile, error) {
	const q = `
		SELECT id, file_type, storage_path, COALESCE(original_name, ''), COALESCE(content_type, ''), COALESCE(size_bytes, 0), created_at
		FROM personnel_work_files
		WHERE school_id = $1 AND personnel_work_id = $2 AND id = $3 AND deleted_at IS NULL`
	var f domain.PersonnelWorkFile
	f.SchoolID = schoolID
	f.PersonnelWorkID = workID
	err := r.db.QueryRow(ctx, q, schoolID, workID, id).
		Scan(&f.ID, &f.FileType, &f.StoragePath, &f.OriginalName, &f.ContentType, &f.SizeBytes, &f.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get work file: %w", err)
	}
	return &f, nil
}

// SoftDeleteFile ลบไฟล์แนบ (scope school_id + personnel_work_id) + audit
// คืน storage_path ที่ถูกลบ เพื่อให้ service ลบ object ออกจาก storage ต่อ
func (r *PersonnelWorkRepository) SoftDeleteFile(ctx context.Context, schoolID, workID, id string, audit domain.AuditEntry) (path string, found bool, err error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	err = tx.QueryRow(ctx,
		`UPDATE personnel_work_files SET deleted_at = now(), updated_at = now()
		 WHERE id = $3 AND personnel_work_id = $2 AND school_id = $1 AND deleted_at IS NULL
		 RETURNING storage_path`,
		schoolID, workID, id).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("repository: soft delete work file: %w", err)
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", false, fmt.Errorf("repository: commit delete work file: %w", err)
	}
	return path, true, nil
}
