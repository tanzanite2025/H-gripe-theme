package antibot

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/metrics"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrChallengeRequired  = errors.New("verification challenge is required")
	ErrChallengeInvalid   = errors.New("verification challenge is invalid")
	ErrRateLimited        = errors.New("verification rate limit exceeded")
	ErrBudgetExceeded     = errors.New("verification delivery budget exceeded")
	ErrCircuitOpen        = errors.New("verification delivery circuit is open")
	ErrServiceUnavailable = errors.New("anti-abuse service is unavailable")
)

type Service struct {
	redis      *redis.Client
	config     config.AntiAbuseConfig
	httpClient *http.Client
}

func New(redisClient *redis.Client, cfg config.AntiAbuseConfig) *Service {
	if redisClient == nil {
		return nil
	}
	return &Service{
		redis:  redisClient,
		config: cfg,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *Service) Required() bool {
	return s != nil && s.config.TurnstileRequired
}

func (s *Service) Guard(ctx context.Context, channel, destination, ip, challengeToken string) error {
	if s == nil {
		return nil
	}
	if err := s.VerifyChallenge(ctx, challengeToken, ip); err != nil {
		return err
	}
	return s.acquireDeliveryBudget(ctx, channel, destination, ip)
}

func (s *Service) VerifyChallenge(ctx context.Context, challengeToken, ip string) error {
	if s == nil {
		return nil
	}
	return s.verifyTurnstile(ctx, challengeToken, ip)
}

func (s *Service) RecordDeliveryResult(channel string, success bool) {
	result := "failed"
	if success {
		result = "success"
	}
	metrics.VerificationSendAttempts.WithLabelValues(normalizeChannel(channel), result).Inc()
}

func (s *Service) verifyTurnstile(ctx context.Context, token, remoteIP string) error {
	if !s.config.TurnstileRequired {
		return nil
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrChallengeRequired
	}
	secret := strings.TrimSpace(s.config.TurnstileSecretKey)
	if secret == "" {
		return ErrServiceUnavailable
	}

	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", token)
	if strings.TrimSpace(remoteIP) != "" {
		form.Set("remoteip", remoteIP)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return ErrServiceUnavailable
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := s.httpClient.Do(request)
	if err != nil {
		return ErrServiceUnavailable
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode != http.StatusOK {
		return ErrServiceUnavailable
	}

	var payload struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return ErrServiceUnavailable
	}
	if !payload.Success {
		return ErrChallengeInvalid
	}
	return nil
}

func (s *Service) acquireDeliveryBudget(ctx context.Context, channel, destination, ip string) error {
	if s.redis == nil {
		return ErrServiceUnavailable
	}

	channel = normalizeChannel(channel)
	destinationKey := digestKey(destination)
	ipKey := digestKey(ip)
	now := time.Now().Unix()
	keys := []string{
		"tanzanite:verification:ip:" + channel + ":" + ipKey,
		"tanzanite:verification:destination:" + channel + ":" + destinationKey,
		"tanzanite:verification:daily:" + channel + ":" + time.Now().UTC().Format("20060102") + ":" + destinationKey,
		"tanzanite:verification:global:" + channel,
		"tanzanite:verification:circuit:" + channel,
	}

	result, err := s.redis.Eval(ctx, deliveryBudgetScript, keys, []interface{}{
		now,
		s.config.VerificationIPWindowSeconds,
		s.config.VerificationDestinationWindowSeconds,
		s.config.VerificationDailyLimit,
		s.config.VerificationGlobalWindowSeconds,
		s.config.VerificationGlobalLimit,
		s.config.VerificationCircuitSeconds,
		uuid.NewString(),
	}).Int()
	if err != nil {
		return ErrServiceUnavailable
	}

	switch result {
	case 0:
		metrics.VerificationSendAttempts.WithLabelValues(channel, "accepted").Inc()
		return nil
	case 1:
		metrics.VerificationBudgetRejections.WithLabelValues(channel, "ip").Inc()
		return ErrRateLimited
	case 2:
		metrics.VerificationBudgetRejections.WithLabelValues(channel, "destination").Inc()
		return ErrRateLimited
	case 3:
		metrics.VerificationBudgetRejections.WithLabelValues(channel, "daily").Inc()
		return ErrRateLimited
	case 4:
		metrics.VerificationBudgetRejections.WithLabelValues(channel, "global").Inc()
		if s.isCircuitOpen(ctx, keys[4]) {
			return ErrCircuitOpen
		}
		return ErrBudgetExceeded
	default:
		return fmt.Errorf("%w: unexpected limiter result", ErrServiceUnavailable)
	}
}

func (s *Service) isCircuitOpen(ctx context.Context, key string) bool {
	if s.redis == nil {
		return false
	}
	value, err := s.redis.Exists(ctx, key).Result()
	return err == nil && value > 0
}

func normalizeChannel(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "unknown"
	}
	return value
}

func digestKey(value string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:])
}

const deliveryBudgetScript = `
local now = tonumber(ARGV[1])
local ip_window = tonumber(ARGV[2])
local destination_window = tonumber(ARGV[3])
local daily_limit = tonumber(ARGV[4])
local global_window = tonumber(ARGV[5])
local global_limit = tonumber(ARGV[6])
local circuit_seconds = tonumber(ARGV[7])
local member = ARGV[8]

if redis.call("EXISTS", KEYS[5]) == 1 then
  return 4
end
if redis.call("EXISTS", KEYS[1]) == 1 then
  return 1
end
if tonumber(redis.call("GET", KEYS[2]) or "0") >= 1 then
  return 2
end
if tonumber(redis.call("GET", KEYS[3]) or "0") >= daily_limit then
  return 3
end

redis.call("ZREMRANGEBYSCORE", KEYS[4], 0, now - global_window)
if redis.call("ZCARD", KEYS[4]) >= global_limit then
  redis.call("SET", KEYS[5], "1", "EX", circuit_seconds)
  return 4
end

redis.call("SET", KEYS[1], "1", "EX", ip_window)
local destination_count = redis.call("INCR", KEYS[2])
if destination_count == 1 then
  redis.call("EXPIRE", KEYS[2], destination_window)
end
local daily_count = redis.call("INCR", KEYS[3])
if daily_count == 1 then
  redis.call("EXPIRE", KEYS[3], 172800)
end
redis.call("ZADD", KEYS[4], now, member)
redis.call("EXPIRE", KEYS[4], global_window + circuit_seconds)
return 0
`
