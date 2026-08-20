package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
)

const (
	DisputeEvidenceStatusReady          = "ready"
	DisputeEvidenceStatusMissing        = "missing"
	DisputeEvidenceStatusManualRequired = "manual_required"
	DisputeEvidenceStatusUnavailable    = "unavailable"
)

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

type DisputeEvidenceSubmissionCheck struct {
	Ready    bool     `json:"ready"`
	Blockers []string `json:"blockers"`
	Warnings []string `json:"warnings"`
}

type DisputePaymentAuthenticationEvidence struct {
	PaymentIntentID    string     `json:"payment_intent_id,omitempty"`
	TransactionID      string     `json:"transaction_id,omitempty"`
	ThreeDSecureResult string     `json:"three_ds_secure_result,omitempty"`
	CVCCheck           string     `json:"cvc_check,omitempty"`
	AVSLine1Check      string     `json:"avs_line1_check,omitempty"`
	AVSPostalCodeCheck string     `json:"avs_postal_code_check,omitempty"`
	AVSMatch           string     `json:"avs_match,omitempty"`
	Source             string     `json:"source,omitempty"`
	ObservedAt         *time.Time `json:"observed_at,omitempty"`
}

func (s *PaymentService) findDisputeEvidenceTransaction(provider string, transactionID *uint, providerPaymentID string, orderID uint) (*paymentdomain.Transaction, error) {
	if s == nil || s.paymentRepo == nil {
		return nil, nil
	}

	if transactionID != nil && *transactionID > 0 {
		transaction, err := s.paymentRepo.FindTransactionByID(*transactionID)
		if err == nil {
			return transaction, nil
		}
		if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}

	if providerPaymentID = strings.TrimSpace(providerPaymentID); providerPaymentID != "" {
		transaction, err := s.paymentRepo.FindTransactionByTransactionID(providerPaymentID)
		if err == nil {
			return transaction, nil
		}
		if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}

	if orderID == 0 {
		return nil, nil
	}
	transactions, err := s.paymentRepo.FindTransactionByOrderID(orderID)
	if err != nil {
		return nil, err
	}
	for index := range transactions {
		if strings.EqualFold(strings.TrimSpace(transactions[index].PaymentMethod), strings.TrimSpace(provider)) {
			return &transactions[index], nil
		}
	}
	return nil, nil
}

func buildStripeAuthenticationEvidence(transaction *paymentdomain.Transaction, paymentIntentID string) *DisputePaymentAuthenticationEvidence {
	if transaction == nil {
		return nil
	}

	result := &DisputePaymentAuthenticationEvidence{
		PaymentIntentID: paymentIntentID,
		TransactionID:   strings.TrimSpace(transaction.TransactionID),
		Source:          "transactions.gateway_response",
		ObservedAt:      disputeTransactionObservedAt(transaction),
	}
	if result.PaymentIntentID == "" {
		result.PaymentIntentID = result.TransactionID
	}

	var payload interface{}
	if strings.TrimSpace(transaction.GatewayResponse) != "" && json.Unmarshal([]byte(transaction.GatewayResponse), &payload) == nil {
		result.ThreeDSecureResult = findStripeThreeDSecureResult(payload)
		result.CVCCheck = findJSONField(payload, "cvc_check")
		result.AVSLine1Check = findJSONField(payload, "address_line1_check")
		result.AVSPostalCodeCheck = firstDisputeNonEmpty(
			findJSONField(payload, "address_postal_code_check"),
			findJSONField(payload, "address_zip_check"),
		)
		result.AVSMatch = summarizeAVSChecks(result.AVSLine1Check, result.AVSPostalCodeCheck)
	}

	if result.ThreeDSecureResult == "" && result.CVCCheck == "" && result.AVSMatch == "" {
		return result
	}
	return result
}

func disputeTransactionObservedAt(transaction *paymentdomain.Transaction) *time.Time {
	if transaction == nil {
		return nil
	}
	if transaction.CompletedAt != nil && !transaction.CompletedAt.IsZero() {
		value := transaction.CompletedAt.UTC()
		return &value
	}
	if !transaction.UpdatedAt.IsZero() {
		value := transaction.UpdatedAt.UTC()
		return &value
	}
	if !transaction.CreatedAt.IsZero() {
		value := transaction.CreatedAt.UTC()
		return &value
	}
	return nil
}

