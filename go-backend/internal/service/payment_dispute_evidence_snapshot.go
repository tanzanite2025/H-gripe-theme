package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/orderevidence"

	paypalapi "github.com/plutov/paypal/v4"
	"github.com/stripe/stripe-go/v76"
)

var (
	ErrOrderEvidenceSubmissionSnapshotStoreUnavailable = errors.New("order evidence submission snapshot store is unavailable")
	ErrOrderEvidenceSubmissionSnapshotInvalid          = errors.New("order evidence submission snapshot is invalid")
)

type disputeEvidenceSnapshotEnvelope struct {
	SchemaVersion          int                                `json:"schema_version"`
	Provider               string                             `json:"provider"`
	LocalDisputeID         uint                               `json:"local_dispute_id"`
	ExternalDisputeID      string                             `json:"external_dispute_id"`
	OrderID                uint                               `json:"order_id"`
	EvidencePackageID      *uint                              `json:"evidence_package_id,omitempty"`
	EvidencePackageVersion int                                `json:"evidence_package_version"`
	EvidencePackageStatus  string                             `json:"evidence_package_status,omitempty"`
	AssembledAt            time.Time                          `json:"assembled_at"`
	Completeness           orderevidence.EvidenceCompleteness `json:"completeness"`
	TrackingContext        *OrderEvidenceTrackingContext      `json:"tracking_context,omitempty"`
	Sources                []OrderEvidenceSourceReference     `json:"sources"`
	EvidenceChecklist      DisputeEvidenceChecklist           `json:"evidence_checklist"`
	Warnings               []string                           `json:"warnings"`
}

type stripeDisputeEvidenceSnapshotRequest struct {
	IncludeCustomerCommunication bool                       `json:"include_customer_communication"`
	AdditionalStatement          string                     `json:"additional_statement,omitempty"`
	ShippingDocumentationFileID  string                     `json:"shipping_documentation_file_id,omitempty"`
	CustomerCommunicationFileID  string                     `json:"customer_communication_file_id,omitempty"`
	ReceiptFileID                string                     `json:"receipt_file_id,omitempty"`
	UncategorizedFileID          string                     `json:"uncategorized_file_id,omitempty"`
	Evidence                     StripeDisputeEvidenceDraft `json:"evidence"`
}

type stripeDisputeEvidenceSnapshotPayload struct {
	disputeEvidenceSnapshotEnvelope
	Request stripeDisputeEvidenceSnapshotRequest `json:"request"`
}

