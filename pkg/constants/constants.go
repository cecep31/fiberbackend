// Package constants provides application-wide constant values.
package constants

// Pagination defaults
const (
	// DefaultLimit is the default number of items per page
	DefaultLimit = 10
	// MaxLimit is the maximum allowed items per page
	MaxLimit = 100
	// DefaultOffset is the default offset for pagination
	DefaultOffset = 0
)

// Post-related constants
const (
	// PostBodyTruncateLength is the number of characters to truncate post body in list views
	PostBodyTruncateLength = 250
	// RandomPostsDefaultLimit is the default limit for random posts
	RandomPostsDefaultLimit = 6
	// RandomPostsMaxLimit is the maximum limit for random posts
	RandomPostsMaxLimit = 20
)

// JWT-related constants
const (
	// JWTTokenExpiry is the duration until JWT token expires
	JWTTokenExpiryHours = 6
	// RefreshTokenLength is the byte length for refresh tokens
	RefreshTokenLength = 64
	// PasswordResetTokenLength is the byte length for password reset tokens
	PasswordResetTokenLength = 32
	// PasswordResetExpiryHours is the duration until password reset token expires
	PasswordResetExpiryHours = 1
)

// Token prefixes
const (
	// RefreshTokenPrefix is the prefix for refresh tokens
	RefreshTokenPrefix = "pl_"
	// PasswordResetTokenPrefix is the prefix for password reset tokens
	PasswordResetTokenPrefix = "pr_"
)

// Context keys for Fiber context
const (
	// UserContextKey is the key for storing user claims in context
	UserContextKey = "user"
	// UserIDContextKey is the key for storing user ID in context
	UserIDContextKey = "user_id"
)
