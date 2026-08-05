package payment

import "time"

// StripeWebhookEvent stores the delivery state of a verified Stripe event.
// The event ID is the idempotency boundary for webhook processing.
type StripeWebhookEvent struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	EventID      string     `gorm:"uniqueIndex;not null" json:"event_id"`
	EventType    string     `gorm:"index;not null" json:"event_type"`
	Status       string     `gorm:"index;not null" json:"status"` // processing, processed, failed
	Payload      string     `gorm:"type:text" json:"-"`
	ErrorMessage string     `gorm:"type:text" json:"error_message,omitempty"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (StripeWebhookEvent) TableName() string {
	return "stripe_webhook_events"
}
