package handler

import (
	"errors"
	"strconv"
	"time"

	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/dto"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type HoldingHandler struct {
	holdingService service.HoldingService
}

func NewHoldingHandler(holdingService service.HoldingService) *HoldingHandler {
	return &HoldingHandler{holdingService: holdingService}
}

func (h *HoldingHandler) respondHoldingError(c fiber.Ctx, message string, err error) error {
	switch {
	case errors.Is(err, apperrors.ErrHoldingNotFound):
		return response.NotFound(c, message, err)
	case errors.Is(err, apperrors.ErrHoldingNotOwned):
		return response.Forbidden(c, message)
	case errors.Is(err, apperrors.ErrHoldingTypeNotFound):
		return response.BadRequest(c, message, err)
	case errors.Is(err, apperrors.ErrHoldingDuplicateSame):
		return response.BadRequest(c, message, err)
	case errors.Is(err, apperrors.ErrHoldingInvalidRange):
		return response.BadRequest(c, message, err)
	default:
		return response.InternalServerError(c, message, err)
	}
}

func (h *HoldingHandler) GetHoldings(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	filter := &dto.HoldingQueryFilter{
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	now := time.Now()
	curMonth := int(now.Month())
	curYear := now.Year()
	filter.Month = &curMonth
	filter.Year = &curYear

	if m := c.Query("month"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v >= 1 && v <= 12 {
			filter.Month = &v
		}
	}
	if y := c.Query("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil {
			filter.Year = &v
		}
	}
	if s := c.Query("sortBy"); s != "" {
		filter.SortBy = s
	}
	if o := c.Query("order"); o != "" {
		filter.SortOrder = o
	}

	holdings, err := h.holdingService.GetHoldings(c.Context(), userID, filter)
	if err != nil {
		return h.respondHoldingError(c, "Failed to get holdings", err)
	}

	return response.Success(c, "Holdings fetched successfully", holdings)
}

func (h *HoldingHandler) GetHoldingByID(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid holding ID", err)
	}

	holding, err := h.holdingService.GetHoldingByID(c.Context(), id, userID)
	if err != nil {
		return h.respondHoldingError(c, "Failed to get holding", err)
	}

	return response.Success(c, "Holding fetched successfully", holding)
}

func (h *HoldingHandler) CreateHolding(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req dto.CreateHoldingRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request format", err)
	}

	if err := bindValidate(c, &req); err != nil {
		return err
	}

	holding, err := h.holdingService.CreateHolding(c.Context(), userID, &req)
	if err != nil {
		return h.respondHoldingError(c, "Failed to create holding", err)
	}

	return response.Created(c, "Holding created successfully", []any{holding})
}

func (h *HoldingHandler) UpdateHolding(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid holding ID", err)
	}

	var req dto.UpdateHoldingRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request format", err)
	}

	holding, err := h.holdingService.UpdateHolding(c.Context(), id, userID, &req)
	if err != nil {
		return h.respondHoldingError(c, "Failed to update holding", err)
	}

	return response.Success(c, "Holding updated successfully", []any{holding})
}

func (h *HoldingHandler) DeleteHolding(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid holding ID", err)
	}

	if err := h.holdingService.DeleteHolding(c.Context(), id, userID); err != nil {
		return h.respondHoldingError(c, "Failed to delete holding", err)
	}

	return response.Success(c, "Holding deleted successfully", nil)
}

func (h *HoldingHandler) GetHoldingTypes(c fiber.Ctx) error {
	types, err := h.holdingService.GetHoldingTypes(c.Context())
	if err != nil {
		return response.InternalServerError(c, "Failed to get holding types", err)
	}

	return response.Success(c, "Holding types fetched successfully", types)
}

func (h *HoldingHandler) GetSummary(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	q := &dto.HoldingSummaryQuery{}
	if m := c.Query("month"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v >= 1 && v <= 12 {
			q.Month = &v
		}
	}
	if y := c.Query("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil {
			q.Year = &v
		}
	}

	summary, err := h.holdingService.GetSummary(c.Context(), userID, q)
	if err != nil {
		return h.respondHoldingError(c, "Failed to get holdings summary", err)
	}

	return response.Success(c, "Holdings summary fetched successfully", summary)
}

func (h *HoldingHandler) GetTrends(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	q := &dto.HoldingTrendsQuery{}
	if y := c.Query("years"); y != "" {
		for _, ys := range splitComma(y) {
			if v, err := strconv.Atoi(ys); err == nil {
				q.Years = append(q.Years, v)
			}
		}
	}

	trends, err := h.holdingService.GetTrends(c.Context(), userID, q)
	if err != nil {
		return h.respondHoldingError(c, "Failed to get holdings trends", err)
	}

	return response.Success(c, "Holdings trends fetched successfully", trends)
}

