package middleware

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"fiberbackend/config"
	"fiberbackend/internal/dto"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type mockUserService struct {
	getAdminByIDFn func(ctx context.Context, id string, deletedOnly bool) (*dto.UserResponse, error)
}

func (m *mockUserService) GetByID(ctx context.Context, id string) (*dto.UserResponse, error) {
	return nil, nil
}

func (m *mockUserService) GetAdminByID(ctx context.Context, id string, deletedOnly bool) (*dto.UserResponse, error) {
	if m.getAdminByIDFn != nil {
		return m.getAdminByIDFn(ctx, id, deletedOnly)
	}
	return nil, nil
}

func (m *mockUserService) GetMe(ctx context.Context, id string) (*dto.CurrentUserResponse, error) {
	return nil, nil
}

func (m *mockUserService) GetByUsername(ctx context.Context, username string) (*dto.UserResponse, error) {
	return nil, nil
}

func (m *mockUserService) GetUsers(ctx context.Context, offset int, limit int, deletedFilter dto.UserDeletedFilter, orderBy dto.UserListOrder, orderDesc bool) ([]*dto.UserResponse, int64, error) {
	return nil, 0, nil
}

func (m *mockUserService) Delete(ctx context.Context, id string) error { return nil }

func (m *mockUserService) Restore(ctx context.Context, id string) (*dto.UserResponse, error) {
	return nil, nil
}

func (m *mockUserService) UploadAvatar(ctx context.Context, userID string, file *multipart.FileHeader) (string, error) {
	return "", nil
}

func (m *mockUserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	return nil, nil
}

func (m *mockUserService) UpdateUser(ctx context.Context, id string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	return nil, nil
}

func newAuthMiddlewareForTest(secret string, users *mockUserService) *AuthMiddleware {
	return NewAuthMiddleware(&config.Config{
		Auth: config.AuthConfig{JWTSecret: secret},
	}, users)
}

func signTestToken(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	if _, ok := claims["exp"]; !ok {
		claims["exp"] = time.Now().Add(time.Hour).Unix()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

// doGet runs a GET against the app and returns the status code, raw body and headers.
func doGet(t *testing.T, app *fiber.App, path string, auth string) (int, string, http.Header) {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, string(raw), resp.Header
}

func decodeAPIResponseBytes(t *testing.T, raw string) response.APIResponse {
	t.Helper()
	var body response.APIResponse
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatalf("decode response: %v body=%q", err, raw)
	}
	return body
}

func TestAuth_MissingHeader(t *testing.T) {
	mw := newAuthMiddlewareForTest("test-secret", &mockUserService{})
	app := fiber.New()
	app.Get("/protected", mw.Auth(), func(c fiber.Ctx) error {
		t.Fatal("next should not run")
		return nil
	})

	status, raw, _ := doGet(t, app, "/protected", "")
	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", status, http.StatusUnauthorized)
	}

	body := decodeAPIResponseBytes(t, raw)
	if body.Success {
		t.Fatal("expected success=false")
	}
	if body.Message != "Missing authorization header" {
		t.Fatalf("message = %q", body.Message)
	}
	if body.Error != "Unauthorized access" {
		t.Fatalf("error field = %q, want standard Unauthorized access", body.Error)
	}
}

func TestAuth_MalformedHeader(t *testing.T) {
	mw := newAuthMiddlewareForTest("test-secret", &mockUserService{})
	app := fiber.New()
	app.Get("/protected", mw.Auth(), func(c fiber.Ctx) error {
		t.Fatal("next should not run")
		return nil
	})

	status, raw, _ := doGet(t, app, "/protected", "Token abc")
	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", status, http.StatusUnauthorized)
	}

	body := decodeAPIResponseBytes(t, raw)
	if body.Message != "Invalid authorization header" {
		t.Fatalf("message = %q", body.Message)
	}
	// Must not leak internal parse details
	if strings.Contains(strings.ToLower(body.Message), "bearer") {
		t.Fatalf("message leaked format detail: %q", body.Message)
	}
}

func TestAuth_InvalidTokenDoesNotLeakDetails(t *testing.T) {
	mw := newAuthMiddlewareForTest("test-secret", &mockUserService{})
	app := fiber.New()
	app.Get("/protected", mw.Auth(), func(c fiber.Ctx) error {
		t.Fatal("next should not run")
		return nil
	})

	status, raw, _ := doGet(t, app, "/protected", "Bearer not-a-valid-jwt")
	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", status, http.StatusUnauthorized)
	}

	body := decodeAPIResponseBytes(t, raw)
	if body.Message != "Invalid or expired token" {
		t.Fatalf("message = %q", body.Message)
	}
	for _, leak := range []string{"token parsing failed", "signature is invalid", "unexpected signing method", "jwt:"} {
		if strings.Contains(strings.ToLower(raw), strings.ToLower(leak)) {
			t.Fatalf("response leaked token details (%q): %s", leak, raw)
		}
	}
}

