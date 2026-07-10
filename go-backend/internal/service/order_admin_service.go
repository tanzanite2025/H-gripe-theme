package service

import (
	"fmt"
	"tanzanite/internal/domain/order"
	"time"
)

func (s *OrderService) GetAdminOrder(id uint) (*order.Order, error) {
	return s.findOrder(id)
}

func (s *OrderService) GetAllOrders(page, pageSize int, status string) ([]order.Order, int64, error) {
	return s.orderRepo.FindAll(page, pageSize, status)
}

func (s *OrderService) ListAdminOrders(page, pageSize int, status, paymentStatus, shippingStatus, search, startDate, endDate string) ([]order.Order, int64, error) {
	return s.orderRepo.FindAllWithFilters(page, pageSize, status, paymentStatus, shippingStatus, search, startDate, endDate)
}

func (s *OrderService) UpdateOrderStatus(id uint, status string) error {
	o, err := s.orderRepo.FindByID(id)
	if err != nil {
		return normalizeOrderError(err)
	}

	if isSystemManagedOrderStatus(status) {
		return fmt.Errorf("%w: %s", ErrSystemManagedOrderStatus, status)
	}

	if !o.CanTransitionTo(status) {
		return fmt.Errorf("invalid status transition from %s to %s", o.Status, status)
	}

	if status == "cancelled" {
		return s.cancelOrderWithRollback(o)
	}

	return s.orderRepo.UpdateStatus(id, status)
}

func isSystemManagedOrderStatus(status string) bool {
	return status == "paid" || status == "refunded"
}

func (s *OrderService) UpdateShippingStatus(id uint, shippingStatus string) error {
	if _, err := s.findOrder(id); err != nil {
		return err
	}

	return s.orderRepo.UpdateShippingStatus(id, shippingStatus)
}

func (s *OrderService) UpdateTrackingInfo(id uint, trackingNumber, carrierCode string) error {
	if _, err := s.findOrder(id); err != nil {
		return err
	}

	return s.orderRepo.UpdateTrackingInfo(id, trackingNumber, carrierCode)
}

func (s *OrderService) UpdateAdminNote(id uint, adminNote string) error {
	o, err := s.findOrder(id)
	if err != nil {
		return err
	}

	o.AdminNote = adminNote
	return s.orderRepo.Update(o)
}

func (s *OrderService) DeleteAdminOrder(id uint) error {
	o, err := s.findOrder(id)
	if err != nil {
		return err
	}

	if o.Status != "cancelled" && o.Status != "refunded" {
		return ErrOrderDeleteNotAllowed
	}

	return s.orderRepo.Delete(id)
}

func (s *OrderService) GetAdminStats() (map[string]interface{}, error) {
	return s.orderRepo.GetStats()
}

func (s *OrderService) GetSalesByDateRange(startDate, endDate time.Time) ([]map[string]interface{}, error) {
	return s.orderRepo.GetSalesByDateRange(startDate, endDate)
}
