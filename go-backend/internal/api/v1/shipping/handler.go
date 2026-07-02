package shipping

import (
	"tanzanite/internal/service"
)

type Handler struct {
	shippingService *service.ShippingService
}

func NewHandler(shippingService *service.ShippingService) *Handler {
	return &Handler{
		shippingService: shippingService,
	}
}
