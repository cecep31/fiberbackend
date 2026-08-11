package handler

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fiberbackend/config"
	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/dto"
	"fiberbackend/internal/model"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authService     service.AuthService
	activityService service.AuthActivityService
	frontendConfig  config.FrontendConfig
}

func NewAuthHandler(authService service.AuthService, activityService service.AuthActivityService, frontendConfig config.FrontendConfig) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		activityService: activityService,
		frontendConfig:  frontendConfig,
	}
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request format", err)
	}

	if err := bindValidate(c, &req); err != nil {
		return err
	}

	user, err := h.authService.Register(c.Context(), req.Email, req.Username, req.Password)
	if errors.Is(err, apperrors.ErrUserExists) {
		return response.Conflict(c, "Registration failed", "Email or username already exists")
	}
	if err != nil {
		return response.InternalServerError(c, "Registration failed", err)
	}

	ipAddress := c.IP()
	userAgent := c.Get(fiber.HeaderUserAgent)
	h.activityService.LogActivity(c.Context(), &user.ID, model.ActivityRegister, model.StatusSuccess, ipAddress, userAgent, nil, nil)

	return response.Created(c, "User registered successfully", map[string]any{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
	})
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var loginReq dto.LoginRequest
	if err := c.Bind().Body(&loginReq); err != nil {
		return response.BadRequest(c, "Invalid request format", err)
	}

	if err := bindValidate(c, &loginReq); err != nil {
		return err
	}

	ipAddress := c.IP()
	userAgent := c.Get(fiber.HeaderUserAgent)

	token, refreshToken, user, err := h.authService.Login(c.Context(), loginReq.Identifier, loginReq.Password, ipAddress, userAgent)
	if errors.Is(err, apperrors.ErrInvalidCredentials) {
		return response.Unauthorized(c, "Invalid identifier or password")
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

func (h *AuthHandler) ForgotPassword(c fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request format", err)
	}

	if err := bindValidate(c, &req); err != nil {
		return err
	}

	ipAddress := c.IP()
	userAgent := c.Get(fiber.HeaderUserAgent)

	err := h.authService.ForgotPassword(c.Context(), req.Email, ipAddress, userAgent)
	if err != nil {
		return response.Success(c, "If the email exists, a password reset link has been sent", nil)
	}

	return response.Success(c, "If the email exists, a password reset link has been sent", nil)
}

func (h *AuthHandler) ResetPassword(c fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request format", err)
	}

	if err := bindValidate(c, &req); err != nil {
		return err
	}

	ipAddress := c.IP()
	userAgent := c.Get(fiber.HeaderUserAgent)

	err := h.authService.ResetPassword(c.Context(), req.Token, req.Password, ipAddress, userAgent)
	if errors.Is(err, apperrors.ErrInvalidToken) {
		return response.BadRequest(c, "Invalid or expired reset token", err)
	}
	if errors.Is(err, apperrors.ErrPasswordResetTokenUsed) {
		return response.BadRequest(c, "Reset token has already been used", err)
	}
	if errors.Is(err, apperrors.ErrPasswordResetTokenExpired) {
		return response.BadRequest(c, "Reset token has expired", err)
	}
	if err != nil {
		return response.InternalServerError(c, "Failed to reset password", err)
	}

	return response.Success(c, "Password reset successful", nil)
}

func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	var req dto.RefreshTokenRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request format", err)
	}

	if err := bindValidate(c, &req); err != nil {
		return err
	}

	ipAddress := c.IP()
	userAgent := c.Get(fiber.HeaderUserAgent)

	token, refreshToken, user, err := h.authService.RefreshToken(c.Context(), req.RefreshToken, ipAddress, userAgent)
	if errors.Is(err, apperrors.ErrInvalidToken) {
		return response.Unauthorized(c, "Invalid or expired refresh token")
	}
	if errors.Is(err, apperrors.ErrTokenExpired) {
		return response.Unauthorized(c, "Refresh token has expired")
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
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req dto.ChangePasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request format", err)
	}

	if err := bindValidate(c, &req); err != nil {
		return err
	}

	ipAddress := c.IP()
	userAgent := c.Get(fiber.HeaderUserAgent)

	err := h.authService.ChangePassword(c.Context(), userID, req.CurrentPassword, req.NewPassword, ipAddress, userAgent)
	if errors.Is(err, apperrors.ErrInvalidCredentials) {
		return response.Unauthorized(c, "Current password is incorrect")
	}
	if err != nil {
		return response.InternalServerError(c, "Failed to change password", err)
	}

	return response.Success(c, "Password changed successfully", nil)
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	var req dto.LogoutRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request format", err)
	}

	if err := bindValidate(c, &req); err != nil {
		return err
	}

	userID, _ := GetUserIDFromClaims(c)
	ipAddress := c.IP()
	userAgent := c.Get(fiber.HeaderUserAgent)

	err := h.authService.Logout(c.Context(), req.RefreshToken)
	if err != nil {
		return response.Success(c, "Logout successful", nil)
	}

	h.activityService.LogActivity(c.Context(), &userID, model.ActivityLogout, model.StatusSuccess, ipAddress, userAgent, nil, nil)

	return response.Success(c, "Logout successful", nil)
}

func (h *AuthHandler) GetProfile(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	user, err := h.authService.GetProfile(c.Context(), userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to get profile", err)
	}

	return response.Success(c, "Profile retrieved successfully", map[string]any{
		"id":              user.ID,
		"email":           user.Email,
		"username":        user.Username,
		"first_name":      user.FirstName,
		"last_name":       user.LastName,
		"image":           user.Image,
		"is_super_admin":  user.IsSuperAdmin,
		"followers_count": user.FollowersCount,
		"following_count": user.FollowingCount,
	})
}

