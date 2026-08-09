package service

import (
	"errors"
	"fmt"
	"time"

	"tanzanite/internal/domain/currency"
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/repository"
)

var (
	ErrLoyaltyProgramConfigNotFound = errors.New("active loyalty program config not found")
	ErrInvalidLoyaltyProgramConfig  = errors.New("invalid loyalty program config")
)

const LoyaltyPointsBaseCurrency = "USD"

type LoyaltyProgramConfigInput struct {
	Enabled                   bool
	Currency                  string
	PurchaseEarnPointsPerUnit int
	ExchangeRatePoints        int
	MinRedeemPoints           int
	MaxValuePerDayCents       int64
	CardExpiryDays            int
	ReferralReferrerPoints    int
	ReferralRefereePoints     int
	CheckInBasePoints         int
	CheckInStreakIntervalDays int
	CheckInStreakBonusPoints  int
	CheckInMaxPoints          int
	RedeemOptions             []LoyaltyProgramOptionInput
	RedeemValuesCents         []int64
	CreatedBy                 *uint
}

type LoyaltyProgramOptionInput struct {
	ValueCents    int64
	Currency      string
	StockQuantity int64
}

type LoyaltyProgramConfigResponse struct {
	ID                        uint                           `json:"id"`
	Version                   int                            `json:"version"`
	Status                    string                         `json:"status"`
	Enabled                   bool                           `json:"enabled"`
	Currency                  string                         `json:"currency"`
	PointsBaseCurrency        string                         `json:"points_base_currency"`
	PurchaseEarnPointsPerUnit int                            `json:"purchase_earn_points_per_currency_unit"`
	ExchangeRatePoints        int                            `json:"exchange_rate_points"`
	MinRedeemPoints           int                            `json:"min_redeem_points"`
	MaxValuePerDayCents       int64                          `json:"max_value_per_day_cents"`
	MaxValuePerDay            float64                        `json:"max_value_per_day"`
	CardExpiryDays            int                            `json:"card_expiry_days"`
	ReferralReferrerPoints    int                            `json:"referral_referrer_points"`
	ReferralRefereePoints     int                            `json:"referral_referee_points"`
	CheckInBasePoints         int                            `json:"checkin_base_points"`
	CheckInStreakIntervalDays int                            `json:"checkin_streak_interval_days"`
	CheckInStreakBonusPoints  int                            `json:"checkin_streak_bonus_points"`
	CheckInMaxPoints          int                            `json:"checkin_max_points"`
	RedeemOptions             []LoyaltyProgramOptionResponse `json:"redeem_options"`
	AvailableCurrencies       []currency.CurrencyOption      `json:"available_currencies"`
	CreatedAt                 time.Time                      `json:"created_at"`
	UpdatedAt                 time.Time                      `json:"updated_at"`
}

type LoyaltyProgramOptionResponse struct {
	ID                uint    `json:"id"`
	ValueCents        int64   `json:"value_cents"`
	Value             float64 `json:"value"`
	Currency          string  `json:"currency"`
	PointsRequired    int     `json:"points_required"`
	StockQuantity     int64   `json:"stock_quantity"`
	RedeemedQuantity  int64   `json:"redeemed_quantity"`
	RemainingQuantity int64   `json:"remaining_quantity"`
	Label             string  `json:"label"`
	Status            string  `json:"status"`
}

type LoyaltyProgramService struct {
	repo *repository.LoyaltyProgramRepository
}

func NewLoyaltyProgramService(repo *repository.LoyaltyProgramRepository) *LoyaltyProgramService {
	return &LoyaltyProgramService{repo: repo}
}

func (s *LoyaltyProgramService) ConfigureCurrencyPolicy(policy *CurrencyPolicyService) {
}

func (s *LoyaltyProgramService) GetActive() (*loyalty.ProgramConfig, error) {
	config, err := s.repo.FindActive()
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrLoyaltyProgramConfigNotFound
		}
		return nil, err
	}
	if err := validateProgramConfig(config); err != nil {
		return nil, err
	}
	return config, nil
}

func (s *LoyaltyProgramService) GetPublicConfig() (*LoyaltyProgramConfigResponse, error) {
	config, err := s.GetActive()
	if err != nil {
		return nil, err
	}
	response := programConfigResponse(config)
	return &response, nil
}

