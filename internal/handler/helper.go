package handler

import (
	"context"
	"errors"
	"strconv"

	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"
	"fiberbackend/pkg/validator"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

var validate = validator.NewValidator()

// bindValidate runs app validation on a bound request struct and maps
// validation failures to a structured 422 response (mirrors echo's c.Validate).
func bindValidate(c fiber.Ctx, req any) error {
	if err := validate.Validate(req); err != nil {
		return response.FromValidateError(c, err)
	}
	return nil
}

// GetUserIDFromClaims reads the authenticated user ID from JWT claims
// previously stored in the context by the auth middleware.
func GetUserIDFromClaims(c fiber.Ctx) (string, bool) {
	userClaims := c.Locals("user")
	if userClaims == nil {
		return "", false
	}

	switch v := userClaims.(type) {
	case jwt.MapClaims:
		userID, exists := v["user_id"]
		if !exists {
			return "", false
		}
		userIDStr, ok := userID.(string)
		if !ok {
			return "", false
		}
		return userIDStr, true
	case *jwt.Token:
		claims, ok := v.Claims.(jwt.MapClaims)
		if !ok {
			return "", false
		}
		userID, exists := claims["user_id"]
		if !exists {
			return "", false
		}
		userIDStr, ok := userID.(string)
		if !ok {
			return "", false
		}
		return userIDStr, true
	case map[string]any:
		userID, exists := v["user_id"]
		if !exists {
			return "", false
		}
		userIDStr, ok := userID.(string)
		if !ok {
			return "", false
		}
		return userIDStr, true
	}
	return "", false
}

// ParsePaginationParams parses limit/offset query params with a default
// limit capped at 100.
func ParsePaginationParams(c fiber.Ctx, defaultLimit int) (limit, offset int) {
	limit = defaultLimit
	offset = 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 100 {
		limit = 100
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}

// respondSpamRejected answers an anti-spam rejection with the shared 422
// envelope used across handlers.
func respondSpamRejected(c fiber.Ctx, message, field string, err error) error {
	reason := err.Error()
	if rejection, ok := errors.AsType[*apperrors.SpamRejection](err); ok {
		reason = rejection.Reason
	}
	return response.SpamRejected(c, message, field, reason)
}

// isSuperAdmin reports whether userID is currently a super admin. It always
// reads the database: the is_super_admin JWT claim stays in the token until it
// expires, so trusting it would keep a demoted admin privileged.
func isSuperAdmin(ctx context.Context, userService service.UserService, userID string) bool {
	if userService == nil || userID == "" {
		return false
	}
	user, err := userService.GetAdminByID(ctx, userID, false)
	if err != nil || user.IsSuperAdmin == nil {
		return false
	}
	return *user.IsSuperAdmin
}

// respondError maps well-known domain errors to their HTTP status and falls
// back to 500 for anything else.
func respondError(c fiber.Ctx, message string, err error) error {
	switch {
	case errors.Is(err, apperrors.ErrUserNotFound),
		errors.Is(err, apperrors.ErrPostNotFound),
		errors.Is(err, apperrors.ErrCommentNotFound),
		errors.Is(err, apperrors.ErrTagNotFound),
		errors.Is(err, apperrors.ErrNotificationNotFound):
		return response.NotFound(c, message, err)
	case errors.Is(err, apperrors.ErrNotAuthor),
		errors.Is(err, apperrors.ErrCommentNotOwned),
		errors.Is(err, apperrors.ErrPostNotOwned):
		return response.Forbidden(c, message)
	case errors.Is(err, apperrors.ErrContentSpamRejected):
		return respondSpamRejected(c, "Content rejected by anti-spam policy", "text", err)
	case errors.Is(err, apperrors.ErrUserExists),
		errors.Is(err, apperrors.ErrAlreadyFollowing):
		return response.Conflict(c, message, err.Error())
	case errors.Is(err, apperrors.ErrCannotFollowSelf),
		errors.Is(err, apperrors.ErrNotFollowing),
		errors.Is(err, apperrors.ErrEmptyPostID),
		errors.Is(err, apperrors.ErrDateRangeTooLarge):
		return response.BadRequest(c, message, err)
	default:
		return response.InternalServerError(c, message, err)
	}
}
