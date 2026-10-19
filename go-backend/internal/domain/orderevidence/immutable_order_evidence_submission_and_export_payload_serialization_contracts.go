package orderevidence

import (
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/shipping"
	"encoding/json"
	"time"
)

// These payload contracts also preserve the original Go field order used by
// schema version one hashes before storage-independent hashing was introduced.
// Changes require a versioned compatibility contract; service DTOs alias these types.
type DisputeEvidenceSnapshotEnvelope struct {
	SchemaVersion          int                            `json:"schema_version"`
	Provider               string                         `json:"provider"`
	LocalDisputeID         uint                           `json:"local_dispute_id"`
	ExternalDisputeID      string                         `json:"external_dispute_id"`
	OrderID                uint                           `json:"order_id"`
	EvidencePackageID      *uint                          `json:"evidence_package_id,omitempty"`
	EvidencePackageVersion int                            `json:"evidence_package_version"`
	EvidencePackageStatus  string                         `json:"evidence_package_status,omitempty"`
	AssembledAt            time.Time                      `json:"assembled_at"`
	Completeness           EvidenceCompleteness           `json:"completeness"`
	TrackingContext        *OrderEvidenceTrackingContext  `json:"tracking_context,omitempty"`
	Sources                []OrderEvidenceSourceReference `json:"sources"`
	EvidenceChecklist      DisputeEvidenceChecklist       `json:"evidence_checklist"`
	Warnings               []string                       `json:"warnings"`
}

type StripeDisputeEvidenceSnapshotRequest struct {
	IncludeCustomerCommunication bool                       `json:"include_customer_communication"`
	AdditionalStatement          string                     `json:"additional_statement,omitempty"`
	ShippingDocumentationFileID  string                     `json:"shipping_documentation_file_id,omitempty"`
	CustomerCommunicationFileID  string                     `json:"customer_communication_file_id,omitempty"`
	ReceiptFileID                string                     `json:"receipt_file_id,omitempty"`
	UncategorizedFileID          string                     `json:"uncategorized_file_id,omitempty"`
	Evidence                     StripeDisputeEvidenceDraft `json:"evidence"`
}

type StripeDisputeEvidenceSnapshotPayload struct {
	DisputeEvidenceSnapshotEnvelope
	Request StripeDisputeEvidenceSnapshotRequest `json:"request"`
}

type PayPalDisputeEvidenceSnapshotRequest struct {
	AdditionalStatement string                          `json:"additional_statement,omitempty"`
	OverrideWarnings    bool                            `json:"override_warnings"`
	Evidence            PayPalDisputeEvidenceDraft      `json:"evidence"`
	Notes               string                          `json:"notes"`
	TrackingCarrierName string                          `json:"tracking_carrier_name,omitempty"`
	TrackingNumber      string                          `json:"tracking_number,omitempty"`
	Documents           []PayPalDisputeEvidenceDocument `json:"documents,omitempty"`
	DocumentWarnings    []string                        `json:"document_warnings,omitempty"`
	SubmissionWarnings  []string                        `json:"submission_warnings,omitempty"`
}

type PayPalDisputeEvidenceSnapshotPayload struct {
	DisputeEvidenceSnapshotEnvelope
	Request PayPalDisputeEvidenceSnapshotRequest `json:"request"`
}

type OrderEvidenceExportSnapshotPayload struct {
	SchemaVersion   int                            `json:"schema_version"`
	ExportType      string                         `json:"export_type"`
	ExportedAt      time.Time                      `json:"exported_at"`
	Order           *order.Order                   `json:"order"`
	Package         *OrderEvidenceExportPackage    `json:"package"`
	Completeness    EvidenceCompleteness           `json:"completeness"`
	TrackingContext *OrderEvidenceTrackingContext  `json:"tracking_context,omitempty"`
	TrackingEvents  []shipping.TrackingEvent       `json:"tracking_events"`
	Sources         []OrderEvidenceSourceReference `json:"sources"`
	Warnings        []string                       `json:"warnings,omitempty"`
}

