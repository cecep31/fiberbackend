package di

import (
	"context"
	"fiberbackend/config"
	"fiberbackend/internal/handler"
	"fiberbackend/internal/middleware"
	"fiberbackend/internal/platform/cache"
	"fiberbackend/internal/platform/database"
	"fiberbackend/internal/platform/email"
	"fiberbackend/internal/platform/market"
	"fiberbackend/internal/platform/openrouter"
	"fiberbackend/internal/platform/queue"
	"fiberbackend/internal/platform/realtime"
	"fiberbackend/internal/platform/storage"
	"fiberbackend/internal/repository"
	"fiberbackend/internal/routes"
	"fiberbackend/internal/service"

	"gorm.io/gorm"
)

// Container holds the manually wired application dependencies.
type Container struct {
	Config  *config.Config
	Cleanup *CleanupManager
	Routes  *routes.Routes
	db      *gorm.DB
	hub     *realtime.Hub
}

// NewContainer creates a manually wired application container.
func NewContainer(cfg *config.Config) (*Container, error) {
	cleanup := NewCleanupManager()

	db, err := database.NewDatabase(cfg)
	if err != nil {
		return nil, err
	}
	cleanup.Register(func() error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	})

	redisCache := cache.NewRedisCache(cfg)
	if redisCache != nil {
		cleanup.Register(func() error {
			return redisCache.Close()
		})
	}

	// Realtime fan-out for guild chat SSE streams; relays through Redis pub/sub
	// when a cache is configured and keeps in-process subscribers otherwise.
	hub := realtime.NewHub(redisCache)
	hub.Start()
	cleanup.Register(hub.Close)

	taskQueue := queue.NewService(cfg.Queue)
	cleanup.Register(func() error {
		return taskQueue.Close()
	})

	s3Storage := storage.NewS3Storage(cfg)
	emailService := email.NewService(cfg.Email, taskQueue)
	cleanup.Register(func() error {
		return emailService.Close()
	})
	taskQueue.Start()

	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	authRepo := repository.NewAuthRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	tagRepo := repository.NewTagRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	postViewRepo := repository.NewPostViewRepository(db)
	postLikeRepo := repository.NewPostLikeRepository(db)
	userFollowRepo := repository.NewUserFollowRepository(db)
	chatConversationRepo := repository.NewChatConversationRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	bookmarkRepo := repository.NewBookmarkRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	authActivityLogRepo := repository.NewAuthActivityLogRepository(db)
	passwordResetTokenRepo := repository.NewPasswordResetTokenRepository(db)
	corporateActionRepo := repository.NewCorporateActionRepository(db)
	postReportRepo := repository.NewPostReportRepository(db)
	postStatsRepo := repository.NewPostStatsRepository(db)
	tagStatsRepo := repository.NewTagStatsRepository(db)
	userStatsRepo := repository.NewUserStatsRepository(db)
	guildRepo := repository.NewGuildRepository(db)
	guildChannelRepo := repository.NewGuildChannelRepository(db)

	authActivityService := service.NewAuthActivityService(authActivityLogRepo)
	openRouterClient := openrouter.NewClient(cfg.OpenRouter)
	userService := service.NewUserService(userRepo, s3Storage)
	tagService := service.NewTagService(tagRepo, redisCache)
	postService := service.NewPostService(postRepo, tagService, s3Storage, redisCache)
	authService := service.NewAuthService(authRepo, userRepo, sessionRepo, passwordResetTokenRepo, authActivityService, cfg, redisCache, emailService)
	notificationService := service.NewNotificationService(notificationRepo)
	commentService := service.NewCommentService(commentRepo, postRepo, notificationService)
	postViewService := service.NewPostViewService(postViewRepo, postRepo, postLikeRepo, redisCache)
	postLikeService := service.NewPostLikeService(postLikeRepo, postRepo)
	userFollowService := service.NewUserFollowService(userFollowRepo, userRepo, notificationService)
	chatConversationService := service.NewChatConversationService(chatConversationRepo, openRouterClient, cfg)
	rawQuoteClient := market.NewRapidAPIQuoteClient(cfg.MarketData.RapidAPIKey, nil)
	quoteClient := market.NewCachedQuoteClient(rawQuoteClient, redisCache)
	holdingService := service.NewHoldingService(holdingRepo, quoteClient, redisCache)
	exchangeRateService := service.NewExchangeRateService(quoteClient, redisCache)
	bookmarkService := service.NewBookmarkService(bookmarkRepo, postRepo)
	postReportService := service.NewPostReportService(postReportRepo, postRepo)
	postStatsService := service.NewPostStatsService(postStatsRepo)
	tagStatsService := service.NewTagStatsService(tagStatsRepo)
	userStatsService := service.NewUserStatsService(userStatsRepo)
	guildService := service.NewGuildService(guildRepo)
	guildChannelService := service.NewGuildChannelService(guildChannelRepo, hub)

	// Corporate actions: IDX
	idxCorporateClient := market.NewRapidAPIIDXClient(cfg.MarketData.RapidAPIKey, nil)
	corporateActionService := service.NewCorporateActionService(idxCorporateClient, corporateActionRepo)

	userHandler := handler.NewUserHandler(userService, userFollowService)
	postHandler := handler.NewPostHandler(postService, postViewService)
	authHandler := handler.NewAuthHandler(authService, authActivityService, cfg.Frontend)
	tagHandler := handler.NewTagHandler(tagService)
	commentHandler := handler.NewCommentHandler(commentService, userService)
	postViewHandler := handler.NewPostViewHandler(postViewService, userService)
	postLikeHandler := handler.NewPostLikeHandler(postLikeService)
	userFollowHandler := handler.NewUserFollowHandler(userFollowService)
	chatConversationHandler := handler.NewChatConversationHandler(chatConversationService)
	holdingHandler := handler.NewHoldingHandler(holdingService)
	exchangeRateHandler := handler.NewExchangeRateHandler(exchangeRateService)
	bookmarkHandler := handler.NewBookmarkHandler(bookmarkService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	corporateActionHandler := handler.NewCorporateActionHandler(corporateActionService)
	postReportHandler := handler.NewPostReportHandler(postReportService)
	postStatsHandler := handler.NewPostStatsHandler(postStatsService)
	tagStatsHandler := handler.NewTagStatsHandler(tagStatsService)
	userStatsHandler := handler.NewUserStatsHandler(userStatsService)
	guildHandler := handler.NewGuildHandler(guildService)
	guildChannelHandler := handler.NewGuildChannelHandler(guildChannelService)

	authMiddleware := middleware.NewAuthMiddleware(cfg, userService)
	appRoutes := routes.NewRoutes(
		cfg,
		redisCache,
		userHandler,
		postHandler,
		authHandler,
		authMiddleware,
		tagHandler,
		commentHandler,
		postViewHandler,
		postLikeHandler,
		userFollowHandler,
		chatConversationHandler,
		holdingHandler,
		exchangeRateHandler,
		bookmarkHandler,
		notificationHandler,
		corporateActionHandler,
		guildHandler,
		guildChannelHandler,
		postReportHandler,
		postStatsHandler,
		tagStatsHandler,
		userStatsHandler,
	)

	return &Container{
		Config:  cfg,
		Cleanup: cleanup,
		Routes:  appRoutes,
		db:      db,
		hub:     hub,
	}, nil
}

// PingDB checks that the database connection is alive.
func (c *Container) PingDB(ctx context.Context) error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// GetCleanupManager retrieves the cleanup manager from the container.
func GetCleanupManager(container *Container) (*CleanupManager, error) {
	if container == nil {
		return nil, nil
	}
	return container.Cleanup, nil
}

// CloseStreams ends every open realtime (SSE) stream. Register it with the
// server shutdown hook so clients disconnect cleanly.
func (c *Container) CloseStreams() {
	if c == nil || c.hub == nil {
		return
	}
	_ = c.hub.Close()
}
