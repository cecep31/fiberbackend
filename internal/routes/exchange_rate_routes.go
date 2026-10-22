package routes

import (
	"github.com/gofiber/fiber/v3"
)

func (r *Routes) setupExchangeRateRoutes(api fiber.Router) {
	exchangeRates := api.Group("/exchange-rates", r.authMiddleware.Auth())
	{
		exchangeRates.Get("", r.exchangeRateHandler.GetRate)
	}
}
