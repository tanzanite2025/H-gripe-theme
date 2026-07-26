package service

import (
	"context"
	"fmt"
	"tanzanite/internal/repository"

	"tanzanite/internal/pkg/antifraud"
)

type PaymentService struct {
	txManager   *repository.TxManager
	paymentRepo *repository.PaymentRepository
	orderRepo   *repository.OrderRepository
	risk        *antifraud.Service
}

func (s *PaymentService) ConfigureRisk(orderRepo *repository.OrderRepository, risk *antifraud.Service) {
	s.orderRepo = orderRepo
	s.risk = risk
}

func (s *PaymentService) RecordGatewayPaymentFailure(ctx context.Context, provider, orderNumber string) error {
	if s.risk == nil || s.orderRepo == nil {
		return nil
	}
	if orderNumber == "" {
		return nil
	}
	o, err := s.orderRepo.FindByOrderNumberForVerification(orderNumber)
	if err != nil {
		return err
	}
	s.risk.RecordProviderFailure(provider)
	return s.risk.RecordFailure(ctx, "user:"+fmt.Sprint(o.UserID))
}

func NewPaymentService(txManager *repository.TxManager, paymentRepo *repository.PaymentRepository) *PaymentService {
	return &PaymentService{
		txManager:   txManager,
		paymentRepo: paymentRepo,
	}
}
