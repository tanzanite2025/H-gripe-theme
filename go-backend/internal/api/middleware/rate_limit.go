package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 简单的速率限制器
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     int           // 每分钟请求数
	window   time.Duration // 时间窗口
}

type Visitor struct {
	lastSeen time.Time
	count    int
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     requestsPerMinute,
		window:   time.Minute,
	}

	// 定期清理过期访客
	go rl.cleanupVisitors()

	return rl
}

// RateLimit 速率限制中间件
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		visitor, exists := rl.visitors[ip]

		if !exists {
			rl.visitors[ip] = &Visitor{
				lastSeen: time.Now(),
				count:    1,
			}
			rl.mu.Unlock()
			c.Next()
			return
		}

		// 检查时间窗口
		if time.Since(visitor.lastSeen) > rl.window {
			visitor.count = 1
			visitor.lastSeen = time.Now()
			rl.mu.Unlock()
			c.Next()
			return
		}

		// 检查请求次数
		if visitor.count >= rl.rate {
			rl.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		visitor.count++
		rl.mu.Unlock()
		c.Next()
	}
}

// cleanupVisitors 清理过期访客
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, visitor := range rl.visitors {
			if time.Since(visitor.lastSeen) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}
