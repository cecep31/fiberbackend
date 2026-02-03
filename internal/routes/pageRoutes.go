package routes

import "github.com/gofiber/fiber/v3"

func (r *Routes) setupPageRoutes(v1 fiber.Router) {
	pages := v1.Group("/pages", r.authMiddleware.Auth())
	{
		pages.Post("", r.pageHandler.CreatePage)
		pages.Get("/:id", r.pageHandler.GetPage)
		pages.Put("/:id", r.pageHandler.UpdatePage)
		pages.Delete("/:id", r.pageHandler.DeletePage)
		pages.Get("/workspace/:workspace_id", r.pageHandler.GetWorkspacePages)
		pages.Get("/children/:parent_id", r.pageHandler.GetChildPages)
	}
}
