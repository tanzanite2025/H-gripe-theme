package service

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/recommendation"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
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
	require.Equal(t, "trending_available", result.Items[0].Slot)
	require.Equal(t, "popular_available", result.Items[0].Reason)
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

func TestNormalizeRecommendationQueryKeepsUnicodeBoundary(t *testing.T) {
	query := strings.Repeat("轮", 200)
	normalized := normalizeRecommendationQuery(query)

	require.True(t, utf8.ValidString(normalized))
	require.Len(t, []rune(normalized), 160)
}

func TestRecommendationServicePrioritizesMatchingSpecsOnProductDetail(t *testing.T) {
	db, productService := newTestProductService(t)
	productTypeID, brakeTypeSpecID := seedRecommendationProductType(t, db)
	context := seedRecommendationTypedProduct(t, db, "context-wheel", "Context Wheel", productTypeID, 1)
	matching := seedRecommendationTypedProduct(t, db, "matching-wheel", "Matching Wheel", productTypeID, 2)
	other := seedRecommendationTypedProduct(t, db, "other-wheel", "Other Wheel", productTypeID, 20)
	seedRecommendationSpecValue(t, db, context.ID, brakeTypeSpecID, "disc")
	seedRecommendationSpecValue(t, db, matching.ID, brakeTypeSpecID, "disc")
	seedRecommendationSpecValue(t, db, other.ID, brakeTypeSpecID, "rim")

	recommendationService := NewRecommendationService(productService)
	result, err := recommendationService.Recommend(RecommendationRequest{
		Surface:   "product_detail_bottom",
		ProductID: &context.ID,
		Limit:     2,
	})

	require.NoError(t, err)
	require.Len(t, result.Items, 2)
	require.Equal(t, matching.ID, result.Items[0].ProductID)
	require.Equal(t, "similar_products", result.Items[0].Slot)
	require.Equal(t, "matching_specs", result.Items[0].Reason)
}

func TestRecommendationServiceUsesPersonalBehaviorSignals(t *testing.T) {
	db, productService := newTestProductService(t)
	require.NoError(t, db.AutoMigrate(&recommendation.Event{}))
	passive := seedRecommendationProduct(t, db, "passive-wheel", "Passive Wheel", "en", true, 8, false, 100)
	signaled := seedRecommendationProduct(t, db, "signaled-wheel", "Signaled Wheel", "en", true, 8, false, 1)
	now := time.Now().UTC()
	seedRecommendationEvent(t, db, "event_signaled_cart", "add_to_cart", "anon_behavior", signaled.ID, now)

	recommendationService := NewRecommendationService(productService, repository.NewRecommendationEventRepository(db))
	result, err := recommendationService.Recommend(RecommendationRequest{
		Surface:     "shop_index_bottom",
		AnonymousID: "anon_behavior",
		Limit:       2,
	})

	require.NoError(t, err)
	require.Len(t, result.Items, 2)
	require.Equal(t, signaled.ID, result.Items[0].ProductID)
	require.Equal(t, "personalized", result.Items[0].Slot)
	require.Equal(t, "your_recent_activity", result.Items[0].Reason)
	require.Equal(t, passive.ID, result.Items[1].ProductID)
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

func seedRecommendationProductType(t *testing.T, db *gorm.DB) (uint, uint) {
	t.Helper()

	productType := product.ProductType{
		Name:      "Wheelset",
		Slug:      "wheelset_recommendation_test",
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&productType).Error)

	brakeType := product.SpecDefinition{
		ProductTypeID: productType.ID,
		Group:         "Compatibility",
		Name:          "Brake Type",
		Slug:          "brake_type",
		FieldType:     "select",
		IsFilterable:  true,
		IsVisible:     true,
		SortOrder:     10,
	}
	require.NoError(t, db.Create(&brakeType).Error)
	return productType.ID, brakeType.ID
}

func seedRecommendationTypedProduct(
	t *testing.T,
	db *gorm.DB,
	slug string,
	name string,
	productTypeID uint,
	viewCount int,
) product.Product {
	t.Helper()

	item := seedRecommendationProduct(t, db, slug, name, "en", true, 8, false, viewCount)
	item.ProductTypeID = &productTypeID
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", item.ID).Update("product_type_id", productTypeID).Error)
	return item
}

func seedRecommendationSpecValue(t *testing.T, db *gorm.DB, productID uint, specDefinitionID uint, value string) {
	t.Helper()

	require.NoError(t, db.Create(&product.ProductSpecValue{
		ProductID:        productID,
		SpecDefinitionID: specDefinitionID,
		Value:            value,
	}).Error)
}

func seedRecommendationEvent(
	t *testing.T,
	db *gorm.DB,
	eventID string,
	eventType string,
	anonymousID string,
	productID uint,
	occurredAt time.Time,
) {
	t.Helper()

	require.NoError(t, db.Create(&recommendation.Event{
		EventID:      eventID,
		EventType:    eventType,
		AnonymousID:  anonymousID,
		ProductID:    &productID,
		MetadataJSON: datatypes.JSON([]byte(`{}`)),
		OccurredAt:   occurredAt,
		ReceivedAt:   occurredAt,
	}).Error)
}
