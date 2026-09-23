package loyalty

import "time"

// ProgramConfig is an immutable version of the loyalty rules.
// A new admin save creates a new version instead of mutating historical rules.
type ProgramConfig struct {
	ID                        uint      `gorm:"primarykey" json:"id"`
	Version                   int       `gorm:"uniqueIndex;not null" json:"version"`
	Status                    string    `gorm:"index;not null;default:'active'" json:"status"` // active, archived
	Enabled                   bool      `gorm:"not null;default:true" json:"enabled"`
	Currency                  string    `gorm:"not null" json:"currency"`
	PurchaseEarnPointsPerUnit int       `gorm:"column:purchase_earn_points_per_currency_unit;not null;default:1" json:"purchase_earn_points_per_currency_unit"`
	ExchangeRatePoints        int       `gorm:"not null" json:"exchange_rate_points"`
	ReferralReferrerPoints    int       `gorm:"not null" json:"referral_referrer_points"`
	ReferralRefereePoints     int       `gorm:"not null" json:"referral_referee_points"`
	CheckInBasePoints         int       `gorm:"column:checkin_base_points;not null" json:"checkin_base_points"`
	CheckInStreakIntervalDays int       `gorm:"column:checkin_streak_interval_days;not null" json:"checkin_streak_interval_days"`
	CheckInStreakBonusPoints  int       `gorm:"column:checkin_streak_bonus_points;not null" json:"checkin_streak_bonus_points"`
	CheckInMaxPoints          int       `gorm:"column:checkin_max_points;not null" json:"checkin_max_points"`
	CreatedBy                 *uint     `json:"created_by"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

func (ProgramConfig) TableName() string {
	return "loyalty_program_configs"
}
