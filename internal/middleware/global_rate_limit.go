package middleware

import (
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/time/rate"
)

// defaultGlobalRateLimitIdleExpiry matches Echo's RateLimiterMemoryStore
// default ExpiresIn.
const defaultGlobalRateLimitIdleExpiry = 3 * time.Minute

// GlobalRateLimiter limits each client IP (c.IP()) with an in-memory token
// bucket: rps tokens per second, burst 2*rps. idleExpiry only sets how long an
// idle visitor is remembered (0 = 3 minutes); it is not a counting window. A
// denied request gets 429 with the shared envelope and a Retry-After header.
// It is a port of echobackend's GlobalRateLimiter (Echo's
// RateLimiterMemoryStore) so both backends throttle identically.
func GlobalRateLimiter(rps int, idleExpiry time.Duration) fiber.Handler {
	store := newTokenBucketStore(rps, idleExpiry)
	retryAfter := strconv.Itoa(globalRetryAfterSeconds(rps))

	return func(c fiber.Ctx) error {
		if store.allow(c.IP(), time.Now()) {
			return c.Next()
		}
		c.Set(fiber.HeaderRetryAfter, retryAfter)
		return response.TooManyRequests(c, response.MsgTooManyAttempts)
	}
}

// globalRetryAfterSeconds is the time until a drained bucket refills one
// token (1/rps), rounded up and never below one second.
func globalRetryAfterSeconds(rps int) int {
	if rps <= 0 {
		return 1
	}
	return max(int(math.Ceil(1/float64(rps))), 1)
}

type tokenBucketVisitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type tokenBucketStore struct {
	mu          sync.Mutex
	visitors    map[string]*tokenBucketVisitor
	rate        rate.Limit
	burst       int
	expiresIn   time.Duration
	lastCleanup time.Time
}

func newTokenBucketStore(rps int, idleExpiry time.Duration) *tokenBucketStore {
	if idleExpiry <= 0 {
		idleExpiry = defaultGlobalRateLimitIdleExpiry
	}
	return &tokenBucketStore{
		visitors:    make(map[string]*tokenBucketVisitor),
		rate:        rate.Limit(rps),
		burst:       rps * 2,
		expiresIn:   idleExpiry,
		lastCleanup: time.Now(),
	}
}

func (s *tokenBucketStore) allow(identifier string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	visitor, ok := s.visitors[identifier]
	if !ok {
		visitor = &tokenBucketVisitor{limiter: rate.NewLimiter(s.rate, s.burst)}
		// c.IP() may alias fasthttp's request buffer (X-Forwarded-For), which
		// is reused after the request; the map key must own its bytes.
		s.visitors[strings.Clone(identifier)] = visitor
	}
	visitor.lastSeen = now
	if now.Sub(s.lastCleanup) > s.expiresIn {
		s.cleanupLocked(now)
	}
	return visitor.limiter.AllowN(now, 1)
}

func (s *tokenBucketStore) cleanupLocked(now time.Time) {
	for id, visitor := range s.visitors {
		if now.Sub(visitor.lastSeen) > s.expiresIn {
			delete(s.visitors, id)
		}
	}
	s.lastCleanup = now
}
