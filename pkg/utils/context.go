// Package utils provides common utility functions for the application.
package utils

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// ContextKey defines the type for context keys to avoid collisions.
type ContextKey string

const (
	// UserContextKey is the key for storing user claims in context.
	UserContextKey ContextKey = "user"
	// UserIDContextKey is the key for storing user ID in context.
	UserIDContextKey ContextKey = "user_id"
)

// GetUserIDFromContext extracts the user ID from the Fiber context.
// Returns the user ID string and a boolean indicating success.
func GetUserIDFromContext(c fiber.Ctx) (string, bool) {
	userID := c.Locals(string(UserIDContextKey))
	if userID == nil {
		return "", false
	}
	return fmt.Sprintf("%v", userID), true
}

// GetUserClaimsFromContext extracts the JWT claims from the Fiber context.
// Returns the claims map and a boolean indicating success.
func GetUserClaimsFromContext(c fiber.Ctx) (jwt.MapClaims, bool) {
	userClaims := c.Locals(string(UserContextKey))
	if userClaims == nil {
		return nil, false
	}
	claims, ok := userClaims.(jwt.MapClaims)
	return claims, ok
}

// GetUserIDFromClaims extracts the user ID from JWT claims.
func GetUserIDFromClaims(claims jwt.MapClaims) (string, bool) {
	if claims == nil {
		return "", false
	}
	userID, exists := claims["user_id"]
	if !exists {
		return "", false
	}
	return fmt.Sprintf("%v", userID), true
}

// RequireUserID extracts the user ID from context or returns an error response.
// This is a helper function for handlers that require authentication.
func RequireUserID(c fiber.Ctx) (string, error) {
	userID, ok := GetUserIDFromContext(c)
	if !ok || userID == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
	}
	return userID, nil
}
