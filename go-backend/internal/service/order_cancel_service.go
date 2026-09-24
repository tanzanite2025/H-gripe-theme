package service

import (
	"commerce-platform/internal/domain/order"
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/repository"
	"encoding/json"
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

		cancelledAt := time.Now().UTC()
		cancelled, err := repos.Order.MarkCancelledIfPendingUnpaid(lockedOrder.ID, cancelledAt)
		if err != nil {
			return err
		}
		if !cancelled {
			return ErrOrderCancellationConflict
		}
		if err := enqueueOrderCancelledDomainEvent(repos.Outbox, lockedOrder, cancelledAt); err != nil {
			return err
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
	neverCancel, _, err := orderConfigurationPolicies(o)
	if err != nil {
		return fmt.Errorf("cannot validate configured option cancellation policy: %w", err)
	}
	if neverCancel {
		return ErrConfiguredOptionCancellationNotAllowed
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
	variantItemsMap := make(map[uint]int)
	for _, item := range o.Items {
		if item.VariantID == nil {
			return nil, fmt.Errorf("[CRITICAL] Missing variant for order item %d", item.ID)
		}
		if productdomain.NormalizeFulfillmentMode(item.FulfillmentMode) == productdomain.FulfillmentModeStock {
			variantItemsMap[*item.VariantID] += item.Quantity
		}
		if len(item.ConfigurationSnapshotData) > 0 && string(item.ConfigurationSnapshotData) != "{}" {
			var configuration ProductConfigurationSnapshot
			if err := json.Unmarshal(item.ConfigurationSnapshotData, &configuration); err != nil {
				return nil, fmt.Errorf("[CRITICAL] Failed to parse configuration snapshot for order item %d: %w", item.ID, err)
			}
			for _, allocation := range configuration.InventoryAllocations {
				if allocation.VariantID == 0 || allocation.Quantity <= 0 {
					return nil, fmt.Errorf("[CRITICAL] Invalid component inventory allocation for order item %d", item.ID)
				}
				variantItemsMap[allocation.VariantID] += allocation.Quantity * item.Quantity
			}
		}
	}
	if len(variantItemsMap) > 0 {
		productIDs, err := repos.Product.IncrementVariantStocks(variantItemsMap)
		if err != nil {
			return nil, fmt.Errorf("[CRITICAL] Failed to restore reserved stock: %w", err)
		}
		affectedProductIDs = append(affectedProductIDs, productIDs...)
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
