package outbox

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	EventStatusPending    = "pending"
	EventStatusProcessing = "processing"
	EventStatusProcessed  = "processed"
	EventStatusFailed     = "failed"
	EventStatusUnknown    = "unknown"
	EventStatusDeadLetter = "dead_letter"

	EventTypeOrderPaid                        = "order.paid"
	EventTypeOrderConfirmationEmail           = "order.confirmation_email"
	EventTypeOrderShippingNotificationEmail   = "order.shipping_notification_email"
	EventTypeReferralOrderPaid                = "referral.order_paid"
	EventTypeReferralOrderDelivered           = "referral.order_delivered"
	EventTypeReferralOrderInvalidated         = "referral.order_invalidated"
	EventTypeVerifiedConversion               = "conversion.verified"
	EventTypePaymentRiskLevelChanged          = "payment.risk_level_changed"
	EventTypePaymentRiskFailOpen              = "payment.risk_fail_open"
	EventTypePaymentRefundPending             = "payment.refund_pending"
	EventTypePaymentRefundCompleted           = "payment.refund_completed"
	EventTypePaymentRefundFailed              = "payment.refund_failed"
	EventTypeMerchantProductUpsert            = "merchant.product_upsert"
	EventTypeMerchantProductWithdraw          = "merchant.product_withdraw"
	EventTypeMerchantOfferRevalidate          = "merchant.offer_revalidate"
	EventTypeProductCacheInvalidate           = "product.cache_invalidate"
	EventTypeCustomerServiceRealtime          = "customer_service.realtime"
	EventTypeCustomerServiceAvatarCleanup     = "customer_service.avatar_cleanup"
	EventTypeStorefrontRouteCatalogChanged    = "storefront.route_catalog_changed"
	AggregateTypeOrder                        = "order"
	AggregateTypePaymentRiskProvider          = "payment_risk_provider"
	AggregateTypePayment                      = "payment"
	AggregateTypeProduct                      = "product"
	AggregateTypeProductCache                 = "product_cache"
	AggregateTypeProductSpecificationTemplate = "product_specification_template"
	AggregateTypeProductBrand                 = "product_brand"
	AggregateTypeInformationTemplate          = "product_information_template"
	AggregateTypeMerchantOffer                = "merchant_offer"
	AggregateTypeCustomerServiceConversation  = "customer_service_conversation"
	AggregateTypeCustomerServiceAgentProfile  = "customer_service_agent_profile"
	AggregateTypeStorefrontRouteCatalogEntry  = "storefront_route_catalog_entry"
	DefaultEventMaxAttempt                    = 10
)

