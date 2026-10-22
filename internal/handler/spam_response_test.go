package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "fiberbackend/internal/apperror"

	"github.com/gofiber/fiber/v3"
)

// The frontend reads the rejection reason from errors[].tag == "spam"; this
// pins the envelope it depends on (shared with axumbackend).
func TestRespondError_SpamRejectionEnvelope(t *testing.T) {
	app := fiber.New()
	var respErr error
	app.Post("/", func(c fiber.Ctx) error {
		err := apperrors.NewSpamRejection(apperrors.ErrContentSpamRejected, "contains banned keyword \"gacor\"")
		respErr = respondError(c, "Failed to create comment", err)
		return nil
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if respErr != nil {
		t.Fatalf("respondError returned %v", respErr)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	var body struct {
		Message string `json:"message"`
		Error   string `json:"error"`
		Errors  []struct {
			Field   string `json:"field"`
			Message string `json:"message"`
			Tag     string `json:"tag"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error != "Content rejected by anti-spam policy" {
		t.Errorf("error = %q", body.Error)
	}
	if len(body.Errors) != 1 || body.Errors[0].Tag != "spam" || body.Errors[0].Field != "text" ||
		body.Errors[0].Message != "contains banned keyword \"gacor\"" {
		t.Errorf("errors = %+v, want one spam entry for field text carrying the bare reason", body.Errors)
	}
}
