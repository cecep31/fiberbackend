package routes

import "github.com/gofiber/fiber/v3"

func (r *Routes) setupHoldingsRoutes(v1 fiber.Router) {
	holdings := v1.Group("/holdings")
	holdings.Use(r.authMiddleware.Auth())
	{
		holdings.Post("", r.holdingHandler.Create)
		holdings.Get("", r.holdingHandler.List)
		holdings.Get("/:id", r.holdingHandler.GetByID)
		holdings.Put("/:id", r.holdingHandler.Update)
		holdings.Delete("/:id", r.holdingHandler.Delete)
	}
}
