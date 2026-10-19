package product

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/api/middleware"
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/cache"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPublicCarbonRimCategoryLoadsTranslationsForProductDetailListsAndBreadcrumbs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&productdomain.ProductCategory{}, &productdomain.ProductCategoryTranslation{},
		&productdomain.ProductSpecificationTemplate{}, &productdomain.SpecDefinition{},
		&productdomain.Product{}, &productdomain.ProductMedia{},
		&productdomain.ProductSpecValue{}, &productdomain.ProductVariant{},
	))
	root := productdomain.ProductCategory{ID: 317, Name: "Wheel Components", Slug: "wheel-components", Depth: 1, IsEnabled: true}
	require.NoError(t, db.Create(&root).Error)
	category := productdomain.ProductCategory{ID: 975, ParentID: &root.ID, Name: "Carbon Rim", Slug: "carbon-rim", Depth: 2, IsEnabled: true}
	require.NoError(t, db.Create(&category).Error)
	require.NoError(t, db.Create(&[]productdomain.ProductCategoryTranslation{
		{ProductCategoryID: root.ID, Locale: "zh_cn", Name: "车轮配件"},
		{ProductCategoryID: category.ID, Locale: "en", Name: "Carbon Fiber Rims"},
		{ProductCategoryID: category.ID, Locale: "zh_cn", Name: "碳纤维车圈"},
	}).Error)
	tubelessReady := true
	for _, locale := range []string{"en", "zh_cn", "fr"} {
		item := productdomain.Product{
			Name: "R45", Slug: "r45-" + locale, Locale: locale, Status: "active", Currency: "USD",
			ProductType: productdomain.ProductTypeCarbonRim, ProductCategoryID: &category.ID,
			CarbonRimSpecification: &productdomain.CarbonRimSpecification{SchemaVersion: 1, Model: "R45", WheelSize: "700c", Material: "carbon_fiber", RidingUses: []string{"road"}, BeadSeatDiameterMM: 622, DepthMM: 45, InternalWidthMM: 25, ExternalWidthMM: 32, BeadStructure: "hooked", TubelessReady: &tubelessReady},
			Variants:               []productdomain.ProductVariant{{SKU: "R45-" + locale, PriceMinor: 10000, Currency: "USD", IsActive: true, IsDefault: true, Stock: 1, Weight: 450, CarbonRimSpecification: &productdomain.CarbonRimVariantSpecification{SpokeHoleCount: 24, DrillingPattern: "uniform", FinishCode: "matte-black", NetWeightGrams: 430}}},
		}
		require.NoError(t, db.Create(&item).Error)
	}
	redisServer := miniredis.RunT(t)
	redisCache, err := cache.Init(config.RedisConfig{Addrs: []string{redisServer.Addr()}})
	require.NoError(t, err)
	t.Cleanup(func() { _ = redisCache.Close() })
	productService := service.NewProductService(repository.NewProductRepository(db), redisCache, 60)
	handler := NewHandler(productService)
	handler.ConfigureProductCategoryService(service.NewProductCategoryService(repository.NewProductCategoryRepository(db)))
	router := gin.New()
	router.Use(middleware.I18n())
	router.GET("/products", handler.ListProducts)
	router.GET("/products/:id", handler.GetProduct)

	for _, scenario := range []struct{ locale, name string }{{"en", "Carbon Fiber Rims"}, {"zh_cn", "碳纤维车圈"}, {"fr", "Carbon Rim"}} {
		t.Run(scenario.locale, func(t *testing.T) {
			requestPublicProduct := func(path string, envelope any) {
				t.Helper()
				request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
				request.Header.Set("Accept-Language", scenario.locale)
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)
				require.Equal(t, http.StatusOK, response.Code, response.Body.String())
				require.NotContains(t, response.Body.String(), `"category_translations"`)
				require.NoError(t, json.Unmarshal(response.Body.Bytes(), envelope))
			}
			var detail struct{ Data PublicProduct }
			requestPublicProduct("/products/r45-"+scenario.locale, &detail)
			require.NotNil(t, detail.Data.Category)
			require.Equal(t, uint(975), detail.Data.Category.ID)
			require.Equal(t, "carbon-rim", detail.Data.Category.Slug)
			require.Equal(t, scenario.name, detail.Data.Category.Name)
			require.Nil(t, detail.Data.ProductSpecificationTemplate)
			require.NotNil(t, detail.Data.CarbonRimSpecification)
			require.Equal(t, "700c", detail.Data.CarbonRimSpecification.WheelSize)
			require.Equal(t, "carbon_fiber", detail.Data.CarbonRimSpecification.Material)
			require.Equal(t, []string{"road"}, detail.Data.CarbonRimSpecification.RidingUses)
			projectedFacts := map[string]string{}
			for _, value := range detail.Data.SpecValues {
				projectedFacts[value.Definition.Slug] = value.Value
			}
			require.NotEmpty(t, projectedFacts["wheel_size"])
			require.NotEmpty(t, projectedFacts["material"])
			require.NotEmpty(t, projectedFacts["riding_uses"])
			require.NotNil(t, detail.Data.Breadcrumb)
			require.Equal(t, "ready", detail.Data.Breadcrumb.Status)
			require.Equal(t, scenario.name, detail.Data.Breadcrumb.Items[3].Name)
			require.True(t, redisServer.Exists("product:slug:r45-"+scenario.locale+":"+scenario.locale))
			require.NoError(t, db.Model(&productdomain.Product{}).Where("slug = ?", "r45-"+scenario.locale).Update("name", "Database change after cache population").Error)
			var cachedDetail struct{ Data PublicProduct }
			requestPublicProduct("/products/r45-"+scenario.locale, &cachedDetail)
			require.NotNil(t, cachedDetail.Data.Category)
			require.Equal(t, "R45", cachedDetail.Data.Name, "the second response must come from Redis")
			require.Equal(t, scenario.name, cachedDetail.Data.Category.Name, "category translation must survive a Redis cache hit")
			for _, path := range []string{"/products", "/products?product_category=carbon-rim", "/products?product_category=wheel-components"} {
				var list struct{ Data []PublicProduct }
				requestPublicProduct(path, &list)
				require.Len(t, list.Data, 1)
				require.NotNil(t, list.Data[0].Category)
				require.Equal(t, scenario.name, list.Data[0].Category.Name)
			}
		})
	}
}
