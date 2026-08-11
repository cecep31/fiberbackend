package handler

import (
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type PostViewHandler struct {
	postViewService service.PostViewService
}

func NewPostViewHandler(postViewService service.PostViewService) *PostViewHandler {
	return &PostViewHandler{postViewService: postViewService}
}

func (h *PostViewHandler) RecordView(c fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID is required", nil)
	}

	var userID string
	if uid, ok := GetUserIDFromClaims(c); ok {
		userID = uid
	}

	ipAddress := c.IP()
	userAgent := c.Get(fiber.HeaderUserAgent)

	err := h.postViewService.RecordView(c.Context(), postID, userID, &ipAddress, &userAgent)
	if err != nil {
		return response.InternalServerError(c, "Failed to record view", err)
	}

	return response.Success(c, "View recorded successfully", nil)
}

func (h *PostViewHandler) GetPostViews(c fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID is required", nil)
	}

	limit, offset := ParsePaginationParams(c, 10)

	views, total, err := h.postViewService.GetViewsByPostID(c.Context(), postID, limit, offset)
	if err != nil {
		return response.InternalServerError(c, "Failed to get post views", err)
	}

	meta := response.CalculatePaginationMeta(total, offset, limit)

	return response.SuccessWithMeta(c, "Successfully retrieved post views", views, meta)
}

func (h *PostViewHandler) GetPostViewStats(c fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID is required", nil)
	}

	stats, err := h.postViewService.GetViewStats(c.Context(), postID)
	if err != nil {
		return response.InternalServerError(c, "Failed to get view statistics", err)
	}

	return response.Success(c, "Successfully retrieved view statistics", stats)
}

func (h *PostViewHandler) CheckUserViewed(c fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID is required", nil)
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "Authentication required")
	}

	hasViewed, err := h.postViewService.HasUserViewedPost(c.Context(), postID, userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to check view status", err)
	}

	return response.Success(c, "Successfully checked view status", map[string]bool{
		"has_viewed": hasViewed,
	})
}
