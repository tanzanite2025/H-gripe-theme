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

type BehaviorEventCleanupScheduler struct {
	behaviorEventService *service.BehaviorEventService
	interval             time.Duration
	cancel               context.CancelFunc
	done                 chan struct{}
	once                 sync.Once
}

func NewBehaviorEventCleanupScheduler(behaviorEventService *service.BehaviorEventService, cfg config.WorkerConfig) *BehaviorEventCleanupScheduler {
	intervalSeconds := cfg.BehaviorEventCleanupIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 86400
	}

	return &BehaviorEventCleanupScheduler{
		behaviorEventService: behaviorEventService,
		interval:             time.Duration(intervalSeconds) * time.Second,
		done:                 make(chan struct{}),
	}
}

func (s *BehaviorEventCleanupScheduler) Start(ctx context.Context) {
	if s == nil || s.behaviorEventService == nil {
		return
	}

	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		defer close(s.done)
		logger.Info("behavior event cleanup scheduler started",
			zap.Duration("interval", s.interval),
		)

		s.cleanupOnce()

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-runCtx.Done():
				logger.Info("behavior event cleanup scheduler stopped")
				return
			case <-ticker.C:
				s.cleanupOnce()
			}
		}
	}()
}

func (s *BehaviorEventCleanupScheduler) Stop() {
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

func (s *BehaviorEventCleanupScheduler) cleanupOnce() {
	result, err := s.behaviorEventService.CleanupExpiredEvents(time.Now().UTC())
	if err != nil {
		logger.Error("behavior event cleanup failed", zap.Error(err))
		return
	}
	if result.TotalDeleted > 0 {
		logger.Info("behavior event cleanup completed",
			zap.Int64("deleted_low_intent", result.DeletedLowIntent),
			zap.Int64("deleted_standard_intent", result.DeletedStandardIntent),
			zap.Int64("deleted_high_intent", result.DeletedHighIntent),
			zap.Int64("total_deleted", result.TotalDeleted),
		)
	}
}
