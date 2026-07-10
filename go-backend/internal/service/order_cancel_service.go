package service

import (
	"errors"
	"fmt"
	"tanzanite/internal/domain/order"
	"tanzanite/internal/repository"
)

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
			if item.VariantID == nil {
				return fmt.Errorf("[CRITICAL] Missing variant for order item %d", item.ID)
			}
			if err := repos.Product.IncrementVariantStock(*item.VariantID, item.Quantity); err != nil {
				return fmt.Errorf("[CRITICAL] Failed to restore stock for variant %d: %w", *item.VariantID, err)
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
