package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/chumko-platform/backend/internal/auth"
	"github.com/chumko-platform/backend/internal/domain"
)

// --- fakes ---

type fakeRepo struct {
	schools   map[string]string // code -> schoolID
	users     map[string]*domain.User
	semesters map[string]string // schoolID -> semesterID
	groups    map[string][]domain.WorkGroupMembership
	touched   []string
}

// key รวม school_id เพื่อจำลองการ scope (กันข้ามโรงเรียน) ใน fake
func userKey(schoolID, key string) string { return schoolID + "/" + key }

func (f *fakeRepo) SchoolIDByCode(_ context.Context, code string) (string, error) {
	return f.schools[code], nil
}

func (f *fakeRepo) UserByUsername(_ context.Context, schoolID, username string) (*domain.User, error) {
	return f.users[userKey(schoolID, username)], nil
}

func (f *fakeRepo) UserByID(_ context.Context, schoolID, userID string) (*domain.User, error) {
	return f.users[userKey(schoolID, userID)], nil
}

func (f *fakeRepo) CurrentSemesterID(_ context.Context, schoolID string) (string, error) {
	return f.semesters[schoolID], nil
}

func (f *fakeRepo) WorkGroupsForUser(_ context.Context, schoolID, userID string) ([]domain.WorkGroupMembership, error) {
	return f.groups[userKey(schoolID, userID)], nil
}

func (f *fakeRepo) TouchLastLogin(_ context.Context, _, userID string) error {
	f.touched = append(f.touched, userID)
	return nil
}

type fakeRefreshStore struct {
	data map[string]string
}

func newFakeRefreshStore() *fakeRefreshStore { return &fakeRefreshStore{data: map[string]string{}} }

func (s *fakeRefreshStore) Save(_ context.Context, token, value string, _ time.Duration) error {
	s.data[token] = value
	return nil
}

func (s *fakeRefreshStore) Lookup(_ context.Context, token string) (string, error) {
	return s.data[token], nil
}

func (s *fakeRefreshStore) Delete(_ context.Context, token string) error {
	delete(s.data, token)
	return nil
}

// --- helpers ---

func hashPwd(t *testing.T, pwd string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	return string(h)
}

func newServiceWithUser(t *testing.T, user *domain.User, code, password string) (*AuthService, *fakeRepo, *fakeRefreshStore) {
	t.Helper()
	repo := &fakeRepo{
		schools:   map[string]string{code: user.SchoolID},
		users:     map[string]*domain.User{},
		semesters: map[string]string{user.SchoolID: "sem-current"},
		groups:    map[string][]domain.WorkGroupMembership{},
	}
	user.PasswordHash = hashPwd(t, password)
	repo.users[userKey(user.SchoolID, user.Username)] = user
	repo.users[userKey(user.SchoolID, user.ID)] = user

	store := newFakeRefreshStore()
	tm := auth.NewTokenManager("secret", 15*time.Minute)
	svc := NewAuthService(repo, tm, store, 7*24*time.Hour)
	return svc, repo, store
}

func sampleUser() *domain.User {
	return &domain.User{
		ID:            "user-1",
		SchoolID:      "school-A",
		Username:      "teacher1",
		Role:          domain.RoleTeacher,
		IsSchoolAdmin: false,
		IsActive:      true,
	}
}

// --- tests ---

