package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/storage"
)

func TestPostgreSQLHistoricalMediaInventoryUsesEnforcedReadOnlyRepeatableSnapshotDuringConcurrentWrites(t *testing.T) {
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	require.NoError(t, env.db.AutoMigrate(&product.ProductBrand{}))
	asset := createVerifiedRegistrationForHistoricalMediaReadOnlyInventoryTest(t, env.db, "catalog/snapshot.png")
	first := product.ProductBrand{Name: "Before snapshot", Slug: "before-snapshot", LogoURL: asset.URL}
	require.NoError(t, env.db.Create(&first).Error)
	config := &storage.Config{Type: storage.StorageTypeLocal, BaseURL: "https://media.example.test"}
	audit := NewHistoricalMediaReadOnlyInventoryService(env.db, config, config, "https://shop.example.test")
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	type transactionSettings struct {
		ReadOnly  string
		Isolation string
	}
	observed := make(chan transactionSettings, 1)
	releaseChannel := make(chan struct{})
	var releaseOnce, pauseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
	defer release()
	callbackName := "observe_historical_inventory_read_only_snapshot_" + uuid.NewString()
	require.NoError(t, env.db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table != "media_assets" {
			return
		}
		pauseOnce.Do(func() {
			var settings transactionSettings
			if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT current_setting('transaction_read_only') AS read_only, current_setting('transaction_isolation') AS isolation").Scan(&settings).Error; err != nil {
				tx.AddError(err)
				return
			}
			observed <- settings
			select {
			case <-releaseChannel:
			case <-ctx.Done():
				tx.AddError(ctx.Err())
			}
		})
	}))
	defer env.db.Callback().Query().Remove(callbackName)
	type auditResult struct {
		report HistoricalMediaReadOnlyInventoryReport
		err    error
	}
	result := make(chan auditResult, 1)
	go func() {
		report, err := audit.AuditHistoricalRegistrationsAndBusinessReferences(ctx, HistoricalMediaReadOnlyInventoryOptions{BatchSize: 1, MaxRecordsPerSource: 100, MaxIssues: 100})
		result <- auditResult{report, err}
	}()
	select {
	case settings := <-observed:
		require.Equal(t, "on", settings.ReadOnly)
		require.Equal(t, "repeatable read", settings.Isolation)
	case <-ctx.Done():
		t.Fatal("audit did not enter its read-only database snapshot")
	}
	require.NoError(t, env.db.WithContext(ctx).Create(&product.ProductBrand{Name: "After snapshot", Slug: "after-snapshot", LogoURL: "/uploads/newly-added.png"}).Error)
	release()
	var outcome auditResult
	select {
	case outcome = <-result:
	case <-ctx.Done():
		t.Fatal("historical media audit timed out")
	}
	require.NoError(t, outcome.err)
	require.Equal(t, "repeatable_read_read_only", outcome.report.SnapshotMode)
	var brandsScanned int
	for _, source := range outcome.report.Sources {
		if source.Table == "product_brands" {
			brandsScanned += source.RowsScanned
		}
	}
	require.Equal(t, 1, brandsScanned, "concurrent commits cannot change the audit snapshot")
	require.Zero(t, outcome.report.IssueCounts["business_managed_url_not_registered"])
	var currentBrands int64
	require.NoError(t, env.db.Model(&product.ProductBrand{}).Count(&currentBrands).Error)
	require.EqualValues(t, 2, currentBrands, "the audit never blocks or overwrites the concurrent writer")
}
