package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"fiberbackend/internal/dto"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type mockArticleService struct {
	available  bool
	generateFn func(ctx context.Context, req *dto.GenerateArticleRequest) (<-chan string, <-chan dto.GeneratedArticle, <-chan error, error)
}

func (m *mockArticleService) Available() bool { return m.available }

func (m *mockArticleService) GenerateArticleStream(ctx context.Context, req *dto.GenerateArticleRequest) (<-chan string, <-chan dto.GeneratedArticle, <-chan error, error) {
	if m.generateFn == nil {
		panic("GenerateArticleStream must not be called")
	}
	return m.generateFn(ctx, req)
}

// articleStreamWith returns service channels that deliver chunks, then err
// (if any) or the article, then close — like the real producer goroutine.
func articleStreamWith(chunks []string, streamErr error, done *dto.GeneratedArticle) (<-chan string, <-chan dto.GeneratedArticle, <-chan error) {
	chunkCh := make(chan string, len(chunks))
	for _, c := range chunks {
		chunkCh <- c
	}
	close(chunkCh)
	complete := make(chan dto.GeneratedArticle, 1)
	errCh := make(chan error, 1)
	if streamErr != nil {
		errCh <- streamErr
	} else if done != nil {
		complete <- *done
	}
	close(complete)
	close(errCh)
	return chunkCh, complete, errCh
}

func serveArticleRoute(t *testing.T, svc *mockArticleService, body, actorID string) (int, string, response.APIResponse) {
	t.Helper()
	h := NewAIHandler(svc)
	status, raw := serveAuthRoute(t, func(app *fiber.App, _ *AuthHandler) {
		app.Post("/api/ai/generate-article", h.GenerateArticle)
	}, nil, http.MethodPost, "/api/ai/generate-article", body, actorID)
	var out response.APIResponse
	_ = json.Unmarshal([]byte(raw), &out) // SSE bodies are not JSON; callers check raw then
	return status, raw, out
}

func assertInOrder(t *testing.T, raw string, want ...string) {
	t.Helper()
	pos := 0
	for _, w := range want {
		i := strings.Index(raw[pos:], w)
		if i < 0 {
			t.Fatalf("missing %q after offset %d in stream:\n%s", w, pos, raw)
		}
		pos += i + len(w)
	}
}

func TestGenerateArticleStreamsChunksThenComplete(t *testing.T) {
	var gotReq dto.GenerateArticleRequest
	svc := &mockArticleService{
		available: true,
		generateFn: func(ctx context.Context, req *dto.GenerateArticleRequest) (<-chan string, <-chan dto.GeneratedArticle, <-chan error, error) {
			gotReq = *req
			chunks, complete, errCh := articleStreamWith([]string{"<h1>T</h1>", "<p>b</p>"}, nil, &dto.GeneratedArticle{Content: "<h1>T</h1><p>b</p>", Model: "test-model"})
			return chunks, complete, errCh, nil
		},
	}
	status, raw, _ := serveArticleRoute(t, svc, `{"prompt":"  Go tips  ","tone":"casual","length":"short","title":" T ","model":"ignored"}`, "user-1")
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %s", status, raw)
	}
	assertInOrder(t, raw,
		`data: {"data":"\u003ch1\u003eT\u003c/h1\u003e","type":"ai_chunk"}`+"\n\n",
		`data: {"data":"\u003cp\u003eb\u003c/p\u003e","type":"ai_chunk"}`+"\n\n",
		`data: {"data":{"content":"\u003ch1\u003eT\u003c/h1\u003e\u003cp\u003eb\u003c/p\u003e","model":"test-model"},"type":"ai_complete"}`+"\n\n",
		"data: [DONE]\n\n",
	)
	if !strings.HasPrefix(raw, `data: {"data":`) || !strings.HasSuffix(raw, "data: [DONE]\n\n") || strings.Contains(raw, `"type":"error"`) {
		t.Fatalf("stream must start with ai_chunk, end with [DONE] and carry no error:\n%s", raw)
	}
	if gotReq.Prompt != "Go tips" || gotReq.Tone != "casual" || gotReq.Length != "short" || gotReq.Title != "T" {
		t.Fatalf("request not trimmed/bound: %+v", gotReq)
	}
}

