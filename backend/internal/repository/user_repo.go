// Package repository เข้าถึง DB เท่านั้น — ทุก query บังคับ scope ด้วย school_id
// (ห้ามมี business logic ที่นี่)
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chumko-platform/backend/internal/domain"
)

// UserRepository คือ implementation บน PostgreSQL ของ service.UserRepository
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository สร้าง repository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// SchoolIDByCode resolve โรงเรียนจาก code (เฉพาะที่ยัง active และไม่ถูกลบ)
func (r *UserRepository) SchoolIDByCode(ctx context.Context, code string) (string, error) {
	const q = `
		SELECT id FROM schools
		WHERE code = $1 AND is_active = TRUE AND deleted_at IS NULL`
	var id string
	err := r.db.QueryRow(ctx, q, code).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("repository: school by code: %w", err)
	}
	return id, nil
}

// UserByUsername หา user ในโรงเรียน (scope ด้วย school_id เสมอ)
func (r *UserRepository) UserByUsername(ctx context.Context, schoolID, username string) (*domain.User, error) {
	const q = `
		SELECT id, school_id, username, password_hash, role, is_school_admin, is_active
		FROM users
		WHERE school_id = $1 AND username = $2 AND deleted_at IS NULL`
	return r.scanUser(ctx, q, schoolID, username)
}

// UserByID หา user ตาม id ภายในโรงเรียน (scope ด้วย school_id เสมอ — กันข้ามโรงเรียน)
func (r *UserRepository) UserByID(ctx context.Context, schoolID, userID string) (*domain.User, error) {
	const q = `
		SELECT id, school_id, username, password_hash, role, is_school_admin, is_active
		FROM users
		WHERE school_id = $1 AND id = $2 AND deleted_at IS NULL`
	return r.scanUser(ctx, q, schoolID, userID)
}

func (r *UserRepository) scanUser(ctx context.Context, q string, args ...any) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx, q, args...).Scan(
		&u.ID, &u.SchoolID, &u.Username, &u.PasswordHash, &u.Role, &u.IsSchoolAdmin, &u.IsActive,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: scan user: %w", err)
	}
	return &u, nil
}

// CreateKiosk สร้างบัญชี role 'kiosk' (เครื่องสแกนหน้า) — password hash มาจาก service
func (r *UserRepository) CreateKiosk(ctx context.Context, schoolID, username, passwordHash string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (school_id, username, password_hash, role) VALUES ($1, $2, $3, 'kiosk') RETURNING id`,
		schoolID, username, passwordHash).Scan(&id)
	if err != nil {
		return "", mapUniqueViolation(err)
	}
	return id, nil
}

// ListKiosk คืนบัญชี kiosk ทั้งหมดของโรงเรียน (ที่ยังไม่ถูกลบ)
func (r *UserRepository) ListKiosk(ctx context.Context, schoolID string) ([]domain.UserBrief, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, username, is_active, created_at::text FROM users
		 WHERE school_id = $1 AND role = 'kiosk' AND deleted_at IS NULL ORDER BY created_at`,
		schoolID)
	if err != nil {
		return nil, fmt.Errorf("repository: list kiosk users: %w", err)
	}
	defer rows.Close()

	var out []domain.UserBrief
	for rows.Next() {
		var u domain.UserBrief
		if err := rows.Scan(&u.ID, &u.Username, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan kiosk user: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// DeleteKiosk ลบบัญชี kiosk (soft delete) — เฉพาะ role kiosk เท่านั้น
func (r *UserRepository) DeleteKiosk(ctx context.Context, schoolID, userID string) (bool, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE users SET deleted_at = now(), updated_at = now()
		 WHERE school_id = $1 AND id = $2 AND role = 'kiosk' AND deleted_at IS NULL`,
		schoolID, userID)
	if err != nil {
		return false, fmt.Errorf("repository: delete kiosk user: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// CurrentSemesterID คืน semester ที่ active ของโรงเรียน
func (r *UserRepository) CurrentSemesterID(ctx context.Context, schoolID string) (string, error) {
	const q = `
		SELECT id FROM semesters
		WHERE school_id = $1 AND is_current = TRUE AND deleted_at IS NULL
		LIMIT 1`
	var id string
	err := r.db.QueryRow(ctx, q, schoolID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("repository: current semester: %w", err)
	}
	return id, nil
}

// WorkGroupsForUser คืนกลุ่มงานที่ user สังกัด (scope ด้วย school_id)
func (r *UserRepository) WorkGroupsForUser(ctx context.Context, schoolID, userID string) ([]domain.WorkGroupMembership, error) {
	const q = `
		SELECT wg.id, wg.code, wg.name, uwg.is_group_admin
		FROM user_work_groups uwg
		JOIN work_groups wg ON wg.id = uwg.work_group_id
		WHERE uwg.school_id = $1 AND uwg.user_id = $2
		  AND uwg.deleted_at IS NULL AND wg.deleted_at IS NULL
		ORDER BY wg.code`
	rows, err := r.db.Query(ctx, q, schoolID, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: work groups: %w", err)
	}
	defer rows.Close()

	var out []domain.WorkGroupMembership
	for rows.Next() {
		var m domain.WorkGroupMembership
		if err := rows.Scan(&m.WorkGroupID, &m.Code, &m.Name, &m.IsGroupAdmin); err != nil {
			return nil, fmt.Errorf("repository: scan work group: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: work groups rows: %w", err)
	}
	return out, nil
}

// TouchLastLogin อัปเดต last_login_at เป็นเวลาปัจจุบัน (scope ด้วย school_id)
func (r *UserRepository) TouchLastLogin(ctx context.Context, schoolID, userID string) error {
	const q = `
		UPDATE users SET last_login_at = now(), updated_at = now()
		WHERE school_id = $1 AND id = $2 AND deleted_at IS NULL`
	if _, err := r.db.Exec(ctx, q, schoolID, userID); err != nil {
		return fmt.Errorf("repository: touch last login: %w", err)
	}
	return nil
}
