package service

import (
	"testing"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSEOResourceServiceValidatesCanonicalURLs(t *testing.T) {
	service := &SEOResourceService{}
	service.ConfigureCanonicalBaseURL("https://store.example.test")

	require.NoError(t, service.validateCanonicalURL("https://store.example.test/blog/release"))
	require.ErrorIs(t, service.validateCanonicalURL("http://store.example.test/blog/release"), ErrInvalidSEOCanonicalURL)
	require.ErrorIs(t, service.validateCanonicalURL("https://other.example.test/blog/release"), ErrInvalidSEOCanonicalURL)
	require.ErrorIs(t, service.validateCanonicalURL("https://store.example.test/blog/release?ref=home"), ErrInvalidSEOCanonicalURL)
	require.NoError(t, service.validateCanonicalURL(""))
}

func TestSEOResourceServiceCanonicalizesProductDiagnosticImages(t *testing.T) {
	service := NewSEOResourceService(nil, nil)
	service.ConfigureMediaService(NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30))

	diagnostics, err := service.ProductDiagnostics(product.Product{
		Name:     "Carbon Wheel",
		Slug:     "carbon-wheel",
		Locale:   "en",
		Status:   "active",
		Price:    199,
		Currency: "USD",
		Stock:    1,
		Media: []product.ProductMedia{{
			MediaType: "image",
			IsVisible: true,
			URL:       "http://media.internal:8080/uploads/products/carbon-wheel.jpg",
		}},
	})

	require.NoError(t, err)
	require.Equal(
		t,
		[]string{"https://shop.example.test/uploads/products/carbon-wheel.jpg"},
		diagnostics.StructuredData.Image,
	)
	require.Equal(t, "unavailable", diagnostics.Breadcrumb.Status)
	require.Equal(t, "category_service_unavailable", diagnostics.BreadcrumbReason)
	require.Equal(t, "blocked", diagnostics.BreadcrumbSSR)
}

func TestSEOResourceServiceProjectsProductBreadcrumbDiagnostics(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&product.ProductCategory{},
		&product.ProductCategoryTranslation{},
	))

	root := product.ProductCategory{
		Name:      "Road Wheels",
		Slug:      "road-wheels",
		Depth:     1,
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&root).Error)
	child := product.ProductCategory{
		ParentID:  &root.ID,
		Name:      "Climbing",
		Slug:      "climbing",
		Depth:     2,
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&child).Error)
	require.NoError(t, db.Create(&[]product.ProductCategoryTranslation{
		{ProductCategoryID: root.ID, Locale: "zh_cn", Name: "公路轮组"},
		{ProductCategoryID: child.ID, Locale: "zh_cn", Name: "爬坡"},
	}).Error)

	categoryService := NewProductCategoryService(repository.NewProductCategoryRepository(db))
	seoService := NewSEOResourceService(nil, nil, categoryService)
	diagnostics, err := seoService.ProductDiagnostics(product.Product{
		ID:                41,
		Locale:            "zh_cn",
		Name:              "Hyper 45",
		Slug:              "hyper-45",
		Status:            "active",
		ProductCategoryID: &child.ID,
	})
	require.NoError(t, err)
	require.Equal(t, "ready", diagnostics.Breadcrumb.Status)
	require.True(t, diagnostics.BreadcrumbComplete)
	require.Equal(t, "ready", diagnostics.BreadcrumbSSR)
	require.Empty(t, diagnostics.BreadcrumbReason)
	require.NotNil(t, diagnostics.PrimaryCategory)
	require.Equal(t, "/zh_cn/shop/road-wheels", diagnostics.PrimaryCategory.Path)
	require.Equal(t, "/zh_cn/products/hyper-45", diagnostics.Breadcrumb.Items[len(diagnostics.Breadcrumb.Items)-1].Path)
	require.Equal(t, []string{"Home", "Shop", "公路轮组", "爬坡", "Hyper 45"}, []string{
		diagnostics.Breadcrumb.Items[0].Name,
		diagnostics.Breadcrumb.Items[1].Name,
		diagnostics.Breadcrumb.Items[2].Name,
		diagnostics.Breadcrumb.Items[3].Name,
		diagnostics.Breadcrumb.Items[4].Name,
	})
}
