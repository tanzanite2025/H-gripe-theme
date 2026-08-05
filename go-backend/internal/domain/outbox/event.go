package outbox

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	EventStatusPending    = "pending"
	EventStatusProcessing = "processing"
	EventStatusProcessed  = "processed"
	EventStatusFailed     = "failed"
	EventStatusDeadLetter = "dead_letter"

	EventTypeOrderPaid               = "order.paid"
	EventTypeVerifiedConversion      = "conversion.verified"
	EventTypePaymentRiskLevelChanged = "payment.risk_level_changed"
	AggregateTypeOrder               = "order"
	AggregateTypePaymentRiskProvider = "payment_risk_provider"
	DefaultEventMaxAttempt           = 10
)

type Event struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	EventKey      string         `gorm:"size:160;uniqueIndex;not null" json:"event_key"`
	EventType     string         `gorm:"size:80;not null;index" json:"event_type"`
	AggregateType string         `gorm:"size:80;not null;index" json:"aggregate_type"`
	AggregateID   string         `gorm:"size:80;not null;index" json:"aggregate_id"`
	Payload       datatypes.JSON `gorm:"not null" json:"payload"`
	Status        string         `gorm:"size:20;not null;default:'pending';index" json:"status"`
	Attempts      int            `gorm:"not null;default:0" json:"attempts"`
	MaxAttempts   int            `gorm:"not null;default:10" json:"max_attempts"`
	AvailableAt   time.Time      `gorm:"not null;index" json:"available_at"`
	LockedAt      *time.Time     `gorm:"index" json:"locked_at"`
	LockedBy      string         `gorm:"size:128;index" json:"locked_by"`
	ProcessedAt   *time.Time     `gorm:"index" json:"processed_at"`
	LastError     string         `gorm:"type:text" json:"last_error"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Event) TableName() string {
	return "outbox_events"
}

func (e *Event) BeforeCreate(_ *gorm.DB) error {
	if e.Status == "" {
		e.Status = EventStatusPending
	}
	if e.MaxAttempts <= 0 {
		e.MaxAttempts = DefaultEventMaxAttempt
	}
	if len(e.Payload) == 0 {
		e.Payload = datatypes.JSON([]byte("{}"))
	}
	if e.AvailableAt.IsZero() {
		e.AvailableAt = time.Now().UTC()
	}
	return nil
}

type OrderPaidPayload struct {
	OrderID              uint      `json:"order_id"`
	OrderNumber          string    `json:"order_number"`
	UserID               uint      `json:"user_id"`
	PaymentTransactionID string    `json:"payment_transaction_id"`
	PaymentMethod        string    `json:"payment_method"`
	Amount               float64   `json:"amount"`
	Currency             string    `json:"currency"`
	PaidAt               time.Time `json:"paid_at"`
	CustomerEmail        string    `json:"customer_email,omitempty"`
	CustomerName         string    `json:"customer_name,omitempty"`
	ShippingCountry      string    `json:"shipping_country,omitempty"`
}

// VerifiedConversionPayload intentionally excludes customer contact details
// and is emitted only after a payment provider verification succeeds.
type VerifiedConversionPayload struct {
	OrderID     uint                           `json:"order_id"`
	Amount      float64                        `json:"amount"`
	Currency    string                         `json:"currency"`
	VerifiedAt  time.Time                      `json:"verified_at"`
	Attribution *VerifiedConversionAttribution `json:"attribution,omitempty"`
}

type VerifiedConversionAttribution struct {
	Source      string `json:"source,omitempty"`
	Medium      string `json:"medium,omitempty"`
	Campaign    string `json:"campaign,omitempty"`
	Term        string `json:"term,omitempty"`
	Content     string `json:"content,omitempty"`
	ClickIDKind string `json:"click_id_kind,omitempty"`
	ClickID     string `json:"click_id,omitempty"`
}
