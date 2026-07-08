package service

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"tanzanite/internal/domain/payment"
	"tanzanite/internal/repository"
	"time"
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

type VerifiedGatewayPaymentInput struct {
	Provider        string
	OrderNumber     string
	TransactionID   string
	PaymentMethod   string
	Amount          float64
	Currency        string
	GatewayResponse string
}

type VerifiedGatewayRefundInput struct {
	Provider        string
	OrderNumber     string
	TransactionID   string
	RefundID        string
	Amount          float64
	Currency        string
	GatewayResponse string
}

func (s *PaymentService) ListPaymentMethods(enabledOnly bool) ([]payment.PaymentMethod, error) {
	return s.paymentRepo.FindAllPaymentMethods(enabledOnly)
}

func (s *PaymentService) GetPaymentMethod(id uint) (*payment.PaymentMethod, error) {
	return s.paymentRepo.FindPaymentMethodByID(id)
}

func (s *PaymentService) ListTaxRates() ([]payment.TaxRate, error) {
	return s.paymentRepo.FindAllTaxRates()
}

func (s *PaymentService) GetTaxRate(id uint) (*payment.TaxRate, error) {
	return s.paymentRepo.FindTaxRateByID(id)
}

func (s *PaymentService) CalculateTax(amount float64, country, state string) (float64, float64, error) {
	taxRate, err := s.paymentRepo.FindTaxRateByLocation(country, state)
	if err != nil {
		return 0, 0, nil
	}

	tax := amount * taxRate.Rate / 100
	return taxRate.Rate, tax, nil
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

func (s *PaymentService) GetRefund(id uint) (*payment.Refund, error) {
	return s.paymentRepo.FindRefundByID(id)
}

func (s *PaymentService) GetOrderRefunds(orderID uint) ([]payment.Refund, error) {
	return s.paymentRepo.FindRefundsByOrderID(orderID)
}

func (s *PaymentService) CreateAdminRefund(refund *payment.Refund, adminUserID uint) error {
	if refund == nil {
		return errors.New("refund is required")
	}
	if refund.OrderID == 0 {
		return errors.New("order_id is required")
	}
	if refund.TransactionID == 0 {
		return errors.New("transaction_id is required")
	}
	if refund.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		transaction, err := repos.Payment.FindTransactionByIDForUpdate(refund.TransactionID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return errors.New("transaction not found")
			}
			return err
		}
		if transaction.OrderID != refund.OrderID {
			return errors.New("transaction does not belong to order")
		}
		if transaction.Status != "completed" {
			return errors.New("transaction is not refundable")
		}

		o, err := repos.Order.FindByIDForUpdate(refund.OrderID)
		if err != nil {
			return normalizeOrderError(err)
		}
		if o.PaymentStatus != "paid" {
			return errors.New("order is not paid")
		}
		if o.Status == "refunded" {
			return errors.New("order is already refunded")
		}

		reservedAmount, err := repos.Payment.SumRefundAmountByTransactionID(transaction.ID, "pending", "completed")
		if err != nil {
			return err
		}
		if refund.Amount-(transaction.Amount-reservedAmount) > 0.01 {
			return fmt.Errorf("refund amount %.2f exceeds refundable amount %.2f", refund.Amount, transaction.Amount-reservedAmount)
		}

		refund.Status = "pending"
		refund.RefundID = nil
		refund.GatewayResponse = ""
		refund.CompletedAt = nil
		refund.RefundedBy = adminUserID

		return repos.Payment.CreateRefund(refund)
	})
}

func (s *PaymentService) RecordVerifiedGatewayRefund(input VerifiedGatewayRefundInput) error {
	if input.Provider == "" {
		return errors.New("provider is required")
	}
	if input.TransactionID == "" {
		return errors.New("transaction_id is required")
	}
	if input.RefundID == "" {
		return errors.New("refund_id is required")
	}
	if input.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if existing, err := repos.Payment.FindRefundByRefundID(input.RefundID); err == nil {
			if existing.Status == "completed" {
				return nil
			}
			return errors.New("refund id is already used")
		} else if !repository.IsRecordNotFound(err) {
			return err
		}

		transaction, err := repos.Payment.FindTransactionByTransactionIDForUpdate(input.TransactionID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return errors.New("transaction not found")
			}
			return err
		}
		if transaction.Status == "refunded" {
			return errors.New("transaction is already fully refunded")
		}
		if transaction.Status != "completed" {
			return errors.New("transaction is not refundable")
		}
		if input.Currency != "" && transaction.Currency != "" && !strings.EqualFold(input.Currency, transaction.Currency) {
			return fmt.Errorf("refund currency %s does not match transaction currency %s", input.Currency, transaction.Currency)
		}

		o, err := repos.Order.FindByIDForUpdate(transaction.OrderID)
		if err != nil {
			return normalizeOrderError(err)
		}
		if input.OrderNumber != "" && o.OrderNumber != input.OrderNumber {
			return errors.New("refund order_number does not match transaction order")
		}
		if o.PaymentStatus == "refunded" {
			return errors.New("order is already refunded")
		}
		if o.PaymentStatus != "paid" {
			return errors.New("order is not paid")
		}

		completedAmount, err := repos.Payment.SumRefundAmountByTransactionID(transaction.ID, "completed")
		if err != nil {
			return err
		}
		if input.Amount-(transaction.Amount-completedAmount) > 0.01 {
			return fmt.Errorf("refund amount %.2f exceeds refundable amount %.2f", input.Amount, transaction.Amount-completedAmount)
		}

		now := time.Now()
		refundID := input.RefundID
		pendingRefund, err := repos.Payment.FindPendingRefundByTransactionAndAmount(transaction.ID, input.Amount)
		if err != nil && !repository.IsRecordNotFound(err) {
			return err
		}
		if pendingRefund != nil && err == nil {
			pendingRefund.Status = "completed"
			pendingRefund.RefundID = &refundID
			pendingRefund.GatewayResponse = input.GatewayResponse
			pendingRefund.CompletedAt = &now
			if err := repos.Payment.UpdateRefund(pendingRefund); err != nil {
				return err
			}
		} else {
			refund := &payment.Refund{
				OrderID:         transaction.OrderID,
				TransactionID:   transaction.ID,
				RefundID:        &refundID,
				Amount:          input.Amount,
				Status:          "completed",
				GatewayResponse: input.GatewayResponse,
				CompletedAt:     &now,
			}
			if err := repos.Payment.CreateRefund(refund); err != nil {
				return err
			}
		}

		if completedAmount+input.Amount >= transaction.Amount-0.01 {
			transaction.Status = "refunded"
			if err := repos.Payment.UpdateTransaction(transaction); err != nil {
				return err
			}
		}

		orderRefundedAmount, err := repos.Payment.SumRefundAmountByOrderID(o.ID, "completed")
		if err != nil {
			return err
		}
		if orderRefundedAmount >= o.TotalAmount-0.01 {
			if err := repos.Order.UpdatePaymentStatus(o.ID, "refunded"); err != nil {
				return err
			}
			return repos.Order.UpdateStatus(o.ID, "refunded")
		}

		return nil
	})
}
