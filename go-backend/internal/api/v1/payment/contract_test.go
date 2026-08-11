package payment

import (
	paymentdomain "commerce-platform/internal/domain/payment"
	"encoding/json"
	"strings"
	"testing"
)

func TestPaymentMethodResponseOmitsSettings(t *testing.T) {
	body, err := json.Marshal(paymentMethodToResponse(paymentdomain.PaymentMethod{
		ID:       1,
		Name:     "Stripe",
		Code:     "stripe",
		Settings: `{"secret_key":"sk_live_secret"}`,
	}))
	if err != nil {
		t.Fatalf("marshal payment method response: %v", err)
	}

	payload := string(body)
	if strings.Contains(payload, "settings") || strings.Contains(payload, "sk_live_secret") {
		t.Fatalf("payment method response leaked settings: %s", payload)
	}
	if !strings.Contains(payload, `"available"`) {
		t.Fatalf("payment method response missing availability field: %s", payload)
	}
}

func TestTransactionResponseOmitsGatewayResponse(t *testing.T) {
	body, err := json.Marshal(transactionToResponse(paymentdomain.Transaction{
		ID:              1,
		OrderID:         2,
		TransactionID:   "txn_123",
		GatewayResponse: `{"client_secret":"pi_secret"}`,
	}))
	if err != nil {
		t.Fatalf("marshal transaction response: %v", err)
	}

	payload := string(body)
	if strings.Contains(payload, "gateway_response") || strings.Contains(payload, "pi_secret") {
		t.Fatalf("transaction response leaked gateway response: %s", payload)
	}
}

func TestRefundResponseOmitsGatewayResponse(t *testing.T) {
	body, err := json.Marshal(refundToResponse(paymentdomain.Refund{
		ID:                  1,
		OrderID:             2,
		TransactionID:       3,
		GatewayResponse:     `{"secret":"refund_secret"}`,
		CalculationSnapshot: `{"internal":"refund_policy"}`,
	}))
	if err != nil {
		t.Fatalf("marshal refund response: %v", err)
	}

	payload := string(body)
	if strings.Contains(payload, "gateway_response") || strings.Contains(payload, "refund_secret") || strings.Contains(payload, "calculation_snapshot") || strings.Contains(payload, "refund_policy") {
		t.Fatalf("refund response leaked gateway response: %s", payload)
	}
}
