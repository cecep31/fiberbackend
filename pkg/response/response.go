package response

import (
	"log/slog"
	"net/http"

	"fiberbackend/pkg/logger"
	"fiberbackend/pkg/validator"

	"github.com/gofiber/fiber/v3"
)

// apiLogger is the logger for response package
var apiLogger = logger.New(slog.LevelInfo, "text", false)

// APIResponse represents the standard API response format
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
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
	apiLogger.Info(message,
		logger.String("type", "success"),
		logger.String("request_id", c.Get(fiber.HeaderXRequestID)),
	)

	return c.Status(http.StatusOK).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithMeta sends a successful response with metadata
func SuccessWithMeta(c fiber.Ctx, message string, data any, meta any) error {
	apiLogger.Info(message,
		logger.String("type", "success"),
		logger.String("request_id", c.Get(fiber.HeaderXRequestID)),
	)

	return c.Status(http.StatusOK).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created sends a created response
func Created(c fiber.Ctx, message string, data any) error {
	apiLogger.Info(message,
		logger.String("type", "created"),
		logger.String("request_id", c.Get(fiber.HeaderXRequestID)),
	)

	return c.Status(http.StatusCreated).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// HandleBindError handles errors from c.Bind().Body(). If the error is validation
// errors from the StructValidator, returns 400 with Message "Validation failed" and
// Data set to the list of field errors. Otherwise returns BadRequest.
func HandleBindError(c fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		apiLogger.Warn("Validation failed",
			logger.String("request_id", c.Get(fiber.HeaderXRequestID)),
			logger.Int("error_count", len(validationErrors.Errors)),
		)
		return c.Status(http.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   validationErrors.Error(),
			Data:    validationErrors.Errors,
		})
	}
	return BadRequest(c, "Invalid request format", err)
}

// BadRequest sends a bad request error response
func BadRequest(c fiber.Ctx, message string, err error) error {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	apiLogger.Warn(message,
		logger.String("request_id", c.Get(fiber.HeaderXRequestID)),
		logger.String("error", errorMsg),
	)

	return c.Status(http.StatusBadRequest).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   errorMsg,
	})
}

// Unauthorized sends an unauthorized error response
func Unauthorized(c fiber.Ctx, message string) error {
	apiLogger.Warn(message,
		logger.String("request_id", c.Get(fiber.HeaderXRequestID)),
		logger.String("type", "unauthorized"),
	)

	return c.Status(http.StatusUnauthorized).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   "Unauthorized access",
	})
}

// Forbidden sends a forbidden error response
func Forbidden(c fiber.Ctx, message string) error {
	apiLogger.Warn(message,
		logger.String("request_id", c.Get(fiber.HeaderXRequestID)),
		logger.String("type", "forbidden"),
	)

	return c.Status(http.StatusForbidden).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   "Access forbidden",
	})
}

// NotFound sends a not found error response
func NotFound(c fiber.Ctx, message string, err error) error {
	errorMsg := "Resource not found"
	if err != nil {
		errorMsg = err.Error()
	}

	apiLogger.Warn(message,
		logger.String("request_id", c.Get(fiber.HeaderXRequestID)),
		logger.String("error", errorMsg),
	)

	return c.Status(http.StatusNotFound).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   errorMsg,
	})
}

// InternalServerError sends an internal server error response
func InternalServerError(c fiber.Ctx, message string, err error) error {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	apiLogger.Error(message,
		logger.String("request_id", c.Get(fiber.HeaderXRequestID)),
		logger.String("error", errorMsg),
	)

	return c.Status(http.StatusInternalServerError).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   errorMsg,
	})
}

// ValidationError sends a validation error response
func ValidationError(c fiber.Ctx, message string, err error) error {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	apiLogger.Warn(message,
		logger.String("request_id", c.Get(fiber.HeaderXRequestID)),
		logger.String("error", errorMsg),
	)

	return c.Status(http.StatusUnprocessableEntity).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   errorMsg,
	})
}

// CalculatePaginationMeta calculates pagination metadata
func CalculatePaginationMeta(totalItems int64, offset, limit int) PaginationMeta {
	totalPages := int(totalItems) / limit
	if int(totalItems)%limit > 0 {
		totalPages++
	}

	return PaginationMeta{
		TotalItems: int(totalItems),
		Offset:     offset,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
