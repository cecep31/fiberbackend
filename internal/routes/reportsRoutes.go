package routes

import (
	"github.com/gofiber/fiber/v3"
)

func (r *Routes) setupReportRoutes(api fiber.Router) {
	reports := api.Group("/reports", r.authMiddleware.Auth(), r.authMiddleware.AuthAdmin())
	{
		reports.Get("/overview", r.reportHandler.GetOverview)
		reports.Get("/users", r.reportHandler.GetUsers)
		reports.Get("/posts", r.reportHandler.GetPosts)
		reports.Get("/engagement", r.reportHandler.GetEngagement)
	}
}
