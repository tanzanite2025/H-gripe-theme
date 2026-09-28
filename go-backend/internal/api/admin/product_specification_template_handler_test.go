package admin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDeleteSystemManagedProductSpecificationTemplateReturnsConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&product.ProductSpecificationTemplate{},
		&product.SpecDefinition{},
		&product.ProductSpecOptionItem{},
	))

	template := product.ProductSpecificationTemplate{
		Name:            "Schwalbe Tire",
		Slug:            "schwalbe_tire",
		IsEnabled:       true,
		IsSystemManaged: true,
	}
	require.NoError(t, db.Create(&template).Error)

	handler := NewProductHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router := gin.New()
	router.DELETE("/api/admin/product-specification-templates/:id", handler.DeleteProductSpecificationTemplate)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/api/admin/product-specification-templates/"+fmt.Sprint(template.ID), nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusConflict, recorder.Code, recorder.Body.String())
	var persisted product.ProductSpecificationTemplate
	require.NoError(t, db.First(&persisted, template.ID).Error)
	require.True(t, persisted.IsSystemManaged)
}
