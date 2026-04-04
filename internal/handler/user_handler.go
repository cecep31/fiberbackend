package handler

import (
	"context"
	"errors"

	"fiberbackend/internal/model"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"
	"fiberbackend/pkg/utils"
	"fiberbackend/pkg/validator"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	userService       service.UserService
	userFollowService service.UserFollowService
}

func NewUserHandler(userService service.UserService, userFollowService service.UserFollowService) *UserHandler {
	return &UserHandler{
		userService:       userService,
		userFollowService: userFollowService,
	}
}

func (h *UserHandler) GetByID(c fiber.Ctx) error {
	userID := c.Params("id")

	// Get current user ID from JWT if authenticated
	currentUserID, _ := utils.GetUserIDFromContext(c)

	// Get user with follow status
	userResponse, err := h.userFollowService.GetUserWithFollowStatus(c.Context(), userID, currentUserID)
	if err != nil {
		return response.InternalServerError(c, "Failed to retrieve user", err)
	}

	return response.Success(c, "Successfully retrieved user", userResponse)
}

// GetByUsername returns a public user profile by username with profile details preloaded.
// When the caller is authenticated, follow status relative to the target user is included.
func (h *UserHandler) GetByUsername(c fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return response.BadRequest(c, "Username is required", nil)
	}

	currentUserID, _ := utils.GetUserIDFromContext(c)

	userResponse, err := h.userService.GetByUsernameWithProfile(c.Context(), username)
	if err != nil {
		if errors.Is(err, utils.ErrUserNotFound) {
			return response.NotFound(c, "User not found", err)
		}
		return response.InternalServerError(c, "Failed to retrieve user", err)
	}

	if err := h.attachFollowStatusIfViewer(c.Context(), currentUserID, userResponse); err != nil {
		return response.InternalServerError(c, "Failed to retrieve user", err)
	}

	return response.Success(c, "Successfully retrieved user", userResponse)
}

func (h *UserHandler) attachFollowStatusIfViewer(ctx context.Context, viewerID string, target *model.UserResponse) error {
	if viewerID == "" || viewerID == target.ID {
		return nil
	}
	isFollowing, err := h.userFollowService.IsFollowing(ctx, viewerID, target.ID)
	if err != nil {
		return err
	}
	target.IsFollowing = &isFollowing
	return nil
}

func (h *UserHandler) GetUsers(c fiber.Ctx) error {
	// Validate and sanitize pagination parameters
	limit, offset, err := validator.ValidatePaginationWithDefaults(c.Query("limit"), c.Query("offset"))
	if err != nil {
		return response.BadRequest(c, "Invalid pagination parameters", err)
	}

	users, total, err := h.userService.GetUsers(c.Context(), offset, limit)
	if err != nil {
		return response.InternalServerError(c, "Failed to retrieve users", err)
	}

	meta := response.CalculatePaginationMeta(total, offset, limit)
	return response.SuccessWithMeta(c, "Successfully retrieved users", users, meta)
}

// DeleteUser deletes a user by ID
func (h *UserHandler) DeleteUser(c fiber.Ctx) error {
	id := c.Params("id")
	err := h.userService.Delete(c.Context(), id)
	if err != nil {
		return response.InternalServerError(c, "Failed to delete user", err)
	}

	return response.Success(c, "Successfully deleted user", nil)
}

// GetMe returns the current authenticated user's information
func (h *UserHandler) GetMe(c fiber.Ctx) error {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Invalid or missing user context")
	}

	userResponse, err := h.userService.GetByID(c.Context(), userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to retrieve user", err)
	}

	return response.Success(c, "Successfully retrieved current user", userResponse)
}
