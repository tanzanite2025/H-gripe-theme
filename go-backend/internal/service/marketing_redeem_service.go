package service

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/repository"
	"time"
)

type RedeemResult struct {
	GiftCardID      uint       `json:"giftcard_id"`
	CardCode        string     `json:"card_code"`
	Balance         float64    `json:"balance"`
	PointsSpent     int        `json:"points_spent"`
	PointsRemaining int        `json:"points_remaining"`
	ExpiresAt       *time.Time `json:"expires_at"`
}

type RedeemGiftCardOption struct {
	ID             int     `json:"id"`
	Label          string  `json:"label"`
	GiftCardValue  float64 `json:"giftcard_value"`
	PointsRequired int     `json:"points_required"`
	Status         string  `json:"status"`
}

func (s *MarketingService) ListRedeemGiftCardOptions(redeemCfg *setting.RedeemSettings) []RedeemGiftCardOption {
	if redeemCfg == nil || !redeemCfg.Enabled || redeemCfg.ExchangeRate <= 0 {
		return []RedeemGiftCardOption{}
	}

	options := make([]RedeemGiftCardOption, 0, len(redeemCfg.PresetValues))
	for idx, value := range redeemCfg.PresetValues {
		if value <= 0 {
			continue
		}
		pointsRequired := int(math.Round(value * float64(redeemCfg.ExchangeRate)))
		if pointsRequired < redeemCfg.MinPoints {
			continue
		}
		options = append(options, RedeemGiftCardOption{
			ID:             idx + 1,
			Label:          fmt.Sprintf("$%.2f Gift Card", value),
			GiftCardValue:  value,
			PointsRequired: pointsRequired,
			Status:         "active",
		})
	}

	return options
}

func (s *MarketingService) RedeemPointsForGiftCard(
	userID uint,
	giftCardValue float64,
	redeemCfg *setting.RedeemSettings,
) (*RedeemResult, error) {
	if redeemCfg == nil {
		return nil, errors.New("[CRITICAL] Redeem settings are required")
	}
	if !redeemCfg.Enabled {
		return nil, errors.New("[CRITICAL] Point redemption is disabled")
	}
	if redeemCfg.ExchangeRate <= 0 {
		return nil, errors.New("[CRITICAL] Redeem exchange rate must be greater than zero")
	}
	if giftCardValue <= 0 {
		return nil, errors.New("[CRITICAL] Gift card value must be greater than zero")
	}
	if !isAllowedRedeemValue(giftCardValue, redeemCfg.PresetValues) {
		return nil, fmt.Errorf("[CRITICAL] Gift card value %.2f is not an allowed redeem preset", giftCardValue)
	}

	pointsToSpend := int(math.Round(giftCardValue * float64(redeemCfg.ExchangeRate)))

	if pointsToSpend < redeemCfg.MinPoints {
		return nil, fmt.Errorf("[CRITICAL] Minimum points required to redeem is %d", redeemCfg.MinPoints)
	}

	var giftcard coupon.GiftCard
	var transaction *loyalty.LoyaltyTransaction

	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		userLoyalty, err := repos.Loyalty.FindOrCreateUserLoyaltyForUpdate(userID)
		if err != nil {
			return fmt.Errorf("[CRITICAL] Failed to retrieve user loyalty data: %v", err)
		}

		if userLoyalty.AvailablePoints < pointsToSpend {
			return fmt.Errorf("[CRITICAL] Insufficient points: available %d, required %d", userLoyalty.AvailablePoints, pointsToSpend)
		}

		todayStart := time.Now().Truncate(24 * time.Hour)
		todayEnd := todayStart.Add(24 * time.Hour)

		sumPoints, err := repos.Loyalty.SumTransactionPointsByUser(userID, "spend", "giftcard", todayStart, todayEnd)
		if err != nil {
			return fmt.Errorf("[CRITICAL] Failed to verify daily limit: %v", err)
		}

		todayRedeemedValue := math.Abs(float64(sumPoints)) / float64(redeemCfg.ExchangeRate)
		if redeemCfg.MaxValuePerDay > 0 && todayRedeemedValue+giftCardValue > redeemCfg.MaxValuePerDay {
			return fmt.Errorf("[CRITICAL] Daily limit exceeded. Limit: %.2f, Redeemed: %.2f, Attempted: %.2f", redeemCfg.MaxValuePerDay, todayRedeemedValue, giftCardValue)
		}

		cardCode := "REDEEM-" + generateRedeemCode(12)
		var expiresAt *time.Time
		if redeemCfg.CardExpiryDays > 0 {
			t := time.Now().AddDate(0, 0, redeemCfg.CardExpiryDays)
			expiresAt = &t
		}

		giftcard = coupon.GiftCard{
			Code:         cardCode,
			InitialValue: giftCardValue,
			Balance:      giftCardValue,
			Currency:     "USD",
			Status:       "active",
			ExpiresAt:    expiresAt,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := repos.Coupon.CreateGiftCard(&giftcard); err != nil {
			return fmt.Errorf("[CRITICAL] Failed to create gift card: %v", err)
		}

		transaction, err = repos.Loyalty.AdjustUserPointsInCurrentTx(
			userID,
			-pointsToSpend,
			"spend",
			"giftcard",
			giftcard.ID,
			fmt.Sprintf("Redeemed gift card %s with %d points", cardCode, pointsToSpend),
		)
		if err != nil {
			return fmt.Errorf("[CRITICAL] Failed to deduct points: %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RedeemResult{
		GiftCardID:      giftcard.ID,
		CardCode:        giftcard.Code,
		Balance:         giftcard.Balance,
		PointsSpent:     pointsToSpend,
		PointsRemaining: transaction.Balance,
		ExpiresAt:       giftcard.ExpiresAt,
	}, nil
}

func isAllowedRedeemValue(value float64, presets []float64) bool {
	for _, preset := range presets {
		if math.Abs(value-preset) < 0.000001 {
			return true
		}
	}
	return false
}

func generateRedeemCode(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
