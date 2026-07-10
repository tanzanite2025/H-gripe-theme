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
}

var (
	ErrOrderNotFound            = errors.New("order not found")
	ErrOrderDeleteNotAllowed    = errors.New("only cancelled or refunded orders can be deleted")
	ErrSystemManagedOrderStatus = errors.New("order status is managed by payment workflow")
)

func NewOrderService(
	txManager *repository.TxManager,
	orderRepo *repository.OrderRepository,
	checkout *CheckoutService,
) *OrderService {
	return &OrderService{
		txManager: txManager,
		orderRepo: orderRepo,
		checkout:  checkout,
	}
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
