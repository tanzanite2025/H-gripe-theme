package shipping

import (
	"time"

	"gorm.io/datatypes"
)

// QuoteSnapshot is the immutable, short-lived server-side contract consumed
// by checkout. QuoteData contains the complete set of plans shown to the buyer.
type QuoteSnapshot struct {
	ID          string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	RequestHash string         `gorm:"type:char(64);not null;index" json:"-"`
	RateVersion string         `gorm:"type:char(64);not null" json:"rate_version"`
	QuoteData   datatypes.JSON `gorm:"column:quote_data;type:jsonb;not null" json:"-"`
	ExpiresAt   time.Time      `gorm:"not null;index" json:"expires_at"`
	CreatedAt   time.Time      `gorm:"not null" json:"created_at"`
}

func (QuoteSnapshot) TableName() string {
	return "shipping_quote_snapshots"
}
