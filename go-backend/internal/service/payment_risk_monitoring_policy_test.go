package service

import (
	"testing"
	"time"

	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/config"

	"github.com/stretchr/testify/require"
)

func TestPaymentRiskMonitoringPolicyKeepsSmallSamplesNormal(t *testing.T) {
	policy := paymentRiskPolicy{config: paymentRiskMonitoringTestConfig()}
	now := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)

	snapshot := policy.Evaluate(PaymentRiskMetrics{
		Provider:               string(paymentdomain.PaymentRiskProviderStripe),
		WindowDays:             30,
		WindowStart:            now.AddDate(0, 0, -30),
		WindowEnd:              now,
		SuccessfulPaymentCount: 2,
		DisputeCount:           1,
	}, now)

	require.Equal(t, paymentdomain.PaymentRiskLevelNormal, snapshot.Level)
	require.Equal(t, "collect_more_volume_and_review_efw", snapshot.RecommendedAction)
	require.InDelta(t, 0.5, snapshot.DisputeActivityRate, 0.000001)
	require.Empty(t, decodePaymentRiskReasons(snapshot.ReasonsJSON))
}

func TestPaymentRiskMonitoringPolicyEscalatesCriticalDisputeActivity(t *testing.T) {
	policy := paymentRiskPolicy{config: paymentRiskMonitoringTestConfig()}
	now := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)

	snapshot := policy.Evaluate(PaymentRiskMetrics{
		Provider:               string(paymentdomain.PaymentRiskProviderStripe),
		WindowDays:             30,
		WindowStart:            now.AddDate(0, 0, -30),
		WindowEnd:              now,
		SuccessfulPaymentCount: 1000,
		DisputeCount:           8,
	}, now)

	require.Equal(t, paymentdomain.PaymentRiskLevelCritical, snapshot.Level)
	require.Equal(t, "force_3ds_for_high_risk_and_operator_review", snapshot.RecommendedAction)
	require.InDelta(t, 0.008, snapshot.DisputeActivityRate, 0.000001)
	require.Equal(t, []string{"dispute_activity_rate_critical"}, decodePaymentRiskReasons(snapshot.ReasonsJSON))
}

func TestPaymentRiskMonitoringPolicyEscalatesRefundWarning(t *testing.T) {
	policy := paymentRiskPolicy{config: paymentRiskMonitoringTestConfig()}
	now := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)

	snapshot := policy.Evaluate(PaymentRiskMetrics{
		Provider:               string(paymentdomain.PaymentRiskProviderPayPal),
		WindowDays:             30,
		WindowStart:            now.AddDate(0, 0, -30),
		WindowEnd:              now,
		SuccessfulPaymentCount: 1000,
		RefundCount:            80,
	}, now)

	require.Equal(t, paymentdomain.PaymentRiskLevelWarning, snapshot.Level)
	require.Equal(t, "force_3ds_for_high_risk_and_notify_operator", snapshot.RecommendedAction)
	require.InDelta(t, 0.08, snapshot.RefundRate, 0.000001)
	require.Equal(t, []string{"refund_rate_warning"}, decodePaymentRiskReasons(snapshot.ReasonsJSON))
}

func paymentRiskMonitoringTestConfig() config.PaymentRiskMonitoringConfig {
	return config.PaymentRiskMonitoringConfig{
		Enabled:                     true,
		WindowDays:                  30,
		MinimumSuccessfulPayments:   20,
		WarningDisputeActivityRate:  0.005,
		CriticalDisputeActivityRate: 0.008,
		WarningEarlyFraudRate:       0.005,
		CriticalEarlyFraudRate:      0.009,
		WarningRefundRate:           0.08,
		CriticalRefundRate:          0.15,
		AutoStepUpEnabled:           true,
	}
}
