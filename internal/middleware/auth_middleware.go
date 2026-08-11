package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"fiberbackend/config"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/applog"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

var log = applog.Component("auth")

// AuthMiddleware provides authentication middleware for Fiber.
//
// It stores the validated JWT claims in the request locals under the "user"
// key; handlers read them via handler.GetUserIDFromClaims.
type AuthMiddleware struct {
	conf        *config.Config
	userService service.UserService
}

// NewAuthMiddleware creates a new instance of AuthMiddleware.
func NewAuthMiddleware(conf *config.Config, userService service.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		conf:        conf,
		userService: userService,
	}
}

// Auth validates JWT tokens and sets user claims in the context.
func (a *AuthMiddleware) Auth() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			log.Warn("auth: missing authorization header", "path", c.Path(), "remote_ip", c.IP())
			return response.Unauthorized(c, "Missing authorization header")
		}

		tokenString, err := extractBearerToken(authHeader)
		if err != nil {
			log.Warn("auth: malformed authorization header", "path", c.Path(), "remote_ip", c.IP(), "error", err)
			return response.Unauthorized(c, "Invalid authorization header")
		}

		claims, err := validateToken(tokenString, a.conf.Auth.JWTSecret)
		if err != nil {
			// Log the real parse/validation error server-side only; never expose it to clients.
			log.Warn("auth: invalid token", "path", c.Path(), "remote_ip", c.IP(), "error", err)
			return response.Unauthorized(c, "Invalid or expired token")
		}

		c.Locals("user", claims)
		return c.Next()
	}
}

// OptionalAuth validates JWT tokens if present but does not require them.
func (a *AuthMiddleware) OptionalAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		tokenString, err := extractBearerToken(authHeader)
		if err != nil {
			return c.Next()
		}

		claims, err := validateToken(tokenString, a.conf.Auth.JWTSecret)
		if err != nil {
			return c.Next()
		}

		c.Locals("user", claims)
		return c.Next()
	}
}

// AuthAdmin validates that the user has admin privileges (DB-checked).
// Must be chained after Auth().
func (a *AuthMiddleware) AuthAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		userClaims := c.Locals("user")
		if userClaims == nil {
			return response.Unauthorized(c, "Authentication required")
		}

		claims, ok := userClaims.(jwt.MapClaims)
		if !ok {
			return response.Unauthorized(c, "Authentication required")
		}

		userID, err := getUserIDFromClaims(claims)
		if err != nil {
			log.Warn("auth: admin check failed to resolve user id", "path", c.Path(), "remote_ip", c.IP(), "error", err)
			return response.Unauthorized(c, "Authentication required")
		}

		isSuperAdmin, err := a.isSuperAdminFromDB(c.Context(), userID)
		if err != nil {
			log.Warn("auth: failed to validate admin privileges", "path", c.Path(), "remote_ip", c.IP(), "user_id", userID, "error", err)
			return response.Unauthorized(c, "Failed to validate privileges")
		}

		if !isSuperAdmin {
			log.Warn("auth: insufficient privileges", "path", c.Path(), "remote_ip", c.IP(), "user_id", userID)
			return response.Forbidden(c, "Insufficient privileges")
		}

		return c.Next()
	}
}

// extractBearerToken extracts the token from the Authorization header.
func extractBearerToken(authHeader string) (string, error) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid token format, expected 'Bearer <token>'")
	}
	return parts[1], nil
}

// validateToken validates the JWT token and returns the claims.
func validateToken(tokenString, jwtSecret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token parsing failed: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func getUserIDFromClaims(claims jwt.MapClaims) (string, error) {
	userID, exists := claims["user_id"]
	if !exists {
		return "", errors.New("unauthorized: user ID not found in token")
	}

	switch v := userID.(type) {
	case string:
		if v == "" {
			return "", errors.New("unauthorized: invalid user ID in token")
		}
		return v, nil
	default:
		return "", errors.New("unauthorized: invalid user ID format in token")
	}
}

func (a *AuthMiddleware) isSuperAdminFromDB(ctx context.Context, userID string) (bool, error) {
	user, err := a.userService.GetAdminByID(ctx, userID, false)
	if err != nil {
		return false, err
	}

	if user.IsSuperAdmin == nil {
		return false, nil
	}

	return *user.IsSuperAdmin, nil
}
