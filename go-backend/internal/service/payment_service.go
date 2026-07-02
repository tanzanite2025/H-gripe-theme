package service

import (
	"errors"
	"fmt"
	"math"
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
	if refund.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if refund.Status == "" {
		refund.Status = "pending"
	}
	refund.RefundedBy = adminUserID

	return s.paymentRepo.CreateRefund(refund)
}

func (s *PaymentService) UpdateRefundStatus(id uint, status string) error {
	refund, err := s.paymentRepo.FindRefundByID(id)
	if err != nil {
		return err
	}

	refund.Status = status
	return s.paymentRepo.UpdateRefund(refund)
}

func (s *PaymentService) CreatePaymentMethod(method *payment.PaymentMethod) error {
	return s.paymentRepo.CreatePaymentMethod(method)
}

func (s *PaymentService) UpdatePaymentMethod(method *payment.PaymentMethod) error {
	return s.paymentRepo.UpdatePaymentMethod(method)
}

func (s *PaymentService) DeletePaymentMethod(id uint) error {
	return s.paymentRepo.DeletePaymentMethod(id)
}

func (s *PaymentService) CreateTaxRate(rate *payment.TaxRate) error {
	return s.paymentRepo.CreateTaxRate(rate)
}

func (s *PaymentService) UpdateTaxRate(rate *payment.TaxRate) error {
	return s.paymentRepo.UpdateTaxRate(rate)
}

func (s *PaymentService) DeleteTaxRate(id uint) error {
	return s.paymentRepo.DeleteTaxRate(id)
}
