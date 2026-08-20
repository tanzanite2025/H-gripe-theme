package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/pkg/config"
	appLogger "commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/pkg/visitorcookie"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const quickBuyRateLimitRedisPrefix = "commerce_platform:quick-buy-rate-limit:v1"

var quickBuyRateLimitScript = redis.NewScript(`
local now_ms = tonumber(ARGV[1])
local ip_refill_per_ms = tonumber(ARGV[2])
local ip_capacity = tonumber(ARGV[3])
local ip_ttl_seconds = tonumber(ARGV[4])
local session_refill_per_ms = tonumber(ARGV[5])
local session_capacity = tonumber(ARGV[6])
local session_ttl_seconds = tonumber(ARGV[7])
local check_session = tonumber(ARGV[8])

local function current_bucket(key, refill_per_ms, capacity)
  local values = redis.call("HMGET", key, "tokens", "updated_at_ms")
  local tokens = tonumber(values[1])
  local updated_at_ms = tonumber(values[2])
  if tokens == nil or updated_at_ms == nil then
    tokens = capacity
    updated_at_ms = now_ms
  end
  if updated_at_ms > now_ms then
    updated_at_ms = now_ms
  end
  local elapsed_ms = now_ms - updated_at_ms
  tokens = math.min(capacity, tokens + (elapsed_ms * refill_per_ms))
  return tokens, updated_at_ms
end

local ip_tokens, ip_updated_at_ms = current_bucket(KEYS[1], ip_refill_per_ms, ip_capacity)
local session_tokens = 0
local session_updated_at_ms = now_ms
if check_session == 1 then
  session_tokens, session_updated_at_ms = current_bucket(KEYS[2], session_refill_per_ms, session_capacity)
end

if ip_tokens < 1 then
  local retry_ms = math.ceil((1 - ip_tokens) / ip_refill_per_ms)
  return {0, "ip", math.ceil(retry_ms / 1000), math.floor(ip_tokens)}
end
if check_session == 1 and session_tokens < 1 then
  local retry_ms = math.ceil((1 - session_tokens) / session_refill_per_ms)
  return {0, "session", math.ceil(retry_ms / 1000), math.floor(session_tokens)}
end

ip_tokens = ip_tokens - 1
redis.call("HSET", KEYS[1], "tokens", ip_tokens, "updated_at_ms", now_ms)
redis.call("EXPIRE", KEYS[1], ip_ttl_seconds)

local remaining = math.floor(ip_tokens)
if check_session == 1 then
  session_tokens = session_tokens - 1
  redis.call("HSET", KEYS[2], "tokens", session_tokens, "updated_at_ms", now_ms)
  redis.call("EXPIRE", KEYS[2], session_ttl_seconds)
  remaining = math.min(remaining, math.floor(session_tokens))
end

return {1, "", 0, remaining}
`)

type QuickBuyRateLimitDecision struct {
	Allowed        bool
	Dimension      string
	RetryAfter     time.Duration
	Remaining      int
	UnavailableErr error
}

type quickBuyRateLimitStore interface {
	Allow(context.Context, quickBuyRateLimitEvaluation) (QuickBuyRateLimitDecision, error)
}

type quickBuyRateLimitEvaluation struct {
	Keys []string
	Args []interface{}
}

type redisQuickBuyRateLimitStore struct {
	client redis.UniversalClient
}

type QuickBuyRateLimiter struct {
	store quickBuyRateLimitStore
	cfg   config.QuickBuyRateLimitConfig
	now   func() time.Time
}

func NewQuickBuyRateLimiter(redisClient redis.UniversalClient, cfg config.QuickBuyRateLimitConfig) *QuickBuyRateLimiter {
	if redisClient == nil {
		return &QuickBuyRateLimiter{cfg: cfg, now: time.Now}
	}
	return &QuickBuyRateLimiter{
		store: redisQuickBuyRateLimitStore{client: redisClient},
		cfg:   cfg,
		now:   time.Now,
	}
}

func QuickBuyRateLimit(redisClient redis.UniversalClient, cfg config.QuickBuyRateLimitConfig) gin.HandlerFunc {
	return NewQuickBuyRateLimiter(redisClient, cfg).Middleware()
}

func (l *QuickBuyRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if l == nil || !l.cfg.Enabled {
			c.Next()
			return
		}

		decision, err := l.Allow(c.Request.Context(), quickBuyRateLimitIdentityFromContext(c))
		if err != nil {
			appLogger.Warn("quick buy rate limit unavailable",
				zap.Error(err),
				zap.Bool("fail_open", l.cfg.FailOpen),
				zap.String("path", c.Request.URL.Path),
			)
			if l.cfg.FailOpen {
				c.Next()
				return
			}
			c.Header("Cache-Control", "no-store, max-age=0")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "quick_buy_rate_limit_unavailable",
				"message": "QUICK Buy protection is temporarily unavailable. Please try again later.",
			})
			c.Abort()
			return
		}

		if !decision.Allowed {
			retryAfterSeconds := retryAfterSeconds(decision.RetryAfter)
			c.Header("Retry-After", strconv.Itoa(retryAfterSeconds))
			c.Header("Cache-Control", "no-store, max-age=0")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "quick_buy_rate_limited",
				"message":     "Too many QUICK Buy requests. Please try again later.",
				"dimension":   decision.Dimension,
				"retry_after": retryAfterSeconds,
			})
			c.Abort()
			return
		}

		if decision.Remaining >= 0 {
			c.Header("X-RateLimit-Remaining", strconv.Itoa(decision.Remaining))
		}
		c.Next()
	}
}

