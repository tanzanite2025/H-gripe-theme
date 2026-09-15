package admin

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/currency"
	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	paymentdomain "commerce-platform/internal/domain/payment"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type orderHandlerAuditRecorder struct {
	logs []audit.AuditLog
}

func (r *orderHandlerAuditRecorder) CreateAuditLog(log *audit.AuditLog) error {
	if log != nil {
		r.logs = append(r.logs, *log)
	}
	return nil
}

func TestFulfillOrderRecordsSuccessfulFulfillmentAudit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, handler, providerID, carrierServiceID := newOrderFulfillmentAuditHandler(t)
	auditRecorder := &orderHandlerAuditRecorder{}
	handler.ConfigureAuditService(auditRecorder)

	orderRecord := orderdomain.Order{
		OrderNumber:    "ORDER-FULFILL-AUDIT-SUCCESS",
		UserID:         42,
		Status:         "processing",
		PaymentStatus:  "paid",
		ShippingStatus: "pending",
		Currency:       "USD",
		TotalAmount:    1200,
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	seedReadyFulfillmentEvidenceForAdminTest(t, db, &orderRecord)

	context := newOrderFulfillmentAuditContext(
		t,
		orderRecord.ID,
		`{"tracking_number":"TRACK-FULFILL-AUDIT","tracking_provider_id":`+
			formatUint(providerID)+`,"carrier_service_id":`+
			formatUint(carrierServiceID)+`,"signature_confirmed":true}`,
	)

	handler.FulfillOrder(context)

	require.Equal(t, http.StatusOK, context.Writer.Status())
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	assert.Equal(t, "execute", log.Action)
	assert.Equal(t, "order_fulfillment", log.Resource)
	assert.Equal(t, orderRecord.ID, log.ResourceID)
	assert.Equal(t, "success", log.Status)
	assert.Equal(t, uint(7), log.UserID)
	assert.Contains(t, log.Changes, `"signature_confirmed":true`)
	assert.Contains(t, log.OldValue, `"shipping_status":"pending"`)
	assert.Contains(t, log.NewValue, `"shipping_status":"shipped"`)
	assert.NotContains(t, log.Changes, "api_key")
	assert.NotContains(t, log.Changes, "webhook_secret")
}

func TestFulfillOrderRecordsFailedFulfillmentAuditWithoutChangingOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, handler, providerID, _ := newOrderFulfillmentAuditHandler(t)
	auditRecorder := &orderHandlerAuditRecorder{}
	handler.ConfigureAuditService(auditRecorder)

	orderRecord := orderdomain.Order{
		OrderNumber:    "ORDER-FULFILL-AUDIT-FAILED",
		UserID:         42,
		Status:         "processing",
		PaymentStatus:  "pending",
		ShippingStatus: "pending",
		Currency:       "USD",
		TotalAmount:    1200,
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	context := newOrderFulfillmentAuditContext(
		t,
		orderRecord.ID,
		`{"tracking_number":"TRACK-FULFILL-AUDIT-FAILED","tracking_provider_id":`+
			formatUint(providerID)+`}`,
	)

	handler.FulfillOrder(context)

	assert.NotEqual(t, http.StatusOK, context.Writer.Status())
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	assert.Equal(t, "order_fulfillment", log.Resource)
	assert.Equal(t, "failed", log.Status)
	assert.Contains(t, log.ErrorMessage, "paid")
	assert.Contains(t, log.OldValue, `"shipping_status":"pending"`)
	assert.NotContains(t, log.NewValue, `"shipping_status":"shipped"`)

	var stored orderdomain.Order
	require.NoError(t, db.First(&stored, orderRecord.ID).Error)
	assert.Equal(t, "processing", stored.Status)
	assert.Equal(t, "pending", stored.ShippingStatus)
}

func TestRespondOrderServiceErrorMapsFulfillmentHoldToConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	respondOrderServiceError(
		context,
		service.ErrOrderFulfillmentOnHold,
		"fallback",
		http.StatusInternalServerError,
	)

	assert.Equal(t, http.StatusConflict, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "payment dispute or review")
}

