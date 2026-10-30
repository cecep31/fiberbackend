package middleware

import (
	"errors"

	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

// ErrorHandler is the app-wide Fiber error handler (fiber.Config.ErrorHandler).
// It replaces Fiber's default text/plain body ("Cannot GET /x", "Request
// Entity Too Large") with the shared response envelope so router-level errors
// (unknown route 404, wrong method 405), the fasthttp body-limit 413,
// recovered panics and any other error a handler returns without writing a
// response all look like every other API error. echobackend's
// HTTPErrorHandler and axumbackend's router fallbacks produce the same bodies.
func ErrorHandler(c fiber.Ctx, err error) error {
	code, message := errorStatus(err)

	if code >= fiber.StatusInternalServerError {
		httpLog.Error("unhandled request error",
			"method", c.Method(),
			"uri", c.OriginalURL(),
			"status", code,
			"error", err,
		)
	}

	if c.Method() == fiber.MethodHead {
		return c.SendStatus(code)
	}
	return response.HTTPStatus(c, code, message)
}

// errorStatus returns the HTTP status and message carried by err: the
// *fiber.Error code/message, or 500 with no message for any other error.
func errorStatus(err error) (int, string) {
	if fe, ok := errors.AsType[*fiber.Error](err); ok && fe.Code != 0 {
		return fe.Code, fe.Message
	}
	return fiber.StatusInternalServerError, ""
}
