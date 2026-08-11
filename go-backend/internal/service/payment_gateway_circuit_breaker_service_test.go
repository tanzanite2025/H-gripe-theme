package service

import (
	"context"
	"testing"
	"time"

	"commerce-platform/internal/pkg/config"

	"github.com/stretchr/testify/require"
)

type fakePaymentGatewayCircuitBreakerStore struct {
	open              bool
	retryAfter        time.Duration
	sampleCount       int64
	failureCount      int64
	openedDuration    time.Duration
	recordedProviders []string
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
) error {
	s.open = true
	s.retryAfter = duration
	s.openedDuration = duration
	return nil
}

func testPaymentGatewayCircuitBreakerConfig() config.PaymentGatewayCircuitBreakerConfig {
	return config.PaymentGatewayCircuitBreakerConfig{
		Enabled:              true,
		WindowSeconds:        60,
		FailureRateThreshold: 0.15,
		MinimumSampleCount:   20,
		OpenDurationSeconds:  30,
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
