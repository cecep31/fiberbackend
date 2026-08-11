package routes

import (
	"github.com/gofiber/fiber/v3"
)

func (r *Routes) setupNotificationRoutes(api fiber.Router) {
	notifications := api.Group("/notifications", r.authMiddleware.Auth())
	{
		notifications.Get("", r.notificationHandler.GetNotifications)
		notifications.Get("/unread-count", r.notificationHandler.GetUnreadCount)
		notifications.Patch("/:id/read", r.notificationHandler.MarkAsRead)
		notifications.Patch("/read-all", r.notificationHandler.MarkAllAsRead)
	}
}
