package service

import (
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/post"
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSitemapIncludesPublishedProductsAndCategories(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&post.Post{},
		&productdomain.Product{},
		&productdomain.ProductVariant{},
		&productdomain.ProductCategory{},
	))

	now := time.Now().UTC()
	require.NoError(t, db.Create(&post.Post{
		Title: "Published guide", Slug: "published-guide", Status: "published", Locale: "en", AuthorID: 1,
		PublishedAt: &now, UpdatedAt: now,
	}).Error)
	category := &productdomain.ProductCategory{Name: "Road Wheels", Slug: "road-wheels", IsEnabled: true, Depth: 1, UpdatedAt: now}
	require.NoError(t, db.Create(category).Error)
	product := &productdomain.Product{
		Name: "Carbon Wheelset", Slug: "carbon-wheelset", Status: "active", Locale: "en", Currency: "USD", UpdatedAt: now,
		ProductCategoryID: &category.ID,
	}
	require.NoError(t, db.Create(product).Error)
	require.NoError(t, db.Create(&productdomain.ProductVariant{
		ProductID: product.ID, SKU: "CARBON-WS", Currency: "USD", IsActive: true,
	}).Error)

	sitemap := NewSitemapServiceWithCatalog(
		repository.NewPostRepository(db),
		repository.NewProductRepository(db),
		repository.NewProductCategoryRepository(db),
		"https://shop.example",
	)

	hreflang, err := sitemap.GenerateHreflangSitemap()
	require.NoError(t, err)
	for _, expected := range []string{
		"https://shop.example/resources/blog/published-guide",
		"https://shop.example/products/carbon-wheelset",
		"https://shop.example/shop/road-wheels",
	} {
		if !strings.Contains(hreflang, expected) {
			t.Fatalf("hreflang sitemap does not contain %q:\n%s", expected, hreflang)
		}
	}

	simple, err := sitemap.GenerateSimpleSitemap("en")
	require.NoError(t, err)
	for _, expected := range []string{
		"https://shop.example/resources/blog/published-guide",
		"https://shop.example/products/carbon-wheelset",
		"https://shop.example/shop/road-wheels",
	} {
		if !strings.Contains(simple, expected) {
			t.Fatalf("simple sitemap does not contain %q:\n%s", expected, simple)
		}
	}
}
