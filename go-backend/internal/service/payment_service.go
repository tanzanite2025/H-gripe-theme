package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"tanzanite/internal/domain/payment"
	"tanzanite/internal/pkg/antifraud"
	"tanzanite/internal/repository"
)

type PaymentService struct {
	txManager                      *repository.TxManager
	paymentRepo                    *repository.PaymentRepository
	orderRepo                      *repository.OrderRepository
	shippingRepo                   *repository.ShippingRepository
	ticketRepo                     *repository.TicketRepository
	risk                           *antifraud.Service
	stripeDisputeEvidenceSubmitter stripeDisputeEvidenceSubmitter
}

func (s *PaymentService) ConfigureRisk(orderRepo *repository.OrderRepository, risk *antifraud.Service) {
	s.orderRepo = orderRepo
	s.risk = risk
}

func (s *PaymentService) ConfigureEvidenceSources(orderRepo *repository.OrderRepository, shippingRepo *repository.ShippingRepository, ticketRepo *repository.TicketRepository) {
	if orderRepo != nil {
		s.orderRepo = orderRepo
	}
	s.shippingRepo = shippingRepo
	s.ticketRepo = ticketRepo
}

type GatewayPaymentAttemptInput struct {
	Provider        string
	OrderNumber     string
	TransactionID   string
	PaymentMethod   string
	Status          string
	Amount          float64
	Currency        string
	GatewayResponse string
	ErrorMessage    string
}

func (s *PaymentService) RecordGatewayPaymentFailure(ctx context.Context, provider, orderNumber, transactionID string) error {
	if strings.TrimSpace(orderNumber) == "" {
		return nil
	}
	if strings.TrimSpace(transactionID) != "" {
		if err := s.RecordGatewayPaymentAttempt(GatewayPaymentAttemptInput{
			Provider:      provider,
			OrderNumber:   orderNumber,
			TransactionID: transactionID,
			Status:        "failed",
			ErrorMessage:  "payment intent failed",
		}); err != nil {
			return err
		}
	}

	if s.risk != nil {
		s.risk.RecordProviderFailure(provider)
	}
	return nil
}

func NewPaymentService(txManager *repository.TxManager, paymentRepo *repository.PaymentRepository) *PaymentService {
	return &PaymentService{
		txManager:   txManager,
		paymentRepo: paymentRepo,
	}
}

func (s *PaymentService) RecordGatewayPaymentAttempt(input GatewayPaymentAttemptInput) error {
	input.Provider = strings.TrimSpace(input.Provider)
	input.OrderNumber = strings.TrimSpace(input.OrderNumber)
	input.TransactionID = strings.TrimSpace(input.TransactionID)
	if input.Provider == "" {
		return errors.New("provider is required")
	}
	if input.OrderNumber == "" {
		return errors.New("order_number is required")
	}
	if input.TransactionID == "" {
		return errors.New("transaction_id is required")
	}
	if input.PaymentMethod == "" {
		input.PaymentMethod = input.Provider
	}
	input.Status = normalizeGatewayAttemptStatus(input.Status)
	input.Currency = normalizePaymentCurrency(input.Currency)

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		o, err := repos.Order.FindByOrderNumberForVerification(input.OrderNumber)
		if err != nil {
			return normalizeOrderError(err)
		}
		expectedCurrency, err := orderPaymentCurrency(o)
		if err != nil {
			return err
		}
		if input.Currency == "" {
			input.Currency = expectedCurrency
		} else if input.Currency != expectedCurrency {
			return fmt.Errorf("payment currency %s does not match order currency %s", input.Currency, expectedCurrency)
		}
		if input.Amount > 0 && math.Abs(o.TotalAmount-input.Amount) >= gatewayPaymentAmountTolerance {
			return fmt.Errorf("payment amount %.2f does not match order total %.2f", input.Amount, o.TotalAmount)
		}

		if existing, err := repos.Payment.FindTransactionByTransactionIDForUpdate(input.TransactionID); err == nil {
			if existing.Status == "completed" || existing.Status == "refunded" || existing.Status == "expired" {
				return nil
			}
			existing.OrderID = o.ID
			existing.PaymentMethod = input.PaymentMethod
			if input.Amount > 0 {
				existing.Amount = input.Amount
			} else if existing.Amount <= 0 {
				existing.Amount = o.TotalAmount
			}
			existing.Currency = input.Currency
			existing.Status = input.Status
			existing.GatewayResponse = input.GatewayResponse
			existing.ErrorMessage = input.ErrorMessage
			return repos.Payment.UpdateTransaction(existing)
		} else if !repository.IsRecordNotFound(err) {
			return err
		}

		amount := input.Amount
		if amount <= 0 {
			amount = o.TotalAmount
		}
		return repos.Payment.CreateTransaction(&payment.Transaction{
			OrderID:         o.ID,
			TransactionID:   input.TransactionID,
			PaymentMethod:   input.PaymentMethod,
			Amount:          amount,
			Currency:        input.Currency,
			Status:          input.Status,
			GatewayResponse: input.GatewayResponse,
			ErrorMessage:    input.ErrorMessage,
		})
	})
}
