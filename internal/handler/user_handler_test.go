package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/dto"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

var (
	_ service.UserService       = (*mockUserService)(nil)
	_ service.UserFollowService = (*mockUserFollowService)(nil)
)

// mockUserService stubs service.UserService; unset hooks return zero values
// and record that the service was reached.
type mockUserService struct {
	called         bool
	uploadAvatarFn func(ctx context.Context, userID string, file *multipart.FileHeader) (string, error)
}

func (m *mockUserService) GetByID(ctx context.Context, id string) (*dto.UserResponse, error) {
	m.called = true
	return &dto.UserResponse{ID: id}, nil
}
func (m *mockUserService) GetAdminByID(ctx context.Context, id string, deletedOnly bool) (*dto.UserResponse, error) {
	m.called = true
	return &dto.UserResponse{ID: id}, nil
}
func (m *mockUserService) GetMe(ctx context.Context, id string) (*dto.CurrentUserResponse, error) {
	m.called = true
	return &dto.CurrentUserResponse{ID: id}, nil
}
func (m *mockUserService) GetByUsername(ctx context.Context, username string) (*dto.UserResponse, error) {
	m.called = true
	return &dto.UserResponse{}, nil
}
func (m *mockUserService) GetUsers(ctx context.Context, offset int, limit int, deletedFilter dto.UserDeletedFilter, orderBy dto.UserListOrder, orderDesc bool) ([]*dto.UserResponse, int64, error) {
	m.called = true
	return []*dto.UserResponse{}, 0, nil
}
func (m *mockUserService) Delete(ctx context.Context, id string) error {
	m.called = true
	return nil
}
func (m *mockUserService) Restore(ctx context.Context, id string) (*dto.UserResponse, error) {
	m.called = true
	return &dto.UserResponse{ID: id}, nil
}
func (m *mockUserService) UploadAvatar(ctx context.Context, userID string, file *multipart.FileHeader) (string, error) {
	m.called = true
	if m.uploadAvatarFn != nil {
		return m.uploadAvatarFn(ctx, userID, file)
	}
	return "users/avatars/x.png", nil
}
func (m *mockUserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	m.called = true
	return &dto.UserResponse{}, nil
}
func (m *mockUserService) UpdateUser(ctx context.Context, id string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	m.called = true
	return &dto.UserResponse{ID: id}, nil
}

// mockUserFollowService stubs service.UserFollowService; err (when set) is
// returned by every method.
type mockUserFollowService struct {
	called bool
	err    error
}

func (m *mockUserFollowService) FollowUser(ctx context.Context, followerID, followingID string) (*dto.FollowResponse, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	return &dto.FollowResponse{IsFollowing: true, Message: "Successfully followed user"}, nil
}
func (m *mockUserFollowService) UnfollowUser(ctx context.Context, followerID, followingID string) (*dto.FollowResponse, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	return &dto.FollowResponse{IsFollowing: false, Message: "Successfully unfollowed user"}, nil
}
func (m *mockUserFollowService) IsFollowing(ctx context.Context, followerID, followingID string) (bool, error) {
	m.called = true
	return false, m.err
}
func (m *mockUserFollowService) GetFollowers(ctx context.Context, userID string, limit, offset int) ([]*dto.UserResponse, int64, error) {
	m.called = true
	if m.err != nil {
		return nil, 0, m.err
	}
	return []*dto.UserResponse{}, 0, nil
}
func (m *mockUserFollowService) GetFollowing(ctx context.Context, userID string, limit, offset int) ([]*dto.UserResponse, int64, error) {
	m.called = true
	if m.err != nil {
		return nil, 0, m.err
	}
	return []*dto.UserResponse{}, 0, nil
}
func (m *mockUserFollowService) GetFollowStats(ctx context.Context, userID string) (*dto.UserFollowStats, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	return &dto.UserFollowStats{UserID: userID}, nil
}
func (m *mockUserFollowService) GetMutualFollows(ctx context.Context, userID1, userID2 string) ([]*dto.UserResponse, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	return []*dto.UserResponse{}, nil
}
func (m *mockUserFollowService) GetUserWithFollowStatus(ctx context.Context, userID, currentUserID string, includeAdminFields bool) (*dto.UserResponse, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	return &dto.UserResponse{ID: userID}, nil
}

