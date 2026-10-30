package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fiberbackend/internal/dto"
	"fiberbackend/internal/service"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// stubBookmarkService records ToggleBookmark calls; the other methods are not
// used by these tests.
type stubBookmarkService struct {
	service.BookmarkService
	toggled int
}

func (s *stubBookmarkService) ToggleBookmark(ctx context.Context, postID, userID string, req *dto.ToggleBookmarkRequest) (*dto.ToggleBookmarkResponse, error) {
	s.toggled++
	return &dto.ToggleBookmarkResponse{Action: "removed"}, nil
}

func TestToggleBookmark_OptionalBody(t *testing.T) {
	cases := []struct {
		name        string
		body        string
		contentType string
		wantStatus  int
		wantToggled int
	}{
		{name: "no body, no content type", wantStatus: http.StatusOK, wantToggled: 1},
		{name: "empty JSON body", contentType: "application/json", wantStatus: http.StatusOK, wantToggled: 1},
		{name: "whitespace body", body: "  \n", contentType: "application/json", wantStatus: http.StatusOK, wantToggled: 1},
		{name: "empty object", body: "{}", contentType: "application/json", wantStatus: http.StatusOK, wantToggled: 1},
		{name: "malformed JSON", body: "{", contentType: "application/json", wantStatus: http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &stubBookmarkService{}
			h := NewBookmarkHandler(svc)
			app := fiber.New()
			app.Post("/bookmarks/:post_id", func(c fiber.Ctx) error {
				c.Locals("user", jwt.MapClaims{"user_id": "user-1"})
				return h.ToggleBookmark(c)
			})

			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/bookmarks/0195f3c0-0000-7000-8000-000000000000", body)
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			_ = resp.Body.Close()
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.wantStatus)
			}
			if svc.toggled != tc.wantToggled {
				t.Fatalf("toggled = %d, want %d", svc.toggled, tc.wantToggled)
			}
		})
	}
}
