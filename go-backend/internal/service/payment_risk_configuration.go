package service

import (
	"context"
	"time"

	"commerce-platform/internal/pkg/config"
)

type PaymentRiskThreeDSConfigurationView struct {
	Enabled                                         bool    `json:"enabled"`
	RuntimeAvailable                                bool    `json:"runtime_available"`
	AdaptiveEnabled                                 bool    `json:"adaptive_enabled"`
	LowRiskMaxAmount                                float64 `json:"low_risk_max_amount"`
	AVSBillingShippingMismatchHighValueThresholdUSD float64 `json:"avs_billing_shipping_mismatch_high_value_threshold_usd"`
	TrustedPaidOrders                               int     `json:"trusted_paid_orders"`
	VisitorRiskLookbackDays                         int     `json:"visitor_risk_lookback_days"`
	StepUpRiskScore                                 int     `json:"step_up_risk_score"`
	ChallengeRiskScore                              int     `json:"challenge_risk_score"`
}

type PaymentRiskEngineConfigurationView struct {
	Enabled              bool `json:"enabled"`
	FailureWindowSeconds int  `json:"failure_window_seconds"`
	FailureThreshold     int  `json:"failure_threshold"`
	DelaySeconds         int  `json:"delay_seconds"`
	HighRiskScore        int  `json:"high_risk_score"`
}

type PaymentBINRateLimitConfigurationView struct {
	Enabled              bool `json:"enabled"`
	WindowSeconds        int  `json:"window_seconds"`
	FailureThreshold     int  `json:"failure_threshold"`
	BlockDurationSeconds int  `json:"block_duration_seconds"`
}

type PaymentGatewayCircuitBreakerConfigurationView struct {
	Enabled              bool    `json:"enabled"`
	WindowSeconds        int     `json:"window_seconds"`
	FailureRateThreshold float64 `json:"failure_rate_threshold"`
	MinimumSampleCount   int     `json:"minimum_sample_count"`
	OpenDurationSeconds  int     `json:"open_duration_seconds"`
}

type PaymentProtectionConfigurationView struct {
	Enabled                            bool `json:"enabled"`
	MaxControlDurationHours            int  `json:"max_control_duration_hours"`
	MaxPausePaymentDurationHours       int  `json:"max_pause_payment_duration_hours"`
	MaxGlobalPausePaymentDurationHours int  `json:"max_global_pause_payment_duration_hours"`
}

type PaymentRiskAntiAbuseConfigurationView struct {
	TurnstileRequired                    bool `json:"turnstile_required"`
	TurnstileConfigured                  bool `json:"turnstile_configured"`
	VerificationIPWindowSeconds          int  `json:"verification_ip_window_seconds"`
	VerificationDestinationWindowSeconds int  `json:"verification_destination_window_seconds"`
	VerificationDailyLimit               int  `json:"verification_daily_limit"`
	VerificationGlobalWindowSeconds      int  `json:"verification_global_window_seconds"`
	VerificationGlobalLimit              int  `json:"verification_global_limit"`
	VerificationCircuitSeconds           int  `json:"verification_circuit_seconds"`
}

type PaymentRiskOrderAbuseConfigurationView struct {
	Enabled                     bool `json:"enabled"`
	OrderCreateWindowSeconds    int  `json:"order_create_window_seconds"`
	MaxOrderCreationsPerUser    int  `json:"max_order_creations_per_user"`
	MaxOrderCreationsPerSession int  `json:"max_order_creations_per_session"`
	MaxOrderCreationsPerIP      int  `json:"max_order_creations_per_ip"`
}

type PaymentRiskVisitorConfigurationView struct {
	Enabled              bool `json:"enabled"`
	FlushIntervalSeconds int  `json:"flush_interval_seconds"`
	MaxPendingFacts      int  `json:"max_pending_facts"`
	SamplePathLimit      int  `json:"sample_path_limit"`
	RetentionDays        int  `json:"retention_days"`
}

