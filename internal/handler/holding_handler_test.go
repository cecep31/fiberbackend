package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/dto"
	"fiberbackend/internal/model"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

var _ service.HoldingService = (*mockHoldingService)(nil)

type mockHoldingService struct {
	getMonthlyDataFn func(ctx context.Context, userID string, q *dto.HoldingMonthlyQuery) ([]dto.HoldingMonthlyDataResponse, error)
}

func (m *mockHoldingService) GetHoldings(ctx context.Context, userID string, filter *dto.HoldingQueryFilter) ([]model.Holding, error) {
	return nil, nil
}

func (m *mockHoldingService) GetHoldingByID(ctx context.Context, id int64, userID string) (*model.Holding, error) {
	return nil, nil
}

func (m *mockHoldingService) CreateHolding(ctx context.Context, userID string, req *dto.CreateHoldingRequest) (*model.Holding, error) {
	return nil, nil
}

func (m *mockHoldingService) UpdateHolding(ctx context.Context, id int64, userID string, req *dto.UpdateHoldingRequest) (*model.Holding, error) {
	return nil, nil
}

func (m *mockHoldingService) DeleteHolding(ctx context.Context, id int64, userID string) error {
	return nil
}

func (m *mockHoldingService) GetHoldingTypes(ctx context.Context) ([]model.HoldingType, error) {
	return nil, nil
}

func (m *mockHoldingService) GetSummary(ctx context.Context, userID string, q *dto.HoldingSummaryQuery) (*dto.HoldingSummaryResponse, error) {
	return nil, nil
}

func (m *mockHoldingService) GetTrends(ctx context.Context, userID string, q *dto.HoldingTrendsQuery) ([]dto.HoldingTrendResponse, error) {
	return nil, nil
}

func (m *mockHoldingService) CompareMonths(ctx context.Context, userID string, q *dto.HoldingCompareQuery) (*dto.HoldingMonthComparisonResponse, error) {
	return nil, nil
}

func (m *mockHoldingService) GetMonthlyData(ctx context.Context, userID string, q *dto.HoldingMonthlyQuery) ([]dto.HoldingMonthlyDataResponse, error) {
	if m.getMonthlyDataFn != nil {
		return m.getMonthlyDataFn(ctx, userID, q)
	}
	return nil, nil
}

func (m *mockHoldingService) SyncPrices(ctx context.Context, userID string) (*dto.HoldingSyncResponse, error) {
	return nil, nil
}

func (m *mockHoldingService) DuplicateHoldings(ctx context.Context, userID string, req *dto.DuplicateHoldingRequest) ([]dto.DuplicateResultItem, error) {
	return nil, nil
}

// runHoldingGet performs a GET against h.GetMonthlyData inside a real Fiber
// request; auth sets whether the Locals user claim is present.
func runHoldingGet(t *testing.T, h *HoldingHandler, target string, auth bool) (int, string) {
	t.Helper()
	app := fiber.New()
	app.Get("/api/holdings/monthly", func(c fiber.Ctx) error {
		if auth {
			c.Locals("user", jwt.MapClaims{"user_id": "user-1"})
		}
		return c.Next()
	}, h.GetMonthlyData)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(raw)
}

func TestHoldingHandler_GetMonthlyData_NormalizesNaturalRange(t *testing.T) {
	var captured *dto.HoldingMonthlyQuery
	h := NewHoldingHandler(&mockHoldingService{
		getMonthlyDataFn: func(ctx context.Context, userID string, q *dto.HoldingMonthlyQuery) ([]dto.HoldingMonthlyDataResponse, error) {
			captured = q
			return []dto.HoldingMonthlyDataResponse{}, nil
		},
	})

	status, body := runHoldingGet(t, h, "/api/holdings/monthly?startMonth=1&startYear=2025&endMonth=12&endYear=2025", true)
	if status != http.StatusOK {
		t.Fatalf("status=%d body=%s", status, body)
	}
	if captured == nil {
		t.Fatal("expected service to be called with a query")
	}
	if captured.StartMonth != 12 || captured.StartYear != 2025 {
		t.Fatalf("expected Start to be normalized to Dec 2025, got %d/%d", captured.StartMonth, captured.StartYear)
	}
	if captured.EndMonth != 1 || captured.EndYear != 2025 {
		t.Fatalf("expected End to be normalized to Jan 2025, got %d/%d", captured.EndMonth, captured.EndYear)
	}
}

func TestHoldingHandler_GetMonthlyData_DefaultsEndTo11MonthsBeforeStart(t *testing.T) {
	var captured *dto.HoldingMonthlyQuery
	h := NewHoldingHandler(&mockHoldingService{
		getMonthlyDataFn: func(ctx context.Context, userID string, q *dto.HoldingMonthlyQuery) ([]dto.HoldingMonthlyDataResponse, error) {
			captured = q
			return []dto.HoldingMonthlyDataResponse{}, nil
		},
	})

	status, body := runHoldingGet(t, h, "/api/holdings/monthly?startMonth=5&startYear=2026", true)
	if status != http.StatusOK {
		t.Fatalf("status=%d body=%s", status, body)
	}
	if captured == nil {
		t.Fatal("expected service to be called with a query")
	}
	if captured.StartMonth != 5 || captured.StartYear != 2026 {
		t.Fatalf("expected Start to be May 2026, got %d/%d", captured.StartMonth, captured.StartYear)
	}
	if captured.EndMonth != 6 || captured.EndYear != 2025 {
		t.Fatalf("expected End to be 11 months before start (Jun 2025), got %d/%d", captured.EndMonth, captured.EndYear)
	}
}

