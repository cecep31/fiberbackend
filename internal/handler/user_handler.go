package handler

import (
	"errors"

	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/dto"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

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

	if c.Query("deleted") == "true" {
		userResponse, err := h.userService.GetAdminByID(c.Context(), userID, true)
		if err != nil {
			return response.InternalServerError(c, "Failed to retrieve user", err)
		}
		return response.Success(c, "Successfully retrieved user", userResponse)
	}

	var currentUserID string
	if uid, ok := GetUserIDFromClaims(c); ok {
		currentUserID = uid
	}

	userResponse, err := h.userFollowService.GetUserWithFollowStatus(c.Context(), userID, currentUserID, true)
	if err != nil {
		return response.InternalServerError(c, "Failed to retrieve user", err)
	}

	return response.Success(c, "Successfully retrieved user", userResponse)
}

func (h *UserHandler) GetByUsername(c fiber.Ctx) error {
	username := c.Params("username")

	var currentUserID string
	if uid, ok := GetUserIDFromClaims(c); ok {
		currentUserID = uid
	}

	user, err := h.userService.GetByUsername(c.Context(), username)
	if err != nil {
		return response.InternalServerError(c, "Failed to retrieve user", err)
	}

	userResponse, err := h.userFollowService.GetUserWithFollowStatus(c.Context(), user.ID, currentUserID, false)
	if err != nil {
		return response.InternalServerError(c, "Failed to retrieve user", err)
	}

	return response.Success(c, "Successfully retrieved user", userResponse)
}

func (h *UserHandler) GetUsers(c fiber.Ctx) error {
	deletedFilter, err := dto.ParseUserDeletedFilter(c.Query("deleted"))
	if err != nil {
		return response.BadRequest(c, "Invalid deleted filter", err)
	}

	limit, offset := ParsePaginationParams(c, 10)

	users, total, err := h.userService.GetUsers(c.Context(), offset, limit, deletedFilter)
	if err != nil {
		return response.InternalServerError(c, "Failed to retrieve users", err)
	}

	meta := response.CalculatePaginationMeta(total, offset, limit)
	return response.SuccessWithMeta(c, "Successfully retrieved users", users, meta)
}

func (h *UserHandler) DeleteUser(c fiber.Ctx) error {
	id := c.Params("id")
	err := h.userService.Delete(c.Context(), id)
	if err != nil {
		return response.InternalServerError(c, "Failed to delete user", err)
	}

	return response.Success(c, "Successfully deleted user", nil)
}

func (h *UserHandler) RestoreUser(c fiber.Ctx) error {
	id := c.Params("id")

	userResponse, err := h.userService.Restore(c.Context(), id)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return response.NotFound(c, "Deleted user not found", err)
		}
		if errors.Is(err, apperrors.ErrUserExists) {
			return response.Conflict(c, "Cannot restore user", "Email or username already taken by another active user")
		}
		return response.InternalServerError(c, "Failed to restore user", err)
	}

	return response.Success(c, "Successfully restored user", userResponse)
}

func (h *UserHandler) GetMe(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	userResponse, err := h.userService.GetMe(c.Context(), userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to retrieve user", err)
	}

	return response.Success(c, "Successfully retrieved current user", userResponse)
}
