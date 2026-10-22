package routes

import (
	"time"

	appmiddleware "fiberbackend/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

func (r *Routes) setupAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")
	loginRateLimit := appmiddleware.FixedWindowRateLimiterWithCache(r.cache, "auth:login", 5, 5*time.Minute)
	registerRateLimit := appmiddleware.FixedWindowRateLimiterWithCache(r.cache, "auth:register", 5, 5*time.Minute)
	forgotPasswordRateLimit := appmiddleware.FixedWindowRateLimiterWithCache(r.cache, "auth:forgot-password", 3, 5*time.Minute)
	resetPasswordRateLimit := appmiddleware.FixedWindowRateLimiterWithCache(r.cache, "auth:reset-password", 5, 5*time.Minute)
	refreshRateLimit := appmiddleware.FixedWindowRateLimiterWithCache(r.cache, "auth:refresh", 30, time.Minute)
	checkUsernameRateLimit := appmiddleware.FixedWindowRateLimiterWithCache(r.cache, "auth:check-username", 20, 5*time.Minute)
	oauthExchangeRateLimit := appmiddleware.FixedWindowRateLimiterWithCache(r.cache, "auth:oauth-exchange", 10, time.Minute)
	{
		auth.Post("/register", registerRateLimit, r.authHandler.Register)
		auth.Post("/login", loginRateLimit, r.authHandler.Login)
		auth.Post("/forgot-password", forgotPasswordRateLimit, r.authHandler.ForgotPassword)
		auth.Post("/reset-password", resetPasswordRateLimit, r.authHandler.ResetPassword)
		auth.Post("/refresh", refreshRateLimit, r.authHandler.RefreshToken)
		auth.Post("/check-username", checkUsernameRateLimit, r.authHandler.CheckUsername)
		auth.Post("/logout", r.authMiddleware.Auth(), r.authHandler.Logout)
		auth.Get("/profile", r.authMiddleware.Auth(), r.authHandler.GetProfile)
		auth.Put("/profile", r.authMiddleware.Auth(), r.authHandler.UpdateProfile)
		auth.Delete("/account", r.authMiddleware.Auth(), r.authHandler.DeleteAccount)
		auth.Patch("/password", r.authMiddleware.Auth(), r.authHandler.ChangePassword)
		auth.Get("/activity-logs", r.authMiddleware.Auth(), r.authHandler.GetActivityLogs)
		auth.Get("/activity-logs/recent", r.authMiddleware.Auth(), r.authHandler.GetRecentActivity)
		auth.Get("/activity-logs/failed-logins", r.authMiddleware.Auth(), r.authMiddleware.AuthAdmin(), r.authHandler.GetFailedLogins)
		// Admin session termination (OWASP ASVS v5.0 7.4.5). The platform-wide
		// revoke is a POST, not a DELETE on the collection, so it cannot be
		// triggered by trimming a user id off the URL below — and it takes a
		// confirmation body.
		auth.Post("/sessions/revoke-all", r.authMiddleware.Auth(), r.authMiddleware.AuthAdmin(), r.authHandler.RevokeAllSessions)
		auth.Delete("/sessions/:userId", r.authMiddleware.Auth(), r.authMiddleware.AuthAdmin(), r.authHandler.RevokeUserSessions)
		auth.Get("/oauth/github", r.authHandler.GithubOAuthRedirect)
		auth.Get("/oauth/github/callback", r.authHandler.GithubOAuthCallback)
		auth.Post("/oauth/exchange", oauthExchangeRateLimit, r.authHandler.ExchangeOAuthCode)
	}
}