type OrderEvidenceExportPackage struct {
	ID                         uint                      `json:"id"`
	OrderID                    uint                      `json:"order_id"`
	SnapshotID                 uint                      `json:"snapshot_id"`
	PackageVersion             int                       `json:"package_version"`
	Status                     string                    `json:"status"`
	OrderTotalUSDSnapshotMinor int64                     `json:"order_total_usd_snapshot_minor"`
	IsHighValue                bool                      `json:"is_high_value"`
	HasSpokeTensionQC          bool                      `json:"has_spoke_tension_qc"`
	SchemaVersion              int                       `json:"schema_version"`
	CreatedBy                  uint                      `json:"created_by"`
	LockedAt                   *time.Time                `json:"locked_at,omitempty"`
	CreatedAt                  time.Time                 `json:"created_at"`
	UpdatedAt                  time.Time                 `json:"updated_at"`
	Items                      []OrderEvidenceExportItem `json:"items"`
}

type OrderEvidenceExportItem struct {
	ID             uint                            `json:"id"`
	PackageID      uint                            `json:"package_id"`
	OrderID        uint                            `json:"order_id"`
	OrderItemID    *uint                           `json:"order_item_id,omitempty"`
	SnapshotID     *uint                           `json:"snapshot_id,omitempty"`
	ItemType       string                          `json:"item_type"`
	Status         string                          `json:"status"`
	RequiredReason string                          `json:"required_reason"`
	DataJSON       json.RawMessage                 `json:"data_json"`
	CapturedAt     *time.Time                      `json:"captured_at,omitempty"`
	CapturedBy     uint                            `json:"captured_by"`
	ContentSHA256  string                          `json:"content_sha256"`
	Attachments    []OrderEvidenceExportAttachment `json:"attachments"`
	CreatedAt      time.Time                       `json:"created_at"`
	UpdatedAt      time.Time                       `json:"updated_at"`
}

type OrderEvidenceExportAttachment struct {
	ID               uint      `json:"id"`
	EvidenceItemID   uint      `json:"evidence_item_id"`
	OriginalFilename string    `json:"original_filename"`
	MimeType         string    `json:"mime_type"`
	SizeBytes        int64     `json:"size_bytes"`
	SHA256           string    `json:"sha256"`
	UploadedBy       uint      `json:"uploaded_by"`
	CreatedAt        time.Time `json:"created_at"`
}

type OrderEvidenceTrackingContext struct {
	Shipments            []OrderEvidenceTrackingShipment `json:"shipments,omitempty"`
	LatestDeliveryEvents []OrderEvidenceDeliveryEvent    `json:"latest_delivery_events,omitempty"`
	ProviderPODURLs      []string                        `json:"provider_pod_urls,omitempty"`
	ManualPOD            *OrderEvidenceManualPOD         `json:"manual_pod,omitempty"`
}

type OrderEvidenceTrackingShipment struct {
	ID                  uint       `json:"id"`
	TrackingNumber      string     `json:"tracking_number"`
	ProviderCarrierCode string     `json:"provider_carrier_code"`
	ProviderCode        string     `json:"provider_code,omitempty"`
	ProviderName        string     `json:"provider_name,omitempty"`
	CarrierName         string     `json:"carrier_name,omitempty"`
	CarrierServiceName  string     `json:"carrier_service_name,omitempty"`
	RegistrationStatus  string     `json:"registration_status"`
	SyncStatus          string     `json:"sync_status"`
	EventCount          int        `json:"event_count"`
	LastEventAt         *time.Time `json:"last_event_at,omitempty"`
	LastSyncedAt        *time.Time `json:"last_synced_at,omitempty"`
	Enabled             bool       `json:"enabled"`
}

