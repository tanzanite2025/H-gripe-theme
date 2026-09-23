package payment

import "time"

// PayPalWebhookEvent stores the delivery state of a verified PayPal event.
// EventID is the atomic idempotency boundary for webhook side effects.
type PayPalWebhookEvent struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	EventID      string     `gorm:"uniqueIndex;not null" json:"event_id"`
	EventType    string     `gorm:"index;not null" json:"event_type"`
	Status       string     `gorm:"index;not null" json:"status"`
	Payload      string     `gorm:"type:text" json:"-"`
	ErrorMessage string     `gorm:"type:text" json:"error_message,omitempty"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (PayPalWebhookEvent) TableName() string { return "paypal_webhook_events" }
