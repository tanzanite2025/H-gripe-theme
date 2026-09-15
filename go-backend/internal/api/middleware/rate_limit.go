package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// RateLimiter 限流器
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rps int, burst int) *RateLimiter {
	return newRateLimiter(rate.Limit(rps), burst)
}

func newRateLimiter(requestRate rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     requestRate,
		burst:    burst,
	}
}

// getLimiter 获取或创建限流器
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter
}

// cleanupOldLimiters 清理旧的限流器
func (rl *RateLimiter) cleanupOldLimiters() {
	ticker := time.NewTicker(time.Hour)
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			// 清空所有限流器，让它们在需要时重新创建
			rl.limiters = make(map[string]*rate.Limiter)
			rl.mu.Unlock()
		}
	}()
}

// RateLimit 限流中间件 - 基于IP
func RateLimit(rps int) gin.HandlerFunc {
	limiter := NewRateLimiter(rps, rps*2)
	limiter.cleanupOldLimiters()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := limiter.getLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByUser 限流中间件 - 基于用户ID
func RateLimitByUser(rps int) gin.HandlerFunc {
	limiter := NewRateLimiter(rps, rps*2)
	limiter.cleanupOldLimiters()

	return func(c *gin.Context) {
		// 尝试从context获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			// 如果没有用户ID，使用IP地址
			userID = c.ClientIP()
		}

		key := fmt.Sprintf("user_%v", userID)
		limiter := limiter.getLimiter(key)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByUserPerMinute limits authenticated actions whose payload or
// downstream work is too expensive for a per-second policy.
func RateLimitByUserPerMinute(requestsPerMinute, burst int) gin.HandlerFunc {
	if requestsPerMinute <= 0 {
		panic("requests per minute must be positive")
	}
	if burst < 1 {
		burst = 1
	}

	limiter := newRateLimiter(rate.Limit(float64(requestsPerMinute)/60), burst)
	limiter.cleanupOldLimiters()

	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = c.ClientIP()
		}

		key := fmt.Sprintf("user_%v", userID)
		if !limiter.getLimiter(key).Allow() {
			c.Header("Retry-After", "20")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByIPAndFingerprintPerMinute scopes anonymous traffic to both the
// network address and the browser fingerprint. This prevents a scraper from
// rotating one dimension while retaining the other, while authenticated users
// continue to be isolated by user ID.
func RateLimitByIPAndFingerprintPerMinute(requestsPerMinute, burst int) gin.HandlerFunc {
	if requestsPerMinute <= 0 {
		panic("requests per minute must be positive")
	}
	if burst < 1 {
		burst = 1
	}
	limiter := newRateLimiter(rate.Limit(float64(requestsPerMinute)/60), burst)
	limiter.cleanupOldLimiters()
	return func(c *gin.Context) {
		key := "ip:" + c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			key = fmt.Sprintf("user:%v", userID)
		} else {
			fingerprint := c.GetHeader("X-Device-Fingerprint")
			digest := sha256.Sum256([]byte(fingerprint))
			key += ":fp:" + fmt.Sprintf("%x", digest[:])
		}
		if !limiter.getLimiter(key).Allow() {
			c.Header("Retry-After", "20")
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate_limit_exceeded", "message": "Calculator rate limit exceeded. Please try again later."})
			c.Abort()
			return
		}
		c.Next()
	}
}

var spokeRateLimitScript = redis.NewScript(`
local now = tonumber(ARGV[1])
local refill = tonumber(ARGV[2])
local capacity = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])
local values = redis.call("HMGET", KEYS[1], "tokens", "updated_at")
local tokens = tonumber(values[1])
local updated = tonumber(values[2])
if tokens == nil or updated == nil then
  tokens = capacity
  updated = now
end
if updated > now then updated = now end
tokens = math.min(capacity, tokens + ((now - updated) * refill))
if tokens < 1 then
  redis.call("HSET", KEYS[1], "tokens", tokens, "updated_at", now)
  redis.call("EXPIRE", KEYS[1], ttl)
  return {0, math.ceil((1 - tokens) / refill)}
end
tokens = tokens - 1
redis.call("HSET", KEYS[1], "tokens", tokens, "updated_at", now)
redis.call("EXPIRE", KEYS[1], ttl)
return {1, math.floor(tokens)}
`)

// SpokeRateLimit is the production limiter for calculator/catalog traffic.
// Redis gives every API replica the same token bucket; the in-process limiter
// is retained only for development/test environments without Redis.
func SpokeRateLimit(redisClient redis.UniversalClient) gin.HandlerFunc {
	local := RateLimitByIPAndFingerprintPerMinute(20, 5)
	return func(c *gin.Context) {
		if redisClient == nil {
			local(c)
			return
		}
		fingerprint := c.GetHeader("X-Device-Fingerprint")
		identity := c.ClientIP() + "|" + fingerprint
		if userID, exists := c.Get("user_id"); exists {
			identity = fmt.Sprintf("user:%v", userID)
		}
		digest := sha256.Sum256([]byte(identity))
		key := "commerce_platform:spoke-rate-limit:v1:" + fmt.Sprintf("%x", digest[:])
		values, err := spokeRateLimitScript.Run(c.Request.Context(), redisClient, []string{key}, time.Now().UnixMilli(), 20.0/60000.0, 5, 120).Slice()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "rate_limit_service_unavailable", "message": "Calculator protection is temporarily unavailable"})
			c.Abort()
			return
		}
		if len(values) < 2 || quickBuyRedisScriptInt64(values[0]) != 1 {
			c.Header("Retry-After", "1")
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate_limit_exceeded", "message": "Calculator rate limit exceeded. Please try again later."})
			c.Abort()
			return
		}
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(quickBuyRedisScriptInt64(values[1]), 10))
		c.Next()
	}
}