type paypalDisputeEvidenceSnapshotRequest struct {
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

type paypalDisputeEvidenceSnapshotPayload struct {
	disputeEvidenceSnapshotEnvelope
	Request paypalDisputeEvidenceSnapshotRequest `json:"request"`
}

func (s *PaymentService) findLatestEvidenceSubmissionSnapshot(
	provider string,
	disputeID uint,
) (*orderevidence.OrderEvidenceSubmissionSnapshot, error) {
	if s == nil || s.orderEvidenceSubmissionRepo == nil {
		return nil, ErrOrderEvidenceSubmissionSnapshotStoreUnavailable
	}
	return s.orderEvidenceSubmissionRepo.FindLatestByDispute(provider, disputeID)
}

func buildDisputeEvidenceSnapshotEnvelope(
	provider string,
	disputeID uint,
	externalDisputeID string,
	orderID uint,
	assembly *OrderEvidencePackageAssembly,
	checklist DisputeEvidenceChecklist,
	warnings []string,
) disputeEvidenceSnapshotEnvelope {
	envelope := disputeEvidenceSnapshotEnvelope{
		SchemaVersion:     orderevidence.OrderEvidenceSubmissionSnapshotSchemaVersion,
		Provider:          strings.ToLower(strings.TrimSpace(provider)),
		LocalDisputeID:    disputeID,
		ExternalDisputeID: strings.TrimSpace(externalDisputeID),
		OrderID:           orderID,
		AssembledAt:       time.Now().UTC(),
		Sources:           []OrderEvidenceSourceReference{},
		EvidenceChecklist: checklist,
		Warnings:          append([]string{}, warnings...),
	}
	if assembly == nil {
		return envelope
	}
	envelope.AssembledAt = assembly.AssembledAt.UTC()
	envelope.Completeness = assembly.Completeness
	envelope.TrackingContext = assembly.TrackingContext
	envelope.Sources = append(envelope.Sources, assembly.Sources...)
	if assembly.Package != nil {
		packageID := assembly.Package.ID
		envelope.EvidencePackageID = &packageID
		envelope.EvidencePackageVersion = assembly.Package.PackageVersion
		envelope.EvidencePackageStatus = assembly.Package.Status
	}
	return envelope
}

func buildLockedEvidenceSubmissionSnapshot(
	provider string,
	disputeID uint,
	externalDisputeID string,
	orderID uint,
	assembly *OrderEvidencePackageAssembly,
	checklist DisputeEvidenceChecklist,
	warnings []string,
	payload interface{},
	lockedAt time.Time,
) (*orderevidence.OrderEvidenceSubmissionSnapshot, error) {
	if lockedAt.IsZero() {
		lockedAt = time.Now().UTC()
	}
	lockedAt = lockedAt.UTC()
	envelope := buildDisputeEvidenceSnapshotEnvelope(
		provider,
		disputeID,
		externalDisputeID,
		orderID,
		assembly,
		checklist,
		warnings,
	)
	completePayload, err := wrapEvidenceSubmissionSnapshotPayload(envelope, payload)
	if err != nil {
		return nil, err
	}
	payloadBytes, err := json.Marshal(completePayload)
	if err != nil {
		return nil, fmt.Errorf("%w: marshal payload: %v", ErrOrderEvidenceSubmissionSnapshotInvalid, err)
	}
	hash := sha256.Sum256(payloadBytes)

	var evidencePackageID *uint
	evidencePackageVersion := 0
	if assembly != nil && assembly.Package != nil {
		packageID := assembly.Package.ID
		evidencePackageID = &packageID
		evidencePackageVersion = assembly.Package.PackageVersion
	}
	snapshot := &orderevidence.OrderEvidenceSubmissionSnapshot{
		Provider:               strings.ToLower(strings.TrimSpace(provider)),
		DisputeID:              disputeID,
		OrderID:                orderID,
		EvidencePackageID:      evidencePackageID,
		EvidencePackageVersion: evidencePackageVersion,
		Version:                1,
		Status:                 orderevidence.SubmissionSnapshotStatusLocked,
		SchemaVersion:          orderevidence.OrderEvidenceSubmissionSnapshotSchemaVersion,
		LockedAt:               lockedAt,
		SnapshotData:           payloadBytes,
		SnapshotSHA256:         hex.EncodeToString(hash[:]),
	}
	if err := snapshot.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOrderEvidenceSubmissionSnapshotInvalid, err)
	}
	return snapshot, nil
}

func wrapEvidenceSubmissionSnapshotPayload(
	envelope disputeEvidenceSnapshotEnvelope,
	request interface{},
) (interface{}, error) {
	switch value := request.(type) {
	case stripeDisputeEvidenceSnapshotRequest:
		if envelope.Provider != "stripe" {
			return nil, fmt.Errorf(
				"%w: Stripe request cannot be stored under provider %q",
				ErrOrderEvidenceSubmissionSnapshotInvalid,
				envelope.Provider,
			)
		}
		return stripeDisputeEvidenceSnapshotPayload{
			disputeEvidenceSnapshotEnvelope: envelope,
			Request:                         value,
		}, nil
	case paypalDisputeEvidenceSnapshotRequest:
		if envelope.Provider != "paypal" {
			return nil, fmt.Errorf(
				"%w: PayPal request cannot be stored under provider %q",
				ErrOrderEvidenceSubmissionSnapshotInvalid,
				envelope.Provider,
			)
		}
		return paypalDisputeEvidenceSnapshotPayload{
			disputeEvidenceSnapshotEnvelope: envelope,
			Request:                         value,
		}, nil
	default:
		return nil, fmt.Errorf(
			"%w: unsupported %s submission payload type %T",
			ErrOrderEvidenceSubmissionSnapshotInvalid,
			strings.TrimSpace(envelope.Provider),
			request,
		)
	}
}