func (h *HoldingHandler) CompareMonths(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	now := time.Now()
	q := &dto.HoldingCompareQuery{
		ToMonth: int(now.Month()),
		ToYear:  now.Year(),
	}

	if fm := c.Query("fromMonth"); fm != "" {
		if v, err := strconv.Atoi(fm); err == nil && v >= 1 && v <= 12 {
			q.FromMonth = &v
		}
	}
	if fy := c.Query("fromYear"); fy != "" {
		if v, err := strconv.Atoi(fy); err == nil {
			q.FromYear = &v
		}
	}
	if tm := c.Query("toMonth"); tm != "" {
		if v, err := strconv.Atoi(tm); err == nil && v >= 1 && v <= 12 {
			q.ToMonth = v
		}
	}
	if ty := c.Query("toYear"); ty != "" {
		if v, err := strconv.Atoi(ty); err == nil {
			q.ToYear = v
		}
	}

	switch {
	case q.FromMonth == nil && q.FromYear == nil:
		fromM, fromY := prevMonth(q.ToMonth, q.ToYear)
		q.FromMonth = &fromM
		q.FromYear = &fromY
	case q.FromMonth == nil:
		q.FromMonth = &q.ToMonth
	case q.FromYear == nil:
		q.FromYear = &q.ToYear
	}

	result, err := h.holdingService.CompareMonths(c.Context(), userID, q)
	if err != nil {
		return h.respondHoldingError(c, "Failed to compare months", err)
	}

	return response.Success(c, "Month comparison fetched successfully", result)
}

func (h *HoldingHandler) GetMonthlyData(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	q, err := parseMonthlyQuery(c)
	if err != nil {
		return response.BadRequest(c, "Invalid monthly range", err)
	}

	result, err := h.holdingService.GetMonthlyData(c.Context(), userID, q)
	if err != nil {
		return h.respondHoldingError(c, "Failed to get monthly data", err)
	}

	return response.Success(c, "Holdings monthly data fetched successfully", result)
}

func parseMonthlyQuery(c fiber.Ctx) (*dto.HoldingMonthlyQuery, error) {
	now := time.Now()
	startMonth, startYear := int(now.Month()), now.Year()
	endMonth, endYear := int(now.Month()), now.Year()

	var hasStartMonth, hasStartYear, hasEndMonth, hasEndYear bool

	if sm := c.Query("startMonth"); sm != "" {
		if v, err := strconv.Atoi(sm); err == nil && v >= 1 && v <= 12 {
			startMonth = v
			hasStartMonth = true
		}
	}
	if sy := c.Query("startYear"); sy != "" {
		if v, err := strconv.Atoi(sy); err == nil {
			startYear = v
			hasStartYear = true
		}
	}
	if em := c.Query("endMonth"); em != "" {
		if v, err := strconv.Atoi(em); err == nil && v >= 1 && v <= 12 {
			endMonth = v
			hasEndMonth = true
		}
	}
	if ey := c.Query("endYear"); ey != "" {
		if v, err := strconv.Atoi(ey); err == nil {
			endYear = v
			hasEndYear = true
		}
	}

	// Fill missing start components with current date.
	if !hasStartMonth {
		startMonth = int(now.Month())
	}
	if !hasStartYear {
		startYear = now.Year()
	}

	// If the end is not fully specified, derive it from the start so the range
	// stays predictable. A completely omitted end means "12 months up to start".
	if !hasEndMonth && !hasEndYear {
		endMonth, endYear = prevNMonths(startMonth, startYear, 11)
	} else {
		if !hasEndMonth {
			endMonth = startMonth
		}
		if !hasEndYear {
			endYear = startYear
		}
	}

	q := &dto.HoldingMonthlyQuery{
		StartMonth: startMonth,
		StartYear:  startYear,
		EndMonth:   endMonth,
		EndYear:    endYear,
	}

	// The service expects Start to be the chronologically latest month and End
	// the oldest. Normalize the endpoints so callers can pass them in any order.
	if q.EndYear > q.StartYear || (q.EndYear == q.StartYear && q.EndMonth > q.StartMonth) {
		q.StartMonth, q.EndMonth = q.EndMonth, q.StartMonth
		q.StartYear, q.EndYear = q.EndYear, q.StartYear
	}

	return q, nil
}

func (h *HoldingHandler) SyncPrices(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	result, err := h.holdingService.SyncPrices(c.Context(), userID)
	if err != nil {
		return h.respondHoldingError(c, "Failed to sync prices", err)
	}

	return response.Success(c, "Prices synced successfully for current month", result)
}

func (h *HoldingHandler) DuplicateHoldings(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req dto.DuplicateHoldingRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request format", err)
	}

	if err := bindValidate(c, &req); err != nil {
		return err
	}

	results, err := h.holdingService.DuplicateHoldings(c.Context(), userID, &req)
	if err != nil {
		return h.respondHoldingError(c, "Failed to duplicate holdings", err)
	}

	return response.Created(c, "Holdings duplicated successfully", results)
}

func splitComma(s string) []string {
	var result []string
	for _, v := range splitStr(s, ",") {
		if v != "" {
			result = append(result, v)
		}
	}
	return result
}

func splitStr(s, sep string) []string {
	var result []string
	for {
		idx := indexOfStr(s, sep)
		if idx < 0 {
			result = append(result, s)
			break
		}
		result = append(result, s[:idx])
		s = s[idx+len(sep):]
	}
	return result
}

func indexOfStr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func prevMonth(month, year int) (int, int) {
	if month == 1 {
		return 12, year - 1
	}
	return month - 1, year
}

func prevNMonths(month, year, n int) (int, int) {
	for range n {
		if month == 1 {
			month = 12
			year--
		} else {
			month--
		}
	}
	return month, year
}