func buildDisputeEvidenceChecklist(
	provider string,
	orderRecord *orderdomain.Order,
	shipment *shippingdomain.TrackingShipment,
	events []shippingdomain.TrackingEvent,
	communications []StripeDisputeCommunicationEvidence,
	authentication *DisputePaymentAuthenticationEvidence,
	policyDisclosure *DisputePolicyDisclosureEvidence,
	refunds []DisputeRefundEvidence,
) DisputeEvidenceChecklist {
	provider = strings.ToLower(strings.TrimSpace(provider))
	items := make([]DisputeEvidenceChecklistItem, 0, 7)

	orderReady := orderRecord != nil &&
		strings.TrimSpace(orderRecord.OrderNumber) != "" &&
		orderRecord.TotalAmount > 0 &&
		len(orderRecord.Items) > 0
	orderSummary := "订单号、下单时间、商品明细和金额来自本地订单。"
	orderReason := ""
	orderObservedAt := (*time.Time)(nil)
	if orderReady {
		orderObservedAt = disputeTimePointer(orderRecord.CreatedAt)
		orderSummary = fmt.Sprintf(
			"订单 %s：%d 个商品明细，总额 %.2f %s。",
			orderRecord.OrderNumber,
			len(orderRecord.Items),
			orderRecord.TotalAmount,
			orderRecord.Currency,
		)
	} else {
		orderReason = "拒付记录没有关联完整的本地订单，无法证明采购凭证。"
	}
	items = append(items, DisputeEvidenceChecklistItem{
		Key:           "purchase_order",
		Title:         "采购与订单凭证",
		ProviderField: disputeProviderField(provider, "receipt", "commercial_invoice"),
		Status:        checklistStatus(orderReady, DisputeEvidenceStatusMissing),
		Required:      true,
		Blocker:       true,
		Source:        "orders + order_items",
		ObservedAt:    orderObservedAt,
		Summary:       orderSummary,
		MissingReason: orderReason,
	})

	trackingNumber := disputeTrackingNumber(orderRecord, shipment)
	delivered := deliveredTrackingEvent(events)
	fulfillmentReady := trackingNumber != "" && delivered != nil
	fulfillmentStatus := DisputeEvidenceStatusMissing
	fulfillmentReason := "订单没有物流单号，也没有本地妥投事件。"
	fulfillmentSummary := "没有可用于核验履约的物流记录。"
	var fulfillmentObservedAt *time.Time
	if trackingNumber != "" && delivered != nil {
		fulfillmentStatus = DisputeEvidenceStatusManualRequired
		fulfillmentReason = "本地已有妥投事件，但官方承运商签名/POD 文件尚未在系统内归档。"
		fulfillmentSummary = fmt.Sprintf("物流单号 %s 已出现妥投事件：%s。", trackingNumber, delivered.Description)
		fulfillmentObservedAt = disputeTimePointer(delivered.EventTime)
	} else if trackingNumber != "" {
		fulfillmentStatus = DisputeEvidenceStatusManualRequired
		fulfillmentReason = "已有物流单号，但本地追踪记录没有明确 delivered/signed 事件。"
		fulfillmentSummary = fmt.Sprintf("物流单号 %s 已保存，等待同步妥投轨迹。", trackingNumber)
	}
	items = append(items, DisputeEvidenceChecklistItem{
		Key:            "fulfillment_delivery",
		Title:          "履约与物流妥投证明",
		ProviderField:  disputeProviderField(provider, "shipping_documentation", "proof_of_fulfillment"),
		Status:         fulfillmentStatus,
		Required:       true,
		Blocker:        provider == "paypal" && !fulfillmentReady,
		ManualRequired: fulfillmentStatus == DisputeEvidenceStatusManualRequired,
		Source:         "orders + shipping_tracking_shipments + tracking_events",
		ObservedAt:     fulfillmentObservedAt,
		Summary:        fulfillmentSummary,
		MissingReason:  fulfillmentReason,
	})

	items = append(items, DisputeEvidenceChecklistItem{
		Key:            "visitor_risk",
		Title:          "买家 IP 与设备指纹",
		ProviderField:  "customer_ip / customer_device_fingerprint",
		Status:         DisputeEvidenceStatusUnavailable,
		Required:       true,
		ManualRequired: true,
		Source:         "visitor_risk_service",
		Summary:        "当前只保存访客 IP/UA 哈希和聚合风险事实，没有与订单绑定的原始下单快照。",
		MissingReason:  "尚未接入订单级 IP、User-Agent、设备指纹和下单时间关联；系统不会伪造这项证据。",
	})

	authReady := authentication != nil && (strings.TrimSpace(authentication.ThreeDSecureResult) != "" ||
		strings.TrimSpace(authentication.CVCCheck) != "" ||
		strings.TrimSpace(authentication.AVSMatch) != "")
	authStatus := DisputeEvidenceStatusUnavailable
	authSummary := "PayPal 交易不适用 Stripe PaymentIntent 的 3DS/CVC/AVS 字段。"
	authReason := ""
	if provider == "stripe" {
		authStatus = DisputeEvidenceStatusMissing
		authSummary = "没有从保存的 Stripe 网关回执中解析到 3DS、CVC 或 AVS 结果。"
		authReason = "需要保留包含 payment_method_details.card.checks / three_d_secure.result 的 Stripe 回执，当前数据不足。"
		if authentication != nil {
			authSummary = formatStripeAuthenticationSummary(authentication)
			authReason = "已找到交易回执，但其中没有可核验的 3DS、CVC 或 AVS 结果。"
		}
		if authReady {
			authStatus = DisputeEvidenceStatusReady
			authReason = ""
		}
	}
	items = append(items, DisputeEvidenceChecklistItem{
		Key:            "payment_authentication",
		Title:          "3D Secure / CVC / AVS 鉴权日志",
		ProviderField:  disputeProviderField(provider, "payment_method_details.card", "not_applicable"),
		Status:         authStatus,
		Required:       provider == "stripe",
		ManualRequired: provider == "stripe" && !authReady,
		Source:         "transactions.gateway_response",
		ObservedAt:     authenticationObservedAt(authentication),
		Summary:        authSummary,
		MissingReason:  authReason,
	})

	items = append(items, DisputeEvidenceChecklistItem{
		Key:           "policy_disclosure",
		Title:         "售后条款与退款政策披露",
		ProviderField: "refund_policy_disclosure",
		Status:        DisputeEvidenceStatusMissing,
		Required:      true,
		Source:        "order_policy_disclosures",
		Summary:       "订单没有保存退款政策的历史披露快照。",
		MissingReason: "无法证明买家在下单时看到的具体政策版本；当前政策页面不会被当作历史证据。",
	})
	if policyDisclosure != nil &&
		strings.TrimSpace(policyDisclosure.PolicyHash) != "" &&
		strings.TrimSpace(policyDisclosure.PolicyURL) != "" &&
		!policyDisclosure.DisclosedAt.IsZero() {
		policySummary := fmt.Sprintf(
			"订单已保存退款政策快照：version=%s，locale=%s，URL=%s，disclosed_at=%s。",
			policyDisclosure.PolicyVersion,
			policyDisclosure.Locale,
			policyDisclosure.PolicyURL,
			policyDisclosure.DisclosedAt.UTC().Format(time.RFC3339),
		)
		policyReason := ""
		manualRequired := false
		policyStatus := DisputeEvidenceStatusReady
		if policyDisclosure.ConsentedAt == nil {
			policySummary += " 未记录显式结算页 Consent 时间戳。"
			policyReason = "快照已保存，但建单请求没有显式确认政策同意时间；可补充前台 Consent 事件作为人工佐证。"
			manualRequired = true
			policyStatus = DisputeEvidenceStatusManualRequired
		}
		for index := range items {
			if items[index].Key != "policy_disclosure" {
				continue
			}
			items[index].Status = policyStatus
			items[index].ManualRequired = manualRequired
			items[index].ObservedAt = disputeTimePointer(policyDisclosure.DisclosedAt)
			items[index].Summary = policySummary
			items[index].MissingReason = policyReason
			break
		}
	}

	refundReady := len(refunds) > 0
	refundSummary := "订单没有退款记录。"
	refundReason := "没有可关联的退款事实；这不代表订单没有售后争议，只表示系统没有保存退款记录。"
	if refundReady {
		refundSummary = fmt.Sprintf("已关联 %d 条退款事实，包含金额、状态、支付渠道退款 ID 和商品行快照摘要。", len(refunds))
		refundReason = ""
	}
	items = append(items, DisputeEvidenceChecklistItem{
		Key:           "refund_activity",
		Title:         "退款执行事实",
		ProviderField: disputeProviderField(provider, "uncategorized_text", "notes"),
		Status:        checklistStatus(refundReady, DisputeEvidenceStatusMissing),
		Required:      false,
		Source:        "refunds + refund_line_items",
		ObservedAt:    latestRefundEvidenceAt(refunds),
		Summary:       refundSummary,
		MissingReason: refundReason,
	})

	communicationReady := len(communications) > 0
	communicationObservedAt := latestCommunicationAt(communications)
	communicationSummary := "没有找到可关联的客服消息。"
	communicationReason := "按订单号、客户账号或订单邮箱没有匹配到外部客服消息。"
	if communicationReady {
		communicationSummary = fmt.Sprintf("已关联 %d 条非内部客服消息，最早和最近时间均保留在下方沟通记录中。", len(communications))
		communicationReason = ""
	}
	items = append(items, DisputeEvidenceChecklistItem{
		Key:           "customer_communication",
		Title:         "买家沟通与客服记录",
		ProviderField: disputeProviderField(provider, "customer_communication", "notes"),
		Status:        checklistStatus(communicationReady, DisputeEvidenceStatusMissing),
		Required:      false,
		Source:        "tickets + ticket_messages",
		ObservedAt:    communicationObservedAt,
		Summary:       communicationSummary,
		MissingReason: communicationReason,
	})

	customizationSummary, customizationObservedAt := disputeCustomizationEvidence(orderRecord)
	customizationReady := customizationSummary != ""
	customizationReason := "订单明细没有可识别的 Spoke Calculator 定制参数或定制商品属性。"
	if customizationReady {
		customizationReason = ""
	}
	items = append(items, DisputeEvidenceChecklistItem{
		Key:            "customization_parameters",
		Title:          "辐条 / 定制参数记录",
		ProviderField:  disputeProviderField(provider, "service_documentation", "notes"),
		Status:         checklistStatus(customizationReady, DisputeEvidenceStatusMissing),
		Required:       false,
		ManualRequired: !customizationReady,
		Source:         "order_items.attributes / spoke calculator",
		ObservedAt:     customizationObservedAt,
		Summary:        nonEmptyOr(customizationSummary, "没有找到与本订单关联的 Spoke 定制参数。"),
		MissingReason:  customizationReason,
	})

	return finalizeDisputeEvidenceChecklist(items)
}

