package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	orderdomain "tanzanite/internal/domain/order"
	paymentdomain "tanzanite/internal/domain/payment"
	shippingdomain "tanzanite/internal/domain/shipping"
	ticketdomain "tanzanite/internal/domain/ticket"
	"tanzanite/internal/repository"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/dispute"
)

var (
	ErrStripeDisputeEvidenceNotSubmittable  = errors.New("stripe dispute evidence is not submittable")
	ErrStripeDisputeEvidenceConfirmRequired = errors.New("confirm is required before submitting dispute evidence")
)

type stripeDisputeEvidenceSubmitter interface {
	Update(id string, params *stripe.DisputeParams) (*stripe.Dispute, error)
}

type liveStripeDisputeEvidenceSubmitter struct {
	apiKey string
}

func (s liveStripeDisputeEvidenceSubmitter) Update(id string, params *stripe.DisputeParams) (*stripe.Dispute, error) {
	client := dispute.Client{
		B:   stripe.GetBackend(stripe.APIBackend),
		Key: s.apiKey,
	}
	return client.Update(id, params)
}

type StripeDisputeEvidencePackage struct {
	Dispute        *paymentdomain.StripeDispute         `json:"dispute"`
	Order          *orderdomain.Order                   `json:"order,omitempty"`
	Shipment       *shippingdomain.TrackingShipment     `json:"shipment,omitempty"`
	TrackingEvents []shippingdomain.TrackingEvent       `json:"tracking_events"`
	Communications []StripeDisputeCommunicationEvidence `json:"communications"`
	Evidence       StripeDisputeEvidenceDraft           `json:"evidence"`
	Warnings       []string                             `json:"warnings"`
	CanSubmit      bool                                 `json:"can_submit"`
}

