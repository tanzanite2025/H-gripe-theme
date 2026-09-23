package loyalty

import (
	"time"

	"gorm.io/datatypes"
)

const (
	ReferralStatusPending  = "pending"
	ReferralStatusOrdered  = "ordered"
	ReferralStatusVesting  = "vesting"
	ReferralStatusSettled  = "settled"
	ReferralStatusExpired  = "expired"
	ReferralStatusRevoked  = "revoked"
	ReferralStatusReversed = "reversed"

	ReferralRewardStatusLocked    = "locked"
	ReferralRewardStatusReleased  = "released"
	ReferralRewardStatusForfeited = "forfeited"
	ReferralRewardStatusReversed  = "reversed"

	ReferralRecipientReferrer = "referrer"
	ReferralRecipientReferee  = "referee"

	ReferralRewardTypePoints = "points"

	ReferralBenefitPoints = "points"

	ReferralFraudModeMonitor = "monitor"
	ReferralFraudModeStrict  = "strict"
)

// ReferralIdentity owns the stable share code for one referrer. Aggregate
// counts are derived from referral_records rather than persisted here.
type ReferralIdentity struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	UserID       uint      `gorm:"not null;uniqueIndex" json:"user_id"`
	ReferralCode string    `gorm:"size:16;not null" json:"referral_code"`
	CustomSlug   *string   `gorm:"size:64" json:"custom_slug,omitempty"`
	IsActive     bool      `gorm:"not null;default:true;index" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (ReferralIdentity) TableName() string {
	return "user_referral_identities"
}

// ReferralProgramConfig is an immutable referral-policy version. The referral
// benefit is always a points amount credited to the unified loyalty balance.
type ReferralProgramConfig struct {
	ID      uint   `gorm:"primarykey" json:"id"`
	Version int    `gorm:"not null;uniqueIndex" json:"version"`
	Status  string `gorm:"size:16;not null;index" json:"status"`
	Enabled bool   `gorm:"not null;default:false" json:"enabled"`
	// Currency is used only by the order-qualification rule. It is not a
	// property of referral points and is intentionally omitted from API JSON.
	Currency                     string    `gorm:"size:3;not null;default:'USD'" json:"-"`
	MinOrderAmountMinor          int64     `gorm:"not null;default:20000" json:"min_order_amount_minor"`
	ReferrerRewardPoints         int       `gorm:"not null;default:1000" json:"referrer_reward_points"`
	RefereeBenefitType           string    `gorm:"size:24;not null;default:'points'" json:"referee_benefit_type"`
	RefereeBenefitValue          int64     `gorm:"not null;default:50" json:"referee_benefit_value"`
	VestingPeriodDays            int       `gorm:"not null;default:30" json:"vesting_period_days"`
	UndeliveredFallbackDays      int       `gorm:"not null;default:45" json:"undelivered_fallback_days"`
	AttributionTTLDays           int       `gorm:"not null;default:30" json:"attribution_ttl_days"`
	MonthlyCapPerReferrer        int       `gorm:"not null;default:10" json:"monthly_cap_per_referrer"`
	AntiFraudMode                string    `gorm:"size:16;not null;default:'monitor'" json:"anti_fraud_mode"`
	SourceLoyaltyProgramConfigID *uint     `gorm:"index" json:"source_loyalty_program_config_id,omitempty"`
	CreatedBy                    *uint     `gorm:"index" json:"created_by,omitempty"`
	CreatedAt                    time.Time `json:"created_at"`
}

func (ReferralProgramConfig) TableName() string {
	return "referral_program_configs"
}

// ReferralRecord is the authoritative referral lifecycle. Personal and device
// comparison values are versioned HMACs; raw identifiers do not belong here.
type ReferralRecord struct {
	ID                     uint           `gorm:"primarykey" json:"id"`
	ReferralIdentityID     uint           `gorm:"not null;index" json:"referral_identity_id"`
	ProgramConfigID        uint           `gorm:"not null;index" json:"program_config_id"`
	ReferrerID             uint           `gorm:"not null;index" json:"referrer_id"`
	RefereeID              *uint          `gorm:"index" json:"referee_id,omitempty"`
	ReferralCodeSnapshot   string         `gorm:"size:16;not null;index" json:"referral_code"`
	AttributionSource      string         `gorm:"size:24;not null" json:"attribution_source"`
	RefereeEmailHash       string         `gorm:"size:64;not null;default:''" json:"-"`
	ClientIPHash           string         `gorm:"size:64;not null;default:''" json:"-"`
	ClientIPSubnetHash     string         `gorm:"size:64;not null;default:'';index" json:"-"`
	DeviceFingerprintHash  string         `gorm:"size:64;not null;default:''" json:"-"`
	ShippingAddressHash    string         `gorm:"size:64;not null;default:''" json:"-"`
	ShippingPhoneHash      string         `gorm:"size:64;not null;default:''" json:"-"`
	PaymentFingerprintHash string         `gorm:"size:64;not null;default:''" json:"-"`
	LegacyReferralID       *uint          `gorm:"index" json:"-"`
	HashKeyVersion         int            `gorm:"not null;default:1" json:"-"`
	OrderID                *uint          `gorm:"index" json:"order_id,omitempty"`
	Currency               string         `gorm:"size:3;not null;default:'USD'" json:"currency"`
	OrderAmountMinor       int64          `gorm:"not null;default:0" json:"order_amount_minor"`
	Status                 string         `gorm:"size:16;not null;index" json:"status"`
	RecordVersion          int            `gorm:"not null;default:1" json:"record_version"`
	ExpiresAt              time.Time      `gorm:"not null;index" json:"expires_at"`
	OrderedAt              *time.Time     `json:"ordered_at,omitempty"`
	DeliveredAt            *time.Time     `json:"delivered_at,omitempty"`
	VestingUntil           *time.Time     `gorm:"index" json:"vesting_until,omitempty"`
	SettledAt              *time.Time     `json:"settled_at,omitempty"`
	ExpiredAt              *time.Time     `json:"expired_at,omitempty"`
	RevokedAt              *time.Time     `json:"revoked_at,omitempty"`
	ReversedAt             *time.Time     `json:"reversed_at,omitempty"`
	RevokeReason           string         `gorm:"type:text" json:"revoke_reason,omitempty"`
	ReverseReason          string         `gorm:"type:text" json:"reverse_reason,omitempty"`
	RiskFlags              datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"risk_flags"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
}

func (ReferralRecord) TableName() string {
	return "referral_records"
}

func (r ReferralRecord) CanTransitionTo(next string) bool {
	switch r.Status {
	case ReferralStatusPending:
		return next == ReferralStatusOrdered || next == ReferralStatusExpired || next == ReferralStatusRevoked
	case ReferralStatusOrdered:
		return next == ReferralStatusVesting || next == ReferralStatusRevoked
	case ReferralStatusVesting:
		return next == ReferralStatusSettled || next == ReferralStatusRevoked
	case ReferralStatusSettled:
		return next == ReferralStatusReversed
	default:
		return false
	}
}

// ReferralReward records the financial side effect independently from the
// lifecycle record so retries can be proven idempotent and reversals audited.
type ReferralReward struct {
	ID                   uint           `gorm:"primarykey" json:"id"`
	ReferralRecordID     uint           `gorm:"not null;index" json:"referral_record_id"`
	ProgramConfigID      uint           `gorm:"not null;index" json:"program_config_id"`
	RecipientUserID      uint           `gorm:"not null;index" json:"recipient_user_id"`
	RecipientRole        string         `gorm:"size:16;not null" json:"recipient_role"`
	RewardType           string         `gorm:"size:16;not null" json:"reward_type"`
	PointsAmount         int            `gorm:"not null;default:0" json:"points_amount"`
	LoyaltyTransactionID *uint          `gorm:"index" json:"loyalty_transaction_id,omitempty"`
	IdempotencyKey       string         `gorm:"size:160;not null;uniqueIndex" json:"idempotency_key"`
	Status               string         `gorm:"size:16;not null;index" json:"status"`
	RuleSnapshot         datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"rule_snapshot"`
	ReleasedAt           *time.Time     `json:"released_at,omitempty"`
	ForfeitedAt          *time.Time     `json:"forfeited_at,omitempty"`
	ReversedAt           *time.Time     `json:"reversed_at,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

func (ReferralReward) TableName() string {
	return "referral_reward_logs"
}

// ReferralTransition is append-only evidence for every accepted state change.
type ReferralTransition struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	ReferralRecordID uint           `gorm:"not null;index" json:"referral_record_id"`
	FromStatus       string         `gorm:"size:16;not null" json:"from_status"`
	ToStatus         string         `gorm:"size:16;not null" json:"to_status"`
	Trigger          string         `gorm:"size:40;not null" json:"trigger"`
	Reason           string         `gorm:"type:text" json:"reason,omitempty"`
	ActorType        string         `gorm:"size:24;not null" json:"actor_type"`
	ActorID          *uint          `gorm:"index" json:"actor_id,omitempty"`
	EventKey         *string        `gorm:"size:160" json:"event_key,omitempty"`
	Metadata         datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	CreatedAt        time.Time      `json:"created_at"`
}

func (ReferralTransition) TableName() string {
	return "referral_transitions"
}
