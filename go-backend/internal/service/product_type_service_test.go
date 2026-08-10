package service

import (
	"context"
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

func TestProductServiceCreatesVisualVariantOptionWithoutFixedTemplateOptions(t *testing.T) {
	_, productService := newTestProductService(t)

	created, err := productService.CreateProductType(ProductTypeInput{
		Name:      "Finish Product",
		Slug:      "finish_product_template",
		IsEnabled: true,
		SpecDefinitions: []ProductSpecDefinitionInput{
			{
				Group:           "Appearance",
				Name:            "Finish",
				Slug:            "finish",
				FieldType:       "select",
				Presentation:    "color",
				IsVisible:       true,
				IsVariantOption: true,
				SortOrder:       10,
				Options:         "",
			},
		},
	})

	require.NoError(t, err)
	require.Len(t, created.SpecDefinitions, 1)
	assert.Equal(t, "color", created.SpecDefinitions[0].Presentation)
	assert.JSONEq(t, `[]`, created.SpecDefinitions[0].Options)
}

func TestProductServicePersistsAndReplacesProductTypeTranslations(t *testing.T) {
	_, productService := newTestProductService(t)

	created, err := productService.CreateProductType(ProductTypeInput{
		Name:      "Wheelset",
		Slug:      "wheelset_translation_test",
		IsEnabled: true,
		Translations: []ProductTypeTranslationInput{
			{Locale: "en", Name: "Wheelset"},
			{Locale: "zh-CN", Name: "轮组"},
		},
	})
	require.NoError(t, err)
	require.Len(t, created.Translations, 2)
	assert.Equal(t, "轮组", created.NameForLocale("zh_cn"))

	updated, err := productService.UpdateProductType(created.ID, ProductTypeInput{
		Name:               "Wheelset",
		Slug:               "wheelset_translation_test",
		IsEnabled:          true,
		UpdateTranslations: true,
		Translations: []ProductTypeTranslationInput{
			{Locale: "en", Name: "Wheelset"},
			{Locale: "fr", Name: "Jeu de roues"},
		},
	})
	require.NoError(t, err)
	require.Len(t, updated.Translations, 2)
	assert.Equal(t, "Wheelset", updated.NameForLocale("zh_cn"))
	assert.Equal(t, "Jeu de roues", updated.NameForLocale("fr"))
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

func TestProductServiceUpdatesAndClearsProductTypeImage(t *testing.T) {
	_, productService := newTestProductService(t)
	created, err := productService.CreateProductType(ProductTypeInput{
		Name:      "首饰",
		Slug:      "jewelry_image",
		IsEnabled: true,
	})
	require.NoError(t, err)

	assetID := uint(42)
	updated, err := productService.UpdateProductTypeImage(created.ID, &assetID, "https://cdn.example.com/categories/jewelry.webp")
	require.NoError(t, err)
	require.NotNil(t, updated.ImageMediaAssetID)
	assert.Equal(t, assetID, *updated.ImageMediaAssetID)
	assert.Equal(t, "https://cdn.example.com/categories/jewelry.webp", updated.ImageURL)

	cleared, err := productService.UpdateProductTypeImage(created.ID, nil, "")
	require.NoError(t, err)
	assert.Nil(t, cleared.ImageMediaAssetID)
	assert.Empty(t, cleared.ImageURL)
}

func TestProductServiceCleansUpDetachedProductTypeImageAssets(t *testing.T) {
	_, productService := newTestProductService(t)
	deleter := &recordingMediaAssetDeleter{}
	productService.ConfigureMediaService(deleter)

	created, err := productService.CreateProductType(ProductTypeInput{
		Name:      "首饰",
		Slug:      "jewelry_image_cleanup",
		IsEnabled: true,
	})
	require.NoError(t, err)

	firstAssetID := uint(101)
	_, err = productService.UpdateProductTypeImage(created.ID, &firstAssetID, "https://cdn.example.com/categories/first.webp")
	require.NoError(t, err)

	secondAssetID := uint(102)
	_, err = productService.UpdateProductTypeImage(created.ID, &secondAssetID, "https://cdn.example.com/categories/second.webp")
	require.NoError(t, err)

	_, err = productService.UpdateProductTypeImage(created.ID, nil, "")
	require.NoError(t, err)

	assert.Equal(t, []uint{firstAssetID, secondAssetID}, deleter.ids)
	assert.Equal(t,
		[]string{
			MediaAssetDeleteConfirmation(firstAssetID),
			MediaAssetDeleteConfirmation(secondAssetID),
		},
		deleter.confirmations,
	)
}

func TestProductServiceCleansUpProductTypeImageOnDelete(t *testing.T) {
	_, productService := newTestProductService(t)
	deleter := &recordingMediaAssetDeleter{}
	productService.ConfigureMediaService(deleter)

	created, err := productService.CreateProductType(ProductTypeInput{
		Name:      "首饰",
		Slug:      "jewelry_image_delete_cleanup",
		IsEnabled: true,
	})
	require.NoError(t, err)

	assetID := uint(103)
	_, err = productService.UpdateProductTypeImage(created.ID, &assetID, "https://cdn.example.com/categories/delete.webp")
	require.NoError(t, err)

	require.NoError(t, productService.DeleteProductType(created.ID))
	assert.Equal(t, []uint{assetID}, deleter.ids)
}

func TestProductServiceKeepsProductTypeWhenDetachedImageAssetIsStillReferenced(t *testing.T) {
	_, productService := newTestProductService(t)
	deleter := &recordingMediaAssetDeleter{err: ErrMediaAssetInUse}
	productService.ConfigureMediaService(deleter)

	created, err := productService.CreateProductType(ProductTypeInput{
		Name:      "首饰",
		Slug:      "jewelry_image_shared",
		IsEnabled: true,
	})
	require.NoError(t, err)

	firstAssetID := uint(104)
	_, err = productService.UpdateProductTypeImage(created.ID, &firstAssetID, "https://cdn.example.com/categories/shared-first.webp")
	require.NoError(t, err)

	secondAssetID := uint(105)
	_, err = productService.UpdateProductTypeImage(created.ID, &secondAssetID, "https://cdn.example.com/categories/shared-second.webp")
	require.NoError(t, err)

	assert.Equal(t, []uint{firstAssetID}, deleter.ids)
	updated, err := productService.GetProductType(created.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.ImageMediaAssetID)
	assert.Equal(t, secondAssetID, *updated.ImageMediaAssetID)
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

type recordingMediaAssetDeleter struct {
	ids           []uint
	confirmations []string
	err           error
}

func (d *recordingMediaAssetDeleter) DeleteAsset(_ context.Context, id uint, confirmation string) error {
	d.ids = append(d.ids, id)
	d.confirmations = append(d.confirmations, confirmation)
	return d.err
}
