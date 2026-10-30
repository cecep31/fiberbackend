package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/dto"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// mockChatService embeds the interface so any method a test does not stub
// panics, which doubles as a "service must not be called" assertion.
type mockChatService struct {
	service.ChatConversationService
	createStreamingMessageFn func(ctx context.Context, userID, conversationID string, req *dto.CreateChatMessageRequest) (*dto.ChatStreamResult, <-chan string, <-chan dto.ChatMessageResponse, <-chan error, error)
}

func (m *mockChatService) CreateStreamingMessage(ctx context.Context, userID, conversationID string, req *dto.CreateChatMessageRequest) (*dto.ChatStreamResult, <-chan string, <-chan dto.ChatMessageResponse, <-chan error, error) {
	return m.createStreamingMessageFn(ctx, userID, conversationID, req)
}

func registerChatRoutes(app *fiber.App, h *ChatConversationHandler) {
	app.Get("/api/chat/conversations/:id", h.GetConversation)
	app.Put("/api/chat/conversations/:id", h.UpdateConversation)
	app.Delete("/api/chat/conversations/:id", h.DeleteConversation)
	app.Post("/api/chat/conversations/:conversationId/messages", h.CreateMessage)
	app.Post("/api/chat/conversations/:conversationId/messages/stream", h.CreateMessageStream)
	app.Get("/api/chat/conversations/:conversationId/messages", h.GetMessages)
	app.Get("/api/chat/messages/:messageId", h.GetMessage)
	app.Delete("/api/chat/messages/:messageId", h.DeleteMessage)
}

func serveChatRoute(t *testing.T, svc service.ChatConversationService, method, target, body string) (int, string, response.APIResponse) {
	t.Helper()
	h := NewChatConversationHandler(svc)
	status, raw := serveAuthRoute(t, func(app *fiber.App, _ *AuthHandler) { registerChatRoutes(app, h) }, nil, method, target, body, "user-1")
	var out response.APIResponse
	_ = json.Unmarshal([]byte(raw), &out) // SSE bodies are not JSON; callers check raw then
	return status, raw, out
}

const validChatID = "0190a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a5b"

func TestChatHandlerRejectsMalformedIDs(t *testing.T) {
	cases := []struct {
		method, target, body, want string
	}{
		{http.MethodGet, "/api/chat/conversations/abc", "", MsgInvalidConversationID},
		{http.MethodPut, "/api/chat/conversations/abc", `{"title":"x"}`, MsgInvalidConversationID},
		{http.MethodDelete, "/api/chat/conversations/abc", "", MsgInvalidConversationID},
		{http.MethodPost, "/api/chat/conversations/abc/messages", `{"content":"hi"}`, MsgInvalidConversationID},
		{http.MethodPost, "/api/chat/conversations/abc/messages/stream", `{"content":"hi"}`, MsgInvalidConversationID},
		{http.MethodGet, "/api/chat/conversations/abc/messages", "", MsgInvalidConversationID},
		{http.MethodGet, "/api/chat/messages/abc", "", MsgInvalidMessageID},
		{http.MethodDelete, "/api/chat/messages/abc", "", MsgInvalidMessageID},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.target, func(t *testing.T) {
			status, _, out := serveChatRoute(t, &mockChatService{}, tc.method, tc.target, tc.body)
			if status != http.StatusBadRequest || out.Message != tc.want || out.Success {
				t.Fatalf("got %d %q, want 400 %q", status, out.Message, tc.want)
			}
		})
	}
}

