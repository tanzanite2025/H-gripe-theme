package service

import (
	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/order"
	attributionpkg "commerce-platform/internal/pkg/attribution"
	"commerce-platform/internal/pkg/logger"
	paymentpkg "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/pkg/requestctx"
	"commerce-platform/internal/repository"
	"context"
	"fmt"
	"strings"
	"time"

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
	return s.CreateOrderWithAttribution(
		ctx,
		userID,
		items,
		shippingAddress,
		billingAddress,
		paymentMethod,
		shippingMethod,
		couponCode,
		pointsToUse,
		attributionpkg.Context{},
	)
}

func (s *OrderService) CreateOrderWithAttribution(
	ctx context.Context,
	userID uint,
	items []order.OrderItem,
	shippingAddress order.Address,
	billingAddress order.Address,
	paymentMethod string,
	shippingMethod string,
	couponCode string,
	pointsToUse int,
	attributionContext attributionpkg.Context,
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
	if pointsToUse > 0 {
		config, err := s.checkout.currentLoyaltyProgramConfig()
		if err != nil {
			return nil, err
		}
		quoteInput.LoyaltyProgramConfig = config
	}
	orderNumber, err := s.generateOrderNumber()
	if err != nil {
		return nil, err
	}

	var createdOrder *order.Order
	txErr := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		quote, err := s.checkout.QuoteWithRepositories(quoteInput, repos)
		if err != nil {
			return err
		}
		orderCurrency := quote.Currency
		if provider := paymentpkg.ProviderForPaymentMethod(paymentMethod); provider != "" {
			if err := paymentpkg.ValidateGatewayCurrency(paymentpkg.GatewayType(provider), orderCurrency); err != nil {
				return fmt.Errorf(
					"payment method %s cannot process order currency %s: %w",
					strings.TrimSpace(paymentMethod),
					orderCurrency,
					err,
				)
			}
		}

		shippingMethodSnapshot := strings.TrimSpace(shippingMethod)
		var carrierID *uint
		var carrierServiceID *uint
		if quote.ShippingQuote != nil && quote.ShippingQuote.SelectedOption != nil {
			selectedOption := quote.ShippingQuote.SelectedOption
			if selectedOption.CarrierID > 0 {
				carrierID = uintPtr(selectedOption.CarrierID)
			}
			if selectedOption.CarrierServiceID > 0 {
				carrierServiceID = uintPtr(selectedOption.CarrierServiceID)
			}
			if label := shippingQuoteOptionSnapshot(*selectedOption); label != "" {
				shippingMethodSnapshot = label
			}
		}
		if shippingMethodSnapshot == "" {
			shippingMethodSnapshot = "standard"
		}

		o := &order.Order{
			OrderNumber:      orderNumber,
			UserID:           userID,
			Status:           "pending",
			PaymentMethod:    paymentMethod,
			PaymentStatus:    "unpaid",
			ShippingMethod:   shippingMethodSnapshot,
			ShippingStatus:   "pending",
			CarrierID:        carrierID,
			CarrierServiceID: carrierServiceID,
			SubtotalAmount:   quote.SubtotalAmount,
			TotalAmount:      quote.TotalAmount,
			ShippingFee:      quote.ShippingFee,
			TaxAmount:        quote.TaxAmount,
			DiscountAmount:   quote.DiscountAmount,
			Currency:         orderCurrency,
			CouponCode:       quote.CouponCode,
			PointsUsed:       quote.PointsToUse,
			PointsValue:      quote.PointsDiscount,
			FXSnapshotData:   currency.OrderFXSnapshotJSON(quote.FXSnapshot),
			Items:            quote.Items,
			ShippingAddress:  shippingAddress,
			BillingAddress:   billingAddress,
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
		if err := persistOrderAttribution(repos.OrderAttribution, o.ID, attributionContext); err != nil {
			return fmt.Errorf("[CRITICAL] Failed to save order attribution: %w", err)
		}

		if quote.PointsToUse > 0 {
			if _, err := repos.Loyalty.AdjustUserPointsInCurrentTxWithConfig(
				userID,
				-quote.PointsToUse,
				"spend",
				"order",
				o.ID,
				fmt.Sprintf("Spent %d points on order #%s", quote.PointsToUse, o.OrderNumber),
				quote.ProgramConfigID,
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

func shippingQuoteOptionSnapshot(option ShippingQuoteOption) string {
	parts := []string{}
	carrier := strings.TrimSpace(option.CarrierName)
	serviceName := strings.TrimSpace(option.ServiceName)
	routeName := strings.TrimSpace(option.RouteName)
	serviceCode := strings.TrimSpace(option.ServiceCode)

	if carrier != "" {
		parts = append(parts, carrier)
	}
	if routeName != "" && routeName != serviceName {
		parts = append(parts, routeName)
	}
	if serviceName != "" {
		if serviceCode != "" {
			serviceName = fmt.Sprintf("%s (%s)", serviceName, serviceCode)
		}
		parts = append(parts, serviceName)
	} else if serviceCode != "" {
		parts = append(parts, serviceCode)
	}

	return strings.Join(parts, " / ")
}

func (s *OrderService) generateOrderNumber() (string, error) {
	if s == nil || s.numberGenerator == nil {
		return "", ErrOrderNumberNotConfigured
	}
	value, err := s.numberGenerator.Generate()
	if err != nil {
		return "", err
	}
	if !s.numberGenerator.Validate(value) {
		return "", fmt.Errorf("generated order number failed validation")
	}
	return value, nil
}