type PaymentRiskMonitoringConfigurationView struct {
	PaymentRiskMonitoringPolicyView
	WorkerEnabled         bool `json:"worker_enabled"`
	WorkerIntervalSeconds int  `json:"worker_interval_seconds"`
}

type PaymentRiskConfigurationView struct {
	Monitoring            PaymentRiskMonitoringConfigurationView        `json:"monitoring"`
	ThreeDS               PaymentRiskThreeDSConfigurationView           `json:"three_ds"`
	PaymentRisk           PaymentRiskEngineConfigurationView            `json:"payment_risk"`
	BINRateLimit          PaymentBINRateLimitConfigurationView          `json:"bin_rate_limit"`
	GatewayCircuitBreaker PaymentGatewayCircuitBreakerConfigurationView `json:"gateway_circuit_breaker"`
	Protection            PaymentProtectionConfigurationView            `json:"protection"`
	AntiAbuse             PaymentRiskAntiAbuseConfigurationView         `json:"anti_abuse"`
	OrderAbuse            PaymentRiskOrderAbuseConfigurationView        `json:"order_abuse"`
	VisitorRisk           PaymentRiskVisitorConfigurationView           `json:"visitor_risk"`
}

type PaymentGatewayHealthView struct {
	Provider          string  `json:"provider"`
	Enabled           bool    `json:"enabled"`
	Allowed           bool    `json:"allowed"`
	CircuitOpen       bool    `json:"circuit_open"`
	FailureRate       float64 `json:"failure_rate"`
	SampleCount       int64   `json:"sample_count"`
	FailureCount      int64   `json:"failure_count"`
	RetryAfterSeconds int64   `json:"retry_after_seconds"`
	Error             string  `json:"error,omitempty"`
}

