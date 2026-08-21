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
			AdaptiveEnabled:  true,
			LowRiskMaxAmount: 100,
			AVSBillingShippingMismatchHighValueThresholdUSD: 900,
			TrustedPaidOrders:   1,
			VisitorRiskLookback: 30,
			StepUpRiskScore:     20,
			ChallengeRiskScore:  60,
		},
	)

	require.Equal(t, 900.0, policy.PolicyView().AVSBillingShippingMismatchHighValueThresholdUSD)
}

func TestBuildPaymentRiskConfigurationViewIncludesAvsThreshold(t *testing.T) {
	policy := NewPaymentThreeDSPolicyService(
		&fakeThreeDSOrderHistory{},
		&fakeThreeDSVisitorRisk{},
		&fakeThreeDSPaymentRisk{},
		config.PaymentThreeDSConfig{
			AdaptiveEnabled:  true,
			LowRiskMaxAmount: 100,
			AVSBillingShippingMismatchHighValueThresholdUSD: 900,
			TrustedPaidOrders:   1,
			VisitorRiskLookback: 30,
			StepUpRiskScore:     20,
			ChallengeRiskScore:  60,
		},
	)

	view := BuildPaymentRiskConfigurationView(nil, nil, policy, nil, nil)

	require.Equal(t, 900.0, view.ThreeDS.AVSBillingShippingMismatchHighValueThresholdUSD)
}
