package service

import (
	"testing"

	"tanzanite/internal/domain/product"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCartServiceKeepsProductVariantsAsSeparateLines(t *testing.T) {
	db, cartService := newTestCartService(t)

	productRecord := product.Product{
		SKU:   "RIM-CART",
		Name:  "Cart Rim",
		Slug:  "cart-rim",
		Price: 999,
		Stock: 99,
	}
	require.NoError(t, db.Create(&productRecord).Error)

	blackVariant := product.ProductVariant{
		ProductID:    productRecord.ID,
		SKU:          "RIM-CART-BLK-24H",
		OptionValues: `{"color":"black","spoke_holes":"24"}`,
		Price:        100,
		Stock:        5,
		IsDefault:    true,
		IsActive:     true,
	}
	whiteVariant := product.ProductVariant{
		ProductID:    productRecord.ID,
		SKU:          "RIM-CART-WHT-28H",
		OptionValues: `{"color":"white","spoke_holes":"28"}`,
		Price:        120,
		Stock:        4,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&[]product.ProductVariant{blackVariant, whiteVariant}).Error)
	require.NoError(t, db.Where("sku = ?", blackVariant.SKU).First(&blackVariant).Error)
	require.NoError(t, db.Where("sku = ?", whiteVariant.SKU).First(&whiteVariant).Error)

	cartRecord := product.Cart{SessionID: "cart-variant-test"}
	require.NoError(t, db.Create(&cartRecord).Error)

	require.NoError(t, cartService.AddToCart(cartRecord.ID, productRecord.ID, &blackVariant.ID, 1))
	require.NoError(t, cartService.AddToCart(cartRecord.ID, productRecord.ID, &whiteVariant.ID, 2))
	require.NoError(t, cartService.AddToCart(cartRecord.ID, productRecord.ID, &blackVariant.ID, 1))

	summary, err := repository.NewCartRepository(db).GetSummary(cartRecord.ID)
	require.NoError(t, err)
	require.Len(t, summary.Items, 2)

	itemsByVariant := make(map[uint]product.CartItem, len(summary.Items))
	for _, item := range summary.Items {
		require.NotNil(t, item.VariantID)
		itemsByVariant[*item.VariantID] = item
	}

	assert.Equal(t, 2, itemsByVariant[blackVariant.ID].Quantity)
	assert.InDelta(t, 100, itemsByVariant[blackVariant.ID].Price, 0.001)
	assert.Equal(t, 2, itemsByVariant[whiteVariant.ID].Quantity)
	assert.InDelta(t, 120, itemsByVariant[whiteVariant.ID].Price, 0.001)
}

func newTestCartService(t *testing.T) (*gorm.DB, *CartService) {
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
		&product.ProductType{},
		&product.SpecDefinition{},
		&product.Product{},
		&product.ProductMedia{},
		&product.ProductSpecValue{},
		&product.ProductVariant{},
		&product.Cart{},
		&product.CartItem{},
	))

	return db, NewCartService(repository.NewCartRepository(db), repository.NewProductRepository(db))
}
