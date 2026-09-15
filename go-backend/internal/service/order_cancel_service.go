package service

import (
	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/repository"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func (s *OrderService) CancelOrder(id uint, userID uint) error {
	o, err := s.orderRepo.FindByID(id)
	if err != nil {
		return normalizeOrderError(err)
	}

	if o.UserID != userID {
		return errors.New("unauthorized")
	}

	if err := validateOrderCancellation(o); err != nil {
		return err
	}

	return s.cancelOrderWithRollback(o)
}

func (s *OrderService) CancelOrderByNumber(orderNumber string, userID uint) error {
	if !s.validatesKnownProtectedOrderNumber(orderNumber) {
		return ErrOrderNotFound
	}
	o, err := s.orderRepo.FindByOrderNumber(orderNumber)
	if err != nil {
		return normalizeOrderError(err)
	}

	if o.UserID != userID {
		return errors.New("unauthorized")
	}
	if err := validateOrderCancellation(o); err != nil {
		return err
	}

	return s.cancelOrderWithRollback(o)
}

func (s *OrderService) cancelOrderWithRollback(o *order.Order) error {
	var affectedProductIDs []uint
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		lockedOrder, err := repos.Order.FindByIDForUpdateWithItems(o.ID)
		if err != nil {
			return normalizeOrderError(err)
		}
		if err := validateOrderCancellation(lockedOrder); err != nil {
			return err
		}

		cancelled, err := repos.Order.MarkCancelledIfPendingUnpaid(lockedOrder.ID, time.Now().UTC())
		if err != nil {
			return err
		}
		if !cancelled {
			return ErrOrderCancellationConflict
		}

		productIDs, err := rollbackOrderReservationsInTx(repos, lockedOrder, "cancelled")
		if err != nil {
			return err
		}
		affectedProductIDs = append(affectedProductIDs, productIDs...)
		if err := s.enqueueProductCacheInvalidationInTx(repos, productIDs, "order stock restored cancelled"); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.invalidateProductCacheAfterStockCommit(affectedProductIDs)
	return nil
}

func validateOrderCancellation(o *order.Order) error {
	if o == nil {
		return ErrOrderNotFound
	}
	if orderRequiresProduction(o) {
		switch order.NormalizeProductionStatus(o.ProductionStatus) {
		case order.ProductionStatusStarted, order.ProductionStatusCompleted:
			return ErrProductionStartedCancellationNotAllowed
		}
	}
	if o.Status == "paid" || o.PaymentStatus == "paid" {
		return ErrPaidOrderCancellationNotAllowed
	}
	if o.Status == "cancelled" {
		return ErrOrderCancellationConflict
	}
	if o.Status != "pending" || o.PaymentStatus != "unpaid" {
		return errors.New("order cannot be cancelled")
	}
	return nil
}

