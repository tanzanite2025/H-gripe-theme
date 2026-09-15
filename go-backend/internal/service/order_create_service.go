package service

import (
	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	domainpricing "commerce-platform/internal/domain/pricing"
	productdomain "commerce-platform/internal/domain/product"
	attributionpkg "commerce-platform/internal/pkg/attribution"
	"commerce-platform/internal/pkg/logger"
	paymentpkg "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/pkg/requestctx"
	"commerce-platform/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type OrderCreationOptions struct {
	PolicyLocale                 string
	PolicyURL                    string
	PolicyDisclosureAcknowledged bool
	PolicySource                 string
	IdempotencyKey               string
	IdempotencyRequestHash       string
	ExpectedTotal                *float64
	ShippingQuoteID              string
	SelectedQuotePlanID          string
	CheckoutCartID               uint
	GiftCardCode                 string
	DisplayCurrency              string
}

var ErrOrderPolicyDisclosureFailure = errors.New("order policy disclosure capture failed")
var ErrOrderTotalChanged = errors.New("price has been updated; please review the order total")
var ErrCheckoutCartAlreadyConsumed = errors.New("checkout cart has already been consumed; check the order status before retrying")
var ErrCheckoutCartRepositoryUnavailable = errors.New("checkout cart transaction repository is not configured")
var ErrCheckoutCartConsumptionConflict = errors.New("checkout cart changed while the order was being created")

const orderCreateIdempotencyScope = "order_create"
const zeroTotalSettlementPaymentMethod = "zero_total"

type orderMoneyFields struct {
	PaymentAmountMinor  int64
	SubtotalAmountMinor int64
	ShippingFeeMinor    int64
	TaxAmountMinor      int64
	DiscountAmountMinor int64
	TotalAmountMinor    int64
	PointsValueMinor    int64
}

func moneyMinorFromMajor(value float64, currencyCode string) (int64, error) {
	amount, err := domainmoney.FromMajorFloat(value, currencyCode)
	if err != nil {
		return 0, err
	}
	return amount.AmountMinor(), nil
}

func populateOrderItemMoneyFields(items []order.OrderItem, orderCurrency string) error {
	for i := range items {
		if items[i].Currency == "" {
			items[i].Currency = orderCurrency
		}
		itemCurrency := items[i].Currency
		var err error
		if items[i].PriceMinor, err = moneyMinorFromMajor(items[i].Price, itemCurrency); err != nil {
			return fmt.Errorf("item %d price: %w", i, err)
		}
		if items[i].SubtotalMinor, err = moneyMinorFromMajor(items[i].Subtotal, itemCurrency); err != nil {
			return fmt.Errorf("item %d subtotal: %w", i, err)
		}
		if items[i].TaxAmountMinor, err = moneyMinorFromMajor(items[i].TaxAmount, itemCurrency); err != nil {
			return fmt.Errorf("item %d tax: %w", i, err)
		}
		if items[i].DiscountMinor, err = moneyMinorFromMajor(items[i].Discount, itemCurrency); err != nil {
			return fmt.Errorf("item %d discount: %w", i, err)
		}
		if items[i].TotalMinor, err = moneyMinorFromMajor(items[i].Total, itemCurrency); err != nil {
			return fmt.Errorf("item %d total: %w", i, err)
		}
		// The immutable pricing snapshot is authoritative for order-time
		// discount, tax, and net line totals. Legacy major projections above are
		// intentionally left untouched for historical reporting.
		if pricingSnapshotIncludesTax(items[i].PricingSnapshotData) {
			snapshot, snapshotErr := domainpricing.ParseLineSnapshot(items[i].PricingSnapshotData)
			if snapshotErr != nil {
				return fmt.Errorf("item %d pricing snapshot: %w", i, snapshotErr)
			}
			items[i].DiscountMinor = snapshot.DiscountTotal().AmountMinor()
			items[i].TaxAmountMinor = snapshot.Tax().AmountMinor()
			netWithTax, addErr := snapshot.NetSubtotal().Add(snapshot.Tax())
			if addErr != nil {
				return fmt.Errorf("item %d pricing snapshot total: %w", i, addErr)
			}
			items[i].TotalMinor = netWithTax.AmountMinor()
		}
	}
	return nil
}

