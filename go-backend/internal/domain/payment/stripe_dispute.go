package payment

import (
	"errors"
	"time"

	"commerce-platform/internal/domain/currency"

	"gorm.io/gorm"
)

// StripeDispute is the local operational record for a Stripe chargeback.
type StripeDispute struct {
	ID                        uint           `gorm:"primarykey" json:"id"`
	StripeDisputeID           string         `gorm:"uniqueIndex;not null" json:"stripe_dispute_id"`
	StripeChargeID            string         `gorm:"index" json:"stripe_charge_id"`
	PaymentIntentID           string         `gorm:"index" json:"payment_intent_id"`
	OrderID                   *uint          `gorm:"index" json:"order_id,omitempty"`
	TransactionID             *uint          `gorm:"index" json:"transaction_id,omitempty"`
	Amount                    float64        `gorm:"not null" json:"amount"`
	Currency                  string         `gorm:"not null" json:"currency"`
	Reason                    string         `gorm:"index" json:"reason"`
	Status                    string         `gorm:"index;not null" json:"status"` // needs_response, warning_needs_response, under_review, won, lost, closed
	EvidenceDueAt             *time.Time     `gorm:"index" json:"evidence_due_at,omitempty"`
	RawPayload                string         `gorm:"type:text" json:"-"`
	EvidenceSubmittedAt       *time.Time     `gorm:"index" json:"evidence_submitted_at,omitempty"`
	EvidenceSubmissionPayload string         `gorm:"type:text" json:"-"`
	EvidenceSubmissionError   string         `gorm:"type:text" json:"evidence_submission_error,omitempty"`
	CreatedAt                 time.Time      `json:"created_at"`
	UpdatedAt                 time.Time      `json:"updated_at"`
	DeletedAt                 gorm.DeletedAt `gorm:"index" json:"-"`
}

func (StripeDispute) TableName() string {
	return "stripe_disputes"
}

func (d *StripeDispute) BeforeSave(tx *gorm.DB) error {
	d.Currency = currency.NormalizeCode(d.Currency)
	if !currency.IsValidCode(d.Currency) || !currency.IsCatalogCode(d.Currency) {
		return errors.New("stripe dispute currency must be a supported ISO 4217 code")
	}
	return nil
}
