package middleware

import (
	"context"
	"net/http"
	"time"

	"tanzanite/internal/pkg/cache"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	limiter *redis_rate.Limiter
}

// NewRateLimiter 创建基于 Redis 的速率限制器
func NewRateLimiter(redisCache *cache.RedisCache) *RateLimiter {
	return &RateLimiter{
		limiter: redis_rate.NewLimiter(redisCache.Client()),
	}
}

// RateLimit 速率限制中间件
func (rl *RateLimiter) RateLimit(requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := "rate_limit:" + ip

		res, err := rl.limiter.Allow(context.Background(), key, redis_rate.Limit{
			Rate:   requests,
			Burst:  requests,
			Period: window,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limit error"})
			c.Abort()
			return
		}

		if res.Allowed == 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
