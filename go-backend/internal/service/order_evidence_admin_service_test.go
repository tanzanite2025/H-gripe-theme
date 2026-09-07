package service

import (
	"encoding/json"
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

func TestOrderEvidenceAdminServiceRejectsCrossOrderItemUpdate(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	firstOrder, _ := seedAdminEvidenceOrder(t, db, "TZ-2026-ADMIN-FIRST")
	secondOrder, _ := seedAdminEvidenceOrder(t, db, "TZ-2026-ADMIN-SECOND")
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	adminService := newOrderEvidenceAdminServiceForTest(db, evidenceRepo)

	firstPackage, err := NewOrderEvidenceService().CreateInitialPackage(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		mustAdminEvidenceSnapshot(t, db, firstOrder.ID),
	)
	require.NoError(t, err)
	secondPackage, err := NewOrderEvidenceService().CreateInitialPackage(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		mustAdminEvidenceSnapshot(t, db, secondOrder.ID),
	)
	require.NoError(t, err)

	secondItems, err := evidenceRepo.ListItemsByPackageID(secondPackage.ID)
	require.NoError(t, err)
	require.NotEmpty(t, secondItems)
	secondIdentity := findEvidenceItemByType(secondItems, orderevidence.EvidenceItemTypeProductIdentity)
	require.NotZero(t, secondIdentity.ID)

	_, err = adminService.UpdateItem(firstOrder.ID, OrderEvidenceAdminItemUpdateInput{
		ItemID:     secondIdentity.ID,
		Status:     orderevidence.EvidenceItemStatusComplete,
		DataJSON:   datatypes.JSON([]byte(`{"serial_number":"cross-order"}`)),
		CapturedBy: 7,
	})
	require.ErrorIs(t, err, ErrOrderEvidenceAdminItemNotFound)

	var stored orderevidence.OrderEvidenceItem
	require.NoError(t, db.First(&stored, secondIdentity.ID).Error)
	assert.NotZero(t, firstPackage.ID)
	assert.Equal(t, orderevidence.EvidenceItemStatusMissing, stored.Status)
}

func TestOrderEvidenceAdminServiceReturnsCompletePackageGraphAfterUpdate(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	orderRecord, snapshot := seedAdminEvidenceOrder(t, db, "TZ-2026-ADMIN-GRAPH")
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	adminService := newOrderEvidenceAdminServiceForTest(db, evidenceRepo)
	_, err := NewOrderEvidenceService().CreateInitialPackage(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		snapshot,
	)
	require.NoError(t, err)

	packageResult, err := adminService.GetPackage(orderRecord.ID)
	require.NoError(t, err)
	items := packageResult.Package.Items
	require.NoError(t, err)
	identity := findEvidenceItemByType(items, orderevidence.EvidenceItemTypeProductIdentity)

	_, err = NewOrderEvidenceAttachmentService().RegisterReference(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		OrderEvidenceAttachmentReferenceInput{
			EvidenceItemID:   identity.ID,
			StorageKey:       "order-evidence/1/admin-identity.jpg",
			OriginalFilename: "admin-identity.jpg",
			MimeType:         "image/jpeg",
			SizeBytes:        12,
			SHA256:           "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			UploadedBy:       7,
		},
	)
	require.NoError(t, err)

	result, err := adminService.UpdateItem(orderRecord.ID, OrderEvidenceAdminItemUpdateInput{
		ItemID:     identity.ID,
		Status:     orderevidence.EvidenceItemStatusComplete,
		DataJSON:   datatypes.JSON([]byte(`{"serial_number":"SN-ADMIN"}`)),
		CapturedBy: 7,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Package)
	assert.Equal(t, orderRecord.ID, result.Package.OrderID)
	assert.Equal(t, 50, result.Completeness.Percent)
	assert.Len(t, result.Package.Items, len(items))
}

func TestOrderEvidenceAdminServiceProjectsTrackingContextWithoutProviderSecrets(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&shipping.TrackingProviderConfig{},
		&shipping.Carrier{},
		&shipping.CarrierService{},
		&shipping.TrackingCarrierMapping{},
		&shipping.TrackingShipment{},
		&shipping.TrackingEvent{},
	))

	orderRecord, snapshot := seedAdminEvidenceOrder(t, db, "TZ-2026-ADMIN-TRACKING")
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	_, err := NewOrderEvidenceService().CreateInitialPackage(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		snapshot,
	)
	require.NoError(t, err)

	provider := &shipping.TrackingProviderConfig{
		ProviderCode: "mock",
		ProviderName: "Mock Tracking",
		APIKey:       "provider-secret",
		Enabled:      true,
	}
	require.NoError(t, db.Create(provider).Error)
	shipment := &shipping.TrackingShipment{
		OrderID:             orderRecord.ID,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      "TRACK-ADMIN-100",
		ProviderCarrierCode: "mock-carrier",
		RegistrationStatus:  "registered",
		SyncStatus:          "synced",
		EventCount:          2,
		Enabled:             true,
	}
	require.NoError(t, db.Create(shipment).Error)

	deliveredAt := time.Date(2026, 9, 5, 4, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(&shipping.TrackingEvent{
		OrderID:                orderRecord.ID,
		TrackingNumber:         shipment.TrackingNumber,
		ProviderCarrierCode:    shipment.ProviderCarrierCode,
		Status:                 "delivered",
		Location:               "Seattle",
		RecipientSignatureName: "A. Customer",
		ProofOfDeliveryURL:     "https://provider.test/pod/latest",
		EventTime:              deliveredAt,
	}).Error)
	require.NoError(t, db.Create(&shipping.TrackingEvent{
		OrderID:             orderRecord.ID,
		TrackingNumber:      shipment.TrackingNumber,
		ProviderCarrierCode: shipment.ProviderCarrierCode,
		Status:              "in_transit",
		ProofOfDeliveryURL:  "https://provider.test/pod/older",
		EventTime:           deliveredAt.Add(-time.Hour),
	}).Error)

	adminService := newOrderEvidenceAdminServiceForTest(db, evidenceRepo)
	adminService.ConfigureOrderEvidencePackageAssembler(
		NewOrderEvidencePackageAssembler(
			repository.NewOrderRepository(db),
			evidenceRepo,
			repository.NewShippingRepository(db),
		),
	)

	result, err := adminService.GetPackage(orderRecord.ID)
	require.NoError(t, err)
	require.NotNil(t, result.TrackingContext)
	require.NotNil(t, result.TrackingContext.Shipment)
	require.NotNil(t, result.TrackingContext.LatestDeliveryEvent)
	require.NotNil(t, result.TrackingContext.ManualPOD)

	assert.Equal(t, "TRACK-ADMIN-100", result.TrackingContext.Shipment.TrackingNumber)
	assert.Equal(t, "mock", result.TrackingContext.Shipment.ProviderCode)
	assert.Equal(t, "Mock Tracking", result.TrackingContext.Shipment.ProviderName)
	assert.Equal(t, "delivered", result.TrackingContext.LatestDeliveryEvent.Status)
	assert.Equal(t, "Seattle", result.TrackingContext.LatestDeliveryEvent.Location)
	assert.Equal(t, "https://provider.test/pod/latest", result.TrackingContext.ProviderPODURL)
	assert.Equal(t, orderevidence.EvidenceItemStatusMissing, result.TrackingContext.ManualPOD.Status)
	assert.Equal(t, 0, result.TrackingContext.ManualPOD.AttachmentCount)

	payload, err := json.Marshal(result)
	require.NoError(t, err)
	assert.NotContains(t, string(payload), "provider-secret")
}

