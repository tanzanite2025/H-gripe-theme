package cardtesting

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"commerce-platform/internal/pkg/config"
	paymentpkg "commerce-platform/internal/pkg/payment"

	"github.com/redis/go-redis/v9"
)

var (
	ErrServiceUnavailable = errors.New("card BIN limiter is unavailable")
	ErrBINBlocked         = errors.New("card BIN is temporarily blocked")
)

const paymentIntentBindingTTL = 24 * time.Hour

const (
	defaultWindowSeconds        = 60
	defaultFailureThreshold     = 5
	defaultBlockDurationSeconds = 1800
)

const recordFailureScript = `
local blockTTL = redis.call("TTL", KEYS[2])
if blockTTL == -1 or blockTTL >= 0 then
	return {0, 1, blockTTL}
end

local count = redis.call("INCR", KEYS[1])
if count == 1 then
	redis.call("EXPIRE", KEYS[1], ARGV[1])
end

if count >= tonumber(ARGV[3]) then
	redis.call("SET", KEYS[2], "1", "EX", ARGV[2])
	return {count, 1, tonumber(ARGV[2])}
end

return {count, 0, redis.call("TTL", KEYS[1])}
`

type Decision struct {
	Allowed    bool
	Blocked    bool
	Count      int64
	RetryAfter time.Duration
}

type store interface {
	CheckBlocked(context.Context, string) (bool, time.Duration, error)
	RecordFailure(context.Context, string, string, time.Duration, time.Duration, int64) (int64, bool, time.Duration, error)
	Delete(context.Context, string) error
	BindPaymentIntent(context.Context, string, string, time.Duration) error
	PaymentIntentBinding(context.Context, string) (string, error)
	DeletePaymentIntentBinding(context.Context, string) error
}

type redisStore struct {
	client redis.UniversalClient
}

