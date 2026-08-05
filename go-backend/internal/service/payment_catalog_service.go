package service

import "tanzanite/internal/domain/payment"

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

func (s *PaymentService) CalculateTax(amount float64, country, state string) (float64, float64, error) {
	taxRate, err := s.paymentRepo.FindTaxRateByLocation(country, state)
	if err != nil {
		return 0, 0, nil
	}

	tax := amount * taxRate.Rate / 100
	return taxRate.Rate, tax, nil
}
