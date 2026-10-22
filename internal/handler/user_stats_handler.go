package handler

import (
	"fiberbackend/internal/dto"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type UserStatsHandler struct {
	statsService service.UserStatsService
}

func NewUserStatsHandler(statsService service.UserStatsService) *UserStatsHandler {
	return &UserStatsHandler{statsService: statsService}
}

func (h *UserStatsHandler) GetStats(c fiber.Ctx) error {
	limit, _ := ParsePaginationParams(c, 10)
	stats, err := h.statsService.GetUserStats(c.Context(), dto.DateRangeQuery{
		StartDate: c.Query("startDate"),
		EndDate:   c.Query("endDate"),
	}, limit)
	if err != nil {
		return response.InternalServerError(c, "Failed to fetch user stats", err)
	}
	return response.Success(c, "User stats fetched successfully", stats)
}
