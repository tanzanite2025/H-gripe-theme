package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"tanzanite/internal/pkg/config"
	pgateway "tanzanite/internal/pkg/payment"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ErrPaymentGatewayCircuitBreakerStoreUnavailable = errors.New("payment gateway circuit breaker store is unavailable")

type PaymentGatewayCircuitBreakerDecision struct {
	Allowed      bool
	CircuitOpen  bool
	FailureRate  float64
	SampleCount  int64
	FailureCount int64
	RetryAfter   time.Duration
}

type paymentGatewayCircuitBreakerStore interface {
	GetOpenCircuitRetryAfter(context.Context, string) (bool, time.Duration, error)
	GetGatewayHealthWindowCounts(context.Context, string, time.Time, time.Duration) (int64, int64, error)
	RecordGatewayHealthEventAndGetWindowCounts(context.Context, string, time.Time, bool, time.Duration) (int64, int64, error)
	OpenGatewayCircuit(context.Context, string, time.Duration) error
}

type redisPaymentGatewayCircuitBreakerStore struct {
	client *redis.Client
}

const readGatewayHealthWindowCountsScript = `
local cutoff = tonumber(ARGV[1])
redis.call("ZREMRANGEBYSCORE", KEYS[1], 0, cutoff)
redis.call("ZREMRANGEBYSCORE", KEYS[2], 0, cutoff)
return {redis.call("ZCARD", KEYS[1]), redis.call("ZCARD", KEYS[2])}
`

const recordGatewayHealthEventScript = `
local cutoff = tonumber(ARGV[1])
local score = tonumber(ARGV[2])
local member = ARGV[3]
local retentionSeconds = tonumber(ARGV[4])
local failure = ARGV[5]

redis.call("ZREMRANGEBYSCORE", KEYS[1], 0, cutoff)
redis.call("ZREMRANGEBYSCORE", KEYS[2], 0, cutoff)
redis.call("ZADD", KEYS[1], score, member)
if failure == "1" then
	redis.call("ZADD", KEYS[2], score, member)
end
redis.call("EXPIRE", KEYS[1], retentionSeconds)
redis.call("EXPIRE", KEYS[2], retentionSeconds)

return {redis.call("ZCARD", KEYS[1]), redis.call("ZCARD", KEYS[2])}
`

func (s redisPaymentGatewayCircuitBreakerStore) GetOpenCircuitRetryAfter(
	ctx context.Context,
	provider string,
) (bool, time.Duration, error) {
	ttl, err := s.client.TTL(ctx, paymentGatewayCircuitKey(provider)).Result()
	if err != nil {
		return false, 0, err
	}
	if ttl == -1 {
		return true, 0, nil
	}
	if ttl <= 0 {
		return false, 0, nil
	}
	return true, ttl, nil
}

func (s redisPaymentGatewayCircuitBreakerStore) GetGatewayHealthWindowCounts(
	ctx context.Context,
	provider string,
	now time.Time,
	window time.Duration,
) (int64, int64, error) {
	return s.executeGatewayHealthWindowScript(
		ctx,
		readGatewayHealthWindowCountsScript,
		provider,
		now,
		window,
		"",
		false,
	)
}

func (s redisPaymentGatewayCircuitBreakerStore) RecordGatewayHealthEventAndGetWindowCounts(
	ctx context.Context,
	provider string,
	now time.Time,
	failure bool,
	window time.Duration,
) (int64, int64, error) {
	member := fmt.Sprintf("%d:%s", now.UnixMilli(), uuid.NewString())
	return s.executeGatewayHealthWindowScript(
		ctx,
		recordGatewayHealthEventScript,
		provider,
		now,
		window,
		member,
		failure,
	)
}

func (s redisPaymentGatewayCircuitBreakerStore) executeGatewayHealthWindowScript(
	ctx context.Context,
	script string,
	provider string,
	now time.Time,
	window time.Duration,
	member string,
	failure bool,
) (int64, int64, error) {
	windowSeconds := int64(window / time.Second)
	if windowSeconds <= 0 {
		windowSeconds = 1
	}
	values, err := s.client.Eval(
		ctx,
		script,
		[]string{
			paymentGatewayHealthEventsKey(provider),
			paymentGatewayHealthFailuresKey(provider),
		},
		now.Add(-window).UnixMilli(),
		now.UnixMilli(),
		member,
		windowSeconds+5,
		boolToRedisFlag(failure),
	).Int64Slice()
	if err != nil {
		return 0, 0, err
	}
	if len(values) != 2 {
		return 0, 0, errors.New("invalid payment gateway health window response")
	}
	return values[0], values[1], nil
}

