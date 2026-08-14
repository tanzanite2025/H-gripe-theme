package service

import (
	"context"
	"errors"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/repository"
)

const showcaseUploadOrderListLimit = 100

// ShowcaseUploadOrderOption is the minimal order contract needed by the
// Picture Warehouse upload form. It excludes addresses and line items.
type ShowcaseUploadOrderOption struct {
	ID             uint       `json:"id"`
	OrderNumber    string     `json:"order_number"`
	Status         string     `json:"status"`
	ShippingStatus string     `json:"shipping_status"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	TotalAmount    float64    `json:"total_amount"`
	Currency       string     `json:"currency"`
	Eligible       bool       `json:"eligible"`
}

// ShowcaseUploadEligibilityService owns the business rule that a Picture
// Warehouse upload must be tied to the current user's completed order.
type ShowcaseUploadEligibilityService struct {
	orderRepo *repository.OrderRepository
}

func NewShowcaseUploadEligibilityService(orderRepo *repository.OrderRepository) *ShowcaseUploadEligibilityService {
	return &ShowcaseUploadEligibilityService{orderRepo: orderRepo}
}

func (s *ShowcaseUploadEligibilityService) ListOrderOptions(ctx context.Context, userID uint) ([]ShowcaseUploadOrderOption, error) {
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

	options := make([]ShowcaseUploadOrderOption, 0, len(orders))
	for _, item := range orders {
		options = append(options, showcaseUploadOrderOption(item))
	}
	return options, nil
}

func (s *ShowcaseUploadEligibilityService) RequireEligibleOrder(ctx context.Context, userID, orderID uint) (*order.Order, error) {
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

func showcaseUploadOrderOption(item order.Order) ShowcaseUploadOrderOption {
	return ShowcaseUploadOrderOption{
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
