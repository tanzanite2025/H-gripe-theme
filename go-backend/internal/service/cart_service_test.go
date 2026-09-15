package service

import (
	"errors"
	"sync"
	"testing"

	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPurchasablePriceStockUsesMoneyCurrencyAndRejectsInvalidCurrency(t *testing.T) {
	variant := &product.ProductVariant{ID: 7, Price: 101, Currency: "JPY", Stock: 3}
	price, stock, resolvedVariantID, err := purchasablePriceStock(variant)
	require.NoError(t, err)
	require.Equal(t, int64(101), price.AmountMinor())
	require.Equal(t, domainmoney.MustNew(101, "JPY").Currency(), price.Currency())
	require.Equal(t, 3, stock)
	require.NotNil(t, resolvedVariantID)
	require.Equal(t, uint(7), *resolvedVariantID)

	variant.Currency = "XXX"
	_, _, _, err = purchasablePriceStock(variant)
	require.EqualError(t, err, "product price currency is invalid")
}

func TestCartServiceRecoversWhenCreateLosesIdentityRace(t *testing.T) {
	db, cartService := newConcurrentTestCartService(t)
	userID := uint(42)
	var injected sync.Once

	db.Callback().Query().After("gorm:query").Register("test_cart_identity_race", func(tx *gorm.DB) {
		if tx.Statement.Table != "carts" {
			return
		}
		injected.Do(func() {
			require.NoError(t, db.Exec(
				"INSERT INTO carts (user_id, session_id) VALUES (?, ?)",
				userID,
				"race-winner",
			).Error)
		})
	})

	cart, err := cartService.GetOrCreateCart(&userID, "race-loser")
	require.NoError(t, err)
	require.NotNil(t, cart)
	assert.Equal(t, userID, *cart.UserID)
	assert.Equal(t, "race-winner", cart.SessionID)

	var cartCount int64
	require.NoError(t, db.Model(&product.Cart{}).Where("user_id = ?", userID).Count(&cartCount).Error)
	assert.EqualValues(t, 1, cartCount)
}

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
	assert.Equal(t, int64(10000), itemsByVariant[blackVariant.ID].PriceMinor)
	assert.Equal(t, 2, itemsByVariant[whiteVariant.ID].Quantity)
	assert.Equal(t, int64(12000), itemsByVariant[whiteVariant.ID].PriceMinor)
}

func TestCartServiceAllowsMadeToOrderVariantWithoutStock(t *testing.T) {
	db, cartService := newTestCartService(t)

	productRecord := product.Product{
		SKU:             "MTO-CART",
		Name:            "Made To Order Cart Product",
		Slug:            "mto-cart",
		FulfillmentMode: product.FulfillmentModeMadeToOrder,
		Price:           999,
		Stock:           0,
	}
	require.NoError(t, db.Create(&productRecord).Error)

	variant := product.ProductVariant{
		ProductID: productRecord.ID,
		SKU:       "MTO-CART-VAR",
		Price:     999,
		Stock:     0,
		IsDefault: true,
		IsActive:  true,
	}
	require.NoError(t, db.Create(&variant).Error)

	cartRecord := product.Cart{SessionID: "mto-cart-session"}
	require.NoError(t, db.Create(&cartRecord).Error)

	require.NoError(t, cartService.AddToCart(cartRecord.ID, productRecord.ID, &variant.ID, 3))
	require.NoError(t, cartService.UpdateCartItem(cartRecord.ID, productRecord.ID, &variant.ID, 5))

	summary, err := repository.NewCartRepository(db).GetSummary(cartRecord.ID)
	require.NoError(t, err)
	require.Len(t, summary.Items, 1)
	assert.Equal(t, 5, summary.Items[0].Quantity)
}

func TestCartSummaryReadDoesNotCreateMissingAnonymousCart(t *testing.T) {
	db, cartService := newTestCartService(t)

	summary, err := cartService.GetCartSummary(nil, "")
	require.NoError(t, err)
	require.NotNil(t, summary)
	assert.Zero(t, summary.ItemCount)
	assert.Zero(t, summary.TotalMoney.AmountMinor())
	assert.Empty(t, summary.Items)

	var cartCount int64
	require.NoError(t, db.Model(&product.Cart{}).Count(&cartCount).Error)
	assert.EqualValues(t, 0, cartCount)
}

