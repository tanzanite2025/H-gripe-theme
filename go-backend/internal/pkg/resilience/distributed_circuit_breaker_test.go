package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestDistributedCircuitBreakerSharesOpenStateAndHalfOpenProbe(t *testing.T) {
	miniRedis := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: miniRedis.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	now := time.Date(2026, time.August, 19, 12, 0, 0, 0, time.UTC)
	miniRedis.SetTime(now)
	config := DistributedCircuitBreakerConfig{
		Enabled:          true,
		FailureThreshold: 2,
		FailureWindow:    time.Minute,
		OpenDuration:     5 * time.Second,
		ProbeTimeout:     10 * time.Second,
	}
	firstReplica := NewDistributedCircuitBreaker(client, config)
	secondReplica := NewDistributedCircuitBreaker(client, config)

	permit, err := firstReplica.Acquire(t.Context(), "tracking:17track:https://api.example.test")
	require.NoError(t, err)
	permit.RecordFailure(t.Context())

	permit, err = secondReplica.Acquire(t.Context(), "tracking:17track:https://api.example.test")
	require.NoError(t, err)
	permit.RecordFailure(t.Context())

	_, err = firstReplica.Acquire(t.Context(), "tracking:17track:https://api.example.test")
	require.ErrorIs(t, err, ErrCircuitOpen)
	require.Greater(t, RetryAfter(err), time.Duration(0))

	now = now.Add(6 * time.Second)
	miniRedis.SetTime(now)
	probe, err := firstReplica.Acquire(t.Context(), "tracking:17track:https://api.example.test")
	require.NoError(t, err, "only one replica may make the half-open probe")

	_, err = secondReplica.Acquire(t.Context(), "tracking:17track:https://api.example.test")
	require.ErrorIs(t, err, ErrCircuitOpen)

	probe.RecordSuccess(t.Context())
	_, err = secondReplica.Acquire(t.Context(), "tracking:17track:https://api.example.test")
	require.NoError(t, err, "the shared circuit should close after the successful probe")
}

func TestDistributedCircuitBreakerDoesNotLetInflightSuccessCloseOpenCircuit(t *testing.T) {
	miniRedis := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: miniRedis.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	breaker := NewDistributedCircuitBreaker(client, DistributedCircuitBreakerConfig{
		Enabled:          true,
		FailureThreshold: 1,
		FailureWindow:    time.Minute,
		OpenDuration:     time.Minute,
		ProbeTimeout:     time.Minute,
	})

	failingRequest, err := breaker.Acquire(t.Context(), "outbox-webhook:https://example.test/events")
	require.NoError(t, err)
	inflightSuccess, err := breaker.Acquire(t.Context(), "outbox-webhook:https://example.test/events")
	require.NoError(t, err)

	failingRequest.RecordFailure(t.Context())
	inflightSuccess.RecordSuccess(t.Context())

	_, err = breaker.Acquire(t.Context(), "outbox-webhook:https://example.test/events")
	require.ErrorIs(t, err, ErrCircuitOpen)
}

func TestDistributedCircuitBreakerReturnsStoreUnavailableWhenRedisCannotBeRead(t *testing.T) {
	miniRedis := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: miniRedis.Addr()})
	require.NoError(t, client.Close())

	breaker := NewDistributedCircuitBreaker(client, DistributedCircuitBreakerConfig{Enabled: true})
	_, err := breaker.Acquire(context.Background(), "google-merchant-api")
	require.True(t, errors.Is(err, ErrCircuitStoreUnavailable))
}

func TestDistributedCircuitBreakerFailsClosedWhenEnabledWithoutRedis(t *testing.T) {
	breaker := NewDistributedCircuitBreaker(nil, DistributedCircuitBreakerConfig{Enabled: true})

	_, err := breaker.Acquire(context.Background(), "google-merchant-api")
	require.ErrorIs(t, err, ErrCircuitStoreUnavailable)
}

func TestDistributedCircuitBreakerAllowsWhenDisabledWithoutRedis(t *testing.T) {
	breaker := NewDistributedCircuitBreaker(nil, DistributedCircuitBreakerConfig{Enabled: false})

	permit, err := breaker.Acquire(context.Background(), "google-merchant-api")
	require.NoError(t, err)
	require.False(t, permit.IsProbe())
}
