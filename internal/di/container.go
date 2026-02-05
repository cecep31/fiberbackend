package di

import (
	"fmt"

	"fiberbackend/config"
	"fiberbackend/internal/handler"
	"fiberbackend/internal/middleware"
	"fiberbackend/internal/repository"
	"fiberbackend/internal/routes"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/database"
	"fiberbackend/pkg/storage"

	"go.uber.org/dig"
	"gorm.io/gorm"
)

// Container holds the dependency injection container.
// It's recommended to use a struct to hold the container to avoid global variables.
type Container struct {
	*dig.Container
}

// NewContainer creates a new dependency injection container and registers all the dependencies.
func NewContainer(cfg *config.Config) (*Container, error) {
	container := &Container{dig.New()}

	if err := container.Provide(func() *config.Config { return cfg }); err != nil {
		return nil, fmt.Errorf("failed to provide config: %w", err)
	}

	if err := container.registerDatabase(); err != nil {
		return nil, fmt.Errorf("failed to register database: %w", err)
	}

	if err := container.registerRepositories(); err != nil {
		return nil, fmt.Errorf("failed to register repositories: %w", err)
	}

	if err := container.registerServices(); err != nil {
		return nil, fmt.Errorf("failed to register services: %w", err)
	}

	if err := container.registerHandlers(); err != nil {
		return nil, fmt.Errorf("failed to register handlers: %w", err)
	}

	if err := container.registerRoutes(); err != nil {
		return nil, fmt.Errorf("failed to register routes: %w", err)
	}

	// Provide cleanup manager
	if err := container.Provide(NewCleanupManager); err != nil {
		return nil, fmt.Errorf("failed to provide cleanup manager: %w", err)
	}

	// Provide storage
	if err := container.Provide(storage.NewS3Storage); err != nil {
		return nil, fmt.Errorf("failed to provide storage: %w", err)
	}

	return container, nil
}

func (c *Container) registerDatabase() error {
	if err := c.Provide(func(config *config.Config, cleanup *CleanupManager) *database.DatabaseWrapper {
		db := database.NewDatabase(config)
		cleanup.Register(db)
		return db
	}); err != nil {
		return fmt.Errorf("failed to provide database wrapper: %w", err)
	}

	if err := c.Provide(func(wrapper *database.DatabaseWrapper) *gorm.DB {
		return wrapper.DB
	}); err != nil {
		return fmt.Errorf("failed to provide gorm db: %w", err)
	}
	return nil
}

func (c *Container) registerRepositories() error {
	if err := c.Provide(repository.NewUserRepository); err != nil {
		return fmt.Errorf("failed to provide user repository: %w", err)
	}
	if err := c.Provide(repository.NewPostRepository); err != nil {
		return fmt.Errorf("failed to provide post repository: %w", err)
	}
	if err := c.Provide(repository.NewAuthRepository); err != nil {
		return fmt.Errorf("failed to provide auth repository: %w", err)
	}
	if err := c.Provide(repository.NewSessionRepository); err != nil {
		return fmt.Errorf("failed to provide session repository: %w", err)
	}
	if err := c.Provide(repository.NewTagRepository); err != nil {
		return fmt.Errorf("failed to provide tag repository: %w", err)
	}
	if err := c.Provide(repository.NewCommentRepository); err != nil {
		return fmt.Errorf("failed to provide comment repository: %w", err)
	}
	if err := c.Provide(repository.NewPostViewRepository); err != nil {
		return fmt.Errorf("failed to provide post view repository: %w", err)
	}
	if err := c.Provide(repository.NewPostLikeRepository); err != nil {
		return fmt.Errorf("failed to provide post like repository: %w", err)
	}
	if err := c.Provide(repository.NewUserFollowRepository); err != nil {
		return fmt.Errorf("failed to provide user follow repository: %w", err)
	}
	if err := c.Provide(repository.NewHoldingRepository); err != nil {
		return fmt.Errorf("failed to provide holding repository: %w", err)
	}
	if err := c.Provide(repository.NewChatConversationRepository); err != nil {
		return fmt.Errorf("failed to provide chat conversation repository: %w", err)
	}
	return nil
}

