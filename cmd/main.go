package main

import (
	"context"
	"errors"
	"fiberbackend/config"
	"fiberbackend/internal/di"
	"fiberbackend/internal/middleware"
	"fiberbackend/internal/routes"
	"fiberbackend/pkg/validator"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
)

func main() {
	// load config
	conf, errconf := config.Load()
	if errconf != nil {
		panic(errconf)
	}

	// Initialize dependency container
	container, err := di.NewContainer(conf)
	if err != nil {
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

	// Initialize routes with dependencies
	var newroutes *routes.Routes
	if err := container.Invoke(func(r *routes.Routes) {
		newroutes = r
	}); err != nil {
		panic(err)
	}
	newroutes.Setup(app)

	app.Get("/", helloWorld)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", conf.AppPort)
		if err := app.Listen(":" + conf.AppPort); err != nil && err != http.ErrServerClosed {
			log.Fatalf("shutting down the server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Print("Server is shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown Fiber server
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Cleanup resources
	cleanup, err := di.GetCleanupManager(container)
	if err != nil {
		log.Printf("Failed to get cleanup manager: %v", err)
	} else {
		if err := cleanup.CleanupWithTimeout(5 * time.Second); err != nil {
			log.Printf("Cleanup failed: %v", err)
		} else {
			log.Print("Resources cleaned up successfully")
		}
	}

	log.Print("Server exited")
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
