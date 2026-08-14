package service

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
)

const orderDisputeScanLimit = 1000

var (
	ErrOrderDisputePaymentNotConfigured  = errors.New("order dispute payment service is not configured")
	ErrOrderDisputeNotFound              = errors.New("order dispute not found")
	ErrOrderDisputeEmailConfirmRequired  = errors.New("confirm is required before sending dispute contact email")
	ErrOrderDisputeEmailNotConfigured    = errors.New("order dispute contact email sender is not configured")
	ErrOrderDisputeEmailRecipientMissing = errors.New("order dispute customer email is missing")
	ErrOrderDisputeEmailSubjectRequired  = errors.New("order dispute contact email subject is required")
	ErrOrderDisputeEmailBodyRequired     = errors.New("order dispute contact email body is required")
)

type OrderDisputeListInput struct {
	Page     int
	PageSize int
	Provider string
	Status   string
	Search   string
}

type OrderDisputeCase struct {
	Provider            string                        `json:"provider"`
	DisputeID           uint                          `json:"dispute_id"`
	ProviderDisputeID   string                        `json:"provider_dispute_id"`
	ProviderPaymentID   string                        `json:"provider_payment_id,omitempty"`
	OrderID             *uint                         `json:"order_id,omitempty"`
	OrderNumber         string                        `json:"order_number,omitempty"`
	CustomerName        string                        `json:"customer_name,omitempty"`
	CustomerEmail       string                        `json:"customer_email,omitempty"`
	OrderStatus         string                        `json:"order_status,omitempty"`
	PaymentStatus       string                        `json:"payment_status,omitempty"`
	ShippingStatus      string                        `json:"shipping_status,omitempty"`
	TrackingNumber      string                        `json:"tracking_number,omitempty"`
	Amount              float64                       `json:"amount"`
	Currency            string                        `json:"currency"`
	Reason              string                        `json:"reason,omitempty"`
	Status              string                        `json:"status"`
	State               string                        `json:"state,omitempty"`
	LifeCycleStage      string                        `json:"life_cycle_stage,omitempty"`
	EvidenceDueAt       *time.Time                    `json:"evidence_due_at,omitempty"`
	EvidenceSubmittedAt *time.Time                    `json:"evidence_submitted_at,omitempty"`
	CreatedAt           time.Time                     `json:"created_at"`
	UpdatedAt           time.Time                     `json:"updated_at"`
	HasDeliveredEvent   bool                          `json:"has_delivered_event"`
	DeliveredAt         *time.Time                    `json:"delivered_at,omitempty"`
	NeedsResponse       bool                          `json:"needs_response"`
	EvidenceSummary     OrderDisputeEvidenceSummary   `json:"evidence_summary"`
	SubmissionReady     bool                          `json:"submission_ready"`
	SubmissionBlockers  []string                      `json:"submission_blockers"`
	Warnings            []string                      `json:"warnings"`
	MistakeAssessment   OrderDisputeMistakeAssessment `json:"mistake_assessment"`
	SuggestedAction     string                        `json:"suggested_action"`
	ContactDraft        OrderDisputeContactDraft      `json:"contact_draft"`
}

type OrderDisputeEvidenceSummary struct {
	Complete            bool      `json:"complete"`
	ReadyCount          int       `json:"ready_count"`
	TotalCount          int       `json:"total_count"`
	MissingCount        int       `json:"missing_count"`
	ManualRequiredCount int       `json:"manual_required_count"`
	UnavailableCount    int       `json:"unavailable_count"`
	BlockerCount        int       `json:"blocker_count"`
	MissingItems        []string  `json:"missing_items"`
	ManualItems         []string  `json:"manual_items"`
	LastEvaluatedAt     time.Time `json:"last_evaluated_at"`
}

type OrderDisputeMistakeAssessment struct {
	Level   string   `json:"level"`
	Label   string   `json:"label"`
	Reason  string   `json:"reason"`
	Signals []string `json:"signals"`
}