// streamWith returns service channels that deliver chunks, then err (if any)
// or the completed message, then close — like the real producer goroutine.
func streamWith(chunks []string, streamErr error, done *dto.ChatMessageResponse) (<-chan string, <-chan dto.ChatMessageResponse, <-chan error) {
	chunkCh := make(chan string, len(chunks))
	for _, c := range chunks {
		chunkCh <- c
	}
	close(chunkCh)
	complete := make(chan dto.ChatMessageResponse, 1)
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

func TestChatMessageStreamEndsWithDoneAfterError(t *testing.T) {
	svc := &mockChatService{
		createStreamingMessageFn: func(ctx context.Context, userID, conversationID string, req *dto.CreateChatMessageRequest) (*dto.ChatStreamResult, <-chan string, <-chan dto.ChatMessageResponse, <-chan error, error) {
			chunks, complete, errCh := streamWith([]string{"Hel"}, apperrors.ErrChatAIResponseSaveFailed, nil)
			return &dto.ChatStreamResult{UserMessage: &dto.ChatMessageResponse{ID: "m-1", Role: "user"}}, chunks, complete, errCh, nil
		},
	}
	status, raw, _ := serveChatRoute(t, svc, http.MethodPost, "/api/chat/conversations/"+validChatID+"/messages/stream", `{"content":"hi"}`)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %s", status, raw)
	}
	wantOrder := []string{`"type":"user_message"`, `"type":"ai_chunk"`, `{"data":"Failed to save AI response","type":"error"}`, "data: [DONE]\n\n"}
	pos := 0
	for _, w := range wantOrder {
		i := strings.Index(raw[pos:], w)
		if i < 0 {
			t.Fatalf("missing %q after offset %d in stream:\n%s", w, pos, raw)
		}
		pos += i + len(w)
	}
	if !strings.HasSuffix(raw, "data: [DONE]\n\n") || strings.Contains(raw, "ai_complete") {
		t.Fatalf("stream must end with [DONE] and carry no ai_complete:\n%s", raw)
	}
}

func TestChatMessageStreamEndsWithDoneAfterComplete(t *testing.T) {
	svc := &mockChatService{
		createStreamingMessageFn: func(ctx context.Context, userID, conversationID string, req *dto.CreateChatMessageRequest) (*dto.ChatStreamResult, <-chan string, <-chan dto.ChatMessageResponse, <-chan error, error) {
			chunks, complete, errCh := streamWith([]string{"Hi"}, nil, &dto.ChatMessageResponse{ID: "m-2", Role: "assistant", Content: "Hi"})
			return &dto.ChatStreamResult{UserMessage: &dto.ChatMessageResponse{ID: "m-1", Role: "user"}}, chunks, complete, errCh, nil
		},
	}
	status, raw, _ := serveChatRoute(t, svc, http.MethodPost, "/api/chat/conversations/"+validChatID+"/messages/stream", `{"content":"hi"}`)
	if status != http.StatusOK || !strings.Contains(raw, `"type":"ai_complete"`) || !strings.HasSuffix(raw, "data: [DONE]\n\n") {
		t.Fatalf("got %d:\n%s", status, raw)
	}
}

func TestChatMessageStreamFallbackIs201(t *testing.T) {
	svc := &mockChatService{
		createStreamingMessageFn: func(ctx context.Context, userID, conversationID string, req *dto.CreateChatMessageRequest) (*dto.ChatStreamResult, <-chan string, <-chan dto.ChatMessageResponse, <-chan error, error) {
			return &dto.ChatStreamResult{UserMessage: &dto.ChatMessageResponse{ID: "m-1", Role: "user"}}, nil, nil, nil, nil
		},
	}
	status, _, out := serveChatRoute(t, svc, http.MethodPost, "/api/chat/conversations/"+validChatID+"/messages/stream", `{"content":"hi"}`)
	if status != http.StatusCreated || out.Message != "Message created successfully" {
		t.Fatalf("got %d %q, want 201 fallback", status, out.Message)
	}
}

func TestStreamErrorMessage(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{apperrors.ErrChatAIResponseSaveFailed, "Failed to save AI response"},
		{errors.New(`OpenRouter API error: 429 {"error":{"message":"Rate limited"}}`), "Rate limited"},
		{errors.New("OpenRouter returned an empty response"), "OpenRouter returned an empty response"},
	}
	for _, tc := range cases {
		if got := streamErrorMessage(tc.err); got != tc.want {
			t.Fatalf("streamErrorMessage(%v) = %q, want %q", tc.err, got, tc.want)
		}
	}
}

