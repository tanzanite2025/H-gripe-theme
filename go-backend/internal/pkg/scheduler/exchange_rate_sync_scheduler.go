package scheduler

import (
	"context"
	"errors"
	"sync"
	"time"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/service"

	"go.uber.org/zap"
)

type ExchangeRateSyncScheduler struct {
	exchangeRateService *service.ExchangeRateService
	interval            time.Duration
	cancel              context.CancelFunc
	done                chan struct{}
	once                sync.Once
}

func NewExchangeRateSyncScheduler(exchangeRateService *service.ExchangeRateService, cfg config.WorkerConfig) *ExchangeRateSyncScheduler {
	intervalSeconds := cfg.ExchangeRateSyncIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 86400
	}
	return &ExchangeRateSyncScheduler{
		exchangeRateService: exchangeRateService,
		interval:            time.Duration(intervalSeconds) * time.Second,
		done:                make(chan struct{}),
	}
}

func (s *ExchangeRateSyncScheduler) Start(ctx context.Context) {
	if s == nil || s.exchangeRateService == nil {
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		defer close(s.done)
		logger.Info("exchange rate sync scheduler started", zap.Duration("interval", s.interval))
		s.syncOnce()

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				logger.Info("exchange rate sync scheduler stopped")
				return
			case <-ticker.C:
				s.syncOnce()
			}
		}
	}()
}

func (s *ExchangeRateSyncScheduler) Stop() {
	if s == nil {
		return
	}
	s.once.Do(func() {
		if s.cancel == nil {
			return
		}
		s.cancel()
		if s.done != nil {
			<-s.done
		}
	})
}

func (s *ExchangeRateSyncScheduler) syncOnce() {
	result, err := s.exchangeRateService.Sync()
	if err != nil {
		if errors.Is(err, service.ErrExchangeRateDisabled) ||
			errors.Is(err, service.ErrExchangeRateNotConfigured) ||
			errors.Is(err, service.ErrExchangeRateSyncInProgress) {
			logger.Info("exchange rate sync skipped", zap.Error(err))
			return
		}
		logger.Error("exchange rate sync failed", zap.Error(err))
		return
	}
	logger.Info("exchange rate sync completed",
		zap.String("base_currency", result.Config.BaseCurrency),
		zap.Int("quote_targets", len(result.Config.QuoteCurrencies)),
		zap.Int("rates", len(result.Rates)),
		zap.Time("fetched_at", result.FetchedAt),
	)
	if result.DisplayPriceRefresh != nil {
		logger.Info("product display price snapshots refreshed",
			zap.Int("products_scanned", result.DisplayPriceRefresh.ProductsScanned),
			zap.Int("products_updated", result.DisplayPriceRefresh.ProductsUpdated),
			zap.Int("variants_scanned", result.DisplayPriceRefresh.VariantsScanned),
			zap.Int("variants_updated", result.DisplayPriceRefresh.VariantsUpdated),
			zap.Int("currency_mismatches", result.DisplayPriceRefresh.CurrencyMismatchCount),
		)
	}
	if result.ShippingDisplayPriceRefresh != nil {
		logger.Info("shipping display price snapshots refreshed",
			zap.Int("templates_scanned", result.ShippingDisplayPriceRefresh.TemplatesScanned),
			zap.Int("templates_updated", result.ShippingDisplayPriceRefresh.TemplatesUpdated),
			zap.Int("rules_scanned", result.ShippingDisplayPriceRefresh.RulesScanned),
			zap.Int("rules_updated", result.ShippingDisplayPriceRefresh.RulesUpdated),
			zap.Int("currency_mismatches", result.ShippingDisplayPriceRefresh.CurrencyMismatchCount),
		)
	}
}