type OrderDisputeContactDraft struct {
	CanSend   bool   `json:"can_send"`
	To        string `json:"to,omitempty"`
	Subject   string `json:"subject,omitempty"`
	Body      string `json:"body,omitempty"`
	MailtoURL string `json:"mailto_url,omitempty"`
}

type OrderDisputeOrderAnalysis struct {
	Order        *orderdomain.Order          `json:"order,omitempty"`
	Disputes     []OrderDisputeCase          `json:"disputes"`
	Summary      OrderDisputeAnalysisSummary `json:"summary"`
	ContactDraft *OrderDisputeContactDraft   `json:"contact_draft,omitempty"`
}

type OrderDisputeAnalysisSummary struct {
	Total             int `json:"total"`
	NeedsResponse     int `json:"needs_response"`
	EvidenceBlocked   int `json:"evidence_blocked"`
	EvidenceSubmitted int `json:"evidence_submitted"`
	LikelyMistake     int `json:"likely_mistake"`
	MissingEmail      int `json:"missing_email"`
}

type SendOrderDisputeContactEmailInput struct {
	OrderID   uint
	Provider  string
	DisputeID uint
	Subject   string
	Body      string
	Confirm   bool
}

type SendOrderDisputeContactEmailResult struct {
	OrderID           uint      `json:"order_id"`
	Provider          string    `json:"provider"`
	DisputeID         uint      `json:"dispute_id"`
	ProviderDisputeID string    `json:"provider_dispute_id"`
	To                string    `json:"to"`
	Subject           string    `json:"subject"`
	SentAt            time.Time `json:"sent_at"`
}

type orderDisputeCandidate struct {
	provider string
	stripe   *paymentdomain.StripeDispute
	paypal   *paymentdomain.PayPalDispute
	order    *orderdomain.Order
}

func (s *OrderService) ListOrderDisputeCases(input OrderDisputeListInput) ([]OrderDisputeCase, int64, error) {
	if s == nil || s.payment == nil || s.payment.paymentRepo == nil {
		return nil, 0, ErrOrderDisputePaymentNotConfigured
	}
	input = normalizeOrderDisputeListInput(input)

	candidates, err := s.collectOrderDisputeCandidates(input)
	if err != nil {
		return nil, 0, err
	}
	candidates = filterOrderDisputeCandidates(candidates, input.Search)
	sort.SliceStable(candidates, func(i, j int) bool {
		return compareOrderDisputeCandidates(candidates[i], candidates[j])
	})

	total := int64(len(candidates))
	start, end := paginationWindow(input.Page, input.PageSize, len(candidates))
	if start >= end {
		return []OrderDisputeCase{}, total, nil
	}

	cases := make([]OrderDisputeCase, 0, end-start)
	for _, candidate := range candidates[start:end] {
		item, err := s.orderDisputeCaseFromCandidate(candidate, true)
		if err != nil {
			return nil, 0, err
		}
		cases = append(cases, item)
	}
	return cases, total, nil
}

