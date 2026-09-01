package payment

import (
	"testing"

	pgateway "commerce-platform/internal/pkg/payment"

	"github.com/stretchr/testify/require"
)

func TestStripeLiabilityShiftedFromPaymentIntentSucceeded(t *testing.T) {
	shifted := stripeLiabilityShiftedFromPaymentIntentSucceeded([]byte(`{
		"id": "pi_1",
		"latest_charge": {
			"payment_method_details": {
				"card": {
					"three_d_secure": {
						"liability_shifted": false
					}
				}
			}
		}
	}`), nil)

	require.NotNil(t, shifted)
	require.False(t, *shifted)

	shifted = stripeLiabilityShiftedFromPaymentIntentSucceeded(nil, []byte(`{
		"id": "evt_1",
		"data": {
			"object": {
				"metadata": {
					"liability_shifted": "true"
				}
			}
		}
	}`))

	require.NotNil(t, shifted)
	require.True(t, *shifted)
}

func TestStripeLiabilityShiftedDoesNotInferFromThreeDSecureResult(t *testing.T) {
	shifted := stripeLiabilityShiftedFromPaymentIntentSucceeded([]byte(`{
		"id": "pi_1",
		"latest_charge": {
			"payment_method_details": {
				"card": {
					"three_d_secure": {
						"result": "authenticated"
					}
				}
			}
		}
	}`), nil)

	require.Nil(t, shifted)
}

func TestPayPalVerifiedPaymentFromEventExtractsLiabilityShifted(t *testing.T) {
	resource := []byte(`{
		"id": "PAYPAL-ORDER-1",
		"status": "COMPLETED",
		"purchase_units": [{
			"custom_id": "ORD-PAYPAL-LIABILITY",
			"payments": {
				"captures": [{
					"id": "PAYPAL-CAPTURE-1",
					"status": "COMPLETED",
					"amount": {
						"currency_code": "USD",
						"value": "84.00"
					},
					"liability_shifted": false
				}]
			}
		}]
	}`)

	payment, handled, err := paypalVerifiedPaymentFromEvent(pgateway.PayPalWebhookEvent{
		EventType: paypalCheckoutOrderCompleted,
		Resource:  resource,
	}, nil)

	require.NoError(t, err)
	require.True(t, handled)
	require.NotNil(t, payment.LiabilityShifted)
	require.False(t, *payment.LiabilityShifted)
}

func TestPaymentResponseLiabilityShiftedPrefersExplicitField(t *testing.T) {
	notShifted := false
	shifted := paymentResponseLiabilityShifted(&pgateway.PaymentResponse{
		LiabilityShifted: &notShifted,
		Metadata: map[string]string{
			"liability_shifted": "true",
		},
	}, nil)

	require.NotNil(t, shifted)
	require.False(t, *shifted)
}
