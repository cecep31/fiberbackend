package response

import (
	"errors"
	"net/http"

	"fiberbackend/pkg/applog"
	"fiberbackend/pkg/validator"

	"github.com/gofiber/fiber/v3"
)

var log = applog.Component("api")

// APIResponse represents the standard API response format
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Errors  any    `json:"errors,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	TotalItems int `json:"total_items"`
	Offset     int `json:"offset"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

// Success sends a successful response
func Success(c fiber.Ctx, message string, data any) error {
	return c.Status(http.StatusOK).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithMeta sends a successful response with metadata
func SuccessWithMeta(c fiber.Ctx, message string, data any, meta any) error {
	return c.Status(http.StatusOK).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created sends a created response
func Created(c fiber.Ctx, message string, data any) error {
	return c.Status(http.StatusCreated).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// BadRequest sends a bad request error response
func BadRequest(c fiber.Ctx, message string, err error) error {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	log.Warn("bad request",
		"message", message,
		"error", errorMsg,
	)

	return c.Status(http.StatusBadRequest).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   errorMsg,
	})
}

// Unauthorized sends an unauthorized error response
func Unauthorized(c fiber.Ctx, message string) error {
	log.Warn("unauthorized",
		"message", message,
	)

	return c.Status(http.StatusUnauthorized).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   "Unauthorized access",
	})
}

// Forbidden sends a forbidden error response
func Forbidden(c fiber.Ctx, message string) error {
	log.Warn("forbidden",
		"message", message,
	)

	return c.Status(http.StatusForbidden).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   "Access forbidden",
	})
}

// TooManyRequests sends a 429 rate-limit response.
func TooManyRequests(c fiber.Ctx, message string) error {
	log.Warn("too many requests",
		"message", message,
	)

	return c.Status(http.StatusTooManyRequests).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   "Rate limit exceeded",
	})
}

// NotFound sends a not found error response
func NotFound(c fiber.Ctx, message string, err error) error {
	errorMsg := "Resource not found"
	if err != nil {
		errorMsg = err.Error()
	}

	log.Warn("not found",
		"message", message,
		"error", errorMsg,
	)

	return c.Status(http.StatusNotFound).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   errorMsg,
	})
}

// InternalServerError sends an internal server error response.
// The raw error is logged server-side only; the client receives a generic message
// to avoid leaking internal details (DSN, stack traces, etc.).
func InternalServerError(c fiber.Ctx, message string, err error) error {
	log.Error("internal server error",
		"message", message,
		"error", err,
	)

	return c.Status(http.StatusInternalServerError).JSON(APIResponse{
		Success: false,
		Message: message,
		// Do NOT include err.Error() in the response — avoids leaking internal details.
	})
}

// ValidationError sends a validation error response
func ValidationError(c fiber.Ctx, message string, err error) error {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	log.Warn("validation error",
		"message", message,
		"error", errorMsg,
	)

	return c.Status(http.StatusUnprocessableEntity).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   errorMsg,
	})
}

// FromValidateError maps validation errors to a unified response:
// structured field errors use 422 with Errors populated; otherwise ValidationError fallback.
func FromValidateError(c fiber.Ctx, err error) error {
	var errs validator.ValidationErrors
	if errors.As(err, &errs) {
		log.Warn("validation error",
			"error", errs.Error(),
		)
		return c.Status(http.StatusUnprocessableEntity).JSON(APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   errs.Error(),
			Errors:  errs.Errors,
		})
	}
	return ValidationError(c, "Validation failed", err)
}

// Conflict sends a 409 Conflict response (e.g. duplicate resource).
func Conflict(c fiber.Ctx, message string, conflictError string) error {
	log.Warn("conflict",
		"message", message,
	)
	return c.Status(http.StatusConflict).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   conflictError,
	})
}

// CalculatePaginationMeta calculates pagination metadata.
// Guards against division-by-zero when limit is 0.
func CalculatePaginationMeta(totalItems int64, offset, limit int) PaginationMeta {
	if limit <= 0 {
		limit = 10
	}

	total := int(totalItems)
	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	return PaginationMeta{
		TotalItems: total,
		Offset:     offset,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
