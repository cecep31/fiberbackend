package middleware

import (
	"time"

	"fiberbackend/config"
	"fiberbackend/pkg/applog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
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
// fiber.Config built by AppConfig. Errors returned through this chain (404/405
// from the router, 413 from the body limit, recovered panics) are rendered by
// ErrorHandler, which AppConfig installs as fiber.Config.ErrorHandler.
func InitMiddleware(app *fiber.App, config *config.Config) {
	// Enhanced request logging with structured format
	app.Use(func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		// A returned error is rendered by ErrorHandler after this middleware
		// unwinds, so log the status it is about to send.
		status := c.Response().StatusCode()
		if err != nil {
			status, _ = errorStatus(err)
		}

		latency := time.Since(start)
		httpLog.Info("handled request",
			"method", c.Method(),
			"uri", c.OriginalURL(),
			"status", status,
			"latency_ms", float64(latency.Nanoseconds())/1e6,
			"remote_ip", c.IP(),
		)

		return err
	})

	app.Use(RecoverWithLog())

	app.Use(cors.New(cors.Config{AllowOrigins: config.HTTP.AllowOrigins}))

	// Add security headers (same set and values as echobackend)
	app.Use(SecureHeaders())

	// Global HTTP rate limit (per-IP token bucket; 0 = disabled).
	if config.HTTP.RateLimitRPS > 0 {
		app.Use(GlobalRateLimiter(config.HTTP.RateLimitRPS, config.HTTP.RateLimitWindow))
	}
}