type OrderEvidenceDeliveryEvent struct {
	ID                     uint      `json:"id"`
	TrackingNumber         string    `json:"tracking_number"`
	Status                 string    `json:"status"`
	Location               string    `json:"location,omitempty"`
	Description            string    `json:"description,omitempty"`
	RecipientSignatureName string    `json:"recipient_signature_name,omitempty"`
	ProofOfDeliveryURL     string    `json:"proof_of_delivery_url,omitempty"`
	EventTime              time.Time `json:"event_time"`
}

type OrderEvidenceManualPOD struct {
	ItemID                    uint       `json:"item_id"`
	Status                    string     `json:"status"`
	TrackingNumber            string     `json:"tracking_number,omitempty"`
	TrackingAssociationStatus string     `json:"tracking_association_status"`
	CapturedAt                *time.Time `json:"captured_at,omitempty"`
	CapturedBy                uint       `json:"captured_by,omitempty"`
	AttachmentCount           int        `json:"attachment_count"`
}

type OrderEvidenceSourceReference struct {
	SourceType       string     `json:"source_type"`
	SourceID         uint       `json:"source_id"`
	OrderItemID      *uint      `json:"order_item_id,omitempty"`
	EvidenceItemType string     `json:"evidence_item_type,omitempty"`
	Status           string     `json:"status,omitempty"`
	ContentSHA256    string     `json:"content_sha256,omitempty"`
	CapturedAt       *time.Time `json:"captured_at,omitempty"`
}

type DisputeEvidenceChecklist struct {
	Items               []DisputeEvidenceChecklistItem `json:"items"`
	ReadyCount          int                            `json:"ready_count"`
	TotalCount          int                            `json:"total_count"`
	MissingCount        int                            `json:"missing_count"`
	ManualRequiredCount int                            `json:"manual_required_count"`
	UnavailableCount    int                            `json:"unavailable_count"`
	BlockerCount        int                            `json:"blocker_count"`
	Complete            bool                           `json:"complete"`
	LastEvaluatedAt     time.Time                      `json:"last_evaluated_at"`
}

type DisputeEvidenceChecklistItem struct {
	Key            string     `json:"key"`
	Title          string     `json:"title"`
	ProviderField  string     `json:"provider_field,omitempty"`
	Status         string     `json:"status"`
	Required       bool       `json:"required"`
	Blocker        bool       `json:"blocker"`
	ManualRequired bool       `json:"manual_required"`
	Source         string     `json:"source"`
	ObservedAt     *time.Time `json:"observed_at,omitempty"`
	Summary        string     `json:"summary"`
	MissingReason  string     `json:"missing_reason,omitempty"`
}

type StripeDisputeEvidenceDraft struct {
	CustomerName           string `json:"customer_name"`
	CustomerEmailAddress   string `json:"customer_email_address"`
	BillingAddress         string `json:"billing_address"`
	ShippingAddress        string `json:"shipping_address"`
	ProductDescription     string `json:"product_description"`
	ShippingCarrier        string `json:"shipping_carrier"`
	ShippingDate           string `json:"shipping_date"`
	ShippingTrackingNumber string `json:"shipping_tracking_number"`
	UncategorizedText      string `json:"uncategorized_text"`
	CommunicationSummary   string `json:"communication_summary"`
}

type PayPalDisputeEvidenceDraft struct {
	CustomerName           string `json:"customer_name"`
	CustomerEmailAddress   string `json:"customer_email_address"`
	ShippingAddress        string `json:"shipping_address"`
	ProductDescription     string `json:"product_description"`
	ShippingCarrier        string `json:"shipping_carrier"`
	ShippingDate           string `json:"shipping_date"`
	ShippingTrackingNumber string `json:"shipping_tracking_number"`
	DeliveredAt            string `json:"delivered_at"`
	InvoiceSummary         string `json:"invoice_summary"`
	ProofOfDeliverySummary string `json:"proof_of_delivery_summary"`
	CommunicationSummary   string `json:"communication_summary"`
	Notes                  string `json:"notes"`
}

type PayPalDisputeEvidenceDocument struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256,omitempty"`
}
