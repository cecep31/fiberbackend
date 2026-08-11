package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"fiberbackend/pkg/applog"

	"github.com/gofiber/fiber/v3"
)

var recoverLog = applog.Component("recover")

// RecoverWithLog returns middleware that recovers panics and logs them with slog.
func RecoverWithLog() fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				if rErr, ok := r.(error); ok && errors.Is(rErr, http.ErrAbortHandler) {
					panic(r)
				}

				tmpErr, ok := r.(error)
				if !ok {
					tmpErr = fmt.Errorf("%v", r)
				}

				stack := make([]byte, 4<<10)
				length := runtime.Stack(stack, false)
				recoverLog.Error("panic recovered",
					"error", tmpErr,
					"method", c.Method(),
					"uri", c.OriginalURL(),
					"stack", string(stack[:length]),
				)
				err = tmpErr
			}
		}()
		return c.Next()
	}
}
