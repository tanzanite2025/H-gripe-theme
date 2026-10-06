package scheduler

import (
	"context"
	"errors"
	"testing"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/service"

	"github.com/stretchr/testify/require"
)

type fakeYanwenCatalogSyncRunner struct {
	calls             []string
	failProductsFor   string
	failCountriesFor  string
	failWarehousesFor string
}

func (f *fakeYanwenCatalogSyncRunner) SyncYanwenOfficialProducts(_ context.Context, input service.YanwenAPIConfigInput) (service.YanwenProductSyncSummary, error) {
	f.calls = append(f.calls, input.Environment+":products")
	if input.Environment == f.failProductsFor {
		return service.YanwenProductSyncSummary{}, errors.New("product sync failed")
	}
	return service.YanwenProductSyncSummary{Environment: input.Environment, Scanned: 1, Added: 1}, nil
}

func (f *fakeYanwenCatalogSyncRunner) SyncYanwenOfficialCountries(_ context.Context, input service.YanwenAPIConfigInput) (service.YanwenCountrySyncSummary, error) {
	f.calls = append(f.calls, input.Environment+":countries")
	if input.Environment == f.failCountriesFor {
		return service.YanwenCountrySyncSummary{}, errors.New("country sync failed")
	}
	return service.YanwenCountrySyncSummary{Environment: input.Environment, Scanned: 1, Updated: 1}, nil
}

func (f *fakeYanwenCatalogSyncRunner) SyncYanwenOfficialWarehouses(_ context.Context, input service.YanwenAPIConfigInput) (service.YanwenWarehouseSyncSummary, error) {
	f.calls = append(f.calls, input.Environment+":warehouses")
	if input.Environment == f.failWarehousesFor {
		return service.YanwenWarehouseSyncSummary{}, errors.New("warehouse sync failed")
	}
	return service.YanwenWarehouseSyncSummary{Environment: input.Environment, Scanned: 1, Updated: 1}, nil
}

func TestYanwenCatalogSyncSchedulerRunOnceRefreshesAllCatalogsForBothEnvironments(t *testing.T) {
	runner := &fakeYanwenCatalogSyncRunner{}
	scheduler := NewYanwenCatalogSyncScheduler(runner, newYanwenCatalogSyncTestWorkerConfig())

	result := scheduler.RunOnce(context.Background())

	require.Equal(t, []string{
		"production:products",
		"production:countries",
		"production:warehouses",
		"fat:products",
		"fat:countries",
		"fat:warehouses",
	}, runner.calls)
	require.Len(t, result.Environments, 2)
	require.Empty(t, result.Environments[0].Errors)
	require.Equal(t, 1, result.Environments[0].Products.Scanned)
	require.Empty(t, result.Environments[1].Errors)
}

func TestYanwenCatalogSyncSchedulerRunOnceContinuesOtherCatalogsAfterOneFailure(t *testing.T) {
	runner := &fakeYanwenCatalogSyncRunner{failProductsFor: "production"}
	scheduler := NewYanwenCatalogSyncScheduler(runner, newYanwenCatalogSyncTestWorkerConfig())

	result := scheduler.RunOnce(context.Background())

	require.Len(t, result.Environments, 2)
	require.Equal(t, []string{"production:products", "production:countries", "production:warehouses", "fat:products", "fat:countries", "fat:warehouses"}, runner.calls)
	require.Equal(t, []string{"products: product sync failed"}, result.Environments[0].Errors)
	require.Equal(t, 1, result.Environments[0].Countries.Scanned)
	require.Empty(t, result.Environments[1].Errors)
}

func newYanwenCatalogSyncTestWorkerConfig() config.WorkerConfig {
	return config.WorkerConfig{YanwenCatalogSyncIntervalSeconds: 86400}
}