func (l *QuickBuyRateLimiter) Allow(ctx context.Context, identity quickBuyRateLimitIdentity) (QuickBuyRateLimitDecision, error) {
	decision := QuickBuyRateLimitDecision{
		Allowed:   true,
		Remaining: -1,
	}
	if l == nil || !l.cfg.Enabled {
		return decision, nil
	}
	if l.store == nil {
		return decision, fmt.Errorf("quick buy rate limit redis client is not configured")
	}
	if strings.TrimSpace(identity.IPAddress) == "" {
		return decision, fmt.Errorf("quick buy rate limit identity is missing an IP address")
	}

	result, err := l.store.Allow(ctx, l.evaluation(identity))
	if err != nil {
		return decision, err
	}
	return result, nil
}

func (l *QuickBuyRateLimiter) evaluation(identity quickBuyRateLimitIdentity) quickBuyRateLimitEvaluation {
	now := l.now
	if now == nil {
		now = time.Now
	}

	keys := []string{
		quickBuyRateLimitKey("ip", identity.IPAddress),
		quickBuyRateLimitKey("session", identity.SessionID),
	}
	checkSession := 0
	if strings.TrimSpace(identity.SessionID) != "" {
		checkSession = 1
	}

	return quickBuyRateLimitEvaluation{
		Keys: keys,
		Args: []interface{}{
			now().UTC().UnixMilli(),
			refillPerMillisecond(l.cfg.IPRequestsPerMinute),
			l.cfg.IPBurst,
			bucketTTLSeconds(l.cfg.IPRequestsPerMinute, l.cfg.IPBurst),
			refillPerMillisecond(l.cfg.SessionRequestsPerMinute),
			l.cfg.SessionBurst,
			bucketTTLSeconds(l.cfg.SessionRequestsPerMinute, l.cfg.SessionBurst),
			checkSession,
		},
	}
}

func (s redisQuickBuyRateLimitStore) Allow(ctx context.Context, evaluation quickBuyRateLimitEvaluation) (QuickBuyRateLimitDecision, error) {
	values, err := quickBuyRateLimitScript.Run(ctx, s.client, evaluation.Keys, evaluation.Args...).Slice()
	if err != nil {
		return QuickBuyRateLimitDecision{}, err
	}
	return decodeQuickBuyRateLimitDecision(values), nil
}

func decodeQuickBuyRateLimitDecision(values []interface{}) QuickBuyRateLimitDecision {
	decision := QuickBuyRateLimitDecision{Allowed: true, Remaining: -1}
	if len(values) < 4 {
		return decision
	}
	decision.Allowed = quickBuyRedisScriptInt64(values[0]) == 1
	decision.Dimension = quickBuyRedisScriptString(values[1])
	decision.RetryAfter = time.Duration(quickBuyRedisScriptInt64(values[2])) * time.Second
	decision.Remaining = int(quickBuyRedisScriptInt64(values[3]))
	return decision
}

func quickBuyRedisScriptInt64(value interface{}) int64 {
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	case string:
		parsed, err := strconv.ParseInt(typed, 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	case []byte:
		parsed, err := strconv.ParseInt(string(typed), 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func quickBuyRedisScriptString(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return ""
	}
}

type quickBuyRateLimitIdentity struct {
	IPAddress string
	SessionID string
}

func quickBuyRateLimitIdentityFromContext(c *gin.Context) quickBuyRateLimitIdentity {
	if c == nil {
		return quickBuyRateLimitIdentity{}
	}
	return quickBuyRateLimitIdentity{
		IPAddress: c.ClientIP(),
		SessionID: quickBuyRateLimitSessionID(c),
	}
}

func quickBuyRateLimitSessionID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	for _, value := range []string{
		c.Param("token"),
		c.GetHeader("X-Quick-Buy-Session"),
		c.GetHeader("X-Anonymous-ID"),
	} {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	for _, cookieName := range []string{"session_id", visitorcookie.CustomerServiceVisitorCookie} {
		if value, err := c.Cookie(cookieName); err == nil && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func quickBuyRateLimitKey(scope, value string) string {
	return quickBuyRateLimitRedisPrefix + ":" + strings.TrimSpace(scope) + ":" + quickBuyRateLimitDigest(value)
}

func quickBuyRateLimitDigest(value string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:])
}

func refillPerMillisecond(requestsPerMinute int) float64 {
	if requestsPerMinute <= 0 {
		return 1.0 / 60000.0
	}
	return float64(requestsPerMinute) / 60000.0
}

func bucketTTLSeconds(requestsPerMinute, burst int) int {
	if requestsPerMinute <= 0 || burst <= 0 {
		return 120
	}
	seconds := int(math.Ceil(float64(burst) / (float64(requestsPerMinute) / 60.0) * 2.0))
	if seconds < 60 {
		return 60
	}
	return seconds
}

func retryAfterSeconds(duration time.Duration) int {
	if duration <= 0 {
		return 1
	}
	seconds := int(math.Ceil(duration.Seconds()))
	if seconds < 1 {
		return 1
	}
	return seconds
}
