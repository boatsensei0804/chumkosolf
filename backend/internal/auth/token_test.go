package auth

import (
	"testing"
	"time"
)

func TestTokenManager_IssueParseRoundtrip(t *testing.T) {
	tm := NewTokenManager("test-secret", 15*time.Minute)

	token, err := tm.Issue("user-1", Claims{
		SchoolID:      "school-1",
		Role:          "teacher",
		SemesterID:    "sem-1",
		IsSchoolAdmin: true,
	})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	claims, err := tm.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Errorf("subject = %q, want user-1", claims.Subject)
	}
	if claims.SchoolID != "school-1" {
		t.Errorf("school_id = %q, want school-1", claims.SchoolID)
	}
	if claims.Role != "teacher" {
		t.Errorf("role = %q, want teacher", claims.Role)
	}
	if claims.SemesterID != "sem-1" {
		t.Errorf("semester_id = %q, want sem-1", claims.SemesterID)
	}
	if !claims.IsSchoolAdmin {
		t.Error("is_school_admin = false, want true")
	}
}

func TestTokenManager_RejectsWrongSecret(t *testing.T) {
	issuer := NewTokenManager("secret-a", 15*time.Minute)
	verifier := NewTokenManager("secret-b", 15*time.Minute)

	token, err := issuer.Issue("user-1", Claims{SchoolID: "s1", Role: "teacher"})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if _, err := verifier.Parse(token); err == nil {
		t.Error("expected error parsing token signed with different secret")
	}
}

func TestTokenManager_RejectsExpired(t *testing.T) {
	tm := NewTokenManager("secret", -1*time.Minute) // หมดอายุไปแล้ว
	token, err := tm.Issue("user-1", Claims{SchoolID: "s1", Role: "teacher"})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if _, err := tm.Parse(token); err == nil {
		t.Error("expected error parsing expired token")
	}
}

func TestNewRefreshToken_UniqueAndLong(t *testing.T) {
	a, err := NewRefreshToken()
	if err != nil {
		t.Fatalf("gen: %v", err)
	}
	b, err := NewRefreshToken()
	if err != nil {
		t.Fatalf("gen: %v", err)
	}
	if a == b {
		t.Error("refresh tokens should be unique")
	}
	if len(a) != 64 { // 32 bytes hex
		t.Errorf("len = %d, want 64", len(a))
	}
}
