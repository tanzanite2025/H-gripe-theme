package service

import (
	"context"
	"errors"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
)

func (s *YanwenAPIService) resolveYanwenGatewayCredentials(input YanwenAPIConfigInput) (yanwenGatewayCredentials, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.gatewayConfiguration == nil {
		return yanwenGatewayCredentials{}, errors.New("Yanwen configuration service is not configured")
	}
	return s.gatewayConfiguration.resolveYanwenGatewayCredentials(input)
}

// YanwenAPIService is retained only for package-internal legacy tests. It is
// excluded from production builds; handlers and workers receive one of the
// narrower Yanwen* service types instead.
type YanwenAPIService struct {
	// These legacy fields support old in-package literals and staged dependency
	// configuration. Runtime work is owned by the specialized services below.
	configs           *repository.YanwenAPIConfigRepository
	products          *repository.YanwenProductCatalogRepository
	countries         *repository.YanwenCountryCatalogRepository
	warehouses        *repository.YanwenWarehouseCatalogRepository
	orders            YanwenOrderFactReader
	publishedChannels *repository.YanwenPublishedChannelRepository
	waybills          *repository.YanwenWaybillRepository
	trackingSnapshots *repository.YanwenTrackingSnapshotRepository
	gateway           *YanwenGatewayClient

	gatewayConfiguration *YanwenGatewayConfigurationService
	officialCatalog      *YanwenOfficialCatalogService
	waybillOperations    *YanwenWaybillOperationsService
	trackingOperations   *YanwenTrackingOperationsService
	operationsOverview   *YanwenOperationsOverviewService
}

func NewYanwenAPIService(
	configs *repository.YanwenAPIConfigRepository,
	products *repository.YanwenProductCatalogRepository,
) *YanwenAPIService {
	service := &YanwenAPIService{configs: configs, products: products, gateway: NewYanwenGatewayClient()}
	service.initializeYanwenDomainServices()
	return service
}

// initializeYanwenDomainServices builds the specialized services once. The
// legacy repository fields remain only for old package-internal tests.
func (s *YanwenAPIService) initializeYanwenDomainServices() {
	if s == nil {
		return
	}
	if s.gateway == nil {
		s.gateway = NewYanwenGatewayClient()
	}
	if s.gatewayConfiguration == nil {
		s.gatewayConfiguration = NewYanwenGatewayConfigurationService(s.configs, s.gateway)
	}
	if s.officialCatalog == nil {
		s.officialCatalog = NewYanwenOfficialCatalogService(
			s.gatewayConfiguration,
			s.products,
			s.countries,
			s.warehouses,
			s.publishedChannels,
			s.gateway,
		)
	}
	if s.waybillOperations == nil {
		s.waybillOperations = NewYanwenWaybillOperationsService(
			s.gatewayConfiguration,
			s.products,
			s.countries,
			s.warehouses,
			s.orders,
			s.publishedChannels,
			s.waybills,
			s.trackingSnapshots,
			s.gateway,
		)
	}
	if s.trackingOperations == nil {
		s.trackingOperations = NewYanwenTrackingOperationsService(
			s.gatewayConfiguration,
			s.trackingSnapshots,
			s.gateway,
		)
	}
	if s.operationsOverview == nil {
		s.operationsOverview = NewYanwenOperationsOverviewService(
			s.gatewayConfiguration,
			s.waybills,
			s.trackingSnapshots,
		)
	}
}

// YanwenGatewayConfigurationService returns the configuration-only service.
func (s *YanwenAPIService) YanwenGatewayConfigurationService() *YanwenGatewayConfigurationService {
	s.initializeYanwenDomainServices()
	if s == nil {
		return nil
	}
	return s.gatewayConfiguration
}

// YanwenOfficialCatalogService returns the official-directory service.
func (s *YanwenAPIService) YanwenOfficialCatalogService() *YanwenOfficialCatalogService {
	s.initializeYanwenDomainServices()
	if s == nil {
		return nil
	}
	return s.officialCatalog
}

