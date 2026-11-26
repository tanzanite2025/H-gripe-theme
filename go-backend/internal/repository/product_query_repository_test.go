package repository

import (
	"fmt"
	"testing"
	"time"

	"commerce-platform/internal/domain/product"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSearchPublicUsesStableIDTieBreakerForPagination(t *testing.T) {
	db := newProductQueryTestDB(t)
	repo := NewProductRepository(db)
	updatedAt := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

	products := seedPublicProductsWithSameUpdatedAt(t, db, updatedAt, 3)

	firstPage, total, err := repo.SearchPublic(ProductSearchQuery{
		Locale: "en",
		Status: "active",
		Offset: 0,
		Limit:  2,
	})
	require.NoError(t, err)
	require.Equal(t, int64(3), total)

	secondPage, total, err := repo.SearchPublic(ProductSearchQuery{
		Locale: "en",
		Status: "active",
		Offset: 2,
		Limit:  2,
	})
	require.NoError(t, err)
	require.Equal(t, int64(3), total)

	requireProductIDs(t, firstPage, products[2].ID, products[1].ID)
	requireProductIDs(t, secondPage, products[0].ID)
}

func TestSearchPublicCompactUsesStableIDTieBreakerForPagination(t *testing.T) {
	db := newProductQueryTestDB(t)
	repo := NewProductRepository(db)
	updatedAt := time.Date(2026, 9, 3, 13, 0, 0, 0, time.UTC)

	products := seedPublicProductsWithSameUpdatedAt(t, db, updatedAt, 3)

	firstPage, err := repo.SearchPublicCompact(ProductSearchQuery{
		Locale: "en",
		Status: "active",
		Offset: 0,
		Limit:  2,
	})
	require.NoError(t, err)

	secondPage, err := repo.SearchPublicCompact(ProductSearchQuery{
		Locale: "en",
		Status: "active",
		Offset: 2,
		Limit:  2,
	})
	require.NoError(t, err)

	requireProductIDs(t, firstPage, products[2].ID, products[1].ID)
	requireProductIDs(t, secondPage, products[0].ID)
}

func TestListQuickBuyCandidatesIncludesMadeToOrderProductsWithoutStock(t *testing.T) {
	db := newProductQueryTestDB(t)
	repo := NewProductRepository(db)

	item := product.Product{
		SKU:             "MTO-QUICK-BUY",
		Name:            "Made To Order Quick Buy",
		Slug:            "mto-quick-buy",
		Status:          "active",
		Locale:          "en",
		FulfillmentMode: product.FulfillmentModeMadeToOrder,
	}
	require.NoError(t, db.Create(&item).Error)
	require.NoError(t, db.Create(&product.ProductVariant{
		ProductID: item.ID,
		SKU:       "MTO-QUICK-BUY-VAR",
		IsActive:  true,
		IsDefault: true,
		Stock:     0,
	}).Error)

	results, total, err := repo.ListQuickBuyCandidates(ProductQuickBuyCandidateQuery{
		Locale: "en",
		Offset: 0,
		Limit:  10,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	requireProductIDs(t, results, item.ID)
}

func TestListQuickBuyCandidatesFiltersWheelsetsByDedicatedFrontAndRearSpecifications(t *testing.T) {
	db := newProductQueryTestDB(t)
	repo := NewProductRepository(db)

	mixedDiameter := seedDedicatedWheelsetFilterProduct(t, db, "Mullet Wheelset", "mullet-wheelset", "27.5")
	matchingDiameters := seedDedicatedWheelsetFilterProduct(t, db, "29 Inch Wheelset", "29-inch-wheelset", "29")
	matchingDiameters.WheelsetSpecification.RearWheel.FreehubBody = "hg"
	require.NoError(t, db.Save(&matchingDiameters).Error)

	testCases := []struct {
		name       string
		filters    map[string][]string
		expectedID []uint
	}{
		{
			name:       "legacy whole wheelset size excludes mixed diameters",
			filters:    map[string][]string{product.WheelsetFilterKeyWheelSize: {"29"}},
			expectedID: []uint{matchingDiameters.ID},
		},
		{
			name:       "legacy multi-value wheel size still requires one value on both wheels",
			filters:    map[string][]string{product.WheelsetFilterKeyWheelSize: {"27.5", "29"}},
			expectedID: []uint{matchingDiameters.ID},
		},
		{
			name:       "front wheel size reads front wheel",
			filters:    map[string][]string{product.WheelsetFilterKeyFrontWheelSize: {"29"}},
			expectedID: []uint{mixedDiameter.ID, matchingDiameters.ID},
		},
		{
			name:       "rear wheel size reads rear wheel",
			filters:    map[string][]string{product.WheelsetFilterKeyRearWheelSize: {"27.5"}},
			expectedID: []uint{mixedDiameter.ID},
		},
		{
			name:       "front axle reads front wheel",
			filters:    map[string][]string{product.WheelsetFilterKeyFrontAxle: {"15x110"}},
			expectedID: []uint{mixedDiameter.ID, matchingDiameters.ID},
		},
		{
			name:       "rear axle reads rear wheel",
			filters:    map[string][]string{product.WheelsetFilterKeyRearAxle: {"12x148"}},
			expectedID: []uint{mixedDiameter.ID, matchingDiameters.ID},
		},
		{
			name:       "brake mount requires both wheels",
			filters:    map[string][]string{product.WheelsetFilterKeyBrakeMount: {"center_lock"}},
			expectedID: []uint{mixedDiameter.ID, matchingDiameters.ID},
		},
		{
			name:       "freehub reads rear wheel",
			filters:    map[string][]string{product.WheelsetFilterKeyFreehub: {"micro_spline"}},
			expectedID: []uint{mixedDiameter.ID},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			results, _, err := repo.ListQuickBuyCandidates(ProductQuickBuyCandidateQuery{
				Locale:      "en",
				SpecFilters: testCase.filters,
				Limit:       10,
			})
			require.NoError(t, err)
			actualIDs := productIDs(results)
			require.ElementsMatch(t, testCase.expectedID, actualIDs)
		})
	}
}

func TestListWheelsetCategoryFilterOptionsReadsDedicatedProductSpecifications(t *testing.T) {
	db := newProductQueryTestDB(t)
	repo := NewProductRepository(db)
	require.NoError(t, db.AutoMigrate(&product.ProductCategory{}))

	category := product.ProductCategory{
		Name:      "Wheelsets",
		Slug:      product.ProductTypeWheelset,
		Depth:     1,
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&category).Error)

	emptyFilters, err := repo.ListFilterableSpecificationsWithDynamicValuesForCategory(product.ProductTypeWheelset)
	require.NoError(t, err)
	require.Len(t, emptyFilters, len(product.WheelsetProductLevelSpecificationFilterDefinitions()))
	for _, filter := range emptyFilters {
		require.Empty(t, filter.Values)
	}

	seedDedicatedWheelsetFilterProduct(t, db, "Mullet Wheelset", "mullet-wheelset", "27.5", category.ID)
	seedDedicatedWheelsetFilterProduct(t, db, "29 Inch Wheelset", "29-inch-wheelset", "29", category.ID)

	filters, err := repo.ListFilterableSpecificationsWithDynamicValuesForCategory(product.ProductTypeWheelset)
	require.NoError(t, err)
	require.Len(t, filters, len(product.WheelsetProductLevelSpecificationFilterDefinitions()))
	valuesBySlug := make(map[string][]string, len(filters))
	for _, filter := range filters {
		valuesBySlug[filter.Slug] = filter.Values
	}
	require.Equal(t, []string{"29"}, valuesBySlug[product.WheelsetFilterKeyWheelSize])
	require.Equal(t, []string{"29"}, valuesBySlug[product.WheelsetFilterKeyFrontWheelSize])
	require.Equal(t, []string{"27.5", "29"}, valuesBySlug[product.WheelsetFilterKeyRearWheelSize])
	require.Equal(t, []string{"15x110"}, valuesBySlug[product.WheelsetFilterKeyFrontAxle])
	require.Equal(t, []string{"12x148"}, valuesBySlug[product.WheelsetFilterKeyRearAxle])
	require.Equal(t, []string{"center_lock"}, valuesBySlug[product.WheelsetFilterKeyBrakeMount])
	require.Equal(t, []string{"micro_spline"}, valuesBySlug[product.WheelsetFilterKeyFreehub])
}

func seedDedicatedWheelsetFilterProduct(t *testing.T, db *gorm.DB, name, slug, rearWheelSize string, productCategoryIDs ...uint) product.Product {
	t.Helper()
	var categoryID *uint
	if len(productCategoryIDs) > 0 {
		categoryID = &productCategoryIDs[0]
	}
	item := product.Product{
		ProductType:       product.ProductTypeWheelset,
		ProductCategoryID: categoryID,
		WheelsetSpecification: &product.WheelsetSpecification{
			SchemaVersion: 1,
			FrontWheel: product.WheelsetProductWheelSpecification{
				WheelSize:       "29",
				SpokeHoleCounts: []int{24},
				LacingPatterns:  []string{"2-cross"},
				AxleStandard:    "15x110",
				BrakeInterface:  "center_lock",
			},
			RearWheel: product.WheelsetProductWheelSpecification{
				WheelSize:       rearWheelSize,
				SpokeHoleCounts: []int{24},
				LacingPatterns:  []string{"2-cross"},
				AxleStandard:    "12x148",
				BrakeInterface:  "center_lock",
				FreehubBody:     "micro_spline",
			},
		},
		Name:   name,
		Slug:   slug,
		Status: "active",
		Locale: "en",
	}
	require.NoError(t, db.Create(&item).Error)
	require.NoError(t, db.Create(&product.ProductVariant{
		ProductID:  item.ID,
		SKU:        slug + "-variant",
		IsActive:   true,
		IsDefault:  true,
		Stock:      1,
		PriceMinor: 1000,
	}).Error)
	return item
}

func TestListQuickBuyCandidatesExcludesSpokeRepairKitProducts(t *testing.T) {
	db := newProductQueryTestDB(t)
	repo := NewProductRepository(db)
	require.NoError(t, db.AutoMigrate(&product.ProductCategory{}))

	category := product.ProductCategory{
		Name:      "Spoke Repair Kits",
		Slug:      product.SpokeRepairKitProductCategorySlug,
		Depth:     2,
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&category).Error)

	repairKit := product.Product{
		Name:              "Repair Kit",
		Slug:              "repair-kit",
		Status:            "active",
		Locale:            "en",
		ProductCategoryID: &category.ID,
	}
	require.NoError(t, db.Create(&repairKit).Error)
	require.NoError(t, db.Create(&product.ProductVariant{
		ProductID:  repairKit.ID,
		SKU:        "REPAIR-KIT-VAR",
		IsActive:   true,
		IsDefault:  true,
		Stock:      10,
		PriceMinor: 1000,
	}).Error)

	regular := product.Product{
		Name:   "Regular Product",
		Slug:   "regular-product",
		Status: "active",
		Locale: "en",
	}
	require.NoError(t, db.Create(&regular).Error)
	require.NoError(t, db.Create(&product.ProductVariant{
		ProductID:  regular.ID,
		SKU:        "REGULAR-VAR",
		IsActive:   true,
		IsDefault:  true,
		Stock:      10,
		PriceMinor: 1000,
	}).Error)

	results, total, err := repo.ListQuickBuyCandidates(ProductQuickBuyCandidateQuery{
		Locale: "en",
		Offset: 0,
		Limit:  10,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	requireProductIDs(t, results, regular.ID)
}

func newProductQueryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(
		&product.ProductBrand{},
		&product.ProductSpecificationTemplate{},
		&product.SpecDefinition{},
		&product.CustomsClassificationProfile{},
		&product.ProductInformationTemplate{},
		&product.Product{},
		&product.ProductMedia{},
		&product.ProductSpecValue{},
		&product.ProductVariant{},
	))
	return db
}

