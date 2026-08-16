package service

import (
	"testing"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
)

func TestProductServiceAdminProductCategoryBindingCanBeCleared(t *testing.T) {
	db, productService := newTestProductService(t)
	require.NoError(t, db.AutoMigrate(&product.ProductCategory{}))
	productService.ConfigureProductCategoryRepository(repository.NewProductCategoryRepository(db))

	category := product.ProductCategory{
		Name:      "Wheelsets",
		Slug:      "wheelsets",
		Depth:     1,
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&category).Error)

	created, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductCategoryID: &category.ID,
		Name:              "Carbon Wheelset",
		Slug:              "carbon-wheelset",
		Status:            "active",
		Locale:            "en",
		Variants: []ProductVariantInput{{
			SKU:       "WHEELSET-001",
			Price:     499,
			Stock:     2,
			IsDefault: true,
			IsActive:  boolPtr(true),
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, created.ProductCategoryID)
	require.Equal(t, category.ID, *created.ProductCategoryID)
	require.NotNil(t, created.ProductCategory)

	cleared, err := productService.UpdateAdminProduct(created.ID, ProductUpdateInput{
		UpdateProductCategoryID: true,
		ProductCategoryID:       nil,
	})
	require.NoError(t, err)
	require.Nil(t, cleared.ProductCategoryID)
	require.Nil(t, cleared.ProductCategory)

	var saved product.Product
	require.NoError(t, db.First(&saved, created.ID).Error)
	require.Nil(t, saved.ProductCategoryID)
}

func TestProductServiceSearchPublicFiltersByCategorySubtree(t *testing.T) {
	db, productService := newTestProductService(t)
	require.NoError(t, db.AutoMigrate(&product.ProductCategory{}))
	productService.ConfigureProductCategoryRepository(repository.NewProductCategoryRepository(db))

	parent := product.ProductCategory{
		Name:      "Wheels",
		Slug:      "wheels",
		Depth:     1,
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&parent).Error)

	child := product.ProductCategory{
		ParentID:  &parent.ID,
		Name:      "Road Wheels",
		Slug:      "road-wheels",
		Depth:     2,
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&child).Error)

	sibling := product.ProductCategory{
		Name:      "Tires",
		Slug:      "tires",
		Depth:     1,
		IsEnabled: true,
	}
	require.NoError(t, db.Create(&sibling).Error)

	create := func(categoryID uint, sku, slug string) *product.Product {
		categoryIDCopy := categoryID
		created, err := productService.CreateAdminProduct(ProductCreateInput{
			ProductCategoryID: &categoryIDCopy,
			Name:              sku,
			Slug:              slug,
			Status:            "active",
			Locale:            "en",
			Variants: []ProductVariantInput{{
				SKU:       sku + "-VAR",
				Price:     100,
				Stock:     2,
				IsDefault: true,
				IsActive:  boolPtr(true),
			}},
		})
		require.NoError(t, err)
		return created
	}

	parentProduct := create(parent.ID, "WHEELS-001", "wheels-product")
	childProduct := create(child.ID, "ROAD-WHEELS-001", "road-wheels-product")
	siblingProduct := create(sibling.ID, "TIRES-001", "tires-product")

	results, total, err := productService.SearchPublic(ProductSearchInput{
		Locale:       "en",
		CategorySlug: "wheels",
		Page:         1,
		PageSize:     20,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, results, 2)
	require.ElementsMatch(t, []uint{parentProduct.ID, childProduct.ID}, []uint{results[0].ID, results[1].ID})
	require.NotContains(t, []uint{results[0].ID, results[1].ID}, siblingProduct.ID)
}
