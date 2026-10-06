package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	tireRimReferenceTraversalWindow        = time.Minute
	tireRimReferenceTraversalStreakLimit   = 10
	tireRimReferenceTraversalBodyBytes     = 16 * 1024
	tireRimReferenceTraversalMaxIdentities = 10000
	tireRimReferenceTraversalRedisPrefix   = "commerce:tire-rim-reference-traversal:v1"
	tireRimReferenceTraversalRedisTimeout  = 100 * time.Millisecond
)

// A tire-width API request is an integer millimetre value. Ten adjacent
// transitions in one minute are enough to distinguish a scripted sweep from
// ordinary fitment lookups while leaving the shared 20/minute token bucket as
// the primary budget for normal users.
var tireRimReferenceTraversalObserveScript = redis.NewScript(`
local current = tonumber(ARGV[1])
local last = redis.call("HGET", KEYS[1], "last_width")
local streak = tonumber(redis.call("HGET", KEYS[1], "streak")) or 0
local direction = tonumber(redis.call("HGET", KEYS[1], "direction")) or 0
if last ~= false and math.abs(current - tonumber(last)) == 1 then
  local next_direction = current > tonumber(last) and 1 or -1
  if direction == next_direction then
    streak = streak + 1
  else
    streak = 1
  end
  direction = next_direction
else
  streak = 0
  direction = 0
end
redis.call("HSET", KEYS[1], "last_width", current, "streak", streak, "direction", direction)
redis.call("EXPIRE", KEYS[1], ARGV[2])
return streak
`)

type tireRimReferenceTraversalBucket struct {
	expiresAt time.Time
	lastWidth int
	hasLast   bool
	direction int
	streak    int
}

type tireRimReferenceTraversalTracker struct {
	mu          sync.Mutex
	buckets     map[string]tireRimReferenceTraversalBucket
	redisClient redis.UniversalClient
	fallback    *tireRimReferenceTraversalTracker
}

func newTireRimReferenceTraversalTracker(redisClient redis.UniversalClient) *tireRimReferenceTraversalTracker {
	if redisClient == nil {
		return &tireRimReferenceTraversalTracker{
			buckets: make(map[string]tireRimReferenceTraversalBucket),
		}
	}
	return &tireRimReferenceTraversalTracker{
		redisClient: redisClient,
		fallback: &tireRimReferenceTraversalTracker{
			buckets: make(map[string]tireRimReferenceTraversalBucket),
		},
	}
}

