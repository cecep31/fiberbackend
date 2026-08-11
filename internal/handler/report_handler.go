package handler

import (
	"strconv"

	"fiberbackend/internal/dto"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type ReportHandler struct {
	reportService service.ReportService
}

func NewReportHandler(reportService service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

func (h *ReportHandler) GetOverview(c fiber.Ctx) error {
	query := dto.DateRangeQuery{
		StartDate: c.Query("startDate"),
		EndDate:   c.Query("endDate"),
	}
	overview, err := h.reportService.GetOverviewStats(c.Context())
	if err != nil {
		return response.InternalServerError(c, "Failed to fetch overview report", err)
	}
	engagement, err := h.reportService.GetEngagementMetrics(c.Context(), query)
	if err != nil {
		return response.InternalServerError(c, "Failed to fetch engagement metrics", err)
	}
	return response.Success(c, "Overview report fetched successfully", map[string]any{
		"overview":   overview,
		"engagement": engagement,
	})
}

func (h *ReportHandler) GetUsers(c fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	report, err := h.reportService.GetUserReport(c.Context(), dto.DateRangeQuery{
		StartDate: c.Query("startDate"),
		EndDate:   c.Query("endDate"),
	}, limit)
	if err != nil {
		return response.InternalServerError(c, "Failed to fetch user report", err)
	}
	return response.Success(c, "User report fetched successfully", report)
}

func (h *ReportHandler) GetPosts(c fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	var tagID *int
	if raw := c.Query("tagId"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			tagID = &parsed
		}
	}
	report, err := h.reportService.GetPostReport(c.Context(), dto.DateRangeQuery{
		StartDate: c.Query("startDate"),
		EndDate:   c.Query("endDate"),
	}, limit, tagID)
	if err != nil {
		return response.InternalServerError(c, "Failed to fetch post report", err)
	}
	return response.Success(c, "Post report fetched successfully", report)
}

func (h *ReportHandler) GetEngagement(c fiber.Ctx) error {
	metrics, err := h.reportService.GetEngagementMetrics(c.Context(), dto.DateRangeQuery{
		StartDate: c.Query("startDate"),
		EndDate:   c.Query("endDate"),
	})
	if err != nil {
		return response.InternalServerError(c, "Failed to fetch engagement metrics", err)
	}
	return response.Success(c, "Engagement metrics fetched successfully", metrics)
}
