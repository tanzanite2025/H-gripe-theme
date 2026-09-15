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

type TrackingScheduler struct {
	shippingService *service.ShippingService
	interval        time.Duration
	batchLimit      int
	lockTTL         time.Duration
	redisClient     redis.UniversalClient
	cancel          context.CancelFunc
	done            chan struct{}
	once            sync.Once
}

func NewTrackingScheduler(shippingService *service.ShippingService, cfg config.WorkerConfig, redisClients ...redis.UniversalClient) *TrackingScheduler {
	intervalSeconds := cfg.TrackingPollingIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 300
	}

	batchLimit := cfg.TrackingPollingBatchLimit
	if batchLimit <= 0 {
		batchLimit = 20
	}
	lockTTL := time.Duration(cfg.DistributedLockTTLSeconds) * time.Second
	if lockTTL < 2*time.Duration(intervalSeconds)*time.Second {
		lockTTL = 2 * time.Duration(intervalSeconds) * time.Second
	}
	var redisClient redis.UniversalClient
	if len(redisClients) > 0 {
		redisClient = redisClients[0]
	}

	return &TrackingScheduler{
		shippingService: shippingService,
		interval:        time.Duration(intervalSeconds) * time.Second,
		batchLimit:      batchLimit,
		lockTTL:         lockTTL,
		redisClient:     redisClient,
		done:            make(chan struct{}),
	}
}

func (s *TrackingScheduler) Start(ctx context.Context) {
	if s == nil || s.shippingService == nil {
		return
	}

	s.shippingService.ConfigureTrackingPolling(true, s.interval, s.batchLimit)

	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		defer close(s.done)
		logger.Info("tracking scheduler started",
			zap.Duration("interval", s.interval),
			zap.Duration("lock_ttl", s.lockTTL),
			zap.Int("batch_limit", s.batchLimit),
		)

		s.syncOnce(runCtx)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-runCtx.Done():
				logger.Info("tracking scheduler stopped")
				return
			case <-ticker.C:
				s.syncOnce(runCtx)
			}
		}
	}()
}

func (s *TrackingScheduler) Stop() {
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

func (s *TrackingScheduler) syncOnce(ctx context.Context) {
	lock, acquired, err := acquireSchedulerLock(ctx, s.redisClient, "scheduler:tracking-polling", s.lockTTL)
	if err != nil {
		logger.Error("tracking scheduler lease unavailable", zap.Error(err))
		return
	}
	if !acquired {
		return
	}
	stopLeaseMaintenance := maintainSchedulerLock(lock, s.lockTTL)
	defer stopLeaseMaintenance()
	defer func() {
		if err := lock.release(context.Background()); err != nil {
			logger.Error("tracking scheduler lease release failed", zap.Error(err))
		}
	}()

	startedAt := time.Now()
	s.shippingService.MarkTrackingPollingStarted(startedAt)

	result, err := s.shippingService.SyncDueTrackingShipments(ctx, s.batchLimit)
	if err != nil {
		s.shippingService.MarkTrackingPollingFinished(startedAt, nil, err)
		logger.Error("tracking scheduler sync failed", zap.Error(err))
		return
	}
	s.shippingService.MarkTrackingPollingFinished(startedAt, result, nil)

	if result.Matched > 0 || result.Failed > 0 {
		logger.Info("tracking scheduler sync completed",
			zap.Int("matched", result.Matched),
			zap.Int("synced", result.Synced),
			zap.Int("failed", result.Failed),
		)
	}
}
