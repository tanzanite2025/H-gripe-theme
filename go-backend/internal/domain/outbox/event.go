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
	// EventStatusIgnored is a terminal operator decision. It is deliberately
	// distinct from processed: no external side effect is claimed to have
	// happened when an administrator suppresses a failed event.
	EventStatusIgnored = "ignored"

	EventTypeOrderPaid = "order.paid"
	// Canonical order facts are separate from legacy integration/email command
	// events. Consumers may subscribe to these facts without coupling to a
	// particular delivery channel.
	EventTypeOrderPaymentSucceeded            = "order.payment_succeeded"
	EventTypeOrderPaymentExpired              = "order.payment_expired"
	EventTypeOrderCancelled                   = "order.cancelled"
	EventTypeOrderShipped                     = "order.shipped"
	EventTypeOrderDelivered                   = "order.delivered"
	EventTypeOrderRefunded                    = "order.refunded"
	EventTypeAfterSalesStatusChanged          = "after_sales.status_changed"
	EventTypeOrderCompleted                   = "order.completed"
	EventTypeOrderDisputeContactEmail         = "order.dispute_contact_email"
	EventTypeEmailChallengeDelivery           = "email.challenge_delivery_requested"
	EventTypeTrackingShipmentRegistration     = "tracking.shipment_registration_requested"
	EventTypeReferralOrderPaid                = "referral.order_paid"
	EventTypeReferralOrderDelivered           = "referral.order_delivered"
	EventTypeReferralOrderInvalidated         = "referral.order_invalidated"
	EventTypeVerifiedConversion               = "conversion.verified"
	EventTypePaymentRiskLevelChanged          = "payment.risk_level_changed"
	EventTypePaymentRiskFailOpen              = "payment.risk_fail_open"
	EventTypePaymentRefundPending             = "payment.refund_pending"
	EventTypePaymentRefundCompleted           = "payment.refund_completed"
	EventTypePaymentRefundFailed              = "payment.refund_failed"
	EventTypePaymentRefundExecutionRequested  = "payment.refund_execution_requested"
	EventTypeMerchantProductUpsert            = "merchant.product_upsert"
	EventTypeMerchantProductWithdraw          = "merchant.product_withdraw"
	EventTypeMerchantOfferRevalidate          = "merchant.offer_revalidate"
	EventTypeProductCacheInvalidate           = "product.cache_invalidate"
	EventTypeCustomerServiceRealtime          = "customer_service.realtime"
	EventTypeCustomerServiceAvatarCleanup     = "customer_service.avatar_cleanup"
	EventTypeCustomerServiceRetentionCleanup  = "customer_service.retention_cleanup_requested"
	EventTypeObjectStorageCleanup             = "storage.object_cleanup_requested"
	EventTypeStorefrontRouteCatalogChanged    = "storefront.route_catalog_changed"
	AggregateTypeOrder                        = "order"
	AggregateTypeAfterSalesCase               = "after_sales_case"
	AggregateTypeEmailChallenge               = "email_challenge"
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
	AggregateTypeMediaAsset                   = "media_asset"
	AggregateTypeSiteLogo                     = "site_logo"
	AggregateTypeHomeVisualTileSet            = "home_visual_tile_set"
	AggregateTypeUGCShowcase                  = "ugc_showcase"
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
	AmountMinor          int64     `json:"amount_minor"`
	AmountDisplay        string    `json:"-"`
	Currency             string    `json:"currency"`
	PaidAt               time.Time `json:"paid_at"`
	CustomerEmail        string    `json:"customer_email,omitempty"`
	CustomerName         string    `json:"customer_name,omitempty"`
	ShippingCountry      string    `json:"shipping_country,omitempty"`
}

// CanonicalDomainEventMetadata is embedded in domain-fact payloads. The
// Outbox event key remains the durable idempotency key; keeping the same key in
// the payload makes replay/debug tooling independent from the storage model.
type CanonicalDomainEventMetadata struct {
	SchemaVersion  int       `json:"schema_version"`
	OccurredAt     time.Time `json:"occurred_at"`
	IdempotencyKey string    `json:"idempotency_key"`
}

// NotificationAudienceSnapshot is the immutable customer-facing identity
// captured with a canonical order fact when it is available. Locale is
// optional because legacy orders may not have a stored language preference;
// notification workers fall back to English in that case.
type NotificationAudienceSnapshot struct {
	RecipientEmail string `json:"recipient_email,omitempty"`
	Locale         string `json:"locale,omitempty"`
	CustomerName   string `json:"customer_name,omitempty"`
}

