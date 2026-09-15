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

// QuickBuyAbandonmentScheduler periodically advances stale quick-buy sessions
// to abandoned/expired states for recovery workflows and reporting.
type QuickBuyAbandonmentScheduler struct {
	service                  *service.QuickBuyService
	interval, abandonedAfter time.Duration
	cancel                   context.CancelFunc
	done                     chan struct{}
	once                     sync.Once
}

func NewQuickBuyAbandonmentScheduler(svc *service.QuickBuyService, cfg config.WorkerConfig) *QuickBuyAbandonmentScheduler {
	interval := time.Duration(cfg.QuickBuyAbandonmentIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = time.Hour
	}
	after := time.Duration(cfg.QuickBuyAbandonmentAfterSeconds) * time.Second
	if after <= 0 {
		after = 24 * time.Hour
	}
	return &QuickBuyAbandonmentScheduler{service: svc, interval: interval, abandonedAfter: after, done: make(chan struct{})}
}

func (s *QuickBuyAbandonmentScheduler) Start(ctx context.Context) {
	if s == nil || s.service == nil {
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	go func() {
		defer close(s.done)
		s.sweep()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				s.sweep()
			}
		}
	}()
}

func (s *QuickBuyAbandonmentScheduler) Stop() {
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

func (s *QuickBuyAbandonmentScheduler) sweep() {
	count, err := s.service.SweepAbandonedSessions(time.Now().UTC(), s.abandonedAfter)
	if err != nil {
		logger.Error("quick-buy abandonment sweep failed", zap.Error(err))
		return
	}
	if count > 0 {
		logger.Info("quick-buy sessions marked abandoned", zap.Int64("count", count))
	}
}