func (s *PaymentService) createOrGetEvidenceSubmissionSnapshot(
	snapshot *orderevidence.OrderEvidenceSubmissionSnapshot,
) (*orderevidence.OrderEvidenceSubmissionSnapshot, error) {
	if s == nil || s.orderEvidenceSubmissionRepo == nil {
		return nil, ErrOrderEvidenceSubmissionSnapshotStoreUnavailable
	}
	result, err := s.orderEvidenceSubmissionRepo.CreateOrGetLatest(snapshot)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, ErrOrderEvidenceSubmissionSnapshotInvalid
	}
	if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOrderEvidenceSubmissionSnapshotInvalid, err)
	}
	return result, nil
}

func parseStripeDisputeEvidenceSnapshot(
	snapshot *orderevidence.OrderEvidenceSubmissionSnapshot,
	disputeID uint,
	externalDisputeID string,
	orderID uint,
) (*stripeDisputeEvidenceSnapshotPayload, error) {
	if snapshot == nil {
		return nil, ErrOrderEvidenceSubmissionSnapshotInvalid
	}
	if err := snapshot.Validate(); err != nil {
		return nil, fmt.Errorf("%w: validate Stripe snapshot: %v", ErrOrderEvidenceSubmissionSnapshotInvalid, err)
	}
	if snapshot.Provider != "stripe" || snapshot.DisputeID != disputeID || snapshot.OrderID != orderID {
		return nil, ErrOrderEvidenceSubmissionSnapshotInvalid
	}
	var payload stripeDisputeEvidenceSnapshotPayload
	if err := json.Unmarshal(snapshot.SnapshotData, &payload); err != nil {
		return nil, fmt.Errorf("%w: decode Stripe payload: %v", ErrOrderEvidenceSubmissionSnapshotInvalid, err)
	}
	if payload.Provider != "stripe" ||
		payload.LocalDisputeID != disputeID ||
		payload.OrderID != orderID ||
		strings.TrimSpace(payload.ExternalDisputeID) != strings.TrimSpace(externalDisputeID) {
		return nil, fmt.Errorf("%w: Stripe snapshot scope does not match dispute", ErrOrderEvidenceSubmissionSnapshotInvalid)
	}
	if payload.SchemaVersion != snapshot.SchemaVersion ||
		payload.EvidencePackageVersion != snapshot.EvidencePackageVersion {
		return nil, fmt.Errorf("%w: Stripe snapshot metadata does not match stored snapshot", ErrOrderEvidenceSubmissionSnapshotInvalid)
	}
	return &payload, nil
}

func parsePayPalDisputeEvidenceSnapshot(
	snapshot *orderevidence.OrderEvidenceSubmissionSnapshot,
	disputeID uint,
	externalDisputeID string,
	orderID uint,
) (*paypalDisputeEvidenceSnapshotPayload, error) {
	if snapshot == nil {
		return nil, ErrOrderEvidenceSubmissionSnapshotInvalid
	}
	if err := snapshot.Validate(); err != nil {
		return nil, fmt.Errorf("%w: validate PayPal snapshot: %v", ErrOrderEvidenceSubmissionSnapshotInvalid, err)
	}
	if snapshot.Provider != "paypal" || snapshot.DisputeID != disputeID || snapshot.OrderID != orderID {
		return nil, ErrOrderEvidenceSubmissionSnapshotInvalid
	}
	var payload paypalDisputeEvidenceSnapshotPayload
	if err := json.Unmarshal(snapshot.SnapshotData, &payload); err != nil {
		return nil, fmt.Errorf("%w: decode PayPal payload: %v", ErrOrderEvidenceSubmissionSnapshotInvalid, err)
	}
	if payload.Provider != "paypal" ||
		payload.LocalDisputeID != disputeID ||
		payload.OrderID != orderID ||
		strings.TrimSpace(payload.ExternalDisputeID) != strings.TrimSpace(externalDisputeID) {
		return nil, fmt.Errorf("%w: PayPal snapshot scope does not match dispute", ErrOrderEvidenceSubmissionSnapshotInvalid)
	}
	if payload.SchemaVersion != snapshot.SchemaVersion ||
		payload.EvidencePackageVersion != snapshot.EvidencePackageVersion {
		return nil, fmt.Errorf("%w: PayPal snapshot metadata does not match stored snapshot", ErrOrderEvidenceSubmissionSnapshotInvalid)
	}
	return &payload, nil
}

