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

// ReferralLifecycleScheduler runs the bounded referral lifecycle sweep and
// settles matured records using idempotent reward transactions.
type ReferralLifecycleScheduler struct {
	referralService *service.ReferralService
	interval        time.Duration
	batchLimit      int
	lockTTL         time.Duration
	redisClient     redis.UniversalClient
	cancel          context.CancelFunc
	done            chan struct{}
	once            sync.Once
}

func NewReferralLifecycleScheduler(
	referralService *service.ReferralService,
	cfg config.WorkerConfig,
	redisClients ...redis.UniversalClient,
) *ReferralLifecycleScheduler {
	intervalSeconds := cfg.ReferralLifecycleIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 24 * 60 * 60
	}
	batchLimit := cfg.ReferralLifecycleBatchLimit
	if batchLimit <= 0 || batchLimit > 500 {
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
	return &ReferralLifecycleScheduler{
		referralService: referralService,
		interval:        time.Duration(intervalSeconds) * time.Second,
		batchLimit:      batchLimit,
		lockTTL:         lockTTL,
		redisClient:     redisClient,
		done:            make(chan struct{}),
	}
}

func (s *ReferralLifecycleScheduler) Start(ctx context.Context) {
	if s == nil || s.referralService == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	go func() {
		defer close(s.done)
		logger.Info("referral lifecycle scheduler started",
			zap.Duration("interval", s.interval),
			zap.Duration("lock_ttl", s.lockTTL),
			zap.Int("batch_limit", s.batchLimit),
		)
		s.scanOnce(runCtx)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				logger.Info("referral lifecycle scheduler stopped")
				return
			case <-ticker.C:
				s.scanOnce(runCtx)
			}
		}
	}()
}

func (s *ReferralLifecycleScheduler) Stop() {
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

func (s *ReferralLifecycleScheduler) scanOnce(ctx context.Context) {
	lock, acquired, err := acquireSchedulerLock(ctx, s.redisClient, "scheduler:referral-lifecycle", s.lockTTL)
	if err != nil {
		logger.Error("referral lifecycle scheduler lease unavailable", zap.Error(err))
		return
	}
	if !acquired {
		return
	}
	stopLeaseMaintenance := maintainSchedulerLock(lock, s.lockTTL)
	defer stopLeaseMaintenance()
	defer func() {
		if err := lock.release(context.Background()); err != nil {
			logger.Error("referral lifecycle scheduler lease release failed", zap.Error(err))
		}
	}()
	result, err := s.referralService.ScanLifecycle(ctx, time.Now().UTC(), s.batchLimit)
	if err != nil {
		logger.Error("referral lifecycle scan failed", zap.Error(err))
		return
	}
	if result.Expired > 0 || result.AdvancedToVesting > 0 || result.MaturedVestingCandidates > 0 || result.Settled > 0 {
		logger.Info("referral lifecycle scan completed",
			zap.Int("pending_scanned", result.PendingScanned),
			zap.Int("expired", result.Expired),
			zap.Int("ordered_scanned", result.OrderedScanned),
			zap.Int("advanced_to_vesting", result.AdvancedToVesting),
			zap.Int("matured_vesting_candidates", result.MaturedVestingCandidates),
			zap.Int("settled", result.Settled),
		)
	}
}
