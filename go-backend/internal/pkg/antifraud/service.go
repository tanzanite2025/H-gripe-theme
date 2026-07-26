package antifraud

import (
	"context"
	"errors"
	"strings"
	"time"

	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/metrics"

	"github.com/redis/go-redis/v9"
)

var ErrServiceUnavailable = errors.New("payment risk service is unavailable")

type Signals struct {
	IPCountry      string
	BillingCountry string
	VPNDetected    bool
	Timezone       string
	UserAgent      string
}

type Decision struct {
	Score    int
	HighRisk bool
	Failures int
	Delay    time.Duration
	Reasons  []string
}

type Service struct {
	redis  *redis.Client
	config config.PaymentRiskConfig
}

func New(redisClient *redis.Client, cfg config.PaymentRiskConfig) *Service {
	if redisClient == nil {
		return nil
	}
	return &Service{redis: redisClient, config: cfg}
}

func (s *Service) Evaluate(ctx context.Context, key string, signals Signals) (Decision, error) {
	decision := Decision{
		Score:   0,
		Reasons: []string{},
	}
	if s == nil || s.redis == nil {
		return decision, nil
	}

	failures, err := s.failureCount(ctx, key)
	if err != nil {
		return decision, ErrServiceUnavailable
	}
	return s.evaluateSignals(failures, signals), nil
}

func (s *Service) evaluateSignals(failures int, signals Signals) Decision {
	decision := Decision{
		Score:    0,
		Failures: failures,
		Reasons:  []string{},
	}
	if failures >= s.config.FailureThreshold {
		decision.Score += 40
		decision.Reasons = append(decision.Reasons, "repeated payment failures")
	}
	if normalizeCountry(signals.IPCountry) != "" &&
		normalizeCountry(signals.BillingCountry) != "" &&
		normalizeCountry(signals.IPCountry) != normalizeCountry(signals.BillingCountry) {
		decision.Score += 35
		decision.Reasons = append(decision.Reasons, "country mismatch")
	}
	if signals.VPNDetected {
		decision.Score += 25
		decision.Reasons = append(decision.Reasons, "vpn signal")
	}
	if strings.TrimSpace(signals.UserAgent) == "" {
		decision.Score += 10
		decision.Reasons = append(decision.Reasons, "missing user agent")
	}

	decision.HighRisk = decision.Score >= s.config.HighRiskScore
	if failures >= s.config.FailureThreshold || decision.HighRisk {
		decision.Delay = time.Duration(s.config.DelaySeconds) * time.Second
	}
	return decision
}

func (s *Service) RecordFailure(ctx context.Context, key string) error {
	if s == nil || s.redis == nil {
		return nil
	}
	for _, currentKey := range failureKeys(key) {
		count, err := s.redis.Incr(ctx, failureKey(currentKey)).Result()
		if err != nil {
			return ErrServiceUnavailable
		}
		if count == 1 {
			if err := s.redis.Expire(ctx, failureKey(currentKey), time.Duration(s.config.FailureWindowSeconds)*time.Second).Err(); err != nil {
				return ErrServiceUnavailable
			}
		}
	}
	metrics.PaymentAttempts.WithLabelValues("checkout", "failed").Inc()
	return nil
}

func (s *Service) RecordSuccess(ctx context.Context, key string) error {
	if s == nil || s.redis == nil {
		return nil
	}
	keys := make([]string, 0, 2)
	for _, currentKey := range failureKeys(key) {
		keys = append(keys, failureKey(currentKey))
	}
	if err := s.redis.Del(ctx, keys...).Err(); err != nil {
		return ErrServiceUnavailable
	}
	metrics.PaymentAttempts.WithLabelValues("checkout", "success").Inc()
	return nil
}

func (s *Service) RecordProviderFailure(provider string) {
	metrics.PaymentAttempts.WithLabelValues(strings.ToLower(strings.TrimSpace(provider)), "failed").Inc()
}

func (s *Service) RecordProviderSuccess(provider string) {
	metrics.PaymentAttempts.WithLabelValues(strings.ToLower(strings.TrimSpace(provider)), "success").Inc()
}

func (s *Service) failureCount(ctx context.Context, key string) (int, error) {
	maxFailures := 0
	for _, currentKey := range failureKeys(key) {
		value, err := s.redis.Get(ctx, failureKey(currentKey)).Int()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return 0, err
		}
		if value > maxFailures {
			maxFailures = value
		}
	}
	return maxFailures, nil
}

func failureKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		key = "anonymous"
	}
	return "tanzanite:payment-risk:failures:" + key
}

func failureKeys(key string) []string {
	key = strings.TrimSpace(key)
	if key == "" {
		return []string{"anonymous"}
	}

	keys := []string{key}
	if sessionMarker := strings.Index(key, ":session:"); sessionMarker > 0 {
		keys = append(keys, key[:sessionMarker])
	}
	return keys
}

func normalizeCountry(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}
