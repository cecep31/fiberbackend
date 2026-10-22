package handler

import (
	"strconv"

	"fiberbackend/internal/dto"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type PostStatsHandler struct {
	statsService service.PostStatsService
}

func NewPostStatsHandler(statsService service.PostStatsService) *PostStatsHandler {
	return &PostStatsHandler{statsService: statsService}
}

func dateRangeFromQuery(c fiber.Ctx) dto.DateRangeQuery {
	return dto.DateRangeQuery{
		StartDate: c.Query("startDate"),
		EndDate:   c.Query("endDate"),
	}
}

func (h *PostStatsHandler) GetStats(c fiber.Ctx) error {
	limit, _ := ParsePaginationParams(c, 10)
	var tagID *int
	if raw := c.Query("tagId"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			tagID = &parsed
		}
	}
	stats, err := h.statsService.GetPostStats(c.Context(), dateRangeFromQuery(c), limit, tagID)
	if err != nil {
		return response.InternalServerError(c, "Failed to fetch post stats", err)
	}
	return response.Success(c, "Post stats fetched successfully", stats)
}

func (h *PostStatsHandler) GetEngagement(c fiber.Ctx) error {
	metrics, err := h.statsService.GetEngagementMetrics(c.Context(), dateRangeFromQuery(c))
	if err != nil {
		return response.InternalServerError(c, "Failed to fetch engagement metrics", err)
	}
	return response.Success(c, "Engagement metrics fetched successfully", metrics)
}
