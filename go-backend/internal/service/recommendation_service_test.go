package service

import (
	"testing"

	"tanzanite/internal/domain/product"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRecommendationServiceReturnsAvailableProductsAndRespectsExclusions(t *testing.T) {
	db, productService := newTestProductService(t)
	available := seedRecommendationProduct(t, db, "available", "Available Wheel", "en", true, 8, true, 20)
	featured := seedRecommendationProduct(t, db, "featured", "Featured Wheel", "en", true, 3, true, 1)
	seedRecommendationProduct(t, db, "out-of-stock", "Out of Stock Wheel", "en", true, 0, false, 200)
	seedRecommendationProduct(t, db, "inactive", "Inactive Wheel", "en", false, 12, false, 300)

	recommendationService := NewRecommendationService(productService)
	result, err := recommendationService.Recommend(RecommendationRequest{
		Surface:           "shop_search_drawer",
		Locale:            "en-US",
		Limit:             6,
		ExcludeProductIDs: []uint{available.ID},
	})

	require.NoError(t, err)
	require.NotEmpty(t, result.RequestID)
	require.Equal(t, RecommendationAlgorithmVersion, result.AlgorithmVersion)
	require.Len(t, result.Items, 1)
	require.Equal(t, featured.ID, result.Items[0].ProductID)
	require.Equal(t, "rule_fallback", result.Items[0].Slot)
	require.Equal(t, "available_global", result.Items[0].Reason)
}

func TestRecommendationServiceUsesEnglishFallbackWhenLocaleHasNoProducts(t *testing.T) {
	db, productService := newTestProductService(t)
	englishProduct := seedRecommendationProduct(t, db, "english-only", "English Wheel", "en", true, 2, false, 1)

	recommendationService := NewRecommendationService(productService)
	result, err := recommendationService.Recommend(RecommendationRequest{
		Surface: "homepage",
		Locale:  "fr",
		Limit:   1,
	})

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	require.Equal(t, englishProduct.ID, result.Items[0].ProductID)
}

func TestRecommendationServiceValidatesSurfaceAndLimit(t *testing.T) {
	recommendationService := NewRecommendationService(&ProductService{})

	_, err := recommendationService.Recommend(RecommendationRequest{Limit: 1})
	require.ErrorIs(t, err, ErrRecommendationSurfaceInvalid)

	_, err = recommendationService.Recommend(RecommendationRequest{
		Surface: "homepage",
		Limit:   MaxRecommendationLimit + 1,
	})
	require.ErrorIs(t, err, ErrRecommendationLimitInvalid)
}

func seedRecommendationProduct(
	t *testing.T,
	db *gorm.DB,
	slug string,
	name string,
	locale string,
	active bool,
	stock int,
	featured bool,
	viewCount int,
) product.Product {
	t.Helper()

	status := "active"
	if !active {
		status = "inactive"
	}
	item := product.Product{
		SKU:       slug + "-sku",
		Name:      name,
		Slug:      slug,
		Status:    status,
		Locale:    locale,
		Featured:  featured,
		ViewCount: viewCount,
		Price:     399,
	}
	require.NoError(t, db.Create(&item).Error)

	variant := product.ProductVariant{
		ProductID: item.ID,
		SKU:       slug + "-variant",
		Title:     name,
		Price:     399,
		Stock:     stock,
		IsActive:  active,
		IsDefault: true,
	}
	require.NoError(t, db.Create(&variant).Error)

	return item
}
