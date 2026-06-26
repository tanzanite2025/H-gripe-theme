package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
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
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(rps),
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
		userID, exists := c.Get("userID")
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