func TestChatMessageStreamRejectsInvalidBody(t *testing.T) {
	status, _, out := serveChatRoute(t, &mockChatService{}, http.MethodPost, "/api/chat/conversations/"+validChatID+"/messages/stream", `{}`)
	if status != http.StatusUnprocessableEntity || out.Message != "Validation failed" {
		t.Fatalf("got %d %q, want 422 (and no service call)", status, out.Message)
	}
}

// TestChatMessageStreamCancelsServiceContext: fiber's c.Context() is never
// cancelled, so the handler must give the service its own ctx and cancel it
// once the stream writer ends; otherwise a disconnected client leaves the
// producer goroutine and the upstream request running forever.
func TestChatMessageStreamCancelsServiceContext(t *testing.T) {
	ctxCh := make(chan context.Context, 1)
	svc := &mockChatService{
		createStreamingMessageFn: func(ctx context.Context, userID, conversationID string, req *dto.CreateChatMessageRequest) (*dto.ChatStreamResult, <-chan string, <-chan dto.ChatMessageResponse, <-chan error, error) {
			ctxCh <- ctx
			chunks, complete, errCh := streamWith([]string{"Hi"}, nil, &dto.ChatMessageResponse{ID: "m-2", Role: "assistant"})
			return &dto.ChatStreamResult{UserMessage: &dto.ChatMessageResponse{ID: "m-1", Role: "user"}}, chunks, complete, errCh, nil
		},
	}
	status, raw, _ := serveChatRoute(t, svc, http.MethodPost, "/api/chat/conversations/"+validChatID+"/messages/stream", `{"content":"hi"}`)
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

// TestChatMessageStreamDisconnectStopsProducer drops the client mid-stream on
// a real listener: the service ctx must be cancelled so a producer blocked on
// an unread chunk can exit.
func TestChatMessageStreamDisconnectStopsProducer(t *testing.T) {
	producerDone := make(chan struct{})
	svc := &mockChatService{
		createStreamingMessageFn: func(ctx context.Context, userID, conversationID string, req *dto.CreateChatMessageRequest) (*dto.ChatStreamResult, <-chan string, <-chan dto.ChatMessageResponse, <-chan error, error) {
			chunks := make(chan string)
			complete := make(chan dto.ChatMessageResponse, 1)
			errCh := make(chan error, 1)
			go func() {
				defer close(producerDone)
				defer close(chunks)
				defer close(complete)
				defer close(errCh)
				for {
					select {
					case <-ctx.Done():
						errCh <- ctx.Err()
						return
					case chunks <- strings.Repeat("x", 1024):
					}
				}
			}()
			return &dto.ChatStreamResult{UserMessage: &dto.ChatMessageResponse{ID: "m-1", Role: "user"}}, chunks, complete, errCh, nil
		},
	}

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("user", jwt.MapClaims{"user_id": "user-1"})
		return c.Next()
	})
	registerChatRoutes(app, NewChatConversationHandler(svc))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() { _ = app.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true}) }()
	t.Cleanup(func() { _ = app.ShutdownWithContext(context.Background()) })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+ln.Addr().String()+"/api/chat/conversations/"+validChatID+"/messages/stream", strings.NewReader(`{"content":"hi"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST stream: %v", err)
	}
	buf := make([]byte, 4096)
	if _, err := res.Body.Read(buf); err != nil {
		t.Fatalf("read first frame: %v", err)
	}
	_ = res.Body.Close() // client disconnects mid-stream

	select {
	case <-producerDone:
	case <-time.After(4 * time.Second):
		t.Fatal("producer still blocked after the client disconnected")
	}
}
