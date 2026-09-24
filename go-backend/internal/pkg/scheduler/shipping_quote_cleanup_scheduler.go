package scheduler

import (
	"context"
	"sync"
	"time"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/service"
	"go.uber.org/zap"
)

// ShippingQuoteCleanupScheduler physically purges expired, short-lived
// checkout quote snapshots in bounded batches.
type ShippingQuoteCleanupScheduler struct {
	shippingService *service.ShippingService
	interval        time.Duration
	batchLimit      int
	cancel          context.CancelFunc
	done            chan struct{}
	once            sync.Once
}

func NewShippingQuoteCleanupScheduler(shippingService *service.ShippingService, cfg config.WorkerConfig) *ShippingQuoteCleanupScheduler {
	interval := time.Duration(cfg.ShippingQuoteCleanupIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = time.Hour
	}
	limit := cfg.ShippingQuoteCleanupBatchLimit
	if limit <= 0 {
		limit = 500
	}
	return &ShippingQuoteCleanupScheduler{
		shippingService: shippingService,
		interval:        interval,
		batchLimit:      limit,
		done:            make(chan struct{}),
	}
}

func (s *ShippingQuoteCleanupScheduler) Start(ctx context.Context) {
	if s == nil || s.shippingService == nil {
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	go func() {
		defer close(s.done)
		logger.Info("shipping quote cleanup scheduler started", zap.Duration("interval", s.interval), zap.Int("batch_limit", s.batchLimit))
		s.cleanupOnce()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				logger.Info("shipping quote cleanup scheduler stopped")
				return
			case <-ticker.C:
				s.cleanupOnce()
			}
		}
	}()
}

func (s *ShippingQuoteCleanupScheduler) Stop() {
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

// RunOnce executes one cleanup pass. It is intentionally exported for
// deployment smoke tests and operators running a one-shot maintenance job.
func (s *ShippingQuoteCleanupScheduler) RunOnce() (int64, error) {
	if s == nil || s.shippingService == nil {
		return 0, nil
	}
	return s.shippingService.CleanupExpiredShippingQuoteSnapshots(time.Now().UTC(), s.batchLimit)
}

func (s *ShippingQuoteCleanupScheduler) cleanupOnce() {
	deleted, err := s.RunOnce()
	if err != nil {
		logger.Error("shipping quote cleanup failed", zap.Error(err))
		return
	}
	if deleted > 0 {
		logger.Info("shipping quote cleanup completed", zap.Int64("deleted", deleted))
	}
}