func rollbackOrderReservationsInTx(repos repository.TxRepositories, o *order.Order, reason string) ([]uint, error) {
	var affectedProductIDs []uint
	for _, item := range o.Items {
		if item.VariantID == nil {
			return nil, fmt.Errorf("[CRITICAL] Missing variant for order item %d", item.ID)
		}
		if productdomain.NormalizeFulfillmentMode(item.FulfillmentMode) != productdomain.FulfillmentModeStock {
			continue
		}
		productIDs, err := repos.Product.IncrementVariantStock(*item.VariantID, item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("[CRITICAL] Failed to restore stock for variant %d: %w", *item.VariantID, err)
		}
		affectedProductIDs = append(affectedProductIDs, productIDs...)
	}

	if o.PointsUsed > 0 {
		_, err := repos.Loyalty.AdjustUserPointsInCurrentTx(
			o.UserID,
			o.PointsUsed,
			"refund",
			"order",
			o.ID,
			fmt.Sprintf("Order #%s %s points refund", o.OrderNumber, reason),
		)
		if err != nil {
			return nil, fmt.Errorf("[CRITICAL] Failed to refund points: %w", err)
		}
	}

	if o.CouponCode != "" {
		cp, err := repos.Coupon.FindCouponByCode(o.CouponCode)
		if err != nil {
			return nil, fmt.Errorf("[CRITICAL] Failed to find coupon during refund: %w", err)
		}
		if cp != nil {
			if err := repos.Coupon.DecrementUsedCount(cp.ID); err != nil {
				return nil, fmt.Errorf("[CRITICAL] Failed to restore coupon usage limit: %w", err)
			}

			if err := repos.Coupon.ReverseCouponUsageByOrderID(o.ID, time.Now().UTC(), reason); err != nil {
				return nil, fmt.Errorf("[CRITICAL] Failed to reverse coupon usage log: %w", err)
			}
		}
	}

	if err := restoreGiftCardUsageForOrderInTx(repos.Coupon, o.ID, reason); err != nil {
		return nil, err
	}

	if o.CheckoutCartID != nil {
		// Cart restoration is deliberately best-effort. Inventory, payment and
		// coupon rollback must still commit when a customer cart was deleted or
		// the cart subsystem is temporarily unavailable.
		if repos.Cart == nil {
			logger.Error("checkout cart restoration skipped: repository unavailable",
				zap.Uint("order_id", o.ID), zap.Uint("cart_id", *o.CheckoutCartID), zap.String("reason", reason))
		} else if _, err := repos.Cart.FindAuthenticatedUserCartByIDForUpdate(*o.CheckoutCartID, o.UserID); err != nil {
			logger.Error("checkout cart restoration skipped: cart unavailable",
				zap.Uint("order_id", o.ID), zap.Uint("cart_id", *o.CheckoutCartID), zap.String("reason", reason), zap.Error(err))
		} else if err := repos.Cart.RestoreConsumedOrderItemsToOriginalCheckoutCart(*o.CheckoutCartID, o.Items, o.Currency); err != nil {
			logger.Error("checkout cart restoration failed",
				zap.Uint("order_id", o.ID), zap.Uint("cart_id", *o.CheckoutCartID), zap.String("reason", reason), zap.Error(err))
		}
	}

	return affectedProductIDs, nil
}

func restoreGiftCardUsageForOrderInTx(
	couponRepo *repository.CouponRepository,
	orderID uint,
	reason string,
) error {
	return restoreGiftCardAmountForOrderInTx(couponRepo, orderID, 0, nil, reason)
}

func sumGiftCardUsageForOrderInTx(
	couponRepo *repository.CouponRepository,
	orderID uint,
) (float64, error) {
	if couponRepo == nil {
		return 0, errors.New("gift card transaction repository is not configured")
	}
	transactions, err := couponRepo.FindGiftCardTransactionByOrderID(orderID)
	if err != nil {
		return 0, fmt.Errorf("[CRITICAL] Failed to load gift card usage for order %d: %w", orderID, err)
	}

	var usedCents int64
	for _, transaction := range transactions {
		if transaction.Type != "use" {
			continue
		}
		amountCents := transaction.AmountCents
		if amountCents < 0 {
			amountCents = -amountCents
		}
		usedCents += amountCents
	}
	if len(transactions) > 0 {
		usedMoney, err := domainmoney.New(usedCents, transactions[0].Currency)
		if err != nil {
			return 0, err
		}
		amount, err := usedMoney.MajorFloat()
		if err != nil {
			return 0, err
		}
		return amount, nil
	}
	return 0, nil
}

func restoreGiftCardRefundInTx(
	couponRepo *repository.CouponRepository,
	orderID uint,
	refundID uint,
	amount float64,
	reason string,
) error {
	if amount <= 0 {
		return nil
	}
	transactions, err := couponRepo.FindGiftCardTransactionByOrderID(orderID)
	if err != nil {
		return err
	}
	cardCurrency := currency.DefaultPrimaryCurrency
	if len(transactions) > 0 && transactions[0].Currency != "" {
		cardCurrency = transactions[0].Currency
	}
	amountMoney, err := domainmoney.FromMajorFloat(amount, cardCurrency)
	if err != nil {
		return err
	}
	return restoreGiftCardAmountForOrderInTx(
		couponRepo,
		orderID,
		amountMoney.AmountMinor(),
		&refundID,
		reason,
	)
}

