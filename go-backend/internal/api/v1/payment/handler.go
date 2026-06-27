package payment

import (
	"tanzanite/internal/repository"
	"tanzanite/internal/service"
)

type Handler struct {
	paymentRepo  *repository.PaymentRepository
	orderService *service.OrderService
}

func NewHandler(paymentRepo *repository.PaymentRepository, orderService *service.OrderService) *Handler {
	return &Handler{
		paymentRepo:  paymentRepo,
		orderService: orderService,
	}
}
