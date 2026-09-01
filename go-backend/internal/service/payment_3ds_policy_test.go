package service

import (
	"context"
	"errors"
	"testing"

	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/domain/visitor"
	"commerce-platform/internal/pkg/antifraud"
	"commerce-platform/internal/pkg/config"
	appLogger "commerce-platform/internal/pkg/logger"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestPaymentThreeDSPolicyMarksLowAmountAsExemptionCandidate(t *testing.T) {
	policy := newTestPaymentThreeDSPolicy(
		config.PaymentThreeDSConfig{
			AdaptiveEnabled:     true,
			LowRiskMaxAmount:    100,
			TrustedPaidOrders:   1,
			VisitorRiskLookback: 30,
			StepUpRiskScore:     20,
			ChallengeRiskScore:  60,
		},
	)

	decision := policy.Decide(context.Background(), PaymentThreeDSDecisionInput{
		UserID:    10,
		OrderID:   99,
		Amount:    80,
		Currency:  "USD",
		BaseMode:  PaymentThreeDSModeAutomatic,
		IPAddress: "203.0.113.10",
		UserAgent: "Mozilla/5.0",
	})

	require.Equal(t, PaymentThreeDSModeAutomatic, decision.Mode)
	require.True(t, decision.ExemptionCandidate)
	require.Equal(t, "adaptive_exemption_candidate", decision.Strategy)
	require.Contains(t, decision.Reasons, "low_amount")
}

func TestPaymentThreeDSPolicyMarksTrustedCustomerAsExemptionCandidate(t *testing.T) {
	policy := newTestPaymentThreeDSPolicy(
		config.PaymentThreeDSConfig{
			AdaptiveEnabled:     true,
			LowRiskMaxAmount:    100,
			TrustedPaidOrders:   2,
			VisitorRiskLookback: 30,
			StepUpRiskScore:     20,
			ChallengeRiskScore:  60,
		},
	)
	policy.orderHistory.(*fakeThreeDSOrderHistory).count = 2

	decision := policy.Decide(context.Background(), PaymentThreeDSDecisionInput{
		UserID:    10,
		OrderID:   99,
		Amount:    800,
		Currency:  "USD",
		BaseMode:  PaymentThreeDSModeAutomatic,
		IPAddress: "203.0.113.10",
		UserAgent: "Mozilla/5.0",
	})

	require.Equal(t, PaymentThreeDSModeAutomatic, decision.Mode)
	require.True(t, decision.ExemptionCandidate)
	require.Equal(t, "adaptive_exemption_candidate", decision.Strategy)
	require.Contains(t, decision.Reasons, "trusted_customer")
}

func TestPaymentThreeDSPolicyIgnoresHighValueAvsRuleWhenCountriesMatch(t *testing.T) {
	policy := newTestPaymentThreeDSPolicy(
		config.PaymentThreeDSConfig{
			AdaptiveEnabled:     true,
			LowRiskMaxAmount:    0,
			TrustedPaidOrders:   1,
			VisitorRiskLookback: 30,
			StepUpRiskScore:     20,
			ChallengeRiskScore:  60,
		},
	)

	decision := policy.Decide(context.Background(), PaymentThreeDSDecisionInput{
		UserID:          10,
		OrderID:         99,
		Amount:          1000,
		Currency:        "USD",
		BaseMode:        PaymentThreeDSModeAutomatic,
		IPAddress:       "203.0.113.10",
		UserAgent:       "Mozilla/5.0",
		BillingCountry:  "US",
		ShippingCountry: "US",
	})

	require.Equal(t, PaymentThreeDSModeAutomatic, decision.Mode)
	require.Equal(t, "adaptive_default", decision.Strategy)
	require.NotContains(t, decision.Reasons, paymentThreeDSBillingShippingMismatchReason)
}

func TestPaymentThreeDSPolicyHonorsConfiguredAvsThreshold(t *testing.T) {
	policy := newTestPaymentThreeDSPolicy(
		config.PaymentThreeDSConfig{
			AdaptiveEnabled:  true,
			LowRiskMaxAmount: 0,
			AVSBillingShippingMismatchHighValueThresholdUSD: 1000,
			TrustedPaidOrders:   1,
			VisitorRiskLookback: 30,
			StepUpRiskScore:     20,
			ChallengeRiskScore:  60,
		},
	)

	decision := policy.Decide(context.Background(), PaymentThreeDSDecisionInput{
		UserID:          10,
		OrderID:         99,
		Amount:          800,
		Currency:        "USD",
		BaseMode:        PaymentThreeDSModeAutomatic,
		IPAddress:       "203.0.113.10",
		UserAgent:       "Mozilla/5.0",
		BillingCountry:  "US",
		ShippingCountry: "DE",
	})

	require.Equal(t, PaymentThreeDSModeAutomatic, decision.Mode)
	require.Equal(t, "adaptive_default", decision.Strategy)
	require.NotContains(t, decision.Reasons, paymentThreeDSBillingShippingMismatchReason)
}

func TestPaymentThreeDSPolicyChallengesHighValueAvsMismatchEvenWhenAdaptiveRiskIsDisabled(t *testing.T) {
	policy := newTestPaymentThreeDSPolicy(
		config.PaymentThreeDSConfig{
			AdaptiveEnabled: false,
		},
	)

	decision := policy.Decide(context.Background(), PaymentThreeDSDecisionInput{
		UserID:          10,
		OrderID:         99,
		Amount:          801,
		Currency:        "USD",
		BaseMode:        PaymentThreeDSModeAutomatic,
		IPAddress:       "203.0.113.10",
		UserAgent:       "Mozilla/5.0",
		BillingCountry:  "US",
		ShippingCountry: "DE",
	})

	require.Equal(t, PaymentThreeDSModeChallenge, decision.Mode)
	require.Equal(t, "adaptive_challenge", decision.Strategy)
	require.Equal(t, visitor.RiskLevelSuspicious, decision.RiskLevel)
	require.Contains(t, decision.Reasons, paymentThreeDSBillingShippingMismatchReason)
	require.Contains(t, decision.Reasons, "adaptive_3ds_disabled")
}

func TestPaymentThreeDSPolicyChallengesHighPaymentRisk(t *testing.T) {
	policy := newTestPaymentThreeDSPolicy(
		config.PaymentThreeDSConfig{
			AdaptiveEnabled:     true,
			LowRiskMaxAmount:    100,
			TrustedPaidOrders:   1,
			VisitorRiskLookback: 30,
			StepUpRiskScore:     20,
			ChallengeRiskScore:  60,
		},
	)
	policy.paymentRisk.(*fakeThreeDSPaymentRisk).decision = antifraud.Decision{
		Score:    75,
		HighRisk: true,
	}

	decision := policy.Decide(context.Background(), PaymentThreeDSDecisionInput{
		UserID:    10,
		OrderID:   99,
		Amount:    80,
		Currency:  "USD",
		BaseMode:  PaymentThreeDSModeAutomatic,
		IPAddress: "203.0.113.10",
		UserAgent: "Mozilla/5.0",
	})

	require.Equal(t, PaymentThreeDSModeChallenge, decision.Mode)
	require.False(t, decision.ExemptionCandidate)
	require.Equal(t, "adaptive_challenge", decision.Strategy)
	require.Contains(t, decision.Reasons, "payment_risk_high")
}

func TestPaymentThreeDSPolicyStepsUpVisitorWatchDecision(t *testing.T) {
	policy := newTestPaymentThreeDSPolicy(
		config.PaymentThreeDSConfig{
			AdaptiveEnabled:     true,
			LowRiskMaxAmount:    100,
			TrustedPaidOrders:   1,
			VisitorRiskLookback: 30,
			StepUpRiskScore:     20,
			ChallengeRiskScore:  60,
		},
	)
	policy.visitorRisk.(*fakeThreeDSVisitorRisk).assessment = VisitorRiskIdentityAssessment{
		Known:          true,
		RiskLevel:      visitor.RiskLevelNormal,
		DecisionAction: visitor.RiskDecisionActionWatch,
	}

	decision := policy.Decide(context.Background(), PaymentThreeDSDecisionInput{
		UserID:    10,
		OrderID:   99,
		Amount:    80,
		Currency:  "USD",
		BaseMode:  PaymentThreeDSModeAutomatic,
		IPAddress: "203.0.113.10",
		UserAgent: "Mozilla/5.0",
	})

	require.Equal(t, PaymentThreeDSModeAny, decision.Mode)
	require.False(t, decision.ExemptionCandidate)
	require.Equal(t, "adaptive_step_up", decision.Strategy)
	require.Contains(t, decision.Reasons, "visitor_risk_manual_watch")
}

func TestPaymentThreeDSPolicyForwardsDeviceFingerprintToVisitorRisk(t *testing.T) {
	policy := newTestPaymentThreeDSPolicy(
		config.PaymentThreeDSConfig{
			AdaptiveEnabled:     true,
			LowRiskMaxAmount:    100,
			TrustedPaidOrders:   1,
			VisitorRiskLookback: 30,
			StepUpRiskScore:     20,
			ChallengeRiskScore:  60,
		},
	)
	visitorRisk := policy.visitorRisk.(*fakeThreeDSVisitorRisk)

	decision := policy.Decide(context.Background(), PaymentThreeDSDecisionInput{
		UserID:            10,
		OrderID:           99,
		Amount:            200,
		Currency:          "USD",
		BaseMode:          PaymentThreeDSModeAutomatic,
		IPAddress:         "203.0.113.10",
		DeviceFingerprint: "browser-fingerprint-hash",
		UserAgent:         "Mozilla/5.0",
	})

	require.Equal(t, PaymentThreeDSModeAutomatic, decision.Mode)
	require.Equal(t, "browser-fingerprint-hash", visitorRisk.lastInput.DeviceFingerprint)
	require.Equal(t, "203.0.113.10", visitorRisk.lastInput.IPAddress)
}

func TestPaymentThreeDSPolicyChallengesVisitorBlockCandidate(t *testing.T) {
	policy := newTestPaymentThreeDSPolicy(
		config.PaymentThreeDSConfig{
			AdaptiveEnabled:     true,
			LowRiskMaxAmount:    100,
			TrustedPaidOrders:   1,
			VisitorRiskLookback: 30,
			StepUpRiskScore:     20,
			ChallengeRiskScore:  60,
		},
	)
	policy.visitorRisk.(*fakeThreeDSVisitorRisk).assessment = VisitorRiskIdentityAssessment{
		Known:          true,
		RiskLevel:      visitor.RiskLevelNormal,
		DecisionAction: visitor.RiskDecisionActionBlockCandidate,
	}

	decision := policy.Decide(context.Background(), PaymentThreeDSDecisionInput{
		UserID:    10,
		OrderID:   99,
		Amount:    80,
		Currency:  "USD",
		BaseMode:  PaymentThreeDSModeAutomatic,
		IPAddress: "203.0.113.10",
		UserAgent: "Mozilla/5.0",
	})

	require.Equal(t, PaymentThreeDSModeChallenge, decision.Mode)
	require.False(t, decision.ExemptionCandidate)
	require.Equal(t, "adaptive_challenge", decision.Strategy)
	require.Contains(t, decision.Reasons, "visitor_risk_manual_challenge")
}

func TestPaymentThreeDSPolicyStepsUpWhenRiskServiceUnavailable(t *testing.T) {
	core, observedLogs := observer.New(zap.ErrorLevel)
	previousLogger := appLogger.Log
	appLogger.Log = zap.New(core)
	t.Cleanup(func() {
		appLogger.Log = previousLogger
	})

	policy := newTestPaymentThreeDSPolicy(
		config.PaymentThreeDSConfig{
			AdaptiveEnabled:     true,
			LowRiskMaxAmount:    100,
			TrustedPaidOrders:   1,
			VisitorRiskLookback: 30,
			StepUpRiskScore:     20,
			ChallengeRiskScore:  60,
		},
	)
	alerts := &fakeThreeDSFailOpenAlerts{}
	policy.ConfigureFailOpenAlertPublisher(alerts)
	policy.paymentRisk.(*fakeThreeDSPaymentRisk).err = errors.New("redis unavailable")

	decision := policy.Decide(context.Background(), PaymentThreeDSDecisionInput{
		Provider:  "Stripe",
		UserID:    10,
		OrderID:   99,
		Amount:    80,
		Currency:  "USD",
		BaseMode:  PaymentThreeDSModeAutomatic,
		IPAddress: "203.0.113.10",
		UserAgent: "Mozilla/5.0",
	})

	require.Equal(t, PaymentThreeDSModeAny, decision.Mode)
	require.False(t, decision.ExemptionCandidate)
	require.Equal(t, "adaptive_step_up", decision.Strategy)
	require.Contains(t, decision.Reasons, "payment_risk_unavailable")

	require.Len(t, alerts.inputs, 1)
	require.Equal(t, "stripe", alerts.inputs[0].Provider)
	require.Equal(t, "antifraud", alerts.inputs[0].Component)
	require.Equal(t, "payment_3ds_policy_evaluate", alerts.inputs[0].Operation)
	require.Equal(t, "payment_risk_unavailable", alerts.inputs[0].Reason)
	require.Equal(t, "step_up_3ds_any", alerts.inputs[0].FallbackAction)

	entries := observedLogs.FilterMessage("payment risk protection fail-open; Redis-backed antifraud unavailable").All()
	require.Len(t, entries, 1)
	fields := entries[0].ContextMap()
	require.Equal(t, "critical", fields["severity"])
	require.Equal(t, true, fields["fail_open"])
	require.Equal(t, true, fields["redis_unavailable"])
	require.Equal(t, "stripe", fields["provider"])
}

func TestPaymentThreeDSPolicyStepsUpForProviderPortfolioRisk(t *testing.T) {
	policy := newTestPaymentThreeDSPolicy(
		config.PaymentThreeDSConfig{
			AdaptiveEnabled:     true,
			LowRiskMaxAmount:    100,
			TrustedPaidOrders:   1,
			VisitorRiskLookback: 30,
			StepUpRiskScore:     20,
			ChallengeRiskScore:  60,
		},
	)
	policy.ConfigureRiskMonitoring(&fakeThreeDSPortfolioRisk{
		policy: PaymentRiskCheckoutPolicy{
			Provider: string(paymentdomain.PaymentRiskProviderStripe),
			Level:    paymentdomain.PaymentRiskLevelWarning,
			Force3DS: true,
		},
	})

	decision := policy.Decide(context.Background(), PaymentThreeDSDecisionInput{
		Provider:  string(paymentdomain.PaymentRiskProviderStripe),
		UserID:    10,
		OrderID:   99,
		Amount:    80,
		Currency:  "USD",
		BaseMode:  PaymentThreeDSModeAutomatic,
		IPAddress: "203.0.113.10",
		UserAgent: "Mozilla/5.0",
	})

	require.Equal(t, PaymentThreeDSModeAny, decision.Mode)
	require.False(t, decision.ExemptionCandidate)
	require.Equal(t, "warning", decision.PortfolioRiskLevel)
	require.Contains(t, decision.Reasons, "portfolio_risk_step_up")
}

func TestPaymentThreeDSPolicyAppliesManualForce3DSWhenAdaptiveRiskIsDisabled(t *testing.T) {
	policy := NewPaymentThreeDSPolicyService(
		&fakeThreeDSOrderHistory{},
		&fakeThreeDSVisitorRisk{
			assessment: VisitorRiskIdentityAssessment{
				RiskLevel: visitor.RiskLevelNormal,
			},
		},
		&fakeThreeDSPaymentRisk{},
		config.PaymentThreeDSConfig{
			AdaptiveEnabled: false,
		},
	)
	protection := &fakeThreeDSPaymentProtection{
		decision: paymentdomain.PaymentProtectionDecision{
			Force3DS: true,
			Reasons:  []string{"manual_force_3ds_control_41"},
		},
	}
	policy.ConfigurePaymentProtection(protection)

	decision := policy.Decide(context.Background(), PaymentThreeDSDecisionInput{
		Provider:       string(paymentdomain.PaymentRiskProviderStripe),
		BaseMode:       PaymentThreeDSModeAutomatic,
		BillingCountry: "US",
		PaymentMethod:  "card",
		Amount:         80,
		Currency:       "USD",
		IPAddress:      "203.0.113.10",
		UserAgent:      "Mozilla/5.0",
	})

	require.Equal(t, PaymentThreeDSModeAny, decision.Mode)
	require.Equal(t, "manual_protection", decision.Strategy)
	require.Contains(t, decision.Reasons, "manual_force_3ds_control_41")
	require.Contains(t, decision.Reasons, "adaptive_3ds_disabled")
	require.Equal(t, "stripe", protection.input.Provider)
	require.Equal(t, "US", protection.input.Country)
	require.Equal(t, "card", protection.input.PaymentMethod)
}

func TestPaymentThreeDSPolicyPreservesManualProtectionStrategyWithAdaptiveRiskEnabled(t *testing.T) {
	policy := newTestPaymentThreeDSPolicy(
		config.PaymentThreeDSConfig{
			AdaptiveEnabled:     true,
			LowRiskMaxAmount:    100,
			TrustedPaidOrders:   1,
			VisitorRiskLookback: 30,
			StepUpRiskScore:     20,
			ChallengeRiskScore:  60,
		},
	)
	protection := &fakeThreeDSPaymentProtection{
		decision: paymentdomain.PaymentProtectionDecision{
			Force3DS: true,
			Reasons:  []string{"manual_force_3ds_control_42"},
		},
	}
	policy.ConfigurePaymentProtection(protection)

	decision := policy.Decide(context.Background(), PaymentThreeDSDecisionInput{
		Provider:       string(paymentdomain.PaymentRiskProviderStripe),
		BaseMode:       PaymentThreeDSModeAutomatic,
		BillingCountry: "US",
		PaymentMethod:  "card",
		Amount:         800,
		Currency:       "USD",
		IPAddress:      "203.0.113.10",
		UserAgent:      "Mozilla/5.0",
	})

	require.Equal(t, PaymentThreeDSModeAny, decision.Mode)
	require.Equal(t, "manual_protection", decision.Strategy)
	require.Contains(t, decision.Reasons, "manual_force_3ds_control_42")
}

func TestPaymentThreeDSPolicyMetadataIsAuditSafe(t *testing.T) {
	decision := PaymentThreeDSDecision{
		Mode:               PaymentThreeDSModeAutomatic,
		Strategy:           "adaptive_exemption_candidate",
		ExemptionCandidate: true,
		RiskLevel:          visitor.RiskLevelNormal,
		RiskScore:          5,
		Reasons:            []string{"low_amount"},
	}

	metadata := decision.Metadata()

	require.Equal(t, PaymentThreeDSModeAutomatic, metadata["three_ds_mode"])
	require.Equal(t, "adaptive_exemption_candidate", metadata["three_ds_policy"])
	require.Equal(t, "true", metadata["three_ds_exemption_candidate"])
	require.Equal(t, visitor.RiskLevelNormal, metadata["three_ds_risk_level"])
	require.Equal(t, "5", metadata["three_ds_risk_score"])
	require.Equal(t, "normal", metadata["three_ds_portfolio_risk_level"])
	require.Equal(t, "low_amount", metadata["three_ds_reasons"])
}

func newTestPaymentThreeDSPolicy(cfg config.PaymentThreeDSConfig) *PaymentThreeDSPolicyService {
	return NewPaymentThreeDSPolicyService(
		&fakeThreeDSOrderHistory{},
		&fakeThreeDSVisitorRisk{
			assessment: VisitorRiskIdentityAssessment{
				RiskLevel: visitor.RiskLevelNormal,
			},
		},
		&fakeThreeDSPaymentRisk{},
		cfg,
	)
}

type fakeThreeDSOrderHistory struct {
	count int64
	err   error
}

func (f *fakeThreeDSOrderHistory) CountPaidOrdersForUserBefore(userID uint, excludeOrderID uint) (int64, error) {
	return f.count, f.err
}

type fakeThreeDSVisitorRisk struct {
	assessment VisitorRiskIdentityAssessment
	lastInput  VisitorRiskIdentityAssessmentInput
	err        error
}

func (f *fakeThreeDSVisitorRisk) AssessIdentity(ctx context.Context, input VisitorRiskIdentityAssessmentInput) (VisitorRiskIdentityAssessment, error) {
	f.lastInput = input
	if f.assessment.RiskLevel == "" {
		f.assessment.RiskLevel = visitor.RiskLevelNormal
	}
	return f.assessment, f.err
}

type fakeThreeDSPaymentRisk struct {
	decision antifraud.Decision
	err      error
}

func (f *fakeThreeDSPaymentRisk) Evaluate(ctx context.Context, key string, signals antifraud.Signals) (antifraud.Decision, error) {
	return f.decision, f.err
}

type fakeThreeDSPortfolioRisk struct {
	policy   PaymentRiskCheckoutPolicy
	err      error
	provider string
}

func (f *fakeThreeDSPortfolioRisk) CurrentCheckoutPolicy(provider string) (PaymentRiskCheckoutPolicy, error) {
	f.provider = provider
	return f.policy, f.err
}

type fakeThreeDSPaymentProtection struct {
	decision paymentdomain.PaymentProtectionDecision
	err      error
	input    paymentdomain.PaymentProtectionEvaluationInput
}

func (f *fakeThreeDSPaymentProtection) Evaluate(input paymentdomain.PaymentProtectionEvaluationInput) (paymentdomain.PaymentProtectionDecision, error) {
	f.input = input
	return f.decision, f.err
}

type fakeThreeDSFailOpenAlerts struct {
	inputs []PaymentRiskFailOpenAlertInput
	err    error
}

func (f *fakeThreeDSFailOpenAlerts) PublishPaymentRiskFailOpenAlert(ctx context.Context, input PaymentRiskFailOpenAlertInput) error {
	f.inputs = append(f.inputs, input)
	return f.err
}
