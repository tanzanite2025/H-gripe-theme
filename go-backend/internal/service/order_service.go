package service

import (
	"errors"
	"tanzanite/internal/domain/order"
	"tanzanite/internal/repository"
)

type OrderService struct {
	txManager *repository.TxManager
	orderRepo *repository.OrderRepository
	checkout  *CheckoutService
	shipping  *ShippingService
}

var (
	ErrOrderNotFound              = errors.New("order not found")
	ErrOrderDeleteNotAllowed      = errors.New("only cancelled or refunded orders can be deleted")
	ErrSystemManagedOrderStatus   = errors.New("order status is managed by payment workflow")
	ErrTrackingNumberRequired     = errors.New("tracking number is required")
	ErrOrderShippingNotConfigured = errors.New("order shipping service is not configured")
)

func NewOrderService(
	txManager *repository.TxManager,
	orderRepo *repository.OrderRepository,
	checkout *CheckoutService,
	shippingServices ...*ShippingService,
) *OrderService {
	service := &OrderService{
		txManager: txManager,
		orderRepo: orderRepo,
		checkout:  checkout,
	}
	if len(shippingServices) > 0 {
		service.shipping = shippingServices[0]
	}
	return service
}

type OrderTrackingUpdateInput struct {
	TrackingNumber     string
	TrackingProviderID uint
	CarrierID          *uint
	CarrierServiceID   *uint
}

func (s *OrderService) GetOrder(id uint, userID uint) (*order.Order, error) {
	o, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, normalizeOrderError(err)
	}

	if o.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	return o, nil
}

func (s *OrderService) GetUserOrders(userID uint, page, pageSize int) ([]order.Order, int64, error) {
	return s.orderRepo.FindByUserID(userID, page, pageSize)
}

func (s *OrderService) findOrder(id uint) (*order.Order, error) {
	o, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, normalizeOrderError(err)
	}

	return o, nil
}

func normalizeOrderError(err error) error {
	if repository.IsRecordNotFound(err) {
		return ErrOrderNotFound
	}
	return err
}

func (s *OrderService) GetOrderStats(userID uint) (map[string]int64, error) {
	return s.orderRepo.GetOrderStats(userID)
}
