package shipping

import (
	"tanzanite/internal/repository"
)

type Handler struct {
	shippingRepo *repository.ShippingRepository
}

func NewHandler(shippingRepo *repository.ShippingRepository) *Handler {
	return &Handler{
		shippingRepo: shippingRepo,
	}
}
