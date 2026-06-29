package middleware

import (
	"context"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/chumko-platform/backend/internal/auth"
	"github.com/chumko-platform/backend/internal/tenant"
)

type fakeSemVerifier struct {
	valid map[string]bool // key: schoolID|semesterID
}

func (f fakeSemVerifier) SemesterInSchool(_ context.Context, schoolID, semesterID string) (bool, error) {
	return f.valid[schoolID+"|"+semesterID], nil
}

const overrideUUID = "11111111-1111-1111-1111-111111111111"

func setupAuthApp(tm *auth.TokenManager, v SemesterVerifier) *fiber.App {
	app := fiber.New()
	app.Use(RequireAuth(tm, v))
	app.Get("/echo", func(c *fiber.Ctx) error {
		return c.SendString(tenant.SemesterIDFromContext(c.UserContext()))
	})
	return app
}

func tokenFor(t *testing.T, tm *auth.TokenManager) string {
	t.Helper()
	tok, err := tm.Issue("user-1", auth.Claims{SchoolID: "school-A", Role: "teacher", SemesterID: "sem-token"})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	return tok
}

func doReq(t *testing.T, app *fiber.App, token, override string) (int, string) {
	t.Helper()
	req := httptest.NewRequest("GET", "/echo", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	if override != "" {
		req.Header.Set("X-Semester-Id", override)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test req: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func TestRequireAuth_NoOverrideUsesToken(t *testing.T) {
	tm := auth.NewTokenManager("secret-key-secret-key", time.Hour)
	app := setupAuthApp(tm, fakeSemVerifier{valid: map[string]bool{}})
	code, body := doReq(t, app, tokenFor(t, tm), "")
	if code != fiber.StatusOK || body != "sem-token" {
		t.Errorf("code=%d body=%q ควร 200/sem-token", code, body)
	}
}

func TestRequireAuth_ValidOverride(t *testing.T) {
	tm := auth.NewTokenManager("secret-key-secret-key", time.Hour)
	app := setupAuthApp(tm, fakeSemVerifier{valid: map[string]bool{"school-A|" + overrideUUID: true}})
	code, body := doReq(t, app, tokenFor(t, tm), overrideUUID)
	if code != fiber.StatusOK || body != overrideUUID {
		t.Errorf("code=%d body=%q ควร 200 + เทอมที่ override", code, body)
	}
}

func TestRequireAuth_OverrideNotInSchool(t *testing.T) {
	tm := auth.NewTokenManager("secret-key-secret-key", time.Hour)
	// verifier ว่าง → เทอมไม่อยู่ในโรงเรียน
	app := setupAuthApp(tm, fakeSemVerifier{valid: map[string]bool{}})
	code, _ := doReq(t, app, tokenFor(t, tm), overrideUUID)
	if code != fiber.StatusForbidden {
		t.Errorf("code=%d ควร 403 (เทอมไม่อยู่ในโรงเรียน)", code)
	}
}

func TestRequireAuth_OverrideMalformed(t *testing.T) {
	tm := auth.NewTokenManager("secret-key-secret-key", time.Hour)
	app := setupAuthApp(tm, fakeSemVerifier{valid: map[string]bool{}})
	code, _ := doReq(t, app, tokenFor(t, tm), "not-a-uuid")
	if code != fiber.StatusBadRequest {
		t.Errorf("code=%d ควร 400 (รหัสเทอมผิดรูปแบบ)", code)
	}
}
