package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// SubjectAttendanceRepository — เช็คชื่อรายวิชา (รายคาบ, scope school_id + semester_id)
type SubjectAttendanceRepository struct {
	db *pgxpool.Pool
}

func NewSubjectAttendanceRepository(db *pgxpool.Pool) *SubjectAttendanceRepository {
	return &SubjectAttendanceRepository{db: db}
}

// SlotContext คืนห้องและ user ของครูประจำคาบ (ใช้ตรวจสิทธิ์ + หา roster)
func (r *SubjectAttendanceRepository) SlotContext(ctx context.Context, schoolID, slotID string) (classID, teacherUserID string, found bool, err error) {
	const q = `
		SELECT t.class_id, p.user_id
		FROM timetables t
		JOIN teaching_assignments ta ON ta.id = t.teaching_assignment_id
		JOIN personnel p ON p.id = ta.personnel_id
		WHERE t.school_id = $1 AND t.id = $2 AND t.deleted_at IS NULL`
	err = r.db.QueryRow(ctx, q, schoolID, slotID).Scan(&classID, &teacherUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", false, nil
	}
	if err != nil {
		return "", "", false, fmt.Errorf("repository: slot context: %w", err)
	}
	return classID, teacherUserID, true, nil
}

// RosterBySlotDate คืนรายชื่อนักเรียนในห้องของคาบ พร้อมผลเช็คชื่อรายวิชาของวันนั้น
func (r *SubjectAttendanceRepository) RosterBySlotDate(ctx context.Context, schoolID, slotID, classID string, date time.Time) ([]domain.AttendanceRosterEntry, error) {
	// a = เช็คชื่อเข้าเรียนรายวัน (เพื่อโชว์ "มาสาย" ในเช็คชื่อรายวิชา)
	const q = `
		SELECT se.student_id, se.student_no, s.student_code, COALESCE(s.prefix, ''), s.first_name, s.last_name,
			COALESCE(sa.id::text, ''), COALESCE(sa.status, ''), COALESCE(sa.note, ''), COALESCE(a.status, '')
		FROM student_enrollments se
		JOIN students s ON s.id = se.student_id
		LEFT JOIN subject_attendances sa
			ON sa.timetable_id = $3 AND sa.student_id = se.student_id AND sa.date = $4
			   AND sa.school_id = $1 AND sa.deleted_at IS NULL
		LEFT JOIN attendances a
			ON a.student_id = se.student_id AND a.date = $4
			   AND a.school_id = $1 AND a.deleted_at IS NULL
		WHERE se.school_id = $1 AND se.class_id = $2 AND se.deleted_at IS NULL AND s.deleted_at IS NULL
		ORDER BY se.student_no NULLS LAST, s.first_name`
	rows, err := r.db.Query(ctx, q, schoolID, classID, slotID, date)
	if err != nil {
		return nil, fmt.Errorf("repository: subject roster: %w", err)
	}
	defer rows.Close()

	var out []domain.AttendanceRosterEntry
	for rows.Next() {
		var e domain.AttendanceRosterEntry
		if err := rows.Scan(&e.StudentID, &e.StudentNo, &e.StudentCode, &e.Prefix, &e.FirstName, &e.LastName,
			&e.AttendanceID, &e.Status, &e.Note, &e.DailyStatus); err != nil {
			return nil, fmt.Errorf("repository: scan subject roster: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// TeacherSlots คืนคาบที่ครู (user) สอนในเทอม (เรียงวัน/คาบ)
func (r *SubjectAttendanceRepository) TeacherSlots(ctx context.Context, schoolID, semesterID, userID string) ([]domain.TeacherCheckinSlot, error) {
	const q = `
		SELECT t.id, t.day_of_week, t.period_no, s.subject_code, s.name, c.grade_level, c.room_name
		FROM timetables t
		JOIN teaching_assignments ta ON ta.id = t.teaching_assignment_id
		JOIN personnel p ON p.id = ta.personnel_id
		JOIN subjects s ON s.id = ta.subject_id
		JOIN classes c ON c.id = t.class_id
		WHERE t.school_id = $1 AND t.semester_id = $2 AND p.user_id = $3 AND t.deleted_at IS NULL
		ORDER BY t.day_of_week, t.period_no`
	rows, err := r.db.Query(ctx, q, schoolID, semesterID, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: teacher slots: %w", err)
	}
	defer rows.Close()

	var out []domain.TeacherCheckinSlot
	for rows.Next() {
		var s domain.TeacherCheckinSlot
		if err := rows.Scan(&s.SlotID, &s.DayOfWeek, &s.PeriodNo, &s.SubjectCode, &s.SubjectName, &s.GradeLevel, &s.RoomName); err != nil {
			return nil, fmt.Errorf("repository: scan teacher slot: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// CheckedSlotDates คืนคู่ (คาบ, วันที่) ที่มีการเช็คชื่อแล้วของคาบที่ครูสอน
func (r *SubjectAttendanceRepository) CheckedSlotDates(ctx context.Context, schoolID, semesterID, userID string) ([]domain.SlotDate, error) {
	const q = `
		SELECT DISTINCT sa.timetable_id, sa.date
		FROM subject_attendances sa
		JOIN timetables t ON t.id = sa.timetable_id
		JOIN teaching_assignments ta ON ta.id = t.teaching_assignment_id
		JOIN personnel p ON p.id = ta.personnel_id
		WHERE sa.school_id = $1 AND sa.semester_id = $2 AND p.user_id = $3 AND sa.deleted_at IS NULL`
	rows, err := r.db.Query(ctx, q, schoolID, semesterID, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: checked slot dates: %w", err)
	}
	defer rows.Close()

	var out []domain.SlotDate
	for rows.Next() {
		var sd domain.SlotDate
		if err := rows.Scan(&sd.SlotID, &sd.Date); err != nil {
			return nil, fmt.Errorf("repository: scan checked slot date: %w", err)
		}
		out = append(out, sd)
	}
	return out, rows.Err()
}

// SemesterRange คืนวันเริ่ม/จบของเทอม (nil ถ้าไม่ได้กำหนด)
func (r *SubjectAttendanceRepository) SemesterRange(ctx context.Context, schoolID, semesterID string) (start, end *time.Time, err error) {
	const q = `SELECT start_date, end_date FROM semesters WHERE id = $2 AND school_id = $1 AND deleted_at IS NULL`
	err = r.db.QueryRow(ctx, q, schoolID, semesterID).Scan(&start, &end)
	if err != nil {
		return nil, nil, fmt.Errorf("repository: semester range: %w", err)
	}
	return start, end, nil
}

// BulkUpsert บันทึกผลเช็คชื่อรายวิชาทั้งคาบของวันนั้น (upsert ตาม timetable_id+student_id+date) + audit
func (r *SubjectAttendanceRepository) BulkUpsert(ctx context.Context, schoolID, semesterID, slotID string, date time.Time, marks []domain.AttendanceMark, checkedBy string, audit domain.AuditEntry) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO subject_attendances (school_id, semester_id, timetable_id, student_id, date, status, note, checked_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (timetable_id, student_id, date) DO UPDATE
			SET status = EXCLUDED.status, note = EXCLUDED.note, semester_id = EXCLUDED.semester_id,
			    checked_by = EXCLUDED.checked_by, deleted_at = NULL, updated_at = now()`
	for _, m := range marks {
		if _, err := tx.Exec(ctx, q, schoolID, semesterID, slotID, m.StudentID, date, m.Status, nilIfEmpty(m.Note), checkedBy); err != nil {
			return fmt.Errorf("repository: upsert subject attendance: %w", err)
		}
	}
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repository: commit subject attendance: %w", err)
	}
	return nil
}