// RateLimitByUserPerMinuteRedis allows one request per authenticated user in a
// shared Redis-backed minute window. It fails closed when Redis is unavailable
// because the protected endpoint has an external side effect.
func RateLimitByUserPerMinuteRedis(redisClient redis.UniversalClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		if redisClient == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "rate_limit_service_unavailable",
				"message": "Rate limit service is temporarily unavailable",
			})
			c.Abort()
			return
		}

		identity := "ip:" + c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			identity = fmt.Sprintf("user:%v", userID)
		}
		key := "commerce_platform:rate_limit:google_indexing:" + identity

		ctx := context.Background()
		if c.Request != nil {
			ctx = c.Request.Context()
		}
		operationCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		allowed, err := redisClient.SetNX(operationCtx, key, "1", time.Minute).Result()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "rate_limit_service_unavailable",
				"message": "Rate limit service is temporarily unavailable",
			})
			c.Abort()
			return
		}
		if allowed {
			c.Next()
			return
		}

		retryAfter := 60
		if ttl, ttlErr := redisClient.TTL(operationCtx, key).Result(); ttlErr == nil && ttl > 0 {
			retryAfter = int((ttl + time.Second - 1) / time.Second)
			if retryAfter < 1 {
				retryAfter = 1
			}
		}
		c.Header("Retry-After", strconv.Itoa(retryAfter))
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":   "rate_limit_exceeded",
			"message": "Google Indexing notification rate limit exceeded",
		})
		c.Abort()
	}
}

// RateLimitByEndpoint 限流中间件 - 基于端点
func RateLimitByEndpoint(rps int) gin.HandlerFunc {
	limiter := NewRateLimiter(rps, rps*2)
	limiter.cleanupOldLimiters()

	return func(c *gin.Context) {
		key := c.Request.Method + ":" + c.Request.URL.Path
		limiter := limiter.getLimiter(key)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "This endpoint is currently rate limited. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GlobalRateLimit 全局限流中间件
func GlobalRateLimit(rps int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(rps), rps*2)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "Server is busy. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
