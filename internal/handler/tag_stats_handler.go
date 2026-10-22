package handler

import (
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type TagStatsHandler struct {
	statsService service.TagStatsService
}

func NewTagStatsHandler(statsService service.TagStatsService) *TagStatsHandler {
	return &TagStatsHandler{statsService: statsService}
}

func (h *TagStatsHandler) GetStats(c fiber.Ctx) error {
	limit, _ := ParsePaginationParams(c, 10)
	stats, err := h.statsService.GetTagStats(c.Context(), limit)
	if err != nil {
		return response.InternalServerError(c, "Failed to fetch tag stats", err)
	}
	return response.Success(c, "Tag stats fetched successfully", stats)
}
