package di

import (
	"fiberbackend/config"
	"fiberbackend/internal/handler"
	"fiberbackend/internal/middleware"
	"fiberbackend/internal/repository"
	"fiberbackend/internal/routes"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/database"
	"fiberbackend/pkg/storage"
)

// Container holds all wired application dependencies.
type Container struct {
	Routes  *routes.Routes
	Cleanup *CleanupManager
}

// NewContainer creates and wires all dependencies manually without reflection.
func NewContainer(cfg *config.Config) (*Container, error) {
	// --- Infrastructure ---
	cleanup := NewCleanupManager()

	dbWrapper := database.NewDatabase(cfg)
	cleanup.Register(dbWrapper)
	db := dbWrapper.DB

	s3 := storage.NewS3Storage(cfg)

	// --- Repositories ---
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	authRepo := repository.NewAuthRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	tagRepo := repository.NewTagRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	postViewRepo := repository.NewPostViewRepository(db)
	postLikeRepo := repository.NewPostLikeRepository(db)
	userFollowRepo := repository.NewUserFollowRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	chatConvRepo := repository.NewChatConversationRepository(db)

	// --- Services ---
	tagSvc := service.NewTagService(tagRepo)
	userSvc := service.NewUserService(userRepo)
	authSvc := service.NewAuthService(authRepo, userRepo, sessionRepo, cfg)
	postSvc := service.NewPostService(postRepo, tagSvc, s3)
	commentSvc := service.NewCommentService(commentRepo, postRepo)
	postViewSvc := service.NewPostViewService(postViewRepo, postRepo)
	postLikeSvc := service.NewPostLikeService(postLikeRepo, postRepo)
	userFollowSvc := service.NewUserFollowService(userFollowRepo, userRepo)
	holdingSvc := service.NewHoldingService(holdingRepo)
	chatConvSvc := service.NewChatConversationService(chatConvRepo)

	// --- Handlers ---
	userHandler := handler.NewUserHandler(userSvc, userFollowSvc)
	postHandler := handler.NewPostHandler(postSvc, postViewSvc)
	authHandler := handler.NewAuthHandler(authSvc)
	tagHandler := handler.NewTagHandler(tagSvc)
	commentHandler := handler.NewCommentHandler(commentSvc)
	postViewHandler := handler.NewPostViewHandler(postViewSvc)
	postLikeHandler := handler.NewPostLikeHandler(postLikeSvc)
	userFollowHandler := handler.NewUserFollowHandler(userFollowSvc)
	holdingHandler := handler.NewHoldingHandler(holdingSvc)
	chatConvHandler := handler.NewChatConversationHandler(chatConvSvc)

	// --- Middleware & Routes ---
	authMiddleware := middleware.NewAuthMiddleware(cfg)
	r := routes.NewRoutes(
		cfg,
		userHandler,
		postHandler,
		authHandler,
		authMiddleware,
		tagHandler,
		commentHandler,
		postViewHandler,
		postLikeHandler,
		userFollowHandler,
		chatConvHandler,
		holdingHandler,
	)

	return &Container{
		Routes:  r,
		Cleanup: cleanup,
	}, nil
}
