package service

import (
	"sort"
	"time"

	paymentdomain "tanzanite/internal/domain/payment"
	"tanzanite/internal/pkg/config"
)

type PaymentRiskMetrics struct {
	Provider                string
	WindowDays              int
	WindowStart             time.Time
	WindowEnd               time.Time
	SuccessfulPaymentCount  int64
	SuccessfulPaymentAmount float64
	DisputeCount            int64
	DisputeAmount           float64
	EarlyFraudWarningCount  int64
	RefundCount             int64
	RefundAmount            float64
}

type PaymentRiskCheckoutPolicy struct {
	Provider   string
	Level      paymentdomain.PaymentRiskLevel
	Force3DS   bool
	Reason     string
	ComputedAt time.Time
}

type paymentRiskPolicy struct {
	config config.PaymentRiskMonitoringConfig
}

func (p paymentRiskPolicy) Evaluate(metrics PaymentRiskMetrics, computedAt time.Time) *paymentdomain.PaymentRiskSnapshot {
	snapshot := &paymentdomain.PaymentRiskSnapshot{
		Provider:                metrics.Provider,
		WindowDays:              metrics.WindowDays,
		WindowStart:             metrics.WindowStart,
		WindowEnd:               metrics.WindowEnd,
		SuccessfulPaymentCount:  metrics.SuccessfulPaymentCount,
		SuccessfulPaymentAmount: metrics.SuccessfulPaymentAmount,
		DisputeCount:            metrics.DisputeCount,
		DisputeAmount:           metrics.DisputeAmount,
		EarlyFraudWarningCount:  metrics.EarlyFraudWarningCount,
		RefundCount:             metrics.RefundCount,
		RefundAmount:            metrics.RefundAmount,
		Level:                   paymentdomain.PaymentRiskLevelNormal,
		RecommendedAction:       "continue_monitoring",
		ComputedAt:              computedAt,
	}

	if metrics.SuccessfulPaymentCount > 0 {
		denominator := float64(metrics.SuccessfulPaymentCount)
		snapshot.DisputeActivityRate = float64(metrics.DisputeCount) / denominator
		snapshot.EarlyFraudWarningRate = float64(metrics.EarlyFraudWarningCount) / denominator
		snapshot.RefundRate = float64(metrics.RefundCount) / denominator
	}

	// A small sample should remain visible but must not open the circuit based
	// on one transaction. This is an internal alert guard, not a gateway rule.
	if metrics.SuccessfulPaymentCount < int64(p.config.MinimumSuccessfulPayments) {
		snapshot.RecommendedAction = "collect_more_volume_and_review_efw"
		return snapshot
	}

	reasons := make([]string, 0, 3)
	if snapshot.DisputeActivityRate >= p.config.CriticalDisputeActivityRate {
		snapshot.Level = paymentdomain.PaymentRiskLevelCritical
		reasons = append(reasons, "dispute_activity_rate_critical")
	} else if snapshot.DisputeActivityRate >= p.config.WarningDisputeActivityRate {
		snapshot.Level = maxPaymentRiskLevel(snapshot.Level, paymentdomain.PaymentRiskLevelWarning)
		reasons = append(reasons, "dispute_activity_rate_warning")
	}
	if snapshot.EarlyFraudWarningRate >= p.config.CriticalEarlyFraudRate {
		snapshot.Level = paymentdomain.PaymentRiskLevelCritical
		reasons = append(reasons, "early_fraud_warning_rate_critical")
	} else if snapshot.EarlyFraudWarningRate >= p.config.WarningEarlyFraudRate {
		snapshot.Level = maxPaymentRiskLevel(snapshot.Level, paymentdomain.PaymentRiskLevelWarning)
		reasons = append(reasons, "early_fraud_warning_rate_warning")
	}
	if snapshot.RefundRate >= p.config.CriticalRefundRate {
		snapshot.Level = paymentdomain.PaymentRiskLevelCritical
		reasons = append(reasons, "refund_rate_critical")
	} else if snapshot.RefundRate >= p.config.WarningRefundRate {
		snapshot.Level = maxPaymentRiskLevel(snapshot.Level, paymentdomain.PaymentRiskLevelWarning)
		reasons = append(reasons, "refund_rate_warning")
	}

	sort.Strings(reasons)
	snapshot.ReasonsJSON = encodePaymentRiskReasons(reasons)
	switch snapshot.Level {
	case paymentdomain.PaymentRiskLevelCritical:
		snapshot.RecommendedAction = "force_3ds_for_high_risk_and_operator_review"
	case paymentdomain.PaymentRiskLevelWarning:
		snapshot.RecommendedAction = "force_3ds_for_high_risk_and_notify_operator"
	}
	return snapshot
}

func maxPaymentRiskLevel(current, next paymentdomain.PaymentRiskLevel) paymentdomain.PaymentRiskLevel {
	if next == paymentdomain.PaymentRiskLevelCritical || current == paymentdomain.PaymentRiskLevelCritical {
		return paymentdomain.PaymentRiskLevelCritical
	}
	if next == paymentdomain.PaymentRiskLevelWarning || current == paymentdomain.PaymentRiskLevelWarning {
		return paymentdomain.PaymentRiskLevelWarning
	}
	return paymentdomain.PaymentRiskLevelNormal
}
