package service

import (
	"context"
	"errors"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
)

// These forwarding methods preserve the original YanwenAPIService contract for
// package-internal legacy tests while keeping implementation in narrow services.
func (s *YanwenAPIService) ListYanwenWaybills(environment, keyword, status, warehouseCode string) ([]shipping.YanwenWaybill, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.waybillOperations == nil {
		return nil, errors.New("Yanwen waybill service is not configured")
	}
	return s.waybillOperations.ListYanwenWaybills(environment, keyword, status, warehouseCode)
}

func (s *YanwenAPIService) SyncYanwenWaybillOfficialDetails(ctx context.Context, waybillID uint) (*shipping.YanwenWaybill, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.waybillOperations == nil {
		return nil, errors.New("Yanwen waybill service is not configured")
	}
	return s.waybillOperations.SyncYanwenWaybillOfficialDetails(ctx, waybillID)
}

func (s *YanwenAPIService) SyncYanwenWaybillsOfficialDetails(ctx context.Context, waybillIDs []uint) ([]shipping.YanwenWaybill, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.waybillOperations == nil {
		return nil, errors.New("Yanwen waybill service is not configured")
	}
	return s.waybillOperations.SyncYanwenWaybillsOfficialDetails(ctx, waybillIDs)
}

func (s *YanwenAPIService) DownloadYanwenWaybillLabel(ctx context.Context, waybillID uint) (YanwenWaybillLabelResult, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.waybillOperations == nil {
		return YanwenWaybillLabelResult{}, errors.New("Yanwen waybill service is not configured")
	}
	return s.waybillOperations.DownloadYanwenWaybillLabel(ctx, waybillID)
}

func (s *YanwenAPIService) DownloadYanwenWaybillLabelsArchive(ctx context.Context, waybillIDs []uint) (YanwenBatchWaybillLabelArchive, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.waybillOperations == nil {
		return YanwenBatchWaybillLabelArchive{}, errors.New("Yanwen waybill service is not configured")
	}
	return s.waybillOperations.DownloadYanwenWaybillLabelsArchive(ctx, waybillIDs)
}

func (s *YanwenAPIService) CancelYanwenWaybill(ctx context.Context, waybillID uint, note string) (YanwenWaybillCancellationResult, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.waybillOperations == nil {
		return YanwenWaybillCancellationResult{}, errors.New("Yanwen waybill service is not configured")
	}
	return s.waybillOperations.CancelYanwenWaybill(ctx, waybillID, note)
}

func (s *YanwenAPIService) CancelYanwenWaybills(ctx context.Context, waybillIDs []uint, note string) (YanwenBatchWaybillCancellationResult, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.waybillOperations == nil {
		return YanwenBatchWaybillCancellationResult{}, errors.New("Yanwen waybill service is not configured")
	}
	return s.waybillOperations.CancelYanwenWaybills(ctx, waybillIDs, note)
}

func (s *YanwenAPIService) CreateYanwenWaybills(ctx context.Context, requests []YanwenCreateWaybillInput) (YanwenBatchWaybillCreationResult, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.waybillOperations == nil {
		return YanwenBatchWaybillCreationResult{}, errors.New("Yanwen waybill service is not configured")
	}
	return s.waybillOperations.CreateYanwenWaybills(ctx, requests)
}

func (s *YanwenAPIService) CreateYanwenWaybill(ctx context.Context, input YanwenCreateWaybillInput) (*shipping.YanwenWaybill, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.waybillOperations == nil {
		return nil, errors.New("Yanwen waybill service is not configured")
	}
	return s.waybillOperations.CreateYanwenWaybill(ctx, input)
}

// Test and package-internal compatibility helpers retain the historical
// receiver without exposing the waybill implementation to other domains.
func (s *YanwenAPIService) verifyYanwenWaybillCustomsPreflight(ctx context.Context, credentials yanwenGatewayCredentials, countryCode string, request YanwenCreateOrderRequest) error {
	s.initializeYanwenDomainServices()
	if s == nil || s.waybillOperations == nil {
		return errors.New("Yanwen waybill service is not configured")
	}
	return s.waybillOperations.verifyYanwenWaybillCustomsPreflight(ctx, credentials, countryCode, request)
}

func (s *YanwenAPIService) resolveYanwenOrderCountry(environment, value string) (string, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.waybillOperations == nil {
		return "", errors.New("Yanwen waybill service is not configured")
	}
	return s.waybillOperations.resolveYanwenOrderCountry(environment, value)
}