func finalizeDisputeEvidenceChecklist(items []DisputeEvidenceChecklistItem) DisputeEvidenceChecklist {
	checklist := DisputeEvidenceChecklist{
		Items:           items,
		TotalCount:      len(items),
		LastEvaluatedAt: time.Now().UTC(),
	}
	for _, item := range items {
		switch item.Status {
		case DisputeEvidenceStatusReady:
			checklist.ReadyCount++
		case DisputeEvidenceStatusMissing:
			checklist.MissingCount++
		case DisputeEvidenceStatusManualRequired:
			checklist.ManualRequiredCount++
		case DisputeEvidenceStatusUnavailable:
			checklist.UnavailableCount++
		}
		if item.Blocker && item.Status != DisputeEvidenceStatusReady {
			checklist.BlockerCount++
		}
	}
	checklist.Complete = checklist.TotalCount > 0
	for _, item := range items {
		if item.Status == DisputeEvidenceStatusReady {
			continue
		}
		if item.Status == DisputeEvidenceStatusUnavailable && !item.Required {
			continue
		}
		checklist.Complete = false
		break
	}
	return checklist
}

func buildDisputeEvidenceSubmissionCheck(canSubmit bool, checklist DisputeEvidenceChecklist) DisputeEvidenceSubmissionCheck {
	check := DisputeEvidenceSubmissionCheck{
		Ready:    canSubmit && checklist.BlockerCount == 0,
		Blockers: []string{},
		Warnings: []string{},
	}
	if !canSubmit {
		check.Blockers = append(check.Blockers, "当前渠道状态不允许提交卖方证据。")
	}
	for _, item := range checklist.Items {
		if item.Blocker && item.Status != DisputeEvidenceStatusReady {
			check.Blockers = append(check.Blockers, item.Title+"："+nonEmptyOr(item.MissingReason, item.Summary))
		}
		if item.Status != DisputeEvidenceStatusReady && !item.Blocker {
			check.Warnings = append(check.Warnings, item.Title+"："+nonEmptyOr(item.MissingReason, item.Summary))
		}
	}
	return check
}

