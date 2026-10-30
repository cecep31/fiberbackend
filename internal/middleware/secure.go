package middleware

import (
	"net"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// Security headers, byte-identical to echobackend's middleware.SecureConfig
// and axumbackend's security-header layer. Fiber's helmet middleware is not
// used because it always adds extra headers (COEP/COOP/CORP,
// Origin-Agent-Cluster, X-DNS-Prefetch-Control, ...) the other backends do not
// send, and spells the HSTS directive differently.
const (
	headerXSSProtection         = "1; mode=block"
	headerContentTypeNosniff    = "nosniff"
	headerXFrameOptions         = "SAMEORIGIN"
	headerContentSecurityPolicy = "default-src 'self'"
	headerReferrerPolicy        = "strict-origin-when-cross-origin"
	// headerHSTS is only sent over HTTPS (see isHTTPS).
	headerHSTS = "max-age=3600; includeSubdomains"
)

// SecureHeaders sets the shared security headers on every response.
func SecureHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set(fiber.HeaderXXSSProtection, headerXSSProtection)
		c.Set(fiber.HeaderXContentTypeOptions, headerContentTypeNosniff)
		c.Set(fiber.HeaderXFrameOptions, headerXFrameOptions)
		if isHTTPS(c) {
			c.Set(fiber.HeaderStrictTransportSecurity, headerHSTS)
		}
		c.Set(fiber.HeaderContentSecurityPolicy, headerContentSecurityPolicy)
		c.Set(fiber.HeaderReferrerPolicy, headerReferrerPolicy)
		return c.Next()
	}
}

// isHTTPS mirrors Echo v5's default Context.Scheme() == "https": the
// connection itself is TLS, or the direct peer is a trusted proxy (loopback,
// private or link-local address) that forwarded an https scheme. Forwarding
// headers from any other peer are ignored so a client cannot fake HTTPS.
func isHTTPS(c fiber.Ctx) bool {
	if c.RequestCtx().IsTLS() {
		return true
	}
	if !isTrustedSchemePeer(c.RequestCtx().RemoteIP()) {
		return false
	}
	return forwardedScheme(c) == "https"
}

func isTrustedSchemePeer(ip net.IP) bool {
	return ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast())
}

// forwardedScheme follows Echo's schemeFromHeaders precedence: when
// X-Forwarded-Proto is present only its last value counts (added by the
// nearest proxy); otherwise X-Forwarded-Protocol, X-Forwarded-Ssl=on and
// X-Url-Scheme are consulted in that order.
func forwardedScheme(c fiber.Ctx) string {
	header := &c.Request().Header
	if values := header.PeekAll(fiber.HeaderXForwardedProto); len(values) > 0 && len(values[len(values)-1]) > 0 {
		last := string(values[len(values)-1])
		if i := strings.LastIndexByte(last, ','); i != -1 {
			last = last[i+1:]
		}
		if scheme, ok := canonicalProto(strings.TrimSpace(last)); ok {
			return scheme
		}
		return "http"
	}
	if scheme, ok := canonicalProto(string(header.Peek(fiber.HeaderXForwardedProtocol))); ok {
		return scheme
	}
	if string(header.Peek(fiber.HeaderXForwardedSsl)) == "on" {
		return "https"
	}
	if scheme, ok := canonicalProto(string(header.Peek(fiber.HeaderXUrlScheme))); ok {
		return scheme
	}
	return "http"
}

func canonicalProto(proto string) (string, bool) {
	for _, p := range [...]string{"http", "https", "ws", "wss"} {
		if strings.EqualFold(proto, p) {
			return p, true
		}
	}
	return "", false
}
