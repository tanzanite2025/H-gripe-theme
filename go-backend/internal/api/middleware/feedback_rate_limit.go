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
	"sync"
	"time"

	"commerce-platform/internal/pkg/config"
	appLogger "commerce-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const feedbackRateLimitRedisPrefix = "commerce_platform:feedback-rate-limit:v1"

var feedbackRateLimitScript = redis.NewScript(`
local now_ms = tonumber(ARGV[1])
local primary_refill_per_ms = tonumber(ARGV[2])
local primary_capacity = tonumber(ARGV[3])
local primary_ttl_seconds = tonumber(ARGV[4])
local secondary_refill_per_ms = tonumber(ARGV[5])
local secondary_capacity = tonumber(ARGV[6])
local secondary_ttl_seconds = tonumber(ARGV[7])
local check_secondary = tonumber(ARGV[8])
local blocked_ttl_seconds = tonumber(ARGV[9])

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

local primary_tokens, primary_updated_at_ms = current_bucket(KEYS[1], primary_refill_per_ms, primary_capacity)
local secondary_tokens = 0
local secondary_updated_at_ms = now_ms
if check_secondary == 1 then
  secondary_tokens, secondary_updated_at_ms = current_bucket(KEYS[2], secondary_refill_per_ms, secondary_capacity)
end

if primary_tokens < 1 then
  local retry_ms = math.ceil((1 - primary_tokens) / primary_refill_per_ms)
  redis.call("INCR", KEYS[3])
  redis.call("EXPIRE", KEYS[3], blocked_ttl_seconds)
  redis.call("INCR", KEYS[5])
  redis.call("EXPIRE", KEYS[5], blocked_ttl_seconds)
  return {0, "ip", math.ceil(retry_ms / 1000), math.floor(primary_tokens)}
end
if check_secondary == 1 and secondary_tokens < 1 then
  local retry_ms = math.ceil((1 - secondary_tokens) / secondary_refill_per_ms)
  redis.call("INCR", KEYS[4])
  redis.call("EXPIRE", KEYS[4], blocked_ttl_seconds)
  redis.call("INCR", KEYS[5])
  redis.call("EXPIRE", KEYS[5], blocked_ttl_seconds)
  return {0, "user", math.ceil(retry_ms / 1000), math.floor(secondary_tokens)}
end

primary_tokens = primary_tokens - 1
redis.call("HSET", KEYS[1], "tokens", primary_tokens, "updated_at_ms", now_ms)
redis.call("EXPIRE", KEYS[1], primary_ttl_seconds)

local remaining = math.floor(primary_tokens)
if check_secondary == 1 then
  secondary_tokens = secondary_tokens - 1
  redis.call("HSET", KEYS[2], "tokens", secondary_tokens, "updated_at_ms", now_ms)
  redis.call("EXPIRE", KEYS[2], secondary_ttl_seconds)
  remaining = math.min(remaining, math.floor(secondary_tokens))
end

return {1, "", 0, remaining}
`)

type FeedbackRateLimitDecision struct {
	Allowed        bool
	Dimension      string
	RetryAfter     time.Duration
	Remaining      int
	Mode           string
	UnavailableErr error
}

type feedbackRateLimitStore interface {
	Allow(context.Context, feedbackRateLimitEvaluation) (FeedbackRateLimitDecision, error)
}

type feedbackRateLimitEvaluation struct {
	Keys []string
	Args []interface{}
}

type redisFeedbackRateLimitStore struct {
	client redis.UniversalClient
}

type FeedbackRateLimitBlockedSummary struct {
	WindowHours      int   `json:"window_hours"`
	Total            int64 `json:"total"`
	ReadIP           int64 `json:"read_ip"`
	WriteIP          int64 `json:"write_ip"`
	WriteUser        int64 `json:"write_user"`
	RedisUnavailable int64 `json:"redis_unavailable"`
}

type FeedbackRateLimiter struct {
	store         feedbackRateLimitStore
	localFallback *feedbackLocalRateLimitStore
	cfg           config.FeedbackRateLimitConfig
	kind          feedbackRateLimitKind
	now           func() time.Time
}

type feedbackRateLimitKind string

