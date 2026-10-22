package handler

import (
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"fiberbackend/internal/dto"
	"fiberbackend/internal/service"

	"github.com/gofiber/fiber/v3"
)

var _ service.PostService = (*mockPostService)(nil)

type mockPostService struct {
	getPostsForSitemapFn func(ctx context.Context, limit int) ([]*dto.SitemapPost, error)
}

func (m *mockPostService) GetPosts(ctx context.Context, limit int, offset int) ([]*dto.PostResponse, int64, error) {
	return nil, 0, nil
}
func (m *mockPostService) GetPostsFiltered(ctx context.Context, filter *dto.PostQueryFilter) ([]*dto.PostResponse, int64, error) {
	return nil, 0, nil
}
func (m *mockPostService) GetPostsByUsername(ctx context.Context, username string, offset int, limit int) ([]*dto.PostResponse, int64, error) {
	return nil, 0, nil
}
func (m *mockPostService) GetPostsRandom(ctx context.Context, limit int) ([]*dto.PostResponse, error) {
	return nil, nil
}
func (m *mockPostService) GetPostsTrending(ctx context.Context, limit int) ([]*dto.PostResponse, error) {
	return nil, nil
}
func (m *mockPostService) GetPostByID(ctx context.Context, id string) (*dto.PostResponse, error) {
	return nil, nil
}
func (m *mockPostService) GetPostBySlugAndUsername(ctx context.Context, slug string, username string) (*dto.PostResponse, error) {
	return nil, nil
}
func (m *mockPostService) GetPostsByCreatedBy(ctx context.Context, createdBy string, offset int, limit int, search string, published *bool) ([]*dto.PostResponse, int64, error) {
	return nil, 0, nil
}
func (m *mockPostService) GetPostsByTag(ctx context.Context, tag string, limit int, offset int) ([]*dto.PostResponse, int64, error) {
	return nil, 0, nil
}
func (m *mockPostService) GetPostsForYou(ctx context.Context, userID string, offset int, limit int) ([]*dto.PostResponse, int64, error) {
	return nil, 0, nil
}
func (m *mockPostService) DeletePostByID(ctx context.Context, id string) error {
	return nil
}
func (m *mockPostService) DeleteMyPost(ctx context.Context, id string, userID string) error {
	return nil
}
func (m *mockPostService) UploadImagePosts(ctx context.Context, file *multipart.FileHeader) (string, error) {
	return "", nil
}
func (m *mockPostService) CreatePost(ctx context.Context, req *dto.CreatePostRequest, creatorID string) (*dto.PostResponse, error) {
	return nil, nil
}
func (m *mockPostService) UpdatePost(ctx context.Context, id string, req *dto.UpdatePostRequest) (*dto.PostResponse, error) {
	return nil, nil
}
func (m *mockPostService) UpdateMyPost(ctx context.Context, id string, userID string, req *dto.UpdatePostRequest) (*dto.PostResponse, error) {
	return nil, nil
}
func (m *mockPostService) IsAuthor(ctx context.Context, id string, userid string) error {
	return nil
}
func (m *mockPostService) GetPostsForSitemap(ctx context.Context, limit int) ([]*dto.SitemapPost, error) {
	if m.getPostsForSitemapFn != nil {
		return m.getPostsForSitemapFn(ctx, limit)
	}
	return nil, nil
}

func TestGetPostsForSitemap_Limits(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedLimit int
	}{
		{
			name:          "Default limit is 50000",
			query:         "",
			expectedLimit: 50000,
		},
		{
			name:          "Custom limit within bounds",
			query:         "limit=250",
			expectedLimit: 250,
		},
		{
			name:          "Custom limit exceeding max is capped at 50000",
			query:         "limit=100000",
			expectedLimit: 50000,
		},
		{
			name:          "Invalid limit falls back to default 50000",
			query:         "limit=-5",
			expectedLimit: 50000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedLimit int
			mockSvc := &mockPostService{
				getPostsForSitemapFn: func(ctx context.Context, limit int) ([]*dto.SitemapPost, error) {
					capturedLimit = limit
					return []*dto.SitemapPost{}, nil
				},
			}

			h := &PostHandler{
				postService: mockSvc,
			}

			app := fiber.New()
			app.Get("/api/posts/sitemap", h.GetPostsForSitemap)

			target := "/api/posts/sitemap"
			if tt.query != "" {
				target += "?" + tt.query
			}
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status 200, got %d", resp.StatusCode)
			}

			if capturedLimit != tt.expectedLimit {
				t.Errorf("expected limit %d, got %d", tt.expectedLimit, capturedLimit)
			}
		})
	}
}
