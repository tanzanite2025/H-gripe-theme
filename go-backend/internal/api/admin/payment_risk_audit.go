package admin

import (
	"strings"
	"time"

	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	paymentAuditActionCreate    = "create"
	paymentAuditActionExecute   = "execute"
	paymentAuditActionRecompute = "recompute"
	paymentAuditActionRevoke    = "revoke"
	paymentAuditActionSubmit    = "submit"

	paymentAuditResourceRefundExecution      = "payment_refund_execution"
	paymentAuditResourceRefundRecommendation = "payment_refund_recommendation"
	paymentAuditResourceRefundDraft          = "payment_refund"
	paymentAuditResourceDisputeEvidence      = "stripe_dispute_evidence"
	paymentAuditResourcePaymentReview        = "payment_review"
	paymentAuditResourcePaymentMethod        = "payment_method"
	paymentAuditResourceRiskMonitoring       = "payment_risk_monitoring"
	paymentAuditResourceProtectionControl    = "payment_protection_control"
)

func (h *PaymentRefundExecutionHandler) ConfigureAuditService(recorder paymentAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *PaymentRefundRecommendationHandler) ConfigureAuditService(recorder paymentAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *PaymentRiskMonitoringHandler) ConfigureAuditService(recorder paymentAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *PaymentProtectionHandler) ConfigureAuditService(recorder paymentAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *PaymentRefundExecutionHandler) recordRefundExecutionAudit(c *gin.Context, event paymentAdminAuditEvent) {
	if h == nil {
		return
	}
	recordPaymentAdminAudit(h.auditService, c, event)
}

func (h *PaymentRefundRecommendationHandler) recordRefundRecommendationAudit(c *gin.Context, event paymentAdminAuditEvent) {
	if h == nil {
		return
	}
	recordPaymentAdminAudit(h.auditService, c, event)
}

func (h *PaymentRiskMonitoringHandler) recordRiskMonitoringAudit(c *gin.Context, event paymentAdminAuditEvent) {
	if h == nil {
		return
	}
	recordPaymentAdminAudit(h.auditService, c, event)
}

func (h *PaymentProtectionHandler) recordProtectionControlAudit(c *gin.Context, event paymentAdminAuditEvent) {
	if h == nil {
		return
	}
	recordPaymentAdminAudit(h.auditService, c, event)
}

func paymentRefundExecutionAuditDetails(
	refundID uint,
	provider string,
	refund *paymentdomain.Refund,
	transaction *paymentdomain.Transaction,
	execution *paymentdomain.PaymentRefundExecution,
) map[string]interface{} {
	details := map[string]interface{}{
		"refund_id": refundID,
	}
	if provider = strings.TrimSpace(provider); provider != "" {
		details["provider"] = provider
	}
	if refund != nil {
		details["refund_id"] = refund.ID
		details["order_id"] = refund.OrderID
		details["transaction_id"] = refund.TransactionID
		details["amount"] = refund.Amount
		details["refund_status"] = refund.Status
		details["line_item_count"] = len(refund.LineItems)
	}
	if transaction != nil {
		details["order_id"] = transaction.OrderID
		details["transaction_id"] = transaction.ID
		details["provider"] = transaction.PaymentMethod
		details["currency"] = transaction.Currency
		details["transaction_status"] = transaction.Status
	}
	if execution != nil {
		details["execution_id"] = execution.ID
		details["execution_status"] = execution.Status
		details["attempt_count"] = execution.AttemptCount
		details["merchant_order_number"] = execution.MerchantOrderNumber
		details["provider_transaction_id_present"] = strings.TrimSpace(execution.ProviderTransactionID) != ""
		details["provider_refund_id_present"] = strings.TrimSpace(execution.ProviderRefundID) != ""
		details["provider_status"] = execution.ProviderStatus
	}
	return details
}

func paymentRefundDraftAuditDetails(
	orderID uint,
	transactionID uint,
	requestedAmount float64,
	reason string,
	lineItemCount int,
	restockCount int,
	refund *paymentdomain.Refund,
) map[string]interface{} {
	details := map[string]interface{}{
		"order_id":                orderID,
		"transaction_id":          transactionID,
		"requested_amount":        requestedAmount,
		"reason_present":          strings.TrimSpace(reason) != "",
		"reason_length":           len(strings.TrimSpace(reason)),
		"line_item_count":         lineItemCount,
		"restock_line_item_count": restockCount,
		"gateway_refund_executed": false,
	}
	if refund != nil {
		details["refund_id"] = refund.ID
		details["order_id"] = refund.OrderID
		details["transaction_id"] = refund.TransactionID
		details["requested_amount"] = refund.RequestedAmount
		details["net_amount"] = refund.Amount
		details["discount_clawback_amount"] = refund.DiscountClawbackAmount
		details["refund_status"] = refund.Status
		details["line_item_count"] = len(refund.LineItems)
	}
	return details
}

