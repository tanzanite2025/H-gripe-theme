package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HTTPRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "http_requests_total",
			Help:      "Total HTTP requests handled by the API.",
		},
		[]string{"method", "route", "status"},
	)
	HTTPDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "commerce_platform",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
	VerificationSendAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "verification_send_attempts_total",
			Help:      "Verification message send attempts by channel and result.",
		},
		[]string{"channel", "result"},
	)
	VerificationBudgetRejections = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "verification_budget_rejections_total",
			Help:      "Verification sends rejected by per-identity limits or the global budget.",
		},
		[]string{"channel", "reason"},
	)
	PaymentAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "payment_attempts_total",
			Help:      "Payment or checkout attempts by result.",
		},
		[]string{"provider", "result"},
	)
	PaymentRiskDelayed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "payment_risk_delayed_total",
			Help:      "Payment or checkout attempts delayed by the carding risk policy.",
		},
	)
	CommercialIntelligenceActions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "commercial_intelligence_actions_total",
			Help:      "Commercial intelligence behavior protection actions by seed and outcome.",
		},
		[]string{"seed_id", "outcome"},
	)
)

func init() {
	register := func(collector prometheus.Collector, replace func(prometheus.Collector)) {
		if err := prometheus.Register(collector); err != nil {
			if alreadyRegistered, ok := err.(prometheus.AlreadyRegisteredError); ok {
				replace(alreadyRegistered.ExistingCollector)
			}
		}
	}

	register(HTTPRequests, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			HTTPRequests = existing
		}
	})
	register(HTTPDuration, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.HistogramVec); ok {
			HTTPDuration = existing
		}
	})
	register(VerificationSendAttempts, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			VerificationSendAttempts = existing
		}
	})
	register(VerificationBudgetRejections, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			VerificationBudgetRejections = existing
		}
	})
	register(PaymentAttempts, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			PaymentAttempts = existing
		}
	})
	register(PaymentRiskDelayed, func(collector prometheus.Collector) {
		if existing, ok := collector.(prometheus.Counter); ok {
			PaymentRiskDelayed = existing
		}
	})
	register(CommercialIntelligenceActions, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			CommercialIntelligenceActions = existing
		}
	})
}
