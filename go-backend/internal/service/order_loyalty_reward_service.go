package service

import (
	"errors"
	"fmt"
	"math/big"
	"time"

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
		if err := repos.Order.UpdateStatus(id, o.Status, "completed"); err != nil {
			return err
		}
		completedAt := time.Now().UTC()
		return enqueueOrderCompletedOutboxEvent(repos.Outbox, o, completedAt)
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
	if o == nil {
		return 0, nil
	}
	subtotalMoney, err := o.SubtotalMoney()
	if err != nil {
		return 0, nil
	}
	discountMoney, err := o.DiscountMoney()
	if err != nil {
		return 0, nil
	}
	eligibleMoney, err := subtotalMoney.Subtract(discountMoney)
	if err != nil || eligibleMoney.AmountMinor() <= 0 || config == nil || config.PurchaseEarnPointsPerUnit <= 0 {
		return 0, nil
	}

	rewardableBaseMoney := eligibleMoney
	if currency.NormalizeCode(o.Currency) != LoyaltyPointsBaseCurrency {
		// New orders carry an immutable rate captured at checkout. Legacy
		// non-USD orders without a usable snapshot remain completable, but
		// cannot receive a reward without guessing a historical rate.
		snapshot, err := currency.ParseOrderFXSnapshot(o.FXSnapshotData)
		if err != nil ||
			currency.NormalizeCode(snapshot.BaseCurrency) != LoyaltyPointsBaseCurrency ||
			currency.NormalizeCode(snapshot.OrderCurrency) != currency.NormalizeCode(o.Currency) {
			return 0, nil
		}
		rewardableBaseMoney, err = order.OrderAmountToBaseMoney(eligibleMoney, snapshot)
		if err != nil {
			return 0, nil
		}
	}

	// PurchaseEarnPointsPerUnit is points per whole USD unit. Keep the
	// calculation in minor units so fractional minor cannot create points.
	pointsNumerator := new(big.Int).Mul(
		big.NewInt(rewardableBaseMoney.AmountMinor()),
		big.NewInt(int64(config.PurchaseEarnPointsPerUnit)),
	)
	basePointsBig := new(big.Int).Quo(pointsNumerator, big.NewInt(100))
	if !basePointsBig.IsInt64() {
		return 0, errors.New("order loyalty points overflow")
	}
	basePoints64 := basePointsBig.Int64()
	basePoints := int(basePoints64)
	if int64(basePoints) != basePoints64 || basePoints <= 0 {
		return 0, nil
	}

	return basePoints, nil
}
