package antifraud

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
	Score             int
	HighRisk          bool
	Failures          int
	Delay             time.Duration
	Action            string
	ChallengeRequired bool
	ChallengeReason   string
	Reasons           []string
}

const (
	DecisionActionAllow     = "allow"
	DecisionActionDelay     = "delay"
	DecisionActionChallenge = "challenge"
	DecisionActionBlock     = "block"
)

type AttemptIdentity struct {
	Provider    string
	UserID      string
	SessionID   string
	AnonymousID string
	IPAddress   string
	UserAgent   string
}

type AttemptInput struct {
	Identity AttemptIdentity
	Signals  Signals
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
		Action:  DecisionActionAllow,
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

func (s *Service) EvaluateAttempt(ctx context.Context, input AttemptInput) (Decision, error) {
	decision := Decision{
		Score:   0,
		Action:  DecisionActionAllow,
		Reasons: []string{},
	}
	if s == nil || s.redis == nil {
		return decision, nil
	}
	keys := input.Identity.FailureKeys()
	if len(keys) == 0 {
		return decision, nil
	}
	failures, err := s.failureCountForKeys(ctx, keys)
	if err != nil {
		return decision, ErrServiceUnavailable
	}
	return s.evaluateSignals(failures, input.Signals), nil

}

func (s *Service) evaluateSignals(failures int, signals Signals) Decision {
	decision := Decision{
		Score:    0,
		Failures: failures,
		Action:   DecisionActionAllow,
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
	if failures >= s.config.FailureThreshold {
		decision.Action = DecisionActionChallenge
		decision.ChallengeRequired = true
		decision.ChallengeReason = "repeated_payment_failures"
	} else if decision.HighRisk {
		decision.Action = DecisionActionDelay
	}
	if decision.Action == DecisionActionChallenge || decision.Action == DecisionActionDelay {
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

func (s *Service) RecordAttemptFailure(ctx context.Context, identity AttemptIdentity) error {
	if s == nil || s.redis == nil {
		return nil
	}
	return s.recordFailureForKeys(ctx, identity.FailureKeys())

}

func (s *Service) RecordAttemptSuccess(ctx context.Context, identity AttemptIdentity) error {
	if s == nil || s.redis == nil {
		return nil
	}
	return s.recordSuccessForKeys(ctx, identity.FailureKeys())

}

func (s *Service) RecordSuccess(ctx context.Context, key string) error {
	if s == nil || s.redis == nil {
		return nil
	}
	return s.recordSuccessForKeys(ctx, failureKeys(key))
}

func (s *Service) BindPaymentIntent(ctx context.Context, paymentIntentID string, identity AttemptIdentity) error {
	if s == nil || s.redis == nil {
		return nil
	}
	paymentIntentID = strings.TrimSpace(paymentIntentID)
	if paymentIntentID == "" {
		return nil
	}
	keys := identity.FailureKeys()
	if len(keys) == 0 {
		return nil
	}
	payload, err := json.Marshal(paymentIntentBinding{Keys: keys})
	if err != nil {
		return fmt.Errorf("%w: encode payment intent risk binding", ErrServiceUnavailable)
	}
	if err := s.redis.Set(ctx, paymentIntentBindingKey(paymentIntentID), payload, 24*time.Hour).Err(); err != nil {
		return ErrServiceUnavailable
	}
	return nil

}

func (s *Service) RecordPaymentIntentFailure(ctx context.Context, paymentIntentID string) error {
	if s == nil || s.redis == nil {
		return nil
	}
	keys, err := s.paymentIntentFailureKeys(ctx, paymentIntentID)
	if err != nil {
		return err
	}
	return s.recordFailureForKeys(ctx, keys)

}

func (s *Service) RecordPaymentIntentSuccess(ctx context.Context, paymentIntentID string) error {
	if s == nil || s.redis == nil {
		return nil
	}
	keys, err := s.paymentIntentFailureKeys(ctx, paymentIntentID)
	if err != nil {
		return err
	}
	if err := s.recordSuccessForKeys(ctx, keys); err != nil {
		return err
	}
	if strings.TrimSpace(paymentIntentID) != "" {
		if err := s.redis.Del(ctx, paymentIntentBindingKey(paymentIntentID)).Err(); err != nil {
			return ErrServiceUnavailable
		}
	}
	return nil

}

func (s *Service) RecordProviderFailure(provider string) {
	metrics.PaymentAttempts.WithLabelValues(strings.ToLower(strings.TrimSpace(provider)), "failed").Inc()
}

func (s *Service) RecordProviderSuccess(provider string) {
	metrics.PaymentAttempts.WithLabelValues(strings.ToLower(strings.TrimSpace(provider)), "success").Inc()
}

func (s *Service) failureCount(ctx context.Context, key string) (int, error) {
	return s.failureCountForKeys(ctx, failureKeys(key))

}

func (s *Service) failureCountForKeys(ctx context.Context, keys []string) (int, error) {
	maxFailures := 0
	for _, currentKey := range uniqueKeys(keys) {
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

func (s *Service) recordFailureForKeys(ctx context.Context, keys []string) error {
	keys = uniqueKeys(keys)
	if len(keys) == 0 {
		return nil
	}
	for _, currentKey := range keys {
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

func (s *Service) recordSuccessForKeys(ctx context.Context, keys []string) error {
	keys = uniqueKeys(keys)
	if len(keys) == 0 {
		return nil
	}
	redisKeys := make([]string, 0, len(keys))
	for _, currentKey := range keys {
		redisKeys = append(redisKeys, failureKey(currentKey))
	}
	if err := s.redis.Del(ctx, redisKeys...).Err(); err != nil {
		return ErrServiceUnavailable
	}
	metrics.PaymentAttempts.WithLabelValues("checkout", "success").Inc()
	return nil

}

func (s *Service) paymentIntentFailureKeys(ctx context.Context, paymentIntentID string) ([]string, error) {
	paymentIntentID = strings.TrimSpace(paymentIntentID)
	if paymentIntentID == "" {
		return nil, nil
	}
	value, err := s.redis.Get(ctx, paymentIntentBindingKey(paymentIntentID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, ErrServiceUnavailable
	}
	var binding paymentIntentBinding
	if err := json.Unmarshal(value, &binding); err != nil {
		return nil, fmt.Errorf("%w: decode payment intent risk binding", ErrServiceUnavailable)
	}
	return uniqueKeys(binding.Keys), nil

}

type paymentIntentBinding struct {
	Keys []string `json:"keys"`
}

func failureKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		key = "anonymous"
	}
	return "tanzanite:payment-risk:failures:" + key
}

func paymentIntentBindingKey(paymentIntentID string) string {
	return "tanzanite:payment-risk:payment-intent:" + digestKey(paymentIntentID)

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

func (identity AttemptIdentity) FailureKeys() []string {
	keys := []string{}
	provider := normalizeKeyPart(identity.Provider)
	if provider == "" {
		provider = "payment"
	}
	if userID := normalizeKeyPart(identity.UserID); userID != "" {
		keys = append(keys, provider+":user:"+digestKey(userID))
	}
	if sessionID := normalizeKeyPart(identity.SessionID); sessionID != "" {
		keys = append(keys, provider+":session:"+digestKey(sessionID))
	}
	if anonymousID := normalizeKeyPart(identity.AnonymousID); anonymousID != "" {
		keys = append(keys, provider+":anonymous:"+digestKey(anonymousID))
	}
	ip := normalizeKeyPart(identity.IPAddress)
	if ip != "" {
		keys = append(keys, provider+":ip:"+digestKey(ip))
	}
	userAgent := normalizeKeyPart(identity.UserAgent)
	if ip != "" && userAgent != "" {
		keys = append(keys, provider+":ipua:"+digestKey(ip+"|"+userAgent))
	}
	return uniqueKeys(keys)

}

func uniqueKeys(keys []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	return result

}

func normalizeKeyPart(value string) string {
	return strings.ToLower(strings.TrimSpace(value))

}

func digestKey(value string) string {
	sum := sha256.Sum256([]byte(normalizeKeyPart(value)))
	return hex.EncodeToString(sum[:])

}

func normalizeCountry(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}
