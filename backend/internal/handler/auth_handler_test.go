package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/chumko-platform/backend/internal/domain"
	"github.com/chumko-platform/backend/internal/service"
)

type fakeAuthService struct {
	loginResult *service.LoginResult
	loginErr    error
}

func (f *fakeAuthService) Login(_ context.Context, _, _, _ string) (*service.LoginResult, error) {
	return f.loginResult, f.loginErr
}
func (f *fakeAuthService) Refresh(_ context.Context, _ string) (*service.LoginResult, error) {
	return f.loginResult, f.loginErr
}
func (f *fakeAuthService) Logout(_ context.Context, _ string) error { return nil }
func (f *fakeAuthService) Me(_ context.Context, _, _ string) (*service.UserInfo, error) {
	return &service.UserInfo{ID: "user-1"}, nil
}

type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func doRequest(t *testing.T, app *fiber.App, body string) (int, envelope) {
	t.Helper()
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal %q: %v", raw, err)
	}
	return resp.StatusCode, env
}

func newApp(svc AuthService) *fiber.App {
	app := fiber.New()
	h := NewAuthHandler(svc)
	app.Post("/auth/login", h.Login)
	return app
}

func TestLoginHandler_Success(t *testing.T) {
	svc := &fakeAuthService{loginResult: &service.LoginResult{
		AccessToken: "access", RefreshToken: "refresh", TokenType: "Bearer", ExpiresIn: 900,
	}}
	app := newApp(svc)

	status, env := doRequest(t, app, `{"school_code":"CHUMKO","username":"u","password":"p"}`)
	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if !env.Success || env.Error != nil {
		t.Fatalf("expected success envelope, got %+v", env)
	}
}

func TestLoginHandler_ValidationError(t *testing.T) {
	app := newApp(&fakeAuthService{})

	status, env := doRequest(t, app, `{"school_code":"CHUMKO"}`) // ขาด username/password
	if status != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want 400", status)
	}
	if env.Success || env.Error == nil || env.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %+v", env)
	}
}

func TestLoginHandler_InvalidCredentialsMapsTo401(t *testing.T) {
	app := newApp(&fakeAuthService{loginErr: domain.ErrInvalidCredentials})

	status, env := doRequest(t, app, `{"school_code":"CHUMKO","username":"u","password":"bad"}`)
	if status != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", status)
	}
	if env.Error == nil || env.Error.Code != "INVALID_CREDENTIALS" {
		t.Fatalf("expected INVALID_CREDENTIALS, got %+v", env)
	}
}

func TestLoginHandler_BadJSON(t *testing.T) {
	app := newApp(&fakeAuthService{})

	status, _ := doRequest(t, app, `not-json`)
	if status != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want 400", status)
	}
}