func (s *YanwenAPIService) ListYanwenOfficialProducts(environment string) ([]shipping.YanwenProductCatalogEntry, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.officialCatalog == nil {
		return nil, errors.New("Yanwen catalog service is not configured")
	}
	return s.officialCatalog.ListYanwenOfficialProducts(environment)
}

func (s *YanwenAPIService) SyncYanwenOfficialProducts(ctx context.Context, input YanwenAPIConfigInput) (YanwenProductSyncSummary, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.officialCatalog == nil {
		return YanwenProductSyncSummary{}, errors.New("Yanwen catalog service is not configured")
	}
	return s.officialCatalog.SyncYanwenOfficialProducts(ctx, input)
}

func (s *YanwenAPIService) ListYanwenOfficialCountries(environment string) ([]shipping.YanwenCountryCatalogEntry, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.officialCatalog == nil {
		return nil, errors.New("Yanwen catalog service is not configured")
	}
	return s.officialCatalog.ListYanwenOfficialCountries(environment)
}

func (s *YanwenAPIService) SyncYanwenOfficialCountries(ctx context.Context, input YanwenAPIConfigInput) (YanwenCountrySyncSummary, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.officialCatalog == nil {
		return YanwenCountrySyncSummary{}, errors.New("Yanwen catalog service is not configured")
	}
	return s.officialCatalog.SyncYanwenOfficialCountries(ctx, input)
}

func (s *YanwenAPIService) ListYanwenOfficialWarehouses(environment string) ([]shipping.YanwenWarehouseCatalogEntry, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.officialCatalog == nil {
		return nil, errors.New("Yanwen catalog service is not configured")
	}
	return s.officialCatalog.ListYanwenOfficialWarehouses(environment)
}

func (s *YanwenAPIService) SyncYanwenOfficialWarehouses(ctx context.Context, input YanwenAPIConfigInput) (YanwenWarehouseSyncSummary, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.officialCatalog == nil {
		return YanwenWarehouseSyncSummary{}, errors.New("Yanwen catalog service is not configured")
	}
	return s.officialCatalog.SyncYanwenOfficialWarehouses(ctx, input)
}

func (s *YanwenAPIService) GetYanwenOperationsOverview(environment string) (*YanwenOperationsOverview, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.operationsOverview == nil {
		return nil, errors.New("Yanwen operations overview service is not configured")
	}
	return s.operationsOverview.GetYanwenOperationsOverview(environment)
}

func (s *YanwenAPIService) ConfigureYanwenTrackingSnapshotRepository(snapshotRepository *repository.YanwenTrackingSnapshotRepository) {
	if s == nil {
		return
	}
	s.trackingSnapshots = snapshotRepository
	s.initializeYanwenDomainServices()
	s.trackingOperations.snapshots = snapshotRepository
	s.waybillOperations.trackingSnapshots = snapshotRepository
	s.operationsOverview.trackingSnapshots = snapshotRepository
}

func (s *YanwenAPIService) QueryYanwenTracking(ctx context.Context, trackingNumbers []string) ([]YanwenTrackingResult, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.trackingOperations == nil {
		return nil, errors.New("Yanwen tracking service is not configured")
	}
	return s.trackingOperations.QueryYanwenTracking(ctx, trackingNumbers)
}

func (s *YanwenAPIService) yanwenTrackingSnapshotRepository() *repository.YanwenTrackingSnapshotRepository {
	if s == nil {
		return nil
	}
	s.initializeYanwenDomainServices()
	if s.trackingOperations == nil {
		return nil
	}
	return s.trackingOperations.yanwenTrackingSnapshotRepository()
}

func (s *YanwenAPIService) persistYanwenTrackingSnapshots(results []YanwenTrackingResult) error {
	s.initializeYanwenDomainServices()
	if s == nil || s.trackingOperations == nil {
		return errors.New("Yanwen tracking service is not configured")
	}
	return s.trackingOperations.persistYanwenTrackingSnapshots(results)
}

func (s *YanwenAPIService) resolveYanwenTrackingAuthorization() (string, error) {
	s.initializeYanwenDomainServices()
	if s == nil || s.gatewayConfiguration == nil {
		return "", errors.New("Yanwen configuration service is not configured")
	}
	return s.gatewayConfiguration.resolveYanwenTrackingAuthorization()
}
