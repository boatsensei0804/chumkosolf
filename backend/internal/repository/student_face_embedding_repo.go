package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumkosoft/backend/internal/domain"
)

// StudentFaceEmbeddingRepository — embedding ใบหน้านักเรียน (1 รูป = 1 แถว) สำหรับระบบสแกนหน้า
type StudentFaceEmbeddingRepository struct {
	db *pgxpool.Pool
}

func NewStudentFaceEmbeddingRepository(db *pgxpool.Pool) *StudentFaceEmbeddingRepository {
	return &StudentFaceEmbeddingRepository{db: db}
}

// Upsert บันทึก/อัปเดต embedding ของรูป (ตาม photo_id)
func (r *StudentFaceEmbeddingRepository) Upsert(ctx context.Context, schoolID, studentID, photoID string, vector []float32) error {
	const q = `
		INSERT INTO student_face_embeddings (school_id, student_id, photo_id, embedding)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (photo_id) DO UPDATE
			SET student_id = EXCLUDED.student_id, embedding = EXCLUDED.embedding, updated_at = now()`
	if _, err := r.db.Exec(ctx, q, schoolID, studentID, photoID, vector); err != nil {
		return fmt.Errorf("repository: upsert face embedding: %w", err)
	}
	return nil
}

// ListBySchool คืน embedding ทั้งหมดของโรงเรียน (ใช้โหลดเข้า matcher)
func (r *StudentFaceEmbeddingRepository) ListBySchool(ctx context.Context, schoolID string) ([]domain.FaceEmbedding, error) {
	const q = `SELECT student_id, photo_id, embedding FROM student_face_embeddings WHERE school_id = $1`
	rows, err := r.db.Query(ctx, q, schoolID)
	if err != nil {
		return nil, fmt.Errorf("repository: list face embeddings: %w", err)
	}
	defer rows.Close()

	var out []domain.FaceEmbedding
	for rows.Next() {
		var e domain.FaceEmbedding
		if err := rows.Scan(&e.StudentID, &e.PhotoID, &e.Vector); err != nil {
			return nil, fmt.Errorf("repository: scan face embedding: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// DeleteByPhoto ลบ embedding ของรูป (ใช้ตอนลบรูป เพื่อให้ฐานใบหน้าตรงกับรูปจริง)
func (r *StudentFaceEmbeddingRepository) DeleteByPhoto(ctx context.Context, schoolID, photoID string) error {
	if _, err := r.db.Exec(ctx,
		`DELETE FROM student_face_embeddings WHERE school_id = $1 AND photo_id = $2`,
		schoolID, photoID); err != nil {
		return fmt.Errorf("repository: delete face embedding: %w", err)
	}
	return nil
}

// DeleteOrphans ลบ embedding ที่ photo_id ไม่อยู่ในชุด keep (ใช้ตอน reindex เพื่อล้างรูปที่ถูกลบไปแล้ว)
func (r *StudentFaceEmbeddingRepository) DeleteOrphans(ctx context.Context, schoolID string, keepPhotoIDs []string) error {
	if len(keepPhotoIDs) == 0 {
		_, err := r.db.Exec(ctx, `DELETE FROM student_face_embeddings WHERE school_id = $1`, schoolID)
		if err != nil {
			return fmt.Errorf("repository: clear face embeddings: %w", err)
		}
		return nil
	}
	_, err := r.db.Exec(ctx,
		`DELETE FROM student_face_embeddings WHERE school_id = $1 AND NOT (photo_id = ANY($2))`,
		schoolID, keepPhotoIDs)
	if err != nil {
		return fmt.Errorf("repository: delete orphan embeddings: %w", err)
	}
	return nil
}

// CountBySchool นับจำนวน embedding ที่ enroll แล้ว (สำหรับรายงานผล reindex)
func (r *StudentFaceEmbeddingRepository) CountBySchool(ctx context.Context, schoolID string) (int, error) {
	var n int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM student_face_embeddings WHERE school_id = $1`, schoolID).Scan(&n); err != nil {
		return 0, fmt.Errorf("repository: count face embeddings: %w", err)
	}
	return n, nil
}
