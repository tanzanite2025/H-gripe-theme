package recommendation

import (
	"time"

	"gorm.io/datatypes"
)

// Event is an append-only storefront behavior fact.
// It intentionally contains no customer-service or device-fingerprint fields.
type Event struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	EventID      string         `gorm:"size:128;not null;uniqueIndex" json:"event_id"`
	EventType    string         `gorm:"size:64;not null;index" json:"event_type"`
	AnonymousID  string         `gorm:"size:128;index" json:"anonymous_id,omitempty"`
	SessionID    string         `gorm:"size:128;index" json:"session_id,omitempty"`
	UserID       *uint          `gorm:"index" json:"user_id,omitempty"`
	ProductID    *uint          `gorm:"index" json:"product_id,omitempty"`
	CategoryID   *uint          `gorm:"index" json:"category_id,omitempty"`
	Locale       string         `gorm:"size:20;index" json:"locale,omitempty"`
	Path         string         `gorm:"size:1024" json:"path,omitempty"`
	Referrer     string         `gorm:"size:1024" json:"referrer,omitempty"`
	MetadataJSON datatypes.JSON `gorm:"not null" json:"metadata"`
	OccurredAt   time.Time      `gorm:"index;not null" json:"occurred_at"`
	ReceivedAt   time.Time      `gorm:"index;not null" json:"received_at"`
}

func (Event) TableName() string {
	return "recommendation_events"
}
