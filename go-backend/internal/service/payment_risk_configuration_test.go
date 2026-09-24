package service

import (
	"testing"

	"commerce-platform/internal/pkg/config"

	"github.com/stretchr/testify/require"
)

func TestPaymentThreeDSPolicyViewIncludesAvsThreshold(t *testing.T) {
	policy := NewPaymentThreeDSPolicyService(
		&fakeThreeDSOrderHistory{},
		&fakeThreeDSVisitorRisk{},
		&fakeThreeDSPaymentRisk{},
		config.PaymentThreeDSConfig{
			AdaptiveEnabled:       true,
			LowRiskMaxAmountMinor: 10000,
			AVSBillingShippingMismatchHighValueThresholdMinor: 90000,
			TrustedPaidOrders:   1,
			VisitorRiskLookback: 30,
			StepUpRiskScore:     20,
			ChallengeRiskScore:  60,
		},
	)

	require.Equal(t, int64(90000), policy.PolicyView().AVSBillingShippingMismatchHighValueThresholdMinor)
}

func TestBuildPaymentRiskConfigurationViewIncludesAvsThreshold(t *testing.T) {
	policy := NewPaymentThreeDSPolicyService(
		&fakeThreeDSOrderHistory{},
		&fakeThreeDSVisitorRisk{},
		&fakeThreeDSPaymentRisk{},
		config.PaymentThreeDSConfig{
			AdaptiveEnabled:       true,
			LowRiskMaxAmountMinor: 10000,
			AVSBillingShippingMismatchHighValueThresholdMinor: 90000,
			TrustedPaidOrders:   1,
			VisitorRiskLookback: 30,
			StepUpRiskScore:     20,
			ChallengeRiskScore:  60,
		},
	)

	view := BuildPaymentRiskConfigurationView(nil, nil, policy, nil, nil)

	require.Equal(t, int64(90000), view.ThreeDS.AVSBillingShippingMismatchHighValueThresholdMinor)
}
