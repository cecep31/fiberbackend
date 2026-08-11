package handler

import (
	"strconv"

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
