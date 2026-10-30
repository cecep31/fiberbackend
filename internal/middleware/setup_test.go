package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"fiberbackend/config"

	"github.com/gofiber/fiber/v3"
)

type envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

func testConfig() *config.Config {
	return &config.Config{HTTP: config.HTTPConfig{AllowOrigins: []string{"*"}}}
}

// newTestApp builds the production app config + global middleware with one
// route, optionally overriding the body limit to keep 413 tests small.
func newTestApp(t *testing.T, conf *config.Config, bodyLimit int) *fiber.App {
	t.Helper()
	cfg := AppConfig(conf)
	if bodyLimit > 0 {
		cfg.BodyLimit = bodyLimit
	}
	app := fiber.New(cfg)
	InitMiddleware(app, conf)
	app.Get("/api/ping", func(c fiber.Ctx) error {
		return c.SendString(c.IP())
	})
	app.Post("/api/ping", func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})
	return app
}

func decodeEnvelope(t *testing.T, resp *http.Response) envelope {
	t.Helper()
	if ct := resp.Header.Get(fiber.HeaderContentType); !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
		t.Fatalf("Content-Type = %q, want JSON", ct)
	}
	var body envelope
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return body
}

func TestUnknownRouteReturnsEnvelope404(t *testing.T) {
	app := newTestApp(t, testConfig(), 0)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/does-not-exist", nil)
	req.Header.Set(fiber.HeaderOrigin, "https://example.com")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	body := decodeEnvelope(t, resp)
	if body.Success || body.Message != "Route not found" || body.Error != "Resource not found" {
		t.Fatalf("body = %+v", body)
	}
	if got := resp.Header.Get(fiber.HeaderAccessControlAllowOrigin); got == "" {
		t.Fatal("404 response is missing CORS headers")
	}
}

func TestWrongMethodReturnsEnvelope405(t *testing.T) {
	app := newTestApp(t, testConfig(), 0)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/api/ping", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", resp.StatusCode)
	}
	body := decodeEnvelope(t, resp)
	if body.Success || body.Message != "Method not allowed" || body.Error != "Method not allowed" {
		t.Fatalf("body = %+v", body)
	}
}

func TestOversizedBodyReturnsEnvelope413(t *testing.T) {
	// app.Test refuses oversized bodies itself, so go through a real listener
	// where fasthttp's MaxRequestBodySize rejects the request before routing.
	baseURL := startListener(t, newTestApp(t, testConfig(), 16))

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/api/ping", strings.NewReader(strings.Repeat("x", 64)))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set(fiber.HeaderOrigin, "https://example.com")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", resp.StatusCode)
	}
	body := decodeEnvelope(t, resp)
	if body.Success || body.Message != "Request Entity Too Large" || body.Error != "Request body too large" {
		t.Fatalf("body = %+v", body)
	}
	if got := resp.Header.Get(fiber.HeaderAccessControlAllowOrigin); got == "" {
		t.Fatal("413 response is missing CORS headers")
	}
}

func TestHandlerErrorReturnsGenericEnvelope500(t *testing.T) {
	app := fiber.New(AppConfig(testConfig()))
	app.Get("/boom", func(fiber.Ctx) error {
		return io.ErrUnexpectedEOF
	})

	resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/boom", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	body := decodeEnvelope(t, resp)
	if body.Success || body.Message != "Internal server error" || body.Error != "" {
		t.Fatalf("body = %+v (must not leak the error)", body)
	}
}

func TestSecureHeadersMatchEcho(t *testing.T) {
	app := newTestApp(t, testConfig(), 0)

	resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/ping", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	want := map[string]string{
		"X-Xss-Protection":        "1; mode=block",
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "SAMEORIGIN",
		"Content-Security-Policy": "default-src 'self'",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
	}
	for name, value := range want {
		if got := resp.Header.Get(name); got != value {
			t.Errorf("%s = %q, want %q", name, got, value)
		}
	}
	for _, name := range []string{
		"Strict-Transport-Security", // plain HTTP
		"Cross-Origin-Embedder-Policy",
		"Cross-Origin-Opener-Policy",
		"Cross-Origin-Resource-Policy",
		"Origin-Agent-Cluster",
		"X-Dns-Prefetch-Control",
		"X-Download-Options",
		"X-Permitted-Cross-Domain-Policies",
	} {
		if got := resp.Header.Get(name); got != "" {
			t.Errorf("unexpected header %s = %q", name, got)
		}
	}
}