func (s *OrderService) GetOrderDisputeAnalysis(orderID uint) (*OrderDisputeOrderAnalysis, error) {
	orderRecord, err := s.findOrder(orderID)
	if err != nil {
		return nil, err
	}
	if s == nil || s.payment == nil || s.payment.paymentRepo == nil {
		return nil, ErrOrderDisputePaymentNotConfigured
	}

	stripes, err := s.payment.paymentRepo.ListStripeDisputesByOrderID(orderID)
	if err != nil {
		return nil, err
	}
	paypals, err := s.payment.paymentRepo.ListPayPalDisputesByOrderID(orderID)
	if err != nil {
		return nil, err
	}

	candidates := make([]orderDisputeCandidate, 0, len(stripes)+len(paypals))
	for i := range stripes {
		candidates = append(candidates, orderDisputeCandidate{provider: "stripe", stripe: &stripes[i], order: orderRecord})
	}
	for i := range paypals {
		candidates = append(candidates, orderDisputeCandidate{provider: "paypal", paypal: &paypals[i], order: orderRecord})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return compareOrderDisputeCandidates(candidates[i], candidates[j])
	})

	cases := make([]OrderDisputeCase, 0, len(candidates))
	for _, candidate := range candidates {
		item, err := s.orderDisputeCaseFromCandidate(candidate, true)
		if err != nil {
			return nil, err
		}
		cases = append(cases, item)
	}

	analysis := &OrderDisputeOrderAnalysis{
		Order:    orderRecord,
		Disputes: cases,
		Summary:  summarizeOrderDisputeCases(cases),
	}
	for i := range cases {
		if cases[i].ContactDraft.CanSend && cases[i].NeedsResponse {
			analysis.ContactDraft = &cases[i].ContactDraft
			break
		}
	}
	return analysis, nil
}

func (s *OrderService) SendOrderDisputeContactEmail(input SendOrderDisputeContactEmailInput) (*SendOrderDisputeContactEmailResult, error) {
	if !input.Confirm {
		return nil, ErrOrderDisputeEmailConfirmRequired
	}
	if s == nil || s.emailSender == nil {
		return nil, ErrOrderDisputeEmailNotConfigured
	}
	analysis, err := s.GetOrderDisputeAnalysis(input.OrderID)
	if err != nil {
		return nil, err
	}
	provider := normalizeOrderDisputeProvider(input.Provider)
	var selected *OrderDisputeCase
	for i := range analysis.Disputes {
		if analysis.Disputes[i].Provider == provider && analysis.Disputes[i].DisputeID == input.DisputeID {
			selected = &analysis.Disputes[i]
			break
		}
	}
	if selected == nil {
		return nil, ErrOrderDisputeNotFound
	}

	to := strings.TrimSpace(selected.CustomerEmail)
	if to == "" {
		return nil, ErrOrderDisputeEmailRecipientMissing
	}
	subject := strings.TrimSpace(input.Subject)
	if subject == "" {
		return nil, ErrOrderDisputeEmailSubjectRequired
	}
	body := strings.TrimSpace(input.Body)
	if body == "" {
		return nil, ErrOrderDisputeEmailBodyRequired
	}

	if err := s.emailSender.SendEmail([]string{to}, subject, body); err != nil {
		return nil, err
	}
	return &SendOrderDisputeContactEmailResult{
		OrderID:           input.OrderID,
		Provider:          provider,
		DisputeID:         selected.DisputeID,
		ProviderDisputeID: selected.ProviderDisputeID,
		To:                to,
		Subject:           subject,
		SentAt:            time.Now().UTC(),
	}, nil
}

func (s *OrderService) collectOrderDisputeCandidates(input OrderDisputeListInput) ([]orderDisputeCandidate, error) {
	candidates := []orderDisputeCandidate{}
	if input.Provider == "" || input.Provider == "stripe" {
		stripes, _, err := s.payment.paymentRepo.ListStripeDisputes(input.Status, 1, orderDisputeScanLimit)
		if err != nil {
			return nil, err
		}
		for i := range stripes {
			candidates = append(candidates, orderDisputeCandidate{
				provider: "stripe",
				stripe:   &stripes[i],
				order:    s.orderForDispute(stripes[i].OrderID),
			})
		}
	}
	if input.Provider == "" || input.Provider == "paypal" {
		paypals, _, err := s.payment.paymentRepo.ListPayPalDisputes(input.Status, 1, orderDisputeScanLimit)
		if err != nil {
			return nil, err
		}
		for i := range paypals {
			candidates = append(candidates, orderDisputeCandidate{
				provider: "paypal",
				paypal:   &paypals[i],
				order:    s.orderForDispute(paypals[i].OrderID),
			})
		}
	}
	return candidates, nil
}