func (s *LoyaltyProgramService) Update(input LoyaltyProgramConfigInput) (*loyalty.ProgramConfig, error) {
	config := &loyalty.ProgramConfig{
		Enabled:                   input.Enabled,
		Currency:                  input.Currency,
		PurchaseEarnPointsPerUnit: input.PurchaseEarnPointsPerUnit,
		ExchangeRatePoints:        input.ExchangeRatePoints,
		MinRedeemPoints:           input.MinRedeemPoints,
		MaxValuePerDayCents:       input.MaxValuePerDayCents,
		CardExpiryDays:            input.CardExpiryDays,
		ReferralReferrerPoints:    input.ReferralReferrerPoints,
		ReferralRefereePoints:     input.ReferralRefereePoints,
		CheckInBasePoints:         input.CheckInBasePoints,
		CheckInStreakIntervalDays: input.CheckInStreakIntervalDays,
		CheckInStreakBonusPoints:  input.CheckInStreakBonusPoints,
		CheckInMaxPoints:          input.CheckInMaxPoints,
		CreatedBy:                 input.CreatedBy,
	}

	optionInputs := input.RedeemOptions
	if len(optionInputs) == 0 {
		for _, valueCents := range input.RedeemValuesCents {
			optionInputs = append(optionInputs, LoyaltyProgramOptionInput{
				ValueCents:    valueCents,
				Currency:      input.Currency,
				StockQuantity: 0,
			})
		}
	}

	for index, optionInput := range optionInputs {
		optionCurrency := currency.NormalizeCode(optionInput.Currency)
		if optionCurrency == "" {
			optionCurrency = currency.NormalizeCode(input.Currency)
		}
		config.RedeemOptions = append(config.RedeemOptions, loyalty.ProgramRedeemOption{
			ValueCents:    optionInput.ValueCents,
			Currency:      optionCurrency,
			StockQuantity: optionInput.StockQuantity,
			SortOrder:     index,
		})
	}

	if err := validateProgramConfig(config); err != nil {
		return nil, err
	}
	if err := s.repo.CreateVersion(config); err != nil {
		return nil, err
	}
	return config, nil
}

func validateProgramConfig(config *loyalty.ProgramConfig) error {
	if config == nil {
		return ErrInvalidLoyaltyProgramConfig
	}

	config.Currency = currency.NormalizeCode(config.Currency)
	if !currency.IsValidCode(config.Currency) {
		return fmt.Errorf("%w: currency must be a three-letter code", ErrInvalidLoyaltyProgramConfig)
	}
	if !currency.IsCatalogCode(config.Currency) {
		return fmt.Errorf("%w: unsupported currency", ErrInvalidLoyaltyProgramConfig)
	}
	if config.PurchaseEarnPointsPerUnit < 0 {
		return fmt.Errorf("%w: purchase earn points cannot be negative", ErrInvalidLoyaltyProgramConfig)
	}
	if config.ExchangeRatePoints <= 0 {
		return fmt.Errorf("%w: exchange rate must be greater than zero", ErrInvalidLoyaltyProgramConfig)
	}
	if config.MinRedeemPoints < 0 || config.MaxValuePerDayCents < 0 || config.CardExpiryDays < 0 {
		return fmt.Errorf("%w: redemption limits cannot be negative", ErrInvalidLoyaltyProgramConfig)
	}
	if config.ReferralReferrerPoints < 0 || config.ReferralRefereePoints < 0 ||
		config.CheckInBasePoints < 0 || config.CheckInStreakBonusPoints < 0 ||
		config.CheckInMaxPoints < 0 {
		return fmt.Errorf("%w: loyalty points cannot be negative", ErrInvalidLoyaltyProgramConfig)
	}
	if config.CheckInStreakIntervalDays <= 0 {
		return fmt.Errorf("%w: check-in interval must be greater than zero", ErrInvalidLoyaltyProgramConfig)
	}
	if config.CheckInMaxPoints < config.CheckInBasePoints {
		return fmt.Errorf("%w: check-in max points cannot be lower than base points", ErrInvalidLoyaltyProgramConfig)
	}
	if config.Enabled && len(config.RedeemOptions) == 0 {
		return fmt.Errorf("%w: at least one redeem option is required", ErrInvalidLoyaltyProgramConfig)
	}

	seen := make(map[string]struct{}, len(config.RedeemOptions))
	for index := range config.RedeemOptions {
		option := &config.RedeemOptions[index]
		if option.ValueCents <= 0 {
			return fmt.Errorf("%w: redeem option value must be greater than zero", ErrInvalidLoyaltyProgramConfig)
		}
		option.Currency = currency.NormalizeCode(option.Currency)
		if !currency.IsValidCode(option.Currency) || !currency.IsCatalogCode(option.Currency) {
			return fmt.Errorf("%w: redeem option currency is unsupported", ErrInvalidLoyaltyProgramConfig)
		}
		if option.StockQuantity < 0 || option.RedeemedQuantity < 0 {
			return fmt.Errorf("%w: redeem option stock cannot be negative", ErrInvalidLoyaltyProgramConfig)
		}
		if option.RedeemedQuantity > option.StockQuantity {
			return fmt.Errorf("%w: redeemed quantity cannot exceed stock quantity", ErrInvalidLoyaltyProgramConfig)
		}
		key := redeemOptionKey(option.Currency, option.ValueCents)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("%w: duplicate redeem option value", ErrInvalidLoyaltyProgramConfig)
		}
		seen[key] = struct{}{}
		option.SortOrder = index
	}

	return nil
}

