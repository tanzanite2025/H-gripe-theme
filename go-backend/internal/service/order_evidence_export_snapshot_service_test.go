package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestOrderEvidenceExportSnapshotServiceRequiresLockedPackage(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&orderevidence.OrderEvidenceExportSnapshot{}))
	snapshot := servicePlanSnapshot(t, db, false)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	_, err := NewOrderEvidenceService().CreateInitialPackage(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		snapshot,
	)
	require.NoError(t, err)

	exportService := NewOrderEvidenceExportSnapshotService(
		repository.NewOrderEvidenceExportSnapshotRepository(db),
		evidenceRepo,
		NewOrderEvidencePackageAssembler(
			repository.NewOrderRepository(db),
			evidenceRepo,
			nil,
		),
	)
	_, err = exportService.Export(snapshot.OrderID, 7)
	require.ErrorIs(t, err, ErrOrderEvidenceExportRequiresLocked)
}

func TestOrderEvidenceExportSnapshotServiceFreezesSourcesForRepeatedExport(t *testing.T) {
	fixture := newOrderEvidenceExportServiceFixture(t, false)
	originalEventID := seedOrderEvidenceExportTracking(t, fixture.db, fixture.orderID)

	first, err := fixture.exportService.Export(fixture.orderID, 7)
	require.NoError(t, err)
	require.NotNil(t, first)
	require.NoError(t, first.Validate())

	var firstPayload OrderEvidenceExportSnapshotPayload
	require.NoError(t, json.Unmarshal(first.SnapshotData, &firstPayload))
	require.NotNil(t, firstPayload.Order)
	require.NotNil(t, firstPayload.TrackingContext)
	require.Len(t, firstPayload.TrackingContext.Shipments, 1)
	require.Len(t, firstPayload.TrackingEvents, 1)
	assert.Equal(t, fixture.orderNumber, firstPayload.Order.OrderNumber)
	assert.Equal(t, "EXPORT-TRACKING", firstPayload.TrackingContext.Shipments[0].TrackingNumber)
	assert.Equal(t, "original delivery", firstPayload.TrackingEvents[0].Description)
	assert.NotContains(t, string(first.SnapshotData), "storage-key")
	assert.NotContains(t, string(first.SnapshotData), `"storage_key"`)

	require.NoError(t, fixture.db.Model(&order.Order{}).
		Where("id = ?", fixture.orderID).
		Update("order_number", "TZ-2026-EXPORT-CHANGED").Error)
	require.NoError(t, fixture.db.Model(&shipping.TrackingEvent{}).
		Where("id = ?", originalEventID).
		Update("description", "changed delivery").Error)

	second, err := fixture.exportService.Export(fixture.orderID, 99)
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, first.SnapshotData, second.SnapshotData)
	assert.Equal(t, uint(7), second.CreatedBy)
}

func TestOrderEvidenceExportSnapshotServiceRejectsTamperedStoredManifest(t *testing.T) {
	fixture := newOrderEvidenceExportServiceFixture(t, false)
	first, err := fixture.exportService.Export(fixture.orderID, 7)
	require.NoError(t, err)

	require.NoError(t, fixture.db.Exec(
		"UPDATE order_evidence_export_snapshots SET snapshot_data = ? WHERE id = ?",
		[]byte(`{"export_type":"tampered"}`),
		first.ID,
	).Error)

	_, err = fixture.exportService.Export(fixture.orderID, 7)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrOrderEvidenceExportSnapshotInvalid))
	assert.Contains(t, err.Error(), "sha256 does not match")
}

type orderEvidenceExportServiceFixture struct {
	db            *gorm.DB
	orderID       uint
	orderNumber   string
	exportService *OrderEvidenceExportSnapshotService
}