func (s *OrderService) orderForDispute(orderID *uint) *orderdomain.Order {
	if s == nil || orderID == nil || *orderID == 0 || s.orderRepo == nil {
		return nil
	}
	orderRecord, err := s.orderRepo.FindByID(*orderID)
	if err != nil {
		return nil
	}
	return orderRecord
}

func (s *OrderService) orderDisputeCaseFromCandidate(candidate orderDisputeCandidate, includeEvidence bool) (OrderDisputeCase, error) {
	if candidate.provider == "paypal" && candidate.paypal != nil {
		return s.orderDisputeCaseFromPayPal(candidate.paypal, candidate.order, includeEvidence)
	}
	if candidate.stripe == nil {
		return OrderDisputeCase{}, ErrOrderDisputeNotFound
	}
	return s.orderDisputeCaseFromStripe(candidate.stripe, candidate.order, includeEvidence)
}

func (s *OrderService) orderDisputeCaseFromStripe(record *paymentdomain.StripeDispute, orderRecord *orderdomain.Order, includeEvidence bool) (OrderDisputeCase, error) {
	item := OrderDisputeCase{
		Provider:            "stripe",
		DisputeID:           record.ID,
		ProviderDisputeID:   strings.TrimSpace(record.StripeDisputeID),
		ProviderPaymentID:   strings.TrimSpace(record.PaymentIntentID),
		OrderID:             record.OrderID,
		Amount:              record.Amount,
		Currency:            record.Currency,
		Reason:              strings.TrimSpace(record.Reason),
		Status:              strings.TrimSpace(record.Status),
		EvidenceDueAt:       record.EvidenceDueAt,
		EvidenceSubmittedAt: record.EvidenceSubmittedAt,
		CreatedAt:           record.CreatedAt,
		UpdatedAt:           record.UpdatedAt,
		NeedsResponse:       disputeNeedsResponse(record.Status) && record.EvidenceSubmittedAt == nil,
	}
	s.decorateOrderDisputeCase(&item, orderRecord)
	if includeEvidence && s.payment != nil {
		pkg, err := s.payment.BuildStripeDisputeEvidencePackage(record.ID)
		if err != nil {
			return item, err
		}
		s.applyStripeEvidencePackage(&item, pkg)
	}
	finalizeOrderDisputeCase(&item)
	return item, nil
}

func (s *OrderService) orderDisputeCaseFromPayPal(record *paymentdomain.PayPalDispute, orderRecord *orderdomain.Order, includeEvidence bool) (OrderDisputeCase, error) {
	item := OrderDisputeCase{
		Provider:            "paypal",
		DisputeID:           record.ID,
		ProviderDisputeID:   strings.TrimSpace(record.PayPalDisputeID),
		ProviderPaymentID:   strings.TrimSpace(record.ProviderPaymentID),
		OrderID:             record.OrderID,
		Amount:              record.Amount,
		Currency:            record.Currency,
		Reason:              strings.TrimSpace(record.Reason),
		Status:              strings.TrimSpace(record.Status),
		State:               strings.TrimSpace(record.DisputeState),
		LifeCycleStage:      strings.TrimSpace(record.DisputeLifeCycleStage),
		EvidenceSubmittedAt: record.EvidenceSubmittedAt,
		CreatedAt:           record.CreatedAt,
		UpdatedAt:           record.UpdatedAt,
		NeedsResponse:       paypalDisputeNeedsSellerEvidence(record) && record.EvidenceSubmittedAt == nil,
	}
	s.decorateOrderDisputeCase(&item, orderRecord)
	if includeEvidence && s.payment != nil {
		pkg, err := s.payment.BuildPayPalDisputeEvidencePackage(record.ID)
		if err != nil {
			return item, err
		}
		s.applyPayPalEvidencePackage(&item, pkg)
	}
	finalizeOrderDisputeCase(&item)
	return item, nil
}

