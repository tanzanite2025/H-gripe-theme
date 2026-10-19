package order

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/aftersales"
	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestCustomerAfterSalesAttachmentHTTPChecksCurrentOrderOwnerAndLocatorBeforeReading(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), DisableForeignKeyConstraintWhenMigrating: true, IgnoreRelationshipsWhenMigrating: true})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	for _, model := range []any{&orderdomain.Order{}, &aftersales.AfterSalesCase{}, &aftersales.AfterSalesCaseAttachment{}} {
		require.NoError(t, db.Migrator().CreateTable(model))
	}
	order := orderdomain.Order{OrderNumber: "CUSTOMER-ATTACHMENT-HTTP", UserID: 7, Currency: "USD"}
	require.NoError(t, db.Create(&order).Error)
	caseRecord := aftersales.AfterSalesCase{OrderID: order.ID, Type: "request", Status: "closed", Reason: "Historical attachment"}
	require.NoError(t, db.Create(&caseRecord).Error)
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), PrivateLocalPath: t.TempDir(), BaseURL: "https://shop.example.test"})
	require.NoError(t, err)
	reference, err := provider.(storage.PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(t.Context(), strings.NewReader("customer evidence"), "after-sales/pending/proof.png")
	require.NoError(t, err)
	attachment := aftersales.AfterSalesCaseAttachment{CaseID: caseRecord.ID, Kind: "image", StorageURL: reference, Filename: "proof.png", ContentType: "image/png", SizeBytes: 9999}
	require.NoError(t, db.Create(&attachment).Error)
	afterSalesService := service.NewAfterSalesService(repository.NewAfterSalesCaseRepository(db), repository.NewOrderRepository(db), nil)
	afterSalesService.ConfigureAttachmentStorage(provider)
	handler := &Handler{afterSalesService: afterSalesService}
	locator := service.CustomerAfterSalesAttachmentOpaqueLocator(attachment)
	for _, fixture := range []struct {
		name, orderNumber, locator string
		userID                     uint
		status                     int
	}{
		{"owner historical closed case", order.OrderNumber, locator, 7, http.StatusOK},
		{"other account", order.OrderNumber, locator, 8, http.StatusNotFound},
		{"other order", "OTHER-ORDER", locator, 7, http.StatusNotFound},
		{"wrong locator", order.OrderNumber, strings.Repeat("f", 64), 7, http.StatusNotFound},
		{"no identity", order.OrderNumber, locator, 0, http.StatusUnauthorized},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if fixture.userID > 0 {
					c.Set("user_id", fixture.userID)
				}
				c.Next()
			})
			router.GET("/orders/:order_number/after-sales/attachments/:attachment_locator", handler.ServeAfterSalesAttachmentForCurrentOrderOwner)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, fmt.Sprintf("/orders/%s/after-sales/attachments/%s", fixture.orderNumber, fixture.locator), nil))
			require.Equal(t, fixture.status, recorder.Code, recorder.Body.String())
			require.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
			require.NotContains(t, recorder.Body.String(), reference)
			if fixture.status == http.StatusOK {
				require.Equal(t, "customer evidence", recorder.Body.String())
				require.Equal(t, "17", recorder.Header().Get("Content-Length"))
			}
		})
	}
}
