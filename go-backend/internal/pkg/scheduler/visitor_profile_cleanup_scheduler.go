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

type VisitorProfileCleanupScheduler struct {
	visitorProfileService *service.VisitorProfileService
	interval              time.Duration
	cancel                context.CancelFunc
	done                  chan struct{}
	once                  sync.Once
}

func NewVisitorProfileCleanupScheduler(visitorProfileService *service.VisitorProfileService, cfg config.WorkerConfig) *VisitorProfileCleanupScheduler {
	intervalSeconds := cfg.VisitorProfileCleanupIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 86400
	}

	return &VisitorProfileCleanupScheduler{
		visitorProfileService: visitorProfileService,
		interval:              time.Duration(intervalSeconds) * time.Second,
		done:                  make(chan struct{}),
	}
}

func (s *VisitorProfileCleanupScheduler) Start(ctx context.Context) {
	if s == nil || s.visitorProfileService == nil {
		return
	}

	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		defer close(s.done)
		logger.Info("visitor profile cleanup scheduler started",
			zap.Duration("interval", s.interval),
		)

		s.cleanupOnce()

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-runCtx.Done():
				logger.Info("visitor profile cleanup scheduler stopped")
				return
			case <-ticker.C:
				s.cleanupOnce()
			}
		}
	}()
}

func (s *VisitorProfileCleanupScheduler) Stop() {
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

func (s *VisitorProfileCleanupScheduler) cleanupOnce() {
	result, err := s.visitorProfileService.CleanupExpiredProfiles(time.Now().UTC())
	if err != nil {
		logger.Error("visitor profile cleanup failed", zap.Error(err))
		return
	}
	if result.TotalChanged > 0 {
		logger.Info("visitor profile cleanup completed",
			zap.Int64("deleted_candidates", result.DeletedCandidates),
			zap.Int64("archived_anonymous", result.ArchivedAnonymous),
			zap.Int64("total_changed", result.TotalChanged),
		)
	}
}
