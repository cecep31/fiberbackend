package handler

import (
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"
	"fiberbackend/pkg/utils"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type PostLikeHandler struct {
	postLikeService service.PostLikeService
}

func NewPostLikeHandler(postLikeService service.PostLikeService) *PostLikeHandler {
	return &PostLikeHandler{postLikeService: postLikeService}
}

// LikePost likes a post
func (h *PostLikeHandler) LikePost(c fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID is required", nil)
	}

	userID, err := utils.RequireUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User authentication required")
	}

	// Like the post
	err = h.postLikeService.LikePost(c.Context(), postID, userID)
	if err != nil {
		if err.Error() == "user has already liked this post" {
			return response.BadRequest(c, "You have already liked this post", nil)
		}
		return response.InternalServerError(c, "Failed to like post", err)
	}

	return response.Success(c, "Post liked successfully", nil)
}

// UnlikePost unlikes a post
func (h *PostLikeHandler) UnlikePost(c fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID is required", nil)
	}

	userID, err := utils.RequireUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User authentication required")
	}

	// Unlike the post
	err = h.postLikeService.UnlikePost(c.Context(), postID, userID)
	if err != nil {
		if err.Error() == "user has not liked this post" {
			return response.BadRequest(c, "You have not liked this post", nil)
		}
		return response.InternalServerError(c, "Failed to unlike post", err)
	}

	return response.Success(c, "Post unliked successfully", nil)
}

// GetPostLikes gets likes for a specific post
func (h *PostLikeHandler) GetPostLikes(c fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID is required", nil)
	}

	// Parse pagination parameters
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 10 // default
	offset := 0 // default

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get likes
	likes, total, err := h.postLikeService.GetLikesByPostID(c.Context(), postID, limit, offset)
	if err != nil {
		return response.InternalServerError(c, "Failed to get post likes", err)
	}

	// Convert to response format
	likeResponses := make([]interface{}, len(likes))
	for i, like := range likes {
		likeResponses[i] = like.ToResponse()
	}

	responseData := map[string]any{
		"likes":  likeResponses,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	return response.Success(c, "Post likes retrieved successfully", responseData)
}

// GetPostLikeStats gets like statistics for a post
func (h *PostLikeHandler) GetPostLikeStats(c fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID is required", nil)
	}

	stats, err := h.postLikeService.GetLikeStats(c.Context(), postID)
	if err != nil {
		return response.InternalServerError(c, "Failed to get like stats", err)
	}

	return response.Success(c, "Like stats retrieved successfully", stats)
}

// CheckUserLiked checks if the current user has liked a post
func (h *PostLikeHandler) CheckUserLiked(c fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID is required", nil)
	}

	userID, err := utils.RequireUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User authentication required")
	}

	hasLiked, err := h.postLikeService.HasUserLikedPost(c.Context(), postID, userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to check like status", err)
	}

	responseData := map[string]interface{}{
		"has_liked": hasLiked,
		"post_id":   postID,
		"user_id":   userID,
	}

	return response.Success(c, "Like status retrieved successfully", responseData)
}
