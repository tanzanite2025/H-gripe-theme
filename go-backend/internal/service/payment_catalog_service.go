package service

import (
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/payment"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func (s *PaymentService) ListPaymentMethods(enabledOnly bool) ([]payment.PaymentMethod, error) {
	return s.paymentRepo.FindAllPaymentMethods(enabledOnly)
}

func (s *PaymentService) GetPaymentMethod(id uint) (*payment.PaymentMethod, error) {
	return s.paymentRepo.FindPaymentMethodByID(id)
}

func (s *PaymentService) CreatePaymentMethod(method *payment.PaymentMethod) error {
	return s.paymentRepo.CreatePaymentMethod(method)
}

func (s *PaymentService) UpdatePaymentMethod(method *payment.PaymentMethod) error {
	existing, err := s.paymentRepo.FindPaymentMethodByID(method.ID)
	if err != nil {
		return err
	}

	existing.Name = method.Name
	existing.Code = method.Code
	existing.Icon = method.Icon
	existing.Description = method.Description
	existing.FeeType = method.FeeType
	existing.FeeValue = method.FeeValue
	existing.MinAmount = method.MinAmount
	existing.MaxAmount = method.MaxAmount
	existing.Enabled = method.Enabled
	existing.SortOrder = method.SortOrder
	existing.Settings = method.Settings

	return s.paymentRepo.UpdatePaymentMethod(existing)
}

func (s *PaymentService) DeletePaymentMethod(id uint) error {
	return s.paymentRepo.DeletePaymentMethod(id)
}

func (s *PaymentService) ListTaxRates() ([]payment.TaxRate, error) {
	return s.paymentRepo.FindAllTaxRates(false)
}

func (s *PaymentService) GetTaxRate(id uint) (*payment.TaxRate, error) {
	return s.paymentRepo.FindTaxRateByID(id)
}

func (s *PaymentService) ListPublicTaxRates() ([]payment.TaxRate, error) {
	return s.paymentRepo.FindAllTaxRates(true)
}

func (s *PaymentService) GetPublicTaxRate(id uint) (*payment.TaxRate, error) {
	rate, err := s.paymentRepo.FindTaxRateByID(id)
	if err != nil {
		return nil, err
	}
	if !rate.Enabled {
		return nil, ErrPaymentNotFound
	}
	return rate, nil
}

// CalculateTaxMoney is the transactional tax path. It keeps the taxable
// amount and computed tax in one currency-specific minor-unit model.
func (s *PaymentService) CalculateTaxMoney(amountMoney domainmoney.Money, country, state string, postalCodes ...string) (float64, domainmoney.Money, error) {
	if err := amountMoney.Validate(); err != nil {
		return 0, domainmoney.Money{}, fmt.Errorf("invalid tax amount: %w", err)
	}
	taxRate, err := s.paymentRepo.FindTaxRateByLocation(country, state, postalCodes...)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, domainmoney.MustNew(0, amountMoney.Currency().String()), nil
		}
		return 0, domainmoney.Money{}, fmt.Errorf("failed to load tax rate for %s/%s: %w", country, state, err)
	}
	if taxRate == nil {
		return 0, domainmoney.Money{}, errors.New("tax rate lookup returned no result")
	}
	taxMoney, err := taxRate.CalculateTaxMoney(amountMoney)
	if err != nil {
		return 0, domainmoney.Money{}, err
	}
	return taxRate.Rate, taxMoney, nil
}
