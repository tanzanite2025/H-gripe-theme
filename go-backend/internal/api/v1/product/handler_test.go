package product

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestGetProductAllowsNumericSlugWithoutIDLookup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&productdomain.ProductSpecificationTemplate{},
		&productdomain.ProductSpecificationTemplateTranslation{},
		&productdomain.SpecDefinition{},
		&productdomain.Product{},
		&productdomain.ProductMedia{},
		&productdomain.ProductSpecValue{},
		&productdomain.ProductVariant{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	item := productdomain.Product{
		SKU:    "NUMERIC-SLUG",
		Name:   "Numeric Slug Product",
		Slug:   "321313121",
		Status: "active",
		Locale: "en",
		Price:  99,
		Stock:  5,
		Variants: []productdomain.ProductVariant{
			{
				SKU:       "NUMERIC-SLUG-VAR",
				Title:     "Default",
				Price:     99,
				Stock:     5,
				IsActive:  true,
				IsDefault: true,
			},
		},
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed product: %v", err)
	}

	router := gin.New()
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products/:id", handler.GetProduct)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/products/321313121", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected numeric slug to return 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"slug":"321313121"`) {
		t.Fatalf("expected response to include numeric slug, got %s", response.Body.String())
	}
}

func TestListProductSpecificationTemplatesDoesNotExposeSpecDefinitions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&productdomain.ProductSpecificationTemplate{},
		&productdomain.ProductSpecificationTemplateTranslation{},
		&productdomain.SpecDefinition{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	productSpecificationTemplate := productdomain.ProductSpecificationTemplate{
		Name:      "Wheelset",
		Slug:      "wheelset",
		IsEnabled: true,
		Translations: []productdomain.ProductSpecificationTemplateTranslation{
			{Locale: "zh_cn", Name: "轮组"},
		},
		SpecDefinitions: []productdomain.SpecDefinition{
			{Name: "Material", Slug: "material", FieldType: "text", IsVisible: true},
		},
	}
	if err := db.Create(&productSpecificationTemplate).Error; err != nil {
		t.Fatalf("seed product specification template: %v", err)
	}

	router := gin.New()
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products/specification-templates", handler.ListProductSpecificationTemplates)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/products/specification-templates", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected product specification templates to return 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"slug":"wheelset"`) {
		t.Fatalf("expected response to include public product specification template, got %s", response.Body.String())
	}
	if strings.Contains(response.Body.String(), "spec_definitions") {
		t.Fatalf("public product specification template index exposes specs: %s", response.Body.String())
	}
}

func TestListCategoriesUsesRequestedLocaleAndFallsBackToDefaultName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&productdomain.ProductCategory{},
		&productdomain.ProductCategoryTranslation{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	category := productdomain.ProductCategory{
		Name:      "Wheel Parts",
		Slug:      "wheel-parts",
		Depth:     1,
		IsEnabled: true,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("seed category: %v", err)
	}
	if err := db.Create(&productdomain.ProductCategoryTranslation{
		ProductCategoryID: category.ID,
		Locale:            "zh_cn",
		Name:              "轮组部件",
	}).Error; err != nil {
		t.Fatalf("seed category translation: %v", err)
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("locale", "zh_cn")
		c.Next()
	})
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	handler.ConfigureProductCategoryService(
		service.NewProductCategoryService(repository.NewProductCategoryRepository(db)),
	)
	router.GET("/products/categories", handler.ListCategories)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/products/categories", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected categories to return 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"name":"轮组部件"`) {
		t.Fatalf("expected translated category name, got %s", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"locale":"zh_cn"`) {
		t.Fatalf("expected response locale, got %s", response.Body.String())
	}
}