type StripeDisputeCommunicationEvidence struct {
	ID          uint      `json:"id"`
	TicketID    uint      `json:"ticket_id"`
	Sender      string    `json:"sender"`
	IsStaff     bool      `json:"is_staff"`
	MessageType string    `json:"message_type"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
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

type SubmitStripeDisputeEvidenceInput struct {
	DisputeID                    uint
	APIKey                       string
	Confirm                      bool
	Submit                       bool
	IncludeCustomerCommunication bool
	AdditionalStatement          string
	ShippingDocumentationFileID  string
	CustomerCommunicationFileID  string
	ReceiptFileID                string
	UncategorizedFileID          string
}

type SubmitStripeDisputeEvidenceResult struct {
	DisputeID       uint       `json:"dispute_id"`
	StripeDisputeID string     `json:"stripe_dispute_id"`
	StripeStatus    string     `json:"stripe_status"`
	SubmittedAt     *time.Time `json:"submitted_at,omitempty"`
	Staged          bool       `json:"staged"`
}

type stripeDisputeEvidenceSubmissionAudit struct {
	Submit                       bool                       `json:"submit"`
	IncludeCustomerCommunication bool                       `json:"include_customer_communication"`
	AdditionalStatement          string                     `json:"additional_statement,omitempty"`
	ShippingDocumentationFileID  string                     `json:"shipping_documentation_file_id,omitempty"`
	CustomerCommunicationFileID  string                     `json:"customer_communication_file_id,omitempty"`
	ReceiptFileID                string                     `json:"receipt_file_id,omitempty"`
	UncategorizedFileID          string                     `json:"uncategorized_file_id,omitempty"`
	Evidence                     StripeDisputeEvidenceDraft `json:"evidence"`
	SubmittedAt                  time.Time                  `json:"submitted_at"`
}

func (s *PaymentService) BuildStripeDisputeEvidencePackage(disputeID uint) (*StripeDisputeEvidencePackage, error) {
	record, err := s.GetStripeDispute(disputeID)
	if err != nil {
		return nil, err
	}

	pkg := &StripeDisputeEvidencePackage{
		Dispute:        record,
		TrackingEvents: []shippingdomain.TrackingEvent{},
		Communications: []StripeDisputeCommunicationEvidence{},
		Warnings:       []string{},
		CanSubmit:      disputeNeedsResponse(record.Status),
	}

	if record.OrderID == nil || s.orderRepo == nil {
		pkg.Warnings = append(pkg.Warnings, "Stripe dispute is not linked to a local order yet.")
		pkg.Evidence = buildStripeDisputeEvidenceDraft(pkg)
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
		pkg.Warnings = append(pkg.Warnings, "No local tracking events are available. Sync the shipment before submitting evidence.")
	} else if deliveredTrackingEvent(pkg.TrackingEvents) == nil {
		pkg.Warnings = append(pkg.Warnings, "Tracking events do not contain a clear delivered event.")
	}
	if len(pkg.Communications) == 0 {
		pkg.Warnings = append(pkg.Warnings, "No linked customer communication was found by order number, customer account, or order email.")
	}
	pkg.Warnings = append(pkg.Warnings, "Carrier official proof-of-delivery PDF is not configured yet. Upload the DHL/FedEx PDF to Stripe as dispute_evidence and paste its File ID before submitting.")

	pkg.Evidence = buildStripeDisputeEvidenceDraft(pkg)
	return pkg, nil
}

func (s *PaymentService) SubmitStripeDisputeEvidence(ctx context.Context, input SubmitStripeDisputeEvidenceInput) (*SubmitStripeDisputeEvidenceResult, error) {
	if !input.Confirm {
		return nil, ErrStripeDisputeEvidenceConfirmRequired
	}
	if strings.TrimSpace(input.APIKey) == "" {
		return nil, errors.New("stripe api key is required")
	}
	pkg, err := s.BuildStripeDisputeEvidencePackage(input.DisputeID)
	if err != nil {
		return nil, err
	}
	if !pkg.CanSubmit {
		return nil, ErrStripeDisputeEvidenceNotSubmittable
	}

	now := time.Now().UTC()
	params := stripeDisputeEvidenceParams(pkg, input)
	params.Context = ctx
	params.Submit = stripe.Bool(input.Submit)
	params.Metadata = map[string]string{
		"tanzanite_dispute_id": fmt.Sprint(pkg.Dispute.ID),
		"tanzanite_order_id":   stripeDisputeEvidenceOrderID(pkg.Order),
	}

	audit := stripeDisputeEvidenceSubmissionAudit{
		Submit:                       input.Submit,
		IncludeCustomerCommunication: input.IncludeCustomerCommunication,
		AdditionalStatement:          strings.TrimSpace(input.AdditionalStatement),
		ShippingDocumentationFileID:  strings.TrimSpace(input.ShippingDocumentationFileID),
		CustomerCommunicationFileID:  strings.TrimSpace(input.CustomerCommunicationFileID),
		ReceiptFileID:                strings.TrimSpace(input.ReceiptFileID),
		UncategorizedFileID:          strings.TrimSpace(input.UncategorizedFileID),
		Evidence:                     pkg.Evidence,
		SubmittedAt:                  now,
	}
	payloadBytes, _ := json.Marshal(audit)
	payload := string(payloadBytes)

	submitter := s.stripeDisputeEvidenceSubmitter
	if submitter == nil {
		submitter = liveStripeDisputeEvidenceSubmitter{apiKey: strings.TrimSpace(input.APIKey)}
	}

	updated, err := submitter.Update(pkg.Dispute.StripeDisputeID, params)
	if err != nil {
		_ = s.paymentRepo.UpdateStripeDisputeEvidenceSubmission(pkg.Dispute.ID, nil, payload, err.Error(), "")
		return nil, err
	}

	status := pkg.Dispute.Status
	if updated != nil && updated.Status != "" {
		status = string(updated.Status)
	}
	var submittedAt *time.Time
	if input.Submit {
		submittedAt = &now
	}
	if err := s.paymentRepo.UpdateStripeDisputeEvidenceSubmission(pkg.Dispute.ID, submittedAt, payload, "", status); err != nil {
		return nil, err
	}

	return &SubmitStripeDisputeEvidenceResult{
		DisputeID:       pkg.Dispute.ID,
		StripeDisputeID: pkg.Dispute.StripeDisputeID,
		StripeStatus:    status,
		SubmittedAt:     submittedAt,
		Staged:          !input.Submit,
	}, nil
}

func stripeDisputeEvidenceParams(pkg *StripeDisputeEvidencePackage, input SubmitStripeDisputeEvidenceInput) *stripe.DisputeParams {
	draft := pkg.Evidence
	if input.IncludeCustomerCommunication && strings.TrimSpace(draft.CommunicationSummary) != "" {
		draft.UncategorizedText = joinEvidenceSections(draft.UncategorizedText, "Customer communication summary:\n"+draft.CommunicationSummary)
	}
	if strings.TrimSpace(input.AdditionalStatement) != "" {
		draft.UncategorizedText = joinEvidenceSections(draft.UncategorizedText, "Operator statement:\n"+strings.TrimSpace(input.AdditionalStatement))
	}

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
	setStripeString(&evidence.ShippingDocumentation, input.ShippingDocumentationFileID)
	setStripeString(&evidence.CustomerCommunication, input.CustomerCommunicationFileID)
	setStripeString(&evidence.Receipt, input.ReceiptFileID)
	setStripeString(&evidence.UncategorizedFile, input.UncategorizedFileID)

	return &stripe.DisputeParams{Evidence: evidence}
}

func buildStripeDisputeEvidenceDraft(pkg *StripeDisputeEvidencePackage) StripeDisputeEvidenceDraft {
	if pkg == nil || pkg.Order == nil {
		return StripeDisputeEvidenceDraft{}
	}
	orderRecord := pkg.Order
	draft := StripeDisputeEvidenceDraft{
		CustomerName:           disputeCustomerName(orderRecord),
		CustomerEmailAddress:   disputeCustomerEmail(orderRecord),
		BillingAddress:         formatDisputeAddress(orderRecord.BillingAddress),
		ShippingAddress:        formatDisputeAddress(orderRecord.ShippingAddress),
		ProductDescription:     disputeProductDescription(orderRecord),
		ShippingCarrier:        disputeShippingCarrier(orderRecord, pkg.Shipment),
		ShippingDate:           disputeShippingDate(orderRecord, pkg.TrackingEvents),
		ShippingTrackingNumber: disputeTrackingNumber(orderRecord, pkg.Shipment),
		CommunicationSummary:   disputeCommunicationSummary(pkg.Communications),
	}
	draft.UncategorizedText = disputeUncategorizedText(orderRecord, pkg.Dispute, pkg.TrackingEvents)
	return draft
}

func disputeProductDescription(orderRecord *orderdomain.Order) string {
	if orderRecord == nil {
		return ""
	}
	lines := make([]string, 0, len(orderRecord.Items)+2)
	lines = append(lines, fmt.Sprintf("Order %s for physical bicycle components / carbon wheelset products.", orderRecord.OrderNumber))
	for _, item := range orderRecord.Items {
		if strings.TrimSpace(item.ProductName) == "" {
			continue
		}
		sku := strings.TrimSpace(item.SKU)
		if sku != "" {
			sku = " SKU: " + sku
		}
		lines = append(lines, fmt.Sprintf("- %s%s x%d, line total %.2f", item.ProductName, sku, item.Quantity, item.Total))
	}
	lines = append(lines, fmt.Sprintf("Order total: %.2f.", orderRecord.TotalAmount))
	return truncateEvidenceText(strings.Join(lines, "\n"), 20000)
}

func disputeUncategorizedText(orderRecord *orderdomain.Order, disputeRecord *paymentdomain.StripeDispute, events []shippingdomain.TrackingEvent) string {
	if orderRecord == nil {
		return ""
	}
	lines := []string{
		fmt.Sprintf("Local order number: %s", orderRecord.OrderNumber),
		fmt.Sprintf("Local order ID: %d", orderRecord.ID),
		fmt.Sprintf("Order status: %s; payment status: %s; shipping status: %s.", orderRecord.Status, orderRecord.PaymentStatus, orderRecord.ShippingStatus),
		fmt.Sprintf("Order created at: %s", orderRecord.CreatedAt.UTC().Format(time.RFC3339)),
	}
	if orderRecord.PaidAt != nil {
		lines = append(lines, fmt.Sprintf("Paid at: %s", orderRecord.PaidAt.UTC().Format(time.RFC3339)))
	}
	if orderRecord.ShippedAt != nil {
		lines = append(lines, fmt.Sprintf("Shipped at: %s", orderRecord.ShippedAt.UTC().Format(time.RFC3339)))
	}
	if orderRecord.CompletedAt != nil {
		lines = append(lines, fmt.Sprintf("Completed at: %s", orderRecord.CompletedAt.UTC().Format(time.RFC3339)))
	}
	if disputeRecord != nil {
		lines = append(lines, fmt.Sprintf("Stripe dispute reason: %s; disputed amount: %.2f %s.", disputeRecord.Reason, disputeRecord.Amount, disputeRecord.Currency))
	}
	if delivered := deliveredTrackingEvent(events); delivered != nil {
		lines = append(lines, fmt.Sprintf("Delivered tracking event: %s | %s | %s | %s", delivered.EventTime.UTC().Format(time.RFC3339), delivered.Status, delivered.Location, delivered.Description))
	}
	if len(events) > 0 {
		lines = append(lines, "Tracking timeline:")
		for _, event := range limitTrackingEvents(events, 12) {
			lines = append(lines, fmt.Sprintf("- %s | %s | %s | %s", event.EventTime.UTC().Format(time.RFC3339), event.Status, event.Location, event.Description))
		}
	}
	return truncateEvidenceText(strings.Join(lines, "\n"), 20000)
}

func disputeCommunicationEvidence(messages []ticketdomain.TicketMessage) []StripeDisputeCommunicationEvidence {
	items := make([]StripeDisputeCommunicationEvidence, 0, len(messages))
	for _, message := range messages {
		content := strings.TrimSpace(message.Content)
		if content == "" || message.IsInternal {
			continue
		}
		items = append(items, StripeDisputeCommunicationEvidence{
			ID:          message.ID,
			TicketID:    message.TicketID,
			Sender:      disputeMessageSender(message),
			IsStaff:     message.IsStaff,
			MessageType: strings.TrimSpace(message.MessageType),
			Content:     truncateEvidenceText(content, 1000),
			CreatedAt:   message.CreatedAt,
		})
	}
	return items
}

func disputeCommunicationSummary(items []StripeDisputeCommunicationEvidence) string {
	if len(items) == 0 {
		return ""
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("- %s | %s | %s", item.CreatedAt.UTC().Format(time.RFC3339), item.Sender, truncateEvidenceText(item.Content, 500)))
	}
	return truncateEvidenceText(strings.Join(lines, "\n"), 12000)
}

func disputeMessageSender(message ticketdomain.TicketMessage) string {
	if message.IsStaff {
		return "Tanzanite support"
	}
	if message.User != nil {
		name := strings.TrimSpace(strings.TrimSpace(message.User.FirstName) + " " + strings.TrimSpace(message.User.LastName))
		if name != "" {
			return name
		}
		if strings.TrimSpace(message.User.Email) != "" {
			return strings.TrimSpace(message.User.Email)
		}
	}
	return "Customer"
}

func disputeCustomerName(orderRecord *orderdomain.Order) string {
	if orderRecord == nil {
		return ""
	}
	name := strings.TrimSpace(strings.TrimSpace(orderRecord.ShippingAddress.FirstName) + " " + strings.TrimSpace(orderRecord.ShippingAddress.LastName))
	if name != "" {
		return name
	}
	return strings.TrimSpace(strings.TrimSpace(orderRecord.BillingAddress.FirstName) + " " + strings.TrimSpace(orderRecord.BillingAddress.LastName))
}

func disputeCustomerEmail(orderRecord *orderdomain.Order) string {
	for _, email := range disputeOrderEmails(orderRecord) {
		if strings.TrimSpace(email) != "" {
			return strings.TrimSpace(email)
		}
	}
	return ""
}

func disputeOrderEmails(orderRecord *orderdomain.Order) []string {
	if orderRecord == nil {
		return nil
	}
	emails := []string{
		strings.ToLower(strings.TrimSpace(orderRecord.ShippingAddress.Email)),
		strings.ToLower(strings.TrimSpace(orderRecord.BillingAddress.Email)),
	}
	seen := map[string]bool{}
	result := make([]string, 0, len(emails))
	for _, email := range emails {
		if email == "" || seen[email] {
			continue
		}
		seen[email] = true
		result = append(result, email)
	}
	return result
}

func formatDisputeAddress(address orderdomain.Address) string {
	parts := []string{
		strings.TrimSpace(strings.TrimSpace(address.FirstName) + " " + strings.TrimSpace(address.LastName)),
		strings.TrimSpace(address.Company),
		strings.TrimSpace(address.Address1),
		strings.TrimSpace(address.Address2),
		strings.TrimSpace(address.City),
		strings.TrimSpace(address.State),
		strings.TrimSpace(address.PostalCode),
		strings.TrimSpace(address.Country),
	}
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return strings.Join(result, ", ")
}

func disputeShippingCarrier(orderRecord *orderdomain.Order, shipment *shippingdomain.TrackingShipment) string {
	if shipment != nil {
		if shipment.Carrier != nil && strings.TrimSpace(shipment.Carrier.Name) != "" {
			return strings.TrimSpace(shipment.Carrier.Name)
		}
		if shipment.Mapping != nil && strings.TrimSpace(shipment.Mapping.ProviderCarrierName) != "" {
			return strings.TrimSpace(shipment.Mapping.ProviderCarrierName)
		}
		if strings.TrimSpace(shipment.ProviderCarrierCode) != "" {
			return strings.TrimSpace(shipment.ProviderCarrierCode)
		}
	}
	if orderRecord == nil {
		return ""
	}
	if strings.TrimSpace(orderRecord.ProviderCarrierName) != "" {
		return strings.TrimSpace(orderRecord.ProviderCarrierName)
	}
	return strings.TrimSpace(orderRecord.ProviderCarrierCode)
}

func disputeShippingDate(orderRecord *orderdomain.Order, events []shippingdomain.TrackingEvent) string {
	if orderRecord != nil && orderRecord.ShippedAt != nil {
		return orderRecord.ShippedAt.UTC().Format("2006-01-02")
	}
	if len(events) == 0 {
		return ""
	}
	oldest := events[0].EventTime
	for _, event := range events {
		if event.EventTime.Before(oldest) {
			oldest = event.EventTime
		}
	}
	if oldest.IsZero() {
		return ""
	}
	return oldest.UTC().Format("2006-01-02")
}

func disputeTrackingNumber(orderRecord *orderdomain.Order, shipment *shippingdomain.TrackingShipment) string {
	if shipment != nil && strings.TrimSpace(shipment.TrackingNumber) != "" {
		return strings.TrimSpace(shipment.TrackingNumber)
	}
	if orderRecord == nil {
		return ""
	}
	return strings.TrimSpace(orderRecord.TrackingNumber)
}

func deliveredTrackingEvent(events []shippingdomain.TrackingEvent) *shippingdomain.TrackingEvent {
	for i := range events {
		status := strings.ToLower(events[i].Status + " " + events[i].Description)
		if strings.Contains(status, "delivered") || strings.Contains(status, "signed") {
			return &events[i]
		}
	}
	return nil
}

func limitTrackingEvents(events []shippingdomain.TrackingEvent, limit int) []shippingdomain.TrackingEvent {
	if limit <= 0 || len(events) <= limit {
		return events
	}
	return events[:limit]
}

func setStripeString(target **string, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	*target = stripe.String(value)
}

func joinEvidenceSections(parts ...string) string {
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return strings.Join(result, "\n\n")
}

func truncateEvidenceText(value string, limit int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if limit <= 0 || len(runes) <= limit {
		return value
	}
	if limit <= 20 {
		return string(runes[:limit])
	}
	return string(runes[:limit-14]) + "\n[truncated]"
}

func stripeDisputeEvidenceOrderID(orderRecord *orderdomain.Order) string {
	if orderRecord == nil {
		return ""
	}
	return fmt.Sprint(orderRecord.ID)
}
