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

// CustomerServiceRetentionScheduler performs bounded, retryable physical
// cleanup. The retention service remains the authority for dependency holds,
// recovery-window checks and audit records.
type CustomerServiceRetentionScheduler struct {
	retentionService *service.CustomerServiceRetentionService
	interval         time.Duration
	batchLimit       int
	cancel           context.CancelFunc
	done             chan struct{}
	once             sync.Once
	started          bool
	stopped          bool
	mu               sync.Mutex
}

func NewCustomerServiceRetentionScheduler(retentionService *service.CustomerServiceRetentionService, cfg config.WorkerConfig) *CustomerServiceRetentionScheduler {
	if retentionService != nil {
		retentionService.ConfigureLegacyWorker(cfg.CustomerServiceRetentionEnabled)
	}
	intervalSeconds := cfg.CustomerServiceRetentionIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 86400
	}
	batchLimit := cfg.CustomerServiceRetentionBatchLimit
	if batchLimit <= 0 {
		batchLimit = 100
	}
	return &CustomerServiceRetentionScheduler{
		retentionService: retentionService,
		interval:         time.Duration(intervalSeconds) * time.Second,
		batchLimit:       batchLimit,
		done:             make(chan struct{}),
	}
}

func (s *CustomerServiceRetentionScheduler) Start(ctx context.Context) {
	if s == nil || s.retentionService == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	if s.started || s.stopped {
		s.mu.Unlock()
		cancel()
		return
	}
	s.started = true
	s.cancel = cancel
	s.mu.Unlock()
	go func() {
		defer close(s.done)
		logger.Info("customer-service retention scheduler started",
			zap.Duration("interval", s.interval), zap.Int("batch_limit", s.batchLimit))
		lastRun := time.Time{}
		for {
			pollInterval := time.Minute
			if cfg, err := s.retentionService.RuntimeConfig(); err == nil && cfg.IntervalSeconds > 0 && cfg.IntervalSeconds < 60 {
				pollInterval = time.Duration(cfg.IntervalSeconds) * time.Second
			}
			timer := time.NewTimer(pollInterval)
			select {
			case <-runCtx.Done():
				timer.Stop()
				logger.Info("customer-service retention scheduler stopped")
				return
			case <-timer.C:
				cfg, err := s.retentionService.RuntimeConfig()
				if err == nil && cfg.Enabled && (lastRun.IsZero() || time.Since(lastRun) >= time.Duration(cfg.IntervalSeconds)*time.Second) {
					s.cleanupOnce()
					lastRun = time.Now()
				}
			}
		}
	}()
}

func (s *CustomerServiceRetentionScheduler) Stop() {
	if s == nil {
		return
	}
	s.once.Do(func() {
		s.mu.Lock()
		started := s.started
		if !started {
			// Stop-before-Start is a valid lifecycle in tests and during
			// startup failure handling; do not wait on a goroutine that was
			// never launched.
			s.stopped = true
			s.mu.Unlock()
			return
		}
		cancel := s.cancel
		s.mu.Unlock()
		if cancel != nil {
			cancel()
		}
		if s.done != nil {
			<-s.done
		}
	})
}

// RunOnce executes one bounded pass synchronously. It is useful for
// operational commands and deterministic scheduler tests; Start continues to
// use the same path for periodic execution.
func (s *CustomerServiceRetentionScheduler) RunOnce() (service.CustomerServiceRetentionBatchResult, error) {
	if s == nil || s.retentionService == nil {
		return service.CustomerServiceRetentionBatchResult{}, nil
	}
	limit := s.batchLimit
	if cfg, err := s.retentionService.RuntimeConfig(); err == nil {
		if !cfg.Enabled {
			return service.CustomerServiceRetentionBatchResult{}, nil
		}
		if cfg.BatchLimit > 0 {
			limit = cfg.BatchLimit
		}
	}
	return s.retentionService.PurgeEligibleBatch(limit)
}

func (s *CustomerServiceRetentionScheduler) cleanupOnce() {
	result, err := s.RunOnce()
	if err != nil {
		logger.Error("customer-service retention cleanup failed", zap.Error(err))
		return
	}
	if result.Scanned > 0 || result.Errors > 0 {
		logger.Info("customer-service retention cleanup completed",
			zap.Int("scanned", result.Scanned), zap.Int("purged", result.Purged),
			zap.Int("blocked", result.Blocked), zap.Int("errors", result.Errors))
	}
}
