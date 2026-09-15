package repository

import (
	"testing"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/product"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRestoreConsumedOrderItemsToOriginalCheckoutCartUsesAtomicVariantUpsert(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, db.AutoMigrate(&product.Cart{}, &product.CartItem{}))

	userID := uint(42)
	variantID := uint(7)
	cart := product.Cart{UserID: &userID}
	require.NoError(t, db.Create(&cart).Error)
	require.NoError(t, db.Create(&product.CartItem{
		CartID:     cart.ID,
		ProductID:  3,
		VariantID:  &variantID,
		Quantity:   1,
		PriceMinor: 9000,
		Currency:   "USD",
	}).Error)

	repo := NewCartRepository(db)
	err = repo.RestoreConsumedOrderItemsToOriginalCheckoutCart(
		cart.ID,
		[]order.OrderItem{{
			ProductID: 3,
			VariantID: &variantID,
			Quantity:  2,
			Price:     100,
		}},
		"USD",
	)
	require.NoError(t, err)

	var restoredItems []product.CartItem
	require.NoError(t, db.Where("cart_id = ?", cart.ID).Find(&restoredItems).Error)
	require.Len(t, restoredItems, 1)
	require.Equal(t, 3, restoredItems[0].Quantity)
	require.Equal(t, int64(10000), restoredItems[0].PriceMinor)
	require.Equal(t, "USD", restoredItems[0].Currency)
}

func TestCartRepositoryGetSummaryUsesCurrencyMinorUnits(t *testing.T) {
	db := newCartRepositorySummaryTestDB(t)
	cart := product.Cart{SessionID: "summary-jpy"}
	require.NoError(t, db.Create(&cart).Error)
	variantID := uint(1)
	require.NoError(t, db.Create(&product.CartItem{
		CartID:     cart.ID,
		ProductID:  1,
		VariantID:  &variantID,
		Quantity:   2,
		PriceMinor: 10000,
		Currency:   "JPY",
	}).Error)

	summary, err := NewCartRepository(db).GetSummary(cart.ID)
	require.NoError(t, err)
	require.Equal(t, 2, summary.ItemCount)
	require.Equal(t, int64(20000), summary.TotalMoney.AmountMinor())
}

func TestCartRepositoryGetSummaryRejectsInvalidMoneyRows(t *testing.T) {
	tests := []struct {
		name       string
		priceMinor int64
		currency   string
		quantity   int
		errText    string
	}{
		{name: "invalid currency", priceMinor: 1000, currency: "XXX", quantity: 1, errText: "price"},
		{name: "negative amount", priceMinor: -1, currency: "USD", quantity: 1, errText: "price"},
		{name: "invalid quantity", priceMinor: 1000, currency: "USD", quantity: 0, errText: "quantity"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newCartRepositorySummaryTestDB(t)
			cart := product.Cart{SessionID: "summary-invalid-" + tt.name}
			require.NoError(t, db.Create(&cart).Error)
			variantID := uint(1)
			item := &product.CartItem{
				CartID:     cart.ID,
				ProductID:  1,
				VariantID:  &variantID,
				Quantity:   tt.quantity,
				PriceMinor: tt.priceMinor,
				Currency:   tt.currency,
			}
			require.NoError(t, db.Create(item).Error)

			summary, err := NewCartRepository(db).GetSummary(cart.ID)
			require.Nil(t, summary)
			require.ErrorContains(t, err, tt.errText)
		})
	}
}

func TestCartRepositoryGetSummaryRejectsMixedCurrencies(t *testing.T) {
	db := newCartRepositorySummaryTestDB(t)
	cart := product.Cart{SessionID: "summary-mixed"}
	require.NoError(t, db.Create(&cart).Error)
	variantID := uint(1)
	secondVariantID := uint(2)
	require.NoError(t, db.Create(&[]product.CartItem{
		{CartID: cart.ID, ProductID: 1, VariantID: &variantID, Quantity: 1, PriceMinor: 1000, Currency: "USD"},
		{CartID: cart.ID, ProductID: 1, VariantID: &secondVariantID, Quantity: 1, PriceMinor: 1000, Currency: "EUR"},
	}).Error)

	summary, err := NewCartRepository(db).GetSummary(cart.ID)
	require.Nil(t, summary)
	require.ErrorContains(t, err, "mixed currencies")
}

func newCartRepositorySummaryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&product.Product{},
		&product.ProductMedia{},
		&product.ProductVariant{},
		&product.Cart{},
		&product.CartItem{},
	))
	productRecord := product.Product{SKU: "SUMMARY-PRODUCT", Name: "Summary Product", Slug: "summary-product", Price: 10}
	require.NoError(t, db.Create(&productRecord).Error)
	variants := []product.ProductVariant{
		{ProductID: productRecord.ID, SKU: "SUMMARY-VARIANT-1", OptionValues: `{"slot":"one"}`, Price: 10, IsActive: true},
		{ProductID: productRecord.ID, SKU: "SUMMARY-VARIANT-2", OptionValues: `{"slot":"two"}`, Price: 10, IsActive: true},
	}
	require.NoError(t, db.Create(&variants).Error)
	require.Equal(t, uint(1), productRecord.ID)
	require.Equal(t, uint(1), variants[0].ID)
	require.Equal(t, uint(2), variants[1].ID)
	return db
}
