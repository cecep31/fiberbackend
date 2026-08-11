package handler

import (
	"errors"

	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type ExchangeRateHandler struct {
	exchangeRateService service.ExchangeRateService
}

func NewExchangeRateHandler(exchangeRateService service.ExchangeRateService) *ExchangeRateHandler {
	return &ExchangeRateHandler{exchangeRateService: exchangeRateService}
}

func (h *ExchangeRateHandler) GetRate(c fiber.Ctx) error {
	from := c.Query("from")
	to := c.Query("to")

	result, err := h.exchangeRateService.GetRate(c.Context(), from, to)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCurrencyPair) {
			return response.BadRequest(c, "Invalid currency pair", err)
		}
		return response.InternalServerError(c, "Failed to get exchange rate", err)
	}

	return response.Success(c, "Exchange rate fetched successfully", result)
}