func TestAuth_ValidToken(t *testing.T) {
	const secret = "test-secret"
	mw := newAuthMiddlewareForTest(secret, &mockUserService{})

	var gotUserID string
	app := fiber.New()
	app.Get("/protected", mw.Auth(), func(c fiber.Ctx) error {
		claims, ok := c.Locals("user").(jwt.MapClaims)
		if !ok {
			t.Fatal("user claims not set")
		}
		gotUserID, _ = claims["user_id"].(string)
		return c.SendStatus(http.StatusNoContent)
	})

	token := signTestToken(t, secret, jwt.MapClaims{"user_id": "user-1"})
	status, raw, _ := doGet(t, app, "/protected", "Bearer "+token)
	if status != http.StatusNoContent {
		t.Fatalf("status = %d, want %d body=%s", status, http.StatusNoContent, raw)
	}
	if gotUserID != "user-1" {
		t.Fatalf("user_id = %q, want user-1", gotUserID)
	}
}

func TestAuthAdmin_ForbiddenUsesStandardResponse(t *testing.T) {
	const secret = "test-secret"
	falseVal := false
	users := &mockUserService{
		getAdminByIDFn: func(ctx context.Context, id string, deletedOnly bool) (*dto.UserResponse, error) {
			return &dto.UserResponse{ID: id, IsSuperAdmin: &falseVal}, nil
		},
	}
	mw := newAuthMiddlewareForTest(secret, users)

	app := fiber.New()
	app.Get("/admin", func(c fiber.Ctx) error {
		c.Locals("user", jwt.MapClaims{"user_id": "user-1"})
		return c.Next()
	}, mw.AuthAdmin(), func(c fiber.Ctx) error {
		t.Fatal("next should not run")
		return nil
	})

	status, raw, _ := doGet(t, app, "/admin", "")
	if status != http.StatusForbidden {
		t.Fatalf("status = %d, want %d body=%s", status, http.StatusForbidden, raw)
	}

	body := decodeAPIResponseBytes(t, raw)
	if body.Message != "Insufficient privileges" {
		t.Fatalf("message = %q", body.Message)
	}
	if body.Error != "Access forbidden" {
		t.Fatalf("error field = %q", body.Error)
	}
}

func TestAuthAdmin_AllowsSuperAdmin(t *testing.T) {
	trueVal := true
	users := &mockUserService{
		getAdminByIDFn: func(ctx context.Context, id string, deletedOnly bool) (*dto.UserResponse, error) {
			return &dto.UserResponse{ID: id, IsSuperAdmin: &trueVal}, nil
		},
	}
	mw := newAuthMiddlewareForTest("test-secret", users)

	app := fiber.New()
	called := false
	app.Get("/admin", func(c fiber.Ctx) error {
		c.Locals("user", jwt.MapClaims{"user_id": "admin-1"})
		return c.Next()
	}, mw.AuthAdmin(), func(c fiber.Ctx) error {
		called = true
		return c.SendStatus(http.StatusNoContent)
	})

	status, _, _ := doGet(t, app, "/admin", "")
	if !called {
		t.Fatal("next was not called")
	}
	if status != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", status, http.StatusNoContent)
	}
}

func TestAuthAdmin_IgnoresStaleSuperAdminClaim(t *testing.T) {
	falseVal := false
	users := &mockUserService{
		getAdminByIDFn: func(ctx context.Context, id string, deletedOnly bool) (*dto.UserResponse, error) {
			return &dto.UserResponse{ID: id, IsSuperAdmin: &falseVal}, nil
		},
	}
	mw := newAuthMiddlewareForTest("test-secret", users)

	app := fiber.New()
	called := false
	app.Get("/admin", func(c fiber.Ctx) error {
		// The token still claims admin, but the user has since been demoted.
		c.Locals("user", jwt.MapClaims{"user_id": "admin-1", "is_super_admin": true})
		return c.Next()
	}, mw.AuthAdmin(), func(c fiber.Ctx) error {
		called = true
		return c.SendStatus(http.StatusNoContent)
	})

	status, _, _ := doGet(t, app, "/admin", "")
	if called {
		t.Fatal("next was called for a demoted admin")
	}
	if status != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", status, http.StatusForbidden)
	}
}
