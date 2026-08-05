package payment

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	paymentdomain "tanzanite/internal/domain/payment"
	pgateway "tanzanite/internal/pkg/payment"
	"tanzanite/internal/service"

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
	if orderReference := firstJSONPathString(resource, "invoice_id", "custom_id", "reference_id"); orderReference != "" {
		metadata["order_reference"] = orderReference
	}

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
	if err := h.recordAndRefreshPaymentRiskEvent(c, riskInput); err != nil {
		return true, err
	}
	return true, h.enqueueRefundRecommendation(riskInput)
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
