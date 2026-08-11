package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/invoice"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/repository"

	paypalapi "github.com/plutov/paypal/v4"
)

var (
	ErrPayPalDisputeEvidenceNotSubmittable = errors.New("paypal dispute evidence is not submittable")
	ErrPayPalDisputeEvidenceTrackingNeeded = errors.New("paypal dispute evidence requires tracking information")
	ErrPayPalDisputeEvidenceConfigRequired = errors.New("paypal client id and secret are required")
	ErrPayPalDisputeInvoiceUnavailable     = errors.New("paypal dispute commercial invoice is unavailable")
)

type PayPalDisputeEvidenceSubmitter interface {
	ProvideEvidence(ctx context.Context, disputeID string, params *paypalapi.DisputeProvideEvidenceParams) error
}

type PayPalDisputeEvidenceDocumentStorage interface {
	UploadFromReader(ctx context.Context, reader io.Reader, filename string) (string, error)
}

type PayPalDisputeInvoiceSellerProfileProvider interface {
	SellerProfile() (invoice.SellerProfile, error)
}

type livePayPalDisputeEvidenceSubmitter struct {
	config pgateway.Config
}

func (s livePayPalDisputeEvidenceSubmitter) ProvideEvidence(ctx context.Context, disputeID string, params *paypalapi.DisputeProvideEvidenceParams) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(s.config.APIKey) == "" || strings.TrimSpace(s.config.SecretKey) == "" {
		return ErrPayPalDisputeEvidenceConfigRequired
	}
	apiBase := paypalapi.APIBaseSandBox
	if strings.EqualFold(s.config.Environment, "production") {
		apiBase = paypalapi.APIBaseLive
	}
	client, err := paypalapi.NewClient(strings.TrimSpace(s.config.APIKey), strings.TrimSpace(s.config.SecretKey), apiBase)
	if err != nil {
		return err
	}
	client.SetHTTPClient(&http.Client{Timeout: 15 * time.Second})
	return client.DisputeProvideEvidence(ctx, disputeID, params)
}

