package payment

import "time"

// PaymentReview is an internal hold/review workflow independent from Stripe's
// payment state. It records why an order needs human attention and the final
// operator decision.
type PaymentReview struct {
	ID              uint       `gorm:"primarykey" json:"id"`
	OrderID         *uint      `gorm:"index" json:"order_id,omitempty"`
	TransactionID   *uint      `gorm:"index" json:"transaction_id,omitempty"`
	DisputeID       *uint      `gorm:"index" json:"dispute_id,omitempty"`
	PaymentIntentID string     `gorm:"index" json:"payment_intent_id,omitempty"`
	StripeReviewID  string     `gorm:"index" json:"stripe_review_id,omitempty"`
	Status          string     `gorm:"index;not null" json:"status"` // pending, approved, rejected, cancelled
	Reason          string     `gorm:"index;not null" json:"reason"`
	Source          string     `gorm:"index;not null" json:"source"` // radar, operator, dispute
	Notes           string     `gorm:"type:text" json:"notes"`
	AssignedToID    *uint      `gorm:"index" json:"assigned_to_id,omitempty"`
	ReviewedByID    *uint      `gorm:"index" json:"reviewed_by_id,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (PaymentReview) TableName() string {
	return "payment_reviews"
}