func seedPublicProductsWithSameUpdatedAt(t *testing.T, db *gorm.DB, updatedAt time.Time, count int) []product.Product {
	t.Helper()

	products := make([]product.Product, 0, count)
	for index := 0; index < count; index++ {
		item := product.Product{
			SKU:        fmt.Sprintf("STABLE-PAGE-%d", index),
			Name:       fmt.Sprintf("Stable Page %d", index),
			Slug:       fmt.Sprintf("stable-page-%d", index),
			Status:     "active",
			Locale:     "en",
			PriceMinor: 10000,
			Stock:      1,
			CreatedAt:  updatedAt,
			UpdatedAt:  updatedAt,
		}
		require.NoError(t, db.Create(&item).Error)
		require.NoError(t, db.Create(&product.ProductVariant{
			ProductID:  item.ID,
			SKU:        fmt.Sprintf("STABLE-PAGE-%d-VAR", index),
			Title:      "Default",
			PriceMinor: 10000,
			Stock:      1,
			IsActive:   true,
			IsDefault:  true,
			CreatedAt:  updatedAt,
			UpdatedAt:  updatedAt,
		}).Error)
		products = append(products, item)
	}
	require.NoError(t, db.Model(&product.Product{}).Where("id IN ?", productIDs(products)).UpdateColumn("updated_at", updatedAt).Error)
	return products
}

func productIDs(products []product.Product) []uint {
	ids := make([]uint, 0, len(products))
	for _, item := range products {
		ids = append(ids, item.ID)
	}
	return ids
}

func requireProductIDs(t *testing.T, products []product.Product, expectedIDs ...uint) {
	t.Helper()

	actualIDs := make([]uint, 0, len(products))
	for _, item := range products {
		actualIDs = append(actualIDs, item.ID)
	}
	require.Equal(t, expectedIDs, actualIDs)
}
