package service

import (
	"errors"

	"commerce-platform/internal/repository"
)

// YanwenWaybillOperationsService owns the Yanwen waybill ledger and the
// order-facing operations needed to create, synchronize, label, and cancel a
// Yanwen shipment. It does not own configuration, catalog synchronization, or
// generic order/shipping persistence.
type YanwenWaybillOperationsService struct {
	configuration     *YanwenGatewayConfigurationService
	products          *repository.YanwenProductCatalogRepository
	countries         *repository.YanwenCountryCatalogRepository
	warehouses        *repository.YanwenWarehouseCatalogRepository
	orders            YanwenOrderFactReader
	publishedChannels *repository.YanwenPublishedChannelRepository
	waybills          *repository.YanwenWaybillRepository
	trackingSnapshots *repository.YanwenTrackingSnapshotRepository
	gateway           *YanwenGatewayClient
}

func NewYanwenWaybillOperationsService(
	configuration *YanwenGatewayConfigurationService,
	products *repository.YanwenProductCatalogRepository,
	countries *repository.YanwenCountryCatalogRepository,
	warehouses *repository.YanwenWarehouseCatalogRepository,
	orders YanwenOrderFactReader,
	publishedChannels *repository.YanwenPublishedChannelRepository,
	waybills *repository.YanwenWaybillRepository,
	trackingSnapshots *repository.YanwenTrackingSnapshotRepository,
	gateway *YanwenGatewayClient,
) *YanwenWaybillOperationsService {
	if gateway == nil {
		gateway = NewYanwenGatewayClient()
	}
	return &YanwenWaybillOperationsService{
		configuration:     configuration,
		products:          products,
		countries:         countries,
		warehouses:        warehouses,
		orders:            orders,
		publishedChannels: publishedChannels,
		waybills:          waybills,
		trackingSnapshots: trackingSnapshots,
		gateway:           gateway,
	}
}

func (s *YanwenWaybillOperationsService) resolveYanwenGatewayCredentials(input YanwenAPIConfigInput) (yanwenGatewayCredentials, error) {
	if s == nil || s.configuration == nil {
		return yanwenGatewayCredentials{}, errors.New("Yanwen gateway configuration service is not configured")
	}
	return s.configuration.resolveYanwenGatewayCredentials(input)
}
