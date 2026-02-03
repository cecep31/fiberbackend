package routes

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func (r *Routes) setupAuthRoutes(v1 fiber.Router) {
	auth := v1.Group("/auth")
	loginLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 5 * time.Minute,
	})
	{
		auth.Post("/register", r.authHandler.Register)
		auth.Post("/login", loginLimiter, r.authHandler.Login)
		auth.Post("/check-username", r.authHandler.CheckUsername)
		auth.Post("/forgot-password", loginLimiter, r.authHandler.ForgotPassword)
		auth.Post("/reset-password", r.authHandler.ResetPassword)
		auth.Post("/refresh", r.authHandler.RefreshToken)
		auth.Put("/change-password", r.authMiddleware.Auth(), r.authHandler.ChangePassword)
	}
}
