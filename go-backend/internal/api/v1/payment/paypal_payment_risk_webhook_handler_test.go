package payment

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPayPalDisputeHelpersReadTransactionAmountAndTimestamp(t *testing.T) {
	resource := map[string]interface{}{
		"create_time": "2026-07-31T10:15:00Z",
		"dispute_amount": map[string]interface{}{
			"value":         "249.90",
			"currency_code": "USD",
		},
		"disputed_transactions": []interface{}{
			map[string]interface{}{
				"seller_transaction_id": "7AB12345CD678901E",
			},
		},
	}

	amount, currency := paypalDisputeAmount(resource)

	require.InDelta(t, 249.90, amount, 0.000001)
	require.Equal(t, "USD", currency)
	require.Equal(t, "7AB12345CD678901E", paypalDisputePaymentID(resource))
	require.Equal(
		t,
		time.Date(2026, time.July, 31, 10, 15, 0, 0, time.UTC),
		paypalRiskOccurredAt(resource),
	)
}
