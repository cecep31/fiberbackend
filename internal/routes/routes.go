package routes

import (
	"fiberbackend/config"
	"fiberbackend/internal/handler"
	"fiberbackend/internal/middleware"
	"fiberbackend/internal/platform/cache"

	"github.com/gofiber/fiber/v3"
)

type Routes struct {
	config                  *config.Config
	cache                   *cache.RedisCache
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
	holdingHandler          *handler.HoldingHandler
	exchangeRateHandler     *handler.ExchangeRateHandler
	bookmarkHandler         *handler.BookmarkHandler
	notificationHandler     *handler.NotificationHandler
	reportHandler           *handler.ReportHandler
	corporateActionHandler  *handler.CorporateActionHandler
}

func NewRoutes(
	config *config.Config,
	redisCache *cache.RedisCache,
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
	holdingHandler *handler.HoldingHandler,
	exchangeRateHandler *handler.ExchangeRateHandler,
	bookmarkHandler *handler.BookmarkHandler,
	notificationHandler *handler.NotificationHandler,
	reportHandler *handler.ReportHandler,
	corporateActionHandler *handler.CorporateActionHandler,
) *Routes {
	return &Routes{
		config:                  config,
		cache:                   redisCache,
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
		holdingHandler:          holdingHandler,
		exchangeRateHandler:     exchangeRateHandler,
		bookmarkHandler:         bookmarkHandler,
		notificationHandler:     notificationHandler,
		reportHandler:           reportHandler,
		corporateActionHandler:  corporateActionHandler,
	}
}

func (r *Routes) Setup(app *fiber.App) {
	// API Group
	api := app.Group("/api")
	r.setupAPIRoutes(api)
}

func (r *Routes) setupAPIRoutes(api fiber.Router) {
	r.setupUserRoutes(api)
	r.setupPostRoutes(api)
	r.setupAuthRoutes(api)
	r.setupTagRoutes(api)
	r.setupChatConversationRoutes(api)
	r.setupHoldingRoutes(api)
	r.setupExchangeRateRoutes(api)
	r.setupBookmarkRoutes(api)
	r.setupNotificationRoutes(api)
	r.setupReportRoutes(api)
}

func (r *Routes) setupChatConversationRoutes(api fiber.Router) {
	conversations := api.Group("/chat/conversations")
	{
		conversations.Post("", r.authMiddleware.Auth(), r.chatConversationHandler.CreateConversation)
		conversations.Post("/stream", r.authMiddleware.Auth(), r.chatConversationHandler.CreateConversationStream)
		conversations.Get("", r.authMiddleware.Auth(), r.chatConversationHandler.GetConversations)
		conversations.Get("/:id", r.authMiddleware.Auth(), r.chatConversationHandler.GetConversation)
		conversations.Put("/:id", r.authMiddleware.Auth(), r.chatConversationHandler.UpdateConversation)
		conversations.Delete("/:id", r.authMiddleware.Auth(), r.chatConversationHandler.DeleteConversation)
		conversations.Post("/:conversationId/messages", r.authMiddleware.Auth(), r.chatConversationHandler.CreateMessage)
		conversations.Post("/:conversationId/messages/stream", r.authMiddleware.Auth(), r.chatConversationHandler.CreateMessageStream)
		conversations.Get("/:conversationId/messages", r.authMiddleware.Auth(), r.chatConversationHandler.GetMessages)
	}

	messages := api.Group("/chat/messages")
	{
		messages.Get("/:messageId", r.authMiddleware.Auth(), r.chatConversationHandler.GetMessage)
		messages.Delete("/:messageId", r.authMiddleware.Auth(), r.chatConversationHandler.DeleteMessage)
	}
}