func stripeDisputeEvidenceAuditDetails(
	disputeID uint,
	submit bool,
	includeCustomerCommunication bool,
	additionalStatement string,
	shippingDocumentationFileID string,
	customerCommunicationFileID string,
	receiptFileID string,
	uncategorizedFileID string,
	result *service.SubmitStripeDisputeEvidenceResult,
) map[string]interface{} {
	details := map[string]interface{}{
		"dispute_id":                          disputeID,
		"submit":                              submit,
		"include_customer_communication":      includeCustomerCommunication,
		"additional_statement_present":        strings.TrimSpace(additionalStatement) != "",
		"additional_statement_length":         len(strings.TrimSpace(additionalStatement)),
		"shipping_documentation_file_present": strings.TrimSpace(shippingDocumentationFileID) != "",
		"customer_communication_file_present": strings.TrimSpace(customerCommunicationFileID) != "",
		"receipt_file_present":                strings.TrimSpace(receiptFileID) != "",
		"uncategorized_file_present":          strings.TrimSpace(uncategorizedFileID) != "",
	}
	if result != nil {
		details["dispute_id"] = result.DisputeID
		details["stripe_dispute_id_present"] = strings.TrimSpace(result.StripeDisputeID) != ""
		details["stripe_status"] = result.StripeStatus
		details["staged"] = result.Staged
		details["submitted"] = result.SubmittedAt != nil
	}
	return details
}

func paymentReviewAuditDetails(
	reviewID uint,
	status string,
	reason string,
	notes string,
	orderID *uint,
	transactionID *uint,
	disputeID *uint,
	paymentIntentID string,
	assignedToID *uint,
	review *paymentdomain.PaymentReview,
) map[string]interface{} {
	details := map[string]interface{}{
		"review_id":                 reviewID,
		"requested_status":          strings.TrimSpace(status),
		"reason_present":            strings.TrimSpace(reason) != "",
		"reason_length":             len(strings.TrimSpace(reason)),
		"notes_present":             strings.TrimSpace(notes) != "",
		"notes_length":              len(strings.TrimSpace(notes)),
		"payment_intent_id_present": strings.TrimSpace(paymentIntentID) != "",
	}
	if orderID != nil {
		details["order_id"] = *orderID
	}
	if transactionID != nil {
		details["transaction_id"] = *transactionID
	}
	if disputeID != nil {
		details["dispute_id"] = *disputeID
	}
	if assignedToID != nil {
		details["assigned_to_id"] = *assignedToID
	}
	if review != nil {
		details["review_id"] = review.ID
		details["status"] = review.Status
		details["source"] = review.Source
		details["stripe_review_id_present"] = strings.TrimSpace(review.StripeReviewID) != ""
		details["reviewed"] = review.ReviewedAt != nil
		if review.OrderID != nil {
			details["order_id"] = *review.OrderID
		}
		if review.TransactionID != nil {
			details["transaction_id"] = *review.TransactionID
		}
		if review.DisputeID != nil {
			details["dispute_id"] = *review.DisputeID
		}
		if review.AssignedToID != nil {
			details["assigned_to_id"] = *review.AssignedToID
		}
	}
	return details
}

func paymentMethodAuditDetails(methodID uint, method paymentdomain.PaymentMethod) map[string]interface{} {
	details := map[string]interface{}{
		"payment_method_id":   methodID,
		"code":                strings.ToLower(strings.TrimSpace(method.Code)),
		"name_present":        strings.TrimSpace(method.Name) != "",
		"icon_present":        strings.TrimSpace(method.Icon) != "",
		"description_present": strings.TrimSpace(method.Description) != "",
		"description_length":  len(strings.TrimSpace(method.Description)),
		"fee_type":            strings.ToLower(strings.TrimSpace(method.FeeType)),
		"fee_value":           method.FeeValue,
		"min_amount":          method.MinAmount,
		"max_amount":          method.MaxAmount,
		"enabled":             method.Enabled,
		"sort_order":          method.SortOrder,
		"settings_present":    strings.TrimSpace(method.Settings) != "",
		"settings_length":     len(strings.TrimSpace(method.Settings)),
	}
	if method.ID != 0 {
		details["payment_method_id"] = method.ID
	}
	return details
}

