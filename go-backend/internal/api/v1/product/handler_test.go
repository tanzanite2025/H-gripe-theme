package product

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	productdomain "tanzanite/internal/domain/product"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

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
		&productdomain.ProductType{},
		&productdomain.ProductTypeTranslation{},
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
