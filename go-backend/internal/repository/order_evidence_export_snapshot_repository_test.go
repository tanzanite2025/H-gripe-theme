package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOrderEvidenceExportSnapshotRepositoryReusesOneVersion(t *testing.T) {
	db, orderID, packageID := newOrderEvidenceExportSnapshotRepositoryFixture(t)
	repo := NewOrderEvidenceExportSnapshotRepository(db)

	first, err := repo.CreateOrGetForPackage(repositoryTestExportSnapshot(orderID, packageID))
	require.NoError(t, err)
	second, err := repo.CreateOrGetForPackage(repositoryTestExportSnapshot(orderID, packageID))
	require.NoError(t, err)

	require.NotNil(t, first)
	require.NotNil(t, second)
	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, 1, second.Version)

	found, err := repo.FindByPackageID(packageID)
	require.NoError(t, err)
	assert.Equal(t, first.ID, found.ID)
	assert.Equal(t, first.SnapshotSHA256, found.SnapshotSHA256)

	var count int64
	require.NoError(t, db.Model(&orderevidence.OrderEvidenceExportSnapshot{}).
		Where("evidence_package_id = ?", packageID).
		Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestOrderEvidenceExportSnapshotRepositoryConcurrentRequestsReuseOneVersion(t *testing.T) {
	db, orderID, packageID := newOrderEvidenceExportSnapshotRepositoryFixture(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	repo := NewOrderEvidenceExportSnapshotRepository(db)

	const workerCount = 6
	results := make(chan *orderevidence.OrderEvidenceExportSnapshot, workerCount)
	errs := make(chan error, workerCount)
	var waitGroup sync.WaitGroup
	for index := 0; index < workerCount; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			result, err := repo.CreateOrGetForPackage(repositoryTestExportSnapshot(orderID, packageID))
			if err != nil {
				errs <- err
				return
			}
			results <- result
		}()
	}
	waitGroup.Wait()
	close(results)
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}

	var firstID uint
	for result := range results {
		require.NotNil(t, result)
		if firstID == 0 {
			firstID = result.ID
			continue
		}
		assert.Equal(t, firstID, result.ID)
	}
	require.NotZero(t, firstID)

	var count int64
	require.NoError(t, db.Model(&orderevidence.OrderEvidenceExportSnapshot{}).
		Where("evidence_package_id = ?", packageID).
		Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestOrderEvidenceExportSnapshotRepositoryCannotUpdateOrDeleteSnapshot(t *testing.T) {
	db, orderID, packageID := newOrderEvidenceExportSnapshotRepositoryFixture(t)
	repo := NewOrderEvidenceExportSnapshotRepository(db)
	snapshot, err := repo.CreateOrGetForPackage(repositoryTestExportSnapshot(orderID, packageID))
	require.NoError(t, err)

	err = db.Model(&orderevidence.OrderEvidenceExportSnapshot{}).
		Where("id = ?", snapshot.ID).
		Update("created_by", 99).Error
	require.ErrorIs(t, err, orderevidence.ErrOrderEvidenceExportSnapshotImmutable)

	err = db.Delete(&orderevidence.OrderEvidenceExportSnapshot{}, snapshot.ID).Error
	require.ErrorIs(t, err, orderevidence.ErrOrderEvidenceExportSnapshotImmutable)

	var count int64
	require.NoError(t, db.Model(&orderevidence.OrderEvidenceExportSnapshot{}).
		Where("id = ?", snapshot.ID).
		Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func newOrderEvidenceExportSnapshotRepositoryFixture(t *testing.T) (*gorm.DB, uint, uint) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(
		&order.Order{},
		&orderevidence.OrderEvidencePackage{},
		&orderevidence.OrderEvidenceExportSnapshot{},
	))

	orderRecord := &order.Order{
		OrderNumber: "TZ-2026-EXPORT-REPOSITORY",
		TotalAmountMinor: 80000,
		Currency:    "USD",
	}
	require.NoError(t, db.Create(orderRecord).Error)
	lockedAt := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	pkg := &orderevidence.OrderEvidencePackage{
		OrderID:                    orderRecord.ID,
		SnapshotID:                 1,
		PackageVersion:             1,
		Status:                     orderevidence.PackageStatusLocked,
		OrderTotalUSDSnapshotMinor: 80000,
		SchemaVersion:              orderevidence.OrderEvidencePackageSchemaVersion,
		LockedAt:                   &lockedAt,
	}
	require.NoError(t, db.Create(pkg).Error)
	return db, orderRecord.ID, pkg.ID
}

func repositoryTestExportSnapshot(orderID, packageID uint) *orderevidence.OrderEvidenceExportSnapshot {
	data := datatypes.JSON([]byte(fmt.Sprintf(
		`{"export_type":"order_evidence_manifest","order_id":%d,"package_id":%d,"schema_version":1}`,
		orderID,
		packageID,
	)))
	hash := sha256.Sum256(data)
	return &orderevidence.OrderEvidenceExportSnapshot{
		OrderID:                orderID,
		EvidencePackageID:      packageID,
		EvidencePackageVersion: 1,
		Version:                1,
		Status:                 orderevidence.ExportSnapshotStatusLocked,
		SchemaVersion:          orderevidence.OrderEvidenceExportSnapshotSchemaVersion,
		LockedAt:               time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC),
		SnapshotData:           data,
		SnapshotSHA256:         hex.EncodeToString(hash[:]),
		CreatedBy:              7,
	}
}