const (
	feedbackRateLimitRead              feedbackRateLimitKind = "read"
	feedbackRateLimitWrite             feedbackRateLimitKind = "write"
	feedbackRateLimitModeRedis                               = "redis"
	feedbackRateLimitModeLocalFallback                       = "local_fallback"
	feedbackLocalRateLimitMaxBuckets                         = 20000
)

var defaultFeedbackLocalRateLimitStore = newFeedbackLocalRateLimitStore(time.Now)

func NewFeedbackReadRateLimiter(redisClient redis.UniversalClient, cfg config.FeedbackRateLimitConfig) *FeedbackRateLimiter {
	return newFeedbackRateLimiter(redisClient, cfg, feedbackRateLimitRead)
}

func NewFeedbackWriteRateLimiter(redisClient redis.UniversalClient, cfg config.FeedbackRateLimitConfig) *FeedbackRateLimiter {
	return newFeedbackRateLimiter(redisClient, cfg, feedbackRateLimitWrite)
}

func newFeedbackRateLimiter(redisClient redis.UniversalClient, cfg config.FeedbackRateLimitConfig, kind feedbackRateLimitKind) *FeedbackRateLimiter {
	limiter := &FeedbackRateLimiter{
		cfg:           cfg,
		kind:          kind,
		now:           time.Now,
		localFallback: defaultFeedbackLocalRateLimitStore,
	}
	if redisClient != nil {
		limiter.store = redisFeedbackRateLimitStore{client: redisClient}
	}
	return limiter
}

func FeedbackReadRateLimit(redisClient redis.UniversalClient, cfg config.FeedbackRateLimitConfig) gin.HandlerFunc {
	return NewFeedbackReadRateLimiter(redisClient, cfg).Middleware()
}

func FeedbackWriteRateLimit(redisClient redis.UniversalClient, cfg config.FeedbackRateLimitConfig) gin.HandlerFunc {
	return NewFeedbackWriteRateLimiter(redisClient, cfg).Middleware()
}

