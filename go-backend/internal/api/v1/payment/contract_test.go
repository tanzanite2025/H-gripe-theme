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
		AmountMinor:     1234,
		Currency:        "USD",
		GatewayResponse: `{"client_secret":"pi_secret"}`,
	}))
	if err != nil {
		t.Fatalf("marshal transaction response: %v", err)
	}

	payload := string(body)
	if strings.Contains(payload, "gateway_response") || strings.Contains(payload, "pi_secret") {
		t.Fatalf("transaction response leaked gateway response: %s", payload)
	}
	if !strings.Contains(payload, `"amount_minor":1234`) || strings.Contains(payload, `"amount":`) {
		t.Fatalf("transaction response must expose canonical minor amount only: %s", payload)
	}
}

func TestRefundResponseOmitsGatewayResponse(t *testing.T) {
	providerRefundID := "re_contract_1"
	body, err := json.Marshal(refundToResponse(paymentdomain.Refund{
		ID:                             1,
		OrderID:                        2,
		TransactionID:                  3,
		RefundID:                       &providerRefundID,
		SettlementAmountMinor:          115000,
		SettlementCurrency:             "USD",
		SettlementBalanceTransactionID: "txn_contract_1",
		FXGainLossMinor:                10000,
		FXGainLossCurrency:             "USD",
		GatewayResponse:                `{"secret":"refund_secret"}`,
		CalculationSnapshot:            `{"internal":"refund_policy"}`,
	}))
	if err != nil {
		t.Fatalf("marshal refund response: %v", err)
	}

	payload := string(body)
	if strings.Contains(payload, "gateway_response") || strings.Contains(payload, "refund_secret") || strings.Contains(payload, "calculation_snapshot") || strings.Contains(payload, "refund_policy") {
		t.Fatalf("refund response leaked gateway response: %s", payload)
	}
	for _, fragment := range []string{`"settlement_amount_minor":115000`, `"settlement_currency":"USD"`, `"settlement_balance_transaction_id":"txn_contract_1"`, `"fx_gain_loss_minor":10000`, `"fx_gain_loss_currency":"USD"`} {
		if !strings.Contains(payload, fragment) {
			t.Fatalf("refund response missing settlement field %s: %s", fragment, payload)
		}
	}
}

func TestRefundLineItemResponseUsesMinorUnitsOnly(t *testing.T) {
	body, err := json.Marshal(refundToResponse(paymentdomain.Refund{
		Currency: "USD",
		LineItems: []paymentdomain.RefundLineItem{{
			Currency:          "USD",
			UnitPriceMinor:    1234,
			LineSubtotalMinor: 2468,
			LineTaxMinor:      200,
			LineDiscountMinor: 100,
			LineTotalMinor:    2568,
		}},
	}))
	if err != nil {
		t.Fatalf("marshal refund response: %v", err)
	}
	payload := string(body)
	if !strings.Contains(payload, `"unit_price_minor":1234`) || !strings.Contains(payload, `"line_total_minor":2568`) {
		t.Fatalf("refund line response missing canonical minor fields: %s", payload)
	}
	for _, legacy := range []string{"unit_price", "line_subtotal_amount", "line_tax_amount", "line_discount_amount", "line_total_amount"} {
		if strings.Contains(payload, `"`+legacy+`"`) {
			t.Fatalf("refund line response exposed legacy field %s: %s", legacy, payload)
		}
	}
}
