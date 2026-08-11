package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/repository"
)

type RedeemGiftCardRequest struct {
	OptionID           uint
	GiftCardValueCents int64
	IdempotencyKey     string
}

type RedeemResult struct {
	RedemptionID       uint       `json:"redemption_id"`
	GiftCardID         uint       `json:"giftcard_id"`
	CardCode           string     `json:"card_code"`
	Balance            float64    `json:"balance"`
	BalanceCents       int64      `json:"balance_cents"`
	GiftCardValueCents int64      `json:"giftcard_value_cents"`
	PointsSpent        int        `json:"points_spent"`
	PointsRemaining    int        `json:"points_remaining"`
	ExpiresAt          *time.Time `json:"expires_at"`
}

type RedeemGiftCardOption struct {
	ID                 uint    `json:"id"`
	Label              string  `json:"label"`
	GiftCardValue      float64 `json:"giftcard_value"`
	GiftCardValueCents int64   `json:"giftcard_value_cents"`
	Currency           string  `json:"currency"`
	PointsRequired     int     `json:"points_required"`
	RemainingQuantity  int64   `json:"remaining_quantity"`
	Status             string  `json:"status"`
}

func (s *MarketingService) ListRedeemGiftCardOptionsFromConfig(config *loyalty.ProgramConfig) []RedeemGiftCardOption {
	if config == nil || !config.Enabled {
		return []RedeemGiftCardOption{}
	}

	options := make([]RedeemGiftCardOption, 0, len(config.RedeemOptions))
	for _, option := range config.RedeemOptions {
		pointsRequired, err := PointsForGiftCardValue(option.ValueCents, config.ExchangeRatePoints)
		if err != nil || pointsRequired < config.MinRedeemPoints || option.RemainingQuantity() <= 0 {
			continue
		}
		value := float64(option.ValueCents) / 100
		options = append(options, RedeemGiftCardOption{
			ID:                 option.ID,
			Label:              fmt.Sprintf("%s %.2f Gift Card", option.Currency, value),
			GiftCardValue:      value,
			GiftCardValueCents: option.ValueCents,
			Currency:           option.Currency,
			PointsRequired:     pointsRequired,
			RemainingQuantity:  option.RemainingQuantity(),
			Status:             "active",
		})
	}
	return options
}

