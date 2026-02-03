package middleware

import (
	"log"
	"time"

	"fiberbackend/config"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func InitMiddleware(app *fiber.App, config *config.Config) {
	// Recover first so any panic in downstream middleware/handlers is caught
	app.Use(recover.New())
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

	// Enhanced request logging with structured format
	app.Use(func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)
		log.Printf(
			"handled request method=%s uri=%s request_id=%s status=%d latency=%.3f ms remote_ip=%s",
			c.Method(),
			c.OriginalURL(),
			c.Get(fiber.HeaderXRequestID),
			c.Response().StatusCode(),
			float64(latency.Nanoseconds())/1000000,
			c.IP(),
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

	app.Use(compress.New())
	app.Use(cors.New(cors.Config{AllowOrigins: []string{"*"}}))
}