func paymentMethodOldValue(method *paymentdomain.PaymentMethod) map[string]interface{} {
	if method == nil {
		return nil
	}
	return paymentMethodAuditDetails(method.ID, *method)
}

func paymentRefundRecommendationAuditDetails(
	recommendationID uint,
	requestedStatus string,
	decisionNotes string,
	recommendation *paymentdomain.PaymentRefundRecommendation,
	refund *paymentdomain.Refund,
) map[string]interface{} {
	details := map[string]interface{}{
		"recommendation_id":       recommendationID,
		"requested_status":        strings.TrimSpace(requestedStatus),
		"decision_notes_present":  strings.TrimSpace(decisionNotes) != "",
		"decision_notes_length":   len(strings.TrimSpace(decisionNotes)),
		"creates_pending_refund":  refund != nil,
		"gateway_refund_executed": false,
	}
	if recommendation != nil {
		details["recommendation_id"] = recommendation.ID
		details["provider"] = recommendation.Provider
		details["source_kind"] = string(recommendation.SourceKind)
		details["recommended_action"] = recommendation.RecommendedAction
		details["recommended_amount"] = recommendation.RecommendedAmount
		details["currency"] = recommendation.Currency
		details["priority"] = recommendation.Priority
		details["status"] = recommendation.Status
		if recommendation.OrderID != nil {
			details["order_id"] = *recommendation.OrderID
		}
		if recommendation.TransactionID != nil {
			details["transaction_id"] = *recommendation.TransactionID
		}
		if recommendation.LinkedRefundID != nil {
			details["linked_refund_id"] = *recommendation.LinkedRefundID
		}
	}
	if refund != nil {
		details["refund_id"] = refund.ID
		details["refund_amount"] = refund.Amount
		details["refund_status"] = refund.Status
		details["order_id"] = refund.OrderID
		details["transaction_id"] = refund.TransactionID
	}
	return details
}

func paymentRiskRecomputeAuditDetails(providers []string, reports map[string]*service.PaymentRiskReport) map[string]interface{} {
	details := map[string]interface{}{
		"providers": providers,
	}
	if len(reports) == 0 {
		return details
	}
	reportSummaries := make(map[string]interface{}, len(reports))
	for provider, report := range reports {
		if report == nil || report.Snapshot == nil {
			reportSummaries[provider] = map[string]interface{}{"snapshot": false}
			continue
		}
		snapshot := report.Snapshot
		reportSummaries[provider] = map[string]interface{}{
			"snapshot":                   true,
			"level":                      snapshot.Level,
			"window_days":                snapshot.WindowDays,
			"successful_payment_count":   snapshot.SuccessfulPaymentCount,
			"dispute_count":              snapshot.DisputeCount,
			"early_fraud_warning_count":  snapshot.EarlyFraudWarningCount,
			"refund_count":               snapshot.RefundCount,
			"dispute_activity_rate":      snapshot.DisputeActivityRate,
			"early_fraud_warning_rate":   snapshot.EarlyFraudWarningRate,
			"refund_rate":                snapshot.RefundRate,
			"recommended_action_present": strings.TrimSpace(snapshot.RecommendedAction) != "",
		}
	}
	details["reports"] = reportSummaries
	return details
}

func paymentProtectionControlRequestAuditDetails(
	action string,
	scopeType string,
	scopeValue string,
	reason string,
	expiresAt time.Time,
) map[string]interface{} {
	details := map[string]interface{}{
		"action":         strings.ToLower(strings.TrimSpace(action)),
		"scope_type":     strings.ToLower(strings.TrimSpace(scopeType)),
		"scope_value":    strings.TrimSpace(scopeValue),
		"reason_present": strings.TrimSpace(reason) != "",
		"reason_length":  len(strings.TrimSpace(reason)),
	}
	if !expiresAt.IsZero() {
		details["expires_at"] = expiresAt.UTC()
	}
	return details
}

func paymentProtectionControlAuditDetails(controlID uint, control *paymentdomain.PaymentProtectionControl) map[string]interface{} {
	details := map[string]interface{}{
		"control_id": controlID,
	}
	if control == nil {
		return details
	}
	details["control_id"] = control.ID
	details["action"] = string(control.Action)
	details["scope_type"] = string(control.ScopeType)
	details["scope_value"] = control.ScopeValue
	details["expires_at"] = control.ExpiresAt.UTC()
	details["enabled"] = control.Enabled
	details["active"] = control.Active
	details["status"] = control.Status
	details["reason_present"] = strings.TrimSpace(control.Reason) != ""
	details["reason_length"] = len(strings.TrimSpace(control.Reason))
	return details
}
