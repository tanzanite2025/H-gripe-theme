package payment

import (
	"errors"
	"time"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"gorm.io/gorm"
)

const (
	PaymentRefundRecommendationStatusPending   = "pending"
	PaymentRefundRecommendationStatusAccepted  = "accepted"
	PaymentRefundRecommendationStatusDismissed = "dismissed"
	PaymentRefundRecommendationStatusCancelled = "cancelled"
)

const (
	PaymentRefundRecommendationPriorityNormal = "normal"
	PaymentRefundRecommendationPriorityHigh   = "high"
)

const (
	PaymentRefundRecommendationActionReviewRefundBeforeDispute = "review_refund_before_dispute"
	PaymentRefundRecommendationActionReviewRefundOrEvidence    = "review_refund_or_evidence"
)

// PaymentRefundRecommendation is an operator-facing recommendation created by
// risk webhooks. It never executes a gateway refund by itself.
type PaymentRefundRecommendation struct {
	ID                 uint                 `gorm:"primarykey" json:"id"`
	Provider           string               `gorm:"index;not null" json:"provider"`
	SourceKind         PaymentRiskEventKind `gorm:"index;not null" json:"source_kind"`
	ExternalReference  string               `gorm:"not null" json:"external_reference"`
	WebhookEventID     string               `gorm:"index" json:"webhook_event_id"`
	RiskEventID        *uint                `gorm:"index" json:"risk_event_id,omitempty"`
	OrderID            *uint                `gorm:"index" json:"order_id,omitempty"`
	TransactionID      *uint                `gorm:"index" json:"transaction_id,omitempty"`
	LinkedRefundID     *uint                `gorm:"index" json:"linked_refund_id,omitempty"`
	ProviderPaymentID  string               `gorm:"index" json:"provider_payment_id,omitempty"`
	PaymentIntentID    string               `gorm:"index" json:"payment_intent_id,omitempty"`
	ChargeID           string               `gorm:"index" json:"charge_id,omitempty"`
	RecommendedAction  string               `gorm:"index;not null" json:"recommended_action"`
	RecommendedAmountMinor int64            `gorm:"column:recommended_amount_minor;not null;default:0" json:"recommended_amount_minor"`
	Currency           string               `gorm:"not null;default:''" json:"currency"`
	Priority           string               `gorm:"index;not null;default:'normal'" json:"priority"`
	Status             string               `gorm:"index;not null;default:'pending'" json:"status"`
	Reason             string               `gorm:"not null" json:"reason"`
	ProviderReason     string               `gorm:"not null;default:''" json:"provider_reason"`
	ReviewBy           *time.Time           `json:"review_by,omitempty"`
	DecisionNotes      string               `gorm:"type:text" json:"decision_notes"`
	ReviewedByID       *uint                `gorm:"index" json:"reviewed_by_id,omitempty"`
	ReviewedAt         *time.Time           `json:"reviewed_at,omitempty"`
	RefundCreatedByID  *uint                `gorm:"index" json:"refund_created_by_id,omitempty"`
	RefundCreatedAt    *time.Time           `json:"refund_created_at,omitempty"`
	SourceMetadataJSON string               `gorm:"type:text" json:"-"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

func (r *PaymentRefundRecommendation) BeforeSave(tx *gorm.DB) error {
	r.Currency = currency.NormalizeCode(r.Currency)
	if r.Currency != "" && !currency.IsCatalogCode(r.Currency) {
		return errors.New("payment refund recommendation currency must be a supported ISO 4217 code")
	}
	if r.RecommendedAmountMinor < 0 {
		return errors.New("payment refund recommendation amount cannot be negative")
	}
	return nil
}

func (r PaymentRefundRecommendation) RecommendedAmountMoney() (domainmoney.Money, error) {
	return domainmoney.New(r.RecommendedAmountMinor, r.Currency)
}

func (PaymentRefundRecommendation) TableName() string {
	return "payment_refund_recommendations"
}
