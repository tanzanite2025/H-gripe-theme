package admin

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/storage"
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

func TestOrderEvidenceHandlerRejectsCrossOrderItem(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, firstOrderID, _, secondItemID := newAdminOrderEvidenceHandlerFixture(t)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	adminService := newOrderEvidenceAdminServiceForHandlerTest(db, evidenceRepo)
	handler := NewOrderEvidenceHandler(adminService)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(
		http.MethodPatch,
		"/api/admin/orders/1/evidence/items/2",
		bytes.NewBufferString(`{"status":"complete","data_json":{"serial_number":"bad"}}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")
	context.Params = gin.Params{
		{Key: "id", Value: strconv.FormatUint(uint64(firstOrderID), 10)},
		{Key: "item_id", Value: strconv.FormatUint(uint64(secondItemID), 10)},
	}
	context.Set("user_id", uint(7))

	handler.UpdateItem(context)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestOrderEvidenceRoutesSeparateViewAndEditPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", string(auth.RoleViewer))
		c.Set("user_id", uint(7))
		c.Next()
	})
	registerOrderEvidenceRoutes(router.Group(""), NewOrderEvidenceHandler(nil))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/orders/1/evidence/lock", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestOrderEvidenceHandlerListsOrdersAndRejectsInvalidFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, firstOrderID, _, _ := newAdminOrderEvidenceHandlerFixture(t)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	adminService := newOrderEvidenceAdminServiceForHandlerTest(db, evidenceRepo)
	handler := NewOrderEvidenceHandler(adminService)

	t.Run("lists order evidence summary", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		context.Request = httptest.NewRequest(
			http.MethodGet,
			"/api/admin/order-evidence?search=HANDLER-FIRST",
			nil,
		)

		handler.ListOrders(context)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Contains(t, recorder.Body.String(), `"order_id":`+strconv.FormatUint(uint64(firstOrderID), 10))
		require.Contains(t, recorder.Body.String(), `"total":1`)
	})

	t.Run("rejects invalid boolean filter", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		context.Request = httptest.NewRequest(
			http.MethodGet,
			"/api/admin/order-evidence?high_value=maybe",
			nil,
		)

		handler.ListOrders(context)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
	})
}

func TestOrderEvidenceHandlerUploadsAttachmentAndReturnsMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, orderID, itemID, _ := newAdminOrderEvidenceHandlerFixture(t)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	adminService, attachmentService := newOrderEvidenceServicesForHandlerTest(db, evidenceRepo, newAdminOrderEvidenceHandlerStorage(t))
	handler := NewOrderEvidenceHandler(adminService)
	handler.ConfigureAttachmentService(attachmentService)

	request := newAdminEvidenceMultipartRequest(t, http.MethodPost, "/api/admin/orders/"+strconv.FormatUint(uint64(orderID), 10)+"/evidence/items/"+strconv.FormatUint(uint64(itemID), 10)+"/attachments", "identity.png")
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = request
	context.Params = gin.Params{
		{Key: "id", Value: strconv.FormatUint(uint64(orderID), 10)},
		{Key: "item_id", Value: strconv.FormatUint(uint64(itemID), 10)},
	}
	context.Set("user_id", uint(7))

	handler.UploadAttachment(context)

	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"original_filename":"identity.png"`)
	require.Contains(t, recorder.Body.String(), `"mime_type":"image/png"`)

	var attachments []orderevidence.OrderEvidenceAttachment
	require.NoError(t, db.Where("evidence_item_id = ?", itemID).Find(&attachments).Error)
	require.Len(t, attachments, 1)
	assert.Contains(t, attachments[0].StorageKey, "order-evidence/"+strconv.FormatUint(uint64(orderID), 10)+"/"+strconv.FormatUint(uint64(itemID), 10)+"/")
	assert.Equal(t, "identity.png", attachments[0].OriginalFilename)
}

func TestOrderEvidenceHandlerRejectsCrossOrderAttachmentUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, firstOrderID, _, secondItemID := newAdminOrderEvidenceHandlerFixture(t)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	adminService, attachmentService := newOrderEvidenceServicesForHandlerTest(db, evidenceRepo, newAdminOrderEvidenceHandlerStorage(t))
	handler := NewOrderEvidenceHandler(adminService)
	handler.ConfigureAttachmentService(attachmentService)

	request := newAdminEvidenceMultipartRequest(t, http.MethodPost, "/api/admin/orders/"+strconv.FormatUint(uint64(firstOrderID), 10)+"/evidence/items/"+strconv.FormatUint(uint64(secondItemID), 10)+"/attachments", "cross-order.png")
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = request
	context.Params = gin.Params{
		{Key: "id", Value: strconv.FormatUint(uint64(firstOrderID), 10)},
		{Key: "item_id", Value: strconv.FormatUint(uint64(secondItemID), 10)},
	}
	context.Set("user_id", uint(7))

	handler.UploadAttachment(context)

	require.Equal(t, http.StatusNotFound, recorder.Code, recorder.Body.String())
}

