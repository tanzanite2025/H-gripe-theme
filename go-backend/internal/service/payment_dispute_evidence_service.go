package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	domainmoney "commerce-platform/internal/domain/money"
	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	paymentdomain "commerce-platform/internal/domain/payment"
	shippingdomain "commerce-platform/internal/domain/shipping"
	ticketdomain "commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/repository"

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
	Dispute               *paymentdomain.StripeDispute          `json:"dispute"`
	Order                 *orderdomain.Order                    `json:"order,omitempty"`
	FulfillmentEvidence   *OrderEvidencePackageAssembly         `json:"fulfillment_evidence,omitempty"`
	PolicyDisclosure      *DisputePolicyDisclosureEvidence      `json:"policy_disclosure,omitempty"`
	Refunds               []DisputeRefundEvidence               `json:"refunds"`
	Shipments             []shippingdomain.TrackingShipment     `json:"-"`
	TrackingEvents        []shippingdomain.TrackingEvent        `json:"-"`
	TrackingContext       *OrderEvidenceTrackingContext         `json:"tracking_context,omitempty"`
	TrackingEventEvidence []OrderEvidenceDeliveryEvent          `json:"tracking_events"`
	Communications        []StripeDisputeCommunicationEvidence  `json:"communications"`
	Authentication        *DisputePaymentAuthenticationEvidence `json:"authentication,omitempty"`
	Evidence              StripeDisputeEvidenceDraft            `json:"evidence"`
	EvidenceChecklist     DisputeEvidenceChecklist              `json:"evidence_checklist"`
	SubmissionCheck       DisputeEvidenceSubmissionCheck        `json:"submission_check"`
	Warnings              []string                              `json:"warnings"`
	CanSubmit             bool                                  `json:"can_submit"`
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
	DisputeID               uint       `json:"dispute_id"`
	StripeDisputeID         string     `json:"stripe_dispute_id"`
	StripeStatus            string     `json:"stripe_status"`
	SubmittedAt             *time.Time `json:"submitted_at,omitempty"`
	Staged                  bool       `json:"staged"`
	EvidenceSnapshotID      uint       `json:"evidence_snapshot_id"`
	EvidenceSnapshotVersion int        `json:"evidence_snapshot_version"`
	EvidenceSnapshotSHA256  string     `json:"evidence_snapshot_sha256"`
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
	EvidenceSnapshotID           uint                       `json:"evidence_snapshot_id"`
	EvidenceSnapshotVersion      int                        `json:"evidence_snapshot_version"`
	EvidenceSnapshotSHA256       string                     `json:"evidence_snapshot_sha256"`
}

