package response

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pkgvalidator "fiberbackend/pkg/validator"

	"github.com/gofiber/fiber/v3"
)

// runWith executes fn as an HTTP handler inside a real Fiber app, closes the
// response body, and returns the status code plus raw body bytes.
func runWith(t *testing.T, fn func(c fiber.Ctx) error) (int, []byte) {
	t.Helper()
	app := fiber.New()
	app.Get("/", fn)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, raw
}

func decode(t *testing.T, raw []byte) APIResponse {
	t.Helper()
	var resp APIResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, string(raw))
	}
	return resp
}

func TestSuccess(t *testing.T) {
	code, raw := runWith(t, func(c fiber.Ctx) error {
		return Success(c, "ok", map[string]string{"hello": "world"})
	})
	if code != http.StatusOK {
		t.Fatalf("status = %d, want %d", code, http.StatusOK)
	}
	body := decode(t, raw)
	if !body.Success {
		t.Errorf("expected Success=true, got false")
	}
	if body.Message != "ok" {
		t.Errorf("Message = %q, want %q", body.Message, "ok")
	}
	if body.Data == nil {
		t.Errorf("expected Data populated")
	}
}

func TestSuccessWithMeta(t *testing.T) {
	meta := PaginationMeta{TotalItems: 5, Offset: 0, Limit: 10, TotalPages: 1}
	code, raw := runWith(t, func(c fiber.Ctx) error {
		return SuccessWithMeta(c, "ok", []int{1, 2, 3}, meta)
	})
	if code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
	body := decode(t, raw)
	if body.Meta == nil {
		t.Fatalf("expected meta in response")
	}
}

func TestCreated(t *testing.T) {
	code, _ := runWith(t, func(c fiber.Ctx) error {
		return Created(c, "created", nil)
	})
	if code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", code, http.StatusCreated)
	}
}

func TestBadRequest(t *testing.T) {
	code, raw := runWith(t, func(c fiber.Ctx) error {
		return BadRequest(c, "bad input", errors.New("missing field"))
	})
	if code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", code, http.StatusBadRequest)
	}
	body := decode(t, raw)
	if body.Success {
		t.Errorf("expected Success=false")
	}
	if body.Error != "missing field" {
		t.Errorf("expected error %q, got %q", "missing field", body.Error)
	}
}

func TestBadRequest_NilError(t *testing.T) {
	_, raw := runWith(t, func(c fiber.Ctx) error {
		return BadRequest(c, "bad", nil)
	})
	body := decode(t, raw)
	if body.Error != "" {
		t.Errorf("expected empty error string when err=nil, got %q", body.Error)
	}
}

func TestUnauthorized(t *testing.T) {
	code, raw := runWith(t, func(c fiber.Ctx) error {
		return Unauthorized(c, "no token")
	})
	if code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", code, http.StatusUnauthorized)
	}
	body := decode(t, raw)
	if body.Error != "Unauthorized access" {
		t.Errorf("Error = %q", body.Error)
	}
}

func TestForbidden(t *testing.T) {
	code, _ := runWith(t, func(c fiber.Ctx) error {
		return Forbidden(c, "no rights")
	})
	if code != http.StatusForbidden {
		t.Fatalf("status = %d", code)
	}
}

func TestNotFound(t *testing.T) {
	code, raw := runWith(t, func(c fiber.Ctx) error {
		return NotFound(c, "missing", errors.New("post not found"))
	})
	if code != http.StatusNotFound {
		t.Fatalf("status = %d", code)
	}
	body := decode(t, raw)
	if body.Error != "post not found" {
		t.Errorf("Error = %q", body.Error)
	}
}

func TestNotFound_NilError(t *testing.T) {
	_, raw := runWith(t, func(c fiber.Ctx) error {
		return NotFound(c, "missing", nil)
	})
	body := decode(t, raw)
	if body.Error != "Resource not found" {
		t.Errorf("expected default error message, got %q", body.Error)
	}
}

// TestInternalServerError_DoesNotLeakError is the documented contract:
// the raw error string MUST NOT appear in the response body — only logged server-side.
func TestInternalServerError_DoesNotLeakError(t *testing.T) {
	secret := "DSN=postgres://user:hunter2@host/db"
	code, raw := runWith(t, func(c fiber.Ctx) error {
		return InternalServerError(c, "boom", errors.New(secret))
	})
	if code != http.StatusInternalServerError {
		t.Fatalf("status = %d", code)
	}
	if strings.Contains(string(raw), secret) {
		t.Fatalf("InternalServerError leaked raw error: %s", string(raw))
	}
	body := decode(t, raw)
	if body.Success {
		t.Errorf("expected Success=false")
	}
	if body.Message != "boom" {
		t.Errorf("Message = %q", body.Message)
	}
}

func TestValidationError(t *testing.T) {
	code, _ := runWith(t, func(c fiber.Ctx) error {
		return ValidationError(c, "invalid", errors.New("field x"))
	})
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", code, http.StatusUnprocessableEntity)
	}
}

func TestFromValidateError_Structured(t *testing.T) {
	verr := pkgvalidator.ValidationErrors{
		Errors: []pkgvalidator.ValidationError{
			{Field: "Email", Message: "Email must be a valid email address", Tag: "email"},
		},
	}
	code, raw := runWith(t, func(c fiber.Ctx) error {
		return FromValidateError(c, verr)
	})
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", code)
	}
	body := decode(t, raw)
	if body.Success {
		t.Errorf("expected Success=false")
	}
	if body.Errors == nil {
		t.Errorf("expected Errors field populated")
	}
}

func TestFromValidateError_Generic(t *testing.T) {
	code, _ := runWith(t, func(c fiber.Ctx) error {
		return FromValidateError(c, errors.New("plain error"))
	})
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", code)
	}
}

func TestConflict(t *testing.T) {
	code, raw := runWith(t, func(c fiber.Ctx) error {
		return Conflict(c, "duplicate", "user already exists")
	})
	if code != http.StatusConflict {
		t.Fatalf("status = %d", code)
	}
	body := decode(t, raw)
	if body.Error != "user already exists" {
		t.Errorf("Error = %q", body.Error)
	}
}

func TestCalculatePaginationMeta(t *testing.T) {
	tests := []struct {
		name       string
		total      int64
		offset     int
		limit      int
		wantPages  int
		wantLimit  int
		wantOffset int
	}{
		{"exact pages", 100, 0, 10, 10, 10, 0},
		{"with remainder", 95, 0, 10, 10, 10, 0},
		{"single page", 5, 0, 10, 1, 10, 0},
		{"zero items", 0, 0, 10, 0, 10, 0},
		{"limit zero -> default 10", 25, 0, 0, 3, 10, 0},
		{"limit negative -> default 10", 25, 0, -5, 3, 10, 0},
		{"offset preserved", 30, 20, 5, 6, 5, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := CalculatePaginationMeta(tt.total, tt.offset, tt.limit)
			if meta.TotalPages != tt.wantPages {
				t.Errorf("TotalPages = %d, want %d", meta.TotalPages, tt.wantPages)
			}
			if meta.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", meta.Limit, tt.wantLimit)
			}
			if meta.Offset != tt.wantOffset {
				t.Errorf("Offset = %d, want %d", meta.Offset, tt.wantOffset)
			}
			if int64(meta.TotalItems) != tt.total {
				t.Errorf("TotalItems = %d, want %d", meta.TotalItems, tt.total)
			}
		})
	}
}
