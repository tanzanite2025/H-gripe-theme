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

const siteQualityReconciliationInterval = 5 * time.Minute

// SiteQualityWorker performs low-frequency target reconciliation, plans due
// targets into durable jobs, and dispatches only jobs already in the queue.
type SiteQualityWorker struct {
	engine           *service.SiteQualityEngineService
	interval         time.Duration
	batch            int
	autoScan         bool
	lastReconciledAt time.Time
	cancel           context.CancelFunc
	done             chan struct{}
	start            sync.Once
	stop             sync.Once
}

func NewSiteQualityWorker(
	engine *service.SiteQualityEngineService,
	cfg config.WorkerConfig,
) *SiteQualityWorker {
	interval := cfg.SiteQualityDispatchIntervalSeconds
	if interval <= 0 {
		interval = 30
	}
	batch := cfg.SiteQualityBatchLimit
	if batch <= 0 {
		batch = 2
	}
	return &SiteQualityWorker{
		engine:   engine,
		interval: time.Duration(interval) * time.Second,
		batch:    batch,
		autoScan: cfg.SiteQualityAutoScanEnabled,
		done:     make(chan struct{}),
	}
}

func (s *SiteQualityWorker) Start(ctx context.Context) {
	if s == nil {
		return
	}
	s.start.Do(func() {
		if s.engine == nil {
			close(s.done)
			return
		}
		runCtx, cancel := context.WithCancel(ctx)
		s.cancel = cancel
		go func() {
			defer close(s.done)
			logger.Info("site quality job worker started",
				zap.Duration("interval", s.interval),
				zap.Int("batch_limit", s.batch),
				zap.Bool("auto_scan_enabled", s.autoScan),
			)
			s.runOnce(runCtx)
			ticker := time.NewTicker(s.interval)
			defer ticker.Stop()
			for {
				select {
				case <-runCtx.Done():
					logger.Info("site quality job worker stopped")
					return
				case <-ticker.C:
					s.runOnce(runCtx)
				}
			}
		}()
	})
}

func (s *SiteQualityWorker) Stop() {
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

func (s *SiteQualityWorker) runOnce(ctx context.Context) {
	now := time.Now().UTC()
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return
		}
	}
	if s.lastReconciledAt.IsZero() || now.Sub(s.lastReconciledAt) >= siteQualityReconciliationInterval {
		synced, err := s.engine.SyncTargetsFromRouteCatalog(now, s.batch*10)
		if err != nil {
			logger.Error("site quality target reconciliation failed", zap.Error(err))
		} else {
			s.lastReconciledAt = now
			logger.Info("site quality target reconciliation completed",
				zap.Int("synced", synced),
			)
		}
	}
	if s.autoScan {
		plan, err := s.engine.PlanDueWork(now, s.batch*10)
		if err != nil {
			logger.Error("site quality due-job planning failed", zap.Error(err))
		} else if plan.EnqueuedJobs > 0 || plan.SkippedJobs > 0 {
			logger.Info("site quality due-job planning completed",
				zap.Int("enqueued", plan.EnqueuedJobs),
				zap.Int("skipped", plan.SkippedJobs),
			)
		}
	}
	result, err := s.engine.ProcessReady(ctx, now, s.batch)
	if err != nil {
		logger.Error("site quality job dispatch failed", zap.Error(err))
		return
	}
	if result.Claimed > 0 {
		logger.Info("site quality job dispatch completed",
			zap.Int("claimed", result.Claimed),
			zap.Int("succeeded", result.Succeeded),
			zap.Int("failed", result.Failed),
			zap.Int("dead_letter", result.DeadLetter),
			zap.String("worker_id", result.WorkerID),
		)
	}
}