const (
	testActorID = "0190a6a0-0000-7000-8000-000000000001"
	testOtherID = "0190a6a0-0000-7000-8000-000000000002"
)

func decodeUserResponse(t *testing.T, raw string) response.APIResponse {
	t.Helper()
	var out response.APIResponse
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("failed to decode response: %v\nbody: %s", err, raw)
	}
	return out
}

func TestUserRoutesRejectMalformedID(t *testing.T) {
	routes := []struct{ method, path string }{
		{http.MethodGet, "/api/users/not-a-uuid"},
		{http.MethodPut, "/api/users/not-a-uuid"},
		{http.MethodDelete, "/api/users/not-a-uuid"},
		{http.MethodPost, "/api/users/not-a-uuid/restore"},
		{http.MethodDelete, "/api/users/not-a-uuid/follow"},
		{http.MethodGet, "/api/users/not-a-uuid/follow-status"},
		{http.MethodGet, "/api/users/not-a-uuid/mutual-follows"},
		{http.MethodGet, "/api/users/not-a-uuid/followers"},
		{http.MethodGet, "/api/users/not-a-uuid/following"},
		{http.MethodGet, "/api/users/not-a-uuid/follow-stats"},
		// Not canonical (upper-case, one digit short) is rejected too.
		{http.MethodGet, "/api/users/0190A6A0-0000-7000-8000-00000000000/followers"},
	}
	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			users := &mockUserService{}
			follows := &mockUserFollowService{}
			body := `{"username":"alice","email":"alice@example.com"}`
			status, raw := serveUserRoute(t, users, follows, rt.method, rt.path, body, "application/json", testActorID)
			if status != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body %s", status, raw)
			}
			out := decodeUserResponse(t, raw)
			if out.Success || out.Message != MsgInvalidUserID || out.Error != apperrors.ErrInvalidUserID.Error() {
				t.Fatalf("unexpected body %+v", out)
			}
			if users.called || follows.called {
				t.Fatal("service must not be called for a malformed id")
			}
		})
	}
}

func TestFollowHandlersMapDomainErrors(t *testing.T) {
	cases := []struct {
		name       string
		method     string
		path       string
		body       string
		err        error
		wantStatus int
		wantMsg    string
		wantError  string
	}{
		{"follow unknown user", http.MethodPost, "/api/users/follow", `{"user_id":"` + testOtherID + `"}`, apperrors.ErrUserNotFound, http.StatusNotFound, "Failed to follow user", "user not found"},
		{"follow twice", http.MethodPost, "/api/users/follow", `{"user_id":"` + testOtherID + `"}`, apperrors.ErrAlreadyFollowing, http.StatusConflict, "Failed to follow user", "already following this user"},
		{"follow self", http.MethodPost, "/api/users/follow", `{"user_id":"` + testActorID + `"}`, apperrors.ErrCannotFollowSelf, http.StatusBadRequest, "Failed to follow user", "cannot follow yourself"},
		{"unfollow not followed", http.MethodDelete, "/api/users/" + testOtherID + "/follow", "", apperrors.ErrNotFollowing, http.StatusBadRequest, "Failed to unfollow user", "not following this user"},
		{"follow-stats unknown user", http.MethodGet, "/api/users/" + testOtherID + "/follow-stats", "", apperrors.ErrUserNotFound, http.StatusNotFound, "Failed to get follow statistics", "user not found"},
		{"followers db failure", http.MethodGet, "/api/users/" + testOtherID + "/followers", "", errors.New("db down"), http.StatusInternalServerError, "Failed to get followers", ""},
		{"following db failure", http.MethodGet, "/api/users/" + testOtherID + "/following", "", errors.New("db down"), http.StatusInternalServerError, "Failed to get following", ""},
		{"follow-status db failure", http.MethodGet, "/api/users/" + testOtherID + "/follow-status", "", errors.New("db down"), http.StatusInternalServerError, "Failed to check follow status", ""},
		{"mutual db failure", http.MethodGet, "/api/users/" + testOtherID + "/mutual-follows", "", errors.New("db down"), http.StatusInternalServerError, "Failed to get mutual follows", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			follows := &mockUserFollowService{err: tc.err}
			status, raw := serveUserRoute(t, &mockUserService{}, follows, tc.method, tc.path, tc.body, "application/json", testActorID)
			if status != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body %s", status, tc.wantStatus, raw)
			}
			out := decodeUserResponse(t, raw)
			if out.Success || out.Message != tc.wantMsg || out.Error != tc.wantError {
				t.Fatalf("unexpected body %+v", out)
			}
			if strings.Contains(raw, "db down") {
				t.Fatalf("internal error leaked: %s", raw)
			}
		})
	}
}