func TestUpdateTrackingInfoRecordsTrackingCorrectionAudit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, handler, providerID, carrierServiceID := newOrderFulfillmentAuditHandler(t)
	auditRecorder := &orderHandlerAuditRecorder{}
	handler.ConfigureAuditService(auditRecorder)

	orderRecord := orderdomain.Order{
		OrderNumber:    "ORDER-TRACKING-CORRECTION-AUDIT",
		UserID:         42,
		Status:         "processing",
		PaymentStatus:  "paid",
		ShippingStatus: "pending",
		Currency:       "USD",
		TotalAmount:    100,
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	seedReadyFulfillmentEvidenceForAdminTest(t, db, &orderRecord)

	fulfillmentContext := newOrderFulfillmentAuditContext(
		t,
		orderRecord.ID,
		`{"tracking_number":"TRACK-AUDIT-ORIGINAL","tracking_provider_id":`+
			formatUint(providerID)+`,"carrier_service_id":`+
			formatUint(carrierServiceID)+`}`,
	)
	handler.FulfillOrder(fulfillmentContext)
	require.Equal(t, http.StatusOK, fulfillmentContext.Writer.Status())
	auditRecorder.logs = nil

	correctionContext := newOrderTrackingAuditContext(
		t,
		orderRecord.ID,
		`{"tracking_number":"TRACK-AUDIT-CORRECTED","tracking_provider_id":`+
			formatUint(providerID)+`,"carrier_service_id":`+
			formatUint(carrierServiceID)+`}`,
	)
	handler.UpdateTrackingInfo(correctionContext)

	require.Equal(t, http.StatusOK, correctionContext.Writer.Status())
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	assert.Equal(t, adminAuditActionUpdate, log.Action)
	assert.Equal(t, adminAuditResourceOrderTracking, log.Resource)
	assert.Equal(t, orderRecord.ID, log.ResourceID)
	assert.Equal(t, adminAuditStatusSuccess, log.Status)
	assert.Contains(t, log.Changes, `"operation":"tracking_correction"`)
	assert.Contains(t, log.OldValue, `"tracking_number":"TRACK-AUDIT-ORIGINAL"`)
	assert.Contains(t, log.NewValue, `"tracking_number":"TRACK-AUDIT-CORRECTED"`)
	assert.NotContains(t, log.Changes, "api_key")
	assert.NotContains(t, log.Changes, "webhook_secret")
}

func newOrderFulfillmentAuditHandler(t *testing.T) (*gorm.DB, *OrderHandler, uint, uint) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(
		&orderdomain.Order{},
		&orderdomain.OrderItem{},
		&orderevidence.OrderEvidenceSnapshot{},
		&orderevidence.OrderEvidencePackage{},
		&orderevidence.OrderEvidenceItem{},
		&orderevidence.OrderEvidenceAttachment{},
		&paymentdomain.StripeDispute{},
		&paymentdomain.PayPalDispute{},
		&paymentdomain.PaymentReview{},
		&shippingdomain.Carrier{},
		&shippingdomain.CarrierService{},
		&shippingdomain.TrackingProviderConfig{},
		&shippingdomain.TrackingCarrierMapping{},
		&shippingdomain.TrackingShipment{},
		&shippingdomain.TrackingEvent{},
	))

	orderRepo := repository.NewOrderRepository(db)
	shippingRepo := repository.NewShippingRepository(db)
	shippingService := service.NewShippingService(shippingRepo)
	shippingService.ConfigureOrderRepository(orderRepo)
	txManager := repository.NewTxManager(
		db,
		orderRepo,
		repository.NewProductRepository(db),
		repository.NewCouponRepository(db),
		repository.NewLoyaltyRepository(db),
		repository.NewPaymentRepository(db),
		shippingRepo,
	)
	txManager.ConfigureOrderEvidenceSnapshotRepository(repository.NewOrderEvidenceSnapshotRepository(db))
	txManager.ConfigureOrderEvidenceRepository(repository.NewOrderEvidenceRepository(db))
	orderService := service.NewOrderService(txManager, orderRepo, nil, shippingService)
	orderService.ConfigureOrderEvidence(service.NewOrderEvidenceService())

	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode: "test",
		ProviderName: "Test Tracking",
		Enabled:      true,
	}
	require.NoError(t, db.Create(&provider).Error)
	carrier := shippingdomain.Carrier{
		Name:    "DHL",
		Code:    "DHL",
		Enabled: true,
	}
	require.NoError(t, db.Create(&carrier).Error)
	carrierService := shippingdomain.CarrierService{
		CarrierID:            carrier.ID,
		ServiceCode:          "DHL-TEST",
		ServiceName:          "DHL Test",
		BillingMode:          "actual_weight",
		VolumetricDivisor:    6000,
		MinChargeWeightGrams: 1,
		Enabled:              true,
	}
	require.NoError(t, db.Create(&carrierService).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingCarrierMapping{
		ProviderID:          provider.ID,
		Scope:               "carrier_service",
		CarrierServiceID:    &carrierService.ID,
		ProviderCarrierCode: "DHL-TEST",
		Enabled:             true,
	}).Error)

	return db, NewOrderHandler(orderService), provider.ID, carrierService.ID
}

