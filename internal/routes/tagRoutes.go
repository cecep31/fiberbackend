package routes

import "github.com/gofiber/fiber/v3"

func (r *Routes) setupTagRoutes(v1 fiber.Router) {
	tags := v1.Group("/tags")
	{
		tags.Post("", r.authMiddleware.Auth(), r.tagHandler.CreateTag)
		tags.Get("", r.tagHandler.GetTags)
		tags.Get("/:id", r.tagHandler.GetTagByID)
		tags.Put("/:id", r.authMiddleware.Auth(), r.authMiddleware.AuthAdmin(), r.tagHandler.UpdateTag)
		tags.Delete("/:id", r.authMiddleware.Auth(), r.authMiddleware.AuthAdmin(), r.tagHandler.DeleteTag)
	}
}
