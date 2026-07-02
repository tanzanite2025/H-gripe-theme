package service

import (
	"context"
	"errors"
	"fmt"
	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/domain/order"
	"tanzanite/internal/pkg/logger"
	"tanzanite/internal/pkg/requestctx"
	"tanzanite/internal/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrderService struct {
	txManager *repository.TxManager
	orderRepo *repository.OrderRepository
	checkout  *CheckoutService
}

var (
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderDeleteNotAllowed = errors.New("only cancelled or refunded orders can be deleted")
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

func (s *OrderService) CreateOrder(
	ctx context.Context,
	userID uint,
	items []order.OrderItem,
	shippingAddress order.Address,
	billingAddress order.Address,
	paymentMethod string,
	shippingMethod string,
	couponCode string,
	pointsToUse int,
) (*order.Order, error) {
	traceID := ""
	if ctx != nil {
		if tid, ok := requestctx.TraceID(ctx); ok {
			traceID = tid
		}
	}
	logger.Info("CreateOrder started", zap.String("trace_id", traceID), zap.Uint("user_id", userID))

	quoteInput := CheckoutQuoteInput{
		UserID:          userID,
		Items:           items,
		ShippingAddress: shippingAddress,
		CouponCode:      couponCode,
		PointsToUse:     pointsToUse,
	}

	var createdOrder *order.Order
	txErr := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		quote, err := s.checkout.QuoteWithRepositories(quoteInput, repos)
		if err != nil {
			return err
		}

		o := &order.Order{
			OrderNumber:     s.generateOrderNumber(),
			UserID:          userID,
			Status:          "pending",
			PaymentMethod:   paymentMethod,
			PaymentStatus:   "unpaid",
			ShippingMethod:  shippingMethod,
			ShippingStatus:  "pending",
			SubtotalAmount:  quote.SubtotalAmount,
			TotalAmount:     quote.TotalAmount,
			ShippingFee:     quote.ShippingFee,
			TaxAmount:       quote.TaxAmount,
			DiscountAmount:  quote.DiscountAmount,
			CouponCode:      quote.CouponCode,
			PointsUsed:      quote.PointsToUse,
			PointsValue:     quote.PointsDiscount,
			Items:           quote.Items,
			ShippingAddress: shippingAddress,
			BillingAddress:  billingAddress,
		}

		itemsMap := make(map[uint]int)
		for _, item := range quote.Items {
			itemsMap[item.ProductID] += item.Quantity
		}
		if err := repos.Product.DecrementStocks(itemsMap); err != nil {
			return fmt.Errorf("[CRITICAL] Failed to deduct stock in bulk: %w", err)
		}

		if err := repos.Order.Create(o); err != nil {
			return fmt.Errorf("[CRITICAL] Failed to create order in database: %w", err)
		}
		createdOrder = o

		if quote.PointsToUse > 0 {
			if _, err := repos.Loyalty.AdjustUserPointsInCurrentTx(
				userID,
				-quote.PointsToUse,
				"spend",
				"order",
				o.ID,
				fmt.Sprintf("Spent %d points on order #%s", quote.PointsToUse, o.OrderNumber),
			); err != nil {
				return fmt.Errorf("[CRITICAL] Failed to deduct points for order ID %d: %w", o.ID, err)
			}
		}

		if quote.Coupon != nil {
			if err := repos.Coupon.IncrementUsedCount(quote.Coupon.ID); err != nil {
				return fmt.Errorf("[CRITICAL] Failed to increment usage count for coupon ID %d: %w", quote.Coupon.ID, err)
			}

			usage := &coupon.CouponUsage{
				CouponID:  quote.Coupon.ID,
				UserID:    userID,
				OrderID:   o.ID,
				Discount:  quote.CouponDiscount,
				CreatedAt: time.Now(),
			}
			if err := repos.Coupon.CreateCouponUsage(usage); err != nil {
				return fmt.Errorf("[CRITICAL] Failed to record coupon usage for coupon ID %d: %w", quote.Coupon.ID, err)
			}
		}

		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return createdOrder, nil
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

func (s *OrderService) GetAdminOrder(id uint) (*order.Order, error) {
	return s.findOrder(id)
}

func (s *OrderService) GetUserOrders(userID uint, page, pageSize int) ([]order.Order, int64, error) {
	return s.orderRepo.FindByUserID(userID, page, pageSize)
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

	if !o.CanTransitionTo(status) {
		return fmt.Errorf("invalid status transition from %s to %s", o.Status, status)
	}

	if status == "cancelled" {
		return s.cancelOrderWithRollback(o)
	}

	return s.orderRepo.UpdateStatus(id, status)
}

func (s *OrderService) UpdatePaymentStatus(id uint, paymentStatus string) error {
	if _, err := s.findOrder(id); err != nil {
		return err
	}

	return s.orderRepo.UpdatePaymentStatus(id, paymentStatus)
}

func (s *OrderService) MarkOrderPaidByNumber(orderNumber string) error {
	if orderNumber == "" {
		return errors.New("order number is required")
	}

	o, err := s.orderRepo.FindByOrderNumber(orderNumber)
	if err != nil {
		return normalizeOrderError(err)
	}

	switch o.Status {
	case "cancelled", "refunded":
		return fmt.Errorf("cannot mark %s order as paid", o.Status)
	case "processing", "shipped", "completed":
		if o.PaymentStatus == "paid" {
			return nil
		}
		return s.orderRepo.UpdatePaymentStatus(o.ID, "paid")
	case "pending", "paid":
		return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
			if err := repos.Order.UpdatePaymentStatus(o.ID, "paid"); err != nil {
				return err
			}
			if o.Status == "pending" {
				if err := repos.Order.UpdateStatus(o.ID, "paid"); err != nil {
					return err
				}
			}
			return repos.Order.UpdateStatus(o.ID, "processing")
		})
	default:
		return fmt.Errorf("unsupported order status %s for paid webhook", o.Status)
	}
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

func (s *OrderService) CancelOrder(id uint, userID uint) error {
	o, err := s.orderRepo.FindByID(id)
	if err != nil {
		return normalizeOrderError(err)
	}

	if o.UserID != userID {
		return errors.New("unauthorized")
	}

	if o.Status != "pending" && o.Status != "paid" {
		return errors.New("order cannot be cancelled")
	}

	return s.cancelOrderWithRollback(o)
}

func (s *OrderService) cancelOrderWithRollback(o *order.Order) error {
	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if err := repos.Order.UpdateStatus(o.ID, "cancelled"); err != nil {
			return err
		}

		for _, item := range o.Items {
			if err := repos.Product.IncrementStock(item.ProductID, item.Quantity); err != nil {
				return fmt.Errorf("[CRITICAL] Failed to restore stock for product %d: %w", item.ProductID, err)
			}
		}

		if o.PointsUsed > 0 {
			_, err := repos.Loyalty.AdjustUserPointsInCurrentTx(
				o.UserID,
				o.PointsUsed,
				"refund",
				"order",
				o.ID,
				fmt.Sprintf("Order #%s cancelled points refund", o.OrderNumber),
			)
			if err != nil {
				return fmt.Errorf("[CRITICAL] Failed to refund points: %w", err)
			}
		}

		if o.CouponCode != "" {
			cp, err := repos.Coupon.FindCouponByCode(o.CouponCode)
			if err != nil {
				return fmt.Errorf("[CRITICAL] Failed to find coupon during refund: %w", err)
			}
			if cp != nil {
				if err := repos.Coupon.DecrementUsedCount(cp.ID); err != nil {
					return fmt.Errorf("[CRITICAL] Failed to restore coupon usage limit: %w", err)
				}

				if err := repos.Coupon.DeleteCouponUsageByOrderID(o.ID); err != nil {
					return fmt.Errorf("[CRITICAL] Failed to delete coupon usage log: %w", err)
				}
			}
		}

		return nil
	})
}

func (s *OrderService) findOrder(id uint) (*order.Order, error) {
	o, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, normalizeOrderError(err)
	}

	return o, nil
}

func normalizeOrderError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrOrderNotFound
	}
	return err
}

func (s *OrderService) GetOrderStats(userID uint) (map[string]int64, error) {
	return s.orderRepo.GetOrderStats(userID)
}

func (s *OrderService) generateOrderNumber() string {
	return fmt.Sprintf("ORD%s%s", time.Now().Format("20060102"), uuid.New().String()[:8])
}
