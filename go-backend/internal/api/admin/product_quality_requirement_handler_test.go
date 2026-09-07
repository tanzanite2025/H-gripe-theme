package admin

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"commerce-platform/internal/domain/auth"
	productdomain "commerce-platform/internal/domain/product"
	productrequirement "commerce-platform/internal/domain/productrequirement"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProductQualityRequirementHandlerRejectsForeignVariant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, firstProductID, secondVariantID := newProductQualityRequirementHandlerFixture(t)
	handler := NewProductQualityRequirementHandler(
		service.NewProductQualityRequirementService(
			repository.NewProductQualityRequirementRepository(db),
		),
	)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(
		http.MethodPut,
		"/api/admin/products/"+strconv.FormatUint(uint64(firstProductID), 10)+"/fulfillment-requirements/spoke-tension-qc",
		bytes.NewBufferString(`{"variant_id":`+strconv.FormatUint(uint64(secondVariantID), 10)+`,"spoke_tension_qc_required":true,"rule_version":"variant-v1"}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")
	context.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(firstProductID), 10)}}
	context.Set("user_id", uint(7))

	handler.UpsertSpokeTensionQC(context)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestProductQualityRequirementRoutesSeparateViewAndEditPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", string(auth.RoleViewer))
		c.Set("user_id", uint(7))
		c.Next()
	})
	registerProductQualityRequirementRoutes(router.Group(""), NewProductQualityRequirementHandler(nil))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPut,
		"/products/1/fulfillment-requirements/spoke-tension-qc",
		bytes.NewBufferString(`{"spoke_tension_qc_required":true,"rule_version":"v1"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func newProductQualityRequirementHandlerFixture(t *testing.T) (*gorm.DB, uint, uint) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&productdomain.Product{},
		&productdomain.ProductVariant{},
		&productrequirement.ProductQualityRequirementRule{},
	))

	first := productdomain.Product{
		SKU:      "SKU-REQUIREMENT-HANDLER-FIRST",
		Name:     "Requirement Handler First",
		Slug:     "requirement-handler-first",
		Currency: "USD",
		Price:    100,
		Stock:    5,
	}
	second := productdomain.Product{
		SKU:      "SKU-REQUIREMENT-HANDLER-SECOND",
		Name:     "Requirement Handler Second",
		Slug:     "requirement-handler-second",
		Currency: "USD",
		Price:    100,
		Stock:    5,
	}
	require.NoError(t, db.Create(&first).Error)
	require.NoError(t, db.Create(&second).Error)

	variant := productdomain.ProductVariant{
		ProductID:    second.ID,
		SKU:          "SKU-REQUIREMENT-HANDLER-SECOND-VARIANT",
		Title:        "Second variant",
		OptionValues: "{}",
		Currency:     "USD",
		Price:        100,
		Stock:        5,
		Weight:       9000,
		IsDefault:    true,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&variant).Error)
	return db, first.ID, variant.ID
}
