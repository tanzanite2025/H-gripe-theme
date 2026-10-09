package scheduler

import (
	"context"
	"sync"
	"time"

	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/service"

	"go.uber.org/zap"
)

// TaxRateSourceSnapshotSyncScheduler periodically checks whether the
// administrator-enabled reference-data snapshot is due for refresh.
type TaxRateSourceSnapshotSyncScheduler struct {
	snapshotService *service.TaxRateSourceSnapshotService
	interval        time.Duration
	cancel          context.CancelFunc
	done            chan struct{}
	startOnce       sync.Once
	stopOnce        sync.Once
}

func NewTaxRateSourceSnapshotSyncScheduler(
	snapshotService *service.TaxRateSourceSnapshotService,
) *TaxRateSourceSnapshotSyncScheduler {
	return &TaxRateSourceSnapshotSyncScheduler{
		snapshotService: snapshotService,
		interval:        time.Hour,
		done:            make(chan struct{}),
	}
}

func (s *TaxRateSourceSnapshotSyncScheduler) Start(ctx context.Context) {
	if s == nil || s.snapshotService == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	s.startOnce.Do(func() {
		runCtx, cancel := context.WithCancel(ctx)
		s.cancel = cancel
		go func() {
			defer close(s.done)
			logger.Info("tax rate source snapshot sync scheduler started", zap.Duration("interval", s.interval))
			s.syncIfDue(runCtx)
			ticker := time.NewTicker(s.interval)
			defer ticker.Stop()
			for {
				select {
				case <-runCtx.Done():
					logger.Info("tax rate source snapshot sync scheduler stopped")
					return
				case <-ticker.C:
					s.syncIfDue(runCtx)
				}
			}
		}()
	})
}

func (s *TaxRateSourceSnapshotSyncScheduler) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		if s.cancel == nil {
			return
		}
		s.cancel()
		<-s.done
	})
}

func (s *TaxRateSourceSnapshotSyncScheduler) syncIfDue(ctx context.Context) {
	result, synced, err := s.snapshotService.SyncIfDue(ctx)
	if err != nil {
		logger.Error("tax rate source snapshot sync failed", zap.Error(err))
		return
	}
	if !synced || result == nil || result.Snapshot == nil {
		return
	}
	logger.Info("tax rate source snapshot sync completed",
		zap.String("provider", result.Snapshot.ProviderCode),
		zap.String("version", result.Snapshot.Version),
		zap.Bool("changed", result.Changed),
		zap.Int("countries", result.Snapshot.CountryCount),
		zap.Int("rates", result.Snapshot.RateCount),
		zap.Time("checked_at", result.CheckedAt),
	)
}
