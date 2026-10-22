package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"fiberbackend/config"
	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/model"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

var (
	_ service.AuthService         = (*mockAuthService)(nil)
	_ service.AuthActivityService = (*mockAuthActivityService)(nil)
)

type mockAuthService struct {
	loginFn              func(ctx context.Context, identifier, password, ipAddress, userAgent string) (string, string, *model.User, error)
	getGithubOAuthURLFn  func(state string) string
	revokeUserSessionsFn func(ctx context.Context, targetUserID, actorID, ipAddress, userAgent string) (int64, error)
	revokeAllSessionsFn  func(ctx context.Context, actorID, ipAddress, userAgent string) (int64, error)
}

func (m *mockAuthService) Register(ctx context.Context, email, username, password, firstName, lastName string) (*model.User, error) {
	return nil, nil
}

func (m *mockAuthService) Login(ctx context.Context, identifier, password, ipAddress, userAgent string) (string, string, *model.User, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, identifier, password, ipAddress, userAgent)
	}
	return "", "", nil, nil
}

func (m *mockAuthService) ForgotPassword(ctx context.Context, email, ipAddress, userAgent string) error {
	return nil
}

func (m *mockAuthService) ResetPassword(ctx context.Context, token, password, ipAddress, userAgent string) error {
	return nil
}

func (m *mockAuthService) RefreshToken(ctx context.Context, refreshToken, ipAddress, userAgent string) (string, string, *model.User, error) {
	return "", "", nil, nil
}

func (m *mockAuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword, ipAddress, userAgent string) error {
	return nil
}

func (m *mockAuthService) Logout(ctx context.Context, refreshToken string) error {
	return nil
}

func (m *mockAuthService) RevokeUserSessions(ctx context.Context, targetUserID, actorID, ipAddress, userAgent string) (int64, error) {
	if m.revokeUserSessionsFn != nil {
		return m.revokeUserSessionsFn(ctx, targetUserID, actorID, ipAddress, userAgent)
	}
	return 0, nil
}

func (m *mockAuthService) RevokeAllSessions(ctx context.Context, actorID, ipAddress, userAgent string) (int64, error) {
	if m.revokeAllSessionsFn != nil {
		return m.revokeAllSessionsFn(ctx, actorID, ipAddress, userAgent)
	}
	return 0, nil
}

func (m *mockAuthService) GetProfile(ctx context.Context, userID string) (*model.User, error) {
	return nil, nil
}

func (m *mockAuthService) GetGithubOAuthURL(state string) string {
	if m.getGithubOAuthURLFn != nil {
		return m.getGithubOAuthURLFn(state)
	}
	return "https://github.com/login/oauth/authorize?state=" + state
}

func (m *mockAuthService) GetGithubToken(_ context.Context, code string) (string, error) {
	return "", nil
}

func (m *mockAuthService) SignInWithGithub(ctx context.Context, githubUser *service.GithubUser, ipAddress, userAgent string) (string, string, *model.User, error) {
	return "", "", nil, nil
}

func (m *mockAuthService) CreateOAuthExchangeCode(ctx context.Context, accessToken, refreshToken string, user *model.User) (string, error) {
	return "", nil
}

func (m *mockAuthService) ExchangeOAuthCode(ctx context.Context, code string) (string, string, *model.User, error) {
	return "", "", nil, nil
}

func (m *mockAuthService) UpdateProfile(ctx context.Context, userID, username, firstName, lastName string) (*model.User, error) {
	return nil, nil
}

func (m *mockAuthService) DeleteAccount(ctx context.Context, userID string) error {
	return nil
}

func (m *mockAuthService) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	return false, nil
}

type mockAuthActivityService struct{}

func (m *mockAuthActivityService) LogActivity(ctx context.Context, userID *string, activityType, status, ipAddress, userAgent string, errorMessage *string, metadata map[string]any) {
}

func (m *mockAuthActivityService) GetActivityLogs(ctx context.Context, userID, activityType string, limit, offset int) ([]*model.AuthActivityLog, int64, error) {
	return nil, 0, nil
}

func (m *mockAuthActivityService) GetRecentActivity(ctx context.Context, userID string, limit int) ([]*model.AuthActivityLog, error) {
	return nil, nil
}

func (m *mockAuthActivityService) GetFailedLogins(ctx context.Context, since time.Time, limit, offset int) ([]*model.AuthActivityLog, int64, error) {
	return nil, 0, nil
}

// postJSON drives a request through the app and returns the status code plus
// the raw response body.
func postJSON(t *testing.T, app *fiber.App, method, target, body string) (int, string, http.Header) {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), method, target, bytes.NewBufferString(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
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

func decodeAuthResponse(t *testing.T, raw string) response.APIResponse {
	t.Helper()
	var out response.APIResponse
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("failed to decode response: %v\nbody: %s", err, raw)
	}
	return out
}