func (s redisPaymentGatewayCircuitBreakerStore) OpenGatewayCircuit(
	ctx context.Context,
	provider string,
	duration time.Duration,
) error {
	return s.client.Set(ctx, paymentGatewayCircuitKey(provider), "open", duration).Err()
}

type PaymentGatewayCircuitBreakerService struct {
	store  paymentGatewayCircuitBreakerStore
	config config.PaymentGatewayCircuitBreakerConfig
	now    func() time.Time
}

func NewPaymentGatewayCircuitBreakerService(
	redisClient *redis.Client,
	cfg config.PaymentGatewayCircuitBreakerConfig,
) *PaymentGatewayCircuitBreakerService {
	if redisClient == nil {
		return nil
	}
	return newPaymentGatewayCircuitBreakerServiceWithStore(
		redisPaymentGatewayCircuitBreakerStore{client: redisClient},
		cfg,
	)
}

func newPaymentGatewayCircuitBreakerServiceWithStore(
	store paymentGatewayCircuitBreakerStore,
	cfg config.PaymentGatewayCircuitBreakerConfig,
) *PaymentGatewayCircuitBreakerService {
	return &PaymentGatewayCircuitBreakerService{
		store:  store,
		config: normalizePaymentGatewayCircuitBreakerConfig(cfg),
		now:    time.Now,
	}
}

func (s *PaymentGatewayCircuitBreakerService) IsGatewayPaymentAttemptAllowed(
	ctx context.Context,
	provider string,
) (PaymentGatewayCircuitBreakerDecision, error) {
	decision := PaymentGatewayCircuitBreakerDecision{Allowed: true}
	if s == nil || !s.config.Enabled || s.store == nil {
		return decision, nil
	}

	provider = normalizePaymentGatewayHealthProvider(provider)
	if provider == "" {
		return decision, nil
	}

	open, retryAfter, err := s.store.GetOpenCircuitRetryAfter(ctx, provider)
	if err != nil {
		return decision, fmt.Errorf("%w: read circuit state: %v", ErrPaymentGatewayCircuitBreakerStoreUnavailable, err)
	}
	if !open {
		return decision, nil
	}

	decision = PaymentGatewayCircuitBreakerDecision{
		Allowed:     false,
		CircuitOpen: true,
		RetryAfter:  retryAfter,
	}
	sampleCount, failureCount, countsErr := s.store.GetGatewayHealthWindowCounts(
		ctx,
		provider,
		s.currentTime(),
		s.gatewayHealthWindow(),
	)
	if countsErr == nil {
		healthDecision := paymentGatewayCircuitBreakerDecisionFromWindowCounts(
			sampleCount,
			failureCount,
			s.config.FailureRateThreshold,
			s.config.MinimumSampleCount,
		)
		decision.FailureRate = healthDecision.FailureRate
		decision.SampleCount = healthDecision.SampleCount
		decision.FailureCount = healthDecision.FailureCount
	}
	return decision, nil
}

func (s *PaymentGatewayCircuitBreakerService) RecordSuccessfulGatewayAPIResponse(
	ctx context.Context,
	provider string,
) error {
	if s == nil || !s.config.Enabled || s.store == nil {
		return nil
	}
	provider = normalizePaymentGatewayHealthProvider(provider)
	if provider == "" {
		return nil
	}
	_, _, err := s.store.RecordGatewayHealthEventAndGetWindowCounts(
		ctx,
		provider,
		s.currentTime(),
		false,
		s.gatewayHealthWindow(),
	)
	if err != nil {
		return fmt.Errorf("%w: record successful gateway response: %v", ErrPaymentGatewayCircuitBreakerStoreUnavailable, err)
	}
	return nil
}

func (s *PaymentGatewayCircuitBreakerService) RecordFailedGatewayAPIResponse(
	ctx context.Context,
	provider string,
) (PaymentGatewayCircuitBreakerDecision, error) {
	decision := PaymentGatewayCircuitBreakerDecision{Allowed: true}
	if s == nil || !s.config.Enabled || s.store == nil {
		return decision, nil
	}
	provider = normalizePaymentGatewayHealthProvider(provider)
	if provider == "" {
		return decision, nil
	}

	sampleCount, failureCount, err := s.store.RecordGatewayHealthEventAndGetWindowCounts(
		ctx,
		provider,
		s.currentTime(),
		true,
		s.gatewayHealthWindow(),
	)
	if err != nil {
		return decision, fmt.Errorf("%w: record failed gateway response: %v", ErrPaymentGatewayCircuitBreakerStoreUnavailable, err)
	}

	decision = paymentGatewayCircuitBreakerDecisionFromWindowCounts(
		sampleCount,
		failureCount,
		s.config.FailureRateThreshold,
		s.config.MinimumSampleCount,
	)
	if !decision.CircuitOpen {
		return decision, nil
	}

	openDuration := s.gatewayCircuitOpenDuration()
	if err := s.store.OpenGatewayCircuit(ctx, provider, openDuration); err != nil {
		return decision, fmt.Errorf("%w: open gateway circuit: %v", ErrPaymentGatewayCircuitBreakerStoreUnavailable, err)
	}
	decision.Allowed = false
	decision.RetryAfter = openDuration
	return decision, nil
}