func TestOrderEvidenceHandlerRejectsAttachmentUploadAfterPackageLock(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, orderID, itemID, _ := newAdminOrderEvidenceHandlerFixture(t)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	adminService, attachmentService := newOrderEvidenceServicesForHandlerTest(db, evidenceRepo, newAdminOrderEvidenceHandlerStorage(t))

	completeAdminEvidencePackageForHandlerTest(t, adminService, attachmentService, evidenceRepo, orderID)
	lockResult, err := adminService.LockPackage(orderID)
	require.NoError(t, err)
	require.Equal(t, orderevidence.PackageStatusLocked, lockResult.Package.Status)

	handler := NewOrderEvidenceHandler(adminService)
	handler.ConfigureAttachmentService(attachmentService)
	request := newAdminEvidenceMultipartRequest(t, http.MethodPost, "/api/admin/orders/"+strconv.FormatUint(uint64(orderID), 10)+"/evidence/items/"+strconv.FormatUint(uint64(itemID), 10)+"/attachments", "after-lock.png")
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = request
	context.Params = gin.Params{
		{Key: "id", Value: strconv.FormatUint(uint64(orderID), 10)},
		{Key: "item_id", Value: strconv.FormatUint(uint64(itemID), 10)},
	}
	context.Set("user_id", uint(7))

	handler.UploadAttachment(context)

	require.Equal(t, http.StatusConflict, recorder.Code, recorder.Body.String())
}

func TestOrderEvidenceHandlerServesAttachmentWithinOrderScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, orderID, itemID, _ := newAdminOrderEvidenceHandlerFixture(t)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	adminService, attachmentService := newOrderEvidenceServicesForHandlerTest(db, evidenceRepo, newAdminOrderEvidenceHandlerStorage(t))
	fileHeader := newAdminEvidencePNGFileHeader(t, "identity.png")
	_, _, err := evidenceRepo.FindItemAndPackageByIDForOrder(itemID, orderID)
	require.NoError(t, err)
	attachment, err := attachmentService.Upload(context.Background(), orderID, itemID, fileHeader, 7)
	require.NoError(t, err)
	raw := readAdminEvidenceFileHeader(t, fileHeader)

	handler := NewOrderEvidenceHandler(adminService)
	handler.ConfigureAttachmentService(attachmentService)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(
		http.MethodGet,
		"/api/admin/orders/"+strconv.FormatUint(uint64(orderID), 10)+"/evidence/items/"+strconv.FormatUint(uint64(itemID), 10)+"/attachments/"+strconv.FormatUint(uint64(attachment.ID), 10),
		nil,
	)
	context.Params = gin.Params{
		{Key: "id", Value: strconv.FormatUint(uint64(orderID), 10)},
		{Key: "item_id", Value: strconv.FormatUint(uint64(itemID), 10)},
		{Key: "attachment_id", Value: strconv.FormatUint(uint64(attachment.ID), 10)},
	}

	handler.ServeAttachment(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "image/png", recorder.Header().Get("Content-Type"))
	assert.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
	assert.Equal(t, raw, recorder.Body.Bytes())
}