func (l *FeedbackRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if l == nil || !l.cfg.Enabled {
			c.Next()
			return
		}

		identity := feedbackRateLimitIdentityFromContext(c)
		decision, err := l.Allow(c.Request.Context(), identity)
		if err != nil {
			appLogger.Warn("feedback rate limit unavailable",
				zap.Error(err),
				zap.Bool("fail_open", l.cfg.FailOpen),
				zap.String("kind", string(l.kind)),
				zap.String("path", c.Request.URL.Path),
			)
			if fallbackDecision, fallbackErr := l.allowLocalFallback(identity); fallbackErr == nil {
				decision = fallbackDecision
			} else {
				appLogger.Warn("feedback local fallback rate limit unavailable",
					zap.Error(fallbackErr),
					zap.Bool("fail_open", l.cfg.FailOpen),
					zap.String("kind", string(l.kind)),
					zap.String("path", c.Request.URL.Path),
				)
				if l.cfg.FailOpen {
					c.Next()
					return
				}
				c.Header("Cache-Control", "no-store, max-age=0")
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error":   "feedback_rate_limit_unavailable",
					"message": "Feedback protection is temporarily unavailable. Please try again later.",
				})
				c.Abort()
				return
			}
		}

		if decision.Mode != "" {
			c.Header("X-Feedback-RateLimit-Mode", decision.Mode)
		}
		if decision.Remaining >= 0 {
			c.Header("X-RateLimit-Remaining", strconv.Itoa(decision.Remaining))
		}

		if !decision.Allowed {
			retryAfterSeconds := retryAfterSeconds(decision.RetryAfter)
			c.Header("Retry-After", strconv.Itoa(retryAfterSeconds))
			c.Header("Cache-Control", "no-store, max-age=0")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "feedback_rate_limited",
				"message":     "Too many feedback requests. Please try again later.",
				"dimension":   decision.Dimension,
				"mode":        decision.Mode,
				"retry_after": retryAfterSeconds,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (l *FeedbackRateLimiter) Allow(ctx context.Context, identity feedbackRateLimitIdentity) (FeedbackRateLimitDecision, error) {
	decision := FeedbackRateLimitDecision{
		Allowed:   true,
		Remaining: -1,
	}
	if l == nil || !l.cfg.Enabled {
		return decision, nil
	}
	if l.store == nil {
		return decision, fmt.Errorf("feedback rate limit redis client is not configured")
	}
	if strings.TrimSpace(identity.IPAddress) == "" {
		return decision, fmt.Errorf("feedback rate limit identity is missing an IP address")
	}

	result, err := l.store.Allow(ctx, l.evaluation(identity))
	if err != nil {
		return decision, err
	}
	result.Mode = feedbackRateLimitModeRedis
	return result, nil
}

func (l *FeedbackRateLimiter) allowLocalFallback(identity feedbackRateLimitIdentity) (FeedbackRateLimitDecision, error) {
	localFallback := l.localFallback
	if localFallback == nil {
		localFallback = defaultFeedbackLocalRateLimitStore
	}
	now := l.now
	if now == nil {
		now = time.Now
	}
	currentTime := now().UTC()
	localFallback.RecordRedisUnavailable(currentTime, l.kind)
	return localFallback.Allow(currentTime, l.kind, l.cfg, identity)
}

func (l *FeedbackRateLimiter) evaluation(identity feedbackRateLimitIdentity) feedbackRateLimitEvaluation {
	now := l.now
	if now == nil {
		now = time.Now
	}
	currentTime := now().UTC()

	primaryRequestsPerMinute := l.cfg.ReadIPRequestsPerMinute
	primaryBurst := l.cfg.ReadIPBurst
	secondaryScope := ""
	secondaryValue := ""
	secondaryRequestsPerMinute := l.cfg.WriteUserRequestsPerMinute
	secondaryBurst := l.cfg.WriteUserBurst
	if l.kind == feedbackRateLimitWrite {
		primaryRequestsPerMinute = l.cfg.WriteIPRequestsPerMinute
		primaryBurst = l.cfg.WriteIPBurst
		secondaryScope = "user"
		secondaryValue = identity.UserID
	}

	keys := []string{
		feedbackRateLimitKey(string(l.kind), "ip", identity.IPAddress),
		feedbackRateLimitKey(string(l.kind), secondaryScope, secondaryValue),
		feedbackRateLimitBlockedKey(currentTime, l.kind, "ip"),
		feedbackRateLimitBlockedKey(currentTime, l.kind, "user"),
		feedbackRateLimitBlockedKey(currentTime, l.kind, "all"),
	}
	checkSecondary := 0
	if strings.TrimSpace(secondaryValue) != "" {
		checkSecondary = 1
	}

	return feedbackRateLimitEvaluation{
		Keys: keys,
		Args: []interface{}{
			currentTime.UnixMilli(),
			refillPerMillisecond(primaryRequestsPerMinute),
			primaryBurst,
			bucketTTLSeconds(primaryRequestsPerMinute, primaryBurst),
			refillPerMillisecond(secondaryRequestsPerMinute),
			secondaryBurst,
			bucketTTLSeconds(secondaryRequestsPerMinute, secondaryBurst),
			checkSecondary,
			int((48 * time.Hour).Seconds()),
		},
	}
}

func (s redisFeedbackRateLimitStore) Allow(ctx context.Context, evaluation feedbackRateLimitEvaluation) (FeedbackRateLimitDecision, error) {
	values, err := feedbackRateLimitScript.Run(ctx, s.client, evaluation.Keys, evaluation.Args...).Slice()
	if err != nil {
		return FeedbackRateLimitDecision{}, err
	}
	return decodeFeedbackRateLimitDecision(values), nil
}

func decodeFeedbackRateLimitDecision(values []interface{}) FeedbackRateLimitDecision {
	decision := FeedbackRateLimitDecision{Allowed: true, Remaining: -1}
	if len(values) < 4 {
		return decision
	}
	decision.Allowed = quickBuyRedisScriptInt64(values[0]) == 1
	decision.Dimension = quickBuyRedisScriptString(values[1])
	decision.RetryAfter = time.Duration(quickBuyRedisScriptInt64(values[2])) * time.Second
	decision.Remaining = int(quickBuyRedisScriptInt64(values[3]))
	return decision
}

type feedbackRateLimitIdentity struct {
	IPAddress string
	UserID    string
}

func feedbackRateLimitIdentityFromContext(c *gin.Context) feedbackRateLimitIdentity {
	if c == nil {
		return feedbackRateLimitIdentity{}
	}
	return feedbackRateLimitIdentity{
		IPAddress: c.ClientIP(),
		UserID:    feedbackRateLimitUserID(c),
	}
}

func feedbackRateLimitUserID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	userID := c.GetUint("user_id")
	if userID == 0 {
		return ""
	}
	return strconv.FormatUint(uint64(userID), 10)
}

