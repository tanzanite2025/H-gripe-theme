package service

import "tanzanite/internal/domain/payment"

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
