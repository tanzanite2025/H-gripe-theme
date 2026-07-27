package loyalty

import "time"

// ProgramConfig is an immutable version of the loyalty and redemption rules.
// A new admin save creates a new version instead of mutating historical rules.
type ProgramConfig struct {
	ID                        uint                  `gorm:"primarykey" json:"id"`
	Version                   int                   `gorm:"uniqueIndex;not null" json:"version"`
	Status                    string                `gorm:"index;not null;default:'active'" json:"status"` // active, archived
	Enabled                   bool                  `gorm:"not null;default:true" json:"enabled"`
	Currency                  string                `gorm:"not null;default:'USD'" json:"currency"`
	ExchangeRatePoints        int                   `gorm:"not null" json:"exchange_rate_points"`
	MinRedeemPoints           int                   `gorm:"not null" json:"min_redeem_points"`
	MaxValuePerDayCents       int64                 `gorm:"not null" json:"max_value_per_day_cents"`
	CardExpiryDays            int                   `gorm:"not null" json:"card_expiry_days"`
	ReferralReferrerPoints    int                   `gorm:"not null" json:"referral_referrer_points"`
	ReferralRefereePoints     int                   `gorm:"not null" json:"referral_referee_points"`
	CheckInBasePoints         int                   `gorm:"not null" json:"checkin_base_points"`
	CheckInStreakIntervalDays int                   `gorm:"not null" json:"checkin_streak_interval_days"`
	CheckInStreakBonusPoints  int                   `gorm:"not null" json:"checkin_streak_bonus_points"`
	CheckInMaxPoints          int                   `gorm:"not null" json:"checkin_max_points"`
	CreatedBy                 *uint                 `json:"created_by"`
	CreatedAt                 time.Time             `json:"created_at"`
	UpdatedAt                 time.Time             `json:"updated_at"`
	RedeemOptions             []ProgramRedeemOption `gorm:"foreignKey:ConfigID" json:"redeem_options"`
}

func (ProgramConfig) TableName() string {
	return "loyalty_program_configs"
}

// ProgramRedeemOption is a value that belongs to one immutable config version.
type ProgramRedeemOption struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	ConfigID   uint      `gorm:"not null;index" json:"config_id"`
	ValueCents int64     `gorm:"not null" json:"value_cents"`
	SortOrder  int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ProgramRedeemOption) TableName() string {
	return "loyalty_program_redeem_options"
}

// DefaultProgramConfig is used only for schema bootstrap and tests.
func DefaultProgramConfig() ProgramConfig {
	return ProgramConfig{
		Version:                   1,
		Status:                    "active",
		Enabled:                   true,
		Currency:                  "USD",
		ExchangeRatePoints:        100,
		MinRedeemPoints:           1000,
		MaxValuePerDayCents:       50000,
		CardExpiryDays:            365,
		ReferralReferrerPoints:    100,
		ReferralRefereePoints:     50,
		CheckInBasePoints:         10,
		CheckInStreakIntervalDays: 7,
		CheckInStreakBonusPoints:  5,
		CheckInMaxPoints:          50,
	}
}
