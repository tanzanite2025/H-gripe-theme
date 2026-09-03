package service

import (
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/repository"
	"errors"
	"fmt"
	"time"
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

	return affectedProductIDs, nil
}
