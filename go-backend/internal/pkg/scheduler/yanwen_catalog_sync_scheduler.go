package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/service"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// YanwenCatalogSyncRunner is the narrow Yanwen-only catalog synchronization
// boundary used by the scheduler. It never reads or writes generic logistics
// catalog data.
type YanwenCatalogSyncRunner interface {
	SyncYanwenOfficialProducts(context.Context, service.YanwenAPIConfigInput) (service.YanwenProductSyncSummary, error)
	SyncYanwenOfficialCountries(context.Context, service.YanwenAPIConfigInput) (service.YanwenCountrySyncSummary, error)
	SyncYanwenOfficialWarehouses(context.Context, service.YanwenAPIConfigInput) (service.YanwenWarehouseSyncSummary, error)
}

// YanwenCatalogSyncEnvironmentResult contains the three official catalog
// results for one Yanwen environment. A failed catalog keeps its zero summary
// and records an operator-readable error while the other catalogs continue.
type YanwenCatalogSyncEnvironmentResult struct {
	Environment string
	Products    service.YanwenProductSyncSummary
	Countries   service.YanwenCountrySyncSummary
	Warehouses  service.YanwenWarehouseSyncSummary
	Errors      []string
}

// YanwenCatalogSyncRunResult contains one result for each environment visited
// by a scheduler run.
type YanwenCatalogSyncRunResult struct {
	Environments []YanwenCatalogSyncEnvironmentResult
}

// YanwenCatalogSyncScheduler periodically refreshes the official Yanwen
// product, country and warehouse catalogs for FAT and production. It has a
// dedicated distributed lease and never invokes the generic logistics sync.
type YanwenCatalogSyncScheduler struct {
	syncRunner  YanwenCatalogSyncRunner
	interval    time.Duration
	lockTTL     time.Duration
	redisClient redis.UniversalClient
	cancel      context.CancelFunc
	done        chan struct{}
	once        sync.Once
}

func NewYanwenCatalogSyncScheduler(
	syncRunner YanwenCatalogSyncRunner,
	cfg config.WorkerConfig,
	redisClients ...redis.UniversalClient,
) *YanwenCatalogSyncScheduler {
	intervalSeconds := cfg.YanwenCatalogSyncIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 86400
	}
	lockTTL := time.Duration(cfg.DistributedLockTTLSeconds) * time.Second
	if lockTTL < 2*time.Duration(intervalSeconds)*time.Second {
		lockTTL = 2 * time.Duration(intervalSeconds) * time.Second
	}
	var redisClient redis.UniversalClient
	if len(redisClients) > 0 {
		redisClient = redisClients[0]
	}
	return &YanwenCatalogSyncScheduler{
		syncRunner:  syncRunner,
		interval:    time.Duration(intervalSeconds) * time.Second,
		lockTTL:     lockTTL,
		redisClient: redisClient,
		done:        make(chan struct{}),
	}
}

func (s *YanwenCatalogSyncScheduler) Start(ctx context.Context) {
	if s == nil || s.syncRunner == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	go func() {
		defer close(s.done)
		logger.Info("Yanwen catalog sync scheduler started",
			zap.Duration("interval", s.interval),
			zap.Duration("lock_ttl", s.lockTTL),
		)
		s.syncOnce(runCtx)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				logger.Info("Yanwen catalog sync scheduler stopped")
				return
			case <-ticker.C:
				s.syncOnce(runCtx)
			}
		}
	}()
}

func (s *YanwenCatalogSyncScheduler) Stop() {
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

// RunOnce performs one complete FAT/production catalog refresh without taking
// the scheduler lease. Tests and controlled operational commands can use it;
// the background scheduler calls it only after acquiring its own lease.
func (s *YanwenCatalogSyncScheduler) RunOnce(ctx context.Context) YanwenCatalogSyncRunResult {
	result := YanwenCatalogSyncRunResult{
		Environments: make([]YanwenCatalogSyncEnvironmentResult, 0, 2),
	}
	if s == nil || s.syncRunner == nil {
		return result
	}
	if ctx == nil {
		ctx = context.Background()
	}
	for _, environment := range []string{
		shipping.YanwenPublishedChannelEnvironmentProduction,
		shipping.YanwenPublishedChannelEnvironmentFAT,
	} {
		environmentResult := YanwenCatalogSyncEnvironmentResult{Environment: environment}
		syncInput := service.YanwenAPIConfigInput{Environment: environment}
		if err := ctx.Err(); err != nil {
			environmentResult.Errors = append(environmentResult.Errors, err.Error())
			result.Environments = append(result.Environments, environmentResult)
			break
		}
		if summary, err := s.syncRunner.SyncYanwenOfficialProducts(ctx, syncInput); err != nil {
			environmentResult.Errors = append(environmentResult.Errors, fmt.Sprintf("products: %v", err))
		} else {
			environmentResult.Products = summary
		}
		if summary, err := s.syncRunner.SyncYanwenOfficialCountries(ctx, syncInput); err != nil {
			environmentResult.Errors = append(environmentResult.Errors, fmt.Sprintf("countries: %v", err))
		} else {
			environmentResult.Countries = summary
		}
		if summary, err := s.syncRunner.SyncYanwenOfficialWarehouses(ctx, syncInput); err != nil {
			environmentResult.Errors = append(environmentResult.Errors, fmt.Sprintf("warehouses: %v", err))
		} else {
			environmentResult.Warehouses = summary
		}
		result.Environments = append(result.Environments, environmentResult)
	}
	return result
}

func (s *YanwenCatalogSyncScheduler) syncOnce(ctx context.Context) {
	lock, acquired, err := acquireSchedulerLock(ctx, s.redisClient, "scheduler:yanwen-catalog-sync", s.lockTTL)
	if err != nil {
		logger.Error("Yanwen catalog sync scheduler lease unavailable", zap.Error(err))
		return
	}
	if !acquired {
		return
	}
	stopLeaseMaintenance := maintainSchedulerLock(lock, s.lockTTL)
	defer stopLeaseMaintenance()
	defer func() {
		if err := lock.release(context.Background()); err != nil {
			logger.Error("Yanwen catalog sync scheduler lease release failed", zap.Error(err))
		}
	}()

	result := s.RunOnce(ctx)
	for _, environmentResult := range result.Environments {
		if len(environmentResult.Errors) > 0 {
			logger.Error("Yanwen catalog sync failed for one environment",
				zap.String("environment", environmentResult.Environment),
				zap.Strings("errors", environmentResult.Errors),
			)
			continue
		}
		logger.Info("Yanwen catalog sync completed",
			zap.String("environment", environmentResult.Environment),
			zap.Int("products_scanned", environmentResult.Products.Scanned),
			zap.Int("products_added", environmentResult.Products.Added),
			zap.Int("products_updated", environmentResult.Products.Updated),
			zap.Int("countries_scanned", environmentResult.Countries.Scanned),
			zap.Int("countries_added", environmentResult.Countries.Added),
			zap.Int("countries_updated", environmentResult.Countries.Updated),
			zap.Int("warehouses_scanned", environmentResult.Warehouses.Scanned),
			zap.Int("warehouses_added", environmentResult.Warehouses.Added),
			zap.Int("warehouses_updated", environmentResult.Warehouses.Updated),
		)
	}
}
