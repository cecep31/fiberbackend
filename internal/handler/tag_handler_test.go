package handler

import (
	"context"
	"net/http"
	"testing"

	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/dto"
	"fiberbackend/internal/model"
	"fiberbackend/internal/service"

	"github.com/gofiber/fiber/v3"
)

var _ service.TagService = (*mockTagService)(nil)

// mockTagService routes UpdateTag/DeleteTag through hooks; the other methods
// are unused by these tests.
type mockTagService struct {
	service.TagService
	updateTagFn func(ctx context.Context, id uint, req *dto.UpdateTagRequest) (*model.Tag, error)
	deleteTagFn func(ctx context.Context, id uint) error
}

func (m *mockTagService) UpdateTag(ctx context.Context, id uint, req *dto.UpdateTagRequest) (*model.Tag, error) {
	return m.updateTagFn(ctx, id, req)
}

func (m *mockTagService) DeleteTag(ctx context.Context, id uint) error {
	return m.deleteTagFn(ctx, id)
}

func serveTagRoute(t *testing.T, svc service.TagService, method, target, body string) (int, string) {
	t.Helper()
	h := NewTagHandler(svc)
	status, raw := serveAuthRoute(t, func(app *fiber.App, _ *AuthHandler) {
		app.Put("/api/tags/:id", h.UpdateTag)
		app.Delete("/api/tags/:id", h.DeleteTag)
	}, nil, method, target, body, "admin-1")
	return status, decodeAuthResponse(t, raw).Message
}

func TestTagHandlerUpdateAndDeleteMissingTagIs404(t *testing.T) {
	for _, sentinel := range []error{apperrors.ErrTagNotFound, apperrors.ErrInvalidTagID} {
		svc := &mockTagService{
			updateTagFn: func(ctx context.Context, id uint, req *dto.UpdateTagRequest) (*model.Tag, error) {
				return nil, sentinel
			},
			deleteTagFn: func(ctx context.Context, id uint) error { return sentinel },
		}
		if status, msg := serveTagRoute(t, svc, http.MethodPut, "/api/tags/0", `{"name":"go"}`); status != http.StatusNotFound || msg != "Tag not found" {
			t.Errorf("PUT (%v): got %d %q, want 404 Tag not found", sentinel, status, msg)
		}
		if status, msg := serveTagRoute(t, svc, http.MethodDelete, "/api/tags/0", ""); status != http.StatusNotFound || msg != "Tag not found" {
			t.Errorf("DELETE (%v): got %d %q, want 404 Tag not found", sentinel, status, msg)
		}
	}
}

func TestTagHandlerUpdateReturnsPersistedTag(t *testing.T) {
	svc := &mockTagService{
		updateTagFn: func(ctx context.Context, id uint, req *dto.UpdateTagRequest) (*model.Tag, error) {
			return &model.Tag{ID: id, Name: req.Name}, nil
		},
	}
	if status, _ := serveTagRoute(t, svc, http.MethodPut, "/api/tags/7", `{"name":"rust"}`); status != http.StatusOK {
		t.Fatalf("got %d, want 200", status)
	}
}
