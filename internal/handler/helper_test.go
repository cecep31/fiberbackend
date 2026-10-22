package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// runInCtx runs fn inside a real Fiber request whose query string is rawQuery
// (normalized) and whose Locals have been prepared by setup.
func runInCtx(t *testing.T, rawQuery string, setup func(c fiber.Ctx), fn func(c fiber.Ctx)) {
	t.Helper()
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		if setup != nil {
			setup(c)
		}
		fn(c)
		return nil
	})

	target := "/"
	if rawQuery != "" {
		target = "/?" + rawQuery
	}
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
	// httptest.NewRequest already parses the query, but ensure normalized form
	req.URL.RawQuery = url.Values(req.URL.Query()).Encode()
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	_ = resp.Body.Close()
}

func TestGetUserIDFromClaims_NoUser(t *testing.T) {
	runInCtx(t, "", nil, func(c fiber.Ctx) {
		if id, ok := GetUserIDFromClaims(c); ok || id != "" {
			t.Fatalf("expected empty/false, got (%q,%v)", id, ok)
		}
	})
}

func TestGetUserIDFromClaims_MapClaims(t *testing.T) {
	runInCtx(t, "", func(c fiber.Ctx) {
		c.Locals("user", jwt.MapClaims{"user_id": "user-123"})
	}, func(c fiber.Ctx) {
		id, ok := GetUserIDFromClaims(c)
		if !ok || id != "user-123" {
			t.Fatalf("got (%q,%v)", id, ok)
		}
	})
}

func TestGetUserIDFromClaims_MapClaims_MissingKey(t *testing.T) {
	runInCtx(t, "", func(c fiber.Ctx) {
		c.Locals("user", jwt.MapClaims{"sub": "user-123"})
	}, func(c fiber.Ctx) {
		if _, ok := GetUserIDFromClaims(c); ok {
			t.Fatal("expected missing key to return false")
		}
	})
}

func TestGetUserIDFromClaims_MapClaims_NonString(t *testing.T) {
	runInCtx(t, "", func(c fiber.Ctx) {
		c.Locals("user", jwt.MapClaims{"user_id": 42})
	}, func(c fiber.Ctx) {
		if _, ok := GetUserIDFromClaims(c); ok {
			t.Fatal("expected non-string user_id to return false")
		}
	})
}

func TestGetUserIDFromClaims_JWTToken(t *testing.T) {
	runInCtx(t, "", func(c fiber.Ctx) {
		tok := &jwt.Token{Claims: jwt.MapClaims{"user_id": "user-456"}}
		c.Locals("user", tok)
	}, func(c fiber.Ctx) {
		id, ok := GetUserIDFromClaims(c)
		if !ok || id != "user-456" {
			t.Fatalf("got (%q,%v)", id, ok)
		}
	})
}

func TestGetUserIDFromClaims_JWTToken_BadClaims(t *testing.T) {
	runInCtx(t, "", func(c fiber.Ctx) {
		tok := &jwt.Token{Claims: jwt.RegisteredClaims{}}
		c.Locals("user", tok)
	}, func(c fiber.Ctx) {
		if _, ok := GetUserIDFromClaims(c); ok {
			t.Fatal("expected non-MapClaims to return false")
		}
	})
}

func TestGetUserIDFromClaims_PlainMap(t *testing.T) {
	runInCtx(t, "", func(c fiber.Ctx) {
		c.Locals("user", map[string]any{"user_id": "user-789"})
	}, func(c fiber.Ctx) {
		id, ok := GetUserIDFromClaims(c)
		if !ok || id != "user-789" {
			t.Fatalf("got (%q,%v)", id, ok)
		}
	})
}

func TestGetUserIDFromClaims_UnknownType(t *testing.T) {
	runInCtx(t, "", func(c fiber.Ctx) {
		c.Locals("user", "just-a-string")
	}, func(c fiber.Ctx) {
		if _, ok := GetUserIDFromClaims(c); ok {
			t.Fatal("expected unknown type to return false")
		}
	})
}

func TestParsePaginationParams(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		def        int
		wantLimit  int
		wantOffset int
	}{
		{"defaults", "", 25, 25, 0},
		{"valid both", "limit=50&offset=10", 25, 50, 10},
		{"limit cap at 100", "limit=500", 25, 100, 0},
		{"invalid limit -> default", "limit=abc", 25, 25, 0},
		{"non-positive limit -> default", "limit=0", 25, 25, 0},
		{"negative offset ignored", "offset=-5", 25, 25, 0},
		{"non-numeric offset ignored", "offset=foo", 25, 25, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInCtx(t, tt.query, nil, func(c fiber.Ctx) {
				limit, offset := ParsePaginationParams(c, tt.def)
				if limit != tt.wantLimit || offset != tt.wantOffset {
					t.Fatalf("got (%d,%d), want (%d,%d)", limit, offset, tt.wantLimit, tt.wantOffset)
				}
			})
		})
	}
}