func newOrderEvidenceAdminServiceForTest(
	db *gorm.DB,
	evidenceRepo *repository.OrderEvidenceRepository,
) *OrderEvidenceAdminService {
	orderRepo := repository.NewOrderRepository(db)
	txManager := repository.NewTxManager(
		db,
		orderRepo,
		repository.NewProductRepository(db),
		repository.NewCouponRepository(db),
		repository.NewLoyaltyRepository(db),
		repository.NewPaymentRepository(db),
	)
	txManager.ConfigureOrderEvidenceRepository(evidenceRepo)
	return NewOrderEvidenceAdminService(txManager, orderRepo, evidenceRepo, NewOrderEvidenceService())
}

func seedAdminEvidenceOrder(
	t *testing.T,
	db *gorm.DB,
	orderNumber string,
) (*order.Order, *orderevidence.OrderEvidenceSnapshot) {
	t.Helper()
	record := &order.Order{
		OrderNumber:    orderNumber,
		TotalAmount:    800,
		Currency:       "USD",
		FXSnapshotData: servicePlanFXSnapshot(),
	}
	require.NoError(t, db.Create(record).Error)
	variantID := uint(22)
	item := order.OrderItem{
		OrderID:     record.ID,
		ProductID:   10,
		VariantID:   &variantID,
		ProductName: "Configured Product",
		SKU:         orderNumber + "-SKU",
		Quantity:    1,
		Price:       800,
		WeightGrams: 9000,
		Attributes:  `{"finish":"black"}`,
	}
	require.NoError(t, db.Create(&item).Error)
	record.Items = []order.OrderItem{item}

	snapshot, err := orderevidence.BuildOrderEvidenceSnapshot(record, []orderevidence.SnapshotItemInput{
		{Item: item},
	}, record.CreatedAt)
	require.NoError(t, err)
	require.NoError(t, db.Create(snapshot).Error)
	return record, snapshot
}

func mustAdminEvidenceSnapshot(t *testing.T, db *gorm.DB, orderID uint) *orderevidence.OrderEvidenceSnapshot {
	t.Helper()
	var snapshot orderevidence.OrderEvidenceSnapshot
	require.NoError(t, db.Where("order_id = ?", orderID).First(&snapshot).Error)
	return &snapshot
}
