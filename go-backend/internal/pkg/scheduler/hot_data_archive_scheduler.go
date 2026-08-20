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

type HotDataArchiveScheduler struct {
	archiveService *service.HotDataArchiveService
	interval       time.Duration
	cancel         context.CancelFunc
	done           chan struct{}
	once           sync.Once
}

func NewHotDataArchiveScheduler(
	archiveService *service.HotDataArchiveService,
	cfg config.WorkerConfig,
) *HotDataArchiveScheduler {
	intervalSeconds := cfg.HotDataArchiveIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 86400
	}
	return &HotDataArchiveScheduler{
		archiveService: archiveService,
		interval:       time.Duration(intervalSeconds) * time.Second,
		done:           make(chan struct{}),
	}
}

func (s *HotDataArchiveScheduler) Start(ctx context.Context) {
	if s == nil || s.archiveService == nil {
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		defer close(s.done)
		logger.Info("hot data archive scheduler started",
			zap.Duration("interval", s.interval),
			zap.Duration("retention", service.HotDataArchiveRetention),
		)
		s.archiveOnce(runCtx)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				logger.Info("hot data archive scheduler stopped")
				return
			case <-ticker.C:
				s.archiveOnce(runCtx)
			}
		}
	}()
}

func (s *HotDataArchiveScheduler) Stop() {
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

func (s *HotDataArchiveScheduler) archiveOnce(ctx context.Context) {
	result, err := s.archiveService.ArchiveExpiredTerminalData(ctx, time.Now().UTC())
	if err != nil {
		logger.Error("hot data archive failed", zap.Error(err))
		return
	}
	if result.SiteQualityRunsArchived == 0 && result.AfterSalesEventsArchived == 0 {
		return
	}
	logger.Info("hot data archive completed",
		zap.Int("site_quality_runs_archived", result.SiteQualityRunsArchived),
		zap.Int("after_sales_events_archived", result.AfterSalesEventsArchived),
		zap.Time("cutoff", result.Cutoff),
	)
}
