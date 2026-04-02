package middleware

import (
	"time"

	"fiberbackend/config"
	"fiberbackend/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func InitMiddleware(app *fiber.App, config *config.Config) {
	// Create logger for middleware
	log := logger.New(config.ParseLogLevel(), config.LogFormat, false)

	// Recover first so any panic in downstream middleware/handlers is caught
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))
	app.Use(requestid.New())

	// Add security headers
	app.Use(helmet.New(helmet.Config{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))

	// Request logging with structured format
	app.Use(func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		latency := time.Since(start)
		statusCode := c.Response().StatusCode()

		// Log with structured fields
		log.Info("request handled",
			logger.String("method", c.Method()),
			logger.String("uri", c.OriginalURL()),
			logger.String("request_id", c.Get(fiber.HeaderXRequestID)),
			logger.Int("status", statusCode),
			logger.Duration("latency_ms", latency),
			logger.String("remote_ip", c.IP()),
			logger.String("user_agent", c.Get(fiber.HeaderUserAgent)),
		)

		return err
	})

	// Enhanced rate limiting with custom store and configuration
	if config.RateLimiterMax > 0 {
		app.Use(limiter.New(limiter.Config{
			Max:        config.RateLimiterMax,
			Expiration: 1 * time.Minute,
		}))
	}

	app.Use(cors.New(cors.Config{AllowOrigins: []string{"*"}}))
}