// avatarMultipart builds a well-formed multipart body whose avatar part is a
// PNG signature followed by padding bytes.
func avatarMultipart(t *testing.T, padding int) (string, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("avatar", "a.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("\x89PNG\r\n\x1a\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(bytes.Repeat([]byte{0}, padding)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.String(), w.FormDataContentType()
}

func TestUploadAvatarTooLargeFromServiceIs413(t *testing.T) {
	users := &mockUserService{
		uploadAvatarFn: func(ctx context.Context, userID string, file *multipart.FileHeader) (string, error) {
			return "", apperrors.ErrAvatarFileTooLarge
		},
	}
	body, contentType := avatarMultipart(t, 0)
	status, raw := serveUserRoute(t, users, &mockUserFollowService{}, http.MethodPost, "/api/users/me/image", body, contentType, testActorID)
	if status != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413; body %s", status, raw)
	}
	out := decodeUserResponse(t, raw)
	if out.Success || out.Message != msgAvatarTooLarge || out.Error != msgAvatarTooLarge {
		t.Fatalf("unexpected body %+v", out)
	}
}

func TestUploadAvatarOversizedBodyIs413BeforeParsing(t *testing.T) {
	users := &mockUserService{}
	body, contentType := avatarMultipart(t, maxAvatarBodyBytes)
	status, raw := serveUserRoute(t, users, &mockUserFollowService{}, http.MethodPost, "/api/users/me/image", body, contentType, testActorID)
	if status != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413; body %s", status, raw)
	}
	out := decodeUserResponse(t, raw)
	if out.Message != msgAvatarTooLarge || out.Error != msgAvatarTooLarge {
		t.Fatalf("unexpected body %+v", out)
	}
	if users.called {
		t.Fatal("service must not be called for an oversized body")
	}
}

func TestUploadAvatarSuccess(t *testing.T) {
	body, contentType := avatarMultipart(t, 0)
	status, raw := serveUserRoute(t, &mockUserService{}, &mockUserFollowService{}, http.MethodPost, "/api/users/me/image", body, contentType, testActorID)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %s", status, raw)
	}
	if out := decodeUserResponse(t, raw); out.Message != "Profile picture updated successfully" {
		t.Fatalf("unexpected body %+v", out)
	}
}

func serveUserRoute(t *testing.T, users *mockUserService, follows *mockUserFollowService, method, target, body, contentType, actorID string) (int, string) {
	t.Helper()

	uh := NewUserHandler(users, follows)
	fh := NewUserFollowHandler(follows)

	// The production body limit is 10 MB; fiber's 4 MB default would reject
	// the oversized-avatar case before the handler sees it.
	app := fiber.New(fiber.Config{BodyLimit: 10 * 1024 * 1024})
	if actorID != "" {
		app.Use(func(c fiber.Ctx) error {
			c.Locals("user", jwt.MapClaims{"user_id": actorID})
			return c.Next()
		})
	}
	app.Post("/api/users/me/image", uh.UploadAvatar)
	app.Get("/api/users/:id", uh.GetByID)
	app.Put("/api/users/:id", uh.UpdateUser)
	app.Delete("/api/users/:id", uh.DeleteUser)
	app.Post("/api/users/:id/restore", uh.RestoreUser)
	app.Post("/api/users/follow", fh.FollowUser)
	app.Delete("/api/users/:id/follow", fh.UnfollowUser)
	app.Get("/api/users/:id/follow-status", fh.CheckFollowStatus)
	app.Get("/api/users/:id/mutual-follows", fh.GetMutualFollows)
	app.Get("/api/users/:id/followers", fh.GetFollowers)
	app.Get("/api/users/:id/following", fh.GetFollowing)
	app.Get("/api/users/:id/follow-stats", fh.GetFollowStats)

	req := httptest.NewRequestWithContext(context.Background(), method, target, strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, contentType)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, string(raw)
}