func TestAuthHandlerLoginSuccess(t *testing.T) {
	username := "cecep"
	h := NewAuthHandler(&mockAuthService{
		loginFn: func(ctx context.Context, identifier, password, ipAddress, userAgent string) (string, string, *model.User, error) {
			if identifier != "cecep" || password != "secret123" {
				t.Fatalf("unexpected credentials %q/%q", identifier, password)
			}
			return "access-token", "refresh-token", &model.User{
				ID:       "user-1",
				Email:    "cecep@example.com",
				Username: &username,
			}, nil
		},
	}, &mockAuthActivityService{}, config.FrontendConfig{})

	app := fiber.New()
	app.Post("/api/auth/login", h.Login)

	status, raw, _ := postJSON(t, app, http.MethodPost, "/api/auth/login", `{"identifier":"cecep","password":"secret123"}`)
	if status != http.StatusOK {
		t.Fatalf("status=%d body=%s", status, raw)
	}
	out := decodeAuthResponse(t, raw)
	if !out.Success || out.Message != "Login successful" {
		t.Fatalf("unexpected response: %+v", out)
	}
}

func TestAuthHandlerLoginInvalidCredentials(t *testing.T) {
	h := NewAuthHandler(&mockAuthService{
		loginFn: func(ctx context.Context, identifier, password, ipAddress, userAgent string) (string, string, *model.User, error) {
			return "", "", nil, apperrors.ErrInvalidCredentials
		},
	}, &mockAuthActivityService{}, config.FrontendConfig{})

	app := fiber.New()
	app.Post("/api/auth/login", h.Login)

	status, raw, _ := postJSON(t, app, http.MethodPost, "/api/auth/login", `{"identifier":"cecep","password":"secret123"}`)
	if status != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", status, raw)
	}
	out := decodeAuthResponse(t, raw)
	if out.Success || out.Message != "Invalid identifier or password" {
		t.Fatalf("unexpected response: %+v", out)
	}
}

func TestAuthHandlerGetProfileRequiresUser(t *testing.T) {
	h := NewAuthHandler(&mockAuthService{}, &mockAuthActivityService{}, config.FrontendConfig{})
	app := fiber.New()
	app.Get("/api/auth/profile", h.GetProfile)

	status, raw, _ := postJSON(t, app, http.MethodGet, "/api/auth/profile", "")
	if status != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", status, raw)
	}
}

func TestAuthHandlerGithubOAuthRedirectSetsStateCookie(t *testing.T) {
	var stateFromService string
	h := NewAuthHandler(&mockAuthService{
		getGithubOAuthURLFn: func(state string) string {
			stateFromService = state
			return "https://github.com/login/oauth/authorize?state=" + state
		},
	}, &mockAuthActivityService{}, config.FrontendConfig{URL: "https://pilput.net"})

	app := fiber.New()
	app.Get("/api/auth/oauth/github", h.GithubOAuthRedirect)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/auth/oauth/github", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	if stateFromService == "" {
		t.Fatal("expected generated OAuth state to be passed to auth service")
	}
	if location := resp.Header.Get(fiber.HeaderLocation); !strings.Contains(location, stateFromService) {
		t.Fatalf("redirect location %q does not contain state %q", location, stateFromService)
	}
	if cookie := resp.Cookies()[0]; cookie.Name != "github_oauth_state" || cookie.Value != stateFromService || !cookie.HttpOnly || !cookie.Secure {
		t.Fatalf("unexpected oauth cookie: %+v", cookie)
	}
}

func TestAppendQueryParamPreservesExistingQuery(t *testing.T) {
	got := appendQueryParam("https://pilput.net/auth/callback?from=github", "code", "oc_123")

	if !strings.Contains(got, "from=github") || !strings.Contains(got, "code=oc_123") {
		t.Fatalf("query params not preserved, got %q", got)
	}
}

// serveAuthRoute registers one route and drives a request through the router,
// which is the only way a handler reading c.Params sees its path parameters.
// actorID, when non-empty, stands in for what the auth middleware would set.
func serveAuthRoute(
	t *testing.T,
	register func(app *fiber.App, h *AuthHandler),
	h *AuthHandler,
	method, target, body, actorID string,
) (int, string) {
	t.Helper()

	app := fiber.New()
	if actorID != "" {
		app.Use(func(c fiber.Ctx) error {
			c.Locals("user", jwt.MapClaims{"user_id": actorID})
			return c.Next()
		})
	}
	register(app, h)

	status, raw, _ := postJSON(t, app, method, target, body)
	return status, raw
}

func registerRevokeUser(app *fiber.App, h *AuthHandler) {
	app.Delete("/api/auth/sessions/:userId", h.RevokeUserSessions)
}

func registerRevokeAll(app *fiber.App, h *AuthHandler) {
	app.Post("/api/auth/sessions/revoke-all", h.RevokeAllSessions)
}

