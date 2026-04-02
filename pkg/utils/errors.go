package utils

import (
	"errors"
	"fmt"
)

// AppError represents a structured application error
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"` // Not serialized to JSON
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Error codes
const (
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeConflict         = "CONFLICT"
	ErrCodeInternal         = "INTERNAL_ERROR"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
)

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError tries to convert an error to AppError
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// ValidationError creates a validation error
func ValidationError(message string, err error) *AppError {
	return NewAppError(ErrCodeValidationFailed, message, err)
}

// NotFoundError creates a not found error
func NotFoundError(message string, err error) *AppError {
	return NewAppError(ErrCodeNotFound, message, err)
}

// ConflictError creates a conflict error
func ConflictError(message string, err error) *AppError {
	return NewAppError(ErrCodeConflict, message, err)
}

// InternalError creates an internal error
func InternalError(message string, err error) *AppError {
	return NewAppError(ErrCodeInternal, message, err)
}

// UnauthorizedError creates an unauthorized error
func UnauthorizedError(message string, err error) *AppError {
	return NewAppError(ErrCodeUnauthorized, message, err)
}

// ForbiddenError creates a forbidden error
func ForbiddenError(message string, err error) *AppError {
	return NewAppError(ErrCodeForbidden, message, err)
}

// ============================================================================
// Domain Errors - Centralized error definitions for the application
// ============================================================================

// Auth Errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenExpired       = errors.New("token has expired")
)

// User Errors
var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUsernameTaken      = errors.New("username is already taken")
)

// Post Errors
var (
	ErrPostNotFound = errors.New("post not found")
)

// Comment Errors
var (
	ErrCommentNotFound = errors.New("comment not found")
)

// Tag Errors
var (
	ErrTagNotFound = errors.New("tag not found")
	ErrTagExists   = errors.New("tag already exists")
)

// Session Errors
var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session has expired")
)
