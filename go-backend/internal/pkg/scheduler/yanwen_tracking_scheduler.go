package scheduler

import (
	"context"
	"sync"
	"time"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/service"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// YanwenTrackingScheduler runs only the Yanwen production tracking workflow.
// It has its own Redis lease and never invokes the generic tracking scheduler.
type YanwenTrackingScheduler struct {
	pollingService *service.YanwenTrackingPollingService
	interval       time.Duration
	lockTTL        time.Duration
	redisClient    redis.UniversalClient
	cancel         context.CancelFunc
	done           chan struct{}
	once           sync.Once
}

func NewYanwenTrackingScheduler(
	pollingService *service.YanwenTrackingPollingService,
	cfg config.WorkerConfig,
	redisClients ...redis.UniversalClient,
) *YanwenTrackingScheduler {
	intervalSeconds := cfg.YanwenTrackingPollingIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 300
	}
	lockTTL := time.Duration(cfg.DistributedLockTTLSeconds) * time.Second
	if lockTTL < 2*time.Duration(intervalSeconds)*time.Second {
		lockTTL = 2 * time.Duration(intervalSeconds) * time.Second
	}
	var redisClient redis.UniversalClient
	if len(redisClients) > 0 {
		redisClient = redisClients[0]
	}
	return &YanwenTrackingScheduler{
		pollingService: pollingService,
		interval:       time.Duration(intervalSeconds) * time.Second,
		lockTTL:        lockTTL,
		redisClient:    redisClient,
		done:           make(chan struct{}),
	}
}

func (s *YanwenTrackingScheduler) Start(ctx context.Context) {
	if s == nil || s.pollingService == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	go func() {
		defer close(s.done)
		logger.Info("Yanwen tracking scheduler started",
			zap.Duration("interval", s.interval),
			zap.Duration("lock_ttl", s.lockTTL),
		)
		s.pollOnce(runCtx)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				logger.Info("Yanwen tracking scheduler stopped")
				return
			case <-ticker.C:
				s.pollOnce(runCtx)
			}
		}
	}()
}

func (s *YanwenTrackingScheduler) Stop() {
	if s == nil {
		return
	}
	s.once.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		if s.done != nil {
			<-s.done
		}
	})
}

func (s *YanwenTrackingScheduler) pollOnce(ctx context.Context) {
	lock, acquired, err := acquireSchedulerLock(ctx, s.redisClient, "scheduler:yanwen-tracking-polling", s.lockTTL)
	if err != nil {
		logger.Error("Yanwen tracking scheduler lease unavailable", zap.Error(err))
		return
	}
	if !acquired {
		return
	}
	stopLeaseMaintenance := maintainSchedulerLock(lock, s.lockTTL)
	defer stopLeaseMaintenance()
	defer func() {
		if err := lock.release(context.Background()); err != nil {
			logger.Error("Yanwen tracking scheduler lease release failed", zap.Error(err))
		}
	}()
	result, err := s.pollingService.PollDueYanwenTrackingTargets(ctx)
	if err != nil {
		logger.Error("Yanwen tracking scheduler poll failed",
			zap.Error(err),
			zap.Int("claimed", result.Claimed),
			zap.Int("synced", result.Synced),
			zap.Int("failed", result.Failed),
		)
		return
	}
	if result.Claimed > 0 {
		logger.Info("Yanwen tracking scheduler poll completed",
			zap.Int("claimed", result.Claimed),
			zap.Int("synced", result.Synced),
			zap.Int("terminal", result.Terminal),
			zap.Int("failed", result.Failed),
		)
	}
}