type Event struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	EventKey       string         `gorm:"size:160;uniqueIndex;not null" json:"event_key"`
	EventType      string         `gorm:"size:80;not null;index" json:"event_type"`
	AggregateType  string         `gorm:"size:80;not null;index" json:"aggregate_type"`
	AggregateID    string         `gorm:"size:80;not null;index" json:"aggregate_id"`
	Payload        datatypes.JSON `gorm:"not null" json:"payload"`
	Status         string         `gorm:"size:20;not null;default:'pending';index" json:"status"`
	Attempts       int            `gorm:"not null;default:0" json:"attempts"`
	MaxAttempts    int            `gorm:"not null;default:10" json:"max_attempts"`
	AvailableAt    time.Time      `gorm:"not null;index" json:"available_at"`
	LockedAt       *time.Time     `gorm:"index" json:"locked_at"`
	LockedBy       string         `gorm:"size:128;index" json:"locked_by"`
	ProcessedAt    *time.Time     `gorm:"index" json:"processed_at"`
	LastAttemptAt  *time.Time     `gorm:"index" json:"last_attempt_at"`
	UncertainAt    *time.Time     `gorm:"index" json:"uncertain_at"`
	ReconcileAfter *time.Time     `gorm:"index" json:"reconcile_after"`
	LastError      string         `gorm:"type:text" json:"last_error"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
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

type OrderConfirmationEmailPayload struct {
	RecipientEmail string    `json:"recipient_email"`
	CustomerName   string    `json:"customer_name,omitempty"`
	OrderID        uint      `json:"order_id"`
	OrderNumber    string    `json:"order_number"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`
	PaidAt         time.Time `json:"paid_at"`
}

type OrderShippingNotificationEmailPayload struct {
	RecipientEmail string    `json:"recipient_email"`
	CustomerName   string    `json:"customer_name,omitempty"`
	OrderID        uint      `json:"order_id"`
	OrderNumber    string    `json:"order_number"`
	CarrierName    string    `json:"carrier_name,omitempty"`
	TrackingNumber string    `json:"tracking_number"`
	TrackingURL    string    `json:"tracking_url,omitempty"`
	ShippedAt      time.Time `json:"shipped_at"`
}

type ReferralOrderPaidPayload struct {
	OrderID     uint      `json:"order_id"`
	UserID      uint      `json:"user_id"`
	AmountMinor int64     `json:"amount_minor"`
	Currency    string    `json:"currency"`
	PaidAt      time.Time `json:"paid_at"`
}

type ReferralOrderDeliveredPayload struct {
	OrderID     uint      `json:"order_id"`
	DeliveredAt time.Time `json:"delivered_at"`
	Source      string    `json:"source"`
}

type ReferralOrderInvalidatedPayload struct {
	OrderID    uint      `json:"order_id"`
	OccurredAt time.Time `json:"occurred_at"`
	Reason     string    `json:"reason"`
	Source     string    `json:"source"`
	Reference  string    `json:"reference,omitempty"`
}

// PaymentRefundPayload is the durable state-transition envelope for refund
// intents and executions. Amounts are always minor units; consumers must not
// infer currency scale from a floating-point value.
type PaymentRefundPayload struct {
	RefundID             uint      `json:"refund_id"`
	OrderID              uint      `json:"order_id"`
	TransactionID        uint      `json:"transaction_id"`
	Provider             string    `json:"provider,omitempty"`
	ProviderRefundID     string    `json:"provider_refund_id,omitempty"`
	RefundStatus         string    `json:"refund_status"`
	ExecutionStatus      string    `json:"execution_status,omitempty"`
	AmountMinor          int64     `json:"amount_minor"`
	GiftCardAmountMinor  int64     `json:"gift_card_amount_minor"`
	RequestedAmountMinor int64     `json:"requested_amount_minor"`
	Currency             string    `json:"currency"`
	Reason               string    `json:"reason,omitempty"`
	ErrorMessage         string    `json:"error_message,omitempty"`
	OccurredAt           time.Time `json:"occurred_at"`
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

type MerchantProductSyncPayload struct {
	ProductID uint   `json:"product_id"`
	Reason    string `json:"reason,omitempty"`
}

type MerchantOfferRevalidatePayload struct {
	OfferID uint   `json:"offer_id"`
	Reason  string `json:"reason,omitempty"`
}

type ProductCacheInvalidatePayload struct {
	ProductIDs                     []uint `json:"product_ids,omitempty"`
	ProductSpecificationTemplateID uint   `json:"product_specification_template_id,omitempty"`
	ProductBrandID                 uint   `json:"product_brand_id,omitempty"`
	ProductInformationTemplateID   uint   `json:"product_information_template_id,omitempty"`
	Reason                         string `json:"reason,omitempty"`
}

// CustomerServiceRealtimePayload is the durable, display-safe envelope passed
// from the customer-service write transaction to realtime delivery workers.
// HTTP history remains the authoritative source after clients receive it.
type CustomerServiceRealtimePayload struct {
	Type           string                       `json:"type"`
	EventID        string                       `json:"event_id"`
	Audience       string                       `json:"audience,omitempty"`
	TicketID       uint                         `json:"ticket_id"`
	ConversationID string                       `json:"conversation_id,omitempty"`
	OccurredAt     time.Time                    `json:"occurred_at"`
	Actor          CustomerServiceRealtimeActor `json:"actor"`
	Payload        json.RawMessage              `json:"payload,omitempty"`
}

type CustomerServiceRealtimeActor struct {
	Kind      string `json:"kind"`
	UserID    *uint  `json:"user_id,omitempty"`
	Anonymous bool   `json:"anonymous,omitempty"`
}

// CustomerServiceAvatarCleanupPayload contains only a first-party URL. The
// cleanup handler re-validates its dedicated storage namespace before delete.
type CustomerServiceAvatarCleanupPayload struct {
	URL string `json:"url"`
}