func buildOrderMoneyFields(quote *CheckoutQuote, orderCurrency, paymentCurrency string, paymentAmount float64) (orderMoneyFields, error) {
	if quote == nil {
		return orderMoneyFields{}, errors.New("checkout quote is required")
	}
	var result orderMoneyFields
	var err error
	if result.PaymentAmountMinor, err = moneyMinorFromMajor(paymentAmount, paymentCurrency); err != nil {
		return orderMoneyFields{}, fmt.Errorf("payment amount: %w", err)
	}
	if result.SubtotalAmountMinor, err = moneyMinorFromMajor(quote.SubtotalAmount, orderCurrency); err != nil {
		return orderMoneyFields{}, fmt.Errorf("subtotal amount: %w", err)
	}
	if result.ShippingFeeMinor, err = moneyMinorFromMajor(quote.ShippingFee, orderCurrency); err != nil {
		return orderMoneyFields{}, fmt.Errorf("shipping fee: %w", err)
	}
	if result.TaxAmountMinor, err = moneyMinorFromMajor(quote.TaxAmount, orderCurrency); err != nil {
		return orderMoneyFields{}, fmt.Errorf("tax amount: %w", err)
	}
	if result.DiscountAmountMinor, err = moneyMinorFromMajor(quote.DiscountAmount, orderCurrency); err != nil {
		return orderMoneyFields{}, fmt.Errorf("discount amount: %w", err)
	}
	if result.TotalAmountMinor, err = moneyMinorFromMajor(quote.TotalAmount, orderCurrency); err != nil {
		return orderMoneyFields{}, fmt.Errorf("total amount: %w", err)
	}
	if result.PointsValueMinor, err = moneyMinorFromMajor(quote.PointsDiscount, orderCurrency); err != nil {
		return orderMoneyFields{}, fmt.Errorf("points value: %w", err)
	}
	return result, nil
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

	if s.orderEvidenceSnapshot == nil || s.orderEvidence == nil {
		return nil, ErrOrderEvidenceNotConfigured
	}

	idempotencyKey := strings.TrimSpace(options.IdempotencyKey)
	idempotencyRequestHash := strings.TrimSpace(options.IdempotencyRequestHash)
	if idempotencyKey != "" && idempotencyRequestHash == "" {
		return nil, ErrOrderIdempotencyHashRequired
	}

	var affectedProductIDs []uint
	quoteInput := CheckoutQuoteInput{
		UserID:              userID,
		Items:               items,
		ShippingAddress:     shippingAddress,
		ShippingQuoteID:     options.ShippingQuoteID,
		SelectedQuotePlanID: options.SelectedQuotePlanID,
		CouponCode:          couponCode,
		GiftCardCode:        options.GiftCardCode,
		DisplayCurrency:     options.DisplayCurrency,
		PaymentMethod:       paymentMethod,
		PointsToUse:         pointsToUse,
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

		var lockedCheckoutCartItemIDs []uint
		if options.CheckoutCartID > 0 {
			if repos.Cart == nil {
				return ErrCheckoutCartRepositoryUnavailable
			}
			if _, err := repos.Cart.FindAuthenticatedUserCartByIDForUpdate(options.CheckoutCartID, userID); err != nil {
				if repository.IsRecordNotFound(err) {
					return ErrCheckoutCartAlreadyConsumed
				}
				return fmt.Errorf("lock checkout cart %d: %w", options.CheckoutCartID, err)
			}
			cartItems, err := repos.Cart.FindCheckoutCartItemsByIDForUpdate(options.CheckoutCartID)
			if err != nil {
				return fmt.Errorf("lock checkout cart items for cart %d: %w", options.CheckoutCartID, err)
			}
			if len(cartItems) == 0 {
				return ErrCheckoutCartAlreadyConsumed
			}
			quoteInput.Items = make([]order.OrderItem, len(cartItems))
			lockedCheckoutCartItemIDs = make([]uint, len(cartItems))
			for i, cartItem := range cartItems {
				cartPriceMoney, priceErr := cartItem.PriceMoney()
				if priceErr != nil {
					return fmt.Errorf("cart item %d price: %w", cartItem.ID, priceErr)
				}
				cartPrice, priceErr := cartPriceMoney.MajorFloat()
				if priceErr != nil {
					return fmt.Errorf("cart item %d price: %w", cartItem.ID, priceErr)
				}
				quoteInput.Items[i] = order.OrderItem{
					ProductID:         cartItem.ProductID,
					VariantID:         cartItem.VariantID,
					Quantity:          cartItem.Quantity,
					Currency:          cartPriceMoney.Currency().String(),
					Price:             cartPrice,
					ConfigurationData: append([]byte(nil), cartItem.ConfigurationData...),
					ConfigurationHash: cartItem.ConfigurationHash,
				}
				lockedCheckoutCartItemIDs[i] = cartItem.ID
			}
		}

		quote, err := s.checkout.QuoteWithRepositories(quoteInput, repos)
		if err != nil {
			return err
		}
		quote.Items, err = attachPricingSnapshotsToOrderItems(quote.Items, quote.PricingSnapshot)
		if err != nil {
			return fmt.Errorf("persist order-item pricing snapshots: %w", err)
		}
		if err := populateOrderItemMoneyFields(quote.Items, quote.Currency); err != nil {
			return fmt.Errorf("persist order-item money snapshot: %w", err)
		}
		orderPricingSnapshot, err := marshalOrderPricingSnapshot(quote)
		if err != nil {
			return fmt.Errorf("persist order pricing snapshot: %w", err)
		}
		if options.ExpectedTotal != nil {
			quoteTotalMoney, quoteTotalErr := domainmoney.FromMajorFloat(quote.TotalAmount, quote.Currency)
			expectedTotalMoney, expectedTotalErr := domainmoney.FromMajorFloat(*options.ExpectedTotal, quote.Currency)
			if quoteTotalErr != nil || expectedTotalErr != nil || quoteTotalMoney.AmountMinor() != expectedTotalMoney.AmountMinor() {
				return ErrOrderTotalChanged
			}
		}
		quoteTotalMoney, quoteTotalErr := domainmoney.FromMajorFloat(quote.TotalAmount, quote.Currency)
		if quoteTotalErr != nil {
			return fmt.Errorf("parse checkout total amount: %w", quoteTotalErr)
		}
		orderCurrency := quote.Currency
		paymentCurrency := quote.PaymentCurrency
		paymentAmount := quote.PaymentAmount
		if paymentCurrency == "" {
			paymentCurrency = orderCurrency
		}
		if paymentAmount <= 0 && quote.TotalAmount > 0 {
			paymentAmount = quote.TotalAmount
		}
		orderMoneyFields, moneyErr := buildOrderMoneyFields(quote, orderCurrency, paymentCurrency, paymentAmount)
		if moneyErr != nil {
			return fmt.Errorf("persist order money snapshot: %w", moneyErr)
		}
		if provider := paymentpkg.ProviderForPaymentMethod(paymentMethod); provider != "" && quote.TotalAmount > 0 {
			if err := paymentpkg.ValidateGatewayCurrency(paymentpkg.GatewayType(provider), paymentCurrency); err != nil {
				return fmt.Errorf(
					"payment method %s cannot process order currency %s: %w",
					strings.TrimSpace(paymentMethod),
					paymentCurrency,
					err,
				)
			}
		}

		shippingMethodSnapshot := strings.TrimSpace(shippingMethod)
		var carrierID *uint
		var carrierServiceID *uint
		var shippingQuoteID string
		var shippingQuotePlanID string
		shippingPlanSnapshot := datatypes.JSON([]byte(`{}`))
		if quote.ShippingQuote != nil && quote.ShippingQuote.SelectedPlan != nil {
			selectedPlan := quote.ShippingQuote.SelectedPlan
			shippingQuoteID = quote.ShippingQuote.ID
			shippingQuotePlanID = selectedPlan.ID
			if len(selectedPlan.Legs) == 1 {
				leg := selectedPlan.Legs[0]
				if leg.CarrierID > 0 {
					carrierID = uintPtr(leg.CarrierID)
				}
				if leg.CarrierServiceID > 0 {
					carrierServiceID = uintPtr(leg.CarrierServiceID)
				}
			}
			encodedPlan, err := json.Marshal(selectedPlan)
			if err != nil {
				return fmt.Errorf("encode selected shipping plan snapshot: %w", err)
			}
			shippingPlanSnapshot = datatypes.JSON(encodedPlan)
			if label := shippingQuotePlanSnapshot(*selectedPlan); label != "" {
				shippingMethodSnapshot = label
			}
		}
		if shippingMethodSnapshot == "" {
			shippingMethodSnapshot = "standard"
		}
		fulfillmentMode := order.ResolveFulfillmentMode(quote.Items)
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
			OrderNumber:              orderNumber,
			UserID:                   userID,
			Status:                   orderStatus,
			PaymentMethod:            paymentMethod,
			PaymentStatus:            paymentStatus,
			PaymentCurrency:          paymentCurrency,
			PaymentAmount:            paymentAmount,
			PaymentAmountMinor:       orderMoneyFields.PaymentAmountMinor,
			ShippingMethod:           shippingMethodSnapshot,
			ShippingStatus:           "pending",
			FulfillmentMode:          fulfillmentMode,
			ProductionStatus:         order.DefaultProductionStatus(fulfillmentMode),
			SignatureRequired:        order.ResolveSignatureRequired(quoteTotalMoney, quote.FXSnapshot),
			CarrierID:                carrierID,
			CarrierServiceID:         carrierServiceID,
			ShippingQuoteID:          shippingQuoteID,
			ShippingQuotePlanID:      shippingQuotePlanID,
			ShippingPlanSnapshotData: shippingPlanSnapshot,
			SubtotalAmount:           quote.SubtotalAmount,
			SubtotalAmountMinor:      orderMoneyFields.SubtotalAmountMinor,
			TotalAmount:              quote.TotalAmount,
			TotalAmountMinor:         orderMoneyFields.TotalAmountMinor,
			ShippingFee:              quote.ShippingFee,
			ShippingFeeMinor:         orderMoneyFields.ShippingFeeMinor,
			TaxAmount:                quote.TaxAmount,
			TaxAmountMinor:           orderMoneyFields.TaxAmountMinor,
			DiscountAmount:           quote.DiscountAmount,
			DiscountAmountMinor:      orderMoneyFields.DiscountAmountMinor,
			Currency:                 orderCurrency,
			CouponCode:               quote.CouponCode,
			PointsUsed:               quote.PointsToUse,
			PointsValue:              quote.PointsDiscount,
			PointsValueMinor:         orderMoneyFields.PointsValueMinor,
			FXSnapshotData:           currency.OrderFXSnapshotJSON(quote.FXSnapshot),
			PricingSnapshotData:      orderPricingSnapshot,
			Items:                    quote.Items,
			ShippingAddress:          shippingAddress,
			BillingAddress:           billingAddress,
			PaidAt:                   paidAt,
		}
		if options.CheckoutCartID > 0 {
			checkoutCartID := options.CheckoutCartID
			o.CheckoutCartID = &checkoutCartID
		}

		variantItemsMap := make(map[uint]int)
		for _, item := range quote.Items {
			if item.VariantID == nil {
				return fmt.Errorf("[CRITICAL] Missing variant for product ID %d", item.ProductID)
			}
			if productdomain.NormalizeFulfillmentMode(item.FulfillmentMode) == productdomain.FulfillmentModeStock {
				variantItemsMap[*item.VariantID] += item.Quantity
			}
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
		snapshot, err := s.orderEvidenceSnapshot.CreateForOrder(repos, o)
		if err != nil {
			return fmt.Errorf("[CRITICAL] Failed to create order evidence snapshot: %w", err)
		}
		if _, err := s.orderEvidence.CreateInitialPackage(repos, snapshot); err != nil {
			return fmt.Errorf("[CRITICAL] Failed to create initial order evidence package: %w", err)
		}
		if quote.GiftCard != nil && quote.GiftCardDiscountCents > 0 {
			if err := consumeGiftCardForOrder(repos.Coupon, quote.GiftCard, o, quote.GiftCardDiscountCents); err != nil {
				return fmt.Errorf("[CRITICAL] Failed to consume gift card for order ID %d: %w", o.ID, err)
			}
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
				Email:     coupon.NormalizeEmail(shippingAddress.Email),
				OrderID:   o.ID,
				Discount:  quote.CouponDiscount,
				CreatedAt: time.Now(),
			}
			if err := repos.Coupon.CreateCouponUsage(usage); err != nil {
				return fmt.Errorf("[CRITICAL] Failed to record coupon usage for coupon ID %d: %w", quote.Coupon.ID, err)
			}
		}

		if options.CheckoutCartID > 0 {
			deletedItemCount, err := repos.Cart.ConsumeLockedCheckoutCartItemRows(options.CheckoutCartID, lockedCheckoutCartItemIDs)
			if err != nil {
				return fmt.Errorf("consume checkout cart %d: %w", options.CheckoutCartID, err)
			}
			if deletedItemCount != int64(len(lockedCheckoutCartItemIDs)) {
				return ErrCheckoutCartConsumptionConflict
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

func consumeGiftCardForOrder(
	couponRepo *repository.CouponRepository,
	giftCard *coupon.GiftCard,
	o *order.Order,
	amountCents int64,
) error {
	if couponRepo == nil {
		return errors.New("gift card repository is not configured")
	}
	if giftCard == nil || o == nil {
		return errors.New("gift card and order are required")
	}
	if amountCents <= 0 {
		return nil
	}
	if !giftCard.IsValid() {
		return repository.ErrGiftCardInsufficientBalance
	}

	balance, err := giftCard.BalanceMoney()
	if err != nil {
		return fmt.Errorf("invalid gift card money: %w", err)
	}
	debit, err := domainmoney.New(amountCents, giftCard.Currency)
	if err != nil {
		return fmt.Errorf("invalid gift card debit: %w", err)
	}
	remaining, err := balance.Subtract(debit)
	if err != nil || remaining.AmountMinor() < 0 {
		return repository.ErrGiftCardInsufficientBalance
	}
	giftCard.BalanceCents = remaining.AmountMinor()
	if giftCard.BalanceCents == 0 {
		giftCard.Status = "used"
	}
	if err := couponRepo.UpdateGiftCard(giftCard); err != nil {
		return err
	}

	return couponRepo.CreateGiftCardTransaction(&coupon.GiftCardTransaction{
		GiftCardID:   giftCard.ID,
		Currency:     giftCard.Currency,
		OrderID:      o.ID,
		Type:         "use",
		AmountCents:  -amountCents,
		BalanceCents: giftCard.BalanceCents,
		Note:         fmt.Sprintf("Gift card applied to order #%s", o.OrderNumber),
	})
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
		AmountMinor:     0,
		Currency:        orderCurrency,
		Status:          "completed",
		GatewayResponse: `{"settlement":"zero_total","reason":"discounts_cover_total"}`,
		CompletedAt:     &settledAt,
	}
	if _, err := repos.Payment.CreateTransactionIfAbsent(transaction); err != nil {
		return err
	}
	settlement, err := domainmoney.New(0, orderCurrency)
	if err != nil {
		return err
	}
	return enqueueOrderPaidOutboxEvent(repos.Outbox, o, VerifiedGatewayPaymentInput{
		Provider:      zeroTotalSettlementPaymentMethod,
		OrderNumber:   o.OrderNumber,
		TransactionID: transactionID,
		PaymentMethod: zeroTotalSettlementPaymentMethod,
		Amount:        settlement,
	}, settledAt)
}

func zeroTotalOrderTransactionID(orderID uint) string {
	return fmt.Sprintf("zero-total-%d", orderID)
}

func shippingQuotePlanSnapshot(plan ShippingQuotePlan) string {
	labels := make([]string, 0, len(plan.Legs))
	for _, leg := range plan.Legs {
		if label := shippingQuoteLegSnapshot(leg); label != "" {
			labels = append(labels, label)
		}
	}
	return strings.Join(labels, " + ")
}

func shippingQuoteLegSnapshot(option ShippingQuoteLeg) string {
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
