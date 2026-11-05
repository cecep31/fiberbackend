package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fiberbackend/config"
	"fiberbackend/internal/handler"
	"fiberbackend/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

func TestGenerateArticleRouteRequiresAuth(t *testing.T) {
	conf := &config.Config{Auth: config.AuthConfig{JWTSecret: "test-secret-test-secret-test-secret-1234"}}
	r := &Routes{
		config:         conf,
		authMiddleware: middleware.NewAuthMiddleware(conf, nil),
		aiHandler:      handler.NewAIHandler(nil), // must not be reached
	}
	app := fiber.New()
	r.setupAIRoutes(app.Group("/api"))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/ai/generate-article", strings.NewReader(`{"prompt":"hello world"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}