func BuildPaymentRiskConfigurationView(
	cfg *config.Config,
	monitoring *PaymentRiskMonitoringService,
	threeDS *PaymentThreeDSPolicyService,
	protection *PaymentProtectionService,
	circuitBreaker *PaymentGatewayCircuitBreakerService,
) PaymentRiskConfigurationView {
	var source config.Config
	if cfg != nil {
		source = *cfg
	}

	monitoringPolicy := PaymentRiskMonitoringPolicyView{}
	if monitoring != nil {
		monitoringPolicy = monitoring.PolicyView()
	} else {
		monitoringPolicy = paymentRiskMonitoringPolicyViewFromConfig(
			normalizePaymentRiskMonitoringConfig(source.PaymentRiskMonitoring),
			false,
		)
	}

	threeDSConfig := normalizePaymentThreeDSConfig(source.PaymentThreeDS)
	threeDSRuntimeAvailable := threeDS != nil
	if threeDS != nil {
		threeDSConfigView := threeDS.PolicyView()
		threeDSConfig = config.PaymentThreeDSConfig{
			AdaptiveEnabled:  threeDSConfigView.AdaptiveEnabled,
			LowRiskMaxAmount: threeDSConfigView.LowRiskMaxAmount,
			AVSBillingShippingMismatchHighValueThresholdUSD: threeDSConfigView.AVSBillingShippingMismatchHighValueThresholdUSD,
			TrustedPaidOrders:   threeDSConfigView.TrustedPaidOrders,
			VisitorRiskLookback: threeDSConfigView.VisitorRiskLookbackDays,
			StepUpRiskScore:     threeDSConfigView.StepUpRiskScore,
			ChallengeRiskScore:  threeDSConfigView.ChallengeRiskScore,
		}
	}

	protectionPolicy := PaymentProtectionConfigurationView{
		Enabled:                            source.PaymentProtection.Enabled,
		MaxControlDurationHours:            source.PaymentProtection.MaxControlDurationHours,
		MaxPausePaymentDurationHours:       source.PaymentProtection.MaxPausePaymentDurationHours,
		MaxGlobalPausePaymentDurationHours: source.PaymentProtection.MaxGlobalPausePaymentDurationHours,
	}
	if protection != nil {
		summary := protection.PolicySummary()
		protectionPolicy = PaymentProtectionConfigurationView{
			Enabled:                            protection.Enabled(),
			MaxControlDurationHours:            intValue(summary["max_control_duration_hours"]),
			MaxPausePaymentDurationHours:       intValue(summary["max_pause_payment_duration_hours"]),
			MaxGlobalPausePaymentDurationHours: intValue(summary["max_global_pause_payment_duration_hours"]),
		}
	}

	circuitConfig := source.PaymentGatewayCircuitBreaker
	if circuitBreaker != nil {
		circuitPolicy := circuitBreaker.PolicyView()
		circuitConfig.Enabled = circuitPolicy.Enabled
		circuitConfig.WindowSeconds = circuitPolicy.WindowSeconds
		circuitConfig.FailureRateThreshold = circuitPolicy.FailureRateThreshold
		circuitConfig.MinimumSampleCount = circuitPolicy.MinimumSampleCount
		circuitConfig.OpenDurationSeconds = circuitPolicy.OpenDurationSeconds
	} else {
		circuitConfig = normalizePaymentGatewayCircuitBreakerConfig(circuitConfig)
	}

	return PaymentRiskConfigurationView{
		Monitoring: PaymentRiskMonitoringConfigurationView{
			PaymentRiskMonitoringPolicyView: monitoringPolicy,
			WorkerEnabled:                   source.Worker.PaymentRiskMonitoringEnabled,
			WorkerIntervalSeconds:           source.Worker.PaymentRiskMonitoringIntervalSeconds,
		},
		ThreeDS: PaymentRiskThreeDSConfigurationView{
			Enabled:          threeDSConfig.AdaptiveEnabled,
			RuntimeAvailable: threeDSRuntimeAvailable,
			AdaptiveEnabled:  threeDSConfig.AdaptiveEnabled,
			LowRiskMaxAmount: threeDSConfig.LowRiskMaxAmount,
			AVSBillingShippingMismatchHighValueThresholdUSD: threeDSConfig.AVSBillingShippingMismatchHighValueThresholdUSD,
			TrustedPaidOrders:       threeDSConfig.TrustedPaidOrders,
			VisitorRiskLookbackDays: threeDSConfig.VisitorRiskLookback,
			StepUpRiskScore:         threeDSConfig.StepUpRiskScore,
			ChallengeRiskScore:      threeDSConfig.ChallengeRiskScore,
		},
		PaymentRisk: PaymentRiskEngineConfigurationView{
			Enabled:              source.PaymentRisk.FailureThreshold > 0 || source.PaymentRisk.HighRiskScore > 0,
			FailureWindowSeconds: source.PaymentRisk.FailureWindowSeconds,
			FailureThreshold:     source.PaymentRisk.FailureThreshold,
			DelaySeconds:         source.PaymentRisk.DelaySeconds,
			HighRiskScore:        source.PaymentRisk.HighRiskScore,
		},
		BINRateLimit: PaymentBINRateLimitConfigurationView{
			Enabled:              source.PaymentBINRateLimit.Enabled,
			WindowSeconds:        source.PaymentBINRateLimit.WindowSeconds,
			FailureThreshold:     source.PaymentBINRateLimit.FailureThreshold,
			BlockDurationSeconds: source.PaymentBINRateLimit.BlockDurationSeconds,
		},
		GatewayCircuitBreaker: PaymentGatewayCircuitBreakerConfigurationView{
			Enabled:              circuitConfig.Enabled,
			WindowSeconds:        circuitConfig.WindowSeconds,
			FailureRateThreshold: circuitConfig.FailureRateThreshold,
			MinimumSampleCount:   circuitConfig.MinimumSampleCount,
			OpenDurationSeconds:  circuitConfig.OpenDurationSeconds,
		},
		Protection: protectionPolicy,
		AntiAbuse: PaymentRiskAntiAbuseConfigurationView{
			TurnstileRequired:                    source.AntiAbuse.TurnstileRequired,
			TurnstileConfigured:                  source.AntiAbuse.TurnstileSecretKey != "",
			VerificationIPWindowSeconds:          source.AntiAbuse.VerificationIPWindowSeconds,
			VerificationDestinationWindowSeconds: source.AntiAbuse.VerificationDestinationWindowSeconds,
			VerificationDailyLimit:               source.AntiAbuse.VerificationDailyLimit,
			VerificationGlobalWindowSeconds:      source.AntiAbuse.VerificationGlobalWindowSeconds,
			VerificationGlobalLimit:              source.AntiAbuse.VerificationGlobalLimit,
			VerificationCircuitSeconds:           source.AntiAbuse.VerificationCircuitSeconds,
		},
		OrderAbuse: PaymentRiskOrderAbuseConfigurationView{
			Enabled:                     source.OrderAbuse.Enabled,
			OrderCreateWindowSeconds:    source.OrderAbuse.OrderCreateWindowSeconds,
			MaxOrderCreationsPerUser:    source.OrderAbuse.MaxOrderCreationsPerUser,
			MaxOrderCreationsPerSession: source.OrderAbuse.MaxOrderCreationsPerSession,
			MaxOrderCreationsPerIP:      source.OrderAbuse.MaxOrderCreationsPerIP,
		},
		VisitorRisk: PaymentRiskVisitorConfigurationView{
			Enabled:              source.VisitorRisk.Enabled,
			FlushIntervalSeconds: source.VisitorRisk.FlushIntervalSeconds,
			MaxPendingFacts:      source.VisitorRisk.MaxPendingFacts,
			SamplePathLimit:      source.VisitorRisk.SamplePathLimit,
			RetentionDays:        source.VisitorRisk.RetentionDays,
		},
	}
}

