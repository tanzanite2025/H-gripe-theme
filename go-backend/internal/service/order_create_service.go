package service

import (
	"context"
	"fmt"
	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/domain/order"
	"tanzanite/internal/pkg/logger"
	"tanzanite/internal/pkg/requestctx"
	"tanzanite/internal/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

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

		variantItemsMap := make(map[uint]int)
		for _, item := range quote.Items {
			if item.VariantID == nil {
				return fmt.Errorf("[CRITICAL] Missing variant for product ID %d", item.ProductID)
			}
			variantItemsMap[*item.VariantID] += item.Quantity
		}
		if err := repos.Product.DecrementVariantStocks(variantItemsMap); err != nil {
			return fmt.Errorf("[CRITICAL] Failed to deduct variant stock in bulk: %w", err)
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

func (s *OrderService) generateOrderNumber() string {
	return fmt.Sprintf("ORD%s%s", time.Now().Format("20060102"), uuid.New().String()[:8])
}