type OrderPaymentSucceededPayload struct {
	CanonicalDomainEventMetadata
	NotificationAudienceSnapshot
	OrderID               uint   `json:"order_id"`
	OrderNumber           string `json:"order_number"`
	PaymentTransactionID  string `json:"payment_transaction_id"`
	PreviousOrderStatus   string `json:"previous_order_status"`
	NewOrderStatus        string `json:"new_order_status"`
	PreviousPaymentStatus string `json:"previous_payment_status"`
	NewPaymentStatus      string `json:"new_payment_status"`
	AmountMinor           int64  `json:"amount_minor"`
	Currency              string `json:"currency"`
	PaymentMethod         string `json:"payment_method,omitempty"`
}

type OrderPaymentExpiredPayload struct {
	CanonicalDomainEventMetadata
	NotificationAudienceSnapshot
	OrderID               uint   `json:"order_id"`
	OrderNumber           string `json:"order_number"`
	PreviousOrderStatus   string `json:"previous_order_status"`
	NewOrderStatus        string `json:"new_order_status"`
	PreviousPaymentStatus string `json:"previous_payment_status"`
	NewPaymentStatus      string `json:"new_payment_status"`
}

type OrderCancelledPayload struct {
	CanonicalDomainEventMetadata
	NotificationAudienceSnapshot
	OrderID             uint   `json:"order_id"`
	OrderNumber         string `json:"order_number"`
	PreviousOrderStatus string `json:"previous_order_status"`
	NewOrderStatus      string `json:"new_order_status"`
	PaymentStatus       string `json:"payment_status"`
}

type OrderShippedPayload struct {
	CanonicalDomainEventMetadata
	NotificationAudienceSnapshot
	OrderID                uint      `json:"order_id"`
	OrderNumber            string    `json:"order_number"`
	ShipmentID             uint      `json:"shipment_id"`
	CarrierName            string    `json:"carrier_name,omitempty"`
	TrackingNumber         string    `json:"tracking_number"`
	TrackingURL            string    `json:"tracking_url,omitempty"`
	PreviousOrderStatus    string    `json:"previous_order_status"`
	NewOrderStatus         string    `json:"new_order_status"`
	PreviousShippingStatus string    `json:"previous_shipping_status"`
	NewShippingStatus      string    `json:"new_shipping_status"`
	ShippedAt              time.Time `json:"shipped_at"`
}

type OrderDeliveredPayload struct {
	CanonicalDomainEventMetadata
	NotificationAudienceSnapshot
	OrderID                uint      `json:"order_id"`
	OrderNumber            string    `json:"order_number"`
	TrackingNumber         string    `json:"tracking_number,omitempty"`
	PreviousShippingStatus string    `json:"previous_shipping_status"`
	NewShippingStatus      string    `json:"new_shipping_status"`
	DeliveredAt            time.Time `json:"delivered_at"`
	Source                 string    `json:"source,omitempty"`
}

type OrderRefundedPayload struct {
	CanonicalDomainEventMetadata
	NotificationAudienceSnapshot
	OrderID          uint   `json:"order_id"`
	OrderNumber      string `json:"order_number,omitempty"`
	RefundID         uint   `json:"refund_id"`
	TransactionID    uint   `json:"transaction_id"`
	Provider         string `json:"provider,omitempty"`
	ProviderRefundID string `json:"provider_refund_id,omitempty"`
	AmountMinor      int64  `json:"amount_minor"`
	Currency         string `json:"currency"`
	RefundStatus     string `json:"refund_status"`
}

type AfterSalesStatusChangedPayload struct {
	CanonicalDomainEventMetadata
	NotificationAudienceSnapshot
	CaseID           uint       `json:"case_id"`
	OrderID          uint       `json:"order_id"`
	OrderNumber      string     `json:"order_number,omitempty"`
	CaseType         string     `json:"case_type"`
	PreviousStatus   string     `json:"previous_status"`
	NewStatus        string     `json:"new_status"`
	TransitionID     uint       `json:"transition_id"`
	Resolution       string     `json:"resolution,omitempty"`
	UpdatedBy        uint       `json:"updated_by"`
	ReturnShipmentID uint       `json:"return_shipment_id,omitempty"`
	Carrier          string     `json:"carrier,omitempty"`
	TrackingNumber   string     `json:"tracking_number,omitempty"`
	TrackingURL      string     `json:"tracking_url,omitempty"`
	LabelURL         string     `json:"label_url,omitempty"`
	WarehouseName    string     `json:"warehouse_name,omitempty"`
	WarehouseAddress string     `json:"warehouse_address,omitempty"`
	ShippedAt        *time.Time `json:"shipped_at,omitempty"`
	ReceivedAt       *time.Time `json:"received_at,omitempty"`
}

