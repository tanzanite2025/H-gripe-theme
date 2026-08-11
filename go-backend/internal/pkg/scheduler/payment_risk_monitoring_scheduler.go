package scheduler

import (
	"context"
	"sync"
	"time"

	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/service"

	"go.uber.org/zap"
)

type PaymentRiskMonitoringScheduler struct {
	monitoring *service.PaymentRiskMonitoringService
	interval   time.Duration
	cancel     context.CancelFunc
	done       chan struct{}
	startOnce  sync.Once
	stopOnce   sync.Once
}

func NewPaymentRiskMonitoringScheduler(
	monitoring *service.PaymentRiskMonitoringService,
	cfg config.WorkerConfig,
) *PaymentRiskMonitoringScheduler {
	intervalSeconds := cfg.PaymentRiskMonitoringIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 3600
	}
	return &PaymentRiskMonitoringScheduler{
		monitoring: monitoring,
		interval:   time.Duration(intervalSeconds) * time.Second,
		done:       make(chan struct{}),
	}
}

func (s *PaymentRiskMonitoringScheduler) Start(ctx context.Context) {
	if s == nil {
		return
	}

	s.startOnce.Do(func() {
		if s.monitoring == nil || !s.monitoring.Enabled() {
			close(s.done)
			return
		}

		runCtx, cancel := context.WithCancel(ctx)
		s.cancel = cancel
		go func() {
			defer close(s.done)
			logger.Info("payment risk monitoring scheduler started", zap.Duration("interval", s.interval))

			s.recomputeOnce(runCtx)
			ticker := time.NewTicker(s.interval)
			defer ticker.Stop()

			for {
				select {
				case <-runCtx.Done():
					logger.Info("payment risk monitoring scheduler stopped")
					return
				case <-ticker.C:
					s.recomputeOnce(runCtx)
				}
			}
		}()
	})
}

func (s *PaymentRiskMonitoringScheduler) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		if s.done != nil {
			<-s.done
		}
	})
}

func (s *PaymentRiskMonitoringScheduler) recomputeOnce(ctx context.Context) {
	for _, provider := range []string{
		string(paymentdomain.PaymentRiskProviderStripe),
		string(paymentdomain.PaymentRiskProviderPayPal),
	} {
		report, err := s.monitoring.RecomputeProvider(ctx, provider, time.Now().UTC())
		if err != nil {
			logger.Error("payment risk monitoring recompute failed", zap.String("provider", provider), zap.Error(err))
			continue
		}
		if report != nil && report.Snapshot != nil {
			logger.Info("payment risk monitoring recomputed",
				zap.String("provider", provider),
				zap.String("level", string(report.Snapshot.Level)),
				zap.Int64("successful_payments", report.Snapshot.SuccessfulPaymentCount),
				zap.Int64("disputes", report.Snapshot.DisputeCount),
				zap.Int64("early_fraud_warnings", report.Snapshot.EarlyFraudWarningCount),
				zap.Int64("refunds", report.Snapshot.RefundCount),
			)
		}
	}
}
