package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/subosito/gotenv"
)

// Config represents the application configuration.
type Config struct {
	AppPort string

	// JWT
	JWTSecret        string
	JWTExpirationHrs int // JWT access token expiration in hours

	// Database
	DatabaseURL        string
	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxLifetime    time.Duration
	SlowQueryThreshold time.Duration

	// Rate limiter (0 = disabled)
	RateLimiterMax int
	RateLimiterTTL int

	Debug bool
}

// Load reads configuration from environment variables (with .env fallback).
// Returns an error if required fields are missing or invalid.
func Load() (*Config, error) {
	gotenv.Load()

	cfg := &Config{
		AppPort:            env("PORT", "8080"),
		JWTSecret:          env("JWT_SECRET", ""),
		JWTExpirationHrs:   envInt("JWT_EXPIRATION_HRS", 6),
		DatabaseURL:        env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
		MaxOpenConns:       envInt("MAX_OPEN_CONNS", 30),
		MaxIdleConns:       envInt("MAX_IDLE_CONNS", 1),
		ConnMaxLifetime:    envDuration("CONN_MAX_LIFETIME", 30*time.Minute),
		SlowQueryThreshold: envDuration("DB_SLOW_QUERY_THRESHOLD", 300*time.Millisecond),
		RateLimiterMax:     envInt("RATE_LIMITER_MAX", 0),
		RateLimiterTTL:     envInt("RATE_LIMITER_TTL", 60),
		Debug:              envBool("DEBUG", false),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate ensures required fields are present and valid.
func (c *Config) validate() error {
	switch {
	case c.JWTSecret == "":
		return fmt.Errorf("JWT_SECRET is required")
	case c.JWTSecret == "your-secret-key" || len(c.JWTSecret) < 32:
		return fmt.Errorf("JWT_SECRET must be at least 32 characters and not a placeholder value")
	case c.DatabaseURL == "":
		return fmt.Errorf("DATABASE_URL is required")
	}
	return nil
}

// env returns the value of the environment variable key, or defaultValue if not set.
func env(key, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultValue
}

// envInt returns the environment variable key parsed as int, or defaultValue on failure.
func envInt(key string, defaultValue int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

// envBool returns the environment variable key parsed as bool, or defaultValue on failure.
func envBool(key string, defaultValue bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultValue
}

// envDuration returns the environment variable key parsed as time.Duration, or defaultValue on failure.
func envDuration(key string, defaultValue time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultValue
}
