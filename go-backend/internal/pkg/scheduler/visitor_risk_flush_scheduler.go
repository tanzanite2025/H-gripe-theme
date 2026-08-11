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

type VisitorRiskFlushScheduler struct {
	visitorRiskService *service.VisitorRiskService
	interval           time.Duration
	cancel             context.CancelFunc
	done               chan struct{}
	once               sync.Once
	lastCleanupAt      time.Time
}

func NewVisitorRiskFlushScheduler(visitorRiskService *service.VisitorRiskService, cfg config.VisitorRiskConfig) *VisitorRiskFlushScheduler {
	intervalSeconds := cfg.FlushIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 60
	}

	return &VisitorRiskFlushScheduler{
		visitorRiskService: visitorRiskService,
		interval:           time.Duration(intervalSeconds) * time.Second,
		done:               make(chan struct{}),
	}
}

func (s *VisitorRiskFlushScheduler) Start(ctx context.Context) {
	if s == nil || s.visitorRiskService == nil || !s.visitorRiskService.Enabled() {
		return
	}

	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		defer close(s.done)
		logger.Info("visitor risk flush scheduler started",
			zap.Duration("interval", s.interval),
		)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-runCtx.Done():
				s.flushOnce(context.Background())
				logger.Info("visitor risk flush scheduler stopped")
				return
			case <-ticker.C:
				s.flushOnce(runCtx)
			}
		}
	}()
}

func (s *VisitorRiskFlushScheduler) Stop() {
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

func (s *VisitorRiskFlushScheduler) flushOnce(ctx context.Context) {
	result, err := s.visitorRiskService.Flush(ctx)
	if err != nil {
		logger.Error("visitor risk flush failed", zap.Error(err))
		return
	}
	if result.FlushedFacts > 0 {
		logger.Info("visitor risk flush completed",
			zap.Int("flushed_facts", result.FlushedFacts),
		)
	}
	s.cleanupExpiredFacts()
}

func (s *VisitorRiskFlushScheduler) cleanupExpiredFacts() {
	now := time.Now().UTC()
	if !s.lastCleanupAt.IsZero() && now.Sub(s.lastCleanupAt) < 24*time.Hour {
		return
	}
	s.lastCleanupAt = now

	result, err := s.visitorRiskService.CleanupExpiredFacts(now)
	if err != nil {
		logger.Error("visitor risk cleanup failed", zap.Error(err))
		return
	}
	if result.DeletedFacts > 0 || result.DeletedDecisions > 0 {
		logger.Info("visitor risk cleanup completed",
			zap.Int64("deleted_facts", result.DeletedFacts),
			zap.Int64("deleted_decisions", result.DeletedDecisions),
			zap.Time("cutoff_day", result.CutoffDay),
		)
	}
}