func TestGenerateArticleStreamEndsWithDoneAfterError(t *testing.T) {
	svc := &mockArticleService{
		available: true,
		generateFn: func(ctx context.Context, req *dto.GenerateArticleRequest) (<-chan string, <-chan dto.GeneratedArticle, <-chan error, error) {
			chunks, complete, errCh := articleStreamWith([]string{"<h1>"}, errors.New(`OpenRouter API error: 429 {"error":{"message":"Rate limit exceeded"}}`), nil)
			return chunks, complete, errCh, nil
		},
	}
	status, raw, _ := serveArticleRoute(t, svc, `{"prompt":"Go tips"}`, "user-1")
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %s", status, raw)
	}
	assertInOrder(t, raw, `"type":"ai_chunk"`, `data: {"data":"Rate limit exceeded","type":"error"}`+"\n\n", "data: [DONE]\n\n")
	if !strings.HasSuffix(raw, "data: [DONE]\n\n") || strings.Contains(raw, "ai_complete") {
		t.Fatalf("stream must end with [DONE] and carry no ai_complete:\n%s", raw)
	}
}

func TestGenerateArticleUnavailableIs503(t *testing.T) {
	status, _, out := serveArticleRoute(t, &mockArticleService{available: false}, `{"prompt":"Go tips"}`, "user-1")
	if status != http.StatusServiceUnavailable || out.Success || out.Message != "AI article generation is not available" {
		t.Fatalf("got %d %+v, want 503 %q", status, out, MsgAIArticleUnavailable)
	}
}

func TestGenerateArticleRequiresUser(t *testing.T) {
	status, _, out := serveArticleRoute(t, &mockArticleService{available: true}, `{"prompt":"Go tips"}`, "")
	if status != http.StatusUnauthorized || out.Message != "User not authenticated" {
		t.Fatalf("got %d %q, want 401", status, out.Message)
	}
}

func TestGenerateArticleRejectsInvalidBody(t *testing.T) {
	cases := []struct {
		name, body string
		want       int
	}{
		{"missing prompt", `{}`, http.StatusUnprocessableEntity},
		{"prompt too short after trim", `{"prompt":"  ab  "}`, http.StatusUnprocessableEntity},
		{"prompt too long", `{"prompt":"` + strings.Repeat("a", 2001) + `"}`, http.StatusUnprocessableEntity},
		{"bad tone", `{"prompt":"Go tips","tone":"angry"}`, http.StatusUnprocessableEntity},
		{"bad length", `{"prompt":"Go tips","length":"huge"}`, http.StatusUnprocessableEntity},
		{"title too long", `{"prompt":"Go tips","title":"` + strings.Repeat("t", 201) + `"}`, http.StatusUnprocessableEntity},
		{"malformed json", `{"prompt":`, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// generateFn is nil: reaching the service would panic.
			status, raw, out := serveArticleRoute(t, &mockArticleService{available: true}, tc.body, "user-1")
			if status != tc.want || out.Success {
				t.Fatalf("got %d, want %d; body %s", status, tc.want, raw)
			}
		})
	}
}

func TestGenerateArticleCancelsServiceContext(t *testing.T) {
	ctxCh := make(chan context.Context, 1)
	svc := &mockArticleService{
		available: true,
		generateFn: func(ctx context.Context, req *dto.GenerateArticleRequest) (<-chan string, <-chan dto.GeneratedArticle, <-chan error, error) {
			ctxCh <- ctx
			chunks, complete, errCh := articleStreamWith([]string{"x"}, nil, &dto.GeneratedArticle{Content: "x", Model: "m"})
			return chunks, complete, errCh, nil
		},
	}
	status, raw, _ := serveArticleRoute(t, svc, `{"prompt":"Go tips"}`, "user-1")
	if status != http.StatusOK || !strings.HasSuffix(raw, "data: [DONE]\n\n") {
		t.Fatalf("got %d:\n%s", status, raw)
	}
	ctx := <-ctxCh
	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("service ctx not cancelled after the stream ended")
	}
}
