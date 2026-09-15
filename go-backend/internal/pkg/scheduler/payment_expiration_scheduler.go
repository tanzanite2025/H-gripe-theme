package scheduler

import (
	"context"
	"sync"
	"time"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/service"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type PaymentExpirationScheduler struct {
	orderService *service.OrderService
	interval     time.Duration
	ttl          time.Duration
	batchLimit   int
	lockTTL      time.Duration
	redisClient  redis.UniversalClient
	cancel       context.CancelFunc
	done         chan struct{}
	once         sync.Once
}

func NewPaymentExpirationScheduler(orderService *service.OrderService, cfg config.WorkerConfig, redisClients ...redis.UniversalClient) *PaymentExpirationScheduler {
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
	lockTTL := time.Duration(cfg.DistributedLockTTLSeconds) * time.Second
	if lockTTL < 2*time.Duration(intervalSeconds)*time.Second {
		lockTTL = 2 * time.Duration(intervalSeconds) * time.Second
	}
	var redisClient redis.UniversalClient
	if len(redisClients) > 0 {
		redisClient = redisClients[0]
	}

	return &PaymentExpirationScheduler{
		orderService: orderService,
		interval:     time.Duration(intervalSeconds) * time.Second,
		ttl:          time.Duration(ttlSeconds) * time.Second,
		batchLimit:   batchLimit,
		lockTTL:      lockTTL,
		redisClient:  redisClient,
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
			zap.Duration("lock_ttl", s.lockTTL),
			zap.Int("batch_limit", s.batchLimit),
		)

		s.cleanupOnceContext(runCtx)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-runCtx.Done():
				logger.Info("payment expiration scheduler stopped")
				return
			case <-ticker.C:
				s.cleanupOnceContext(runCtx)
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
	s.cleanupOnceContext(context.Background())
}

func (s *PaymentExpirationScheduler) cleanupOnceContext(ctx context.Context) {
	lock, acquired, err := acquireSchedulerLock(ctx, s.redisClient, "scheduler:payment-expiration", s.lockTTL)
	if err != nil {
		logger.Error("payment expiration scheduler lease unavailable", zap.Error(err))
		return
	}
	if !acquired {
		return
	}
	stopLeaseMaintenance := maintainSchedulerLock(lock, s.lockTTL)
	defer stopLeaseMaintenance()
	defer func() {
		if err := lock.release(context.Background()); err != nil {
			logger.Error("payment expiration scheduler lease release failed", zap.Error(err))
		}
	}()
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
