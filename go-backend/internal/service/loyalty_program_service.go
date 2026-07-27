package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/repository"
)

var (
	ErrLoyaltyProgramConfigNotFound = errors.New("active loyalty program config not found")
	ErrInvalidLoyaltyProgramConfig  = errors.New("invalid loyalty program config")
)

type LoyaltyProgramConfigInput struct {
	Enabled                   bool
	Currency                  string
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
	RedeemValuesCents         []int64
	CreatedBy                 *uint
}

type LoyaltyProgramConfigResponse struct {
	ID                        uint                           `json:"id"`
	Version                   int                            `json:"version"`
	Status                    string                         `json:"status"`
	Enabled                   bool                           `json:"enabled"`
	Currency                  string                         `json:"currency"`
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
	CreatedAt                 time.Time                      `json:"created_at"`
	UpdatedAt                 time.Time                      `json:"updated_at"`
}

type LoyaltyProgramOptionResponse struct {
	ID             uint    `json:"id"`
	ValueCents     int64   `json:"value_cents"`
	Value          float64 `json:"value"`
	PointsRequired int     `json:"points_required"`
	Label          string  `json:"label"`
	Status         string  `json:"status"`
}

type LoyaltyProgramService struct {
	repo *repository.LoyaltyProgramRepository
}

func NewLoyaltyProgramService(repo *repository.LoyaltyProgramRepository) *LoyaltyProgramService {
	return &LoyaltyProgramService{repo: repo}
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

	for index, valueCents := range input.RedeemValuesCents {
		config.RedeemOptions = append(config.RedeemOptions, loyalty.ProgramRedeemOption{
			ValueCents: valueCents,
			SortOrder:  index,
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

	config.Currency = strings.ToUpper(strings.TrimSpace(config.Currency))
	if len(config.Currency) != 3 {
		return fmt.Errorf("%w: currency must be a three-letter code", ErrInvalidLoyaltyProgramConfig)
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
	if len(config.RedeemOptions) == 0 {
		return fmt.Errorf("%w: at least one redeem option is required", ErrInvalidLoyaltyProgramConfig)
	}

	seen := make(map[int64]struct{}, len(config.RedeemOptions))
	for index := range config.RedeemOptions {
		option := &config.RedeemOptions[index]
		if option.ValueCents <= 0 {
			return fmt.Errorf("%w: redeem option value must be greater than zero", ErrInvalidLoyaltyProgramConfig)
		}
		if _, exists := seen[option.ValueCents]; exists {
			return fmt.Errorf("%w: duplicate redeem option value", ErrInvalidLoyaltyProgramConfig)
		}
		seen[option.ValueCents] = struct{}{}
		option.SortOrder = index
	}

	return nil
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
		CreatedAt:                 config.CreatedAt,
		UpdatedAt:                 config.UpdatedAt,
	}

	for _, option := range config.RedeemOptions {
		pointsRequired, _ := PointsForGiftCardValue(option.ValueCents, config.ExchangeRatePoints)
		value := float64(option.ValueCents) / 100
		response.RedeemOptions = append(response.RedeemOptions, LoyaltyProgramOptionResponse{
			ID:             option.ID,
			ValueCents:     option.ValueCents,
			Value:          value,
			PointsRequired: pointsRequired,
			Label:          fmt.Sprintf("%s %.2f Gift Card", config.Currency, value),
			Status:         redeemOptionPublicStatus(config, pointsRequired),
		})
	}

	return response
}

func redeemOptionPublicStatus(config *loyalty.ProgramConfig, pointsRequired int) string {
	if config == nil || !config.Enabled || pointsRequired < config.MinRedeemPoints {
		return "inactive"
	}
	return "active"
}
