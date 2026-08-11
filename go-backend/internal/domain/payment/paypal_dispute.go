package payment

import (
	"errors"
	"time"

	"tanzanite/internal/domain/currency"

	"gorm.io/gorm"
)

// PayPalDispute is the local operational record for a PayPal customer dispute.
type PayPalDispute struct {
	ID                        uint           `gorm:"primarykey" json:"id"`
	PayPalDisputeID           string         `gorm:"column:paypal_dispute_id;uniqueIndex;not null" json:"paypal_dispute_id"`
	OrderID                   *uint          `gorm:"index" json:"order_id,omitempty"`
	TransactionID             *uint          `gorm:"index" json:"transaction_id,omitempty"`
	ProviderPaymentID         string         `gorm:"index" json:"provider_payment_id"`
	Amount                    float64        `gorm:"not null" json:"amount"`
	Currency                  string         `gorm:"not null" json:"currency"`
	Reason                    string         `gorm:"index" json:"reason"`
	Status                    string         `gorm:"index;not null" json:"status"`
	DisputeState              string         `gorm:"index" json:"dispute_state"`
	DisputeLifeCycleStage     string         `gorm:"index" json:"dispute_life_cycle_stage"`
	RawPayload                string         `gorm:"type:text" json:"-"`
	EvidenceSubmittedAt       *time.Time     `gorm:"index" json:"evidence_submitted_at,omitempty"`
	EvidenceSubmissionPayload string         `gorm:"type:text" json:"-"`
	EvidenceSubmissionError   string         `gorm:"type:text" json:"evidence_submission_error,omitempty"`
	CreatedAt                 time.Time      `json:"created_at"`
	UpdatedAt                 time.Time      `json:"updated_at"`
	DeletedAt                 gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PayPalDispute) TableName() string {
	return "paypal_disputes"
}

func (d *PayPalDispute) BeforeSave(tx *gorm.DB) error {
	d.Currency = currency.NormalizeCode(d.Currency)
	if !currency.IsValidCode(d.Currency) || !currency.IsCatalogCode(d.Currency) {
		return errors.New("paypal dispute currency must be a supported ISO 4217 code")
	}
	return nil
}