func restoreGiftCardAmountForOrderInTx(
	couponRepo *repository.CouponRepository,
	orderID uint,
	targetCents int64,
	refundID *uint,
	reason string,
) error {
	if couponRepo == nil {
		return errors.New("gift card transaction repository is not configured")
	}

	transactions, err := couponRepo.FindGiftCardTransactionByOrderID(orderID)
	if err != nil {
		return fmt.Errorf("[CRITICAL] Failed to load gift card usage for order %d: %w", orderID, err)
	}
	usedByCard := make(map[uint]int64)
	refundedByCard := make(map[uint]int64)
	cardOrder := make([]uint, 0)
	seenCards := make(map[uint]struct{})
	linkedRefundCents := int64(0)
	for _, transaction := range transactions {
		switch transaction.Type {
		case "use":
			if _, seen := seenCards[transaction.GiftCardID]; !seen {
				cardOrder = append(cardOrder, transaction.GiftCardID)
				seenCards[transaction.GiftCardID] = struct{}{}
			}
			usedCents := transaction.AmountCents
			if usedCents < 0 {
				usedCents = -usedCents
			}
			usedByCard[transaction.GiftCardID] += usedCents
		case "refund":
			refundedCents := transaction.AmountCents
			if refundedCents < 0 {
				refundedCents = -refundedCents
			}
			if refundID != nil && transaction.RefundID != nil && *transaction.RefundID == *refundID {
				linkedRefundCents += refundedCents
				continue
			}
			refundedByCard[transaction.GiftCardID] += refundedCents
		}
	}

	if targetCents <= 0 {
		for _, cardID := range cardOrder {
			availableCents := usedByCard[cardID] - refundedByCard[cardID]
			if availableCents > 0 {
				targetCents += availableCents
			}
		}
	}
	if refundID != nil {
		targetCents -= linkedRefundCents
		if targetCents <= 0 {
			return nil
		}
	}

	if targetCents <= 0 {
		return nil
	}

	for _, cardID := range cardOrder {
		availableCents := usedByCard[cardID] - refundedByCard[cardID]
		if availableCents <= 0 {
			continue
		}
		restoreCents := availableCents
		if restoreCents > targetCents {
			restoreCents = targetCents
		}

		card, err := couponRepo.FindGiftCardByIDForUpdate(cardID)
		if err != nil {
			return fmt.Errorf("[CRITICAL] Failed to lock gift card %d during order %d rollback: %w", cardID, orderID, err)
		}
		balance, err := card.BalanceMoney()
		if err != nil {
			return fmt.Errorf("[CRITICAL] Invalid gift card money for card %d: %w", card.ID, err)
		}
		credit, err := domainmoney.New(restoreCents, card.Currency)
		if err != nil {
			return fmt.Errorf("[CRITICAL] Invalid gift card restoration for card %d: %w", card.ID, err)
		}
		updatedBalance, err := balance.Add(credit)
		if err != nil {
			return fmt.Errorf("[CRITICAL] Failed to calculate gift card %d restoration: %w", card.ID, err)
		}
		card.BalanceCents = updatedBalance.AmountMinor()
		if card.Status == "used" && card.BalanceCents > 0 {
			card.Status = "active"
		}
		if err := couponRepo.UpdateGiftCard(card); err != nil {
			return fmt.Errorf("[CRITICAL] Failed to restore gift card %d during order %d rollback: %w", card.ID, orderID, err)
		}
		restoration := &coupon.GiftCardTransaction{
			GiftCardID:   card.ID,
			Currency:     card.Currency,
			OrderID:      orderID,
			Type:         "refund",
			AmountCents:  restoreCents,
			BalanceCents: card.BalanceCents,
			Note:         fmt.Sprintf("Gift card restored after order rollback: %s", reason),
		}
		if refundID != nil {
			restoration.RefundID = refundID
			restoration.Note = fmt.Sprintf("Gift card restored after refund #%d: %s", *refundID, reason)
		}
		if err := couponRepo.CreateGiftCardTransaction(restoration); err != nil {
			return fmt.Errorf("[CRITICAL] Failed to record gift card restoration for order %d: %w", orderID, err)
		}
		targetCents -= restoreCents
		if targetCents <= 0 {
			return nil
		}
	}

	return fmt.Errorf(
		"[CRITICAL] Gift card restoration amount %d cents exceeds available usage for order %d",
		targetCents,
		orderID,
	)
}