func paymentRiskMonitoringPolicyViewFromConfig(
	cfg config.PaymentRiskMonitoringConfig,
	alertingEnabled bool,
) PaymentRiskMonitoringPolicyView {
	return PaymentRiskMonitoringPolicyView{
		WindowDays:                  cfg.WindowDays,
		MinimumSuccessfulPayments:   cfg.MinimumSuccessfulPayments,
		WarningDisputeActivityRate:  cfg.WarningDisputeActivityRate,
		CriticalDisputeActivityRate: cfg.CriticalDisputeActivityRate,
		WarningEarlyFraudRate:       cfg.WarningEarlyFraudRate,
		CriticalEarlyFraudRate:      cfg.CriticalEarlyFraudRate,
		WarningRefundRate:           cfg.WarningRefundRate,
		CriticalRefundRate:          cfg.CriticalRefundRate,
		AutoStepUpEnabled:           cfg.AutoStepUpEnabled,
		AlertingEnabled:             alertingEnabled,
	}
}

func intValue(value interface{}) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}

func (s *PaymentGatewayCircuitBreakerService) CurrentHealth(
	ctx context.Context,
	provider string,
) PaymentGatewayHealthView {
	view := PaymentGatewayHealthView{
		Provider: provider,
		Allowed:  true,
	}
	if s == nil {
		return view
	}
	decision, err := s.ReadGatewayHealthWindow(ctx, provider)
	if err != nil {
		view.Error = err.Error()
		return view
	}
	view.Enabled = s.PolicyView().Enabled
	view.Allowed = decision.Allowed
	view.CircuitOpen = decision.CircuitOpen
	view.FailureRate = decision.FailureRate
	view.SampleCount = decision.SampleCount
	view.FailureCount = decision.FailureCount
	view.RetryAfterSeconds = int64(decision.RetryAfter / time.Second)
	return view
}
