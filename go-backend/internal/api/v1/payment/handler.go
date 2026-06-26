package payment

import (
	"tanzanite/internal/repository"
)

type Handler struct {
	paymentRepo *repository.PaymentRepository
	orderRepo   *repository.OrderRepository
}

func NewHandler(paymentRepo *repository.PaymentRepository, orderRepo *repository.OrderRepository) *Handler {
	return &Handler{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
	}
}
