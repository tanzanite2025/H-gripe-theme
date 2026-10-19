package admin

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

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/warranty"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestPrivateOrderAttachmentActualAdministratorRoutesRetainProductAndOrderReadPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&warranty.WarrantyClaim{}, &warranty.WarrantyServiceRecord{}))
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), PrivateLocalPath: t.TempDir(), BaseURL: "https://shop.example.test"})
	require.NoError(t, err)
	reference, err := provider.(storage.PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(t.Context(), strings.NewReader("admin private evidence"), "warranty/proof.png")
	require.NoError(t, err)
	claim := warranty.WarrantyClaim{OrderNumber: "PRIVATE-ADMIN-READ", Images: fmt.Sprintf("[%q]", reference)}
	require.NoError(t, db.Create(&claim).Error)
	warrantyService := service.NewWarrantyService(nil, repository.NewWarrantyRepository(db), repository.NewOrderRepository(db))
	warrantyService.ConfigurePrivateOrderBusinessAttachmentUploads(service.NewPrivateOrderBusinessAttachmentUploadService(nil, provider, "https://shop.example.test"))
	for _, role := range []string{"editor", "support", "manager", "admin"} {
		t.Run(role, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) { c.Set("user_role", role); c.Next() })
			group := router.Group("/api/admin")
			registerContentRoutes(group, nil, nil, nil, nil, nil, nil, nil, nil, nil, NewWarrantyHandler(warrantyService), NewShipmentRecordHandler(nil, nil), nil, nil, nil, nil, nil, nil)
			registerCommerceRoutes(group, nil, NewAfterSalesHandler(nil), nil, nil, nil, nil, nil, nil)
			for _, fixture := range []struct {
				path       string
				permission auth.Permission
			}{
				{fmt.Sprintf("/api/admin/warranty/claims/%d/attachments/images/0", claim.ID), auth.PermProductView},
				{"/api/admin/warranty/shipment-records/0/images/0", auth.PermProductView},
				{"/api/admin/after-sales/0/attachments/0", auth.PermOrderView},
			} {
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, fixture.path, nil))
				if !auth.Role(role).HasPermission(fixture.permission) {
					require.Equal(t, http.StatusForbidden, recorder.Code)
				} else if strings.Contains(fixture.path, "/claims/") {
					require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
				} else {
					require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
				}
			}
		})
	}
}