func TestCartSummarySelfHealsInvalidVariantItems(t *testing.T) {
	db, _ := newTestCartService(t)

	firstProduct := product.Product{SKU: "SELF-HEAL-1", Name: "Self Heal One", Slug: "self-heal-one", Price: 10, Stock: 10}
	secondProduct := product.Product{SKU: "SELF-HEAL-2", Name: "Self Heal Two", Slug: "self-heal-two", Price: 10, Stock: 10}
	require.NoError(t, db.Create(&firstProduct).Error)
	require.NoError(t, db.Create(&secondProduct).Error)
	validVariant := product.ProductVariant{ProductID: firstProduct.ID, SKU: "SELF-HEAL-VALID", OptionValues: `{"color":"valid"}`, Price: 10, Stock: 5, IsActive: true}
	inactiveVariant := product.ProductVariant{ProductID: firstProduct.ID, SKU: "SELF-HEAL-INACTIVE", OptionValues: `{"color":"inactive"}`, Price: 10, Stock: 5, IsActive: false}
	deletedVariant := product.ProductVariant{ProductID: firstProduct.ID, SKU: "SELF-HEAL-DELETED", OptionValues: `{"color":"deleted"}`, Price: 10, Stock: 5, IsActive: true}
	mismatchVariant := product.ProductVariant{ProductID: firstProduct.ID, SKU: "SELF-HEAL-MISMATCH", OptionValues: `{"color":"mismatch"}`, Price: 10, Stock: 5, IsActive: true}
	require.NoError(t, db.Create(&[]product.ProductVariant{validVariant, inactiveVariant, deletedVariant, mismatchVariant}).Error)
	require.NoError(t, db.Model(&product.ProductVariant{}).Where("sku = ?", inactiveVariant.SKU).Update("is_active", false).Error)
	require.NoError(t, db.Where("sku = ?", deletedVariant.SKU).First(&deletedVariant).Error)
	require.NoError(t, db.Where("sku = ?", deletedVariant.SKU).Delete(&product.ProductVariant{}).Error)
	require.NoError(t, db.Where("sku = ?", validVariant.SKU).First(&validVariant).Error)
	require.NoError(t, db.Where("sku = ?", inactiveVariant.SKU).First(&inactiveVariant).Error)
	require.NoError(t, db.Where("sku = ?", mismatchVariant.SKU).First(&mismatchVariant).Error)
	var danglingVariantID uint = 999999
	var mismatchVariantID = mismatchVariant.ID
	cart := product.Cart{SessionID: "self-heal-cart"}
	require.NoError(t, db.Create(&cart).Error)
	require.NoError(t, db.Create(&[]product.CartItem{
		{CartID: cart.ID, ProductID: firstProduct.ID, VariantID: &validVariant.ID, Quantity: 1, PriceMinor: 1000},
		{CartID: cart.ID, ProductID: firstProduct.ID, VariantID: &inactiveVariant.ID, Quantity: 1, PriceMinor: 1000},
		{CartID: cart.ID, ProductID: firstProduct.ID, VariantID: &deletedVariant.ID, Quantity: 1, PriceMinor: 1000},
		{CartID: cart.ID, ProductID: firstProduct.ID, VariantID: &danglingVariantID, Quantity: 1, PriceMinor: 1000},
		{CartID: cart.ID, ProductID: secondProduct.ID, VariantID: &mismatchVariantID, Quantity: 1, PriceMinor: 1000},
	}).Error)

	summary, err := repository.NewCartRepository(db.Debug()).GetSummary(cart.ID)
	require.NoError(t, err)
	require.Len(t, summary.Items, 1)
	require.NotNil(t, summary.Items[0].VariantID)
	assert.Equal(t, validVariant.ID, *summary.Items[0].VariantID)
}

func TestAnonymousCartCreationRequiresSessionID(t *testing.T) {
	_, cartService := newTestCartService(t)

	cart, err := cartService.GetOrCreateCart(nil, "")
	require.Nil(t, cart)
	assert.True(t, errors.Is(err, ErrCartNotFound))
}

func newTestCartService(t *testing.T) (*gorm.DB, *CartService) {
	return newTestCartServiceWithDSN(t, ":memory:", 1)
}

func newConcurrentTestCartService(t *testing.T) (*gorm.DB, *CartService) {
	return newTestCartServiceWithDSN(t, "file:cart-service-race?mode=memory&cache=shared", 4)
}

func newTestCartServiceWithDSN(t *testing.T, dsn string, maxOpenConns int) (*gorm.DB, *CartService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		TranslateError: true,
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(
		&product.ProductSpecificationTemplate{},
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
