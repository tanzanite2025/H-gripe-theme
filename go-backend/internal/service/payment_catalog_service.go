package service

import (
	"commerce-platform/internal/domain/payment"
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
	existing.FeeValueMinor = method.FeeValueMinor
	existing.FeeRateDecimal = method.FeeRateDecimal
	existing.MinAmountMinor = method.MinAmountMinor
	existing.MaxAmountMinor = method.MaxAmountMinor
	existing.Enabled = method.Enabled
	existing.SortOrder = method.SortOrder
	existing.Settings = method.Settings

	return s.paymentRepo.UpdatePaymentMethod(existing)
}

func (s *PaymentService) DeletePaymentMethod(id uint) error {
	return s.paymentRepo.DeletePaymentMethod(id)
}