func TestAuthHandlerRevokeUserSessions(t *testing.T) {
	var gotTarget, gotActor string
	h := NewAuthHandler(&mockAuthService{
		revokeUserSessionsFn: func(ctx context.Context, targetUserID, actorID, ipAddress, userAgent string) (int64, error) {
			gotTarget, gotActor = targetUserID, actorID
			return 3, nil
		},
	}, &mockAuthActivityService{}, config.FrontendConfig{})

	status, raw := serveAuthRoute(t, registerRevokeUser, h, http.MethodDelete, "/api/auth/sessions/user-9", "", "admin-1")

	if status != http.StatusOK {
		t.Fatalf("status=%d body=%s", status, raw)
	}
	if gotTarget != "user-9" {
		t.Errorf("expected the path parameter to reach the service, got %q", gotTarget)
	}
	if gotActor != "admin-1" {
		t.Errorf("expected the acting admin to be passed through, got %q", gotActor)
	}

	out := decodeAuthResponse(t, raw)
	data, ok := out.Data.(map[string]any)
	if !ok {
		t.Fatalf("unexpected data payload: %#v", out.Data)
	}
	if data["revoked_sessions"] != float64(3) {
		t.Errorf("expected the revoked count in the response, got %v", data["revoked_sessions"])
	}
}

func TestAuthHandlerRevokeUserSessionsRequiresUser(t *testing.T) {
	h := NewAuthHandler(&mockAuthService{}, &mockAuthActivityService{}, config.FrontendConfig{})
	status, raw := serveAuthRoute(t, registerRevokeUser, h, http.MethodDelete, "/api/auth/sessions/user-9", "", "")

	if status != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", status, raw)
	}
}

func TestAuthHandlerRevokeUserSessionsNotFound(t *testing.T) {
	h := NewAuthHandler(&mockAuthService{
		revokeUserSessionsFn: func(ctx context.Context, targetUserID, actorID, ipAddress, userAgent string) (int64, error) {
			return 0, apperrors.ErrUserNotFound
		},
	}, &mockAuthActivityService{}, config.FrontendConfig{})

	status, raw := serveAuthRoute(t, registerRevokeUser, h, http.MethodDelete, "/api/auth/sessions/user-9", "", "admin-1")

	if status != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", status, raw)
	}
}

func TestAuthHandlerRevokeUserSessionsInvalidID(t *testing.T) {
	h := NewAuthHandler(&mockAuthService{
		revokeUserSessionsFn: func(ctx context.Context, targetUserID, actorID, ipAddress, userAgent string) (int64, error) {
			return 0, apperrors.ErrInvalidUserID
		},
	}, &mockAuthActivityService{}, config.FrontendConfig{})

	status, raw := serveAuthRoute(t, registerRevokeUser, h, http.MethodDelete, "/api/auth/sessions/not-a-uuid", "", "admin-1")

	if status != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", status, raw)
	}
}

func TestAuthHandlerRevokeAllSessions(t *testing.T) {
	called := false
	h := NewAuthHandler(&mockAuthService{
		revokeAllSessionsFn: func(ctx context.Context, actorID, ipAddress, userAgent string) (int64, error) {
			called = true
			return 12, nil
		},
	}, &mockAuthActivityService{}, config.FrontendConfig{})

	status, raw := serveAuthRoute(t, registerRevokeAll, h, http.MethodPost, "/api/auth/sessions/revoke-all", `{"confirm":true}`, "admin-1")

	if status != http.StatusOK {
		t.Fatalf("status=%d body=%s", status, raw)
	}
	if !called {
		t.Fatal("expected the service to be called")
	}
}

func TestAuthHandlerRevokeAllSessionsRequiresConfirmation(t *testing.T) {
	// Without an explicit acknowledgement the platform-wide logout must not
	// fire, however well-formed the request is.
	for _, body := range []string{`{}`, `{"confirm":false}`} {
		h := NewAuthHandler(&mockAuthService{
			revokeAllSessionsFn: func(ctx context.Context, actorID, ipAddress, userAgent string) (int64, error) {
				t.Errorf("service must not be called for body %s", body)
				return 0, nil
			},
		}, &mockAuthActivityService{}, config.FrontendConfig{})

		status, raw := serveAuthRoute(t, registerRevokeAll, h, http.MethodPost, "/api/auth/sessions/revoke-all", body, "admin-1")
		if status != http.StatusBadRequest {
			t.Errorf("body %s: status=%d body=%s", body, status, raw)
		}
	}
}

func TestAuthHandlerRevokeAllSessionsRequiresUser(t *testing.T) {
	h := NewAuthHandler(&mockAuthService{}, &mockAuthActivityService{}, config.FrontendConfig{})
	status, raw := serveAuthRoute(t, registerRevokeAll, h, http.MethodPost, "/api/auth/sessions/revoke-all", `{"confirm":true}`, "")

	if status != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", status, raw)
	}
}