func (s *OrderService) decorateOrderDisputeCase(item *OrderDisputeCase, orderRecord *orderdomain.Order) {
	if item == nil || orderRecord == nil {
		return
	}
	item.OrderID = &orderRecord.ID
	item.OrderNumber = strings.TrimSpace(orderRecord.OrderNumber)
	item.CustomerName = disputeCustomerName(orderRecord)
	item.CustomerEmail = disputeCustomerEmail(orderRecord)
	item.OrderStatus = strings.TrimSpace(orderRecord.Status)
	item.PaymentStatus = strings.TrimSpace(orderRecord.PaymentStatus)
	item.ShippingStatus = strings.TrimSpace(orderRecord.ShippingStatus)
	item.TrackingNumber = strings.TrimSpace(orderRecord.TrackingNumber)
	if strings.EqualFold(orderRecord.ShippingStatus, "delivered") || strings.EqualFold(orderRecord.Status, "completed") {
		item.HasDeliveredEvent = true
		if orderRecord.CompletedAt != nil {
			item.DeliveredAt = orderRecord.CompletedAt
		}
	}
}

func (s *OrderService) applyStripeEvidencePackage(item *OrderDisputeCase, pkg *StripeDisputeEvidencePackage) {
	if item == nil || pkg == nil {
		return
	}
	if pkg.Order != nil {
		s.decorateOrderDisputeCase(item, pkg.Order)
	}
	item.EvidenceSummary = orderDisputeEvidenceSummary(pkg.EvidenceChecklist)
	item.SubmissionReady = pkg.SubmissionCheck.Ready
	item.SubmissionBlockers = pkg.SubmissionCheck.Blockers
	item.Warnings = append([]string{}, pkg.Warnings...)
	if delivered := deliveredTrackingEvent(pkg.TrackingEvents); delivered != nil {
		item.HasDeliveredEvent = true
		item.DeliveredAt = disputeTimePointer(delivered.EventTime)
	}
}

func (s *OrderService) applyPayPalEvidencePackage(item *OrderDisputeCase, pkg *PayPalDisputeEvidencePackage) {
	if item == nil || pkg == nil {
		return
	}
	if pkg.Order != nil {
		s.decorateOrderDisputeCase(item, pkg.Order)
	}
	item.EvidenceSummary = orderDisputeEvidenceSummary(pkg.EvidenceChecklist)
	item.SubmissionReady = pkg.SubmissionCheck.Ready
	item.SubmissionBlockers = pkg.SubmissionCheck.Blockers
	item.Warnings = append([]string{}, pkg.Warnings...)
	if delivered := deliveredTrackingEvent(pkg.TrackingEvents); delivered != nil {
		item.HasDeliveredEvent = true
		item.DeliveredAt = disputeTimePointer(delivered.EventTime)
	}
}

func finalizeOrderDisputeCase(item *OrderDisputeCase) {
	if item == nil {
		return
	}
	item.MistakeAssessment = assessOrderDisputeMistake(*item)
	item.SuggestedAction = suggestedOrderDisputeAction(*item)
	item.ContactDraft = buildOrderDisputeContactDraft(*item)
}

func orderDisputeEvidenceSummary(checklist DisputeEvidenceChecklist) OrderDisputeEvidenceSummary {
	summary := OrderDisputeEvidenceSummary{
		Complete:            checklist.Complete,
		ReadyCount:          checklist.ReadyCount,
		TotalCount:          checklist.TotalCount,
		MissingCount:        checklist.MissingCount,
		ManualRequiredCount: checklist.ManualRequiredCount,
		UnavailableCount:    checklist.UnavailableCount,
		BlockerCount:        checklist.BlockerCount,
		MissingItems:        []string{},
		ManualItems:         []string{},
		LastEvaluatedAt:     checklist.LastEvaluatedAt,
	}
	for _, item := range checklist.Items {
		if item.Status == DisputeEvidenceStatusMissing || item.Status == DisputeEvidenceStatusUnavailable {
			summary.MissingItems = append(summary.MissingItems, item.Title)
		}
		if item.ManualRequired || item.Status == DisputeEvidenceStatusManualRequired {
			summary.ManualItems = append(summary.ManualItems, item.Title)
		}
	}
	return summary
}

