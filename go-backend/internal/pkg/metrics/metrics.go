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
	HoneypotBlocked = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "honeypot_blocked_total",
			Help:      "Form submissions silently dropped after a honeypot decoy was filled.",
		},
		[]string{"form", "field"},
	)
	HoneypotTimingEvaluations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "honeypot_timing_evaluations_total",
			Help:      "Server-signed form timing token evaluations by bounded result.",
		},
		[]string{"form", "result"},
	)
	HoneypotTimingSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "commerce_platform",
			Name:      "honeypot_timing_seconds",
			Help:      "Observed time from server timing token issuance to form submission.",
			Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 1.5, 2, 3, 5, 10, 30, 60, 300, 600},
		},
		[]string{"form"},
	)
	HoneypotTimingReplays = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "honeypot_timing_replays_total",
			Help:      "Redis-backed shadow checks for repeated server-signed timing tokens.",
		},
		[]string{"form", "result"},
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
	GlobalIPBlockChecks = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "global_ip_block_checks_total",
			Help:      "Application-wide IP block checks by outcome.",
		},
		[]string{"outcome"},
	)
	GlobalIPBlockCacheRefreshes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "global_ip_block_cache_refreshes_total",
			Help:      "Application-wide IP block cache refresh attempts by result.",
		},
		[]string{"result"},
	)
	GlobalIPBlockCacheInvalidations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "global_ip_block_cache_invalidations_total",
			Help:      "Application-wide IP block cache invalidation notifications by direction and result.",
		},
		[]string{"direction", "result"},
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
	CustomerServiceArchivedConversations = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_archived_conversations_total",
			Help:      "Customer-service conversation inbox archive transitions.",
		},
	)
	CustomerServiceReopenedConversations = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_reopened_conversations_total",
			Help:      "Customer-service conversation transitions back to open/inbox.",
		},
	)
	CustomerServiceRetentionSoftDeleted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_soft_deleted_conversations_total",
			Help:      "Customer-service conversations soft-deleted by retention operations.",
		},
	)
	CustomerServiceRetentionPurged = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_purged_conversations_total",
			Help:      "Customer-service conversations physically purged after retention checks.",
		},
	)
	CustomerServiceRetentionEligibility = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_retention_eligibility_total",
			Help:      "Customer-service retention worker eligibility outcomes.",
		},
		[]string{"result"},
	)
	CustomerServiceRetentionAttachmentReferenceSkips = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_retention_attachment_reference_skips_total",
			Help:      "Customer-service retention attachment references skipped by bounded reason.",
		},
		[]string{"reason"},
	)
	CustomerServiceRetentionSearchIndexDeletes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_retention_search_index_deletes_total",
			Help:      "Customer-service retention search-index deletion outcomes.",
		},
		[]string{"result"},
	)
	CustomerServiceRetentionCleanupOutboxEvents = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "commerce_platform",
			Name:      "customer_service_retention_cleanup_outbox_events",
			Help:      "Customer-service retention cleanup Outbox events by durable status.",
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
	register(HoneypotBlocked, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			HoneypotBlocked = existing
		}
	})
	register(HoneypotTimingEvaluations, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			HoneypotTimingEvaluations = existing
		}
	})
	register(HoneypotTimingSeconds, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.HistogramVec); ok {
			HoneypotTimingSeconds = existing
		}
	})
	register(HoneypotTimingReplays, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			HoneypotTimingReplays = existing
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
	register(GlobalIPBlockChecks, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			GlobalIPBlockChecks = existing
		}
	})
	register(GlobalIPBlockCacheRefreshes, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			GlobalIPBlockCacheRefreshes = existing
		}
	})
	register(GlobalIPBlockCacheInvalidations, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			GlobalIPBlockCacheInvalidations = existing
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
	register(CustomerServiceArchivedConversations, func(collector prometheus.Collector) {
		if existing, ok := collector.(prometheus.Counter); ok {
			CustomerServiceArchivedConversations = existing
		}
	})
	register(CustomerServiceReopenedConversations, func(collector prometheus.Collector) {
		if existing, ok := collector.(prometheus.Counter); ok {
			CustomerServiceReopenedConversations = existing
		}
	})
	register(CustomerServiceRetentionSoftDeleted, func(collector prometheus.Collector) {
		if existing, ok := collector.(prometheus.Counter); ok {
			CustomerServiceRetentionSoftDeleted = existing
		}
	})
	register(CustomerServiceRetentionPurged, func(collector prometheus.Collector) {
		if existing, ok := collector.(prometheus.Counter); ok {
			CustomerServiceRetentionPurged = existing
		}
	})
	register(CustomerServiceRetentionEligibility, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			CustomerServiceRetentionEligibility = existing
		}
	})
	register(CustomerServiceRetentionAttachmentReferenceSkips, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			CustomerServiceRetentionAttachmentReferenceSkips = existing
		}
	})
	register(CustomerServiceRetentionSearchIndexDeletes, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.CounterVec); ok {
			CustomerServiceRetentionSearchIndexDeletes = existing
		}
	})
	register(CustomerServiceRetentionCleanupOutboxEvents, func(collector prometheus.Collector) {
		if existing, ok := collector.(*prometheus.GaugeVec); ok {
			CustomerServiceRetentionCleanupOutboxEvents = existing
		}
	})
}
