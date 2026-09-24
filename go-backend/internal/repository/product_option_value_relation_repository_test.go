package repository

import (
	"testing"

	productdomain "commerce-platform/internal/domain/product"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestProductOptionValueRelationsPersistReloadAndNormalizeConflicts(t *testing.T) {
	db := newProductOptionValueRelationRepositoryDB(t)
	owner := createProductOptionValueRelationRepositoryProduct(t, db, "relation-owner", "REL-OWNER")
	values := createProductOptionValueRelationRepositoryValues(t, db, owner.ID, "alpha", "beta", "gamma")

	require.NoError(t, replaceProductOptionValueRelations(db, owner.ID, []productdomain.ProductOptionValueRelation{
		{SourceOptionValueID: values[2].ID, TargetOptionValueID: values[0].ID, RelationType: productdomain.OptionValueRelationConflicts},
		{SourceOptionValueID: values[0].ID, TargetOptionValueID: values[2].ID, RelationType: productdomain.OptionValueRelationConflicts},
		{SourceOptionValueID: values[0].ID, TargetOptionValueID: values[1].ID, RelationType: productdomain.OptionValueRelationRequires},
	}))

	var reloaded productdomain.Product
	repo := NewProductRepository(db)
	require.NoError(t, repo.preloadProductOptionValueRelations(db).First(&reloaded, owner.ID).Error)
	require.Len(t, reloaded.OptionValueRelations, 2)
	require.Equal(t, productdomain.OptionValueRelationConflicts, reloaded.OptionValueRelations[0].RelationType)
	require.Equal(t, values[0].ID, reloaded.OptionValueRelations[0].SourceOptionValueID)
	require.Equal(t, values[2].ID, reloaded.OptionValueRelations[0].TargetOptionValueID)
	require.NotZero(t, reloaded.OptionValueRelations[0].ID)
	require.Equal(t, productdomain.OptionValueRelationRequires, reloaded.OptionValueRelations[1].RelationType)
}

func TestProductOptionValueRelationRepositoryRejectsInvalidOwnershipAndSelfRelation(t *testing.T) {
	db := newProductOptionValueRelationRepositoryDB(t)
	owner := createProductOptionValueRelationRepositoryProduct(t, db, "relation-owner", "REL-OWNER")
	other := createProductOptionValueRelationRepositoryProduct(t, db, "relation-other", "REL-OTHER")
	ownerValues := createProductOptionValueRelationRepositoryValues(t, db, owner.ID, "owner-alpha")
	otherValues := createProductOptionValueRelationRepositoryValues(t, db, other.ID, "other-alpha")

	_, err := normalizeProductOptionValueRelations(db, owner.ID, []productdomain.ProductOptionValueRelation{{
		SourceOptionValueID: ownerValues[0].ID, TargetOptionValueID: ownerValues[0].ID, RelationType: productdomain.OptionValueRelationRequires,
	}})
	require.ErrorIs(t, err, ErrProductOptionValueRelationInvalid)

	_, err = normalizeProductOptionValueRelations(db, owner.ID, []productdomain.ProductOptionValueRelation{{
		SourceOptionValueID: ownerValues[0].ID, TargetOptionValueID: otherValues[0].ID, RelationType: productdomain.OptionValueRelationRequires,
	}})
	require.ErrorIs(t, err, ErrProductOptionValueRelationInvalid)
}

func newProductOptionValueRelationRepositoryDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&productdomain.Product{},
		&productdomain.ProductVariantOptionValue{},
		&productdomain.ProductOptionValueRelation{},
	))
	return db
}

func createProductOptionValueRelationRepositoryProduct(t *testing.T, db *gorm.DB, slug, sku string) productdomain.Product {
	t.Helper()
	item := productdomain.Product{Name: slug, Slug: slug, SKU: sku, Currency: "USD", Status: "active", Locale: "en"}
	require.NoError(t, db.Create(&item).Error)
	return item
}

func createProductOptionValueRelationRepositoryValues(t *testing.T, db *gorm.DB, productID uint, keys ...string) []productdomain.ProductVariantOptionValue {
	t.Helper()
	values := make([]productdomain.ProductVariantOptionValue, 0, len(keys))
	for index, key := range keys {
		values = append(values, productdomain.ProductVariantOptionValue{
			ProductID: productID, SpecDefinitionID: 1, ValueKey: key, Label: key, SortOrder: index, IsEnabled: true,
		})
	}
	require.NoError(t, db.Create(&values).Error)
	return values
}