func assessOrderDisputeMistake(item OrderDisputeCase) OrderDisputeMistakeAssessment {
	signals := []string{}
	if strings.TrimSpace(item.OrderNumber) == "" {
		return OrderDisputeMistakeAssessment{
			Level:   "unlinked_order",
			Label:   "未关联订单",
			Reason:  "拒付记录没有关联本地订单，先处理支付记录与订单绑定。",
			Signals: signals,
		}
	}
	if isResolvedOrderDispute(item) {
		return OrderDisputeMistakeAssessment{
			Level:   "resolved",
			Label:   "已结束",
			Reason:  "渠道状态显示该拒付已经进入结束或复核后状态。",
			Signals: signals,
		}
	}
	if strings.TrimSpace(item.CustomerEmail) == "" {
		signals = append(signals, "订单缺少客户邮箱")
		return OrderDisputeMistakeAssessment{
			Level:   "no_email",
			Label:   "缺少邮箱",
			Reason:  "无法直接向客户确认是否误操作，需要先补齐联系方式。",
			Signals: signals,
		}
	}
	if item.HasDeliveredEvent {
		signals = append(signals, "订单已有妥投/完成信号")
	}
	if orderDisputeReasonLooksUnrecognized(item.Reason) {
		signals = append(signals, "拒付原因接近未识别交易/未授权")
	}
	if item.HasDeliveredEvent && orderDisputeReasonLooksUnrecognized(item.Reason) {
		return OrderDisputeMistakeAssessment{
			Level:   "likely_mistake",
			Label:   "疑似误操作",
			Reason:  "客户可能忘记订单或账单描述，建议先邮件确认，再决定是否提交证据。",
			Signals: signals,
		}
	}
	if item.EvidenceSummary.BlockerCount > 0 {
		signals = append(signals, fmt.Sprintf("%d 项申诉阻断证据未满足", item.EvidenceSummary.BlockerCount))
		return OrderDisputeMistakeAssessment{
			Level:   "evidence_gap",
			Label:   "证据缺口",
			Reason:  "先补齐阻断证据，再判断是否联系客户或提交申诉。",
			Signals: signals,
		}
	}
	return OrderDisputeMistakeAssessment{
		Level:   "needs_review",
		Label:   "待人工判断",
		Reason:  "系统没有足够信号断定是误操作，需要结合订单、物流和客户沟通人工确认。",
		Signals: signals,
	}
}

func suggestedOrderDisputeAction(item OrderDisputeCase) string {
	switch item.MistakeAssessment.Level {
	case "likely_mistake":
		return "先邮件联系客户确认是否误操作，同时保留证据提交窗口。"
	case "no_email":
		return "先补齐客户邮箱或从支付渠道/客服记录中找到联系方式。"
	case "unlinked_order":
		return "先把拒付记录关联到本地订单和交易。"
	case "evidence_gap":
		return "先补齐阻断证据，再进入支付域提交申诉。"
	case "resolved":
		return "保留记录，无需继续联系客户。"
	default:
		if item.NeedsResponse {
			return "人工核对客户身份、物流和沟通记录后决定联系客户或提交申诉。"
		}
		return "持续观察渠道状态。"
	}
}

