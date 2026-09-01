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

type UGCShowcaseCleanupScheduler struct {
	UGCShowcaseService *service.UGCShowcaseService
	interval        time.Duration
	pendingTTL      time.Duration
	batchLimit      int
	cancel          context.CancelFunc
	done            chan struct{}
	once            sync.Once
}

func NewUGCShowcaseCleanupScheduler(UGCShowcaseService *service.UGCShowcaseService, cfg config.WorkerConfig) *UGCShowcaseCleanupScheduler {
	intervalSeconds := cfg.ShowcaseCleanupIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 86400
	}
	pendingTTLSeconds := cfg.ShowcasePendingTTLSeconds
	if pendingTTLSeconds <= 0 {
		pendingTTLSeconds = 30 * 24 * 60 * 60
	}
	batchLimit := cfg.ShowcaseCleanupBatchLimit
	if batchLimit <= 0 {
		batchLimit = 100
	}

	return &UGCShowcaseCleanupScheduler{
		UGCShowcaseService: UGCShowcaseService,
		interval:        time.Duration(intervalSeconds) * time.Second,
		pendingTTL:      time.Duration(pendingTTLSeconds) * time.Second,
		batchLimit:      batchLimit,
		done:            make(chan struct{}),
	}
}

func (s *UGCShowcaseCleanupScheduler) Start(ctx context.Context) {
	if s == nil || s.UGCShowcaseService == nil {
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		defer close(s.done)
		logger.Info("showcase cleanup scheduler started",
			zap.Duration("interval", s.interval),
			zap.Duration("pending_ttl", s.pendingTTL),
			zap.Int("batch_limit", s.batchLimit),
		)

		s.cleanupOnce(runCtx)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-runCtx.Done():
				logger.Info("showcase cleanup scheduler stopped")
				return
			case <-ticker.C:
				s.cleanupOnce(runCtx)
			}
		}
	}()
}

func (s *UGCShowcaseCleanupScheduler) Stop() {
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

func (s *UGCShowcaseCleanupScheduler) cleanupOnce(ctx context.Context) {
	result, err := s.UGCShowcaseService.CleanupExpiredPendingImages(
		ctx,
		time.Now().UTC(),
		s.pendingTTL,
		s.batchLimit,
	)
	if err != nil {
		logger.Error("showcase cleanup failed", zap.Error(err))
		return
	}
	if result.ScannedCandidates > 0 {
		logger.Info("showcase cleanup completed",
			zap.Int("scanned_candidates", result.ScannedCandidates),
			zap.Int("expired_pending_records", result.ExpiredPendingRecords),
			zap.Int("deleted_pending_images", result.DeletedPendingImages),
			zap.Int("retained_failed_images", result.RetainedFailedImages),
			zap.Int("updated_image_references", result.UpdatedImageReferences),
			zap.Time("cutoff", result.Cutoff),
		)
	}
}
