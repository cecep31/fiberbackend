package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fiberbackend/config"
	"fiberbackend/internal/di"
	"fiberbackend/internal/middleware"
	"fiberbackend/pkg/logger"
	"fiberbackend/pkg/validator"

	"github.com/gofiber/fiber/v3"
)

func main() {
	// load config
	conf, errconf := config.Load()
	if errconf != nil {
		panic(errconf)
	}

	// Initialize structured logger
	log := logger.New(conf.ParseLogLevel(), conf.LogFormat, true)

	log.Info("starting application",
		logger.String("port", conf.AppPort),
		logger.String("log_level", conf.LogLevel),
		logger.String("log_format", conf.LogFormat),
		logger.Bool("debug", conf.Debug),
	)

	// Initialize dependency container
	container, err := di.NewContainer(conf)
	if err != nil {
		log.Error("failed to initialize container", logger.Err(err))
		panic(err)
	}

	// Initialize Fiber with custom error handler so API always returns JSON
	app := fiber.New(fiber.Config{
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		BodyLimit:       10 * 1024 * 1024,
		StructValidator: validator.NewValidator(),
		ErrorHandler:    jsonErrorHandler,
	})

	// Setup global middleware first (applies to all routes)
	middleware.InitMiddleware(app, conf)

	// Setup routes
	container.Routes.Setup(app)

	// Health and readiness endpoints
	app.Get("/health", healthCheck)
	app.Get("/ready", readinessCheck)

	app.Get("/", helloWorld)

	// Start server in a goroutine
	go func() {
		log.Info("starting server",
			logger.String("port", conf.AppPort),
			logger.String("address", "http://localhost:"+conf.AppPort),
		)
		if err := app.Listen(":" + conf.AppPort); err != nil && err != http.ErrServerClosed {
			log.Error("server shutdown error", logger.Err(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown Fiber server
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Error("server forced to shutdown", logger.Err(err))
	}

	// Cleanup resources
	if err := container.Cleanup.CleanupWithTimeout(5 * time.Second); err != nil {
		log.Warn("cleanup failed", logger.Err(err))
	} else {
		log.Info("resources cleaned up successfully")
	}

	log.Info("server exited")
}

func healthCheck(c fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(map[string]any{
		"status":  "healthy",
		"success": true,
	})
}

func readinessCheck(c fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(map[string]any{
		"status":  "ready",
		"success": true,
	})
}

func jsonErrorHandler(c fiber.Ctx, err error) error {
	code := http.StatusInternalServerError
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}
	return c.Status(code).JSON(map[string]any{
		"success": false,
		"message": "Request failed",
		"error":   err.Error(),
	})
}

func helloWorld(c fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(map[string]any{
		"message": "Hello, World!",
		"success": true,
	})
}