func checklistStatus(ready bool, missingStatus string) string {
	if ready {
		return DisputeEvidenceStatusReady
	}
	return missingStatus
}

func disputeProviderField(provider, stripeField, paypalField string) string {
	if strings.EqualFold(strings.TrimSpace(provider), "paypal") {
		return paypalField
	}
	return stripeField
}

func disputeTimePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}

func authenticationObservedAt(authentication *DisputePaymentAuthenticationEvidence) *time.Time {
	if authentication == nil {
		return nil
	}
	return authentication.ObservedAt
}

func latestCommunicationAt(items []StripeDisputeCommunicationEvidence) *time.Time {
	var latest time.Time
	for _, item := range items {
		if item.CreatedAt.After(latest) {
			latest = item.CreatedAt
		}
	}
	return disputeTimePointer(latest)
}

func formatStripeAuthenticationSummary(authentication *DisputePaymentAuthenticationEvidence) string {
	if authentication == nil {
		return ""
	}
	parts := []string{}
	if value := strings.TrimSpace(authentication.ThreeDSecureResult); value != "" {
		parts = append(parts, "3DS result="+value)
	}
	if value := strings.TrimSpace(authentication.CVCCheck); value != "" {
		parts = append(parts, "CVC="+value)
	}
	if value := strings.TrimSpace(authentication.AVSMatch); value != "" {
		parts = append(parts, "AVS="+value)
	}
	if len(parts) == 0 {
		return "已找到 Stripe 交易回执，但没有可展示的鉴权结果。"
	}
	return strings.Join(parts, "; ") + "。"
}

