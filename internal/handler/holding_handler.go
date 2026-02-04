package handler

import (
	"fiberbackend/internal/model"
	"fiberbackend/internal/repository"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type HoldingHandler struct {
	holdingService service.HoldingService
}

func NewHoldingHandler(holdingService service.HoldingService) *HoldingHandler {
	return &HoldingHandler{holdingService: holdingService}
}

func (h *HoldingHandler) getUserID(c fiber.Ctx) (string, error) {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok || claims["user_id"] == nil {
		return "", fmt.Errorf("missing user context")
	}
	return fmt.Sprintf("%v", claims["user_id"]), nil
}

func (h *HoldingHandler) Create(c fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Invalid or missing user context")
	}

	var dto model.CreateHoldingDTO
	if err := c.Bind().Body(&dto); err != nil {
		return response.HandleBindError(c, err)
	}

	holding, err := h.holdingService.Create(c.Context(), userID, &dto)
	if err != nil {
		return response.InternalServerError(c, "Failed to create holding", err)
	}
	return response.Created(c, "Holding created", holding)
}

func (h *HoldingHandler) GetByID(c fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Invalid or missing user context")
	}

	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid holding ID", err)
	}

	holding, err := h.holdingService.GetByID(c.Context(), id, userID)
	if err != nil {
		if err == repository.ErrHoldingNotFound {
			return response.NotFound(c, "Holding not found", err)
		}
		return response.InternalServerError(c, "Failed to get holding", err)
	}
	return response.Success(c, "Holding retrieved", holding)
}

func (h *HoldingHandler) List(c fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Invalid or missing user context")
	}

	now := time.Now()
	currentMonth := int(now.Month())
	currentYear := now.Year()

	filter := &model.HoldingQueryFilter{
		Limit:  20,
		Offset: 0,
		Month:  &currentMonth, // default: current month
		Year:   &currentYear,  // default: current year
	}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filter.Limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			filter.Offset = n
		}
	}
	if v := c.Query("month"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 12 {
			filter.Month = &n
		}
	}
	if v := c.Query("year"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 2000 {
			filter.Year = &n
		}
	}

	list, total, err := h.holdingService.ListByUserID(c.Context(), userID, filter)
	if err != nil {
		return response.InternalServerError(c, "Failed to list holdings", err)
	}

	meta := response.PaginationMeta{
		TotalItems: int(total),
		Offset:     filter.Offset,
		Limit:      filter.Limit,
		TotalPages: int(total) / filter.Limit,
	}
	if int(total)%filter.Limit > 0 {
		meta.TotalPages++
	}
	return response.SuccessWithMeta(c, "Holdings retrieved", list, meta)
}

func (h *HoldingHandler) Update(c fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Invalid or missing user context")
	}

	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid holding ID", err)
	}

	var dto model.UpdateHoldingDTO
	if err := c.Bind().Body(&dto); err != nil {
		return response.HandleBindError(c, err)
	}

	holding, err := h.holdingService.Update(c.Context(), id, userID, &dto)
	if err != nil {
		if err == repository.ErrHoldingNotFound {
			return response.NotFound(c, "Holding not found", err)
		}
		return response.InternalServerError(c, "Failed to update holding", err)
	}
	return response.Success(c, "Holding updated", holding)
}

func (h *HoldingHandler) Delete(c fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Invalid or missing user context")
	}

	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "Invalid holding ID", err)
	}

	if err := h.holdingService.Delete(c.Context(), id, userID); err != nil {
		if err == repository.ErrHoldingNotFound {
			return response.NotFound(c, "Holding not found", err)
		}
		return response.InternalServerError(c, "Failed to delete holding", err)
	}
	return response.Success(c, "Holding deleted", nil)
}
