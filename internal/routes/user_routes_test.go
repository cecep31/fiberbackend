package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"fiberbackend/config"
	"fiberbackend/internal/dto"
	"fiberbackend/internal/handler"
	"fiberbackend/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

// stubFollowService answers the public follow reads with empty results.
// The other methods are unused by these tests.
type stubFollowService struct{}

func (stubFollowService) FollowUser(context.Context, string, string) (*dto.FollowResponse, error) {
	return &dto.FollowResponse{}, nil
}

func (stubFollowService) UnfollowUser(context.Context, string, string) (*dto.FollowResponse, error) {
	return &dto.FollowResponse{}, nil
}

func (stubFollowService) IsFollowing(context.Context, string, string) (bool, error) {
	return false, nil
}

func (stubFollowService) GetFollowers(context.Context, string, int, int) ([]*dto.UserResponse, int64, error) {
	return []*dto.UserResponse{}, 0, nil
}

func (stubFollowService) GetFollowing(context.Context, string, int, int) ([]*dto.UserResponse, int64, error) {
	return []*dto.UserResponse{}, 0, nil
}

func (stubFollowService) GetFollowStats(context.Context, string) (*dto.UserFollowStats, error) {
	return &dto.UserFollowStats{}, nil
}

func (stubFollowService) GetMutualFollows(context.Context, string, string) ([]*dto.UserResponse, error) {
	return []*dto.UserResponse{}, nil
}

func (stubFollowService) GetUserWithFollowStatus(context.Context, string, string, bool) (*dto.UserResponse, error) {
	return &dto.UserResponse{}, nil
}

func newUserRoutesTestApp() *fiber.App {
	conf := &config.Config{Auth: config.AuthConfig{JWTSecret: "test-secret-test-secret-test-secret-1234"}}
	r := &Routes{
		config:            conf,
		authMiddleware:    middleware.NewAuthMiddleware(conf, nil),
		userHandler:       &handler.UserHandler{},
		userStatsHandler:  &handler.UserStatsHandler{},
		userFollowHandler: handler.NewUserFollowHandler(stubFollowService{}),
	}
	app := fiber.New()
	r.setupUserRoutes(app.Group("/api"))
	return app
}

// Followers, following and follow-stats are public in echo and axum: a
// Fiber Group created with Auth() must not swallow them.
func TestPublicFollowRoutesDoNotRequireAuth(t *testing.T) {
	app := newUserRoutesTestApp()
	const id = "0190f3a0-0000-7000-8000-000000000001"
	for _, path := range []string{
		"/api/users/" + id + "/followers",
		"/api/users/" + id + "/following",
		"/api/users/" + id + "/follow-stats",
	} {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s without token = %d, want 200", path, resp.StatusCode)
		}
	}
}

// Authenticated user routes, and unknown /api/users/* paths, still answer 401
// without a token (echo's group RouteNotFound handler runs Auth() too).
func TestProtectedUserRoutesRequireAuth(t *testing.T) {
	app := newUserRoutesTestApp()
	const id = "0190f3a0-0000-7000-8000-000000000001"
	cases := []struct{ method, path string }{
		{http.MethodGet, "/api/users/me"},
		{http.MethodGet, "/api/users/stats"},
		{http.MethodGet, "/api/users"},
		{http.MethodPost, "/api/users/follow"},
		{http.MethodDelete, "/api/users/" + id + "/follow"},
		{http.MethodGet, "/api/users/" + id + "/follow-status"},
		{http.MethodGet, "/api/users/" + id + "/mutual-follows"},
		{http.MethodGet, "/api/users/" + id + "/does-not-exist"},
	}
	for _, tc := range cases {
		resp, err := app.Test(httptest.NewRequest(tc.method, tc.path, nil))
		if err != nil {
			t.Fatalf("%s %s: %v", tc.method, tc.path, err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s without token = %d, want 401", tc.method, tc.path, resp.StatusCode)
		}
	}
}
