package quickbuy

import (
	"time"

	"gorm.io/datatypes"
)

const (
	SessionStatusActive      = "active"
	SessionStatusCompleted   = "completed"
	SessionStatusAddedToCart = "added_to_cart"
	SessionStatusOrdered     = "ordered"
	SessionStatusAbandoned   = "abandoned"
	SessionStatusExpired     = "expired"

	ValidationStatusValid   = "valid"
	ValidationStatusWarning = "warning"
	ValidationStatusInvalid = "invalid"
)

type Session struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	SessionToken     string         `gorm:"size:96;uniqueIndex;not null" json:"session_token"`
	FlowID           uint           `gorm:"not null;index" json:"flow_id"`
	FlowVersionID    uint           `gorm:"not null;index" json:"flow_version_id"`
	Locale           string         `gorm:"size:32;not null;default:'en'" json:"locale"`
	MarketCountry    string         `gorm:"size:8;not null;default:''" json:"market_country"`
	Currency         string         `gorm:"size:3;not null;default:'USD'" json:"currency"`
	AnonymousID      string         `gorm:"size:128;not null;default:''" json:"anonymous_id"`
	UserID           *uint          `gorm:"index" json:"user_id,omitempty"`
	Status           string         `gorm:"size:24;not null;default:'active';index" json:"status"`
	ValidationStatus string         `gorm:"size:24;not null;default:'valid'" json:"validation_status"`
	SubtotalSnapshot float64        `gorm:"type:numeric(12,2);not null;default:0" json:"subtotal_snapshot"`
	WeightSnapshotG  int            `gorm:"not null;default:0" json:"weight_snapshot_g"`
	Metadata         datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	ExpiresAt        *time.Time     `json:"expires_at,omitempty"`
	Flow             *Flow          `gorm:"foreignKey:FlowID" json:"flow,omitempty"`
	Version          *Version       `gorm:"foreignKey:FlowVersionID" json:"version,omitempty"`
	Items            []SessionItem  `gorm:"foreignKey:SessionID" json:"items,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func (Session) TableName() string {
	return "quick_buy_sessions"
}

type SessionItem struct {
	ID                uint           `gorm:"primarykey" json:"id"`
	SessionID         uint           `gorm:"not null;index" json:"session_id"`
	StepID            uint           `gorm:"not null;index" json:"step_id"`
	StepKey           string         `gorm:"size:120;not null" json:"step_key"`
	ProductID         uint           `gorm:"not null;index" json:"product_id"`
	VariantID         *uint          `gorm:"index" json:"variant_id,omitempty"`
	Quantity          int            `gorm:"not null;default:1" json:"quantity"`
	UnitPriceSnapshot float64        `gorm:"type:numeric(12,2);not null;default:0" json:"unit_price_snapshot"`
	CurrencySnapshot  string         `gorm:"size:3;not null;default:'USD'" json:"currency_snapshot"`
	WeightSnapshotG   int            `gorm:"not null;default:0" json:"weight_snapshot_g"`
	ProductSnapshot   datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"product_snapshot"`
	VariantSnapshot   datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"variant_snapshot"`
	SortOrder         int            `gorm:"not null;default:100" json:"sort_order"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

func (SessionItem) TableName() string {
	return "quick_buy_session_items"
}
