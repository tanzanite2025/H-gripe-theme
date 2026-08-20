package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/pkg/config"
	pgateway "commerce-platform/internal/pkg/payment"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ErrPaymentGatewayCircuitBreakerStoreUnavailable = errors.New("payment gateway circuit breaker store is unavailable")

type PaymentGatewayCircuitBreakerDecision struct {
	Allowed      bool
	CircuitOpen  bool
	Probe        bool
	PermitToken  string `json:"-"`
	FailureRate  float64
	SampleCount  int64
	FailureCount int64
	RetryAfter   time.Duration
}

type PaymentGatewayCircuitBreakerPolicyView struct {
	Enabled                 bool    `json:"enabled"`
	WindowSeconds           int     `json:"window_seconds"`
	FailureRateThreshold    float64 `json:"failure_rate_threshold"`
	MinimumSampleCount      int     `json:"minimum_sample_count"`
	OpenDurationSeconds     int     `json:"open_duration_seconds"`
	HalfOpenProbeTimeoutSec int     `json:"half_open_probe_timeout_seconds"`
}

type paymentGatewayCircuitBreakerStore interface {
	AcquireGatewayPaymentAttempt(context.Context, string, time.Duration) (bool, bool, time.Duration, string, error)
	GetOpenCircuitRetryAfter(context.Context, string) (bool, time.Duration, error)
	GetGatewayHealthWindowCounts(context.Context, string, time.Time, time.Duration) (int64, int64, error)
	RecordGatewayHealthEventAndGetWindowCounts(context.Context, string, time.Time, bool, time.Duration) (int64, int64, error)
	OpenGatewayCircuit(context.Context, string, time.Duration, time.Duration, string) (bool, error)
	CloseGatewayCircuitProbe(context.Context, string, string) (bool, error)
	ReleaseGatewayCircuitProbe(context.Context, string, string) error
}

type redisPaymentGatewayCircuitBreakerStore struct {
	client redis.UniversalClient
}

const readGatewayHealthWindowCountsScript = `
local redisTime = redis.call("TIME")
local now = (tonumber(redisTime[1]) * 1000) + math.floor(tonumber(redisTime[2]) / 1000)
local window = tonumber(ARGV[1])
local cutoff = now - window
redis.call("ZREMRANGEBYSCORE", KEYS[1], 0, cutoff)
redis.call("ZREMRANGEBYSCORE", KEYS[2], 0, cutoff)
return {redis.call("ZCARD", KEYS[1]), redis.call("ZCARD", KEYS[2])}
`

const recordGatewayHealthEventScript = `
local redisTime = redis.call("TIME")
local now = (tonumber(redisTime[1]) * 1000) + math.floor(tonumber(redisTime[2]) / 1000)
local window = tonumber(ARGV[1])
local cutoff = now - window
local member = ARGV[2]
local retentionSeconds = tonumber(ARGV[3])
local failure = ARGV[4]

redis.call("ZREMRANGEBYSCORE", KEYS[1], 0, cutoff)
redis.call("ZREMRANGEBYSCORE", KEYS[2], 0, cutoff)
redis.call("ZADD", KEYS[1], now, member)
if failure == "1" then
	redis.call("ZADD", KEYS[2], now, member)
end
redis.call("EXPIRE", KEYS[1], retentionSeconds)
redis.call("EXPIRE", KEYS[2], retentionSeconds)

return {redis.call("ZCARD", KEYS[1]), redis.call("ZCARD", KEYS[2])}
`

const acquireGatewayPaymentAttemptScript = `
local redisTime = redis.call("TIME")
local now = (tonumber(redisTime[1]) * 1000) + math.floor(tonumber(redisTime[2]) / 1000)
local state = redis.call("GET", KEYS[1])

if not state then
	return {1, 0, 0}
end

local decodedOK, decoded = pcall(cjson.decode, state)
if not decodedOK or type(decoded) ~= "table" then
	local legacyTTL = redis.call("PTTL", KEYS[1])
	if legacyTTL > 0 then
		return {0, legacyTTL, 0}
	end
	redis.call("DEL", KEYS[1])
	return {1, 0, 0}
end

local openUntil = tonumber(decoded["open_until_ms"])
if not openUntil then
	redis.call("DEL", KEYS[1])
	return {1, 0, 0}
end

if openUntil > now then
	return {0, openUntil - now, 0}
end

local probeTTL = tonumber(ARGV[2])
if redis.call("SET", KEYS[2], ARGV[1], "PX", probeTTL, "NX") then
	return {1, 0, 1}
end

local retryAfter = redis.call("PTTL", KEYS[2])
if retryAfter <= 0 then
	retryAfter = probeTTL
end
return {0, retryAfter, 1}
`

