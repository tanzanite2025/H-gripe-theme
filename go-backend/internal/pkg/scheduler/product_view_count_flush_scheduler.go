package scheduler

import (
	"context"
	"time"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/service"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type ProductViewCountFlushScheduler struct {
	productService *service.ProductService
	redisClient    redis.UniversalClient
	interval       time.Duration
	batchLockTTL   time.Duration
	cancel         context.CancelFunc
	done           chan struct{}
}

var productViewCountFlushReleaseScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
end
return 0
`)

func NewProductViewCountFlushScheduler(productService *service.ProductService, redisClient redis.UniversalClient, cfg config.WorkerConfig) *ProductViewCountFlushScheduler {
	seconds := cfg.ProductViewCountFlushIntervalSeconds
	if seconds <= 0 {
		seconds = 10
	}
	return &ProductViewCountFlushScheduler{productService: productService, redisClient: redisClient, interval: time.Duration(seconds) * time.Second, batchLockTTL: time.Duration(seconds*2) * time.Second, done: make(chan struct{})}
}

func (s *ProductViewCountFlushScheduler) Start(ctx context.Context) {
	if s == nil || s.productService == nil {
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	go func() {
		defer close(s.done)
		logger.Info("product view count flush scheduler started", zap.Duration("interval", s.interval))
		s.flushOnce(runCtx)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				logger.Info("product view count flush scheduler stopped")
				return
			case <-ticker.C:
				s.flushOnce(runCtx)
			}
		}
	}()
}

func (s *ProductViewCountFlushScheduler) Stop() {
	if s == nil {
		return
	}
	if s.cancel == nil {
		return
	}
	s.cancel()
	if s.done != nil {
		<-s.done
	}
}

func (s *ProductViewCountFlushScheduler) flushOnce(ctx context.Context) {
	if s.redisClient != nil {
		token := uuid.NewString()
		ok, err := s.redisClient.SetNX(ctx, "scheduler:product-view-count-flush", token, s.batchLockTTL).Result()
		if err != nil || !ok {
			return
		}
		defer productViewCountFlushReleaseScript.Run(ctx, s.redisClient, []string{"scheduler:product-view-count-flush"}, token)
	}
	count, err := s.productService.FlushProductViewCounts(ctx)
	if err != nil {
		logger.Error("product view count flush failed", zap.Error(err))
		return
	}
	if count > 0 {
		logger.Info("product view counts flushed", zap.Int("products", count))
	}
}
