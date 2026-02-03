package handler

import (
	"fiberbackend/internal/model"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type UserFollowHandler struct {
	userFollowService service.UserFollowService
}

func NewUserFollowHandler(userFollowService service.UserFollowService) *UserFollowHandler {
	return &UserFollowHandler{userFollowService: userFollowService}
}

// FollowUser follows a user
func (h *UserFollowHandler) FollowUser(c fiber.Ctx) error {
	// Get current user ID from JWT
	userClaims := c.Locals("user")
	if userClaims == nil {
		return response.Unauthorized(c, "Authentication required")
	}

	claims, ok := userClaims.(jwt.MapClaims)
	if !ok {
		return response.InternalServerError(c, "Invalid user context", nil)
	}

	currentUserID, exists := claims["user_id"]
	if !exists {
		return response.InternalServerError(c, "User ID not found in token", nil)
	}

	currentUserIDStr := fmt.Sprintf("%v", currentUserID)

	// Get user ID to follow from request body
	var followReq model.FollowRequest
	if err := c.Bind().Body(&followReq); err != nil {
		return response.HandleBindError(c, err)
	}

	// Follow the user
	followResponse, err := h.userFollowService.FollowUser(c.Context(), currentUserIDStr, followReq.UserID)
	if err != nil {
		return response.InternalServerError(c, "Failed to follow user", err)
	}

	return response.Success(c, followResponse.Message, followResponse)
}

// UnfollowUser unfollows a user
func (h *UserFollowHandler) UnfollowUser(c fiber.Ctx) error {
	// Get current user ID from JWT
	userClaims := c.Locals("user")
	if userClaims == nil {
		return response.Unauthorized(c, "Authentication required")
	}

	claims, ok := userClaims.(jwt.MapClaims)
	if !ok {
		return response.InternalServerError(c, "Invalid user context", nil)
	}

	currentUserID, exists := claims["user_id"]
	if !exists {
		return response.InternalServerError(c, "User ID not found in token", nil)
	}

	currentUserIDStr := fmt.Sprintf("%v", currentUserID)

	// Get user ID to unfollow from URL parameter
	userIDToUnfollow := c.Params("id")
	if userIDToUnfollow == "" {
		return response.BadRequest(c, "User ID is required", nil)
	}

	// Unfollow the user
	followResponse, err := h.userFollowService.UnfollowUser(c.Context(), currentUserIDStr, userIDToUnfollow)
	if err != nil {
		return response.InternalServerError(c, "Failed to unfollow user", err)
	}

	return response.Success(c, followResponse.Message, followResponse)
}

// GetFollowers gets followers of a user
func (h *UserFollowHandler) GetFollowers(c fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return response.BadRequest(c, "User ID is required", nil)
	}

	// Parse pagination parameters
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get followers
	followers, total, err := h.userFollowService.GetFollowers(c.Context(), userID, limit, offset)
	if err != nil {
		return response.InternalServerError(c, "Failed to get followers", err)
	}

	// Calculate pagination meta
	meta := response.CalculatePaginationMeta(total, offset, limit)

	return response.SuccessWithMeta(c, "Successfully retrieved followers", followers, meta)
}

// GetFollowing gets users that a user is following
func (h *UserFollowHandler) GetFollowing(c fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return response.BadRequest(c, "User ID is required", nil)
	}

	// Parse pagination parameters
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get following
	following, total, err := h.userFollowService.GetFollowing(c.Context(), userID, limit, offset)
	if err != nil {
		return response.InternalServerError(c, "Failed to get following", err)
	}

	// Calculate pagination meta
	meta := response.CalculatePaginationMeta(total, offset, limit)

	return response.SuccessWithMeta(c, "Successfully retrieved following", following, meta)
}

// GetFollowStats gets follow statistics for a user
func (h *UserFollowHandler) GetFollowStats(c fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return response.BadRequest(c, "User ID is required", nil)
	}

	// Get follow statistics
	stats, err := h.userFollowService.GetFollowStats(c.Context(), userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to get follow statistics", err)
	}

	return response.Success(c, "Successfully retrieved follow statistics", stats)
}

// CheckFollowStatus checks if current user is following a specific user
func (h *UserFollowHandler) CheckFollowStatus(c fiber.Ctx) error {
	// Get current user from JWT
	userClaims := c.Locals("user")
	if userClaims == nil {
		return response.Unauthorized(c, "Authentication required")
	}

	claims, ok := userClaims.(jwt.MapClaims)
	if !ok {
		return response.InternalServerError(c, "Invalid user context", nil)
	}

	currentUserID, exists := claims["user_id"]
	if !exists {
		return response.InternalServerError(c, "User ID not found in token", nil)
	}

	currentUserIDStr := fmt.Sprintf("%v", currentUserID)

	// Get target user ID
	targetUserID := c.Params("id")
	if targetUserID == "" {
		return response.BadRequest(c, "User ID is required", nil)
	}

	// Check follow status
	isFollowing, err := h.userFollowService.IsFollowing(c.Context(), currentUserIDStr, targetUserID)
	if err != nil {
		return response.InternalServerError(c, "Failed to check follow status", err)
	}

	return response.Success(c, "Successfully checked follow status", map[string]bool{
		"is_following": isFollowing,
	})
}

// GetMutualFollows gets mutual follows between current user and another user
func (h *UserFollowHandler) GetMutualFollows(c fiber.Ctx) error {
	// Get current user from JWT
	userClaims := c.Locals("user")
	if userClaims == nil {
		return response.Unauthorized(c, "Authentication required")
	}

	claims, ok := userClaims.(jwt.MapClaims)
	if !ok {
		return response.InternalServerError(c, "Invalid user context", nil)
	}

	currentUserID, exists := claims["user_id"]
	if !exists {
		return response.InternalServerError(c, "User ID not found in token", nil)
	}

	currentUserIDStr := fmt.Sprintf("%v", currentUserID)

	// Get other user ID
	otherUserID := c.Params("id")
	if otherUserID == "" {
		return response.BadRequest(c, "User ID is required", nil)
	}

	// Get mutual follows
	mutualFollows, err := h.userFollowService.GetMutualFollows(c.Context(), currentUserIDStr, otherUserID)
	if err != nil {
		return response.InternalServerError(c, "Failed to get mutual follows", err)
	}

	return response.Success(c, "Successfully retrieved mutual follows", mutualFollows)
}
