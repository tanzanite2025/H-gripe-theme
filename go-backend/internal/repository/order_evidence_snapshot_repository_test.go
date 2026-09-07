package repository

import (
	"crypto/sha256"
	"encoding/hex"
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

func TestOrderEvidenceSnapshotRepositoryCreatesFindsAndProtectsSnapshot(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&order.Order{}, &orderevidence.OrderEvidenceSnapshot{}))

	orderRecord := order.Order{
		OrderNumber: "TZ-2026-REPOSITORY-EVIDENCE",
		TotalAmount: 750,
		Currency:    "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	repo := NewOrderEvidenceSnapshotRepository(db)
	snapshot := repositoryTestEvidenceSnapshot(orderRecord.ID)
	require.NoError(t, repo.Create(snapshot))

	found, err := repo.FindByOrderID(orderRecord.ID)
	require.NoError(t, err)
	assert.Equal(t, snapshot.OrderID, found.OrderID)
	assert.Equal(t, snapshot.SnapshotSHA256, found.SnapshotSHA256)
	require.NoError(t, found.VerifyIntegrity())

	duplicate := repositoryTestEvidenceSnapshot(orderRecord.ID)
	require.Error(t, repo.Create(duplicate))
}

func TestOrderEvidenceSnapshotCannotBeUpdatedOrDeleted(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&order.Order{}, &orderevidence.OrderEvidenceSnapshot{}))

	orderRecord := order.Order{
		OrderNumber: "TZ-2026-REPOSITORY-IMMUTABLE",
		TotalAmount: 100,
		Currency:    "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	repo := NewOrderEvidenceSnapshotRepository(db)
	snapshot := repositoryTestEvidenceSnapshot(orderRecord.ID)
	require.NoError(t, repo.Create(snapshot))

	err = db.Model(&orderevidence.OrderEvidenceSnapshot{}).
		Where("id = ?", snapshot.ID).
		Update("is_high_value", true).Error
	require.ErrorIs(t, err, orderevidence.ErrOrderEvidenceSnapshotImmutable)

	err = db.Delete(&orderevidence.OrderEvidenceSnapshot{}, snapshot.ID).Error
	require.ErrorIs(t, err, orderevidence.ErrOrderEvidenceSnapshotImmutable)

	var count int64
	require.NoError(t, db.Model(&orderevidence.OrderEvidenceSnapshot{}).
		Where("id = ?", snapshot.ID).
		Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func repositoryTestEvidenceSnapshot(orderID uint) *orderevidence.OrderEvidenceSnapshot {
	data := datatypes.JSON([]byte(`{"items":[],"schema_version":1}`))
	hash := sha256.Sum256(data)
	return &orderevidence.OrderEvidenceSnapshot{
		OrderID:           orderID,
		SchemaVersion:     orderevidence.OrderEvidenceSnapshotSchemaVersion,
		ConfirmedAt:       time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC),
		Currency:          "USD",
		OrderTotalAmount:  100,
		OrderTotalUSD:     100,
		IsHighValue:       false,
		HasSpokeTensionQC: false,
		SnapshotData:      data,
		SnapshotSHA256:    hex.EncodeToString(hash[:]),
	}
}
