package scheduler

import (
	"context"
	"sync"
	"time"

	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/logger"
	"tanzanite/internal/service"

	"go.uber.org/zap"
)

type TrackingScheduler struct {
	shippingService *service.ShippingService
	interval        time.Duration
	batchLimit      int
	cancel          context.CancelFunc
	done            chan struct{}
	once            sync.Once
}

func NewTrackingScheduler(shippingService *service.ShippingService, cfg config.WorkerConfig) *TrackingScheduler {
	intervalSeconds := cfg.TrackingPollingIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 300
	}

	batchLimit := cfg.TrackingPollingBatchLimit
	if batchLimit <= 0 {
		batchLimit = 20
	}

	return &TrackingScheduler{
		shippingService: shippingService,
		interval:        time.Duration(intervalSeconds) * time.Second,
		batchLimit:      batchLimit,
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
