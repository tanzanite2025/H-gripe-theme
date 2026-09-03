package service

import (
	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	attributionpkg "commerce-platform/internal/pkg/attribution"
	"commerce-platform/internal/pkg/logger"
	paymentpkg "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/pkg/requestctx"
	"commerce-platform/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

type OrderCreationOptions struct {
	PolicyLocale                 string
	PolicyURL                    string
	PolicyDisclosureAcknowledged bool
	PolicySource                 string
	IdempotencyKey               string
	IdempotencyRequestHash       string
}

var ErrOrderPolicyDisclosureFailure = errors.New("order policy disclosure capture failed")

const orderCreateIdempotencyScope = "order_create"
const zeroTotalSettlementPaymentMethod = "zero_total"

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
	return s.CreateOrderWithAttributionAndOptions(
		ctx,
		userID,
		items,
		shippingAddress,
		billingAddress,
		paymentMethod,
		shippingMethod,
		couponCode,
		pointsToUse,
		attributionContext,
		OrderCreationOptions{},
	)
}

func (s *OrderService) CreateOrderWithAttributionAndOptions(
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
	options OrderCreationOptions,
) (*order.Order, error) {
	traceID := ""
	if ctx != nil {
		if tid, ok := requestctx.TraceID(ctx); ok {
			traceID = tid
		}
	}
	logger.Info("CreateOrder started", zap.String("trace_id", traceID), zap.Uint("user_id", userID))

	idempotencyKey := strings.TrimSpace(options.IdempotencyKey)
	idempotencyRequestHash := strings.TrimSpace(options.IdempotencyRequestHash)
	if idempotencyKey != "" && idempotencyRequestHash == "" {
		return nil, ErrOrderIdempotencyHashRequired
	}

	var affectedProductIDs []uint
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

	var createdOrder *order.Order
	txErr := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		var idempotencyRecord *order.OrderIdempotency
		if idempotencyKey != "" {
			if repos.OrderIdempotency == nil {
				return ErrOrderIdempotencyUnavailable
			}
			record := &order.OrderIdempotency{
				UserID:         userID,
				Scope:          orderCreateIdempotencyScope,
				IdempotencyKey: idempotencyKey,
				RequestHash:    idempotencyRequestHash,
			}
			claimed, err := repos.OrderIdempotency.TryCreate(record)
			if err != nil {
				return err
			}
			if !claimed {
				existing, err := repos.OrderIdempotency.FindByUserScopeKey(userID, orderCreateIdempotencyScope, idempotencyKey)
				if repository.IsRecordNotFound(err) {
					claimed, err = repos.OrderIdempotency.TryCreate(record)
					if err != nil {
						return err
					}
					if claimed {
						idempotencyRecord = record
					} else {
						existing, err = repos.OrderIdempotency.FindByUserScopeKey(userID, orderCreateIdempotencyScope, idempotencyKey)
					}
				}
				if err != nil {
					return err
				}
				if idempotencyRecord == nil {
					if strings.TrimSpace(existing.RequestHash) != idempotencyRequestHash {
						return ErrOrderIdempotencyConflict
					}
					if existing.OrderID != nil && *existing.OrderID > 0 {
						replayedOrder, err := repos.Order.FindByID(*existing.OrderID)
						if err != nil {
							return normalizeOrderError(err)
						}
						createdOrder = replayedOrder
						return nil
					}
					return ErrOrderIdempotencyInProgress
				}
			} else {
				idempotencyRecord = record
			}
		}

		quote, err := s.checkout.QuoteWithRepositories(quoteInput, repos)
		if err != nil {
			return err
		}
		orderCurrency := quote.Currency
		if provider := paymentpkg.ProviderForPaymentMethod(paymentMethod); provider != "" && quote.TotalAmount > 0 {
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
		isZeroTotalOrder := quote.TotalAmount <= 0
		orderStatus := "pending"
		paymentStatus := "unpaid"
		var paidAt *time.Time
		if isZeroTotalOrder {
			paidAtValue := time.Now().UTC()
			orderStatus = "processing"
			paymentStatus = "paid"
			paidAt = &paidAtValue
		}

		orderNumber, err := s.generateOrderNumber()
		if err != nil {
			return err
		}

		o := &order.Order{
			OrderNumber:      orderNumber,
			UserID:           userID,
			Status:           orderStatus,
			PaymentMethod:    paymentMethod,
			PaymentStatus:    paymentStatus,
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
			PaidAt:           paidAt,
		}

		variantItemsMap := make(map[uint]int)
		for _, item := range quote.Items {
			if item.VariantID == nil {
				return fmt.Errorf("[CRITICAL] Missing variant for product ID %d", item.ProductID)
			}
			variantItemsMap[*item.VariantID] += item.Quantity
		}
		productIDs, err := repos.Product.DecrementVariantStocks(variantItemsMap)
		if err != nil {
			return fmt.Errorf("[CRITICAL] Failed to deduct variant stock in bulk: %w", err)
		}
		affectedProductIDs = append(affectedProductIDs, productIDs...)
		if err := s.enqueueProductCacheInvalidationInTx(repos, productIDs, "order stock deducted"); err != nil {
			return fmt.Errorf("[CRITICAL] Failed to enqueue product cache invalidation: %w", err)
		}

		if err := repos.Order.Create(o); err != nil {
			return fmt.Errorf("[CRITICAL] Failed to create order in database: %w", err)
		}
		if isZeroTotalOrder {
			if err := settleZeroTotalOrderInTx(repos, o, orderCurrency, paidAt); err != nil {
				return fmt.Errorf("[CRITICAL] Failed to settle zero-total order ID %d: %w", o.ID, err)
			}
		}
		if idempotencyRecord != nil {
			if err := repos.OrderIdempotency.BindOrderID(idempotencyRecord.ID, o.ID); err != nil {
				return fmt.Errorf("[CRITICAL] Failed to bind order idempotency record: %w", err)
			}
		}
		createdOrder = o
		if s.refundCancellationPolicy != nil {
			if repos.PolicyDisclosure == nil || repos.Setting == nil {
				return fmt.Errorf("%w: repositories are not configured", ErrOrderPolicyDisclosureFailure)
			}
			var consentedAt *time.Time
			if options.PolicyDisclosureAcknowledged {
				value := time.Now().UTC()
				consentedAt = &value
			}
			disclosure, err := s.refundCancellationPolicy.BuildOrderDisclosure(
				repos.Setting,
				o.ID,
				options.PolicyLocale,
				options.PolicyURL,
				options.PolicySource,
				consentedAt,
			)
			if err != nil {
				return fmt.Errorf("[CRITICAL] %w: capture refund and cancellation policy disclosure: %w", ErrOrderPolicyDisclosureFailure, err)
			}
			if err := repos.PolicyDisclosure.Create(disclosure); err != nil {
				return fmt.Errorf("[CRITICAL] %w: save refund and cancellation policy disclosure: %w", ErrOrderPolicyDisclosureFailure, err)
			}
		}
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

	s.invalidateProductCacheAfterStockCommit(affectedProductIDs)

	return createdOrder, nil
}

func settleZeroTotalOrderInTx(repos repository.TxRepositories, o *order.Order, orderCurrency string, paidAt *time.Time) error {
	if repos.Payment == nil {
		return errors.New("payment transaction repository is not configured")
	}
	if o == nil {
		return errors.New("order is required")
	}
	settledAt := time.Now().UTC()
	if paidAt != nil && !paidAt.IsZero() {
		settledAt = paidAt.UTC()
	}
	transactionID := zeroTotalOrderTransactionID(o.ID)
	transaction := &paymentdomain.Transaction{
		OrderID:         o.ID,
		TransactionID:   transactionID,
		PaymentMethod:   zeroTotalSettlementPaymentMethod,
		Amount:          0,
		Currency:        orderCurrency,
		Status:          "completed",
		GatewayResponse: `{"settlement":"zero_total","reason":"discounts_cover_total"}`,
		CompletedAt:     &settledAt,
	}
	if _, err := repos.Payment.CreateTransactionIfAbsent(transaction); err != nil {
		return err
	}
	return enqueueOrderPaidOutboxEvent(repos.Outbox, o, VerifiedGatewayPaymentInput{
		Provider:      zeroTotalSettlementPaymentMethod,
		OrderNumber:   o.OrderNumber,
		TransactionID: transactionID,
		PaymentMethod: zeroTotalSettlementPaymentMethod,
		Amount:        0,
		Currency:      orderCurrency,
	}, settledAt)
}

func zeroTotalOrderTransactionID(orderID uint) string {
	return fmt.Sprintf("zero-total-%d", orderID)
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
