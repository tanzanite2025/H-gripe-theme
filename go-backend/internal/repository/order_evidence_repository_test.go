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

func TestOrderEvidenceRepositoryCreatesAndReadsPackageGraph(t *testing.T) {
	db := newOrderEvidenceRepositoryTestDB(t)
	orderRecord, snapshot := seedOrderEvidenceRepositoryOrderAndSnapshot(t, db)
	repo := NewOrderEvidenceRepository(db)

	itemData := datatypes.JSON([]byte(`{"snapshot_id":1}`))
	itemHash := sha256.Sum256(itemData)
	orderItemID := uint(1)
	items := []orderevidence.OrderEvidenceItem{
		{
			OrderID:        orderRecord.ID,
			OrderItemID:    &orderItemID,
			ItemType:       orderevidence.EvidenceItemTypeConfigurationConfirmation,
			Status:         orderevidence.EvidenceItemStatusComplete,
			RequiredReason: orderevidence.EvidenceRequiredReasonBase,
			SnapshotID:     &snapshot.ID,
			DataJSON:       itemData,
			CapturedAt:     timePtrForRepositoryEvidence(time.Now().UTC()),
			ContentSHA256:  hex.EncodeToString(itemHash[:]),
		},
		{
			OrderID:        orderRecord.ID,
			OrderItemID:    &orderItemID,
			ItemType:       orderevidence.EvidenceItemTypeProductIdentity,
			Status:         orderevidence.EvidenceItemStatusMissing,
			RequiredReason: orderevidence.EvidenceRequiredReasonBase,
		},
		{
			OrderID:        orderRecord.ID,
			ItemType:       orderevidence.EvidenceItemTypeSignedPOD,
			Status:         orderevidence.EvidenceItemStatusMissing,
			RequiredReason: orderevidence.EvidenceRequiredReasonBase,
		},
	}
	pkg := &orderevidence.OrderEvidencePackage{
		OrderID:                    orderRecord.ID,
		SnapshotID:                 snapshot.ID,
		PackageVersion:             1,
		Status:                     orderevidence.PackageStatusIncomplete,
		OrderTotalUSDSnapshotMinor: 10000,
		SchemaVersion:              orderevidence.OrderEvidencePackageSchemaVersion,
	}
	require.NoError(t, repo.CreatePackageWithItems(pkg, items))
	require.NotZero(t, pkg.ID)

	foundPackage, err := repo.FindLatestPackageByOrderID(orderRecord.ID)
	require.NoError(t, err)
	assert.Equal(t, pkg.ID, foundPackage.ID)
	assert.Equal(t, snapshot.ID, foundPackage.SnapshotID)

	foundItems, err := repo.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	require.Len(t, foundItems, 3)

	attachment := &orderevidence.OrderEvidenceAttachment{
		EvidenceItemID:   foundItems[0].ID,
		StorageKey:       "order-evidence/1/config.json",
		OriginalFilename: "config.json",
		MimeType:         "application/json",
		SizeBytes:        20,
		SHA256:           "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	require.NoError(t, repo.CreateAttachment(attachment))
	foundAttachments, err := repo.ListAttachmentsByItemID(foundItems[0].ID)
	require.NoError(t, err)
	require.Len(t, foundAttachments, 1)
	assert.Equal(t, attachment.StorageKey, foundAttachments[0].StorageKey)
}

func TestOrderEvidenceRepositoryRejectsCrossOrderItems(t *testing.T) {
	db := newOrderEvidenceRepositoryTestDB(t)
	orderRecord, snapshot := seedOrderEvidenceRepositoryOrderAndSnapshot(t, db)
	repo := NewOrderEvidenceRepository(db)

	pkg := &orderevidence.OrderEvidencePackage{
		OrderID:                    orderRecord.ID,
		SnapshotID:                 snapshot.ID,
		PackageVersion:             1,
		Status:                     orderevidence.PackageStatusIncomplete,
		OrderTotalUSDSnapshotMinor: 10000,
		SchemaVersion:              orderevidence.OrderEvidencePackageSchemaVersion,
	}
	items := []orderevidence.OrderEvidenceItem{{
		OrderID:        orderRecord.ID + 99,
		ItemType:       orderevidence.EvidenceItemTypeSignedPOD,
		Status:         orderevidence.EvidenceItemStatusMissing,
		RequiredReason: orderevidence.EvidenceRequiredReasonBase,
	}}
	require.ErrorContains(t, repo.CreatePackageWithItems(pkg, items), "does not match package order_id")

	var packageCount int64
	require.NoError(t, db.Model(&orderevidence.OrderEvidencePackage{}).Count(&packageCount).Error)
	assert.Zero(t, packageCount)
}

func TestOrderEvidenceRepositoryListsOrdersWithAndWithoutEvidencePackages(t *testing.T) {
	db := newOrderEvidenceRepositoryTestDB(t)
	firstOrder, snapshot := seedOrderEvidenceRepositoryOrderAndSnapshot(t, db)
	secondOrder := order.Order{
		OrderNumber:      "TZ-2026-REPOSITORY-NO-PACKAGE",
		TotalAmountMinor: 90000,
		Currency:         "USD",
		ShippingAddress: order.Address{
			Email: "second@example.com",
		},
		FXSnapshotData: repositoryEvidenceFXSnapshot(),
	}
	require.NoError(t, db.Create(&secondOrder).Error)

	repo := NewOrderEvidenceRepository(db)
	pkg := &orderevidence.OrderEvidencePackage{
		OrderID:                    firstOrder.ID,
		SnapshotID:                 snapshot.ID,
		PackageVersion:             1,
		Status:                     orderevidence.PackageStatusIncomplete,
		OrderTotalUSDSnapshotMinor: 10000,
		SchemaVersion:              orderevidence.OrderEvidencePackageSchemaVersion,
	}
	items := []orderevidence.OrderEvidenceItem{
		{
			OrderID:        firstOrder.ID,
			ItemType:       orderevidence.EvidenceItemTypeOutboundWeightPackaging,
			Status:         orderevidence.EvidenceItemStatusMissing,
			RequiredReason: orderevidence.EvidenceRequiredReasonBase,
		},
		{
			OrderID:        firstOrder.ID,
			ItemType:       orderevidence.EvidenceItemTypeSignedPOD,
			Status:         orderevidence.EvidenceItemStatusComplete,
			RequiredReason: orderevidence.EvidenceRequiredReasonBase,
			DataJSON:       datatypes.JSON([]byte(`{"delivered":true}`)),
			CapturedAt:     timePtrForRepositoryEvidence(time.Now().UTC()),
			ContentSHA256:  repositoryEvidenceHash(`{"delivered":true}`),
		},
	}
	require.NoError(t, repo.CreatePackageWithItems(pkg, items))

	rows, total, err := repo.ListAdminOrders(OrderEvidenceAdminListQuery{
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, rows, 2)

	byOrderID := make(map[uint]OrderEvidenceAdminListRow, len(rows))
	for _, row := range rows {
		byOrderID[row.OrderID] = row
	}

	firstRow := byOrderID[firstOrder.ID]
	assert.Equal(t, pkg.ID, firstRow.PackageID)
	assert.Equal(t, orderevidence.PackageStatusIncomplete, firstRow.PackageStatus)
	assert.Equal(t, 2, firstRow.TotalEvidenceItems)
	assert.Equal(t, 1, firstRow.CompleteEvidenceItems)
	assert.Equal(t, 1, firstRow.PendingEvidenceItems)

	secondRow := byOrderID[secondOrder.ID]
	assert.Zero(t, secondRow.PackageID)
	assert.Empty(t, secondRow.PackageStatus)
	assert.Zero(t, secondRow.TotalEvidenceItems)
	assert.False(t, secondRow.IsHighValue)

	rows, total, err = repo.ListAdminOrders(OrderEvidenceAdminListQuery{
		Page:     1,
		PageSize: 10,
		Search:   "second@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, rows, 1)
	assert.Equal(t, secondOrder.ID, rows[0].OrderID)
}

func newOrderEvidenceRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&order.Order{},
		&order.OrderItem{},
		&orderevidence.OrderEvidenceSnapshot{},
		&orderevidence.OrderEvidencePackage{},
		&orderevidence.OrderEvidenceItem{},
		&orderevidence.OrderEvidenceAttachment{},
	))
	return db
}

func seedOrderEvidenceRepositoryOrderAndSnapshot(
	t *testing.T,
	db *gorm.DB,
) (order.Order, orderevidence.OrderEvidenceSnapshot) {
	t.Helper()
	orderRecord := order.Order{
		OrderNumber:      "TZ-2026-REPOSITORY-PACKAGE",
		TotalAmountMinor: 10000,
		Currency:         "USD",
		FXSnapshotData:   repositoryEvidenceFXSnapshot(),
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	snapshot := orderevidence.OrderEvidenceSnapshot{
		OrderID:               orderRecord.ID,
		SchemaVersion:         orderevidence.OrderEvidenceSnapshotSchemaVersion,
		ConfirmedAt:           time.Now().UTC(),
		Currency:              "USD",
		OrderTotalAmountMinor: 10000,
		OrderTotalUSDMinor:    10000,
		SnapshotData:          datatypes.JSON([]byte(`{"schema_version":1,"items":[{}]}`)),
		SnapshotSHA256:        repositoryEvidenceHash(`{"schema_version":1,"items":[{}]}`),
	}
	require.NoError(t, db.Create(&snapshot).Error)
	return orderRecord, snapshot
}

func repositoryEvidenceFXSnapshot() datatypes.JSON {
	return datatypes.JSON([]byte(`{"version":1,"base_currency":"USD","order_currency":"USD","rate_decimal":"1","source":"test","captured_at":"2026-09-04T12:00:00Z"}`))
}

func repositoryEvidenceHash(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func timePtrForRepositoryEvidence(value time.Time) *time.Time {
	return &value
}
