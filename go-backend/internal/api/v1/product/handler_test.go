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

func TestListProductsFiltersFeaturedResultsByProductCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&productdomain.ProductCategory{},
		&productdomain.Product{},
		&productdomain.ProductMedia{},
		&productdomain.ProductVariant{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	wheelsetCategory := productdomain.ProductCategory{
		Name:      "Wheelsets",
		Slug:      "wheelset",
		Depth:     1,
		IsEnabled: true,
	}
	rimCategory := productdomain.ProductCategory{
		Name:      "Rims",
		Slug:      "rim",
		Depth:     1,
		IsEnabled: true,
	}
	if err := db.Create(&wheelsetCategory).Error; err != nil {
		t.Fatalf("seed wheelset category: %v", err)
	}
	if err := db.Create(&rimCategory).Error; err != nil {
		t.Fatalf("seed rim category: %v", err)
	}

	createProduct := func(categoryID uint, slug string, featured bool) {
		categoryIDCopy := categoryID
		item := productdomain.Product{
			ProductCategoryID: &categoryIDCopy,
			SKU:               strings.ToUpper(slug),
			Name:              slug,
			Slug:              slug,
			Status:            "active",
			Locale:            "en",
			Featured:          featured,
			Price:             100,
		}
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("seed product %s: %v", slug, err)
		}
		if err := db.Create(&productdomain.ProductVariant{
			ProductID: item.ID,
			SKU:       strings.ToUpper(slug) + "-VAR",
			Title:     "Default",
			Price:     100,
			Stock:     1,
			IsDefault: true,
			IsActive:  true,
		}).Error; err != nil {
			t.Fatalf("seed product variant %s: %v", slug, err)
		}
	}

	createProduct(wheelsetCategory.ID, "wheelset-featured", true)
	createProduct(wheelsetCategory.ID, "wheelset-regular", false)
	createProduct(rimCategory.ID, "rim-featured", true)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("locale", "en")
		c.Next()
	})
	handler := NewHandler(service.NewProductService(repository.NewProductRepository(db), nil, 0))
	router.GET("/products", handler.ListProducts)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/products?featured=true&product_category=wheelset", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected featured wheelsets response to return 200, got %d: %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, `"slug":"wheelset-featured"`) {
		t.Fatalf("expected featured wheelset in response, got %s", body)
	}
	if strings.Contains(body, `"slug":"wheelset-regular"`) || strings.Contains(body, `"slug":"rim-featured"`) {
		t.Fatalf("response ignored featured wheelset scope: %s", body)
	}
	if !strings.Contains(body, `"page":1`) || !strings.Contains(body, `"total":1`) {
		t.Fatalf("response omitted product pagination metadata: %s", body)
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
		&productdomain.SpecDefinition{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	productSpecificationTemplate := productdomain.ProductSpecificationTemplate{
		Name:      "Wheelset",
		Slug:      "wheelset",
		IsEnabled: true,
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
	if !strings.Contains(response.Body.String(), `"route_path":"/zh_cn/shop/wheel-parts"`) {
		t.Fatalf("expected category route_path, got %s", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"locale":"zh_cn"`) {
		t.Fatalf("expected response locale, got %s", response.Body.String())
	}
}