func (s *PaymentGatewayCircuitBreakerService) ReadGatewayHealthWindow(
	ctx context.Context,
	provider string,
) (PaymentGatewayCircuitBreakerDecision, error) {
	decision := PaymentGatewayCircuitBreakerDecision{Allowed: true}
	if s == nil || !s.config.Enabled || s.store == nil {
		return decision, nil
	}
	provider = normalizePaymentGatewayHealthProvider(provider)
	if provider == "" {
		return decision, nil
	}

	sampleCount, failureCount, err := s.store.GetGatewayHealthWindowCounts(
		ctx,
		provider,
		s.currentTime(),
		s.gatewayHealthWindow(),
	)
	if err != nil {
		return decision, fmt.Errorf("%w: read gateway health window: %v", ErrPaymentGatewayCircuitBreakerStoreUnavailable, err)
	}
	return paymentGatewayCircuitBreakerDecisionFromWindowCounts(
		sampleCount,
		failureCount,
		s.config.FailureRateThreshold,
		s.config.MinimumSampleCount,
	), nil
}

func (s *PaymentGatewayCircuitBreakerService) currentTime() time.Time {
	if s != nil && s.now != nil {
		return s.now().UTC()
	}
	return time.Now().UTC()
}

func (s *PaymentGatewayCircuitBreakerService) gatewayHealthWindow() time.Duration {
	return time.Duration(s.config.WindowSeconds) * time.Second
}

func (s *PaymentGatewayCircuitBreakerService) gatewayCircuitOpenDuration() time.Duration {
	return time.Duration(s.config.OpenDurationSeconds) * time.Second
}

func paymentGatewayCircuitBreakerDecisionFromWindowCounts(
	sampleCount int64,
	failureCount int64,
	failureRateThreshold float64,
	minimumSampleCount int,
) PaymentGatewayCircuitBreakerDecision {
	failureRate := 0.0
	if sampleCount > 0 {
		failureRate = float64(failureCount) / float64(sampleCount)
	}
	circuitOpen := sampleCount >= int64(minimumSampleCount) && failureRate > failureRateThreshold
	return PaymentGatewayCircuitBreakerDecision{
		Allowed:      !circuitOpen,
		CircuitOpen:  circuitOpen,
		FailureRate:  failureRate,
		SampleCount:  sampleCount,
		FailureCount: failureCount,
	}
}

func normalizePaymentGatewayCircuitBreakerConfig(
	cfg config.PaymentGatewayCircuitBreakerConfig,
) config.PaymentGatewayCircuitBreakerConfig {
	if cfg.WindowSeconds <= 0 {
		cfg.WindowSeconds = 60
	}
	if cfg.FailureRateThreshold <= 0 {
		cfg.FailureRateThreshold = 0.15
	}
	if cfg.MinimumSampleCount <= 0 {
		cfg.MinimumSampleCount = 20
	}
	if cfg.OpenDurationSeconds <= 0 {
		cfg.OpenDurationSeconds = 30
	}
	return cfg
}

func normalizePaymentGatewayHealthProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case string(pgateway.GatewayStripe):
		return string(pgateway.GatewayStripe)
	case string(pgateway.GatewayPayPal):
		return string(pgateway.GatewayPayPal)
	case string(pgateway.GatewayAlipay):
		return string(pgateway.GatewayAlipay)
	case string(pgateway.GatewayWechat):
		return string(pgateway.GatewayWechat)
	default:
		return ""
	}
}

func paymentGatewayHealthEventsKey(provider string) string {
	return "tanzanite:payment-gateway-health:" + provider + ":events"
}

func paymentGatewayHealthFailuresKey(provider string) string {
	return "tanzanite:payment-gateway-health:" + provider + ":failures"
}

func paymentGatewayCircuitKey(provider string) string {
	return "tanzanite:payment-gateway-health:" + provider + ":circuit-open"
}

func boolToRedisFlag(value bool) string {
	if value {
		return "1"
	}
	return "0"
}