func summarizeAVSChecks(line1, postal string) string {
	checks := []string{strings.ToLower(strings.TrimSpace(line1)), strings.ToLower(strings.TrimSpace(postal))}
	hasValue := false
	hasFailure := false
	hasPass := false
	for _, check := range checks {
		if check == "" {
			continue
		}
		hasValue = true
		if check == "pass" || check == "matched" {
			hasPass = true
		} else if check == "fail" || check == "unmatched" {
			hasFailure = true
		}
	}
	switch {
	case hasFailure:
		return "mismatch"
	case hasPass && hasValue:
		return "match"
	case hasValue:
		return "partial"
	default:
		return ""
	}
}

func findStripeThreeDSecureResult(value interface{}) string {
	var result string
	walkJSON(value, func(key string, nested interface{}) bool {
		normalizedKey := strings.NewReplacer("-", "_", " ", "_").Replace(strings.ToLower(strings.TrimSpace(key)))
		if !strings.Contains(normalizedKey, "three_d_secure") && normalizedKey != "3ds" && normalizedKey != "three_ds" {
			return false
		}
		if nestedMap, ok := nested.(map[string]interface{}); ok {
			result = firstDisputeNonEmpty(
				findJSONField(nestedMap, "result"),
				findJSONField(nestedMap, "status"),
			)
		} else if stringValue, ok := nested.(string); ok {
			result = strings.TrimSpace(stringValue)
		}
		return result != ""
	})
	return result
}

func findJSONField(value interface{}, wantedKeys ...string) string {
	wanted := map[string]struct{}{}
	for _, key := range wantedKeys {
		wanted[strings.ToLower(strings.TrimSpace(key))] = struct{}{}
	}
	var result string
	walkJSON(value, func(key string, nested interface{}) bool {
		if _, ok := wanted[strings.ToLower(strings.TrimSpace(key))]; !ok {
			return false
		}
		if stringValue, ok := nested.(string); ok {
			result = strings.TrimSpace(stringValue)
		}
		return result != ""
	})
	return result
}

func walkJSON(value interface{}, visit func(string, interface{}) bool) bool {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, nested := range typed {
			if visit(key, nested) {
				return true
			}
			if walkJSON(nested, visit) {
				return true
			}
		}
	case []interface{}:
		for _, nested := range typed {
			if walkJSON(nested, visit) {
				return true
			}
		}
	}
	return false
}

func disputeCustomizationEvidence(orderRecord *orderdomain.Order) (string, *time.Time) {
	if orderRecord == nil {
		return "", nil
	}
	markers := []string{"spoke", "erd", "flange", "lacing", "cross", "nipple", "wheel_type", "hub_model", "rim_model"}
	for _, item := range orderRecord.Items {
		raw := strings.TrimSpace(item.Attributes)
		if raw == "" {
			continue
		}
		lower := strings.ToLower(raw)
		matched := false
		for _, marker := range markers {
			if strings.Contains(lower, marker) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		sku := strings.TrimSpace(item.SKU)
		if sku == "" {
			sku = fmt.Sprintf("商品明细 #%d", item.ID)
		}
		return truncateEvidenceText(fmt.Sprintf("%s：%s", sku, raw), 1800), disputeTimePointer(item.CreatedAt)
	}
	return "", nil
}

func nonEmptyOr(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return strings.TrimSpace(fallback)
}

func firstDisputeNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