// OrderDisputeContactEmailPayload is the durable command for a manually
// requested dispute contact email. The worker sends it only after the request
// transaction has committed, so SMTP latency cannot hold an order request open.
type OrderDisputeContactEmailPayload struct {
	OrderID           uint      `json:"order_id"`
	Provider          string    `json:"provider"`
	DisputeID         uint      `json:"dispute_id"`
	ProviderDisputeID string    `json:"provider_dispute_id"`
	RecipientEmail    string    `json:"recipient_email"`
	Subject           string    `json:"subject"`
	Body              string    `json:"body"`
	RequestedAt       time.Time `json:"requested_at"`
}

// EmailChallengeDeliveryPayload is the durable command for sending a
// one-time verification link after the challenge record has committed.
type EmailChallengeDeliveryPayload struct {
	RecipientEmail   string    `json:"recipient_email"`
	DeliverySubject  string    `json:"delivery_subject"`
	BodyTemplate     string    `json:"body_template"`
	Purpose          string    `json:"purpose"`
	ChallengeSubject string    `json:"challenge_subject"`
	Nonce            string    `json:"nonce"`
	ExpiresAt        time.Time `json:"expires_at"`
	RequestedAt      time.Time `json:"requested_at"`
}

// TrackingShipmentRegistrationPayload is the durable command for registering
// a parcel with an external tracking provider after fulfillment commits.
type TrackingShipmentRegistrationPayload struct {
	Version                  int       `json:"version"`
	OrderID                  uint      `json:"order_id"`
	TrackingProviderID       uint      `json:"tracking_provider_id"`
	TrackingNumber           string    `json:"tracking_number"`
	ProviderCarrierCode      string    `json:"provider_carrier_code"`
	CarrierID                *uint     `json:"carrier_id,omitempty"`
	CarrierServiceID         *uint     `json:"carrier_service_id,omitempty"`
	TrackingCarrierMappingID *uint     `json:"tracking_carrier_mapping_id,omitempty"`
	SourceFingerprint        string    `json:"source_fingerprint"`
	RequestedAt              time.Time `json:"requested_at"`
}

// OrderCompletedPayload is the durable trigger for post-completion
// settlement processors such as loyalty rewards and referral completion.
type OrderCompletedPayload struct {
	CanonicalDomainEventMetadata
	NotificationAudienceSnapshot
	OrderID     uint      `json:"order_id"`
	OrderNumber string    `json:"order_number"`
	CompletedAt time.Time `json:"completed_at"`
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
	RequestedAmountMinor int64     `json:"requested_amount_minor"`
	Currency             string    `json:"currency"`
	Reason               string    `json:"reason,omitempty"`
	ErrorMessage         string    `json:"error_message,omitempty"`
	OccurredAt           time.Time `json:"occurred_at"`
}

// PaymentRefundExecutionRequestedPayload is the durable command consumed by
// the refund execution worker. It contains no gateway credentials; the worker
// resolves those from the current server-side payment configuration.
type PaymentRefundExecutionRequestedPayload struct {
	RefundID       uint      `json:"refund_id"`
	AdminID        uint      `json:"admin_id"`
	Provider       string    `json:"provider"`
	Attempt        int       `json:"attempt"`
	IdempotencyKey string    `json:"idempotency_key"`
	RequestedAt    time.Time `json:"requested_at"`
}

// VerifiedConversionPayload intentionally excludes customer contact details
// and is emitted only after a payment provider verification succeeds.
type VerifiedConversionPayload struct {
	OrderID     uint                           `json:"order_id"`
	AmountMinor int64                          `json:"amount_minor"`
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

// CustomerServiceRetentionCleanupPayload contains the external cleanup work
// that must happen after a retention purge transaction commits. Attachment
// references are captured before ticket messages are deleted because they are
// no longer available to a post-commit worker.
type CustomerServiceRetentionCleanupPayload struct {
	TicketID             uint      `json:"ticket_id"`
	AttachmentReferences []string  `json:"attachment_references"`
	RequestedAt          time.Time `json:"requested_at"`
}

// ObjectStorageCleanupPayload describes durable, post-commit object cleanup.
// ObjectKeys are storage-provider keys rather than public URLs so workers do
// not need to persist or trust an origin that may change between retries.
type ObjectStorageCleanupPayload struct {
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	ObjectKeys   []string  `json:"object_keys,omitempty"`
	RequestedAt  time.Time `json:"requested_at"`
}
