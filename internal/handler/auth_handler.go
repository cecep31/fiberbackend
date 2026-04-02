package handler

import (
	"net/http"

	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authService service.AuthService
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=30"`
	Password string `json:"password" validate:"required,min=8,password"`
}

type CheckUsernameRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8,password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8"`
	NewPassword     string `json:"new_password" validate:"required,min=8,password"`
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.HandleBindError(c, err)
	}

	user, err := h.authService.Register(c.Context(), req.Email, req.Username, req.Password)
	if err == service.ErrUserExists {
		return c.Status(http.StatusConflict).JSON(response.APIResponse{
			Success: false,
			Message: "Registration failed",
			Error:   "Email or username already exists",
		})
	}
	if err != nil {
		return response.InternalServerError(c, "Registration failed", err)
	}

	return response.Created(c, "User registered successfully", map[string]any{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
	})
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var loginReq LoginRequest
	if err := c.Bind().Body(&loginReq); err != nil {
		return response.HandleBindError(c, err)
	}

	token, refreshToken, user, err := h.authService.Login(c.Context(), loginReq.Email, loginReq.Password)
	if err == service.ErrInvalidCredentials {
		return response.Unauthorized(c, "Invalid email or password")
	}
	if err != nil {
		return response.InternalServerError(c, "Login failed", err)
	}

	return response.Success(c, "Login successful", map[string]any{
		"access_token":  token,
		"refresh_token": refreshToken,
		"user": map[string]any{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
		},
	})
}

func (h *AuthHandler) CheckUsername(c fiber.Ctx) error {
	var req CheckUsernameRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.HandleBindError(c, err)
	}

	isAvailable, err := h.authService.CheckUsernameAvailability(c.Context(), req.Username)
	if err != nil {
		return response.InternalServerError(c, "Failed to check username availability", err)
	}

	return response.Success(c, "Username availability checked", map[string]any{
		"username":  req.Username,
		"available": isAvailable,
	})
}

func (h *AuthHandler) ForgotPassword(c fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.HandleBindError(c, err)
	}

	err := h.authService.ForgotPassword(c.Context(), req.Email)
	if err == service.ErrUserNotFound {
		// Return success even if email doesn't exist for security reasons
		return response.Success(c, "If the email exists, a password reset link has been sent", nil)
	}
	if err != nil {
		return response.InternalServerError(c, "Failed to process password reset request", err)
	}

	return response.Success(c, "If the email exists, a password reset link has been sent", nil)
}

func (h *AuthHandler) ResetPassword(c fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.HandleBindError(c, err)
	}

	err := h.authService.ResetPassword(c.Context(), req.Token, req.Password)
	if err == service.ErrInvalidToken {
		return response.BadRequest(c, "Invalid or expired token", nil)
	}
	if err != nil {
		return response.InternalServerError(c, "Failed to reset password", err)
	}

	return response.Success(c, "Password reset successfully", nil)
}

func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.HandleBindError(c, err)
	}

	token, refreshToken, user, err := h.authService.RefreshToken(c.Context(), req.RefreshToken)
	if err == service.ErrInvalidToken {
		return response.Unauthorized(c, "Invalid refresh token")
	}
	if err != nil {
		return response.InternalServerError(c, "Failed to refresh token", err)
	}

	return response.Success(c, "Token refreshed successfully", map[string]any{
		"access_token":  token,
		"refresh_token": refreshToken,
		"user": map[string]any{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
		},
	})
}

func (h *AuthHandler) ChangePassword(c fiber.Ctx) error {
	var req ChangePasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.HandleBindError(c, err)
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Unauthorized(c, "Invalid or missing user context")
	}

	err := h.authService.ChangePassword(c.Context(), userID, req.CurrentPassword, req.NewPassword)
	if err == service.ErrInvalidCredentials {
		return response.Unauthorized(c, "Current password is incorrect")
	}
	if err != nil {
		return response.InternalServerError(c, "Failed to change password", err)
	}

	return response.Success(c, "Password changed successfully", nil)
}
