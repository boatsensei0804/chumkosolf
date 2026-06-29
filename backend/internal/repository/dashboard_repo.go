package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// DashboardRepository — ข้อมูลสรุปสำหรับหน้าแรก (ที่ปรึกษา/เช็คชื่อวันนี้/ตารางสอน)
type DashboardRepository struct {
	db *pgxpool.Pool
}

func NewDashboardRepository(db *pgxpool.Pool) *DashboardRepository {
	return &DashboardRepository{db: db}
}

// AdvisorSummary คืนจำนวนห้องที่ปรึกษา + จำนวนนักเรียนในที่ปรึกษาของครู (user)
func (r *DashboardRepository) AdvisorSummary(ctx context.Context, schoolID, semesterID, userID string) (classCount, studentCount int, err error) {
	const cq = `
		SELECT count(*) FROM class_advisors ca
		JOIN personnel p ON p.id = ca.personnel_id AND p.deleted_at IS NULL
		WHERE ca.school_id = $1 AND ca.semester_id = $2 AND p.user_id = $3 AND ca.deleted_at IS NULL`
	if err = r.db.QueryRow(ctx, cq, schoolID, semesterID, userID).Scan(&classCount); err != nil {
		return 0, 0, fmt.Errorf("repository: advisor class count: %w", err)
	}
	const sq = `
		SELECT count(*) FROM student_enrollments se
		JOIN class_advisors ca ON ca.class_id = se.class_id AND ca.school_id = se.school_id
			AND ca.semester_id = se.semester_id AND ca.deleted_at IS NULL
		JOIN personnel p ON p.id = ca.personnel_id AND p.deleted_at IS NULL
		WHERE se.school_id = $1 AND se.semester_id = $2 AND p.user_id = $3 AND se.deleted_at IS NULL`
	if err = r.db.QueryRow(ctx, sq, schoolID, semesterID, userID).Scan(&studentCount); err != nil {
		return 0, 0, fmt.Errorf("repository: advisee count: %w", err)
	}
	return classCount, studentCount, nil
}

// Advisees คืนรายชื่อนักเรียนในห้องที่ปรึกษาของครู (user) พร้อมสถานะเช็คชื่อเข้าเรียนของวันนี้
func (r *DashboardRepository) Advisees(ctx context.Context, schoolID, semesterID, userID string, date time.Time) ([]domain.Advisee, error) {
	const q = `
		SELECT s.id, s.student_code, COALESCE(s.prefix, ''), s.first_name, s.last_name,
			COALESCE(s.phone, ''), s.national_id_encrypted,
			c.grade_level, c.room_name, COALESCE(a.status, '')
		FROM student_enrollments se
		JOIN class_advisors ca ON ca.class_id = se.class_id AND ca.school_id = se.school_id
			AND ca.semester_id = se.semester_id AND ca.deleted_at IS NULL
		JOIN personnel p ON p.id = ca.personnel_id AND p.deleted_at IS NULL
		JOIN students s ON s.id = se.student_id AND s.deleted_at IS NULL
		JOIN classes c ON c.id = se.class_id AND c.deleted_at IS NULL
		LEFT JOIN attendances a ON a.student_id = se.student_id AND a.date = $4
			AND a.school_id = $1 AND a.deleted_at IS NULL
		WHERE se.school_id = $1 AND se.semester_id = $2 AND p.user_id = $3 AND se.deleted_at IS NULL
		ORDER BY c.grade_level, c.room_name, s.student_code`
	rows, err := r.db.Query(ctx, q, schoolID, semesterID, userID, date)
	if err != nil {
		return nil, fmt.Errorf("repository: advisees: %w", err)
	}
	defer rows.Close()

	var out []domain.Advisee
	for rows.Next() {
		var a domain.Advisee
		if err := rows.Scan(&a.StudentID, &a.StudentCode, &a.Prefix, &a.FirstName, &a.LastName,
			&a.Phone, &a.NationalIDEnc, &a.GradeLevel, &a.RoomName, &a.TodayStatus); err != nil {
			return nil, fmt.Errorf("repository: scan advisee: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// IsAdvisee ตรวจว่านักเรียนอยู่ในห้องที่ปรึกษาของครู (user) ในเทอมนั้นจริง (ใช้คุมสิทธิ์แก้ข้อมูล)
func (r *DashboardRepository) IsAdvisee(ctx context.Context, schoolID, semesterID, userID, studentID string) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1 FROM student_enrollments se
			JOIN class_advisors ca ON ca.class_id = se.class_id AND ca.school_id = se.school_id
				AND ca.semester_id = se.semester_id AND ca.deleted_at IS NULL
			JOIN personnel p ON p.id = ca.personnel_id AND p.deleted_at IS NULL
			WHERE se.school_id = $1 AND se.semester_id = $2 AND p.user_id = $3
				AND se.student_id = $4 AND se.deleted_at IS NULL
		)`
	var ok bool
	if err := r.db.QueryRow(ctx, q, schoolID, semesterID, userID, studentID).Scan(&ok); err != nil {
		return false, fmt.Errorf("repository: is advisee: %w", err)
	}
	return ok, nil
}

// TodayAttendanceCounts คืนจำนวนนักเรียนที่ปรึกษาแยกตามสถานะเช็คชื่อของวันนี้ ("" = ยังไม่เช็ค)
func (r *DashboardRepository) TodayAttendanceCounts(ctx context.Context, schoolID, semesterID, userID string, date time.Time) ([]domain.StatusCount, error) {
	const q = `
		SELECT COALESCE(a.status, ''), count(*)
		FROM student_enrollments se
		JOIN class_advisors ca ON ca.class_id = se.class_id AND ca.school_id = se.school_id
			AND ca.semester_id = se.semester_id AND ca.deleted_at IS NULL
		JOIN personnel p ON p.id = ca.personnel_id AND p.deleted_at IS NULL
		LEFT JOIN attendances a ON a.student_id = se.student_id AND a.date = $4
			AND a.school_id = $1 AND a.deleted_at IS NULL
		WHERE se.school_id = $1 AND se.semester_id = $2 AND p.user_id = $3 AND se.deleted_at IS NULL
		GROUP BY COALESCE(a.status, '')`
	rows, err := r.db.Query(ctx, q, schoolID, semesterID, userID, date)
	if err != nil {
		return nil, fmt.Errorf("repository: today attendance counts: %w", err)
	}
	defer rows.Close()

	var out []domain.StatusCount
	for rows.Next() {
		var sc domain.StatusCount
		if err := rows.Scan(&sc.Status, &sc.Count); err != nil {
			return nil, fmt.Errorf("repository: scan status count: %w", err)
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

// TeacherSlots คืนคาบที่ครู (user) สอนในเทอม (สำหรับตารางสอนหน้าแรก)
func (r *DashboardRepository) TeacherSlots(ctx context.Context, schoolID, semesterID, userID string) ([]domain.TeacherCheckinSlot, error) {
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
		return nil, fmt.Errorf("repository: dashboard teacher slots: %w", err)
	}
	defer rows.Close()

	var out []domain.TeacherCheckinSlot
	for rows.Next() {
		var s domain.TeacherCheckinSlot
		if err := rows.Scan(&s.SlotID, &s.DayOfWeek, &s.PeriodNo, &s.SubjectCode, &s.SubjectName, &s.GradeLevel, &s.RoomName); err != nil {
			return nil, fmt.Errorf("repository: scan dashboard slot: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
