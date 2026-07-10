package service

import (
	"tanzanite/internal/repository"
)

type PaymentService struct {
	txManager   *repository.TxManager
	paymentRepo *repository.PaymentRepository
}

func NewPaymentService(txManager *repository.TxManager, paymentRepo *repository.PaymentRepository) *PaymentService {
	return &PaymentService{
		txManager:   txManager,
		paymentRepo: paymentRepo,
	}
}