func buildOrderDisputeContactDraft(item OrderDisputeCase) OrderDisputeContactDraft {
	to := strings.TrimSpace(item.CustomerEmail)
	if to == "" {
		return OrderDisputeContactDraft{CanSend: false}
	}
	orderNumber := strings.TrimSpace(item.OrderNumber)
	if orderNumber == "" {
		orderNumber = fmt.Sprintf("#%d", derefUint(item.OrderID))
	}
	name := strings.TrimSpace(item.CustomerName)
	if name == "" {
		name = "there"
	}
	amount := fmt.Sprintf("%.2f %s", item.Amount, strings.ToUpper(strings.TrimSpace(item.Currency)))
	subject := fmt.Sprintf("Please confirm your order %s", orderNumber)
	bodyLines := []string{
		fmt.Sprintf("Hi %s,", name),
		"",
		fmt.Sprintf("We noticed a payment dispute for your order %s (%s). Sometimes this happens when a cardholder does not recognize the billing descriptor or opens a dispute by mistake.", orderNumber, amount),
		"",
		"Could you please reply and confirm whether you intended to open this dispute?",
		"",
		"Order details:",
		fmt.Sprintf("- Order number: %s", orderNumber),
		fmt.Sprintf("- Dispute channel: %s", strings.Title(item.Provider)),
	}
	if strings.TrimSpace(item.Reason) != "" {
		bodyLines = append(bodyLines, fmt.Sprintf("- Dispute reason: %s", item.Reason))
	}
	if strings.TrimSpace(item.TrackingNumber) != "" {
		bodyLines = append(bodyLines, fmt.Sprintf("- Tracking number: %s", item.TrackingNumber))
	}
	bodyLines = append(bodyLines,
		"",
		"If this was opened by mistake, please let us know in your reply. If there is a real issue with the order, tell us what happened and we will help resolve it.",
		"",
		"Thank you,",
		"Customer Support",
	)
	body := strings.Join(bodyLines, "\n")
	values := url.Values{}
	values.Set("subject", subject)
	values.Set("body", body)
	return OrderDisputeContactDraft{
		CanSend:   true,
		To:        to,
		Subject:   subject,
		Body:      body,
		MailtoURL: "mailto:" + url.PathEscape(to) + "?" + values.Encode(),
	}
}

func summarizeOrderDisputeCases(cases []OrderDisputeCase) OrderDisputeAnalysisSummary {
	summary := OrderDisputeAnalysisSummary{Total: len(cases)}
	for _, item := range cases {
		if item.NeedsResponse {
			summary.NeedsResponse++
		}
		if item.EvidenceSummary.BlockerCount > 0 {
			summary.EvidenceBlocked++
		}
		if item.EvidenceSubmittedAt != nil {
			summary.EvidenceSubmitted++
		}
		if item.MistakeAssessment.Level == "likely_mistake" {
			summary.LikelyMistake++
		}
		if strings.TrimSpace(item.CustomerEmail) == "" {
			summary.MissingEmail++
		}
	}
	return summary
}

func normalizeOrderDisputeListInput(input OrderDisputeListInput) OrderDisputeListInput {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}
	input.Provider = normalizeOrderDisputeProvider(input.Provider)
	input.Status = strings.TrimSpace(input.Status)
	input.Search = strings.ToLower(strings.TrimSpace(input.Search))
	return input
}

func normalizeOrderDisputeProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "stripe":
		return "stripe"
	case "paypal":
		return "paypal"
	default:
		return ""
	}
}

