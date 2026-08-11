package middleware

import (
	"time"

	"fiberbackend/config"
	"fiberbackend/pkg/applog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

var httpLog = applog.Component("http")

func InitMiddleware(app *fiber.App, config *config.Config) {
	// Recover first so any panic in downstream middleware/handlers is caught
	app.Use(RecoverWithLog())

	// Body limit (10MB) is enforced via fasthttp.MaxRequestBodySize in the
	// fiber.Config passed to fiber.New() (see cmd/main.go).

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
		httpLog.Info("handled request",
			"method", c.Method(),
			"uri", c.OriginalURL(),
			"status", c.Response().StatusCode(),
			"latency_ms", float64(latency.Nanoseconds())/1e6,
			"remote_ip", c.IP(),
		)

		return err
	})

	// Global HTTP rate limit (sustained RPS, token bucket; 0 = disabled)
	if config.HTTP.RateLimitRPS > 0 {
		app.Use(limiter.New(limiter.Config{
			Max:        config.HTTP.RateLimitRPS * 2, // burst = 2x sustained RPS
			Expiration: time.Second,
		}))
	}

	app.Use(cors.New(cors.Config{AllowOrigins: config.HTTP.AllowOrigins}))
}
