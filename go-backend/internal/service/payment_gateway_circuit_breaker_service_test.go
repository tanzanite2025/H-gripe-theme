package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/pkg/config"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type fakePaymentGatewayCircuitBreakerStore struct {
	open              bool
	retryAfter        time.Duration
	probe             bool
	permitToken       string
	sampleCount       int64
	failureCount      int64
	openedDuration    time.Duration
	recordedProviders []string
}

func (s *fakePaymentGatewayCircuitBreakerStore) AcquireGatewayPaymentAttempt(
	_ context.Context,
	_ string,
	_ time.Duration,
) (bool, bool, time.Duration, string, error) {
	if s.open {
		return false, s.probe, s.retryAfter, "", nil
	}
	token := s.permitToken
	if s.probe && token == "" {
		token = "probe-token"
	}
	return true, s.probe, 0, token, nil
}

func (s *fakePaymentGatewayCircuitBreakerStore) GetOpenCircuitRetryAfter(
	_ context.Context,
	_ string,
) (bool, time.Duration, error) {
	return s.open, s.retryAfter, nil
}

func (s *fakePaymentGatewayCircuitBreakerStore) GetGatewayHealthWindowCounts(
	_ context.Context,
	_ string,
	_ time.Time,
	_ time.Duration,
) (int64, int64, error) {
	return s.sampleCount, s.failureCount, nil
}

func (s *fakePaymentGatewayCircuitBreakerStore) RecordGatewayHealthEventAndGetWindowCounts(
	_ context.Context,
	provider string,
	_ time.Time,
	failure bool,
	_ time.Duration,
) (int64, int64, error) {
	s.recordedProviders = append(s.recordedProviders, provider)
	s.sampleCount++
	if failure {
		s.failureCount++
	}
	return s.sampleCount, s.failureCount, nil
}

func (s *fakePaymentGatewayCircuitBreakerStore) OpenGatewayCircuit(
	_ context.Context,
	_ string,
	duration time.Duration,
	_ time.Duration,
	_ string,
) (bool, error) {
	s.open = true
	s.retryAfter = duration
	s.openedDuration = duration
	return true, nil
}

func (s *fakePaymentGatewayCircuitBreakerStore) CloseGatewayCircuitProbe(
	_ context.Context,
	_ string,
	permitToken string,
) (bool, error) {
	if permitToken == "" || (s.permitToken != "" && permitToken != s.permitToken) {
		return false, nil
	}
	s.open = false
	s.probe = false
	return true, nil
}

func (s *fakePaymentGatewayCircuitBreakerStore) ReleaseGatewayCircuitProbe(
	_ context.Context,
	_ string,
	_ string,
) error {
	s.probe = false
	return nil
}

func testPaymentGatewayCircuitBreakerConfig() config.PaymentGatewayCircuitBreakerConfig {
	return config.PaymentGatewayCircuitBreakerConfig{
		Enabled:                 true,
		WindowSeconds:           60,
		FailureRateThreshold:    0.15,
		MinimumSampleCount:      20,
		OpenDurationSeconds:     30,
		HalfOpenProbeTimeoutSec: config.MinimumPaymentGatewayHalfOpenProbeTimeoutSeconds,
	}
}

func TestPaymentGatewayCircuitBreakerOpensOnlyAboveThresholdAfterMinimumSamples(t *testing.T) {
	store := &fakePaymentGatewayCircuitBreakerStore{
		sampleCount:  18,
		failureCount: 2,
	}
	circuitBreaker := newPaymentGatewayCircuitBreakerServiceWithStore(
		store,
		testPaymentGatewayCircuitBreakerConfig(),
	)

	decision, err := circuitBreaker.RecordFailedGatewayAPIResponse(context.Background(), "stripe")
	require.NoError(t, err)
	require.False(t, decision.CircuitOpen, "circuit must wait for minimum sample count")
	require.Equal(t, int64(19), decision.SampleCount)
	require.Equal(t, int64(3), decision.FailureCount)
	require.InDelta(t, 3.0/19.0, decision.FailureRate, 0.0001)
	require.False(t, store.open)

	store.sampleCount = 19
	store.failureCount = 2
	decision, err = circuitBreaker.RecordFailedGatewayAPIResponse(context.Background(), "stripe")
	require.NoError(t, err)
	require.False(t, decision.CircuitOpen, "15%% failure rate must not open the circuit")
	require.Equal(t, int64(20), decision.SampleCount)
	require.Equal(t, int64(3), decision.FailureCount)
	require.InDelta(t, 0.15, decision.FailureRate, 0.0001)
	require.False(t, store.open)

	store.sampleCount = 19
	store.failureCount = 3
	decision, err = circuitBreaker.RecordFailedGatewayAPIResponse(context.Background(), "stripe")
	require.NoError(t, err)
	require.True(t, decision.CircuitOpen)
	require.False(t, decision.Allowed)
	require.Equal(t, 30*time.Second, decision.RetryAfter)
	require.Equal(t, 30*time.Second, store.openedDuration)
}

func TestPaymentGatewayCircuitBreakerAllowsWhenCircuitIsClosed(t *testing.T) {
	store := &fakePaymentGatewayCircuitBreakerStore{}
	circuitBreaker := newPaymentGatewayCircuitBreakerServiceWithStore(
		store,
		testPaymentGatewayCircuitBreakerConfig(),
	)

	decision, err := circuitBreaker.IsGatewayPaymentAttemptAllowed(context.Background(), "stripe")
	require.NoError(t, err)
	require.True(t, decision.Allowed)
	require.False(t, decision.CircuitOpen)
}