func (c *Container) registerServices() error {
	if err := c.Provide(service.NewUserService); err != nil {
		return fmt.Errorf("failed to provide user service: %w", err)
	}
	if err := c.Provide(service.NewPostService); err != nil {
		return fmt.Errorf("failed to provide post service: %w", err)
	}
	if err := c.Provide(service.NewAuthService); err != nil {
		return fmt.Errorf("failed to provide auth service: %w", err)
	}
	if err := c.Provide(service.NewTagService); err != nil {
		return fmt.Errorf("failed to provide tag service: %w", err)
	}
	if err := c.Provide(service.NewCommentService); err != nil {
		return fmt.Errorf("failed to provide comment service: %w", err)
	}
	if err := c.Provide(service.NewPostViewService); err != nil {
		return fmt.Errorf("failed to provide post view service: %w", err)
	}
	if err := c.Provide(service.NewPostLikeService); err != nil {
		return fmt.Errorf("failed to provide post like service: %w", err)
	}
	if err := c.Provide(service.NewUserFollowService); err != nil {
		return fmt.Errorf("failed to provide user follow service: %w", err)
	}
	if err := c.Provide(service.NewHoldingService); err != nil {
		return fmt.Errorf("failed to provide holding service: %w", err)
	}
	if err := c.Provide(service.NewChatConversationService); err != nil {
		return fmt.Errorf("failed to provide chat conversation service: %w", err)
	}
	return nil
}

func (c *Container) registerHandlers() error {
	if err := c.Provide(handler.NewUserHandler); err != nil {
		return fmt.Errorf("failed to provide user handler: %w", err)
	}
	if err := c.Provide(handler.NewPostHandler); err != nil {
		return fmt.Errorf("failed to provide post handler: %w", err)
	}
	if err := c.Provide(handler.NewAuthHandler); err != nil {
		return fmt.Errorf("failed to provide auth handler: %w", err)
	}
	if err := c.Provide(handler.NewTagHandler); err != nil {
		return fmt.Errorf("failed to provide tag handler: %w", err)
	}
	if err := c.Provide(handler.NewCommentHandler); err != nil {
		return fmt.Errorf("failed to provide comment handler: %w", err)
	}
	if err := c.Provide(handler.NewPostViewHandler); err != nil {
		return fmt.Errorf("failed to provide post view handler: %w", err)
	}
	if err := c.Provide(handler.NewPostLikeHandler); err != nil {
		return fmt.Errorf("failed to provide post like handler: %w", err)
	}
	if err := c.Provide(handler.NewUserFollowHandler); err != nil {
		return fmt.Errorf("failed to provide user follow handler: %w", err)
	}
	if err := c.Provide(handler.NewHoldingHandler); err != nil {
		return fmt.Errorf("failed to provide holding handler: %w", err)
	}
	if err := c.Provide(handler.NewChatConversationHandler); err != nil {
		return fmt.Errorf("failed to provide chat conversation handler: %w", err)
	}
	return nil
}

func (c *Container) registerRoutes() error {
	if err := c.Provide(middleware.NewAuthMiddleware); err != nil {
		return fmt.Errorf("failed to provide auth middleware: %w", err)
	}
	if err := c.Provide(routes.NewRoutes); err != nil {
		return fmt.Errorf("failed to provide routes: %w", err)
	}
	return nil
}

// GetCleanupManager retrieves the cleanup manager from the container.
// This function is kept for convenience, but it's recommended to pass the container
// instance around instead of using this function.
func GetCleanupManager(container *Container) (*CleanupManager, error) {
	var cleanup *CleanupManager
	if err := container.Invoke(func(c *CleanupManager) {
		cleanup = c
	}); err != nil {
		return nil, fmt.Errorf("failed to get cleanup manager: %w", err)
	}
	return cleanup, nil
}