func TestHoldingHandler_GetMonthlyData_RequiresAuth(t *testing.T) {
	h := NewHoldingHandler(&mockHoldingService{})

	status, _ := runHoldingGet(t, h, "/api/holdings/monthly", false)
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", status)
	}
}

func runParseMonthlyQuery(target string) (*dto.HoldingMonthlyQuery, error) {
	var (
		q   *dto.HoldingMonthlyQuery
		err error
	)
	app := fiber.New()
	app.Get("/api/holdings/monthly", func(c fiber.Ctx) error {
		q, err = parseMonthlyQuery(c)
		return nil
	})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
	resp, appErr := app.Test(req)
	if appErr != nil {
		panic(appErr)
	}
	_ = resp.Body.Close()
	return q, err
}

func TestParseMonthlyQuery_DefaultsToCurrentDate(t *testing.T) {
	q, err := runParseMonthlyQuery("/api/holdings/monthly")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	now := time.Now()
	if q.StartMonth != int(now.Month()) || q.StartYear != now.Year() {
		t.Fatalf("expected start to default to current date, got %+v", q)
	}
	if q.EndYear > q.StartYear || (q.EndYear == q.StartYear && q.EndMonth > q.StartMonth) {
		t.Fatalf("expected end to be on or before start, got %+v", q)
	}
}

func TestParseMonthlyQuery_DerivesPartialEndFromStart(t *testing.T) {
	q, err := runParseMonthlyQuery("/api/holdings/monthly?startMonth=5&startYear=2026&endYear=2025")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.StartMonth != 5 || q.StartYear != 2026 || q.EndMonth != 5 || q.EndYear != 2025 {
		t.Fatalf("unexpected query: %+v", q)
	}
}

func TestParseMonthlyQuery_NormalizesInvertedDifferentYear(t *testing.T) {
	q, err := runParseMonthlyQuery("/api/holdings/monthly?startMonth=1&startYear=2024&endMonth=12&endYear=2025")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.StartMonth != 12 || q.StartYear != 2025 || q.EndMonth != 1 || q.EndYear != 2024 {
		t.Fatalf("unexpected query: %+v", q)
	}
}

func TestParseMonthlyQuery_RejectsOutOfBoundsYears(t *testing.T) {
	next := strconv.Itoa(time.Now().Year() + 2)
	for _, qs := range []string{
		"startYear=abc",
		"startYear=1989",
		"startYear=" + next,
		"startYear=2147483647&endYear=1990",
		"endYear=1.5",
		"endYear=1989",
		"endYear=" + next,
	} {
		if _, err := runParseMonthlyQuery("/api/holdings/monthly?" + qs); !errors.Is(err, apperrors.ErrHoldingInvalidRange) {
			t.Errorf("%s: expected ErrHoldingInvalidRange, got %v", qs, err)
		}
	}
}

func TestParseMonthlyQuery_AcceptsYearBounds(t *testing.T) {
	maxYear := strconv.Itoa(time.Now().Year() + 1)
	q, err := runParseMonthlyQuery("/api/holdings/monthly?startMonth=1&startYear=" + maxYear + "&endMonth=1&endYear=1990")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.StartYear != time.Now().Year()+1 || q.EndYear != 1990 {
		t.Fatalf("unexpected query: %+v", q)
	}
	// Months stay lenient: garbage is ignored, not rejected.
	if _, err := runParseMonthlyQuery("/api/holdings/monthly?startMonth=abc&endMonth=13"); err != nil {
		t.Fatalf("unexpected error for lenient months: %v", err)
	}
}

func TestHoldingHandler_GetMonthlyData_InvalidYear400(t *testing.T) {
	h := NewHoldingHandler(&mockHoldingService{
		getMonthlyDataFn: func(ctx context.Context, userID string, q *dto.HoldingMonthlyQuery) ([]dto.HoldingMonthlyDataResponse, error) {
			t.Error("service must not be called for an invalid year")
			return nil, nil
		},
	})
	status, raw := runHoldingGet(t, h, "/api/holdings/monthly?startYear=1800", true)
	var body response.APIResponse
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if status != http.StatusBadRequest || body.Message != "Invalid monthly range" || body.Error != apperrors.ErrHoldingInvalidRange.Error() {
		t.Fatalf("status=%d body=%s", status, raw)
	}
}

func TestHoldingHandler_RespondHoldingError_BadRequestMappings(t *testing.T) {
	h := NewHoldingHandler(&mockHoldingService{})
	for _, tc := range []struct {
		err       error
		wantError string
	}{
		{apperrors.ErrHoldingRangeTooLarge, "requested month/year range is too large"},
		{&apperrors.HoldingInvalidDecimalError{Field: "current_value"}, "Invalid decimal value for current_value"},
	} {
		app := fiber.New()
		app.Get("/", func(c fiber.Ctx) error {
			return h.respondHoldingError(c, "Failed to get monthly data", tc.err)
		})
		resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
		if err != nil {
			t.Fatalf("app.Test failed: %v", err)
		}
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		var body response.APIResponse
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest || body.Message != "Failed to get monthly data" || body.Error != tc.wantError {
			t.Fatalf("%v: status=%d body=%s", tc.err, resp.StatusCode, raw)
		}
	}
}