func stripeDisputeEvidenceSnapshotRequestFromPackage(
	pkg *StripeDisputeEvidencePackage,
	input SubmitStripeDisputeEvidenceInput,
) stripeDisputeEvidenceSnapshotRequest {
	draft := pkg.Evidence
	if input.IncludeCustomerCommunication && strings.TrimSpace(draft.CommunicationSummary) != "" {
		draft.UncategorizedText = joinEvidenceSections(
			draft.UncategorizedText,
			"Customer communication summary:\n"+draft.CommunicationSummary,
		)
	}
	if strings.TrimSpace(input.AdditionalStatement) != "" {
		draft.UncategorizedText = joinEvidenceSections(
			draft.UncategorizedText,
			"Operator statement:\n"+strings.TrimSpace(input.AdditionalStatement),
		)
	}
	return stripeDisputeEvidenceSnapshotRequest{
		IncludeCustomerCommunication: input.IncludeCustomerCommunication,
		AdditionalStatement:          strings.TrimSpace(input.AdditionalStatement),
		ShippingDocumentationFileID:  strings.TrimSpace(input.ShippingDocumentationFileID),
		CustomerCommunicationFileID:  strings.TrimSpace(input.CustomerCommunicationFileID),
		ReceiptFileID:                strings.TrimSpace(input.ReceiptFileID),
		UncategorizedFileID:          strings.TrimSpace(input.UncategorizedFileID),
		Evidence:                     draft,
	}
}

func stripeDisputeEvidenceParamsFromSnapshot(
	payload *stripeDisputeEvidenceSnapshotPayload,
) *stripe.DisputeParams {
	if payload == nil {
		return &stripe.DisputeParams{}
	}
	draft := payload.Request.Evidence
	evidence := &stripe.DisputeEvidenceParams{}
	setStripeString(&evidence.CustomerName, draft.CustomerName)
	setStripeString(&evidence.CustomerEmailAddress, draft.CustomerEmailAddress)
	setStripeString(&evidence.BillingAddress, draft.BillingAddress)
	setStripeString(&evidence.ShippingAddress, draft.ShippingAddress)
	setStripeString(&evidence.ProductDescription, draft.ProductDescription)
	setStripeString(&evidence.ShippingCarrier, draft.ShippingCarrier)
	setStripeString(&evidence.ShippingDate, draft.ShippingDate)
	setStripeString(&evidence.ShippingTrackingNumber, draft.ShippingTrackingNumber)
	setStripeString(&evidence.UncategorizedText, truncateEvidenceText(draft.UncategorizedText, 20000))
	setStripeString(&evidence.ShippingDocumentation, payload.Request.ShippingDocumentationFileID)
	setStripeString(&evidence.CustomerCommunication, payload.Request.CustomerCommunicationFileID)
	setStripeString(&evidence.Receipt, payload.Request.ReceiptFileID)
	setStripeString(&evidence.UncategorizedFile, payload.Request.UncategorizedFileID)
	return &stripe.DisputeParams{
		Evidence: evidence,
	}
}