func TestGlobalRateLimiterTokenBucket(t *testing.T) {
	conf := testConfig()
	conf.HTTP.RateLimitRPS = 1 // burst 2
	app := newTestApp(t, conf, 0)

	for i := range 2 {
		resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/ping", nil))
		if err != nil {
			t.Fatalf("request %d failed: %v", i+1, err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d status = %d, want 200", i+1, resp.StatusCode)
		}
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/ping", nil)
	req.Header.Set(fiber.HeaderOrigin, "https://example.com")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("limited request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", resp.StatusCode)
	}
	if got := resp.Header.Get(fiber.HeaderRetryAfter); got != "1" {
		t.Fatalf("Retry-After = %q, want 1", got)
	}
	if got := resp.Header.Get(fiber.HeaderAccessControlAllowOrigin); got == "" {
		t.Fatal("429 response is missing CORS headers")
	}
	body := decodeEnvelope(t, resp)
	if body.Success || body.Message != "Too many attempts. Please try again later." || body.Error != "Rate limit exceeded" {
		t.Fatalf("body = %+v", body)
	}
}

func TestTokenBucketStoreRefillsAndExpiresIdleVisitors(t *testing.T) {
	store := newTokenBucketStore(1, time.Minute)
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	store.lastCleanup = now

	for i := range 2 {
		if !store.allow("192.0.2.1", now) {
			t.Fatalf("request %d inside burst should be allowed", i+1)
		}
	}
	if store.allow("192.0.2.1", now) {
		t.Fatal("request beyond burst should be denied")
	}
	if !store.allow("192.0.2.2", now) {
		t.Fatal("other client must have its own bucket")
	}
	if !store.allow("192.0.2.1", now.Add(time.Second)) {
		t.Fatal("one token should refill after 1/rps seconds")
	}

	store.allow("192.0.2.3", now.Add(3*time.Minute))
	if _, ok := store.visitors["192.0.2.1"]; ok {
		t.Fatal("idle visitor should be evicted after the idle expiry")
	}
}

// startListener serves app on a loopback listener so the peer address is
// 127.0.0.1 (a trusted proxy hop), which app.Test's 0.0.0.0 peer is not.
func startListener(t *testing.T, app *fiber.App) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		_ = app.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true})
	}()
	t.Cleanup(func() {
		_ = app.ShutdownWithContext(context.Background())
	})
	return "http://" + ln.Addr().String()
}

func getIP(t *testing.T, baseURL string, headers map[string]string) string {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/api/ping", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

func TestClientIPWithTrustProxy(t *testing.T) {
	conf := testConfig()
	conf.HTTP.TrustProxy = true
	baseURL := startListener(t, newTestApp(t, conf, 0))

	cases := []struct {
		name string
		xff  string
		want string
	}{
		{"no header falls back to peer", "", "127.0.0.1"},
		{"single client hop", "203.0.113.7", "203.0.113.7"},
		{"rightmost untrusted hop wins over spoofed left entries", "198.51.100.1, 203.0.113.7, 10.0.0.2", "203.0.113.7"},
		{"CGNAT hop is a trusted proxy", "203.0.113.7, 100.64.1.1", "203.0.113.7"},
		{"all trusted returns leftmost", "10.0.0.9, 192.168.1.1", "10.0.0.9"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			headers := map[string]string{}
			if tc.xff != "" {
				headers[fiber.HeaderXForwardedFor] = tc.xff
			}
			if got := getIP(t, baseURL, headers); got != tc.want {
				t.Fatalf("c.IP() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestClientIPWithoutTrustProxyIgnoresHeader(t *testing.T) {
	baseURL := startListener(t, newTestApp(t, testConfig(), 0))

	got := getIP(t, baseURL, map[string]string{fiber.HeaderXForwardedFor: "203.0.113.7"})
	if got != "127.0.0.1" {
		t.Fatalf("c.IP() = %q, want peer 127.0.0.1", got)
	}
}

func TestHSTSOnlyOverForwardedHTTPSFromTrustedPeer(t *testing.T) {
	baseURL := startListener(t, newTestApp(t, testConfig(), 0))

	do := func(headers map[string]string) string {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/api/ping", nil)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		defer resp.Body.Close()
		return resp.Header.Get(fiber.HeaderStrictTransportSecurity)
	}

	if got := do(nil); got != "" {
		t.Fatalf("plain HTTP HSTS = %q, want none", got)
	}
	if got := do(map[string]string{fiber.HeaderXForwardedProto: "https"}); got != "max-age=3600; includeSubdomains" {
		t.Fatalf("forwarded HTTPS HSTS = %q", got)
	}
}
