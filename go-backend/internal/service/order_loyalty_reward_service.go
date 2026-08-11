package service

import (
	"errors"
	"fmt"
	"math"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/repository"
)

func (s *OrderService) completeOrderWithLoyaltyReward(id uint) error {
	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		o, err := repos.Order.FindByIDForUpdateWithItems(id)
		if err != nil {
			return normalizeOrderError(err)
		}
		if !o.CanTransitionTo("completed") {
			return fmt.Errorf("invalid status transition from %s to completed", o.Status)
		}
		if err := repos.Order.UpdateStatus(id, "completed"); err != nil {
			return err
		}
		return s.awardOrderCompletionPoints(repos, o)
	})
}

func (s *OrderService) awardOrderCompletionPoints(repos repository.TxRepositories, o *order.Order) error {
	if repos.Loyalty == nil || o == nil || o.UserID == 0 {
		return nil
	}

	config, err := s.currentOrderRewardProgramConfig(repos)
	if err != nil {
		return err
	}
	if config == nil || !config.Enabled || config.PurchaseEarnPointsPerUnit <= 0 {
		return nil
	}

	existing, err := repos.Loyalty.CountTransactionsByUserTypeSourceAndSourceID(o.UserID, "earn", "order", o.ID)
	if err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}

	points, err := s.calculateOrderCompletionPoints(repos, o, config)
	if err != nil {
		return err
	}
	if points <= 0 {
		return nil
	}

	_, err = repos.Loyalty.AdjustUserPointsInCurrentTxWithConfig(
		o.UserID,
		points,
		"earn",
		"order",
		o.ID,
		fmt.Sprintf("Order #%s completion reward", o.OrderNumber),
		programConfigID(config),
	)
	return err
}

func (s *OrderService) currentOrderRewardProgramConfig(repos repository.TxRepositories) (*loyalty.ProgramConfig, error) {
	if repos.Program != nil {
		config, err := repos.Program.FindActive()
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return nil, nil
			}
			return nil, err
		}
		if err := validateProgramConfig(config); err != nil {
			return nil, err
		}
		return config, nil
	}

	if s == nil || s.checkout == nil {
		return nil, nil
	}
	config, err := s.checkout.currentLoyaltyProgramConfig()
	if err != nil {
		if errors.Is(err, ErrLoyaltyProgramConfigNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return config, nil
}

func (s *OrderService) calculateOrderCompletionPoints(
	_ repository.TxRepositories,
	o *order.Order,
	config *loyalty.ProgramConfig,
) (int, error) {
	eligibleAmount := o.SubtotalAmount - o.DiscountAmount
	if eligibleAmount <= 0 || config == nil || config.PurchaseEarnPointsPerUnit <= 0 {
		return 0, nil
	}
	if currency.NormalizeCode(o.Currency) != LoyaltyPointsBaseCurrency {
		return 0, fmt.Errorf("%w: order completion points require %s order amounts, got %s", ErrInvalidLoyaltyProgramConfig, LoyaltyPointsBaseCurrency, o.Currency)
	}

	basePoints := int(math.Floor(eligibleAmount * float64(config.PurchaseEarnPointsPerUnit)))
	if basePoints <= 0 {
		return 0, nil
	}

	return basePoints, nil
}