func paypalDisputeEvidenceSnapshotRequestFromPackage(
	pkg *PayPalDisputeEvidencePackage,
	input SubmitPayPalDisputeEvidenceInput,
	documents []PayPalDisputeEvidenceDocument,
	documentWarnings []string,
) paypalDisputeEvidenceSnapshotRequest {
	draft := pkg.Evidence
	return paypalDisputeEvidenceSnapshotRequest{
		AdditionalStatement: strings.TrimSpace(input.AdditionalStatement),
		OverrideWarnings:    input.OverrideWarnings,
		Evidence:            draft,
		Notes:               paypalDisputeEvidenceNotes(pkg, draft, input.AdditionalStatement),
		TrackingCarrierName: paypalCarrierName(draft.ShippingCarrier),
		TrackingNumber:      strings.TrimSpace(draft.ShippingTrackingNumber),
		Documents:           append([]PayPalDisputeEvidenceDocument{}, documents...),
		DocumentWarnings:    append([]string{}, documentWarnings...),
		SubmissionWarnings:  append([]string{}, pkg.SubmissionCheck.Warnings...),
	}
}

func paypalDisputeEvidenceParamsFromSnapshot(
	payload *paypalDisputeEvidenceSnapshotPayload,
) *paypalapi.DisputeProvideEvidenceParams {
	if payload == nil {
		return &paypalapi.DisputeProvideEvidenceParams{}
	}
	trackingInfo := []*paypalapi.TrackingInfo{}
	if payload.Request.TrackingNumber != "" {
		trackingInfo = append(trackingInfo, &paypalapi.TrackingInfo{
			CarrierName:    payload.Request.TrackingCarrierName,
			TrackingNumber: payload.Request.TrackingNumber,
		})
	}
	return &paypalapi.DisputeProvideEvidenceParams{
		Evidences: &paypalapi.DisputeEvidence{
			EvidenceType: paypalapi.EvidenceTypeProofOfFulfillment,
			Documents:    paypalEvidenceDocuments(payload.Request.Documents),
			Notes:        truncateEvidenceText(payload.Request.Notes, 4000),
			EvidenceInfo: &paypalapi.DisputeEvidenceInfo{
				TrackingInfo: trackingInfo,
			},
		},
	}
}

func stripeDisputeEvidenceSubmissionAuditFromSnapshot(
	payload *stripeDisputeEvidenceSnapshotPayload,
	submittedAt time.Time,
) stripeDisputeEvidenceSubmissionAudit {
	if payload == nil {
		return stripeDisputeEvidenceSubmissionAudit{SubmittedAt: submittedAt}
	}
	return stripeDisputeEvidenceSubmissionAudit{
		IncludeCustomerCommunication: payload.Request.IncludeCustomerCommunication,
		AdditionalStatement:          payload.Request.AdditionalStatement,
		ShippingDocumentationFileID:  payload.Request.ShippingDocumentationFileID,
		CustomerCommunicationFileID:  payload.Request.CustomerCommunicationFileID,
		ReceiptFileID:                payload.Request.ReceiptFileID,
		UncategorizedFileID:          payload.Request.UncategorizedFileID,
		Evidence:                     payload.Request.Evidence,
		SubmittedAt:                  submittedAt,
	}
}

func paypalDisputeEvidenceSubmissionAuditFromSnapshot(
	payload *paypalDisputeEvidenceSnapshotPayload,
	submittedAt time.Time,
) payPalDisputeEvidenceSubmissionAudit {
	if payload == nil {
		return payPalDisputeEvidenceSubmissionAudit{SubmittedAt: submittedAt}
	}
	return payPalDisputeEvidenceSubmissionAudit{
		AdditionalStatement: payload.Request.AdditionalStatement,
		Evidence:            payload.Request.Evidence,
		EvidenceType:        string(paypalapi.EvidenceTypeProofOfFulfillment),
		Documents:           append([]PayPalDisputeEvidenceDocument{}, payload.Request.Documents...),
		DocumentWarnings:    append([]string{}, payload.Request.DocumentWarnings...),
		SubmissionWarnings:  append([]string{}, payload.Request.SubmissionWarnings...),
		OverrideWarnings:    payload.Request.OverrideWarnings,
		SubmittedAt:         submittedAt,
	}
}
