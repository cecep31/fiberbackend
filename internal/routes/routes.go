package routes

import (
	"fiberbackend/config"
	"fiberbackend/internal/handler"
	"fiberbackend/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

type Routes struct {
	config                  *config.Config
	userHandler             *handler.UserHandler
	postHandler             *handler.PostHandler
	authHandler             *handler.AuthHandler
	authMiddleware          *middleware.AuthMiddleware
	tagHandler              *handler.TagHandler
	commentHandler          *handler.CommentHandler
	postViewHandler         *handler.PostViewHandler
	postLikeHandler         *handler.PostLikeHandler
	userFollowHandler       *handler.UserFollowHandler
	chatConversationHandler *handler.ChatConversationHandler
}

func NewRoutes(
	config *config.Config,
	userHandler *handler.UserHandler,
	postHandler *handler.PostHandler,
	authHandler *handler.AuthHandler,
	authMiddleware *middleware.AuthMiddleware,
	tagHandler *handler.TagHandler,
	commentHandler *handler.CommentHandler,
	postViewHandler *handler.PostViewHandler,
	postLikeHandler *handler.PostLikeHandler,
	userFollowHandler *handler.UserFollowHandler,
	chatConversationHandler *handler.ChatConversationHandler,
) *Routes {
	return &Routes{
		config:                  config,
		userHandler:             userHandler,
		postHandler:             postHandler,
		authHandler:             authHandler,
		authMiddleware:          authMiddleware,
		tagHandler:              tagHandler,
		commentHandler:          commentHandler,
		postViewHandler:         postViewHandler,
		postLikeHandler:         postLikeHandler,
		userFollowHandler:       userFollowHandler,
		chatConversationHandler: chatConversationHandler,
	}
}

func (r *Routes) Setup(app *fiber.App) {
	// API Group
	v1 := app.Group("/v1")
	r.setupV1Routes(v1)
}

func (r *Routes) setupV1Routes(v1 fiber.Router) {
	r.setupUserRoutes(v1)
	r.setupPostRoutes(v1)
	r.setupAuthRoutes(v1)
	r.setupTagRoutes(v1)
	r.setupChatConversationRoutes(v1)
	if r.config.Debug {
		r.setupDebugRoutes(v1)
	}
}

func (r *Routes) setupChatConversationRoutes(v1 fiber.Router) {
	conversations := v1.Group("/chat/conversations")
	{
		conversations.Post("", r.authMiddleware.Auth(), r.chatConversationHandler.CreateConversation)
		conversations.Get("", r.authMiddleware.Auth(), r.chatConversationHandler.GetConversations)
		conversations.Get("/:id", r.authMiddleware.Auth(), r.chatConversationHandler.GetConversation)
		conversations.Put("/:id", r.authMiddleware.Auth(), r.chatConversationHandler.UpdateConversation)
		conversations.Delete("/:id", r.authMiddleware.Auth(), r.chatConversationHandler.DeleteConversation)
	}
}