func seedReadyFulfillmentEvidenceForAdminTest(t *testing.T, db *gorm.DB, orderRecord *orderdomain.Order) {
	t.Helper()

	var storedOrder orderdomain.Order
	require.NoError(t, db.Preload("Items").First(&storedOrder, orderRecord.ID).Error)
	storedOrder.Currency = "USD"
	storedOrder.FXSnapshotData = currency.OrderFXSnapshotJSON(currency.OrderFXSnapshot{
		Version:         currency.OrderFXSnapshotVersion,
		BaseCurrency:    "USD",
		OrderCurrency:   "USD",
		BaseToOrderRate: 1,
		Source:          "admin-fulfillment-test",
		CapturedAt:      time.Now().UTC(),
	})
	require.NoError(t, db.Model(&orderdomain.Order{}).
		Where("id = ?", storedOrder.ID).
		Updates(map[string]interface{}{
			"currency":    storedOrder.Currency,
			"fx_snapshot": storedOrder.FXSnapshotData,
		}).Error)

	variantID := uint(1)
	item := orderdomain.OrderItem{
		OrderID:                storedOrder.ID,
		ProductID:              1,
		VariantID:              &variantID,
		ProductName:            "Fulfillment audit product",
		SKU:                    fmt.Sprintf("FULFILL-AUDIT-%d", storedOrder.ID),
		Quantity:               1,
		Price:                  storedOrder.TotalAmount,
		Subtotal:               storedOrder.TotalAmount,
		Total:                  storedOrder.TotalAmount,
		Attributes:             "{}",
		WeightGrams:            1000,
		DeclaredValue:          float64PtrForAdminFulfillmentTest(storedOrder.TotalAmount),
		DeclaredValueConfirmed: true,
	}
	require.NoError(t, db.Create(&item).Error)
	storedOrder.Items = []orderdomain.OrderItem{item}

	snapshot, err := orderevidence.BuildOrderEvidenceSnapshot(
		&storedOrder,
		[]orderevidence.SnapshotItemInput{{Item: item}},
		time.Now().UTC(),
	)
	require.NoError(t, err)
	require.NoError(t, db.Create(snapshot).Error)

	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	_, err = service.NewOrderEvidenceService().CreateInitialPackage(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		snapshot,
	)
	require.NoError(t, err)

	pkg, err := evidenceRepo.FindLatestPackageByOrderID(storedOrder.ID)
	require.NoError(t, err)
	items, err := evidenceRepo.ListItemsByPackageID(pkg.ID)
	require.NoError(t, err)
	evidenceService := service.NewOrderEvidenceService()
	for _, evidenceItem := range items {
		switch evidenceItem.ItemType {
		case orderevidence.EvidenceItemTypeProductIdentity:
			registerAdminFulfillmentAttachment(t, db, storedOrder.ID, evidenceItem.ID, "identity.jpg", "identity")
			_, err = evidenceService.UpdateItem(
				repository.TxRepositories{OrderEvidence: evidenceRepo},
				service.OrderEvidenceItemUpdateInput{
					OrderID:    storedOrder.ID,
					ItemID:     evidenceItem.ID,
					Status:     orderevidence.EvidenceItemStatusComplete,
					DataJSON:   datatypes.JSON([]byte(`{"capture_note":"dispatch product photo"}`)),
					CapturedBy: 7,
				},
			)
			require.NoError(t, err)
		case orderevidence.EvidenceItemTypeOutboundWeightPackaging:
			registerAdminFulfillmentAttachment(t, db, storedOrder.ID, evidenceItem.ID, "outbound.jpg", "outbound")
			_, err = evidenceService.UpdateItem(
				repository.TxRepositories{OrderEvidence: evidenceRepo},
				service.OrderEvidenceItemUpdateInput{
					OrderID: storedOrder.ID,
					ItemID:  evidenceItem.ID,
					Status:  orderevidence.EvidenceItemStatusComplete,
					DataJSON: datatypes.JSON([]byte(`{
						"schema_version": 1,
						"gross_weight_g": 1200,
						"package_count": 1,
						"packaging_method": "carton"
					}`)),
					CapturedBy: 7,
				},
			)
			require.NoError(t, err)
		}
	}
}

func float64PtrForAdminFulfillmentTest(value float64) *float64 {
	return &value
}
func registerAdminFulfillmentAttachment(
	t *testing.T,
	db *gorm.DB,
	orderID uint,
	itemID uint,
	filename string,
	content string,
) {
	t.Helper()
	hash := sha256.Sum256([]byte(content))
	require.NoError(t, db.Create(&orderevidence.OrderEvidenceAttachment{
		EvidenceItemID:   itemID,
		StorageKey:       fmt.Sprintf("order-evidence/%d/%d/%s", orderID, itemID, filename),
		OriginalFilename: filename,
		MimeType:         "image/jpeg",
		SizeBytes:        int64(len(content)),
		SHA256:           hex.EncodeToString(hash[:]),
		UploadedBy:       7,
	}).Error)
}

func newOrderFulfillmentAuditContext(t *testing.T, orderID uint, body string) *gin.Context {
	t.Helper()
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: formatUint(orderID)}}
	context.Set("user_id", uint(7))
	context.Set("username", "fulfillment-admin")
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/orders/"+formatUint(orderID)+"/fulfillment",
		strings.NewReader(body),
	)
	context.Request.Header.Set("Content-Type", "application/json")
	return context
}

func newOrderTrackingAuditContext(t *testing.T, orderID uint, body string) *gin.Context {
	t.Helper()
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: formatUint(orderID)}}
	context.Set("user_id", uint(7))
	context.Set("username", "fulfillment-admin")
	context.Request = httptest.NewRequest(
		http.MethodPatch,
		"/api/admin/orders/"+formatUint(orderID)+"/tracking",
		strings.NewReader(body),
	)
	context.Request.Header.Set("Content-Type", "application/json")
	return context
}

func formatUint(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}
