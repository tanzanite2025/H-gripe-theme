package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
)

// YanwenOfficialCatalogService owns only the official Yanwen product, country,
// and warehouse directories plus the curated-channel disablement that follows
// an official product removal. It never reads orders or generic logistics data.
type YanwenOfficialCatalogService struct {
	configuration     *YanwenGatewayConfigurationService
	products          *repository.YanwenProductCatalogRepository
	countries         *repository.YanwenCountryCatalogRepository
	warehouses        *repository.YanwenWarehouseCatalogRepository
	publishedChannels *repository.YanwenPublishedChannelRepository
	gateway           *YanwenGatewayClient
}

func NewYanwenOfficialCatalogService(
	configuration *YanwenGatewayConfigurationService,
	products *repository.YanwenProductCatalogRepository,
	countries *repository.YanwenCountryCatalogRepository,
	warehouses *repository.YanwenWarehouseCatalogRepository,
	publishedChannels *repository.YanwenPublishedChannelRepository,
	gateway *YanwenGatewayClient,
) *YanwenOfficialCatalogService {
	if gateway == nil {
		gateway = NewYanwenGatewayClient()
	}
	return &YanwenOfficialCatalogService{
		configuration:     configuration,
		products:          products,
		countries:         countries,
		warehouses:        warehouses,
		publishedChannels: publishedChannels,
		gateway:           gateway,
	}
}

func (s *YanwenOfficialCatalogService) ListYanwenOfficialProducts(environment string) ([]shipping.YanwenProductCatalogEntry, error) {
	environment = normalizeYanwenEnvironment(environment)
	if !isSupportedYanwenEnvironment(environment) {
		return nil, errors.New("Yanwen environment must be fat or production")
	}
	if s == nil || s.products == nil {
		return nil, errors.New("Yanwen product catalog repository is not configured")
	}
	return s.products.FindYanwenProductCatalogEntriesByEnvironment(environment)
}

func (s *YanwenOfficialCatalogService) SyncYanwenOfficialProducts(ctx context.Context, input YanwenAPIConfigInput) (YanwenProductSyncSummary, error) {
	input.Environment = normalizeYanwenEnvironment(input.Environment)
	if !isSupportedYanwenEnvironment(input.Environment) {
		return YanwenProductSyncSummary{}, errors.New("Yanwen environment must be fat or production")
	}
	if s == nil || s.products == nil {
		return YanwenProductSyncSummary{}, errors.New("Yanwen product catalog repository is not configured")
	}
	if s.configuration == nil {
		return YanwenProductSyncSummary{}, errors.New("Yanwen gateway configuration service is not configured")
	}
	credentials, err := s.configuration.resolveYanwenGatewayCredentials(input)
	if err != nil {
		return YanwenProductSyncSummary{}, err
	}
	if s.gateway == nil {
		return YanwenProductSyncSummary{}, errors.New("Yanwen gateway client is not configured")
	}
	products, err := s.gateway.FetchYanwenOfficialProducts(ctx, credentials)
	if err != nil {
		return YanwenProductSyncSummary{}, err
	}
	syncedAt := time.Now().UTC()
	stats, err := s.products.UpsertYanwenProductCatalogEntries(credentials.environment, products, syncedAt)
	if err != nil {
		return YanwenProductSyncSummary{}, fmt.Errorf("save Yanwen product catalog: %w", err)
	}
	var channelsDisabled int64
	if s.publishedChannels != nil {
		productCodes := make([]string, 0, len(products))
		for _, product := range products {
			productCodes = append(productCodes, product.ProductID)
		}
		channelsDisabled, err = s.publishedChannels.DisableMissingOfficialProducts(
			credentials.environment,
			productCodes,
			syncedAt,
		)
		if err != nil {
			return YanwenProductSyncSummary{}, fmt.Errorf("disable Yanwen channels for missing products: %w", err)
		}
	}
	return YanwenProductSyncSummary{
		Environment:      credentials.environment,
		Scanned:          stats.Scanned,
		Added:            stats.Added,
		Updated:          stats.Updated,
		ChannelsDisabled: channelsDisabled,
		SyncedAt:         syncedAt,
	}, nil
}

