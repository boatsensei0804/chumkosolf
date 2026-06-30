// Package service เก็บ business logic ทั้งหมด (เรียก repository, ตรวจสิทธิ์เชิงธุรกิจ)
package service

import (
	"context"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/chumkosoft/backend/internal/auth"
	"github.com/chumkosoft/backend/internal/domain"
)

// UserRepository คือ contract ที่ repository ชั้น DB ต้องทำให้ — ทุก method scope ด้วย school_id
type UserRepository interface {
	// SchoolIDByCode resolve โรงเรียนจาก code; คืน ("", nil) ถ้าไม่พบ
	SchoolIDByCode(ctx context.Context, code string) (string, error)
	// UserByUsername หา user ในโรงเรียนนั้น; คืน (nil, nil) ถ้าไม่พบ
	UserByUsername(ctx context.Context, schoolID, username string) (*domain.User, error)
	// UserByID หา user ตาม id ภายในโรงเรียน; คืน (nil, nil) ถ้าไม่พบ
	UserByID(ctx context.Context, schoolID, userID string) (*domain.User, error)
	// CurrentSemesterID คืน semester ที่ active ของโรงเรียน; คืน ("", nil) ถ้ายังไม่มี
	CurrentSemesterID(ctx context.Context, schoolID string) (string, error)
	// WorkGroupsForUser คืนกลุ่มงานที่ user สังกัด
	WorkGroupsForUser(ctx context.Context, schoolID, userID string) ([]domain.WorkGroupMembership, error)
	// TouchLastLogin อัปเดต last_login_at (best-effort)
	TouchLastLogin(ctx context.Context, schoolID, userID string) error
}

// RefreshStore เก็บ refresh token ฝั่ง server (เพื่อเพิกถอน/หมุนเวียนได้)
// value ที่เก็บคือ "<schoolID>|<userID>"
type RefreshStore interface {
	Save(ctx context.Context, token, value string, ttl time.Duration) error
	Lookup(ctx context.Context, token string) (string, error)
	Delete(ctx context.Context, token string) error
}

// UserInfo คือข้อมูล user ที่ปลอดภัยส่งกลับ frontend (ไม่มี password hash)
type UserInfo struct {
	ID            string                       `json:"id"`
	Username      string                       `json:"username"`
	Role          string                       `json:"role"`
	SchoolID      string                       `json:"school_id"`
	IsSchoolAdmin bool                         `json:"is_school_admin"`
	WorkGroups    []domain.WorkGroupMembership `json:"work_groups"`
}

// LoginResult คือผลของการ login/refresh
type LoginResult struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"` // วินาที
	User         UserInfo `json:"user"`
}

// AuthService รวม use case ของการยืนยันตัวตน
type AuthService struct {
	repo       UserRepository
	tokens     *auth.TokenManager
	refresh    RefreshStore
	refreshTTL time.Duration
}

// NewAuthService สร้าง service
func NewAuthService(repo UserRepository, tokens *auth.TokenManager, refresh RefreshStore, refreshTTL time.Duration) *AuthService {
	return &AuthService{repo: repo, tokens: tokens, refresh: refresh, refreshTTL: refreshTTL}
}

// Login ตรวจ school code + username + password แล้วออก token
func (s *AuthService) Login(ctx context.Context, schoolCode, username, password string) (*LoginResult, error) {
	schoolID, err := s.repo.SchoolIDByCode(ctx, strings.TrimSpace(schoolCode))
	if err != nil {
		return nil, err
	}
	if schoolID == "" {
		// โรงเรียนไม่พบ — คืน invalid credentials กัน enumeration
		return nil, domain.ErrInvalidCredentials
	}

	user, err := s.repo.UserByUsername(ctx, schoolID, strings.TrimSpace(username))
	if err != nil {
		return nil, err
	}
	if user == nil {
		// เทียบ bcrypt กับ dummy hash เพื่อให้เวลาตอบใกล้เคียงกรณีมี user (กัน timing/user enumeration)
		_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(password))
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, domain.ErrUserInactive
	}

	return s.issueFor(ctx, user)
}

// Refresh หมุนเวียน refresh token: ตรวจ token เดิม → ออก access+refresh ใหม่ → ลบอันเก่า
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*LoginResult, error) {
	value, err := s.refresh.Lookup(ctx, refreshToken)
	if err != nil || value == "" {
		return nil, domain.ErrInvalidToken
	}
	schoolID, userID, ok := splitRefreshValue(value)
	if !ok {
		return nil, domain.ErrInvalidToken
	}

	user, err := s.repo.UserByID(ctx, schoolID, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		_ = s.refresh.Delete(ctx, refreshToken)
		return nil, domain.ErrInvalidToken
	}
	if !user.IsActive {
		_ = s.refresh.Delete(ctx, refreshToken)
		return nil, domain.ErrUserInactive
	}

	result, err := s.issueFor(ctx, user)
	if err != nil {
		return nil, err
	}
	// rotation: ลบ token เดิมหลังออกอันใหม่สำเร็จ
	_ = s.refresh.Delete(ctx, refreshToken)
	return result, nil
}

// Logout เพิกถอน refresh token (idempotent)
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.refresh.Delete(ctx, refreshToken)
}

// Me คืนข้อมูลผู้ใช้ปัจจุบันจาก scope ใน context (เรียกหลังผ่าน RequireAuth)
func (s *AuthService) Me(ctx context.Context, schoolID, userID string) (*UserInfo, error) {
	user, err := s.repo.UserByID(ctx, schoolID, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}
	info, err := s.toUserInfo(ctx, user)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// issueFor ออก access + refresh token สำหรับ user (ใช้ร่วมกันระหว่าง Login/Refresh)
func (s *AuthService) issueFor(ctx context.Context, user *domain.User) (*LoginResult, error) {
	semesterID, err := s.repo.CurrentSemesterID(ctx, user.SchoolID)
	if err != nil {
		return nil, err
	}

	access, err := s.tokens.Issue(user.ID, auth.Claims{
		SchoolID:      user.SchoolID,
		Role:          user.Role,
		SemesterID:    semesterID,
		IsSchoolAdmin: user.IsSchoolAdmin,
	})
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	if err := s.refresh.Save(ctx, refreshToken, user.SchoolID+"|"+user.ID, s.refreshTTL); err != nil {
		return nil, err
	}

	// best-effort: ไม่ให้ login ล้มเพราะอัปเดต last_login ไม่สำเร็จ
	if err := s.repo.TouchLastLogin(ctx, user.SchoolID, user.ID); err != nil {
		log.Printf("auth: touch last_login ล้มเหลว (user=%s): %v", user.ID, err)
	}

	info, err := s.toUserInfo(ctx, user)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  access,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.tokens.AccessTTL().Seconds()),
		User:         info,
	}, nil
}

func (s *AuthService) toUserInfo(ctx context.Context, user *domain.User) (UserInfo, error) {
	groups, err := s.repo.WorkGroupsForUser(ctx, user.SchoolID, user.ID)
	if err != nil {
		return UserInfo{}, err
	}
	if groups == nil {
		groups = []domain.WorkGroupMembership{}
	}
	return UserInfo{
		ID:            user.ID,
		Username:      user.Username,
		Role:          user.Role,
		SchoolID:      user.SchoolID,
		IsSchoolAdmin: user.IsSchoolAdmin,
		WorkGroups:    groups,
	}, nil
}

func splitRefreshValue(value string) (schoolID, userID string, ok bool) {
	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// dummyHash คือ bcrypt ของสตริงสุ่ม ใช้เทียบเวลาเมื่อไม่พบ user (constant-time-ish)
var dummyHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