func newOrderEvidenceExportServiceFixture(t *testing.T, requiresQC bool) orderEvidenceExportServiceFixture {
	t.Helper()
	db := newOrderEvidenceServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&shipping.TrackingProviderConfig{},
		&shipping.TrackingShipment{},
		&shipping.TrackingEvent{},
		&orderevidence.OrderEvidenceExportSnapshot{},
	))

	snapshot := servicePlanSnapshot(t, db, requiresQC)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	evidenceService := NewOrderEvidenceService()
	pkg, err := evidenceService.CreateInitialPackage(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		snapshot,
	)
	require.NoError(t, err)

	items, err := evidenceRepo.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	attachmentService := NewOrderEvidenceAttachmentService()
	for _, item := range items {
		if item.ItemType == orderevidence.EvidenceItemTypeConfigurationConfirmation {
			continue
		}
		hash := fmt.Sprintf("%064d", item.ID)
		_, err = attachmentService.RegisterReference(
			repository.TxRepositories{OrderEvidence: evidenceRepo},
			OrderEvidenceAttachmentReferenceInput{
				EvidenceItemID:   item.ID,
				StorageKey:       fmt.Sprintf("order-evidence/%d/%d/%s.jpg", snapshot.OrderID, item.ID, item.ItemType),
				OriginalFilename: item.ItemType + ".jpg",
				MimeType:         "image/jpeg",
				SizeBytes:        32,
				SHA256:           hash,
				UploadedBy:       7,
			},
		)
		require.NoError(t, err)
		_, err = evidenceService.UpdateItem(
			repository.TxRepositories{OrderEvidence: evidenceRepo},
			OrderEvidenceItemUpdateInput{
				ItemID:     item.ID,
				Status:     orderevidence.EvidenceItemStatusComplete,
				DataJSON:   exportEvidenceData(item.ItemType),
				CapturedBy: 7,
			},
		)
		require.NoError(t, err)
	}
	_, _, err = evidenceService.LockPackage(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		pkg.ID,
		time.Date(2026, 9, 5, 13, 0, 0, 0, time.UTC),
	)
	require.NoError(t, err)

	orderRepo := repository.NewOrderRepository(db)
	exportService := NewOrderEvidenceExportSnapshotService(
		repository.NewOrderEvidenceExportSnapshotRepository(db),
		evidenceRepo,
		NewOrderEvidencePackageAssembler(
			orderRepo,
			evidenceRepo,
			repository.NewShippingRepository(db),
		),
	)
	var orderRecord order.Order
	require.NoError(t, db.First(&orderRecord, snapshot.OrderID).Error)
	return orderEvidenceExportServiceFixture{
		db:            db,
		orderID:       snapshot.OrderID,
		orderNumber:   orderRecord.OrderNumber,
		exportService: exportService,
	}
}

func exportEvidenceData(itemType string) datatypes.JSON {
	switch itemType {
	case orderevidence.EvidenceItemTypeOutboundWeightPackaging:
		return datatypes.JSON([]byte(`{
			"schema_version": 1,
			"gross_weight_g": 12640,
			"package_count": 2,
			"packaging_method": "double wall carton"
		}`))
	case orderevidence.EvidenceItemTypeSignedPOD:
		return datatypes.JSON([]byte(`{
			"schema_version": 1,
			"tracking_number": "EXPORT-TRACKING",
			"delivered_at": "2026-09-05T10:30:00Z"
		}`))
	default:
		return datatypes.JSON([]byte(`{"manual_record":"captured at dispatch"}`))
	}
}

func seedOrderEvidenceExportTracking(t *testing.T, db *gorm.DB, orderID uint) uint {
	t.Helper()
	provider := &shipping.TrackingProviderConfig{
		ProviderCode: "export-test",
		ProviderName: "Export Test Tracking",
		APIKey:       "must-not-enter-snapshot",
		Enabled:      true,
	}
	require.NoError(t, db.Create(provider).Error)
	shipment := &shipping.TrackingShipment{
		OrderID:             orderID,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "EXPORT-TRACKING",
		ProviderCarrierCode: "export-carrier",
		RegistrationStatus:  "registered",
		SyncStatus:          "synced",
		Enabled:             true,
	}
	require.NoError(t, db.Create(shipment).Error)
	event := &shipping.TrackingEvent{
		OrderID:             orderID,
		TrackingNumber:      shipment.TrackingNumber,
		ProviderCarrierCode: shipment.ProviderCarrierCode,
		Status:              "delivered",
		Description:         "original delivery",
		ProofOfDeliveryURL:  "https://carrier.example.test/export-pod.pdf",
		EventTime:           time.Date(2026, 9, 5, 14, 0, 0, 0, time.UTC),
	}
	require.NoError(t, db.Create(event).Error)
	return event.ID
}
