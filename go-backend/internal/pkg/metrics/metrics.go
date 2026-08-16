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
	ProductCacheInvalidations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "product_cache_invalidations_total",
			Help:      "Product detail cache invalidation attempts by source and result.",
		},
		[]string{"source", "result"},
	)
	ProductCacheInvalidationKeys = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "commerce_platform",
			Name:      "product_cache_invalidation_keys",
			Help:      "Number of Redis keys deleted per product detail cache invalidation.",
			Buckets:   []float64{0, 1, 2, 5, 10, 25, 50, 100, 250, 500},
		},
		[]string{"source"},
	)
	CustomerServiceRealtimeWebSocketConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_realtime_websocket_connections",
			Help:      "Active customer-service WebSocket connections on this API instance.",
		},
	)
	CustomerServiceRealtimeWebSocketConnectionAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_realtime_websocket_connection_attempts_total",
			Help:      "Customer-service WebSocket connection attempts by result.",
		},
		[]string{"result"},
	)
	CustomerServiceRealtimeWebSocketOutboundOverflows = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_realtime_websocket_outbound_overflows_total",
			Help:      "Customer-service WebSocket connections closed because their outbound queue filled.",
		},
	)
	CustomerServiceRealtimeHubDeliveries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_realtime_hub_deliveries_total",
			Help:      "In-process customer-service realtime fanout attempts by subscription scope and result.",
		},
		[]string{"scope", "result"},
	)
	CustomerServiceRealtimeRelayPublishes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_realtime_relay_publishes_total",
			Help:      "Customer-service Redis Stream publish attempts by result.",
		},
		[]string{"result"},
	)
	CustomerServiceRealtimeRelayReads = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_realtime_relay_reads_total",
			Help:      "Customer-service Redis Stream reads by result.",
		},
		[]string{"result"},
	)
	CustomerServiceRealtimeRelayEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_realtime_relay_events_total",
			Help:      "Customer-service Redis Stream events handled by delivery path and result.",
		},
		[]string{"path", "result"},
	)
	CustomerServiceRealtimeReplayRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_realtime_replay_requests_total",
			Help:      "Customer-service Redis Stream replay requests by result.",
		},
		[]string{"result"},
	)
	CustomerServiceRealtimeOutboxDeliveries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_realtime_outbox_deliveries_total",
			Help:      "Customer-service realtime Outbox handler deliveries by result.",
		},
		[]string{"result"},
	)
	CustomerServiceRealtimeOutboxEvents = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_realtime_outbox_events",
			Help:      "Customer-service realtime Outbox events by durable status.",
		},
		[]string{"status"},
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
	register(ProductCacheInvalidations, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			ProductCacheInvalidations = existing
		}
	})
	register(ProductCacheInvalidationKeys, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.HistogramVec); ok {
			ProductCacheInvalidationKeys = existing
		}
	})
	register(CustomerServiceRealtimeWebSocketConnections, func(collector prometheus.Collector) {
		if existing, ok := collector.(prometheus.Gauge); ok {
			CustomerServiceRealtimeWebSocketConnections = existing
		}
	})
	register(CustomerServiceRealtimeWebSocketConnectionAttempts, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			CustomerServiceRealtimeWebSocketConnectionAttempts = existing
		}
	})
	register(CustomerServiceRealtimeWebSocketOutboundOverflows, func(collector prometheus.Collector) {
		if existing, ok := collector.(prometheus.Counter); ok {
			CustomerServiceRealtimeWebSocketOutboundOverflows = existing
		}
	})
	register(CustomerServiceRealtimeHubDeliveries, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			CustomerServiceRealtimeHubDeliveries = existing
		}
	})
	register(CustomerServiceRealtimeRelayPublishes, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			CustomerServiceRealtimeRelayPublishes = existing
		}
	})
	register(CustomerServiceRealtimeRelayReads, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			CustomerServiceRealtimeRelayReads = existing
		}
	})
	register(CustomerServiceRealtimeRelayEvents, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			CustomerServiceRealtimeRelayEvents = existing
		}
	})
	register(CustomerServiceRealtimeReplayRequests, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			CustomerServiceRealtimeReplayRequests = existing
		}
	})
	register(CustomerServiceRealtimeOutboxDeliveries, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			CustomerServiceRealtimeOutboxDeliveries = existing
		}
	})
	register(CustomerServiceRealtimeOutboxEvents, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.GaugeVec); ok {
			CustomerServiceRealtimeOutboxEvents = existing
		}
	})
}
