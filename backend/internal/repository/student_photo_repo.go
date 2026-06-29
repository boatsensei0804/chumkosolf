package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// StudentPhotoRepository — รูปนักเรียน (หลายรูปต่อคน) ตาราง student_photos
// sync students.photo_path ให้ชี้รูปโปรไฟล์ (is_primary) เสมอ ภายใน transaction เดียว
type StudentPhotoRepository struct {
	db *pgxpool.Pool
}

func NewStudentPhotoRepository(db *pgxpool.Pool) *StudentPhotoRepository {
	return &StudentPhotoRepository{db: db}
}

const studentPhotoCols = `id, student_id, storage_path, COALESCE(content_type, ''), COALESCE(size_bytes, 0), is_primary, created_at`

func scanStudentPhoto(row pgx.Row) (*domain.StudentPhoto, error) {
	var p domain.StudentPhoto
	if err := row.Scan(&p.ID, &p.StudentID, &p.StoragePath, &p.ContentType, &p.SizeBytes, &p.IsPrimary, &p.CreatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

// CountByStudent คืนจำนวนรูปที่ยังไม่ถูกลบของนักเรียน
func (r *StudentPhotoRepository) CountByStudent(ctx context.Context, schoolID, studentID string) (int, error) {
	var n int
	err := r.db.QueryRow(ctx,
		`SELECT count(*) FROM student_photos WHERE school_id = $1 AND student_id = $2 AND deleted_at IS NULL`,
		schoolID, studentID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("repository: count student photos: %w", err)
	}
	return n, nil
}

// ListByStudent คืนรูปทั้งหมดของนักเรียน (รูปโปรไฟล์มาก่อน แล้วเรียงตามเวลา)
func (r *StudentPhotoRepository) ListByStudent(ctx context.Context, schoolID, studentID string) ([]domain.StudentPhoto, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+studentPhotoCols+` FROM student_photos
		 WHERE school_id = $1 AND student_id = $2 AND deleted_at IS NULL
		 ORDER BY is_primary DESC, created_at`,
		schoolID, studentID)
	if err != nil {
		return nil, fmt.Errorf("repository: list student photos: %w", err)
	}
	defer rows.Close()

	var out []domain.StudentPhoto
	for rows.Next() {
		p, err := scanStudentPhoto(rows)
		if err != nil {
			return nil, fmt.Errorf("repository: scan student photo: %w", err)
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// Get คืนรูป 1 รายการ (ตรวจ school + ยังไม่ลบ)
func (r *StudentPhotoRepository) Get(ctx context.Context, schoolID, photoID string) (*domain.StudentPhoto, error) {
	p, err := scanStudentPhoto(r.db.QueryRow(ctx,
		`SELECT `+studentPhotoCols+` FROM student_photos
		 WHERE school_id = $1 AND id = $2 AND deleted_at IS NULL`,
		schoolID, photoID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get student photo: %w", err)
	}
	return p, nil
}

// Add เพิ่มรูปใหม่; ถ้า makePrimary จะตั้งเป็นรูปโปรไฟล์ (เคลียร์ของเดิม + sync students.photo_path)
// กันกรณีอัปโหลดพร้อมกันหลายรูปตอนยังไม่มีรูป: ถ้าชน unique index ของ primary → เพิ่มเป็นรูปธรรมดาแทน
func (r *StudentPhotoRepository) Add(ctx context.Context, schoolID, studentID string, np domain.NewStudentPhoto, makePrimary bool, audit domain.AuditEntry) (string, error) {
	id, err := r.addOnce(ctx, schoolID, studentID, np, makePrimary, audit)
	if err != nil && makePrimary && isUniqueViolation(err, "uq_student_photos_primary") {
		return r.addOnce(ctx, schoolID, studentID, np, false, audit)
	}
	return id, err
}

func (r *StudentPhotoRepository) addOnce(ctx context.Context, schoolID, studentID string, np domain.NewStudentPhoto, makePrimary bool, audit domain.AuditEntry) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if makePrimary {
		if _, err := tx.Exec(ctx,
			`UPDATE student_photos SET is_primary = false, updated_at = now()
			 WHERE school_id = $1 AND student_id = $2 AND is_primary AND deleted_at IS NULL`,
			schoolID, studentID); err != nil {
			return "", fmt.Errorf("repository: clear primary: %w", err)
		}
	}

	var id string
	if err := tx.QueryRow(ctx,
		`INSERT INTO student_photos (school_id, student_id, storage_path, content_type, size_bytes, is_primary)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		schoolID, studentID, np.StoragePath, nilIfEmpty(np.ContentType), np.SizeBytes, makePrimary).Scan(&id); err != nil {
		return "", fmt.Errorf("repository: insert student photo: %w", err)
	}

	if makePrimary {
		if err := syncStudentPhotoPathTx(ctx, tx, schoolID, studentID, np.StoragePath); err != nil {
			return "", err
		}
	}

	audit.TargetID = studentID
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("repository: commit add student photo: %w", err)
	}
	return id, nil
}

// SetPrimary ตั้งรูปที่เลือกเป็นรูปโปรไฟล์ (เคลียร์ของเดิม + sync students.photo_path)
func (r *StudentPhotoRepository) SetPrimary(ctx context.Context, schoolID, studentID, photoID string, audit domain.AuditEntry) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// ต้องเคลียร์ของเดิมก่อน เพราะ partial unique index บังคับ primary ได้ 1 แถว
	if _, err := tx.Exec(ctx,
		`UPDATE student_photos SET is_primary = false, updated_at = now()
		 WHERE school_id = $1 AND student_id = $2 AND is_primary AND deleted_at IS NULL`,
		schoolID, studentID); err != nil {
		return false, fmt.Errorf("repository: clear primary: %w", err)
	}

	var storagePath string
	err = tx.QueryRow(ctx,
		`UPDATE student_photos SET is_primary = true, updated_at = now()
		 WHERE school_id = $1 AND student_id = $2 AND id = $3 AND deleted_at IS NULL
		 RETURNING storage_path`,
		schoolID, studentID, photoID).Scan(&storagePath)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("repository: set primary: %w", err)
	}

	if err := syncStudentPhotoPathTx(ctx, tx, schoolID, studentID, storagePath); err != nil {
		return false, err
	}
	audit.TargetID = studentID
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("repository: commit set primary: %w", err)
	}
	return true, nil
}

// SoftDelete ลบรูป (soft); ถ้าเป็นรูปโปรไฟล์จะเลื่อนรูปล่าสุดที่เหลือเป็นโปรไฟล์แทน (หรือเคลียร์ถ้าหมด)
// คืน storage_path ของรูปที่ลบ เพื่อให้ service ลบ object จริง
func (r *StudentPhotoRepository) SoftDelete(ctx context.Context, schoolID, studentID, photoID string, audit domain.AuditEntry) (string, bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", false, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var storagePath string
	var wasPrimary bool
	err = tx.QueryRow(ctx,
		`SELECT storage_path, is_primary FROM student_photos
		 WHERE school_id = $1 AND student_id = $2 AND id = $3 AND deleted_at IS NULL FOR UPDATE`,
		schoolID, studentID, photoID).Scan(&storagePath, &wasPrimary)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("repository: lock student photo: %w", err)
	}

	if _, err := tx.Exec(ctx,
		`UPDATE student_photos SET deleted_at = now(), is_primary = false, updated_at = now() WHERE id = $1`,
		photoID); err != nil {
		return "", false, fmt.Errorf("repository: soft delete student photo: %w", err)
	}

	if wasPrimary {
		// เลื่อนรูปล่าสุดที่เหลือเป็นโปรไฟล์ใหม่ (ถ้ามี)
		var newID, newPath string
		err := tx.QueryRow(ctx,
			`SELECT id, storage_path FROM student_photos
			 WHERE school_id = $1 AND student_id = $2 AND deleted_at IS NULL
			 ORDER BY created_at DESC LIMIT 1`,
			schoolID, studentID).Scan(&newID, &newPath)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			if err := syncStudentPhotoPathTx(ctx, tx, schoolID, studentID, ""); err != nil {
				return "", false, err
			}
		case err != nil:
			return "", false, fmt.Errorf("repository: pick new primary: %w", err)
		default:
			if _, err := tx.Exec(ctx, `UPDATE student_photos SET is_primary = true, updated_at = now() WHERE id = $1`, newID); err != nil {
				return "", false, fmt.Errorf("repository: promote primary: %w", err)
			}
			if err := syncStudentPhotoPathTx(ctx, tx, schoolID, studentID, newPath); err != nil {
				return "", false, err
			}
		}
	}

	audit.TargetID = studentID
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", false, fmt.Errorf("repository: commit delete student photo: %w", err)
	}
	return storagePath, true, nil
}

// Dataset คืนแถว (นักเรียน × รูป) สำหรับสร้าง dataset สแกนหน้า
// classID ว่าง = ทุกนักเรียนในโรงเรียนที่มีรูป; ถ้าระบุ = เฉพาะนักเรียนในห้องนั้น (ตามเทอม)
// ป้ายห้อง (grade/room) มาจากการจัดห้องของเทอมปัจจุบัน (ว่างถ้าไม่มี)
func (r *StudentPhotoRepository) Dataset(ctx context.Context, schoolID, semesterID, classID string) ([]domain.StudentPhotoRow, error) {
	args := []any{schoolID, nilIfEmpty(semesterID)}
	q := `
		SELECT s.id, s.student_code, COALESCE(s.prefix, ''), s.first_name, s.last_name,
			COALESCE(c.grade_level, ''), COALESCE(c.room_name, ''),
			sp.id, sp.storage_path, sp.is_primary
		FROM student_photos sp
		JOIN students s ON s.id = sp.student_id AND s.deleted_at IS NULL
		LEFT JOIN student_enrollments se ON se.student_id = s.id AND se.school_id = sp.school_id
			AND se.semester_id = $2 AND se.deleted_at IS NULL
		LEFT JOIN classes c ON c.id = se.class_id AND c.deleted_at IS NULL
		WHERE sp.school_id = $1 AND sp.deleted_at IS NULL`
	if classID != "" {
		args = append(args, classID)
		q += ` AND se.class_id = $3`
	}
	q += ` ORDER BY s.first_name, s.last_name, s.id, sp.is_primary DESC, sp.created_at`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("repository: face dataset: %w", err)
	}
	defer rows.Close()

	var out []domain.StudentPhotoRow
	for rows.Next() {
		var r domain.StudentPhotoRow
		if err := rows.Scan(&r.StudentID, &r.StudentCode, &r.Prefix, &r.FirstName, &r.LastName,
			&r.GradeLevel, &r.RoomName, &r.PhotoID, &r.StoragePath, &r.IsPrimary); err != nil {
			return nil, fmt.Errorf("repository: scan dataset row: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// syncStudentPhotoPathTx ตั้ง students.photo_path ให้ตรงกับรูปโปรไฟล์ ("" = ล้าง)
func syncStudentPhotoPathTx(ctx context.Context, q querier, schoolID, studentID, photoPath string) error {
	if _, err := q.Exec(ctx,
		`UPDATE students SET photo_path = $3, updated_at = now() WHERE id = $2 AND school_id = $1`,
		schoolID, studentID, nilIfEmpty(photoPath)); err != nil {
		return fmt.Errorf("repository: sync student photo_path: %w", err)
	}
	return nil
}
