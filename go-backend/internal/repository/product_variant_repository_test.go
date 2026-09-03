package repository

import (
	"regexp"
	"testing"

	"commerce-platform/internal/domain/product"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/glebarez/sqlite"
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
	require.Equal(t, 0, storedProduct.Stock)

	_, err = repo.DecrementVariantStocks(map[uint]int{variant.ID: 1})
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient stock")

	require.NoError(t, db.First(&storedVariant, variant.ID).Error)
	require.Equal(t, 0, storedVariant.Stock)
}

func TestProductRepositoryDecrementVariantStocksUsesStableLockOrder(t *testing.T) {
	repo, mock, cleanup := newMockProductVariantRepository(t)
	defer cleanup()

	variantUpdateSQL := regexp.QuoteMeta(`UPDATE "product_variants" SET "stock"=stock - $1 WHERE (id = $2 AND is_active = $3 AND stock >= $4) AND "product_variants"."deleted_at" IS NULL`)
	mock.ExpectExec(variantUpdateSQL).
		WithArgs(1, uint(10), true, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(variantUpdateSQL).
		WithArgs(2, uint(20), true, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT "product_id" FROM "product_variants" WHERE id IN ($1,$2) AND "product_variants"."deleted_at" IS NULL`)).
		WithArgs(uint(10), uint(20)).
		WillReturnRows(sqlmock.NewRows([]string{"product_id"}).
			AddRow(uint(200)).
			AddRow(uint(100)))

	productStockQuerySQL := regexp.QuoteMeta(`SELECT COALESCE(SUM(stock), 0) FROM "product_variants" WHERE (product_id = $1 AND is_active = $2 AND deleted_at IS NULL) AND "product_variants"."deleted_at" IS NULL`)
	productUpdateSQL := regexp.QuoteMeta(`UPDATE "products" SET "stock"=$1,"updated_at"=$2 WHERE id = $3 AND "products"."deleted_at" IS NULL`)
	mock.ExpectQuery(productStockQuerySQL).
		WithArgs(uint(100), true).
		WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(int64(3)))
	mock.ExpectExec(productUpdateSQL).
		WithArgs(int64(3), sqlmock.AnyArg(), uint(100)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(productStockQuerySQL).
		WithArgs(uint(200), true).
		WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(int64(4)))
	mock.ExpectExec(productUpdateSQL).
		WithArgs(int64(4), sqlmock.AnyArg(), uint(200)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	productIDs, err := repo.DecrementVariantStocks(map[uint]int{20: 2, 10: 1})
	require.NoError(t, err)
	require.Equal(t, []uint{100, 200}, productIDs)
	require.NoError(t, mock.ExpectationsWereMet())
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
