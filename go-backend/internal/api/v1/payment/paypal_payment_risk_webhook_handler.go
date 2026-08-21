package payment

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	paymentdomain "commerce-platform/internal/domain/payment"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) recordPayPalDisputeRiskEvent(c *gin.Context, event pgateway.PayPalWebhookEvent, payload []byte) (bool, error) {
	eventType := strings.ToUpper(strings.TrimSpace(event.EventType))
	if !strings.HasPrefix(eventType, "CUSTOMER.DISPUTE.") {
		return false, nil
	}

	resource := map[string]interface{}{}
	if len(event.Resource) > 0 {
		if err := json.Unmarshal(event.Resource, &resource); err != nil {
			return true, err
		}
	}

	externalReference := firstJSONPathString(resource, "dispute_id", "id")
	if externalReference == "" {
		externalReference = event.ID
	}
	providerPaymentID := paypalDisputePaymentID(resource)
	amount, currency := paypalDisputeAmount(resource)

	metadata := map[string]string{
		"event_type": eventType,
		"status":     firstJSONPathString(resource, "status"),
		"reason":     firstJSONPathString(resource, "reason"),
	}
	if orderReference := paypalDisputeOrderReference(resource); orderReference != "" {
		metadata["order_reference"] = orderReference
	}
	orderReference := metadata["order_reference"]

	riskInput := service.PaymentRiskEventInput{
		Provider:          string(pgateway.GatewayPayPal),
		Kind:              paymentdomain.PaymentRiskEventDispute,
		ExternalReference: externalReference,
		WebhookEventID:    event.ID,
		ProviderPaymentID: providerPaymentID,
		Amount:            amount,
		Currency:          currency,
		OccurredAt:        paypalRiskOccurredAt(resource),
		Payload:           string(payload),
		Metadata:          metadata,
	}
	if err := h.recordAndRefreshPaymentRiskEvent(riskInput); err != nil {
		return true, err
	}
	if err := h.enqueueRefundRecommendation(riskInput); err != nil {
		return true, err
	}

	dispute, err := h.paymentService.RecordPayPalDispute(service.PayPalDisputeInput{
		PayPalDisputeID:       externalReference,
		ProviderPaymentID:     providerPaymentID,
		OrderReference:        orderReference,
		Amount:                amount,
		Currency:              currency,
		Reason:                firstJSONPathString(resource, "reason"),
		Status:                firstJSONPathString(resource, "status"),
		DisputeState:          firstJSONPathString(resource, "dispute_state"),
		DisputeLifeCycleStage: firstJSONPathString(resource, "dispute_life_cycle_stage"),
		RawPayload:            string(payload),
	})
	if err != nil {
		return true, err
	}

	result, submitErr := h.autoSubmitPayPalDisputeEvidence(c, dispute)
	c.Set("paypal_dispute_id", dispute.ID)
	c.Set("paypal_dispute_external_id", dispute.PayPalDisputeID)
	if result != nil {
		c.Set("paypal_dispute_evidence_submitted", true)
		c.Set("paypal_dispute_evidence_tracking_number", result.TrackingNumber)
	}
	if submitErr != nil {
		c.Set("paypal_dispute_evidence_submitted", false)
		c.Set("paypal_dispute_evidence_error", submitErr.Error())
	}
	return true, nil
}

func (h *Handler) autoSubmitPayPalDisputeEvidence(c *gin.Context, dispute *paymentdomain.PayPalDispute) (*service.SubmitPayPalDisputeEvidenceResult, error) {
	if h == nil || h.paymentService == nil || dispute == nil {
		return nil, nil
	}
	config, err := h.loadPaymentGatewayConfiguration(pgateway.GatewayPayPal)
	if err != nil {
		return nil, err
	}
	result, err := h.paymentService.SubmitPayPalDisputeEvidence(c.Request.Context(), service.SubmitPayPalDisputeEvidenceInput{
		DisputeID:   dispute.ID,
		ClientID:    config.APIKey,
		SecretKey:   config.SecretKey,
		Environment: config.Environment,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func firstJSONPathString(resource map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		value, ok := resource[key]
		if !ok {
			continue
		}
		if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
			return strings.TrimSpace(text)
		}
	}
	return ""
}

func paypalDisputePaymentID(resource map[string]interface{}) string {
	transactions, ok := resource["disputed_transactions"].([]interface{})
	if !ok {
		return firstJSONPathString(resource, "transaction_id", "seller_transaction_id")
	}
	for _, raw := range transactions {
		transaction, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if value := firstJSONPathString(transaction, "seller_transaction_id", "buyer_transaction_id", "transaction_id"); value != "" {
			return value
		}
	}
	return ""
}

func paypalDisputeOrderReference(resource map[string]interface{}) string {
	if value := firstJSONPathString(resource, "invoice_id", "custom_id", "reference_id"); value != "" {
		return value
	}
	transactions, ok := resource["disputed_transactions"].([]interface{})
	if !ok {
		return ""
	}
	for _, raw := range transactions {
		transaction, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if value := firstJSONPathString(transaction, "invoice_id", "custom_id", "reference_id"); value != "" {
			return value
		}
		items, ok := transaction["items"].([]interface{})
		if !ok {
			continue
		}
		for _, rawItem := range items {
			item, ok := rawItem.(map[string]interface{})
			if !ok {
				continue
			}
			if value := firstJSONPathString(item, "invoice_id", "custom_id", "reference_id"); value != "" {
				return value
			}
		}
	}
	return ""
}

func paypalDisputeAmount(resource map[string]interface{}) (float64, string) {
	for _, key := range []string{"dispute_amount", "amount"} {
		value, ok := resource[key].(map[string]interface{})
		if !ok {
			continue
		}
		rawAmount, _ := value["value"].(string)
		if rawAmount == "" {
			if numeric, ok := value["value"].(float64); ok {
				rawAmount = strconv.FormatFloat(numeric, 'f', -1, 64)
			}
		}
		amount, err := strconv.ParseFloat(strings.TrimSpace(rawAmount), 64)
		if err != nil || amount <= 0 {
			continue
		}
		currency := firstJSONPathString(value, "currency_code", "currency")
		return amount, strings.ToUpper(currency)
	}
	return 0, ""
}

func paypalRiskOccurredAt(resource map[string]interface{}) time.Time {
	for _, key := range []string{"create_time", "update_time"} {
		if value := firstJSONPathString(resource, key); value != "" {
			if parsed, err := time.Parse(time.RFC3339, value); err == nil {
				return parsed.UTC()
			}
		}
	}
	return time.Now().UTC()
}