func TestOrderEvidenceHandlerExportsLockedSnapshotWithHeadersAndSanitizedAudit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, orderID, _, _ := newAdminOrderEvidenceHandlerFixture(t)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	adminService, attachmentService := newOrderEvidenceServicesForHandlerTest(db, evidenceRepo, newAdminOrderEvidenceHandlerStorage(t))
	completeAdminEvidencePackageForHandlerTest(t, adminService, attachmentService, evidenceRepo, orderID)
	_, err := adminService.LockPackage(orderID)
	require.NoError(t, err)

	auditRecorder := &orderEvidenceAuditRecorder{}
	handler := NewOrderEvidenceHandler(adminService)
	handler.ConfigureExportService(service.NewOrderEvidenceExportSnapshotService(
		repository.NewOrderEvidenceExportSnapshotRepository(db),
		evidenceRepo,
		service.NewOrderEvidencePackageAssembler(
			repository.NewOrderRepository(db),
			evidenceRepo,
			repository.NewShippingRepository(db),
		),
	))
	handler.ConfigureAuditService(auditRecorder)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(
		http.MethodGet,
		"/api/admin/orders/"+strconv.FormatUint(uint64(orderID), 10)+"/evidence/export",
		nil,
	)
	context.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(orderID), 10)}}
	context.Set("user_id", uint(7))
	context.Set("username", "export-admin")

	handler.ExportSnapshot(context)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
	assert.Contains(t, recorder.Header().Get("Content-Disposition"), "attachment")
	assert.Contains(t, recorder.Header().Get("Content-Disposition"), "order-evidence-")
	assert.NotEmpty(t, recorder.Header().Get("X-Order-Evidence-Export-Snapshot-ID"))
	assert.Equal(t, "1", recorder.Header().Get("X-Order-Evidence-Export-Package-Version"))
	assert.Len(t, recorder.Header().Get("X-Order-Evidence-Export-SHA256"), 64)
	assert.Contains(t, recorder.Body.String(), `"export_type":"order_evidence_manifest"`)
	assert.NotContains(t, recorder.Body.String(), `"storage_key"`)
	assert.NotContains(t, recorder.Body.String(), "order-evidence/1/")

	require.Len(t, auditRecorder.logs, 1)
	auditLog := auditRecorder.logs[0]
	assert.Equal(t, "success", auditLog.Status)
	assert.Contains(t, auditLog.NewValue, `"snapshot_sha256"`)
	assert.NotContains(t, auditLog.NewValue, `"storage_key"`)
	assert.NotContains(t, auditLog.NewValue, "order-evidence/")
	assert.NotContains(t, auditLog.Changes, "snapshot_data")
}

func newAdminOrderEvidenceHandlerStorage(t *testing.T) storage.StorageService {
	t.Helper()
	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: t.TempDir(),
		BaseURL:   "http://evidence.test",
	})
	require.NoError(t, err)
	return storageService
}

func newAdminEvidenceMultipartRequest(
	t *testing.T,
	method string,
	path string,
	filename string,
) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	require.NoError(t, png.Encode(part, image.NewRGBA(image.Rect(0, 0, 4, 4))))
	require.NoError(t, writer.Close())

	request := httptest.NewRequestWithContext(context.Background(), method, path, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func newAdminEvidencePNGFileHeader(t *testing.T, filename string) *multipart.FileHeader {
	t.Helper()
	request := newAdminEvidenceMultipartRequest(t, http.MethodPost, "/", filename)
	require.NoError(t, request.ParseMultipartForm(2<<20))
	files := request.MultipartForm.File["file"]
	require.Len(t, files, 1)
	return files[0]
}

func readAdminEvidenceFileHeader(t *testing.T, file *multipart.FileHeader) []byte {
	t.Helper()
	src, err := file.Open()
	require.NoError(t, err)
	defer func() { _ = src.Close() }()
	var body bytes.Buffer
	_, err = body.ReadFrom(src)
	require.NoError(t, err)
	return body.Bytes()
}

func completeAdminEvidencePackageForHandlerTest(
	t *testing.T,
	adminService *service.OrderEvidenceAdminService,
	attachmentService *service.OrderEvidenceAttachmentService,
	evidenceRepo *repository.OrderEvidenceRepository,
	orderID uint,
) {
	t.Helper()
	result, err := adminService.GetPackage(orderID)
	require.NoError(t, err)
	for _, item := range result.Package.Items {
		if item.ItemType == orderevidence.EvidenceItemTypeConfigurationConfirmation {
			continue
		}
		if item.ItemType == orderevidence.EvidenceItemTypeProductIdentity ||
			item.ItemType == orderevidence.EvidenceItemTypeOutboundWeightPackaging ||
			item.ItemType == orderevidence.EvidenceItemTypeSignedPOD {
			_, err := attachmentService.RegisterReference(
				repository.TxRepositories{
					OrderEvidence: evidenceRepo,
				},
				service.OrderEvidenceAttachmentReferenceInput{
					EvidenceItemID:   item.ID,
					StorageKey:       "order-evidence/1/" + string(item.ItemType) + ".jpg",
					OriginalFilename: string(item.ItemType) + ".jpg",
					MimeType:         "image/jpeg",
					SizeBytes:        12,
					SHA256:           "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
					UploadedBy:       7,
				},
			)
			require.NoError(t, err)
		}
		_, err := adminService.UpdateItem(orderID, service.OrderEvidenceAdminItemUpdateInput{
			ItemID:     item.ID,
			Status:     orderevidence.EvidenceItemStatusComplete,
			DataJSON:   handlerEvidenceCompletionData(item.ItemType),
			CapturedBy: 7,
		})
		require.NoError(t, err)
	}
}

func handlerEvidenceCompletionData(itemType string) []byte {
	switch itemType {
	case orderevidence.EvidenceItemTypeOutboundWeightPackaging:
		return []byte(`{
			"schema_version": 1,
			"gross_weight_g": 12640,
			"package_count": 2,
			"packaging_method": "double wall carton"
		}`)
	case orderevidence.EvidenceItemTypeSignedPOD:
		return []byte(`{
			"schema_version": 1,
			"tracking_number": "TRACK-100",
			"delivered_at": "2026-09-05T10:30:00Z"
		}`)
	default:
		return []byte(`{"verified":true}`)
	}
}

func newOrderEvidenceServicesForHandlerTest(
	db *gorm.DB,
	evidenceRepo *repository.OrderEvidenceRepository,
	storageService storage.StorageService,
) (*service.OrderEvidenceAdminService, *service.OrderEvidenceAttachmentService) {
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
	adminService := service.NewOrderEvidenceAdminService(
		txManager,
		orderRepo,
		evidenceRepo,
		service.NewOrderEvidenceService(),
	)
	attachmentService := service.NewConfiguredOrderEvidenceAttachmentService(
		txManager,
		evidenceRepo,
		storageService,
	)
	return adminService, attachmentService
}

func newAdminOrderEvidenceHandlerFixture(t *testing.T) (*gorm.DB, uint, uint, uint) {
	t.Helper()
	db := newAdminOrderEvidenceHandlerDB(t)
	firstOrderID, firstItemID := seedAdminEvidenceOrderForHandler(t, db, "TZ-2026-HANDLER-FIRST")
	_, secondItemID := seedAdminEvidenceOrderForHandler(t, db, "TZ-2026-HANDLER-SECOND")
	return db, firstOrderID, firstItemID, secondItemID
}

func newAdminOrderEvidenceHandlerDB(t *testing.T) *gorm.DB {
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
		&order.OrderItem{},
		&orderevidence.OrderEvidenceSnapshot{},
		&orderevidence.OrderEvidencePackage{},
		&orderevidence.OrderEvidenceItem{},
		&orderevidence.OrderEvidenceAttachment{},
		&orderevidence.OrderEvidenceExportSnapshot{},
		&shipping.TrackingProviderConfig{},
		&shipping.TrackingShipment{},
		&shipping.TrackingEvent{},
	))
	return db
}