const openGatewayCircuitScript = `
local redisTime = redis.call("TIME")
local now = (tonumber(redisTime[1]) * 1000) + math.floor(tonumber(redisTime[2]) / 1000)
local openDuration = tonumber(ARGV[1])
local stateTTL = tonumber(ARGV[2])
local permitToken = ARGV[3]
local generation = ARGV[4]

if permitToken ~= "" then
	if redis.call("GET", KEYS[2]) ~= permitToken then
		return 0
	end
	redis.call("DEL", KEYS[2])
else
	if redis.call("EXISTS", KEYS[2]) == 1 then
		return 0
	end
end

local state = cjson.encode({
	open_until_ms = now + openDuration,
	generation = generation
})
redis.call("SET", KEYS[1], state, "PX", stateTTL)
return 1
`

const closeGatewayCircuitProbeScript = `
if redis.call("GET", KEYS[2]) ~= ARGV[1] then
	return 0
end
redis.call("DEL", KEYS[1])
redis.call("DEL", KEYS[2])
return 1
`

const releaseGatewayCircuitProbeScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`

func (s redisPaymentGatewayCircuitBreakerStore) AcquireGatewayPaymentAttempt(
	ctx context.Context,
	provider string,
	probeLease time.Duration,
) (bool, bool, time.Duration, string, error) {
	if probeLease <= 0 {
		probeLease = time.Second
	}
	token := uuid.NewString()
	values, err := s.client.Eval(
		ctx,
		acquireGatewayPaymentAttemptScript,
		[]string{paymentGatewayCircuitKey(provider), paymentGatewayCircuitProbeKey(provider)},
		token,
		probeLease.Milliseconds(),
	).Int64Slice()
	if err != nil {
		return false, false, 0, "", err
	}
	if len(values) != 3 {
		return false, false, 0, "", errors.New("invalid payment gateway circuit acquire response")
	}
	allowed := values[0] == 1
	probe := values[2] == 1
	permitToken := ""
	if allowed && probe {
		permitToken = token
	}
	return allowed, probe, time.Duration(values[1]) * time.Millisecond, permitToken, nil
}

func (s redisPaymentGatewayCircuitBreakerStore) GetOpenCircuitRetryAfter(
	ctx context.Context,
	provider string,
) (bool, time.Duration, error) {
	probeTTL, err := s.client.PTTL(ctx, paymentGatewayCircuitProbeKey(provider)).Result()
	if err != nil {
		return false, 0, err
	}
	value, err := s.client.Get(ctx, paymentGatewayCircuitKey(provider)).Result()
	if err != nil {
		if err == redis.Nil {
			if probeTTL > 0 {
				return true, time.Duration(probeTTL) * time.Millisecond, nil
			}
			return false, 0, nil
		}
		return false, 0, err
	}
	ttl, err := s.client.TTL(ctx, paymentGatewayCircuitKey(provider)).Result()
	if err != nil {
		return false, 0, err
	}
	if ttl <= 0 {
		if probeTTL > 0 {
			return true, time.Duration(probeTTL) * time.Millisecond, nil
		}
		return false, 0, nil
	}
	var state struct {
		OpenUntilMs int64 `json:"open_until_ms"`
	}
	if err := json.Unmarshal([]byte(value), &state); err == nil && state.OpenUntilMs > 0 {
		now, timeErr := s.client.Time(ctx).Result()
		if timeErr == nil {
			retryAfterMs := state.OpenUntilMs - now.UnixMilli()
			if retryAfterMs <= 0 {
				if probeTTL > 0 {
					return true, time.Duration(probeTTL) * time.Millisecond, nil
				}
				return false, 0, nil
			}
			return true, time.Duration(retryAfterMs) * time.Millisecond, nil
		}
	}
	if probeTTL > 0 {
		return true, time.Duration(probeTTL) * time.Millisecond, nil
	}
	if value == "open" {
		return true, ttl, nil
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
		window,
		"",
		false,
	)
}

func (s redisPaymentGatewayCircuitBreakerStore) RecordGatewayHealthEventAndGetWindowCounts(
	ctx context.Context,
	provider string,
	_ time.Time,
	failure bool,
	window time.Duration,
) (int64, int64, error) {
	member := uuid.NewString()
	return s.executeGatewayHealthWindowScript(
		ctx,
		recordGatewayHealthEventScript,
		provider,
		window,
		member,
		failure,
	)
}

func (s redisPaymentGatewayCircuitBreakerStore) executeGatewayHealthWindowScript(
	ctx context.Context,
	script string,
	provider string,
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
		window.Milliseconds(),
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
	probeLease time.Duration,
	permitToken string,
) (bool, error) {
	if duration <= 0 {
		duration = time.Second
	}
	if probeLease <= 0 {
		probeLease = time.Second
	}
	opened, err := s.client.Eval(
		ctx,
		openGatewayCircuitScript,
		[]string{paymentGatewayCircuitKey(provider), paymentGatewayCircuitProbeKey(provider)},
		duration.Milliseconds(),
		(duration + probeLease).Milliseconds(),
		strings.TrimSpace(permitToken),
		uuid.NewString(),
	).Int64()
	return opened == 1, err
}

func (s redisPaymentGatewayCircuitBreakerStore) CloseGatewayCircuitProbe(
	ctx context.Context,
	provider string,
	permitToken string,
) (bool, error) {
	permitToken = strings.TrimSpace(permitToken)
	if permitToken == "" {
		return false, nil
	}
	closed, err := s.client.Eval(
		ctx,
		closeGatewayCircuitProbeScript,
		[]string{paymentGatewayCircuitKey(provider), paymentGatewayCircuitProbeKey(provider)},
		permitToken,
	).Int64()
	return closed == 1, err
}

func (s redisPaymentGatewayCircuitBreakerStore) ReleaseGatewayCircuitProbe(
	ctx context.Context,
	provider string,
	permitToken string,
) error {
	permitToken = strings.TrimSpace(permitToken)
	if permitToken == "" {
		return nil
	}
	return s.client.Eval(
		ctx,
		releaseGatewayCircuitProbeScript,
		[]string{paymentGatewayCircuitProbeKey(provider)},
		permitToken,
	).Err()
}

type PaymentGatewayCircuitBreakerService struct {
	store  paymentGatewayCircuitBreakerStore
	config config.PaymentGatewayCircuitBreakerConfig
	now    func() time.Time
}

func NewPaymentGatewayCircuitBreakerService(
	redisClient redis.UniversalClient,
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

	allowed, probe, retryAfter, permitToken, err := s.store.AcquireGatewayPaymentAttempt(
		ctx,
		provider,
		s.gatewayCircuitProbeTimeout(),
	)
	if err != nil {
		return decision, fmt.Errorf("%w: acquire circuit permit: %v", ErrPaymentGatewayCircuitBreakerStoreUnavailable, err)
	}
	if allowed {
		decision.Probe = probe
		decision.PermitToken = permitToken
		return decision, nil
	}

	decision = PaymentGatewayCircuitBreakerDecision{
		Allowed:     false,
		CircuitOpen: true,
		Probe:       probe,
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

func (s *PaymentGatewayCircuitBreakerService) ReadGatewayPaymentAttemptAvailability(
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
	permitTokens ...string,
) error {
	if s == nil || !s.config.Enabled || s.store == nil {
		return nil
	}
	provider = normalizePaymentGatewayHealthProvider(provider)
	if provider == "" {
		return nil
	}
	permitToken := firstPaymentGatewayPermitToken(permitTokens)
	if permitToken != "" {
		closed, err := s.store.CloseGatewayCircuitProbe(ctx, provider, permitToken)
		if err != nil {
			return fmt.Errorf("%w: close successful gateway probe: %v", ErrPaymentGatewayCircuitBreakerStoreUnavailable, err)
		}
		if !closed {
			return nil
		}
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

func (s *PaymentGatewayCircuitBreakerService) ReleaseGatewayPaymentAttempt(
	ctx context.Context,
	provider string,
	permitTokens ...string,
) error {
	if s == nil || !s.config.Enabled || s.store == nil {
		return nil
	}
	provider = normalizePaymentGatewayHealthProvider(provider)
	if provider == "" {
		return nil
	}
	permitToken := firstPaymentGatewayPermitToken(permitTokens)
	if permitToken == "" {
		return nil
	}
	if err := s.store.ReleaseGatewayCircuitProbe(ctx, provider, permitToken); err != nil {
		return fmt.Errorf("%w: release gateway probe: %v", ErrPaymentGatewayCircuitBreakerStoreUnavailable, err)
	}
	return nil
}

func (s *PaymentGatewayCircuitBreakerService) RecordFailedGatewayAPIResponse(
	ctx context.Context,
	provider string,
	permitTokens ...string,
) (PaymentGatewayCircuitBreakerDecision, error) {
	decision := PaymentGatewayCircuitBreakerDecision{Allowed: true}
	if s == nil || !s.config.Enabled || s.store == nil {
		return decision, nil
	}
	provider = normalizePaymentGatewayHealthProvider(provider)
	if provider == "" {
		return decision, nil
	}
	permitToken := firstPaymentGatewayPermitToken(permitTokens)

	openDuration := s.gatewayCircuitOpenDuration()
	if permitToken != "" {
		opened, err := s.store.OpenGatewayCircuit(
			ctx,
			provider,
			openDuration,
			s.gatewayCircuitProbeTimeout(),
			permitToken,
		)
		if err != nil {
			return decision, fmt.Errorf("%w: open gateway circuit: %v", ErrPaymentGatewayCircuitBreakerStoreUnavailable, err)
		}
		if !opened {
			return decision, nil
		}
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
	if permitToken != "" {
		decision.Allowed = false
		decision.CircuitOpen = true
		decision.RetryAfter = openDuration
		return decision, nil
	}
	if !decision.CircuitOpen {
		return decision, nil
	}

	opened, err := s.store.OpenGatewayCircuit(
		ctx,
		provider,
		openDuration,
		s.gatewayCircuitProbeTimeout(),
		permitToken,
	)
	if err != nil {
		return decision, fmt.Errorf("%w: open gateway circuit: %v", ErrPaymentGatewayCircuitBreakerStoreUnavailable, err)
	}
	if opened {
		decision.Allowed = false
		decision.CircuitOpen = true
		decision.RetryAfter = openDuration
	}
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

func (s *PaymentGatewayCircuitBreakerService) PolicyView() PaymentGatewayCircuitBreakerPolicyView {
	if s == nil {
		return PaymentGatewayCircuitBreakerPolicyView{}
	}
	return PaymentGatewayCircuitBreakerPolicyView{
		Enabled:                 s.config.Enabled,
		WindowSeconds:           s.config.WindowSeconds,
		FailureRateThreshold:    s.config.FailureRateThreshold,
		MinimumSampleCount:      s.config.MinimumSampleCount,
		OpenDurationSeconds:     s.config.OpenDurationSeconds,
		HalfOpenProbeTimeoutSec: s.config.HalfOpenProbeTimeoutSec,
	}
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

func (s *PaymentGatewayCircuitBreakerService) gatewayCircuitProbeTimeout() time.Duration {
	return time.Duration(s.config.HalfOpenProbeTimeoutSec) * time.Second
}

func firstPaymentGatewayPermitToken(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
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
	if cfg.HalfOpenProbeTimeoutSec <= 0 {
		cfg.HalfOpenProbeTimeoutSec = config.MinimumPaymentGatewayHalfOpenProbeTimeoutSeconds
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
	return paymentGatewayHealthKeyPrefix(provider) + ":events"
}

func paymentGatewayHealthFailuresKey(provider string) string {
	return paymentGatewayHealthKeyPrefix(provider) + ":failures"
}

func paymentGatewayCircuitKey(provider string) string {
	return paymentGatewayHealthKeyPrefix(provider) + ":circuit-open"
}

func paymentGatewayCircuitProbeKey(provider string) string {
	return paymentGatewayHealthKeyPrefix(provider) + ":circuit-probe"
}

func paymentGatewayHealthKeyPrefix(provider string) string {
	return "commerce-platform:payment-gateway-health:{" + strings.TrimSpace(provider) + "}"
}

func boolToRedisFlag(value bool) string {
	if value {
		return "1"
	}
	return "0"
}