func TestLogin_Success(t *testing.T) {
	svc, _, store := newServiceWithUser(t, sampleUser(), "CHUMKO", "secret123")

	res, err := svc.Login(context.Background(), "CHUMKO", "teacher1", "secret123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if res.AccessToken == "" || res.RefreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}
	if res.TokenType != "Bearer" {
		t.Errorf("token_type = %q, want Bearer", res.TokenType)
	}
	if res.User.ID != "user-1" || res.User.Role != domain.RoleTeacher {
		t.Errorf("unexpected user info: %+v", res.User)
	}
	if _, ok := store.data[res.RefreshToken]; !ok {
		t.Error("refresh token should be persisted")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, _, _ := newServiceWithUser(t, sampleUser(), "CHUMKO", "secret123")

	_, err := svc.Login(context.Background(), "CHUMKO", "teacher1", "wrong")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	svc, _, _ := newServiceWithUser(t, sampleUser(), "CHUMKO", "secret123")

	_, err := svc.Login(context.Background(), "CHUMKO", "ghost", "secret123")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_UnknownSchool(t *testing.T) {
	svc, _, _ := newServiceWithUser(t, sampleUser(), "CHUMKO", "secret123")

	_, err := svc.Login(context.Background(), "NOPE", "teacher1", "secret123")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_InactiveUser(t *testing.T) {
	u := sampleUser()
	u.IsActive = false
	svc, _, _ := newServiceWithUser(t, u, "CHUMKO", "secret123")

	_, err := svc.Login(context.Background(), "CHUMKO", "teacher1", "secret123")
	if !errors.Is(err, domain.ErrUserInactive) {
		t.Errorf("err = %v, want ErrUserInactive", err)
	}
}

// TestLogin_SchoolIsolation: ผู้ใช้ของโรงเรียน A เข้าสู่ระบบผ่าน code ของโรงเรียน B ไม่ได้
// (จำลองว่า username ซ้ำกันได้ข้ามโรงเรียน แต่ scope ด้วย school_id กันไว้)
func TestLogin_SchoolIsolation(t *testing.T) {
	svc, repo, _ := newServiceWithUser(t, sampleUser(), "SCHOOL_A", "secret123")
	// เพิ่มโรงเรียน B (code คนละค่า แต่ไม่มี user teacher1 ใน school-B)
	repo.schools["SCHOOL_B"] = "school-B"
	repo.semesters["school-B"] = "sem-B"

	_, err := svc.Login(context.Background(), "SCHOOL_B", "teacher1", "secret123")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("ผู้ใช้โรงเรียน A ไม่ควร login ผ่านโรงเรียน B ได้: err = %v", err)
	}
}

func TestRefresh_RotatesToken(t *testing.T) {
	svc, _, store := newServiceWithUser(t, sampleUser(), "CHUMKO", "secret123")
	login, err := svc.Login(context.Background(), "CHUMKO", "teacher1", "secret123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	oldRefresh := login.RefreshToken

	res, err := svc.Refresh(context.Background(), oldRefresh)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if res.RefreshToken == oldRefresh {
		t.Error("refresh token should rotate")
	}
	if _, ok := store.data[oldRefresh]; ok {
		t.Error("old refresh token should be revoked after rotation")
	}
	if _, ok := store.data[res.RefreshToken]; !ok {
		t.Error("new refresh token should be persisted")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	svc, _, _ := newServiceWithUser(t, sampleUser(), "CHUMKO", "secret123")

	_, err := svc.Refresh(context.Background(), "does-not-exist")
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestLogout_RevokesToken(t *testing.T) {
	svc, _, store := newServiceWithUser(t, sampleUser(), "CHUMKO", "secret123")
	login, err := svc.Login(context.Background(), "CHUMKO", "teacher1", "secret123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if err := svc.Logout(context.Background(), login.RefreshToken); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, ok := store.data[login.RefreshToken]; ok {
		t.Error("refresh token should be removed after logout")
	}
}

func TestMe_ReturnsUserInfo(t *testing.T) {
	svc, _, _ := newServiceWithUser(t, sampleUser(), "CHUMKO", "secret123")

	info, err := svc.Me(context.Background(), "school-A", "user-1")
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	if info.Username != "teacher1" {
		t.Errorf("username = %q, want teacher1", info.Username)
	}
	if info.WorkGroups == nil {
		t.Error("work_groups should be non-nil slice")
	}
}

func TestMe_CrossSchoolBlocked(t *testing.T) {
	svc, _, _ := newServiceWithUser(t, sampleUser(), "CHUMKO", "secret123")

	// ขอข้อมูล user-1 ภายใต้ school-B (ไม่ใช่โรงเรียนของเขา) → ไม่พบ
	_, err := svc.Me(context.Background(), "school-B", "user-1")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("err = %v, want ErrUserNotFound", err)
	}
}