func seedAdminEvidenceOrderForHandler(
	t *testing.T,
	db *gorm.DB,
	orderNumber string,
) (uint, uint) {
	t.Helper()
	record := &order.Order{
		OrderNumber:      orderNumber,
		TotalAmountMinor: 80000,
		Currency:         "USD",
		FXSnapshotData: currency.OrderFXSnapshotJSON(currency.OrderFXSnapshot{
			Version:       currency.OrderFXSnapshotVersion,
			BaseCurrency:  "USD",
			OrderCurrency: "USD",
			RateDecimal:   "1",
			Source:        "handler-test",
			CapturedAt:    time.Now().UTC(),
		}),
	}
	require.NoError(t, db.Create(record).Error)
	variantID := uint(22)
	item := order.OrderItem{
		OrderID:                   record.ID,
		ProductID:                 10,
		VariantID:                 &variantID,
		ProductName:               "Configured Product",
		SKU:                       orderNumber + "-SKU",
		Quantity:                  1,
		PriceMinor:                80000,
		WeightGrams:               9000,
		ConfigurationSnapshotData: datatypes.JSON([]byte(`{"finish":"black"}`)),
	}
	require.NoError(t, db.Create(&item).Error)
	record.Items = []order.OrderItem{item}
	snapshot, err := orderevidence.BuildOrderEvidenceSnapshot(record, []orderevidence.SnapshotItemInput{
		{Item: item},
	}, record.CreatedAt)
	require.NoError(t, err)
	require.NoError(t, db.Create(snapshot).Error)
	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	_, err = service.NewOrderEvidenceService().CreateInitialPackage(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		snapshot,
	)
	require.NoError(t, err)
	var pkg orderevidence.OrderEvidencePackage
	require.NoError(t, db.Where("order_id = ?", record.ID).First(&pkg).Error)
	var evidenceItems []orderevidence.OrderEvidenceItem
	require.NoError(t, db.Where("package_id = ? AND item_type = ?", pkg.ID, orderevidence.EvidenceItemTypeProductIdentity).
		Find(&evidenceItems).Error)
	require.Len(t, evidenceItems, 1)
	return record.ID, evidenceItems[0].ID
}

func newOrderEvidenceAdminServiceForHandlerTest(
	db *gorm.DB,
	evidenceRepo *repository.OrderEvidenceRepository,
) *service.OrderEvidenceAdminService {
	adminService, _ := newOrderEvidenceServicesForHandlerTest(db, evidenceRepo, nil)
	return adminService
}