func (s *MarketingService) RedeemPointsForGiftCard(
	userID uint,
	req RedeemGiftCardRequest,
	config *loyalty.ProgramConfig,
) (*RedeemResult, error) {
	if config == nil {
		return nil, ErrLoyaltyProgramConfigNotFound
	}
	if !config.Enabled {
		return nil, errors.New("point redemption is disabled")
	}
	if req.IdempotencyKey == "" {
		return nil, errors.New("idempotency key is required")
	}
	if s.txManager == nil {
		return nil, errors.New("redemption transaction manager is unavailable")
	}

	selectedOption, err := resolveRedeemOption(req, config)
	if err != nil {
		return nil, err
	}
	valueCents := selectedOption.ValueCents
	pointsToSpend, err := PointsForGiftCardValue(valueCents, config.ExchangeRatePoints)
	if err != nil {
		return nil, err
	}
	if pointsToSpend < config.MinRedeemPoints {
		return nil, fmt.Errorf("minimum points required to redeem is %d", config.MinRedeemPoints)
	}

	var result *RedeemResult
	err = s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.Redemption == nil {
			return errors.New("gift card redemption repository is unavailable")
		}

		existing, findErr := repos.Redemption.FindByUserAndIdempotencyKey(userID, req.IdempotencyKey)
		if findErr == nil {
			result, err = s.redeemResultFromExisting(repos.Loyalty, existing)
			return err
		}
		if !repository.IsRecordNotFound(findErr) {
			return findErr
		}

		userLoyalty, err := repos.Loyalty.FindOrCreateUserLoyaltyForUpdate(userID)
		if err != nil {
			return fmt.Errorf("failed to retrieve user loyalty data: %w", err)
		}
		if userLoyalty.AvailablePoints < pointsToSpend {
			return fmt.Errorf("insufficient points: available %d, required %d", userLoyalty.AvailablePoints, pointsToSpend)
		}

		existing, findErr = repos.Redemption.FindByUserAndIdempotencyKey(userID, req.IdempotencyKey)
		if findErr == nil {
			result, err = s.redeemResultFromExisting(repos.Loyalty, existing)
			return err
		}
		if !repository.IsRecordNotFound(findErr) {
			return findErr
		}

		start, end := localDayBounds(time.Now())
		todayValueCents, err := repos.Redemption.SumValueCentsByUser(userID, start, end)
		if err != nil {
			return fmt.Errorf("failed to verify daily redemption limit: %w", err)
		}
		if config.MaxValuePerDayCents > 0 && todayValueCents+valueCents > config.MaxValuePerDayCents {
			return fmt.Errorf(
				"daily limit exceeded: limit %.2f, redeemed %.2f, attempted %.2f",
				float64(config.MaxValuePerDayCents)/100,
				float64(todayValueCents)/100,
				float64(valueCents)/100,
			)
		}

		if repos.Program == nil {
			return errors.New("loyalty program repository is unavailable")
		}
		consumedOption, err := repos.Program.ConsumeRedeemOption(config.ID, selectedOption.ID)
		if err != nil {
			return fmt.Errorf("failed to reserve gift card stock: %w", err)
		}

		selectedOption = *consumedOption

		var expiresAt *time.Time
		if config.CardExpiryDays > 0 {
			expiry := time.Now().AddDate(0, 0, config.CardExpiryDays)
			expiresAt = &expiry
		}

		giftCard, err := createRedeemedGiftCard(repos.Coupon, userID, valueCents, selectedOption.Currency, expiresAt)
		if err != nil {
			return fmt.Errorf("failed to create gift card: %w", err)
		}

		configID := config.ID
		redemption := &coupon.GiftCardRedemption{
			UserID:             userID,
			GiftCardID:         giftCard.ID,
			ProgramConfigID:    configID,
			IdempotencyKey:     req.IdempotencyKey,
			Currency:           selectedOption.Currency,
			GiftCardValueCents: valueCents,
			PointsSpent:        pointsToSpend,
			Status:             "completed",
		}
		if err := repos.Redemption.Create(redemption); err != nil {
			return fmt.Errorf("failed to create gift card redemption: %w", err)
		}

		loyaltyTransaction, err := repos.Loyalty.AdjustUserPointsInCurrentTxWithConfig(
			userID,
			-pointsToSpend,
			"spend",
			"gift_card_redemption",
			redemption.ID,
			fmt.Sprintf("Redeemed gift card %s with %d points", giftCard.Code, pointsToSpend),
			&configID,
		)
		if err != nil {
			return fmt.Errorf("failed to deduct points: %w", err)
		}

		redemption.LoyaltyTransactionID = &loyaltyTransaction.ID
		if err := repos.Redemption.Update(redemption); err != nil {
			return fmt.Errorf("failed to link loyalty transaction: %w", err)
		}

		redemptionID := redemption.ID
		if err := repos.Coupon.CreateGiftCardTransaction(&coupon.GiftCardTransaction{
			GiftCardID:   giftCard.ID,
			RedemptionID: &redemptionID,
			Type:         "issue",
			AmountCents:  valueCents,
			BalanceCents: valueCents,
			Note:         "Gift card issued through loyalty redemption",
		}); err != nil {
			return fmt.Errorf("failed to create gift card issue transaction: %w", err)
		}

		result = &RedeemResult{
			RedemptionID:       redemption.ID,
			GiftCardID:         giftCard.ID,
			CardCode:           giftCard.Code,
			Balance:            giftCard.Balance,
			BalanceCents:       giftCard.BalanceCents,
			GiftCardValueCents: valueCents,
			PointsSpent:        pointsToSpend,
			PointsRemaining:    loyaltyTransaction.Balance,
			ExpiresAt:          giftCard.ExpiresAt,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func resolveRedeemOption(req RedeemGiftCardRequest, config *loyalty.ProgramConfig) (loyalty.ProgramRedeemOption, error) {
	if req.OptionID > 0 {
		for _, option := range config.RedeemOptions {
			if option.ID == req.OptionID {
				if option.RemainingQuantity() <= 0 {
					return loyalty.ProgramRedeemOption{}, repository.ErrRedeemOptionOutOfStock
				}
				return option, nil
			}
		}
		return loyalty.ProgramRedeemOption{}, errors.New("redeem option does not belong to the active program config")
	}

	if req.GiftCardValueCents <= 0 {
		return loyalty.ProgramRedeemOption{}, errors.New("redeem option is required")
	}
	for _, option := range config.RedeemOptions {
		if option.ValueCents == req.GiftCardValueCents && option.RemainingQuantity() > 0 {
			return option, nil
		}
	}
	return loyalty.ProgramRedeemOption{}, errors.New("gift card value is not an allowed redeem option")
}

func (s *MarketingService) redeemResultFromExisting(
	loyaltyRepo *repository.LoyaltyRepository,
	redemption *coupon.GiftCardRedemption,
) (*RedeemResult, error) {
	if redemption == nil || redemption.GiftCard == nil || redemption.LoyaltyTransactionID == nil {
		return nil, errors.New("existing redemption is incomplete")
	}
	transaction, err := loyaltyRepo.FindTransactionByID(*redemption.LoyaltyTransactionID)
	if err != nil {
		return nil, err
	}
	return &RedeemResult{
		RedemptionID:       redemption.ID,
		GiftCardID:         redemption.GiftCard.ID,
		CardCode:           redemption.GiftCard.Code,
		Balance:            redemption.GiftCard.Balance,
		BalanceCents:       redemption.GiftCard.BalanceCents,
		GiftCardValueCents: redemption.GiftCardValueCents,
		PointsSpent:        redemption.PointsSpent,
		PointsRemaining:    transaction.Balance,
		ExpiresAt:          redemption.GiftCard.ExpiresAt,
	}, nil
}

func localDayBounds(now time.Time) (time.Time, time.Time) {
	location := now.Location()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	return start, start.AddDate(0, 0, 1)
}

func createRedeemedGiftCard(
	couponRepo *repository.CouponRepository,
	userID uint,
	valueCents int64,
	currency string,
	expiresAt *time.Time,
) (*coupon.GiftCard, error) {
	const maxAttempts = 5
	var lastErr error
	ownerUserID := userID

	for attempt := 0; attempt < maxAttempts; attempt++ {
		cardCode, err := generateRedeemCode(16)
		if err != nil {
			return nil, fmt.Errorf("generate gift card code: %w", err)
		}

		giftCard := &coupon.GiftCard{
			Code:         "REDEEM-" + cardCode,
			InitialCents: valueCents,
			BalanceCents: valueCents,
			Currency:     currency,
			Status:       "active",
			OwnerUserID:  &ownerUserID,
			Origin:       "loyalty_redemption",
			ExpiresAt:    expiresAt,
		}

		if err := couponRepo.CreateGiftCard(giftCard); err != nil {
			lastErr = err
			if repository.IsDuplicatedKey(err) {
				continue
			}
			return nil, err
		}

		return giftCard, nil
	}

	return nil, fmt.Errorf("unique gift card code was not generated after %d attempts: %w", maxAttempts, lastErr)
}

func generateRedeemCode(length int) (string, error) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	for index := range randomBytes {
		randomBytes[index] = alphabet[int(randomBytes[index])%len(alphabet)]
	}
	return string(randomBytes), nil
}