func (s *PaymentService) BuildStripeDisputeEvidencePackage(disputeID uint) (*StripeDisputeEvidencePackage, error) {
	record, err := s.GetStripeDispute(disputeID)
	if err != nil {
		return nil, err
	}

	pkg := &StripeDisputeEvidencePackage{
		Dispute:               record,
		Refunds:               []DisputeRefundEvidence{},
		TrackingEvents:        []shippingdomain.TrackingEvent{},
		TrackingEventEvidence: []OrderEvidenceDeliveryEvent{},
		Communications:        []StripeDisputeCommunicationEvidence{},
		Warnings:              []string{},
		CanSubmit:             disputeNeedsResponse(record.Status),
	}
	transaction, err := s.findDisputeEvidenceTransaction("stripe", record.TransactionID, record.PaymentIntentID, disputeOrderID(record.OrderID))
	if err != nil {
		pkg.Warnings = append(pkg.Warnings, "Stripe transaction authentication evidence is unavailable; verify the saved gateway response before submitting.")
	}
	pkg.Authentication = buildStripeAuthenticationEvidence(transaction, record.PaymentIntentID)

	if record.OrderID == nil || s.orderRepo == nil {
		pkg.Warnings = append(pkg.Warnings, "Stripe dispute is not linked to a local order yet.")
		finalizeStripeDisputeEvidencePackage(pkg)
		return pkg, nil
	}

	orderRecord, err := s.orderRepo.FindByID(*record.OrderID)
	if err != nil {
		return nil, err
	}
	pkg.Order = orderRecord
	if s.paymentRepo != nil {
		refunds, err := s.paymentRepo.FindRefundsByOrderID(orderRecord.ID)
		if err != nil {
			return nil, err
		}
		pkg.Refunds = buildRefundEvidence(refunds, orderRecord.Currency)
	}
	if s.policyDisclosureRepo != nil {
		disclosure, err := s.policyDisclosureRepo.FindByOrderID(orderRecord.ID)
		if err == nil {
			pkg.PolicyDisclosure = buildPolicyDisclosureEvidence(disclosure)
		} else if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}
	if pkg.PolicyDisclosure == nil {
		pkg.Warnings = append(pkg.Warnings, "No order-level refund and cancellation policy disclosure snapshot was found; the current policy page will not be used as historical evidence.")
	}

	fulfillmentEvidence, err := s.assembleOrderEvidencePackage(orderRecord.ID)
	if err != nil {
		return nil, err
	}
	if fulfillmentEvidence == nil {
		pkg.Warnings = append(pkg.Warnings, "Order evidence package assembler is not configured; fulfillment evidence was not loaded.")
	} else {
		pkg.FulfillmentEvidence = fulfillmentEvidence
		pkg.Shipments = fulfillmentEvidence.Shipments
		pkg.TrackingEvents = fulfillmentEvidence.TrackingEvents
		pkg.TrackingContext = fulfillmentEvidence.TrackingContext
		pkg.TrackingEventEvidence = projectOrderEvidenceDeliveryEvents(fulfillmentEvidence.TrackingEvents)
		pkg.Warnings = append(pkg.Warnings, fulfillmentEvidence.Warnings...)
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

	if len(pkg.Shipments) == 0 {
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

	finalizeStripeDisputeEvidencePackage(pkg)
	return pkg, nil
}

func (s *PaymentService) SubmitStripeDisputeEvidence(ctx context.Context, input SubmitStripeDisputeEvidenceInput) (*SubmitStripeDisputeEvidenceResult, error) {
	if !input.Confirm {
		return nil, ErrStripeDisputeEvidenceConfirmRequired
	}
	if strings.TrimSpace(input.APIKey) == "" {
		return nil, errors.New("stripe api key is required")
	}
	record, err := s.GetStripeDispute(input.DisputeID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	snapshot, payload, err := s.loadStripeEvidenceSubmissionSnapshot(record, input, now)
	if err != nil {
		return nil, err
	}
	params := stripeDisputeEvidenceParamsFromSnapshot(payload)
	params.Context = ctx
	params.Submit = stripe.Bool(input.Submit)
	params.Metadata = map[string]string{
		"commerce_platform_dispute_id": fmt.Sprint(record.ID),
		"commerce_platform_order_id":   stripeDisputeEvidenceOrderIDFromOrderID(record.OrderID),
	}

	audit := stripeDisputeEvidenceSubmissionAuditFromSnapshot(payload, now)
	audit.Submit = input.Submit
	audit.EvidenceSnapshotID = snapshot.ID
	audit.EvidenceSnapshotVersion = snapshot.Version
	audit.EvidenceSnapshotSHA256 = snapshot.SnapshotSHA256
	payloadBytes, _ := json.Marshal(audit)
	auditPayload := string(payloadBytes)

	submitter := s.stripeDisputeEvidenceSubmitter
	if submitter == nil {
		submitter = liveStripeDisputeEvidenceSubmitter{apiKey: strings.TrimSpace(input.APIKey)}
	}

	updated, err := submitter.Update(record.StripeDisputeID, params)
	if err != nil {
		_ = s.paymentRepo.UpdateStripeDisputeEvidenceSubmission(record.ID, nil, auditPayload, err.Error(), "")
		return nil, err
	}

	status := record.Status
	if input.Submit && updated != nil && updated.Status != "" {
		status = string(updated.Status)
	}
	var submittedAt *time.Time
	if input.Submit {
		submittedAt = &now
	}
	if err := s.paymentRepo.UpdateStripeDisputeEvidenceSubmission(record.ID, submittedAt, auditPayload, "", status); err != nil {
		return nil, err
	}

	return &SubmitStripeDisputeEvidenceResult{
		DisputeID:               record.ID,
		StripeDisputeID:         record.StripeDisputeID,
		StripeStatus:            status,
		SubmittedAt:             submittedAt,
		Staged:                  !input.Submit,
		EvidenceSnapshotID:      snapshot.ID,
		EvidenceSnapshotVersion: snapshot.Version,
		EvidenceSnapshotSHA256:  snapshot.SnapshotSHA256,
	}, nil
}

func (s *PaymentService) loadStripeEvidenceSubmissionSnapshot(
	record *paymentdomain.StripeDispute,
	input SubmitStripeDisputeEvidenceInput,
	lockedAt time.Time,
) (*orderevidence.OrderEvidenceSubmissionSnapshot, *stripeDisputeEvidenceSnapshotPayload, error) {
	if record == nil || record.OrderID == nil || *record.OrderID == 0 {
		return nil, nil, ErrStripeDisputeEvidenceNotSubmittable
	}

	var (
		snapshot *orderevidence.OrderEvidenceSubmissionSnapshot
		err      error
	)
	// A persisted snapshot belongs to a final submission attempt. A non-final
	// Stripe draft must always reflect the current order evidence package.
	if input.Submit {
		snapshot, err = s.findLatestEvidenceSubmissionSnapshot("stripe", record.ID)
		switch {
		case err == nil:
			payload, parseErr := parseStripeDisputeEvidenceSnapshot(
				snapshot,
				record.ID,
				record.StripeDisputeID,
				*record.OrderID,
			)
			return snapshot, payload, parseErr
		case !repository.IsRecordNotFound(err):
			return nil, nil, err
		}
	}

	pkg, err := s.BuildStripeDisputeEvidencePackage(record.ID)
	if err != nil {
		return nil, nil, err
	}
	if !pkg.CanSubmit || pkg.Order == nil {
		return nil, nil, ErrStripeDisputeEvidenceNotSubmittable
	}
	request := stripeDisputeEvidenceSnapshotRequestFromPackage(pkg, input)
	snapshot, err = buildLockedEvidenceSubmissionSnapshot(
		"stripe",
		record.ID,
		record.StripeDisputeID,
		*record.OrderID,
		pkg.FulfillmentEvidence,
		pkg.EvidenceChecklist,
		pkg.Warnings,
		request,
		lockedAt,
	)
	if err != nil {
		return nil, nil, err
	}
	if input.Submit {
		snapshot, err = s.createOrGetEvidenceSubmissionSnapshot(snapshot)
		if err != nil {
			return nil, nil, err
		}
	}
	payload, err := parseStripeDisputeEvidenceSnapshot(
		snapshot,
		record.ID,
		record.StripeDisputeID,
		*record.OrderID,
	)
	return snapshot, payload, err
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
		ShippingCarrier:        disputeShippingCarrier(orderRecord, pkg.Shipments),
		ShippingDate:           disputeShippingDate(orderRecord, pkg.TrackingEvents),
		ShippingTrackingNumber: disputeTrackingNumber(orderRecord, pkg.Shipments),
		CommunicationSummary:   disputeCommunicationSummary(pkg.Communications),
	}
	draft.UncategorizedText = disputeUncategorizedText(orderRecord, pkg.Dispute, pkg.TrackingEvents, pkg.PolicyDisclosure, pkg.Refunds)
	return draft
}

func finalizeStripeDisputeEvidencePackage(pkg *StripeDisputeEvidencePackage) {
	if pkg == nil {
		return
	}
	pkg.Evidence = buildStripeDisputeEvidenceDraft(pkg)
	pkg.EvidenceChecklist = buildDisputeEvidenceChecklist(
		"stripe",
		pkg.Order,
		pkg.Shipments,
		pkg.TrackingEvents,
		pkg.Communications,
		pkg.Authentication,
		pkg.PolicyDisclosure,
		pkg.Refunds,
		DisputeEvidenceChecklistOptions{},
	)
	pkg.SubmissionCheck = buildDisputeEvidenceSubmissionCheck(pkg.CanSubmit, pkg.EvidenceChecklist)
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
		lineTotal := "0"
		if value, err := item.TotalMoney(); err == nil {
			if formatted, formatErr := value.FormatMajor(); formatErr == nil {
				lineTotal = formatted
			}
		}
		lines = append(lines, fmt.Sprintf("- %s%s x%d, line total %s", item.ProductName, sku, item.Quantity, lineTotal))
	}
	orderTotal := "0"
	if value, err := orderRecord.TotalMoney(); err == nil {
		if formatted, formatErr := value.FormatMajor(); formatErr == nil {
			orderTotal = formatted
		}
	}
	lines = append(lines, fmt.Sprintf("Order total: %s.", orderTotal))
	return truncateEvidenceText(strings.Join(lines, "\n"), 20000)
}

func disputeUncategorizedText(orderRecord *orderdomain.Order, disputeRecord *paymentdomain.StripeDispute, events []shippingdomain.TrackingEvent, policyDisclosure *DisputePolicyDisclosureEvidence, refunds []DisputeRefundEvidence) string {
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
		amount := "invalid"
		if money, amountErr := domainmoney.New(disputeRecord.AmountMinor, disputeRecord.Currency); amountErr == nil {
			if formatted, formatErr := money.FormatMajor(); formatErr == nil {
				amount = formatted
			}
		}
		lines = append(lines, fmt.Sprintf("Stripe dispute reason: %s; disputed amount: %s %s.", disputeRecord.Reason, amount, disputeRecord.Currency))
	}
	if policyDisclosure != nil {
		lines = append(lines, fmt.Sprintf(
			"Refund & Cancellation Policy disclosure: version=%s; hash=%s; locale=%s; URL=%s; disclosed_at=%s; consented_at=%s; source=%s.",
			policyDisclosure.PolicyVersion,
			policyDisclosure.PolicyHash,
			policyDisclosure.Locale,
			policyDisclosure.PolicyURL,
			policyDisclosure.DisclosedAt.UTC().Format(time.RFC3339),
			formatOptionalDisputeTime(policyDisclosure.ConsentedAt),
			policyDisclosure.Source,
		))
	}
	if summary := refundEvidenceSummary(refunds); summary != "" {
		lines = append(lines, summary)
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
		return "Store support"
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

func disputeShippingCarrier(orderRecord *orderdomain.Order, shipments []shippingdomain.TrackingShipment) string {
	values := make([]string, 0, len(shipments))
	seen := make(map[string]struct{})
	for _, shipment := range shipments {
		value := ""
		if shipment.Carrier != nil {
			value = strings.TrimSpace(shipment.Carrier.Name)
		}
		if value == "" && shipment.Mapping != nil {
			value = strings.TrimSpace(shipment.Mapping.ProviderCarrierName)
		}
		if value == "" {
			value = strings.TrimSpace(shipment.ProviderCarrierCode)
		}
		if value != "" {
			if _, ok := seen[value]; !ok {
				seen[value] = struct{}{}
				values = append(values, value)
			}
		}
	}
	return strings.Join(values, ", ")
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

func disputeTrackingNumber(orderRecord *orderdomain.Order, shipments []shippingdomain.TrackingShipment) string {
	values := make([]string, 0, len(shipments))
	for _, shipment := range shipments {
		if trackingNumber := strings.TrimSpace(shipment.TrackingNumber); trackingNumber != "" {
			values = append(values, trackingNumber)
		}
	}
	return strings.Join(values, ", ")
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

func stripeDisputeEvidenceOrderIDFromOrderID(orderID *uint) string {
	if orderID == nil {
		return ""
	}
	return fmt.Sprint(*orderID)
}

func disputeOrderID(orderID *uint) uint {
	if orderID == nil {
		return 0
	}
	return *orderID
}