func TestPaymentGatewayCircuitBreakerAllowsSingleRedisHalfOpenProbe(t *testing.T) {
	miniRedis := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: miniRedis.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	circuitBreaker := NewPaymentGatewayCircuitBreakerService(client, testPaymentGatewayCircuitBreakerConfig())
	store := redisPaymentGatewayCircuitBreakerStore{client: client}
	opened, err := store.OpenGatewayCircuit(t.Context(), "stripe", 100*time.Millisecond, time.Minute, "")
	require.NoError(t, err)
	require.True(t, opened)
	time.Sleep(150 * time.Millisecond)

	first, err := circuitBreaker.IsGatewayPaymentAttemptAllowed(t.Context(), "stripe")
	require.NoError(t, err)
	require.True(t, first.Allowed)
	require.True(t, first.Probe)
	require.NotEmpty(t, first.PermitToken)

	second, err := circuitBreaker.IsGatewayPaymentAttemptAllowed(t.Context(), "stripe")
	require.NoError(t, err)
	require.False(t, second.Allowed)
	require.True(t, second.CircuitOpen)
	require.True(t, second.Probe)
	require.Positive(t, second.RetryAfter)

	require.NoError(t, circuitBreaker.RecordSuccessfulGatewayAPIResponse(t.Context(), "stripe", first.PermitToken))
	third, err := circuitBreaker.IsGatewayPaymentAttemptAllowed(t.Context(), "stripe")
	require.NoError(t, err)
	require.True(t, third.Allowed)
	require.False(t, third.Probe)
}

func TestPaymentGatewayCircuitBreakerProbeFailureReopensOnlyForOwnerToken(t *testing.T) {
	miniRedis := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: miniRedis.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	circuitBreaker := NewPaymentGatewayCircuitBreakerService(client, testPaymentGatewayCircuitBreakerConfig())
	store := redisPaymentGatewayCircuitBreakerStore{client: client}
	opened, err := store.OpenGatewayCircuit(t.Context(), "paypal", 100*time.Millisecond, time.Minute, "")
	require.NoError(t, err)
	require.True(t, opened)
	time.Sleep(150 * time.Millisecond)

	probe, err := circuitBreaker.IsGatewayPaymentAttemptAllowed(t.Context(), "paypal")
	require.NoError(t, err)
	require.True(t, probe.Allowed)
	require.True(t, probe.Probe)

	_, err = circuitBreaker.RecordFailedGatewayAPIResponse(t.Context(), "paypal", "wrong-token")
	require.NoError(t, err)
	sampleCount, failureCount, err := store.GetGatewayHealthWindowCounts(
		t.Context(),
		"paypal",
		time.Now().UTC(),
		time.Minute,
	)
	require.NoError(t, err)
	require.Zero(t, sampleCount)
	require.Zero(t, failureCount)
	availability, err := circuitBreaker.ReadGatewayPaymentAttemptAvailability(t.Context(), "paypal")
	require.NoError(t, err)
	require.False(t, availability.Allowed)

	decision, err := circuitBreaker.RecordFailedGatewayAPIResponse(t.Context(), "paypal", probe.PermitToken)
	require.NoError(t, err)
	require.True(t, decision.CircuitOpen)
	require.False(t, decision.Allowed)
	require.Equal(t, 30*time.Second, decision.RetryAfter)
}

func TestPaymentGatewayCircuitBreakerReturnsRetryAfterWhenCircuitIsOpen(t *testing.T) {
	store := &fakePaymentGatewayCircuitBreakerStore{
		open:       true,
		retryAfter: 17 * time.Second,
	}
	circuitBreaker := newPaymentGatewayCircuitBreakerServiceWithStore(
		store,
		testPaymentGatewayCircuitBreakerConfig(),
	)

	decision, err := circuitBreaker.IsGatewayPaymentAttemptAllowed(context.Background(), "paypal")
	require.NoError(t, err)
	require.False(t, decision.Allowed)
	require.True(t, decision.CircuitOpen)
	require.Equal(t, 17*time.Second, decision.RetryAfter)
}

func TestPaymentGatewayHealthEventsUseRedisTime(t *testing.T) {
	miniRedis := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: miniRedis.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	now := time.Date(2026, time.August, 19, 12, 0, 0, 0, time.UTC)
	miniRedis.SetTime(now)
	store := redisPaymentGatewayCircuitBreakerStore{client: client}

	_, _, err := store.RecordGatewayHealthEventAndGetWindowCounts(
		t.Context(),
		"stripe",
		now.Add(time.Hour),
		true,
		time.Minute,
	)
	require.NoError(t, err)

	events, err := client.ZRangeWithScores(t.Context(), paymentGatewayHealthEventsKey("stripe"), 0, -1).Result()
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, float64(now.UnixMilli()), events[0].Score)
}

func TestPaymentGatewayCircuitBreakerKeysShareRedisClusterHashTag(t *testing.T) {
	provider := "stripe"
	keys := []string{
		paymentGatewayHealthEventsKey(provider),
		paymentGatewayHealthFailuresKey(provider),
		paymentGatewayCircuitKey(provider),
		paymentGatewayCircuitProbeKey(provider),
	}

	expectedTag := redisClusterHashTagForTest(keys[0])
	require.NotEmpty(t, expectedTag)
	for _, key := range keys[1:] {
		require.Equal(t, expectedTag, redisClusterHashTagForTest(key))
	}
}

func redisClusterHashTagForTest(key string) string {
	start := strings.IndexByte(key, '{')
	if start < 0 {
		return ""
	}
	endOffset := strings.IndexByte(key[start+1:], '}')
	if endOffset < 0 {
		return ""
	}
	return key[start+1 : start+1+endOffset]
}