func redeemOptionKey(currencyCode string, valueCents int64) string {
	return fmt.Sprintf("%s:%d", currency.NormalizeCode(currencyCode), valueCents)
}

func PointsForGiftCardValue(valueCents int64, exchangeRatePoints int) (int, error) {
	if valueCents <= 0 || exchangeRatePoints <= 0 {
		return 0, ErrInvalidLoyaltyProgramConfig
	}
	points := (valueCents*int64(exchangeRatePoints) + 50) / 100
	maxInt := int64(^uint(0) >> 1)
	if points <= 0 || points > maxInt {
		return 0, fmt.Errorf("%w: calculated points are out of range", ErrInvalidLoyaltyProgramConfig)
	}
	return int(points), nil
}

func programConfigResponse(config *loyalty.ProgramConfig) LoyaltyProgramConfigResponse {
	response := LoyaltyProgramConfigResponse{
		ID:                        config.ID,
		Version:                   config.Version,
		Status:                    config.Status,
		Enabled:                   config.Enabled,
		Currency:                  config.Currency,
		PointsBaseCurrency:        LoyaltyPointsBaseCurrency,
		PurchaseEarnPointsPerUnit: config.PurchaseEarnPointsPerUnit,
		ExchangeRatePoints:        config.ExchangeRatePoints,
		MinRedeemPoints:           config.MinRedeemPoints,
		MaxValuePerDayCents:       config.MaxValuePerDayCents,
		MaxValuePerDay:            float64(config.MaxValuePerDayCents) / 100,
		CardExpiryDays:            config.CardExpiryDays,
		ReferralReferrerPoints:    config.ReferralReferrerPoints,
		ReferralRefereePoints:     config.ReferralRefereePoints,
		CheckInBasePoints:         config.CheckInBasePoints,
		CheckInStreakIntervalDays: config.CheckInStreakIntervalDays,
		CheckInStreakBonusPoints:  config.CheckInStreakBonusPoints,
		CheckInMaxPoints:          config.CheckInMaxPoints,
		RedeemOptions:             make([]LoyaltyProgramOptionResponse, 0, len(config.RedeemOptions)),
		AvailableCurrencies:       currency.Catalog(),
		CreatedAt:                 config.CreatedAt,
		UpdatedAt:                 config.UpdatedAt,
	}

	for _, option := range config.RedeemOptions {
		pointsRequired, _ := PointsForGiftCardValue(option.ValueCents, config.ExchangeRatePoints)
		value := float64(option.ValueCents) / 100
		response.RedeemOptions = append(response.RedeemOptions, LoyaltyProgramOptionResponse{
			ID:                option.ID,
			ValueCents:        option.ValueCents,
			Value:             value,
			Currency:          option.Currency,
			PointsRequired:    pointsRequired,
			StockQuantity:     option.StockQuantity,
			RedeemedQuantity:  option.RedeemedQuantity,
			RemainingQuantity: option.RemainingQuantity(),
			Label:             fmt.Sprintf("%s %.2f Gift Card", option.Currency, value),
			Status:            redeemOptionPublicStatus(config, pointsRequired, option.RemainingQuantity()),
		})
	}

	return response
}

func redeemOptionPublicStatus(config *loyalty.ProgramConfig, pointsRequired int, remainingQuantity int64) string {
	if config == nil || !config.Enabled || pointsRequired < config.MinRedeemPoints || remainingQuantity <= 0 {
		return "inactive"
	}
	return "active"
}
