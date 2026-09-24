package service

import (
	"errors"
	"fmt"
	"time"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/repository"
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
	ReferralReferrerPoints    int
	ReferralRefereePoints     int
	CheckInBasePoints         int
	CheckInStreakIntervalDays int
	CheckInStreakBonusPoints  int
	CheckInMaxPoints          int
	CreatedBy                 *uint
}

type LoyaltyProgramConfigResponse struct {
	ID                        uint                      `json:"id"`
	Version                   int                       `json:"version"`
	Status                    string                    `json:"status"`
	Enabled                   bool                      `json:"enabled"`
	Currency                  string                    `json:"currency"`
	PointsBaseCurrency        string                    `json:"points_base_currency"`
	PurchaseEarnPointsPerUnit int                       `json:"purchase_earn_points_per_currency_unit"`
	ReferralReferrerPoints    int                       `json:"referral_referrer_points"`
	ReferralRefereePoints     int                       `json:"referral_referee_points"`
	CheckInBasePoints         int                       `json:"checkin_base_points"`
	CheckInStreakIntervalDays int                       `json:"checkin_streak_interval_days"`
	CheckInStreakBonusPoints  int                       `json:"checkin_streak_bonus_points"`
	CheckInMaxPoints          int                       `json:"checkin_max_points"`
	AvailableCurrencies       []currency.CurrencyOption `json:"available_currencies"`
	CreatedAt                 time.Time                 `json:"created_at"`
	UpdatedAt                 time.Time                 `json:"updated_at"`
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
		ReferralReferrerPoints:    input.ReferralReferrerPoints,
		ReferralRefereePoints:     input.ReferralRefereePoints,
		CheckInBasePoints:         input.CheckInBasePoints,
		CheckInStreakIntervalDays: input.CheckInStreakIntervalDays,
		CheckInStreakBonusPoints:  input.CheckInStreakBonusPoints,
		CheckInMaxPoints:          input.CheckInMaxPoints,
		CreatedBy:                 input.CreatedBy,
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
	return nil
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
		ReferralReferrerPoints:    config.ReferralReferrerPoints,
		ReferralRefereePoints:     config.ReferralRefereePoints,
		CheckInBasePoints:         config.CheckInBasePoints,
		CheckInStreakIntervalDays: config.CheckInStreakIntervalDays,
		CheckInStreakBonusPoints:  config.CheckInStreakBonusPoints,
		CheckInMaxPoints:          config.CheckInMaxPoints,
		AvailableCurrencies:       currency.Catalog(),
		CreatedAt:                 config.CreatedAt,
		UpdatedAt:                 config.UpdatedAt,
	}

	return response
}