func feedbackRateLimitKey(kind, scope, value string) string {
	parts := []string{
		feedbackRateLimitRedisPrefix,
		strings.TrimSpace(kind),
		strings.TrimSpace(scope),
		feedbackRateLimitDigest(value),
	}
	return strings.Join(parts, ":")
}

func feedbackRateLimitBlockedKey(at time.Time, kind feedbackRateLimitKind, dimension string) string {
	return strings.Join([]string{
		feedbackRateLimitRedisPrefix,
		"blocked",
		at.UTC().Truncate(time.Hour).Format("2006010215"),
		string(kind),
		strings.TrimSpace(dimension),
	}, ":")
}

type feedbackLocalRateLimitBucket struct {
	Tokens    float64
	UpdatedAt time.Time
}

type feedbackLocalRateLimitStore struct {
	mu          sync.Mutex
	buckets     map[string]feedbackLocalRateLimitBucket
	blocked     map[string]int64
	unavailable map[string]int64
	now         func() time.Time
	nextCleanup time.Time
}

func newFeedbackLocalRateLimitStore(now func() time.Time) *feedbackLocalRateLimitStore {
	if now == nil {
		now = time.Now
	}
	return &feedbackLocalRateLimitStore{
		buckets:     make(map[string]feedbackLocalRateLimitBucket),
		blocked:     make(map[string]int64),
		unavailable: make(map[string]int64),
		now:         now,
	}
}

func (s *feedbackLocalRateLimitStore) Allow(at time.Time, kind feedbackRateLimitKind, cfg config.FeedbackRateLimitConfig, identity feedbackRateLimitIdentity) (FeedbackRateLimitDecision, error) {
	decision := FeedbackRateLimitDecision{
		Allowed:   true,
		Mode:      feedbackRateLimitModeLocalFallback,
		Remaining: -1,
	}
	if s == nil {
		return decision, fmt.Errorf("feedback local fallback store is not configured")
	}
	if strings.TrimSpace(identity.IPAddress) == "" {
		return decision, fmt.Errorf("feedback local fallback identity is missing an IP address")
	}
	if at.IsZero() {
		at = s.now().UTC()
	}
	at = at.UTC()

	primaryRequestsPerMinute := cfg.ReadIPRequestsPerMinute
	primaryBurst := cfg.ReadIPBurst
	secondaryValue := ""
	secondaryRequestsPerMinute := cfg.WriteUserRequestsPerMinute
	secondaryBurst := cfg.WriteUserBurst
	if kind == feedbackRateLimitWrite {
		primaryRequestsPerMinute = cfg.WriteIPRequestsPerMinute
		primaryBurst = cfg.WriteIPBurst
		secondaryValue = strings.TrimSpace(identity.UserID)
	}

	primaryRefillPerMillisecond := refillPerMillisecond(primaryRequestsPerMinute)
	secondaryRefillPerMillisecond := refillPerMillisecond(secondaryRequestsPerMinute)
	primaryCapacity := feedbackRateLimitPositiveBurst(primaryBurst)
	secondaryCapacity := feedbackRateLimitPositiveBurst(secondaryBurst)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(at)

	primaryKey := feedbackRateLimitKey(string(kind), "local-ip", identity.IPAddress)
	_, primaryExists := s.buckets[primaryKey]
	if !primaryExists && len(s.buckets) >= feedbackLocalRateLimitMaxBuckets {
		s.recordBlockedLocked(at, kind, "ip")
		return feedbackLocalRateLimitCapacityDecision("ip"), nil
	}
	primaryBucket := s.refilledBucketLocked(primaryKey, primaryRefillPerMillisecond, primaryCapacity, at)
	if primaryBucket.Tokens < 1 {
		s.recordBlockedLocked(at, kind, "ip")
		return FeedbackRateLimitDecision{
			Allowed:    false,
			Dimension:  "ip",
			Mode:       feedbackRateLimitModeLocalFallback,
			RetryAfter: feedbackRateLimitRetryAfter(primaryBucket.Tokens, primaryRefillPerMillisecond),
			Remaining:  int(math.Floor(primaryBucket.Tokens)),
		}, nil
	}

	var secondaryKey string
	var secondaryBucket feedbackLocalRateLimitBucket
	checkSecondary := secondaryValue != ""
	if checkSecondary {
		secondaryKey = feedbackRateLimitKey(string(kind), "local-user", secondaryValue)
		pendingPrimaryInsert := 0
		if !primaryExists {
			pendingPrimaryInsert = 1
		}
		if _, secondaryExists := s.buckets[secondaryKey]; !secondaryExists && len(s.buckets)+pendingPrimaryInsert >= feedbackLocalRateLimitMaxBuckets {
			s.recordBlockedLocked(at, kind, "user")
			return feedbackLocalRateLimitCapacityDecision("user"), nil
		}
		secondaryBucket = s.refilledBucketLocked(secondaryKey, secondaryRefillPerMillisecond, secondaryCapacity, at)
		if secondaryBucket.Tokens < 1 {
			s.recordBlockedLocked(at, kind, "user")
			return FeedbackRateLimitDecision{
				Allowed:    false,
				Dimension:  "user",
				Mode:       feedbackRateLimitModeLocalFallback,
				RetryAfter: feedbackRateLimitRetryAfter(secondaryBucket.Tokens, secondaryRefillPerMillisecond),
				Remaining:  int(math.Floor(secondaryBucket.Tokens)),
			}, nil
		}
	}

	primaryBucket.Tokens--
	primaryBucket.UpdatedAt = at
	s.buckets[primaryKey] = primaryBucket
	remaining := int(math.Floor(primaryBucket.Tokens))
	if checkSecondary {
		secondaryBucket.Tokens--
		secondaryBucket.UpdatedAt = at
		s.buckets[secondaryKey] = secondaryBucket
		if secondaryRemaining := int(math.Floor(secondaryBucket.Tokens)); secondaryRemaining < remaining {
			remaining = secondaryRemaining
		}
	}
	decision.Remaining = remaining
	return decision, nil
}