func (s redisStore) CheckBlocked(ctx context.Context, key string) (bool, time.Duration, error) {
	ttl, err := s.client.TTL(ctx, key).Result()
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

func (s redisStore) RecordFailure(
	ctx context.Context,
	failureKey string,
	blockKey string,
	window time.Duration,
	blockDuration time.Duration,
	threshold int64,
) (int64, bool, time.Duration, error) {
	values, err := s.client.Eval(
		ctx,
		recordFailureScript,
		[]string{failureKey, blockKey},
		int64(window/time.Second),
		int64(blockDuration/time.Second),
		threshold,
	).Int64Slice()
	if err != nil {
		return 0, false, 0, err
	}
	if len(values) != 3 {
		return 0, false, 0, errors.New("invalid card BIN limiter response")
	}
	return values[0], values[1] == 1, time.Duration(values[2]) * time.Second, nil
}

func (s redisStore) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

func (s redisStore) BindPaymentIntent(ctx context.Context, paymentIntentID, failureKey string, ttl time.Duration) error {
	return s.client.Set(ctx, paymentIntentBindingKey(paymentIntentID), failureKey, ttl).Err()
}

func (s redisStore) PaymentIntentBinding(ctx context.Context, paymentIntentID string) (string, error) {
	value, err := s.client.Get(ctx, paymentIntentBindingKey(paymentIntentID)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return value, err
}

func (s redisStore) DeletePaymentIntentBinding(ctx context.Context, paymentIntentID string) error {
	return s.client.Del(ctx, paymentIntentBindingKey(paymentIntentID)).Err()
}

type Service struct {
	store store
	cfg   config.PaymentBINRateLimitConfig
}

func New(redisClient redis.UniversalClient, cfg config.PaymentBINRateLimitConfig) *Service {
	if redisClient == nil {
		return nil
	}
	return &Service{
		store: redisStore{client: redisClient},
		cfg:   normalizeConfig(cfg),
	}
}

func newWithStore(s store, cfg config.PaymentBINRateLimitConfig) *Service {
	return &Service{store: s, cfg: normalizeConfig(cfg)}
}

func (s *Service) Check(ctx context.Context, bin string) (Decision, error) {
	decision := Decision{Allowed: true}
	if s == nil || !s.cfg.Enabled || s.store == nil || strings.TrimSpace(bin) == "" {
		return decision, nil
	}

	normalized, err := paymentpkg.NormalizeCardBIN(bin)
	if err != nil {
		return decision, err
	}
	blocked, retryAfter, err := s.store.CheckBlocked(ctx, blockKey(normalized))
	if err != nil {
		return Decision{}, ErrServiceUnavailable
	}
	if blocked {
		return Decision{
			Allowed:    false,
			Blocked:    true,
			RetryAfter: retryAfter,
		}, nil
	}
	return decision, nil
}

func (s *Service) RecordFailure(ctx context.Context, bin string) (Decision, error) {
	decision := Decision{Allowed: true}
	if s == nil || !s.cfg.Enabled || s.store == nil || strings.TrimSpace(bin) == "" {
		return decision, nil
	}

	normalized, err := paymentpkg.NormalizeCardBIN(bin)
	if err != nil {
		return decision, err
	}
	count, blocked, retryAfter, err := s.store.RecordFailure(
		ctx,
		failureKey(normalized),
		blockKey(normalized),
		time.Duration(s.cfg.WindowSeconds)*time.Second,
		time.Duration(s.cfg.BlockDurationSeconds)*time.Second,
		int64(s.cfg.FailureThreshold),
	)
	if err != nil {
		return Decision{}, ErrServiceUnavailable
	}
	return Decision{
		Allowed:    !blocked,
		Blocked:    blocked,
		Count:      count,
		RetryAfter: retryAfter,
	}, nil
}

func (s *Service) RecordSuccess(ctx context.Context, bin string) error {
	if s == nil || !s.cfg.Enabled || s.store == nil || strings.TrimSpace(bin) == "" {
		return nil
	}
	normalized, err := paymentpkg.NormalizeCardBIN(bin)
	if err != nil {
		return err
	}
	if err := s.store.Delete(ctx, failureKey(normalized)); err != nil {
		return ErrServiceUnavailable
	}
	return nil
}

func (s *Service) BindPaymentIntent(ctx context.Context, paymentIntentID, bin string) error {
	if s == nil || !s.cfg.Enabled || s.store == nil {
		return nil
	}
	paymentIntentID = strings.TrimSpace(paymentIntentID)
	if paymentIntentID == "" || strings.TrimSpace(bin) == "" {
		return nil
	}
	normalized, err := paymentpkg.NormalizeCardBIN(bin)
	if err != nil {
		return err
	}
	if err := s.store.BindPaymentIntent(ctx, paymentIntentID, failureKey(normalized), paymentIntentBindingTTL); err != nil {
		return ErrServiceUnavailable
	}
	return nil
}

func (s *Service) RecordPaymentIntentFailure(ctx context.Context, paymentIntentID string) (Decision, error) {
	if s == nil || !s.cfg.Enabled || s.store == nil {
		return Decision{Allowed: true}, nil
	}
	paymentIntentID = strings.TrimSpace(paymentIntentID)
	if paymentIntentID == "" {
		return Decision{Allowed: true}, nil
	}
	failureKeyValue, err := s.store.PaymentIntentBinding(ctx, paymentIntentID)
	if err != nil {
		return Decision{}, ErrServiceUnavailable
	}
	if failureKeyValue == "" {
		return Decision{Allowed: true}, nil
	}
	count, blocked, retryAfter, err := s.store.RecordFailure(
		ctx,
		failureKeyValue,
		blockKeyFromFailureKey(failureKeyValue),
		time.Duration(s.cfg.WindowSeconds)*time.Second,
		time.Duration(s.cfg.BlockDurationSeconds)*time.Second,
		int64(s.cfg.FailureThreshold),
	)
	if err != nil {
		return Decision{}, ErrServiceUnavailable
	}
	return Decision{
		Allowed:    !blocked,
		Blocked:    blocked,
		Count:      count,
		RetryAfter: retryAfter,
	}, nil
}

func (s *Service) RecordPaymentIntentSuccess(ctx context.Context, paymentIntentID string) error {
	if s == nil || !s.cfg.Enabled || s.store == nil {
		return nil
	}
	paymentIntentID = strings.TrimSpace(paymentIntentID)
	if paymentIntentID == "" {
		return nil
	}
	failureKeyValue, err := s.store.PaymentIntentBinding(ctx, paymentIntentID)
	if err != nil {
		return ErrServiceUnavailable
	}
	if failureKeyValue != "" {
		if err := s.store.Delete(ctx, failureKeyValue); err != nil {
			return ErrServiceUnavailable
		}
	}
	if err := s.store.DeletePaymentIntentBinding(ctx, paymentIntentID); err != nil {
		return ErrServiceUnavailable
	}
	return nil
}

func failureKey(bin string) string {
	return "commerce_platform:payment-risk:bin:" + binLength(bin) + ":" + digest(bin) + ":failures"
}

func blockKey(bin string) string {
	return "commerce_platform:payment-risk:bin:" + binLength(bin) + ":" + digest(bin) + ":blocked"
}

func blockKeyFromFailureKey(key string) string {
	return strings.TrimSuffix(key, ":failures") + ":blocked"
}

func paymentIntentBindingKey(paymentIntentID string) string {
	return "commerce_platform:payment-risk:bin-payment-intent:" + digest(paymentIntentID)
}

func digest(value string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:])
}

func normalizeConfig(cfg config.PaymentBINRateLimitConfig) config.PaymentBINRateLimitConfig {
	if cfg.WindowSeconds <= 0 {
		cfg.WindowSeconds = defaultWindowSeconds
	}
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = defaultFailureThreshold
	}
	if cfg.BlockDurationSeconds <= 0 {
		cfg.BlockDurationSeconds = defaultBlockDurationSeconds
	}
	return cfg
}

func binLength(value string) string {
	if len(value) == 6 {
		return "6"
	}
	return "8"
}
