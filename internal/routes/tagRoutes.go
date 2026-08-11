package routes

import (
	"github.com/gofiber/fiber/v3"
)

func (r *Routes) setupTagRoutes(api fiber.Router) {
	tags := api.Group("/tags")
	{
		tags.Post("", r.authMiddleware.Auth(), r.tagHandler.CreateTag)
		tags.Get("", r.tagHandler.GetTags)
		tags.Get("/trending", r.tagHandler.GetTrendingTags)
		tags.Get("/sitemap", r.tagHandler.GetTagsForSitemap)
		tags.Get("/:id", r.tagHandler.GetTagByID)
		tags.Put("/:id", r.authMiddleware.Auth(), r.authMiddleware.AuthAdmin(), r.tagHandler.UpdateTag)
		tags.Delete("/:id", r.authMiddleware.Auth(), r.authMiddleware.AuthAdmin(), r.tagHandler.DeleteTag)
	}
}
