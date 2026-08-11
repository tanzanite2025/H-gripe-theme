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

type OutboxDispatchScheduler struct {
	outboxService *service.OutboxService
	interval      time.Duration
	batchLimit    int
	cancel        context.CancelFunc
	done          chan struct{}
	once          sync.Once
}

func NewOutboxDispatchScheduler(outboxService *service.OutboxService, cfg config.WorkerConfig) *OutboxDispatchScheduler {
	intervalSeconds := cfg.OutboxDispatchIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 10
	}
	batchLimit := cfg.OutboxDispatchBatchLimit
	if batchLimit <= 0 {
		batchLimit = 100
	}

	if outboxService != nil && cfg.OutboxDispatchLockTimeoutSeconds > 0 {
		outboxService.ConfigureLockTimeout(time.Duration(cfg.OutboxDispatchLockTimeoutSeconds) * time.Second)
	}

	return &OutboxDispatchScheduler{
		outboxService: outboxService,
		interval:      time.Duration(intervalSeconds) * time.Second,
		batchLimit:    batchLimit,
		done:          make(chan struct{}),
	}
}

func (s *OutboxDispatchScheduler) Start(ctx context.Context) {
	if s == nil || s.outboxService == nil {
		return
	}

	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		defer close(s.done)
		logger.Info("outbox dispatch scheduler started",
			zap.Duration("interval", s.interval),
			zap.Int("batch_limit", s.batchLimit),
		)

		s.dispatchOnce(runCtx)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-runCtx.Done():
				logger.Info("outbox dispatch scheduler stopped")
				return
			case <-ticker.C:
				s.dispatchOnce(runCtx)
			}
		}
	}()
}

func (s *OutboxDispatchScheduler) Stop() {
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

func (s *OutboxDispatchScheduler) dispatchOnce(ctx context.Context) {
	result, err := s.outboxService.ProcessPending(ctx, time.Now().UTC(), s.batchLimit)
	if err != nil {
		logger.Error("outbox dispatch failed", zap.Error(err))
		return
	}
	if result.Claimed > 0 {
		logger.Info("outbox dispatch completed",
			zap.Int("claimed", result.Claimed),
			zap.Int("processed", result.Processed),
			zap.Int("failed", result.Failed),
			zap.Int("dead_letter", result.DeadLetter),
			zap.Int("unhandled", result.Unhandled),
			zap.String("worker_id", result.WorkerID),
		)
	}
}
