package routes

import (
	"time"

	"fiberbackend/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

func (r *Routes) setupAIRoutes(api fiber.Router) {
	generateArticleRateLimit := middleware.FixedWindowRateLimiterWithCache(r.cache, "ai:generate-article", 20, time.Hour)

	ai := api.Group("/ai")
	// Super admins only for now (rollout); privileges are read from the DB.
	ai.Post("/generate-article", r.authMiddleware.Auth(), r.authMiddleware.AuthAdmin(), generateArticleRateLimit, r.aiHandler.GenerateArticle)
}
