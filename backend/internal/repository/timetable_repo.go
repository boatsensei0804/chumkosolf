package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// TimetableRepository — ตั้งค่าตารางสอน + นิยามคาบ + ช่องตารางสอน (รายเทอม)
type TimetableRepository struct {
	db *pgxpool.Pool
}

func NewTimetableRepository(db *pgxpool.Pool) *TimetableRepository {
	return &TimetableRepository{db: db}
}

// --- settings ---

// GetSettings คืนค่าตั้งของเทอม (nil ถ้ายังไม่ได้ตั้ง)
func (r *TimetableRepository) GetSettings(ctx context.Context, schoolID, semesterID string) (*domain.TimetableSettings, error) {
	const q = `
		SELECT id, days_per_week, periods_per_day, created_at, updated_at
		FROM timetable_settings
		WHERE school_id = $1 AND semester_id = $2 AND deleted_at IS NULL`
	var s domain.TimetableSettings
	s.SchoolID = schoolID
	s.SemesterID = semesterID
	err := r.db.QueryRow(ctx, q, schoolID, semesterID).Scan(&s.ID, &s.DaysPerWeek, &s.PeriodsPerDay, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get timetable settings: %w", err)
	}
	return &s, nil
}

// ListPeriods คืนนิยามคาบของเทอม (เรียงตามคาบ)
func (r *TimetableRepository) ListPeriods(ctx context.Context, schoolID, semesterID string) ([]domain.PeriodDefinition, error) {
	const q = `
		SELECT id, period_no, COALESCE(label, ''),
			COALESCE(to_char(start_time, 'HH24:MI'), ''), COALESCE(to_char(end_time, 'HH24:MI'), ''),
			is_break
		FROM period_definitions
		WHERE school_id = $1 AND semester_id = $2 AND deleted_at IS NULL
		ORDER BY period_no`
	rows, err := r.db.Query(ctx, q, schoolID, semesterID)
	if err != nil {
		return nil, fmt.Errorf("repository: list periods: %w", err)
	}
	defer rows.Close()

	var out []domain.PeriodDefinition
	for rows.Next() {
		var p domain.PeriodDefinition
		p.SchoolID = schoolID
		p.SemesterID = semesterID
		if err := rows.Scan(&p.ID, &p.PeriodNo, &p.Label, &p.StartTime, &p.EndTime, &p.IsBreak); err != nil {
			return nil, fmt.Errorf("repository: scan period: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// SaveConfig บันทึกค่าตั้ง + นิยามคาบทั้งชุด (upsert คาบที่ส่งมา, soft-delete คาบที่ไม่ได้ส่ง) ใน tx เดียว
func (r *TimetableRepository) SaveConfig(ctx context.Context, schoolID, semesterID string, daysPerWeek, periodsPerDay int, periods []domain.NewPeriodDefinition, audit domain.AuditEntry) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const settingsQ = `
		INSERT INTO timetable_settings (school_id, semester_id, days_per_week, periods_per_day)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (school_id, semester_id)
			DO UPDATE SET days_per_week = EXCLUDED.days_per_week, periods_per_day = EXCLUDED.periods_per_day,
			              deleted_at = NULL, updated_at = now()`
	if _, err := tx.Exec(ctx, settingsQ, schoolID, semesterID, daysPerWeek, periodsPerDay); err != nil {
		return fmt.Errorf("repository: upsert timetable settings: %w", err)
	}

	const periodQ = `
		INSERT INTO period_definitions (school_id, semester_id, period_no, label, start_time, end_time, is_break)
		VALUES ($1, $2, $3, $4, NULLIF($5,'')::time, NULLIF($6,'')::time, $7)
		ON CONFLICT (school_id, semester_id, period_no)
			DO UPDATE SET label = EXCLUDED.label, start_time = EXCLUDED.start_time, end_time = EXCLUDED.end_time,
			              is_break = EXCLUDED.is_break, deleted_at = NULL, updated_at = now()`
	nos := make([]int, 0, len(periods))
	for _, p := range periods {
		if _, err := tx.Exec(ctx, periodQ, schoolID, semesterID, p.PeriodNo, nilIfEmpty(p.Label), p.StartTime, p.EndTime, p.IsBreak); err != nil {
			return fmt.Errorf("repository: upsert period: %w", err)
		}
		nos = append(nos, p.PeriodNo)
	}

	// soft-delete คาบที่ไม่ได้อยู่ในชุดที่ส่งมา
	const pruneQ = `
		UPDATE period_definitions SET deleted_at = now(), updated_at = now()
		WHERE school_id = $1 AND semester_id = $2 AND deleted_at IS NULL AND NOT (period_no = ANY($3))`
	if _, err := tx.Exec(ctx, pruneQ, schoolID, semesterID, nos); err != nil {
		return fmt.Errorf("repository: prune periods: %w", err)
	}

	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repository: commit timetable config: %w", err)
	}
	return nil
}

// --- slots ---

// ListSlotsByClass คืนช่องตารางสอนของห้อง พร้อมชื่อวิชา/ครู
func (r *TimetableRepository) ListSlotsByClass(ctx context.Context, schoolID, classID string) ([]domain.TimetableSlot, error) {
	const q = `
		SELECT t.id, t.day_of_week, t.period_no, t.teaching_assignment_id,
			s.subject_code, s.name,
			COALESCE(p.prefix, ''), p.first_name, p.last_name
		FROM timetables t
		JOIN teaching_assignments ta ON ta.id = t.teaching_assignment_id
		JOIN subjects s ON s.id = ta.subject_id
		JOIN personnel p ON p.id = ta.personnel_id
		WHERE t.school_id = $1 AND t.class_id = $2 AND t.deleted_at IS NULL
		ORDER BY t.day_of_week, t.period_no`
	rows, err := r.db.Query(ctx, q, schoolID, classID)
	if err != nil {
		return nil, fmt.Errorf("repository: list timetable slots: %w", err)
	}
	defer rows.Close()

	var out []domain.TimetableSlot
	for rows.Next() {
		var t domain.TimetableSlot
		t.SchoolID = schoolID
		t.ClassID = classID
		if err := rows.Scan(&t.ID, &t.DayOfWeek, &t.PeriodNo, &t.TeachingAssignmentID,
			&t.SubjectCode, &t.SubjectName, &t.TeacherPrefix, &t.TeacherFirstName, &t.TeacherLastName); err != nil {
			return nil, fmt.Errorf("repository: scan timetable slot: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ListTeachers คืนรายชื่อบุคลากร (ครู/ผู้บริหาร) ที่ใช้งานอยู่ — สำหรับคำนวณครูว่าง
func (r *TimetableRepository) ListTeachers(ctx context.Context, schoolID string) ([]domain.TeacherBrief, error) {
	const q = `
		SELECT p.id, COALESCE(p.prefix, ''), p.first_name, p.last_name
		FROM personnel p
		JOIN users u ON u.id = p.user_id
		WHERE p.school_id = $1 AND p.deleted_at IS NULL AND u.is_active = true
		ORDER BY p.first_name, p.last_name`
	rows, err := r.db.Query(ctx, q, schoolID)
	if err != nil {
		return nil, fmt.Errorf("repository: list teachers: %w", err)
	}
	defer rows.Close()

	var out []domain.TeacherBrief
	for rows.Next() {
		var t domain.TeacherBrief
		if err := rows.Scan(&t.ID, &t.Prefix, &t.FirstName, &t.LastName); err != nil {
			return nil, fmt.Errorf("repository: scan teacher brief: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// BusyTeacherPeriods คืนคู่ (ครู, คาบ) ที่ติดสอนในวันที่ระบุ (ใช้คำนวณครูว่าง)
func (r *TimetableRepository) BusyTeacherPeriods(ctx context.Context, schoolID, semesterID string, dayOfWeek int) ([]domain.TeacherPeriod, error) {
	const q = `
		SELECT DISTINCT ta.personnel_id, t.period_no
		FROM timetables t
		JOIN teaching_assignments ta ON ta.id = t.teaching_assignment_id
		WHERE t.school_id = $1 AND t.semester_id = $2 AND t.day_of_week = $3 AND t.deleted_at IS NULL`
	rows, err := r.db.Query(ctx, q, schoolID, semesterID, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("repository: busy teacher periods: %w", err)
	}
	defer rows.Close()

	var out []domain.TeacherPeriod
	for rows.Next() {
		var tp domain.TeacherPeriod
		if err := rows.Scan(&tp.PersonnelID, &tp.PeriodNo); err != nil {
			return nil, fmt.Errorf("repository: scan teacher period: %w", err)
		}
		out = append(out, tp)
	}
	return out, rows.Err()
}

// TeacherSlotConflict ตรวจว่าครู (personnel) มีคาบสอนห้องอื่นในวัน+คาบเดียวกันแล้วหรือไม่
// (กันครูสอน 2 ห้องพร้อมกัน) — ไม่นับห้องที่กำลังจัด (excludeClassID)
func (r *TimetableRepository) TeacherSlotConflict(ctx context.Context, schoolID, semesterID, personnelID string, dayOfWeek, periodNo int, excludeClassID string) (bool, error) {
	const q = `
		SELECT EXISTS(
			SELECT 1 FROM timetables t
			JOIN teaching_assignments ta ON ta.id = t.teaching_assignment_id
			WHERE t.school_id = $1 AND t.semester_id = $2
			  AND t.day_of_week = $3 AND t.period_no = $4
			  AND ta.personnel_id = $5 AND t.class_id <> $6
			  AND t.deleted_at IS NULL)`
	var conflict bool
	if err := r.db.QueryRow(ctx, q, schoolID, semesterID, dayOfWeek, periodNo, personnelID, excludeClassID).Scan(&conflict); err != nil {
		return false, fmt.Errorf("repository: teacher slot conflict: %w", err)
	}
	return conflict, nil
}

// UpsertSlot ตั้งค่าช่องตาราง (ห้อง×วัน×คาบ) → มอบหมายการสอน + audit
func (r *TimetableRepository) UpsertSlot(ctx context.Context, schoolID, semesterID, classID string, ns domain.NewTimetableSlot, audit domain.AuditEntry) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO timetables (school_id, semester_id, class_id, day_of_week, period_no, teaching_assignment_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (semester_id, class_id, day_of_week, period_no)
			DO UPDATE SET teaching_assignment_id = EXCLUDED.teaching_assignment_id, deleted_at = NULL, updated_at = now()
		RETURNING id`
	var id string
	if err := tx.QueryRow(ctx, q, schoolID, semesterID, classID, ns.DayOfWeek, ns.PeriodNo, ns.TeachingAssignmentID).Scan(&id); err != nil {
		return "", fmt.Errorf("repository: upsert timetable slot: %w", err)
	}
	audit.TargetID = id
	if err := insertAuditTx(ctx, tx, audit); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("repository: commit timetable slot: %w", err)
	}
	return id, nil
}

// DeleteSlot ล้างช่องตาราง (soft delete) + audit
func (r *TimetableRepository) DeleteSlot(ctx context.Context, schoolID, classID, slotID string, audit domain.AuditEntry) (bool, error) {
	return softDeleteWithAudit(ctx, r.db,
		`UPDATE timetables SET deleted_at = now(), updated_at = now()
		 WHERE id = $3 AND class_id = $2 AND school_id = $1 AND deleted_at IS NULL`,
		schoolID, classID, slotID, audit)
}