func feedbackLocalRateLimitCapacityDecision(dimension string) FeedbackRateLimitDecision {
	return FeedbackRateLimitDecision{
		Allowed:    false,
		Dimension:  dimension,
		Mode:       feedbackRateLimitModeLocalFallback,
		RetryAfter: time.Minute,
		Remaining:  0,
	}
}

func (s *feedbackLocalRateLimitStore) BlockedCounts(hours int) FeedbackRateLimitBlockedSummary {
	if hours < 1 || hours > 168 {
		hours = 24
	}
	summary := FeedbackRateLimitBlockedSummary{WindowHours: hours}
	if s == nil {
		return summary
	}
	now := s.now().UTC().Truncate(time.Hour)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(now)

	for offset := 0; offset < hours; offset++ {
		hour := now.Add(-time.Duration(offset) * time.Hour)
		readIP := s.blocked[feedbackLocalRateLimitCounterKey(hour, feedbackRateLimitRead, "ip")]
		writeIP := s.blocked[feedbackLocalRateLimitCounterKey(hour, feedbackRateLimitWrite, "ip")]
		writeUser := s.blocked[feedbackLocalRateLimitCounterKey(hour, feedbackRateLimitWrite, "user")]
		summary.ReadIP += readIP
		summary.WriteIP += writeIP
		summary.WriteUser += writeUser
		summary.Total += readIP + writeIP + writeUser
		summary.RedisUnavailable += s.unavailable[feedbackLocalRateLimitCounterKey(hour, feedbackRateLimitRead, "redis_unavailable")]
		summary.RedisUnavailable += s.unavailable[feedbackLocalRateLimitCounterKey(hour, feedbackRateLimitWrite, "redis_unavailable")]
	}
	return summary
}

