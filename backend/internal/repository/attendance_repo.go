package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumkosoft/backend/internal/domain"
)

// AttendanceRepository — เช็คชื่อเข้าเรียนรายวัน (scope school_id + semester_id)
type AttendanceRepository struct {
	db *pgxpool.Pool
}

func NewAttendanceRepository(db *pgxpool.Pool) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

// RosterByClassDate คืนรายชื่อนักเรียนในห้อง พร้อมผลเช็คชื่อของวันนั้น (LEFT JOIN — ยังไม่เช็ค = ว่าง)
func (r *AttendanceRepository) RosterByClassDate(ctx context.Context, schoolID, classID string, date time.Time) ([]domain.AttendanceRosterEntry, error) {
	const q = `
		SELECT se.student_id, se.student_no, s.student_code, COALESCE(s.prefix, ''), s.first_name, s.last_name,
			COALESCE(a.id::text, ''), COALESCE(a.status, ''), COALESCE(a.note, '')
		FROM student_enrollments se
		JOIN students s ON s.id = se.student_id
		LEFT JOIN attendances a
			ON a.student_id = se.student_id AND a.date = $3 AND a.school_id = $1 AND a.deleted_at IS NULL
		WHERE se.school_id = $1 AND se.class_id = $2 AND se.deleted_at IS NULL AND s.deleted_at IS NULL
		ORDER BY se.student_no NULLS LAST, s.first_name`
	rows, err := r.db.Query(ctx, q, schoolID, classID, date)
	if err != nil {
		return nil, fmt.Errorf("repository: roster by class/date: %w", err)
	}
	defer rows.Close()

	var out []domain.AttendanceRosterEntry
	for rows.Next() {
		var e domain.AttendanceRosterEntry
		if err := rows.Scan(&e.StudentID, &e.StudentNo, &e.StudentCode, &e.Prefix, &e.FirstName, &e.LastName,
			&e.AttendanceID, &e.Status, &e.Note); err != nil {
			return nil, fmt.Errorf("repository: scan roster entry: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// BulkUpsert บันทึกผลเช็คชื่อทั้งห้องของวันนั้น (upsert ตาม student_id+date) + audit ใน tx เดียว
func (r *AttendanceRepository) BulkUpsert(ctx context.Context, schoolID, semesterID, classID string, date time.Time, marks []domain.AttendanceMark, checkedBy string, audit domain.AuditEntry) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO attendances (school_id, semester_id, class_id, student_id, date, status, note, checked_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (student_id, date) DO UPDATE
			SET status = EXCLUDED.status, note = EXCLUDED.note, class_id = EXCLUDED.class_id,
			    semester_id = EXCLUDED.semester_id, checked_by = EXCLUDED.checked_by,
			    deleted_at = NULL, updated_at = now()`
	for _, m := range marks {
		if _, err := tx.Exec(ctx, q, schoolID, semesterID, classID, m.StudentID, date, m.Status, nilIfEmpty(m.Note), checkedBy); err != nil {
			return fmt.Errorf("repository: upsert attendance: %w", err)
		}
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repository: commit attendance: %w", err)
	}
	return nil
}

// StatusForStudentDate คืนสถานะเช็คชื่อเข้าเรียนของนักเรียนในวันนั้น (ถ้ามี) — ใช้กันสแกนซ้ำที่ kiosk
func (r *AttendanceRepository) StatusForStudentDate(ctx context.Context, schoolID, studentID string, date time.Time) (string, bool, error) {
	var status string
	err := r.db.QueryRow(ctx,
		`SELECT status FROM attendances WHERE school_id = $1 AND student_id = $2 AND date = $3 AND deleted_at IS NULL`,
		schoolID, studentID, date).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("repository: attendance status for student: %w", err)
	}
	return status, true, nil
}

// IsClassAdvisorUser ตรวจว่าผู้ใช้ (user) เป็นครูที่ปรึกษาของห้องนี้หรือไม่ (ผ่าน personnel.user_id)
func (r *AttendanceRepository) IsClassAdvisorUser(ctx context.Context, schoolID, classID, userID string) (bool, error) {
	const q = `
		SELECT EXISTS(
			SELECT 1 FROM class_advisors ca
			JOIN personnel p ON p.id = ca.personnel_id
			WHERE ca.school_id = $1 AND ca.class_id = $2 AND p.user_id = $3
			  AND ca.deleted_at IS NULL AND p.deleted_at IS NULL)`
	var ok bool
	if err := r.db.QueryRow(ctx, q, schoolID, classID, userID).Scan(&ok); err != nil {
		return false, fmt.Errorf("repository: is class advisor: %w", err)
	}
	return ok, nil
}
