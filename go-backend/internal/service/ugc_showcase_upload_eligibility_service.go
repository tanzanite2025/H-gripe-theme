package service

import (
	"context"
	"errors"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/repository"
)

const showcaseUploadOrderListLimit = 100

// UGCShowcaseUploadOrderOption is the minimal order contract needed by the
// Picture Warehouse upload form. It excludes addresses and line items.
type UGCShowcaseUploadOrderOption struct {
	ID             uint       `json:"id"`
	OrderNumber    string     `json:"order_number"`
	Status         string     `json:"status"`
	ShippingStatus string     `json:"shipping_status"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	TotalAmount    float64    `json:"total_amount"`
	Currency       string     `json:"currency"`
	Eligible       bool       `json:"eligible"`
}

// UGCShowcaseUploadEligibilityService owns the business rule that a Picture
// Warehouse upload must be tied to the current user's completed order.
type UGCShowcaseUploadEligibilityService struct {
	orderRepo *repository.OrderRepository
}

func NewUGCShowcaseUploadEligibilityService(orderRepo *repository.OrderRepository) *UGCShowcaseUploadEligibilityService {
	return &UGCShowcaseUploadEligibilityService{orderRepo: orderRepo}
}

func (s *UGCShowcaseUploadEligibilityService) ListOrderOptions(ctx context.Context, userID uint) ([]UGCShowcaseUploadOrderOption, error) {
	if s == nil || s.orderRepo == nil {
		return nil, ErrShowcaseUploadEligibilityUnavailable
	}
	if userID == 0 {
		return nil, ErrShowcaseUploadOrderNotEligible
	}

	orders, err := s.orderRepo.FindByUserIDForShowcaseUpload(userID, showcaseUploadOrderListLimit)
	if err != nil {
		return nil, err
	}

	options := make([]UGCShowcaseUploadOrderOption, 0, len(orders))
	for _, item := range orders {
		options = append(options, ugcShowcaseUploadOrderOption(item))
	}
	return options, nil
}

func (s *UGCShowcaseUploadEligibilityService) RequireEligibleOrder(ctx context.Context, userID, orderID uint) (*order.Order, error) {
	if s == nil || s.orderRepo == nil {
		return nil, ErrShowcaseUploadEligibilityUnavailable
	}
	if orderID == 0 {
		return nil, ErrShowcaseUploadOrderRequired
	}
	if userID == 0 {
		return nil, ErrShowcaseUploadOrderNotEligible
	}

	item, err := s.orderRepo.FindByIDAndUserIDForShowcaseUpload(orderID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, ErrShowcaseUploadOrderNotEligible
		}
		return nil, err
	}
	if !isCompletedShowcaseUploadOrder(*item) {
		return nil, ErrShowcaseUploadOrderNotEligible
	}
	return item, nil
}

func ugcShowcaseUploadOrderOption(item order.Order) UGCShowcaseUploadOrderOption {
	return UGCShowcaseUploadOrderOption{
		ID:             item.ID,
		OrderNumber:    item.OrderNumber,
		Status:         item.Status,
		ShippingStatus: item.ShippingStatus,
		CompletedAt:    item.CompletedAt,
		TotalAmount:    item.TotalAmount,
		Currency:       item.Currency,
		Eligible:       isCompletedShowcaseUploadOrder(item),
	}
}

func isCompletedShowcaseUploadOrder(item order.Order) bool {
	return item.Status == "completed" && item.CompletedAt != nil
}
