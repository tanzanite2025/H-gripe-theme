package app

import (
	"time"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/resilience"

	"github.com/redis/go-redis/v9"
)

type outboundHTTPResilience struct {
	retry   resilience.HTTPRetryPolicy
	breaker resilience.CircuitController
}

func newOutboundHTTPResilience(
	redisClient redis.UniversalClient,
	cfg config.OutboundHTTPResilienceConfig,
) outboundHTTPResilience {
	retry := resilience.HTTPRetryPolicy{
		MaxAttempts: cfg.RetryMaxAttempts,
		Backoff: resilience.BackoffPolicy{
			BaseDelay: time.Duration(cfg.RetryBaseDelayMillis) * time.Millisecond,
			MaxDelay:  time.Duration(cfg.RetryMaxDelayMillis) * time.Millisecond,
			Jitter:    time.Duration(cfg.RetryJitterMillis) * time.Millisecond,
		},
		RetryUnsafeMethods: true,
	}
	breaker := resilience.NewDistributedCircuitBreaker(redisClient, resilience.DistributedCircuitBreakerConfig{
		Enabled:          cfg.Enabled,
		FailureThreshold: cfg.FailureThreshold,
		FailureWindow:    time.Duration(cfg.FailureWindowSeconds) * time.Second,
		OpenDuration:     time.Duration(cfg.OpenDurationSeconds) * time.Second,
		ProbeTimeout:     time.Duration(cfg.HalfOpenProbeTimeoutSec) * time.Second,
	})
	return outboundHTTPResilience{
		retry:   retry,
		breaker: breaker,
	}
}
