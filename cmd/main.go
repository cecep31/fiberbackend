package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fiberbackend/config"
	"fiberbackend/internal/di"
	"fiberbackend/internal/middleware"
	"fiberbackend/pkg/applog"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

func main() {
	applog.SetupFromEnv()

	conf, errconf := config.Load()
	if errconf != nil {
		slog.Error("failed to load config", "error", errconf)
		panic(errconf)
	}

	applog.Setup(conf.App.Debug)

	// Initialize dependency container
	container, err := di.NewContainer(conf)
	if err != nil {
		slog.Error("failed to initialize container", "error", err)
		panic(err)
	}

	// Body limit, timeouts, enveloped error responses and client-IP
	// resolution (HTTP_TRUST_PROXY) live in middleware.AppConfig.
	app := fiber.New(middleware.AppConfig(conf))

	// Global middleware must be registered before routes so CORS, security
	// headers, rate limiting, and recover apply to all endpoints.
	middleware.InitMiddleware(app, conf)

	// Initialize routes with manually wired dependencies
	container.Routes.Setup(app)

	app.Get("/", helloWorld)

	// Health check endpoint — used by Docker HEALTHCHECK and load balancers.
	// Returns 200 when the DB is reachable, 503 otherwise.
	app.Get("/health", func(c fiber.Ctx) error {
		return healthCheck(c, container)
	})

	// Start server in a goroutine. A listener failure (e.g. port already in
	// use) must stop the process; otherwise it would keep waiting for a signal
	// while serving nothing.
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("starting server", "port", conf.HTTP.Port)
		addr := ":" + conf.HTTP.Port
		if err := app.Listen(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for an interrupt signal or a server failure, then shut down with a
	// timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	exitCode := 0
	select {
	case <-quit:
		slog.Info("server is shutting down")
	case err := <-serverErr:
		slog.Error("server exited unexpectedly", "error", err)
		exitCode = 1
	}

	// End open SSE (realtime) streams first so graceful shutdown does not
	// block waiting on them.
	container.CloseStreams()

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	// Cleanup resources
	cleanup, err := di.GetCleanupManager(container)
	if err != nil {
		slog.Error("failed to get cleanup manager", "error", err)
	} else if cleanup != nil {
		if err := cleanup.CleanupWithTimeout(5 * time.Second); err != nil {
			slog.Error("cleanup failed", "error", err)
		} else {
			slog.Info("resources cleaned up successfully")
		}
	}

	slog.Info("server exited")
	os.Exit(exitCode)
}

func helloWorld(c fiber.Ctx) error {
	return response.Success(c, "Hello, World!", nil)
}

// healthCheck pings the database and returns 200 OK or 503 Service Unavailable.
func healthCheck(c fiber.Ctx, container *di.Container) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	if err := container.PingDB(ctx); err != nil {
		slog.Warn("health check: database unreachable", "error", err)
		return c.Status(http.StatusServiceUnavailable).JSON(map[string]string{
			"status": "unhealthy",
			"reason": "database unreachable",
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]string{
		"status": "ok",
	})
}
