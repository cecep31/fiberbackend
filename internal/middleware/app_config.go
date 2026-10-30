package middleware

import (
	"time"

	"fiberbackend/config"

	"github.com/gofiber/fiber/v3"
)

// BodyLimit is the global request body cap (10 MB). fasthttp rejects larger
// bodies before routing; ErrorHandler renders that as the shared 413 envelope.
const BodyLimit = 10 * 1024 * 1024

// AppConfig returns the fiber.Config used by cmd/main.go: body limit,
// timeouts, the enveloped ErrorHandler and client-IP resolution.
//
// With HTTP_TRUST_PROXY, c.IP() matches echobackend's
// ExtractIPFromXFFHeader: when the direct peer is a trusted proxy (loopback,
// private, link-local or CGNAT 100.64.0.0/10), X-Forwarded-For is walked
// right-to-left and the first hop outside those ranges is the client IP (the
// leftmost valid hop when all are trusted; the peer when the header is
// missing). EnableIPValidation is required: without it Fiber returns the raw,
// client-controlled header value. Without HTTP_TRUST_PROXY the peer address is
// used and proxy headers are ignored.
func AppConfig(conf *config.Config) fiber.Config {
	cfg := fiber.Config{
		AppName:      "fiberbackend",
		BodyLimit:    BodyLimit,
		ErrorHandler: ErrorHandler,
		// ReadTimeout covers the full request body read. For most endpoints 60 s
		// is fine, but file uploads can be slow on poor connections; the 10 MB
		// body limit bounds abuse. fasthttp applies WriteTimeout once, as an
		// absolute deadline set when the response starts (not per chunk), so
		// long-lived streams such as the guild SSE endpoint refresh the conn's
		// write deadline themselves before every frame.
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	if conf.HTTP.TrustProxy {
		cfg.TrustProxy = true
		cfg.ProxyHeader = fiber.HeaderXForwardedFor
		cfg.EnableIPValidation = true
		cfg.TrustProxyConfig = fiber.TrustProxyConfig{
			Loopback:  true,
			Private:   true,
			LinkLocal: true,
			Proxies:   conf.TrustedProxyCIDRs(),
		}
	}
	return cfg
}
