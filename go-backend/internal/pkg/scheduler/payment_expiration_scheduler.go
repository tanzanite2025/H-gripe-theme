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

type PaymentExpirationScheduler struct {
	orderService *service.OrderService
	interval     time.Duration
	ttl          time.Duration
	batchLimit   int
	cancel       context.CancelFunc
	done         chan struct{}
	once         sync.Once
}

func NewPaymentExpirationScheduler(orderService *service.OrderService, cfg config.WorkerConfig) *PaymentExpirationScheduler {
	intervalSeconds := cfg.PaymentExpirationIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 15 * 60
	}
	ttlSeconds := cfg.PaymentPendingTTLSeconds
	if ttlSeconds <= 0 {
		ttlSeconds = 30 * 60
	}
	batchLimit := cfg.PaymentExpirationBatchLimit
	if batchLimit <= 0 {
		batchLimit = 100
	}

	return &PaymentExpirationScheduler{
		orderService: orderService,
		interval:     time.Duration(intervalSeconds) * time.Second,
		ttl:          time.Duration(ttlSeconds) * time.Second,
		batchLimit:   batchLimit,
		done:         make(chan struct{}),
	}
}

func (s *PaymentExpirationScheduler) Start(ctx context.Context) {
	if s == nil || s.orderService == nil {
		return
	}

	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		defer close(s.done)
		logger.Info("payment expiration scheduler started",
			zap.Duration("interval", s.interval),
			zap.Duration("ttl", s.ttl),
			zap.Int("batch_limit", s.batchLimit),
		)

		s.cleanupOnce()

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-runCtx.Done():
				logger.Info("payment expiration scheduler stopped")
				return
			case <-ticker.C:
				s.cleanupOnce()
			}
		}
	}()
}

func (s *PaymentExpirationScheduler) Stop() {
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

func (s *PaymentExpirationScheduler) cleanupOnce() {
	result, err := s.orderService.ExpireStalePendingPayments(time.Now().UTC(), s.ttl, s.batchLimit)
	if err != nil {
		logger.Error("payment expiration cleanup failed", zap.Error(err))
		return
	}
	if result.ExpiredOrders > 0 {
		logger.Info("payment expiration cleanup completed",
			zap.Int("scanned_candidates", result.ScannedCandidates),
			zap.Int("expired_orders", result.ExpiredOrders),
			zap.Int("skipped_orders", result.SkippedOrders),
			zap.Int64("expired_open_transactions", result.ExpiredOpenTransactions),
			zap.Time("cutoff", result.Cutoff),
		)
	}
}