func (s *feedbackLocalRateLimitStore) RecordRedisUnavailable(at time.Time, kind feedbackRateLimitKind) {
	if s == nil {
		return
	}
	if at.IsZero() {
		at = s.now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(at.UTC())
	s.unavailable[feedbackLocalRateLimitCounterKey(at, kind, "redis_unavailable")]++
}

func (s *feedbackLocalRateLimitStore) refilledBucketLocked(key string, refillPerMillisecond float64, capacity int, at time.Time) feedbackLocalRateLimitBucket {
	bucket, exists := s.buckets[key]
	if !exists || bucket.UpdatedAt.IsZero() {
		return feedbackLocalRateLimitBucket{
			Tokens:    float64(capacity),
			UpdatedAt: at,
		}
	}
	if bucket.UpdatedAt.After(at) {
		bucket.UpdatedAt = at
	}
	elapsedMilliseconds := float64(at.Sub(bucket.UpdatedAt).Milliseconds())
	if elapsedMilliseconds > 0 {
		bucket.Tokens = math.Min(float64(capacity), bucket.Tokens+(elapsedMilliseconds*refillPerMillisecond))
		bucket.UpdatedAt = at
	}
	return bucket
}

func (s *feedbackLocalRateLimitStore) recordBlockedLocked(at time.Time, kind feedbackRateLimitKind, dimension string) {
	s.blocked[feedbackLocalRateLimitCounterKey(at, kind, dimension)]++
}

func (s *feedbackLocalRateLimitStore) cleanupLocked(at time.Time) {
	if !s.nextCleanup.IsZero() && at.Before(s.nextCleanup) {
		return
	}
	s.nextCleanup = at.Add(time.Hour)
	bucketCutoff := at.Add(-2 * time.Hour)
	for key, bucket := range s.buckets {
		if bucket.UpdatedAt.Before(bucketCutoff) {
			delete(s.buckets, key)
		}
	}
	counterCutoff := at.Add(-169 * time.Hour).Format("2006010215")
	for key := range s.blocked {
		if key < counterCutoff {
			delete(s.blocked, key)
		}
	}
	for key := range s.unavailable {
		if key < counterCutoff {
			delete(s.unavailable, key)
		}
	}
}

func feedbackLocalRateLimitCounterKey(at time.Time, kind feedbackRateLimitKind, dimension string) string {
	return strings.Join([]string{
		at.UTC().Truncate(time.Hour).Format("2006010215"),
		string(kind),
		strings.TrimSpace(dimension),
	}, ":")
}

func feedbackRateLimitPositiveBurst(value int) int {
	if value < 1 {
		return 1
	}
	return value
}

func feedbackRateLimitRetryAfter(tokens float64, refillPerMillisecond float64) time.Duration {
	if refillPerMillisecond <= 0 {
		return time.Second
	}
	missingTokens := 1 - tokens
	if missingTokens <= 0 {
		return time.Second
	}
	milliseconds := int64(math.Ceil(missingTokens / refillPerMillisecond))
	if milliseconds < 1 {
		milliseconds = 1
	}
	return time.Duration(milliseconds) * time.Millisecond
}

func FeedbackLocalRateLimitBlockedCounts(hours int) FeedbackRateLimitBlockedSummary {
	return defaultFeedbackLocalRateLimitStore.BlockedCounts(hours)
}

func FeedbackRateLimitBlockedCounts(ctx context.Context, redisClient redis.UniversalClient, hours int) (FeedbackRateLimitBlockedSummary, error) {
	if hours < 1 || hours > 168 {
		hours = 24
	}
	summary := FeedbackRateLimitBlockedSummary{WindowHours: hours}
	if redisClient == nil {
		return summary, fmt.Errorf("feedback rate limit redis client is not configured")
	}

	now := time.Now().UTC().Truncate(time.Hour)
	keys := make([]string, 0, hours*5)
	for offset := 0; offset < hours; offset++ {
		hour := now.Add(-time.Duration(offset) * time.Hour)
		keys = append(keys,
			feedbackRateLimitBlockedKey(hour, feedbackRateLimitRead, "ip"),
			feedbackRateLimitBlockedKey(hour, feedbackRateLimitWrite, "ip"),
			feedbackRateLimitBlockedKey(hour, feedbackRateLimitWrite, "user"),
			feedbackRateLimitBlockedKey(hour, feedbackRateLimitRead, "all"),
			feedbackRateLimitBlockedKey(hour, feedbackRateLimitWrite, "all"),
		)
	}

	values, err := redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		return summary, err
	}

	for index, value := range values {
		count := feedbackRateLimitCounterValue(value)
		switch index % 5 {
		case 0:
			summary.ReadIP += count
		case 1:
			summary.WriteIP += count
		case 2:
			summary.WriteUser += count
		case 3, 4:
			summary.Total += count
		}
	}
	return summary, nil
}

func feedbackRateLimitCounterValue(value interface{}) int64 {
	switch typed := value.(type) {
	case nil:
		return 0
	case int64:
		return typed
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

func feedbackRateLimitDigest(value string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:])
}