func (t *tireRimReferenceTraversalTracker) observe(ctx context.Context, key string, width int, now time.Time) int {
	if t == nil || strings.TrimSpace(key) == "" {
		return 0
	}
	if t.redisClient != nil {
		return t.observeRedis(ctx, key, width, now)
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	bucket, exists := t.buckets[key]
	if !exists || !now.Before(bucket.expiresAt) {
		if !exists && len(t.buckets) >= tireRimReferenceTraversalMaxIdentities {
			tireRimReferenceTraversalCleanupExpiredBuckets(t.buckets, now)
			if len(t.buckets) >= tireRimReferenceTraversalMaxIdentities {
				return 0
			}
		}
		bucket = tireRimReferenceTraversalBucket{expiresAt: now.Add(tireRimReferenceTraversalWindow)}
	}
	if bucket.hasLast && tireRimReferenceTraversalAbsoluteIntegerDifference(width, bucket.lastWidth) == 1 {
		direction := 1
		if width < bucket.lastWidth {
			direction = -1
		}
		if bucket.direction == direction {
			bucket.streak++
		} else {
			bucket.streak = 1
		}
		bucket.direction = direction
	} else {
		bucket.streak = 0
		bucket.direction = 0
	}
	bucket.lastWidth = width
	bucket.hasLast = true
	t.buckets[key] = bucket
	return bucket.streak
}

func (t *tireRimReferenceTraversalTracker) observeRedis(ctx context.Context, key string, width int, now time.Time) int {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	digest := tireRimReferenceTraversalDigest(key)
	redisKey := tireRimReferenceTraversalRedisPrefix + ":" + digest
	operationContext, cancel := context.WithTimeout(tireRimReferenceTraversalContextOrBackground(ctx), tireRimReferenceTraversalRedisTimeout)
	defer cancel()

	result, err := tireRimReferenceTraversalObserveScript.Run(
		operationContext,
		t.redisClient,
		[]string{redisKey},
		width,
		int(tireRimReferenceTraversalWindow/time.Second),
	).Result()
	if err == nil {
		if streak, ok := redisResultInt(result); ok {
			return streak
		}
	}
	if t.fallback != nil {
		return t.fallback.observe(ctx, key, width, now)
	}
	return 0
}

// TireRimReferenceTraversalGuard detects a uniform adjacent-width sweep on the
// tire/rim solve endpoint. It restores the request body before the handler so
// this behavior check never changes the API contract.
func TireRimReferenceTraversalGuard(redisClients ...redis.UniversalClient) gin.HandlerFunc {
	var redisClient redis.UniversalClient
	if len(redisClients) > 0 {
		redisClient = redisClients[0]
	}
	tracker := newTireRimReferenceTraversalTracker(redisClient)

	return func(c *gin.Context) {
		if c == nil || c.Request == nil || c.Request.Method != http.MethodPost ||
			!strings.HasSuffix(strings.TrimSpace(c.Request.URL.Path), "/api/v1/engineering/tire-rim/solve") {
			c.Next()
			return
		}

		width, ok := readTireRimReferenceTraversalWidth(c)
		if !ok {
			c.Next()
			return
		}
		identity := tireRimReferenceTraversalIdentity(c)
		if identity == "" {
			c.Next()
			return
		}
		if tracker.observe(commercialRequestContext(c), identity, width, time.Now().UTC()) >= tireRimReferenceTraversalStreakLimit {
			abortTireRimReferenceTraversal(c)
			return
		}
		c.Next()
	}
}

func readTireRimReferenceTraversalWidth(c *gin.Context) (int, bool) {
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return 0, false
	}
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, tireRimReferenceTraversalBodyBytes+1))
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if err != nil || len(body) > tireRimReferenceTraversalBodyBytes {
		return 0, false
	}
	var payload struct {
		TireWidthMM *int `json:"tire_width_mm"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || payload.TireWidthMM == nil {
		return 0, false
	}
	return *payload.TireWidthMM, true
}

func tireRimReferenceTraversalIdentity(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	ip := strings.TrimSpace(c.ClientIP())
	if ip == "" {
		return ""
	}
	identity := ip + "|" + strings.TrimSpace(c.GetHeader("X-Device-Fingerprint"))
	if userID, exists := c.Get("user_id"); exists {
		if userIdentity := tireRimReferenceUserIdentity(userID); userIdentity != "" {
			identity = "user:" + userIdentity
		}
	}
	return tireRimReferenceTraversalDigest(identity)
}

func tireRimReferenceTraversalDigest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func tireRimReferenceTraversalAbsoluteIntegerDifference(first, second int) int {
	if first >= second {
		return first - second
	}
	return second - first
}

func tireRimReferenceTraversalContextOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func tireRimReferenceTraversalCleanupExpiredBuckets(
	buckets map[string]tireRimReferenceTraversalBucket,
	now time.Time,
) {
	for key, bucket := range buckets {
		if !now.Before(bucket.expiresAt) {
			delete(buckets, key)
		}
	}
}

func tireRimReferenceUserIdentity(value any) string {
	switch typed := value.(type) {
	case uint:
		return strconv.FormatUint(uint64(typed), 10)
	case uint8:
		return strconv.FormatUint(uint64(typed), 10)
	case uint16:
		return strconv.FormatUint(uint64(typed), 10)
	case uint32:
		return strconv.FormatUint(uint64(typed), 10)
	case uint64:
		return strconv.FormatUint(typed, 10)
	case int:
		if typed > 0 {
			return strconv.FormatInt(int64(typed), 10)
		}
	case int8:
		if typed > 0 {
			return strconv.FormatInt(int64(typed), 10)
		}
	case int16:
		if typed > 0 {
			return strconv.FormatInt(int64(typed), 10)
		}
	case int32:
		if typed > 0 {
			return strconv.FormatInt(int64(typed), 10)
		}
	case int64:
		if typed > 0 {
			return strconv.FormatInt(typed, 10)
		}
	case string:
		return strings.TrimSpace(typed)
	}
	return ""
}

func abortTireRimReferenceTraversal(c *gin.Context) {
	c.Header("Retry-After", strconv.Itoa(int(tireRimReferenceTraversalWindow/time.Second)))
	c.Header("Cache-Control", "no-store, max-age=0")
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"error":   "calculator_traversal_rate_limited",
		"message": "Calculator requests are temporarily limited. Please try again later.",
	})
}
