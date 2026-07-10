package service

import (
	"errors"
	"fmt"
	"math"
	"tanzanite/internal/domain/payment"
	"tanzanite/internal/repository"
	"time"
)

type VerifiedGatewayPaymentInput struct {
	Provider        string
	OrderNumber     string
	TransactionID   string
	PaymentMethod   string
	Amount          float64
	Currency        string
	GatewayResponse string
}

func (s *PaymentService) GetTransaction(id uint) (*payment.Transaction, error) {
	return s.paymentRepo.FindTransactionByID(id)
}

func (s *PaymentService) GetOrderTransactions(orderID uint) ([]payment.Transaction, error) {
	return s.paymentRepo.FindTransactionByOrderID(orderID)
}

func (s *PaymentService) RecordVerifiedGatewayPayment(input VerifiedGatewayPaymentInput) error {
	if input.Provider == "" {
		return errors.New("provider is required")
	}
	if input.OrderNumber == "" {
		return errors.New("order_number is required")
	}
	if input.TransactionID == "" {
		return errors.New("transaction_id is required")
	}
	if input.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if input.PaymentMethod == "" {
		input.PaymentMethod = input.Provider
	}
	if input.Currency == "" {
		input.Currency = "USD"
	}

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if _, err := repos.Payment.FindTransactionByTransactionID(input.TransactionID); err == nil {
			return nil
		} else if !repository.IsRecordNotFound(err) {
			return err
		}

		o, err := repos.Order.FindByOrderNumberForVerification(input.OrderNumber)
		if err != nil {
			return normalizeOrderError(err)
		}
		if o.PaymentStatus == "paid" {
			return errors.New("order is already paid")
		}
		if o.Status == "cancelled" || o.Status == "refunded" {
			return fmt.Errorf("cannot mark %s order as paid", o.Status)
		}
		if math.Abs(o.TotalAmount-input.Amount) > 0.01 {
			return fmt.Errorf("payment amount %.2f does not match order total %.2f", input.Amount, o.TotalAmount)
		}

		completedAt := time.Now()
		transaction := &payment.Transaction{
			OrderID:         o.ID,
			TransactionID:   input.TransactionID,
			PaymentMethod:   input.PaymentMethod,
			Amount:          input.Amount,
			Currency:        input.Currency,
			Status:          "completed",
			GatewayResponse: input.GatewayResponse,
			CompletedAt:     &completedAt,
		}
		if err := repos.Payment.CreateTransaction(transaction); err != nil {
			return err
		}
		if err := repos.Order.UpdatePaymentStatus(o.ID, "paid"); err != nil {
			return err
		}
		if o.Status == "pending" || o.Status == "paid" {
			return repos.Order.UpdateStatus(o.ID, "processing")
		}
		return nil
	})
}