func (s *YanwenOfficialCatalogService) ListYanwenOfficialCountries(environment string) ([]shipping.YanwenCountryCatalogEntry, error) {
	environment = normalizeYanwenEnvironment(environment)
	if !isSupportedYanwenEnvironment(environment) {
		return nil, errors.New("Yanwen environment must be fat or production")
	}
	if s == nil || s.countries == nil {
		return nil, errors.New("Yanwen country catalog repository is not configured")
	}
	return s.countries.FindYanwenCountryCatalogEntriesByEnvironment(environment)
}

func (s *YanwenOfficialCatalogService) SyncYanwenOfficialCountries(ctx context.Context, input YanwenAPIConfigInput) (YanwenCountrySyncSummary, error) {
	input.Environment = normalizeYanwenEnvironment(input.Environment)
	if !isSupportedYanwenEnvironment(input.Environment) {
		return YanwenCountrySyncSummary{}, errors.New("Yanwen environment must be fat or production")
	}
	if s == nil || s.countries == nil {
		return YanwenCountrySyncSummary{}, errors.New("Yanwen country catalog repository is not configured")
	}
	if s.configuration == nil {
		return YanwenCountrySyncSummary{}, errors.New("Yanwen gateway configuration service is not configured")
	}
	credentials, err := s.configuration.resolveYanwenGatewayCredentials(input)
	if err != nil {
		return YanwenCountrySyncSummary{}, err
	}
	if s.gateway == nil {
		return YanwenCountrySyncSummary{}, errors.New("Yanwen gateway client is not configured")
	}
	countries, err := s.gateway.FetchYanwenOfficialCountries(ctx, credentials)
	if err != nil {
		return YanwenCountrySyncSummary{}, err
	}
	syncedAt := time.Now().UTC()
	stats, err := s.countries.UpsertYanwenCountryCatalogEntries(credentials.environment, countries, syncedAt)
	if err != nil {
		return YanwenCountrySyncSummary{}, fmt.Errorf("save Yanwen country catalog: %w", err)
	}
	return YanwenCountrySyncSummary{
		Environment: credentials.environment,
		Scanned:     stats.Scanned,
		Added:       stats.Added,
		Updated:     stats.Updated,
		SyncedAt:    syncedAt,
	}, nil
}

func (s *YanwenOfficialCatalogService) ListYanwenOfficialWarehouses(environment string) ([]shipping.YanwenWarehouseCatalogEntry, error) {
	environment = normalizeYanwenEnvironment(environment)
	if !isSupportedYanwenEnvironment(environment) {
		return nil, errors.New("Yanwen environment must be fat or production")
	}
	if s == nil || s.warehouses == nil {
		return nil, errors.New("Yanwen warehouse catalog repository is not configured")
	}
	return s.warehouses.FindYanwenWarehouseCatalogEntriesByEnvironment(environment)
}

func (s *YanwenOfficialCatalogService) SyncYanwenOfficialWarehouses(ctx context.Context, input YanwenAPIConfigInput) (YanwenWarehouseSyncSummary, error) {
	input.Environment = normalizeYanwenEnvironment(input.Environment)
	if !isSupportedYanwenEnvironment(input.Environment) {
		return YanwenWarehouseSyncSummary{}, errors.New("Yanwen environment must be fat or production")
	}
	if s == nil || s.warehouses == nil {
		return YanwenWarehouseSyncSummary{}, errors.New("Yanwen warehouse catalog repository is not configured")
	}
	if s.configuration == nil {
		return YanwenWarehouseSyncSummary{}, errors.New("Yanwen gateway configuration service is not configured")
	}
	credentials, err := s.configuration.resolveYanwenGatewayCredentials(input)
	if err != nil {
		return YanwenWarehouseSyncSummary{}, err
	}
	if s.gateway == nil {
		return YanwenWarehouseSyncSummary{}, errors.New("Yanwen gateway client is not configured")
	}
	warehouses, err := s.gateway.FetchYanwenOfficialWarehouses(ctx, credentials)
	if err != nil {
		return YanwenWarehouseSyncSummary{}, err
	}
	syncedAt := time.Now().UTC()
	stats, err := s.warehouses.UpsertYanwenWarehouseCatalogEntries(credentials.environment, warehouses, syncedAt)
	if err != nil {
		return YanwenWarehouseSyncSummary{}, fmt.Errorf("save Yanwen warehouse catalog: %w", err)
	}
	return YanwenWarehouseSyncSummary{
		Environment: credentials.environment,
		Scanned:     stats.Scanned,
		Added:       stats.Added,
		Updated:     stats.Updated,
		SyncedAt:    syncedAt,
	}, nil
}
