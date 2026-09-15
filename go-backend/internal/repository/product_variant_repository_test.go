package repository

import (
	"regexp"
	"testing"

	"commerce-platform/internal/domain/product"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProductRepositoryDecrementVariantStocksUsesConditionalUpdate(t *testing.T) {
	db := newProductVariantTestDB(t)
	repo := NewProductRepository(db)

	p := &product.Product{
		SKU:      "LIMITED-RIM",
		Name:     "Limited Rim",
		Slug:     "limited-rim",
		Currency: "USD",
		Price:    199.99,
		Stock:    1,
	}
	require.NoError(t, repo.Create(p))

	variant := product.ProductVariant{
		ProductID: p.ID,
		SKU:       "LIMITED-RIM-01",
		Title:     "Limited Rim 01",
		Currency:  "USD",
		Price:     199.99,
		Stock:     1,
		IsActive:  true,
		IsDefault: true,
	}
	require.NoError(t, db.Create(&variant).Error)

	productIDs, err := repo.DecrementVariantStocks(map[uint]int{variant.ID: 1})
	require.NoError(t, err)
	require.Len(t, productIDs, 1)
	require.Equal(t, p.ID, productIDs[0])

	var storedVariant product.ProductVariant
	require.NoError(t, db.First(&storedVariant, variant.ID).Error)
	require.Equal(t, 0, storedVariant.Stock)

	var storedProduct product.Product
	require.NoError(t, db.First(&storedProduct, p.ID).Error)
	require.Equal(t, 1, storedProduct.Stock, "inventory writes must not update the product compatibility summary")

	_, err = repo.DecrementVariantStocks(map[uint]int{variant.ID: 1})
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient stock")

	require.NoError(t, db.First(&storedVariant, variant.ID).Error)
	require.Equal(t, 0, storedVariant.Stock)
}

func TestProductRepositoryDecrementVariantStocksUsesStableLockOrder(t *testing.T) {
	repo, mock, cleanup := newMockProductVariantRepository(t)
	defer cleanup()

	variantLookupSQL := `SELECT id, master_variant_id, is_active FROM "product_variants" WHERE id = \$1.*LIMIT \$2`
	mock.ExpectQuery(variantLookupSQL).
		WithArgs(uint(10), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "master_variant_id", "is_active"}).AddRow(uint(10), nil, true))
	mock.ExpectQuery(variantLookupSQL).
		WithArgs(uint(20), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "master_variant_id", "is_active"}).AddRow(uint(20), nil, true))

	variantUpdateSQL := regexp.QuoteMeta(`UPDATE "product_variants" SET "stock"=stock - $1 WHERE (id = $2 AND is_active = $3 AND stock >= $4) AND "product_variants"."deleted_at" IS NULL`)
	mock.ExpectExec(variantUpdateSQL).
		WithArgs(1, uint(10), true, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(variantUpdateSQL).
		WithArgs(2, uint(20), true, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT "product_id" FROM "product_variants" WHERE (id IN ($1,$2) OR master_variant_id IN ($3,$4)) AND "product_variants"."deleted_at" IS NULL`)).
		WithArgs(uint(10), uint(20), uint(10), uint(20)).
		WillReturnRows(sqlmock.NewRows([]string{"product_id"}).
			AddRow(uint(200)).
			AddRow(uint(100)))

	productIDs, err := repo.DecrementVariantStocks(map[uint]int{20: 2, 10: 1})
	require.NoError(t, err)
	require.Equal(t, []uint{100, 200}, productIDs)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepositoryTranslatedVariantUsesMasterInventory(t *testing.T) {
	db := newProductVariantTestDB(t)
	repo := NewProductRepository(db)

	root := &product.Product{SKU: "MASTER-RIM", Name: "Master Rim", Slug: "master-rim", Currency: "USD", Price: 100, Stock: 10}
	require.NoError(t, repo.Create(root))
	master := product.ProductVariant{ProductID: root.ID, SKU: "MASTER-RIM-VAR", Currency: "USD", Price: 100, Stock: 10, IsActive: true, IsDefault: true}
	require.NoError(t, db.Create(&master).Error)
	translated := &product.Product{SKU: "MASTER-RIM-FR", Name: "Master Rim FR", Slug: "master-rim-fr", Currency: "USD", Price: 100, Stock: 10}
	require.NoError(t, repo.Create(translated))
	masterID := master.ID
	localized := product.ProductVariant{ProductID: translated.ID, MasterVariantID: &masterID, SKU: "MASTER-RIM-VAR-fr", Currency: "USD", Price: 100, Stock: 5, IsActive: true, IsDefault: true}
	require.NoError(t, db.Create(&localized).Error)

	productIDs, err := repo.DecrementVariantStocks(map[uint]int{localized.ID: 3})
	require.NoError(t, err)
	assert.Contains(t, productIDs, root.ID)
	assert.Contains(t, productIDs, translated.ID)

	var storedMaster product.ProductVariant
	require.NoError(t, db.First(&storedMaster, master.ID).Error)
	assert.Equal(t, 7, storedMaster.Stock)
	var storedLocalized product.ProductVariant
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).First(&storedLocalized, localized.ID).Error)
	assert.Equal(t, 5, storedLocalized.Stock, "localized legacy stock must never be treated as inventory")

	var storedRootProduct, storedTranslatedProduct product.Product
	require.NoError(t, db.First(&storedRootProduct, root.ID).Error)
	require.NoError(t, db.First(&storedTranslatedProduct, translated.ID).Error)
	assert.Equal(t, 10, storedRootProduct.Stock)
	assert.Equal(t, 10, storedTranslatedProduct.Stock)
}

func TestProductRepositoryIncrementVariantStockDoesNotWriteProductSummary(t *testing.T) {
	db := newProductVariantTestDB(t)
	repo := NewProductRepository(db)

	record := &product.Product{SKU: "RESTOCK-RIM", Name: "Restock Rim", Slug: "restock-rim", Currency: "USD", Price: 100, Stock: 2}
	require.NoError(t, repo.Create(record))
	variant := product.ProductVariant{ProductID: record.ID, SKU: "RESTOCK-RIM-VAR", Currency: "USD", Price: 100, Stock: 2, IsActive: true, IsDefault: true}
	require.NoError(t, db.Create(&variant).Error)

	productIDs, err := repo.IncrementVariantStock(variant.ID, 3)
	require.NoError(t, err)
	require.Equal(t, []uint{record.ID}, productIDs)

	var storedVariant product.ProductVariant
	require.NoError(t, db.First(&storedVariant, variant.ID).Error)
	require.Equal(t, 5, storedVariant.Stock)
	var storedProduct product.Product
	require.NoError(t, db.First(&storedProduct, record.ID).Error)
	require.Equal(t, 2, storedProduct.Stock)
}

func TestSyncProductSummaryUsesLowestEffectiveActiveVariantPrice(t *testing.T) {
	salePrice := 75.0
	item := &product.Product{}
	variants := []product.ProductVariant{
		{SKU: "DEFAULT", Currency: "USD", Price: 100, Stock: 3, IsActive: true, IsDefault: true},
		{SKU: "SALE", Currency: "USD", Price: 120, SalePrice: &salePrice, Stock: 4, IsActive: true},
		{SKU: "INACTIVE-CHEAP", Currency: "USD", Price: 10, Stock: 8, IsActive: false},
	}

	syncProductSummaryFromVariants(item, variants)

	require.Equal(t, "DEFAULT", item.SKU)
	require.Equal(t, 120.0, item.Price)
	require.Equal(t, &salePrice, item.SalePrice)
	require.Equal(t, 7, item.Stock)
}

func newProductVariantTestDB(t *testing.T) *gorm.DB {
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

	require.NoError(t, db.AutoMigrate(&product.Product{}, &product.ProductVariant{}))
	return db
}

func newMockProductVariantRepository(t *testing.T) (*ProductRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		_ = sqlDB.Close()
	}
	require.NoError(t, err)

	return NewProductRepository(db), mock, func() {
		_ = sqlDB.Close()
	}
}