// YanwenWaybillOperationsService returns the waybill-only application service.
func (s *YanwenAPIService) YanwenWaybillOperationsService() *YanwenWaybillOperationsService {
	s.initializeYanwenDomainServices()
	if s == nil {
		return nil
	}
	return s.waybillOperations
}

// YanwenTrackingOperationsService returns the tracking-only application
// service used by the tracking handler and production polling worker.
func (s *YanwenAPIService) YanwenTrackingOperationsService() *YanwenTrackingOperationsService {
	s.initializeYanwenDomainServices()
	if s == nil {
		return nil
	}
	return s.trackingOperations
}

// YanwenOperationsOverviewService returns the read-only overview service.
func (s *YanwenAPIService) YanwenOperationsOverviewService() *YanwenOperationsOverviewService {
	s.initializeYanwenDomainServices()
	if s == nil {
		return nil
	}
	return s.operationsOverview
}

func (s *YanwenAPIService) yanwenGatewayClient() *YanwenGatewayClient {
	if s == nil {
		return nil
	}
	return s.gateway
}

func (s *YanwenAPIService) ConfigureYanwenCountryCatalogRepository(countries *repository.YanwenCountryCatalogRepository) {
	if s == nil {
		return
	}
	s.countries = countries
	s.initializeYanwenDomainServices()
	s.officialCatalog.countries = countries
	s.waybillOperations.countries = countries
}

func (s *YanwenAPIService) ConfigureYanwenWarehouseCatalogRepository(warehouses *repository.YanwenWarehouseCatalogRepository) {
	if s == nil {
		return
	}
	s.warehouses = warehouses
	s.initializeYanwenDomainServices()
	s.officialCatalog.warehouses = warehouses
	s.waybillOperations.warehouses = warehouses
}

func (s *YanwenAPIService) ConfigureYanwenWaybillRepositories(
	orders YanwenOrderFactReader,
	waybills *repository.YanwenWaybillRepository,
) {
	if s == nil {
		return
	}
	s.orders = orders
	s.waybills = waybills
	s.initializeYanwenDomainServices()
	s.waybillOperations.orders = orders
	s.waybillOperations.waybills = waybills
	s.operationsOverview.waybills = waybills
}

// ConfigureYanwenPublishedChannelRepository supplies the Yanwen-only curated
// channel store used by real waybill creation and catalog synchronization.
func (s *YanwenAPIService) ConfigureYanwenPublishedChannelRepository(channels *repository.YanwenPublishedChannelRepository) {
	if s == nil {
		return
	}
	s.publishedChannels = channels
	s.initializeYanwenDomainServices()
	s.officialCatalog.publishedChannels = channels
	s.waybillOperations.publishedChannels = channels
}

func (s *YanwenAPIService) GetYanwenAPIConfigurationView(environment string) (*shipping.YanwenAPIConfigView, error) {
	s.initializeYanwenDomainServices()
	if s != nil && s.gatewayConfiguration != nil {
		return s.gatewayConfiguration.GetYanwenAPIConfigurationView(environment)
	}
	return nil, errors.New("Yanwen configuration service is not configured")
}

func (s *YanwenAPIService) SaveYanwenAPIConfiguration(input YanwenAPIConfigInput) error {
	s.initializeYanwenDomainServices()
	if s != nil && s.gatewayConfiguration != nil {
		return s.gatewayConfiguration.SaveYanwenAPIConfiguration(input)
	}
	return errors.New("Yanwen configuration service is not configured")
}

func (s *YanwenAPIService) PingYanwenGateway(ctx context.Context, input YanwenAPIConfigInput) (YanwenPingResult, error) {
	s.initializeYanwenDomainServices()
	if s != nil && s.gatewayConfiguration != nil {
		return s.gatewayConfiguration.PingYanwenGateway(ctx, input)
	}
	return YanwenPingResult{}, errors.New("Yanwen configuration service is not configured")
}