type PayPalDisputeEvidencePackage struct {
	Dispute        *paymentdomain.PayPalDispute         `json:"dispute"`
	Order          *orderdomain.Order                   `json:"order,omitempty"`
	Shipment       *shippingdomain.TrackingShipment     `json:"shipment,omitempty"`
	TrackingEvents []shippingdomain.TrackingEvent       `json:"tracking_events"`
	Communications []StripeDisputeCommunicationEvidence `json:"communications"`
	Evidence       PayPalDisputeEvidenceDraft           `json:"evidence"`
	Documents      []PayPalDisputeEvidenceDocument      `json:"documents"`
	Warnings       []string                             `json:"warnings"`
	CanSubmit      bool                                 `json:"can_submit"`
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
	Type string `json:"type"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type SubmitPayPalDisputeEvidenceInput struct {
	DisputeID           uint
	ClientID            string
	SecretKey           string
	Environment         string
	AdditionalStatement string
}

type SubmitPayPalDisputeEvidenceResult struct {
	DisputeID       uint                            `json:"dispute_id"`
	PayPalDisputeID string                          `json:"paypal_dispute_id"`
	PayPalStatus    string                          `json:"paypal_status"`
	SubmittedAt     *time.Time                      `json:"submitted_at,omitempty"`
	TrackingNumber  string                          `json:"tracking_number"`
	EvidenceTypes   []string                        `json:"evidence_types"`
	Documents       []PayPalDisputeEvidenceDocument `json:"documents"`
}

type payPalDisputeEvidenceSubmissionAudit struct {
	AdditionalStatement string                          `json:"additional_statement,omitempty"`
	Evidence            PayPalDisputeEvidenceDraft      `json:"evidence"`
	EvidenceType        string                          `json:"evidence_type"`
	Documents           []PayPalDisputeEvidenceDocument `json:"documents,omitempty"`
	DocumentWarnings    []string                        `json:"document_warnings,omitempty"`
	SubmittedAt         time.Time                       `json:"submitted_at"`
}

func (s *PaymentService) BuildPayPalDisputeEvidencePackage(disputeID uint) (*PayPalDisputeEvidencePackage, error) {
	record, err := s.GetPayPalDispute(disputeID)
	if err != nil {
		return nil, err
	}

	pkg := &PayPalDisputeEvidencePackage{
		Dispute:        record,
		TrackingEvents: []shippingdomain.TrackingEvent{},
		Communications: []StripeDisputeCommunicationEvidence{},
		Warnings:       []string{},
		CanSubmit:      paypalDisputeNeedsSellerEvidence(record) && record.EvidenceSubmittedAt == nil,
	}

	if record.OrderID == nil || s.orderRepo == nil {
		pkg.Warnings = append(pkg.Warnings, "PayPal dispute is not linked to a local order yet.")
		pkg.Evidence = buildPayPalDisputeEvidenceDraft(pkg)
		return pkg, nil
	}

	orderRecord, err := s.orderRepo.FindByID(*record.OrderID)
	if err != nil {
		return nil, err
	}
	pkg.Order = orderRecord

	if s.shippingRepo != nil {
		if shipment, err := s.shippingRepo.FindTrackingShipmentByOrderID(orderRecord.ID); err == nil {
			pkg.Shipment = shipment
		} else if !repository.IsRecordNotFound(err) {
			return nil, err
		}
		events, err := s.shippingRepo.FindTrackingEventsByOrderID(orderRecord.ID)
		if err != nil {
			return nil, err
		}
		pkg.TrackingEvents = events
	}

	if s.ticketRepo != nil {
		messages, err := s.ticketRepo.FindDisputeCandidateMessages(repository.DisputeCommunicationFilter{
			UserID:      orderRecord.UserID,
			Emails:      disputeOrderEmails(orderRecord),
			OrderNumber: orderRecord.OrderNumber,
			Limit:       80,
		})
		if err != nil {
			return nil, err
		}
		pkg.Communications = disputeCommunicationEvidence(messages)
	}

	if strings.TrimSpace(orderRecord.TrackingNumber) == "" && pkg.Shipment == nil {
		pkg.Warnings = append(pkg.Warnings, "No tracking number is available on the order.")
	}
	if len(pkg.TrackingEvents) == 0 {
		pkg.Warnings = append(pkg.Warnings, "No local tracking events are available. Sync the shipment before submitting PayPal evidence.")
	} else if deliveredTrackingEvent(pkg.TrackingEvents) == nil {
		pkg.Warnings = append(pkg.Warnings, "Tracking events do not contain a clear delivered or signed event.")
	}
	if len(pkg.Communications) == 0 {
		pkg.Warnings = append(pkg.Warnings, "No linked customer communication was found by order number, customer account, or order email.")
	}
	pkg.Warnings = append(pkg.Warnings, "Carrier official proof-of-delivery PDF attachment is not configured yet. PayPal will receive structured tracking, signature-event, invoice, and communication notes.")

	pkg.Evidence = buildPayPalDisputeEvidenceDraft(pkg)
	return pkg, nil
}

func (s *PaymentService) SubmitPayPalDisputeEvidence(ctx context.Context, input SubmitPayPalDisputeEvidenceInput) (*SubmitPayPalDisputeEvidenceResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	pkg, err := s.BuildPayPalDisputeEvidencePackage(input.DisputeID)
	if err != nil {
		return nil, err
	}
	if !pkg.CanSubmit {
		return nil, ErrPayPalDisputeEvidenceNotSubmittable
	}

	now := time.Now().UTC()
	documents := []PayPalDisputeEvidenceDocument{}
	documentWarnings := []string{}
	if strings.TrimSpace(pkg.Evidence.ShippingTrackingNumber) != "" {
		if s.paypalDisputeInvoiceAutoAttachEnabled() {
			documents, documentWarnings = s.paypalDisputeEvidenceDocuments(ctx, pkg, now)
			if len(documentWarnings) > 0 {
				pkg.Warnings = append(pkg.Warnings, documentWarnings...)
			}
			pkg.Documents = documents
		} else {
			documentWarnings = append(documentWarnings, "Commercial invoice PDF auto-attachment is not enabled in the payment service configuration; structured PayPal evidence was submitted without the PDF document.")
			pkg.Warnings = append(pkg.Warnings, documentWarnings...)
		}
	}
	params := paypalDisputeEvidenceParams(pkg, input, documents)
	audit := payPalDisputeEvidenceSubmissionAudit{
		AdditionalStatement: strings.TrimSpace(input.AdditionalStatement),
		Evidence:            pkg.Evidence,
		EvidenceType:        string(paypalapi.EvidenceTypeProofOfFulfillment),
		Documents:           documents,
		DocumentWarnings:    documentWarnings,
		SubmittedAt:         now,
	}
	payloadBytes, _ := json.Marshal(audit)
	payload := string(payloadBytes)

	if strings.TrimSpace(pkg.Evidence.ShippingTrackingNumber) == "" {
		_ = s.paymentRepo.UpdatePayPalDisputeEvidenceSubmission(pkg.Dispute.ID, nil, payload, ErrPayPalDisputeEvidenceTrackingNeeded.Error(), "")
		return nil, ErrPayPalDisputeEvidenceTrackingNeeded
	}

	submitter := s.paypalDisputeEvidenceSubmitter
	if submitter == nil {
		if strings.TrimSpace(input.ClientID) == "" || strings.TrimSpace(input.SecretKey) == "" {
			_ = s.paymentRepo.UpdatePayPalDisputeEvidenceSubmission(pkg.Dispute.ID, nil, payload, ErrPayPalDisputeEvidenceConfigRequired.Error(), "")
			return nil, ErrPayPalDisputeEvidenceConfigRequired
		}
		submitter = livePayPalDisputeEvidenceSubmitter{config: pgateway.Config{
			Type:        pgateway.GatewayPayPal,
			APIKey:      strings.TrimSpace(input.ClientID),
			SecretKey:   strings.TrimSpace(input.SecretKey),
			Environment: strings.TrimSpace(input.Environment),
		}}
	}

	if err := submitter.ProvideEvidence(ctx, pkg.Dispute.PayPalDisputeID, params); err != nil {
		_ = s.paymentRepo.UpdatePayPalDisputeEvidenceSubmission(pkg.Dispute.ID, nil, payload, err.Error(), "")
		return nil, err
	}

	submittedAt := now
	if err := s.paymentRepo.UpdatePayPalDisputeEvidenceSubmission(pkg.Dispute.ID, &submittedAt, payload, "", ""); err != nil {
		return nil, err
	}
	return &SubmitPayPalDisputeEvidenceResult{
		DisputeID:       pkg.Dispute.ID,
		PayPalDisputeID: pkg.Dispute.PayPalDisputeID,
		PayPalStatus:    pkg.Dispute.Status,
		SubmittedAt:     &submittedAt,
		TrackingNumber:  pkg.Evidence.ShippingTrackingNumber,
		EvidenceTypes:   []string{string(paypalapi.EvidenceTypeProofOfFulfillment)},
		Documents:       documents,
	}, nil
}

func paypalDisputeEvidenceParams(pkg *PayPalDisputeEvidencePackage, input SubmitPayPalDisputeEvidenceInput, documents []PayPalDisputeEvidenceDocument) *paypalapi.DisputeProvideEvidenceParams {
	draft := pkg.Evidence
	notes := paypalDisputeEvidenceNotes(pkg, draft, input.AdditionalStatement)
	trackingInfo := []*paypalapi.TrackingInfo{}
	if strings.TrimSpace(draft.ShippingTrackingNumber) != "" {
		trackingInfo = append(trackingInfo, &paypalapi.TrackingInfo{
			CarrierName:    paypalCarrierName(draft.ShippingCarrier),
			TrackingNumber: strings.TrimSpace(draft.ShippingTrackingNumber),
		})
	}

	return &paypalapi.DisputeProvideEvidenceParams{
		Evidences: &paypalapi.DisputeEvidence{
			EvidenceType: paypalapi.EvidenceTypeProofOfFulfillment,
			Documents:    paypalEvidenceDocuments(documents),
			Notes:        truncateEvidenceText(notes, 4000),
			EvidenceInfo: &paypalapi.DisputeEvidenceInfo{
				TrackingInfo: trackingInfo,
			},
		},
	}
}

func (s *PaymentService) paypalDisputeEvidenceDocuments(ctx context.Context, pkg *PayPalDisputeEvidencePackage, generatedAt time.Time) ([]PayPalDisputeEvidenceDocument, []string) {
	documents := []PayPalDisputeEvidenceDocument{}
	warnings := []string{}
	if pkg == nil || pkg.Order == nil {
		return documents, warnings
	}
	if s == nil || s.paypalDisputeDocumentStorage == nil {
		warnings = append(warnings, "Commercial invoice PDF storage is not configured; no invoice document was attached.")
		return documents, warnings
	}

	document, options, err := s.paypalDisputeCommercialInvoice(pkg, generatedAt)
	if err != nil {
		warnings = append(warnings, "Commercial invoice PDF was not generated: "+err.Error())
		return documents, warnings
	}
	pdfBytes, err := invoice.RenderCommercialInvoicePDF(document, options.FontPath)
	if err != nil {
		warnings = append(warnings, "Commercial invoice PDF was not rendered: "+err.Error())
		return documents, warnings
	}

	name := paypalEvidenceDocumentName(document.DocumentNumber, "commercial-invoice")
	uploadedURL, err := s.paypalDisputeDocumentStorage.UploadFromReader(ctx, bytes.NewReader(pdfBytes), name)
	if err != nil {
		warnings = append(warnings, "Commercial invoice PDF upload failed: "+err.Error())
		return documents, warnings
	}
	if !paypalEvidenceDocumentURLUsable(uploadedURL) {
		warnings = append(warnings, "Commercial invoice PDF URL is not a public HTTPS URL usable by PayPal; no invoice document was attached.")
		return documents, warnings
	}

	documents = append(documents, PayPalDisputeEvidenceDocument{
		Type: "commercial_invoice",
		Name: name,
		URL:  uploadedURL,
	})
	return documents, warnings
}

func paypalEvidenceDocuments(documents []PayPalDisputeEvidenceDocument) []*paypalapi.Document {
	result := make([]*paypalapi.Document, 0, len(documents))
	for _, document := range documents {
		if strings.TrimSpace(document.Name) == "" || strings.TrimSpace(document.URL) == "" {
			continue
		}
		result = append(result, &paypalapi.Document{
			Name: strings.TrimSpace(document.Name),
			URL:  strings.TrimSpace(document.URL),
		})
	}
	return result
}

func paypalEvidenceDocumentName(documentNumber, prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "paypal-evidence"
	}
	documentNumber = strings.TrimSpace(documentNumber)
	if documentNumber == "" {
		return prefix + ".pdf"
	}
	replacer := strings.NewReplacer("/", "-", "\\", "-", " ", "-")
	clean := replacer.Replace(documentNumber)
	clean = strings.Trim(clean, "-_.")
	if clean == "" {
		return prefix + ".pdf"
	}
	return prefix + "-" + clean + ".pdf"
}

func paypalEvidenceDocumentURLUsable(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed == nil {
		return false
	}
	if !strings.EqualFold(parsed.Scheme, "https") || parsed.Host == "" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsLinkLocalUnicast()
	}
	return true
}

func buildPayPalDisputeEvidenceDraft(pkg *PayPalDisputeEvidencePackage) PayPalDisputeEvidenceDraft {
	if pkg == nil || pkg.Order == nil {
		return PayPalDisputeEvidenceDraft{}
	}
	orderRecord := pkg.Order
	draft := PayPalDisputeEvidenceDraft{
		CustomerName:           disputeCustomerName(orderRecord),
		CustomerEmailAddress:   disputeCustomerEmail(orderRecord),
		ShippingAddress:        formatDisputeAddress(orderRecord.ShippingAddress),
		ProductDescription:     disputeProductDescription(orderRecord),
		ShippingCarrier:        paypalDisputeShippingCarrier(orderRecord, pkg.Shipment),
		ShippingDate:           disputeShippingDate(orderRecord, pkg.TrackingEvents),
		ShippingTrackingNumber: disputeTrackingNumber(orderRecord, pkg.Shipment),
		InvoiceSummary:         paypalDisputeInvoiceSummary(orderRecord),
		CommunicationSummary:   disputeCommunicationSummary(pkg.Communications),
	}
	if delivered := deliveredTrackingEvent(pkg.TrackingEvents); delivered != nil {
		draft.DeliveredAt = delivered.EventTime.UTC().Format(time.RFC3339)
		draft.ProofOfDeliverySummary = fmt.Sprintf("%s | %s | %s | %s", draft.DeliveredAt, delivered.Status, delivered.Location, delivered.Description)
	}
	draft.Notes = paypalDisputeEvidenceNotes(pkg, draft, "")
	return draft
}

func paypalDisputeEvidenceNotes(pkg *PayPalDisputeEvidencePackage, draft PayPalDisputeEvidenceDraft, additionalStatement string) string {
	if pkg == nil || pkg.Order == nil {
		return ""
	}
	orderRecord := pkg.Order
	lines := []string{
		fmt.Sprintf("Evidence package for PayPal dispute %s.", pkg.Dispute.PayPalDisputeID),
		fmt.Sprintf("Local order number: %s; local order ID: %d.", orderRecord.OrderNumber, orderRecord.ID),
		fmt.Sprintf("Order status: %s; payment status: %s; shipping status: %s.", orderRecord.Status, orderRecord.PaymentStatus, orderRecord.ShippingStatus),
		fmt.Sprintf("Invoice summary: %s", draft.InvoiceSummary),
		fmt.Sprintf("Customer: %s <%s>.", draft.CustomerName, draft.CustomerEmailAddress),
		fmt.Sprintf("Ship-to address: %s", draft.ShippingAddress),
		fmt.Sprintf("Product details: %s", draft.ProductDescription),
		fmt.Sprintf("Fulfillment: carrier=%s; tracking_number=%s; shipping_date=%s.", draft.ShippingCarrier, draft.ShippingTrackingNumber, draft.ShippingDate),
	}
	if draft.ProofOfDeliverySummary != "" {
		lines = append(lines, "Proof of delivery / signature event: "+draft.ProofOfDeliverySummary)
	}
	if len(pkg.TrackingEvents) > 0 {
		lines = append(lines, "Tracking timeline:")
		for _, event := range limitTrackingEvents(pkg.TrackingEvents, 12) {
			lines = append(lines, fmt.Sprintf("- %s | %s | %s | %s", event.EventTime.UTC().Format(time.RFC3339), event.Status, event.Location, event.Description))
		}
	}
	if draft.CommunicationSummary != "" {
		lines = append(lines, "Customer communication summary:")
		lines = append(lines, draft.CommunicationSummary)
	}
	if strings.TrimSpace(additionalStatement) != "" {
		lines = append(lines, "Operator statement:")
		lines = append(lines, strings.TrimSpace(additionalStatement))
	}
	return truncateEvidenceText(strings.Join(lines, "\n"), 4000)
}

func paypalDisputeInvoiceSummary(orderRecord *orderdomain.Order) string {
	if orderRecord == nil {
		return ""
	}
	lines := []string{
		fmt.Sprintf("Invoice/order %s total %.2f %s.", orderRecord.OrderNumber, orderRecord.TotalAmount, orderRecord.Currency),
		fmt.Sprintf("Subtotal %.2f; shipping %.2f; tax %.2f; discount %.2f.", orderRecord.SubtotalAmount, orderRecord.ShippingFee, orderRecord.TaxAmount, orderRecord.DiscountAmount),
	}
	if orderRecord.PaidAt != nil {
		lines = append(lines, fmt.Sprintf("Paid at %s.", orderRecord.PaidAt.UTC().Format(time.RFC3339)))
	}
	for _, item := range orderRecord.Items {
		if strings.TrimSpace(item.ProductName) == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s SKU %s x%d line total %.2f.", item.ProductName, item.SKU, item.Quantity, item.Total))
	}
	return truncateEvidenceText(strings.Join(lines, " "), 1200)
}

func paypalDisputeShippingCarrier(orderRecord *orderdomain.Order, shipment *shippingdomain.TrackingShipment) string {
	if shipment != nil {
		if strings.TrimSpace(shipment.ProviderCarrierCode) != "" {
			return strings.TrimSpace(shipment.ProviderCarrierCode)
		}
		if shipment.Mapping != nil && strings.TrimSpace(shipment.Mapping.ProviderCarrierName) != "" {
			return strings.TrimSpace(shipment.Mapping.ProviderCarrierName)
		}
	}
	if orderRecord != nil && strings.TrimSpace(orderRecord.ProviderCarrierCode) != "" {
		return strings.TrimSpace(orderRecord.ProviderCarrierCode)
	}
	return disputeShippingCarrier(orderRecord, shipment)
}

func paypalCarrierName(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	replacer := strings.NewReplacer(" ", "_", "-", "_")
	return replacer.Replace(value)
}

func paypalDisputeNeedsSellerEvidence(record *paymentdomain.PayPalDispute) bool {
	if record == nil {
		return false
	}
	status := strings.ToUpper(strings.TrimSpace(record.Status))
	state := strings.ToUpper(strings.TrimSpace(record.DisputeState))
	switch status {
	case "RESOLVED", "UNDER_REVIEW", "WAITING_FOR_BUYER_RESPONSE":
		return false
	case "WAITING_FOR_SELLER_RESPONSE", "OPEN":
		return true
	}
	switch state {
	case "REQUIRED_ACTION", "OPEN_INQUIRIES":
		return true
	default:
		return status == ""
	}
}
