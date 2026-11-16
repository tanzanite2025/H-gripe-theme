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
		logger.Info("hot data retention and archival scheduler started",
			zap.Duration("interval", s.interval),
			zap.Duration("site_quality_run_retention", service.SiteQualityRunRetention),
			zap.Duration("after_sales_event_archive_retention", service.AfterSalesEventArchiveRetention),
		)
		s.processExpiredDataOnce(runCtx)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				logger.Info("hot data archive scheduler stopped")
				return
			case <-ticker.C:
				s.processExpiredDataOnce(runCtx)
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

func (s *HotDataArchiveScheduler) processExpiredDataOnce(ctx context.Context) {
	result, err := s.archiveService.ProcessExpiredTerminalData(ctx, time.Now().UTC())
	if err != nil {
		logger.Error("hot data retention and archival failed", zap.Error(err))
		return
	}
	if result.SiteQualityRunsPurged == 0 &&
		result.ArchivedSiteQualityRunsPurged == 0 &&
		result.AfterSalesEventsArchived == 0 {
		return
	}
	logger.Info("hot data retention and archival completed",
		zap.Time("site_quality_run_cutoff", result.SiteQualityRunCutoff),
		zap.Int("site_quality_runs_purged", result.SiteQualityRunsPurged),
		zap.Int("archived_site_quality_runs_purged", result.ArchivedSiteQualityRunsPurged),
		zap.Time("after_sales_event_archive_cutoff", result.AfterSalesEventArchiveCutoff),
		zap.Int("after_sales_events_archived", result.AfterSalesEventsArchived),
	)
}
