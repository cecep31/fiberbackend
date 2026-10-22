package routes

import (
	"github.com/gofiber/fiber/v3"
)

func (r *Routes) setupHoldingRoutes(api fiber.Router) {
	holdings := api.Group("/holdings", r.authMiddleware.Auth())
	{
		holdings.Get("", r.holdingHandler.GetHoldings)
		holdings.Get("/summary", r.holdingHandler.GetSummary)
		holdings.Get("/trends", r.holdingHandler.GetTrends)
		holdings.Get("/compare", r.holdingHandler.CompareMonths)
		holdings.Get("/monthly", r.holdingHandler.GetMonthlyData)
		holdings.Get("/calendar", r.corporateActionHandler.GetCalendar)
		holdings.Post("", r.holdingHandler.CreateHolding)
		holdings.Post("/duplicate", r.holdingHandler.DuplicateHoldings)
		holdings.Post("/sync", r.holdingHandler.SyncPrices)
		holdings.Get("/:id", r.holdingHandler.GetHoldingByID)
		holdings.Put("/:id", r.holdingHandler.UpdateHolding)
		holdings.Delete("/:id", r.holdingHandler.DeleteHolding)
	}

	holdingTypes := api.Group("/holding-types", r.authMiddleware.Auth())
	{
		holdingTypes.Get("", r.holdingHandler.GetHoldingTypes)
	}
}
