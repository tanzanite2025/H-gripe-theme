package service

import (
	"errors"
	"tanzanite/internal/domain/payment"
	"tanzanite/internal/repository"
)

type PaymentService struct {
	paymentRepo  *repository.PaymentRepository
	orderService *OrderService
}

func NewPaymentService(paymentRepo *repository.PaymentRepository, orderService *OrderService) *PaymentService {
	return &PaymentService{
		paymentRepo:  paymentRepo,
		orderService: orderService,
	}
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

func (s *PaymentService) CreateGatewayTransaction(transaction *payment.Transaction) error {
	if transaction == nil {
		return errors.New("transaction is required")
	}
	if transaction.OrderID == 0 {
		return errors.New("order_id is required")
	}
	if transaction.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if transaction.PaymentMethod == "" {
		return errors.New("payment_method is required")
	}
	if transaction.Status == "" {
		transaction.Status = "pending"
	}

	return s.paymentRepo.CreateTransaction(transaction)
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
