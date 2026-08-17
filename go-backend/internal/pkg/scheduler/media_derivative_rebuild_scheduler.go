package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/service"

	"go.uber.org/zap"
)

type MediaDerivativeRebuildScheduler struct {
	mediaService *service.MediaService
	interval     time.Duration
	batchLimit   int
	workerID     string
	cancel       context.CancelFunc
	done         chan struct{}
	start        sync.Once
	stop         sync.Once
}

func NewMediaDerivativeRebuildScheduler(
	mediaService *service.MediaService,
	cfg config.WorkerConfig,
) *MediaDerivativeRebuildScheduler {
	intervalSeconds := cfg.MediaDerivativeRebuildIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 5
	}
	batchLimit := cfg.MediaDerivativeRebuildBatchLimit
	if batchLimit <= 0 || batchLimit > 100 {
		batchLimit = 10
	}
	return &MediaDerivativeRebuildScheduler{
		mediaService: mediaService,
		interval:     time.Duration(intervalSeconds) * time.Second,
		batchLimit:   batchLimit,
		workerID:     fmt.Sprintf("media-derivative-rebuild-%d", time.Now().UTC().UnixNano()),
		done:         make(chan struct{}),
	}
}

func (s *MediaDerivativeRebuildScheduler) Start(ctx context.Context) {
	if s == nil {
		return
	}
	s.start.Do(func() {
		if s.mediaService == nil {
			close(s.done)
			return
		}
		runCtx, cancel := context.WithCancel(ctx)
		s.cancel = cancel
		go func() {
			defer close(s.done)
			logger.Info("media derivative rebuild scheduler started",
				zap.Duration("interval", s.interval),
				zap.Int("batch_limit", s.batchLimit),
			)
			s.runOnce(runCtx)
			ticker := time.NewTicker(s.interval)
			defer ticker.Stop()
			for {
				select {
				case <-runCtx.Done():
					logger.Info("media derivative rebuild scheduler stopped")
					return
				case <-ticker.C:
					s.runOnce(runCtx)
				}
			}
		}()
	})
}

func (s *MediaDerivativeRebuildScheduler) Stop() {
	if s == nil {
		return
	}
	s.stop.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		if s.done != nil {
			<-s.done
		}
	})
}

func (s *MediaDerivativeRebuildScheduler) runOnce(ctx context.Context) {
	result, err := s.mediaService.ProcessNextMediaDerivativeRebuild(ctx, s.workerID, s.batchLimit)
	if err != nil {
		logger.Error("media derivative rebuild batch failed", zap.Error(err))
		return
	}
	if result.Claimed {
		logger.Info("media derivative rebuild batch completed",
			zap.Uint("job_id", result.JobID),
			zap.Bool("completed", result.Completed),
			zap.Int("scanned_assets", result.ScannedAssets),
			zap.Int("generated_assets", result.GeneratedAssets),
			zap.Int("generated_derivatives", result.GeneratedDerivatives),
			zap.Int("failed_assets", result.FailedAssets),
			zap.Int64("updated_product_media_rows", result.UpdatedProductMediaRows),
		)
	}
}
