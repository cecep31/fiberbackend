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

// InitMiddleware registers global middleware. Order matters — each entry wraps
// everything registered after it:
//
//  1. Request logger: outermost, so every response (including recovered panics
//     and 429s) is logged with its final status.
//  2. Recover: catches panics in all later middleware and handlers.
//  3. CORS: answers preflight requests before they reach the rate limiter, and
//     sets CORS headers before inner middleware can reject the request, so
//     browsers can read 413/429 error bodies instead of reporting a CORS error.
//  4. Secure headers, rate limit.
//
// Body limit (10 MB) is enforced via fasthttp.MaxRequestBodySize in the
// fiber.Config passed to fiber.New() (see cmd/main.go).
func InitMiddleware(app *fiber.App, config *config.Config) {
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

	app.Use(RecoverWithLog())

	app.Use(cors.New(cors.Config{AllowOrigins: config.HTTP.AllowOrigins}))

	// Add security headers
	app.Use(helmet.New(helmet.Config{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))

	// Global HTTP rate limit (sustained RPS with 2x burst; 0 = disabled).
	// RateLimitWindow overrides the per-key window when set, mirroring echo's
	// memory-store ExpiresIn.
	if config.HTTP.RateLimitRPS > 0 {
		expiration := time.Second
		if config.HTTP.RateLimitWindow > 0 {
			expiration = config.HTTP.RateLimitWindow
		}
		app.Use(limiter.New(limiter.Config{
			Max:        config.HTTP.RateLimitRPS * 2, // burst = 2x sustained RPS
			Expiration: expiration,
		}))
	}
}