func filterOrderDisputeCandidates(candidates []orderDisputeCandidate, search string) []orderDisputeCandidate {
	search = strings.ToLower(strings.TrimSpace(search))
	if search == "" {
		return candidates
	}
	filtered := make([]orderDisputeCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if orderDisputeCandidateMatches(candidate, search) {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

func orderDisputeCandidateMatches(candidate orderDisputeCandidate, search string) bool {
	haystack := []string{candidate.provider}
	if candidate.stripe != nil {
		haystack = append(haystack,
			candidate.stripe.StripeDisputeID,
			candidate.stripe.StripeChargeID,
			candidate.stripe.PaymentIntentID,
			candidate.stripe.Reason,
			candidate.stripe.Status,
		)
	}
	if candidate.paypal != nil {
		haystack = append(haystack,
			candidate.paypal.PayPalDisputeID,
			candidate.paypal.ProviderPaymentID,
			candidate.paypal.Reason,
			candidate.paypal.Status,
			candidate.paypal.DisputeState,
			candidate.paypal.DisputeLifeCycleStage,
		)
	}
	if candidate.order != nil {
		haystack = append(haystack,
			candidate.order.OrderNumber,
			disputeCustomerName(candidate.order),
			disputeCustomerEmail(candidate.order),
			candidate.order.TrackingNumber,
		)
	}
	for _, value := range haystack {
		if strings.Contains(strings.ToLower(strings.TrimSpace(value)), search) {
			return true
		}
	}
	return false
}

func compareOrderDisputeCandidates(left, right orderDisputeCandidate) bool {
	leftRank := orderDisputeCandidateRank(left)
	rightRank := orderDisputeCandidateRank(right)
	if leftRank != rightRank {
		return leftRank < rightRank
	}
	leftDue := orderDisputeCandidateDueAt(left)
	rightDue := orderDisputeCandidateDueAt(right)
	if leftDue != nil && rightDue != nil && !leftDue.Equal(*rightDue) {
		return leftDue.Before(*rightDue)
	}
	if leftDue != nil && rightDue == nil {
		return true
	}
	if leftDue == nil && rightDue != nil {
		return false
	}
	return orderDisputeCandidateUpdatedAt(left).After(orderDisputeCandidateUpdatedAt(right))
}

func orderDisputeCandidateRank(candidate orderDisputeCandidate) int {
	if candidate.stripe != nil {
		if disputeNeedsResponse(candidate.stripe.Status) && candidate.stripe.EvidenceSubmittedAt == nil {
			return 0
		}
		if strings.EqualFold(candidate.stripe.Status, "under_review") {
			return 1
		}
		return 2
	}
	if candidate.paypal != nil {
		if paypalDisputeNeedsSellerEvidence(candidate.paypal) && candidate.paypal.EvidenceSubmittedAt == nil {
			return 0
		}
		status := strings.ToUpper(strings.TrimSpace(candidate.paypal.Status))
		if status == "UNDER_REVIEW" || status == "WAITING_FOR_BUYER_RESPONSE" {
			return 1
		}
		return 2
	}
	return 3
}

func orderDisputeCandidateDueAt(candidate orderDisputeCandidate) *time.Time {
	if candidate.stripe != nil {
		return candidate.stripe.EvidenceDueAt
	}
	return nil
}

func orderDisputeCandidateUpdatedAt(candidate orderDisputeCandidate) time.Time {
	if candidate.stripe != nil {
		return candidate.stripe.UpdatedAt
	}
	if candidate.paypal != nil {
		return candidate.paypal.UpdatedAt
	}
	return time.Time{}
}

func paginationWindow(page, pageSize, length int) (int, int) {
	start := (page - 1) * pageSize
	if start > length {
		return length, length
	}
	end := start + pageSize
	if end > length {
		end = length
	}
	return start, end
}

func orderDisputeReasonLooksUnrecognized(reason string) bool {
	reason = strings.ToLower(strings.TrimSpace(reason))
	return strings.Contains(reason, "fraud") ||
		strings.Contains(reason, "unrecognized") ||
		strings.Contains(reason, "unauthorized") ||
		strings.Contains(reason, "not_recognized") ||
		strings.Contains(reason, "unauthorised")
}

func isResolvedOrderDispute(item OrderDisputeCase) bool {
	status := strings.ToLower(strings.TrimSpace(item.Status))
	state := strings.ToLower(strings.TrimSpace(item.State))
	if item.Provider == "paypal" {
		return status == "resolved" || state == "resolved"
	}
	return status == "won" || status == "lost" || status == "closed"
}

func derefUint(value *uint) uint {
	if value == nil {
		return 0
	}
	return *value
}