func (h *AuthHandler) GetActivityLogs(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	limit, offset := ParsePaginationParams(c, 20)
	activityType := c.Query("activity_type")

	logs, totalCount, err := h.activityService.GetActivityLogs(c.Context(), userID, activityType, limit, offset)
	if err != nil {
		return response.InternalServerError(c, "Failed to get activity logs", err)
	}

	meta := response.CalculatePaginationMeta(totalCount, offset, limit)
	return response.SuccessWithMeta(c, "Activity logs retrieved successfully", logs, meta)
}

func (h *AuthHandler) GetRecentActivity(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	logs, err := h.activityService.GetRecentActivity(c.Context(), userID, limit)
	if err != nil {
		return response.InternalServerError(c, "Failed to get recent activity", err)
	}

	return response.Success(c, "Recent activity retrieved successfully", logs)
}

func (h *AuthHandler) GetFailedLogins(c fiber.Ctx) error {
	_, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	limit, offset := ParsePaginationParams(c, 20)

	hours := 24
	if hParam := c.Query("since_hours"); hParam != "" {
		if parsed, err := strconv.Atoi(hParam); err == nil && parsed > 0 {
			hours = parsed
		}
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	logs, totalCount, err := h.activityService.GetFailedLogins(c.Context(), since, limit, offset)
	if err != nil {
		return response.InternalServerError(c, "Failed to get failed logins", err)
	}

	meta := response.CalculatePaginationMeta(totalCount, offset, limit)
	return response.SuccessWithMeta(c, "Failed logins retrieved successfully", logs, meta)
}

func (h *AuthHandler) GithubOAuthRedirect(c fiber.Ctx) error {
	state, err := generateOAuthState()
	if err != nil {
		return response.InternalServerError(c, "Failed to start GitHub OAuth", err)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "github_oauth_state",
		Value:    state,
		Path:     "/api/auth/oauth/github",
		MaxAge:   int((10 * time.Minute).Seconds()),
		HTTPOnly: true,
		Secure:   strings.HasPrefix(h.frontendConfig.URL, "https://"),
		SameSite: fiber.CookieSameSiteLaxMode,
	})

	authURL := h.authService.GetGithubOAuthURL(state)
	return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(authURL)
}

func (h *AuthHandler) GithubOAuthCallback(c fiber.Ctx) error {
	callbackURL := h.frontendConfig.OAuthCallbackURL
	clearGithubOAuthStateCookie(c)

	code := c.Query("code")
	if code == "" {
		return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(callbackURL + "?error=missing_code")
	}
	state := c.Query("state")
	stateCookie := c.Cookies("github_oauth_state")
	if state == "" || subtle.ConstantTimeCompare([]byte(state), []byte(stateCookie)) != 1 {
		return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(callbackURL + "?error=invalid_state")
	}

	ipAddress := c.IP()
	userAgent := c.Get(fiber.HeaderUserAgent)

	githubToken, err := h.authService.GetGithubToken(c.Context(), code)
	if err != nil {
		h.activityService.LogActivity(c.Context(), nil, model.ActivityOAuthLoginFailed, model.StatusFailure, ipAddress, userAgent, nil, map[string]any{"provider": "github", "error": err.Error()})
		return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(callbackURL + "?error=github_token_failed")
	}

	githubUser, err := fetchGithubUser(c.Context(), githubToken)
	if err != nil {
		h.activityService.LogActivity(c.Context(), nil, model.ActivityOAuthLoginFailed, model.StatusFailure, ipAddress, userAgent, nil, map[string]any{"provider": "github", "error": err.Error()})
		return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(callbackURL + "?error=github_user_failed")
	}

	if githubUser.Email == nil || *githubUser.Email == "" {
		email, err := fetchGithubUserEmail(c.Context(), githubToken)
		if err == nil && email != "" {
			githubUser.Email = &email
		}
	}

	accessToken, refreshToken, user, err := h.authService.SignInWithGithub(c.Context(), githubUser, ipAddress, userAgent)
	if err != nil {
		return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(callbackURL + "?error=oauth_login_failed")
	}

	exchangeCode, err := h.authService.CreateOAuthExchangeCode(c.Context(), accessToken, refreshToken, user)
	if err != nil {
		return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(callbackURL + "?error=oauth_exchange_failed")
	}

	redirectURL := appendQueryParam(callbackURL, "code", exchangeCode)
	return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(redirectURL)
}

func (h *AuthHandler) ExchangeOAuthCode(c fiber.Ctx) error {
	var req dto.OAuthExchangeRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request format", err)
	}

	if err := bindValidate(c, &req); err != nil {
		return err
	}

	accessToken, refreshToken, user, err := h.authService.ExchangeOAuthCode(c.Context(), req.Code)
	if errors.Is(err, apperrors.ErrInvalidToken) {
		return response.Unauthorized(c, "Invalid or expired OAuth code")
	}
	if err != nil {
		return response.InternalServerError(c, "Failed to exchange OAuth code", err)
	}

	return response.Success(c, "OAuth code exchanged successfully", map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": map[string]any{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
		},
	})
}

func fetchGithubUser(ctx context.Context, token string) (*service.GithubUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user service.GithubUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func fetchGithubUserEmail(ctx context.Context, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary {
			return e.Email, nil
		}
	}
	if len(emails) > 0 {
		return emails[0].Email, nil
	}
	return "", nil
}

func generateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func clearGithubOAuthStateCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "github_oauth_state",
		Value:    "",
		Path:     "/api/auth/oauth/github",
		MaxAge:   -1,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
}

func appendQueryParam(rawURL, key, value string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
	return u.String()
}
