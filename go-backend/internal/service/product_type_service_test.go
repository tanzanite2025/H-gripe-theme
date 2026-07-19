package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductServiceCreatesManagedProductType(t *testing.T) {
	_, productService := newTestProductService(t)

	created, err := productService.CreateProductType(ProductTypeInput{
		Name:        "首饰",
		Slug:        "jewelry",
		Description: "首饰类商品",
		SortOrder:   10,
		IsEnabled:   true,
		SpecDefinitions: []ProductSpecDefinitionInput{
			{
				Group:        "材质",
				Name:         "材质",
				Slug:         "material",
				FieldType:    "select",
				IsRequired:   true,
				IsFilterable: true,
				IsVisible:    true,
				SortOrder:    10,
				Options:      `["银","金","银"]`,
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "首饰", created.Name)
	assert.True(t, created.IsEnabled)
	require.Len(t, created.SpecDefinitions, 1)
	assert.JSONEq(t, `["银","金"]`, created.SpecDefinitions[0].Options)
	assert.True(t, created.SpecDefinitions[0].IsVisible)
}

func TestProductServiceUpdatesProductTypeAndReplacesSpecs(t *testing.T) {
	_, productService := newTestProductService(t)
	created, err := productService.CreateProductType(ProductTypeInput{
		Name:      "首饰",
		Slug:      "jewelry",
		IsEnabled: true,
		SpecDefinitions: []ProductSpecDefinitionInput{
			{Name: "材质", Slug: "material", FieldType: "text", IsVisible: true},
		},
	})
	require.NoError(t, err)

	updated, err := productService.UpdateProductType(created.ID, ProductTypeInput{
		Name:      "配饰",
		Slug:      "accessories",
		IsEnabled: false,
		SpecDefinitions: []ProductSpecDefinitionInput{
			{Name: "尺寸", Slug: "size", FieldType: "number", Unit: "mm", IsVisible: false},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "配饰", updated.Name)
	assert.Equal(t, "accessories", updated.Slug)
	assert.False(t, updated.IsEnabled)
	require.Len(t, updated.SpecDefinitions, 1)
	assert.Equal(t, "size", updated.SpecDefinitions[0].Slug)
	assert.False(t, updated.SpecDefinitions[0].IsVisible)
}

func TestProductServiceRejectsDuplicateProductTypeSlug(t *testing.T) {
	_, productService := newTestProductService(t)
	_, err := productService.CreateProductType(ProductTypeInput{Name: "首饰", Slug: "jewelry", IsEnabled: true})
	require.NoError(t, err)

	_, err = productService.CreateProductType(ProductTypeInput{Name: "另一个类型", Slug: "jewelry", IsEnabled: true})
	assert.ErrorIs(t, err, ErrProductTypeSlugExists)
}

func TestProductServiceDeletesProductType(t *testing.T) {
	_, productService := newTestProductService(t)
	created, err := productService.CreateProductType(ProductTypeInput{
		Name:      "首饰",
		Slug:      "jewelry",
		IsEnabled: true,
		SpecDefinitions: []ProductSpecDefinitionInput{
			{Name: "材质", Slug: "material", FieldType: "text", IsVisible: true},
		},
	})
	require.NoError(t, err)

	require.NoError(t, productService.DeleteProductType(created.ID))
	_, err = productService.GetProductType(created.ID)
	assert.ErrorIs(t, err, ErrProductTypeNotFound)
}
