package service

import (
	"context"
	"testing"

	"commerce-platform/internal/domain/product"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductServiceCreatesManagedProductSpecificationTemplate(t *testing.T) {
	_, productService := newTestProductService(t)

	created, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
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

	created, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
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

func TestProductServiceAllowsDynamicSelectSpecificationWithoutSharedOptions(t *testing.T) {
	_, productService := newTestProductService(t)

	created, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name:      "Rim",
		Slug:      "rim_dynamic_values",
		IsEnabled: true,
		SpecDefinitions: []ProductSpecDefinitionInput{
			{
				Group:        "规格",
				Name:         "Rim Depth",
				Slug:         "rim_depth",
				FieldType:    "select",
				IsFilterable: true,
				IsVisible:    true,
				SortOrder:    10,
			},
		},
	})

	require.NoError(t, err)
	require.Len(t, created.SpecDefinitions, 1)
	assert.Equal(t, "select", created.SpecDefinitions[0].FieldType)
	assert.JSONEq(t, `[]`, created.SpecDefinitions[0].Options)
}

func TestProductServicePersistsAndReplacesProductSpecificationTemplateTranslations(t *testing.T) {
	_, productService := newTestProductService(t)

	created, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name:      "Wheelset",
		Slug:      "wheelset_translation_test",
		IsEnabled: true,
		Translations: []ProductSpecificationTemplateTranslationInput{
			{Locale: "en", Name: "Wheelset"},
			{Locale: "zh-CN", Name: "轮组"},
		},
	})
	require.NoError(t, err)
	require.Len(t, created.Translations, 2)
	assert.Equal(t, "轮组", created.NameForLocale("zh_cn"))

	updated, err := productService.UpdateProductSpecificationTemplate(created.ID, ProductSpecificationTemplateInput{
		Name:               "Wheelset",
		Slug:               "wheelset_translation_test",
		IsEnabled:          true,
		UpdateTranslations: true,
		Translations: []ProductSpecificationTemplateTranslationInput{
			{Locale: "en", Name: "Wheelset"},
			{Locale: "fr", Name: "Jeu de roues"},
		},
	})
	require.NoError(t, err)
	require.Len(t, updated.Translations, 2)
	assert.Equal(t, "Wheelset", updated.NameForLocale("zh_cn"))
	assert.Equal(t, "Jeu de roues", updated.NameForLocale("fr"))
}

func TestProductServiceListsPublicProductSpecificationTemplatesWithoutSpecifications(t *testing.T) {
	_, productService := newTestProductService(t)

	created, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name:      "Wheelset",
		Slug:      "wheelset_public_index",
		IsEnabled: true,
		Translations: []ProductSpecificationTemplateTranslationInput{
			{Locale: "zh-CN", Name: "轮组"},
		},
		SpecDefinitions: []ProductSpecDefinitionInput{
			{Name: "Material", Slug: "material", FieldType: "text", IsVisible: true},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.SpecDefinitions)

	publicTypes, err := productService.ListPublicProductSpecificationTemplates(false)
	require.NoError(t, err)
	require.Len(t, publicTypes, 1)
	assert.Empty(t, publicTypes[0].SpecDefinitions)
	assert.Equal(t, "轮组", publicTypes[0].NameForLocale("zh_cn"))
}

func TestProductServiceUpdatesProductSpecificationTemplateAndReplacesSpecs(t *testing.T) {
	_, productService := newTestProductService(t)
	created, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name:      "首饰",
		Slug:      "jewelry",
		IsEnabled: true,
		SpecDefinitions: []ProductSpecDefinitionInput{
			{Name: "材质", Slug: "material", FieldType: "text", IsVisible: true},
		},
	})
	require.NoError(t, err)

	updated, err := productService.UpdateProductSpecificationTemplate(created.ID, ProductSpecificationTemplateInput{
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

func TestProductServiceProductSpecificationTemplateMutationInvalidatesDependentProductCache(t *testing.T) {
	_, productService := newTestProductService(t)
	cache := &recordingProductDetailCache{}
	repo := &fakeProductCacheIdentityRepository{
		productsByProductSpecificationTemplateID: map[uint][]product.Product{
			42: {{ID: 7, Slug: "typed-product", Locale: "en"}},
		},
	}
	productService.productCacheInvalidator = NewProductDetailCacheInvalidator(repo, cache)

	productService.InvalidateProductCacheByProductSpecificationTemplateID(42)

	assert.Equal(t, []uint{42}, repo.requestedProductSpecificationTemplateIDs)
	assert.Contains(t, cache.deletedKeys, "product:7")
	assert.Contains(t, cache.deletedKeys, "product:slug:typed-product:en")
}

func TestProductServiceUpdatesAndClearsProductSpecificationTemplateImage(t *testing.T) {
	_, productService := newTestProductService(t)
	created, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name:      "首饰",
		Slug:      "jewelry_image",
		IsEnabled: true,
	})
	require.NoError(t, err)

	assetID := uint(42)
	updated, err := productService.UpdateProductSpecificationTemplateImage(created.ID, &assetID, "https://cdn.example.com/categories/jewelry.webp")
	require.NoError(t, err)
	require.NotNil(t, updated.ImageMediaAssetID)
	assert.Equal(t, assetID, *updated.ImageMediaAssetID)
	assert.Equal(t, "https://cdn.example.com/categories/jewelry.webp", updated.ImageURL)

	cleared, err := productService.UpdateProductSpecificationTemplateImage(created.ID, nil, "")
	require.NoError(t, err)
	assert.Nil(t, cleared.ImageMediaAssetID)
	assert.Empty(t, cleared.ImageURL)
}

func TestProductServiceCleansUpDetachedProductSpecificationTemplateImageAssets(t *testing.T) {
	_, productService := newTestProductService(t)
	deleter := &recordingMediaAssetDeleter{}
	productService.ConfigureMediaService(deleter)

	created, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name:      "首饰",
		Slug:      "jewelry_image_cleanup",
		IsEnabled: true,
	})
	require.NoError(t, err)

	firstAssetID := uint(101)
	_, err = productService.UpdateProductSpecificationTemplateImage(created.ID, &firstAssetID, "https://cdn.example.com/categories/first.webp")
	require.NoError(t, err)

	secondAssetID := uint(102)
	_, err = productService.UpdateProductSpecificationTemplateImage(created.ID, &secondAssetID, "https://cdn.example.com/categories/second.webp")
	require.NoError(t, err)

	_, err = productService.UpdateProductSpecificationTemplateImage(created.ID, nil, "")
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

func TestProductServiceCleansUpProductSpecificationTemplateImageOnDelete(t *testing.T) {
	_, productService := newTestProductService(t)
	deleter := &recordingMediaAssetDeleter{}
	productService.ConfigureMediaService(deleter)

	created, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name:      "首饰",
		Slug:      "jewelry_image_delete_cleanup",
		IsEnabled: true,
	})
	require.NoError(t, err)

	assetID := uint(103)
	_, err = productService.UpdateProductSpecificationTemplateImage(created.ID, &assetID, "https://cdn.example.com/categories/delete.webp")
	require.NoError(t, err)

	require.NoError(t, productService.DeleteProductSpecificationTemplate(created.ID))
	assert.Equal(t, []uint{assetID}, deleter.ids)
}

func TestProductServiceKeepsProductSpecificationTemplateWhenDetachedImageAssetIsStillReferenced(t *testing.T) {
	_, productService := newTestProductService(t)
	deleter := &recordingMediaAssetDeleter{err: ErrMediaAssetInUse}
	productService.ConfigureMediaService(deleter)

	created, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name:      "首饰",
		Slug:      "jewelry_image_shared",
		IsEnabled: true,
	})
	require.NoError(t, err)

	firstAssetID := uint(104)
	_, err = productService.UpdateProductSpecificationTemplateImage(created.ID, &firstAssetID, "https://cdn.example.com/categories/shared-first.webp")
	require.NoError(t, err)

	secondAssetID := uint(105)
	_, err = productService.UpdateProductSpecificationTemplateImage(created.ID, &secondAssetID, "https://cdn.example.com/categories/shared-second.webp")
	require.NoError(t, err)

	assert.Equal(t, []uint{firstAssetID}, deleter.ids)
	updated, err := productService.GetProductSpecificationTemplate(created.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.ImageMediaAssetID)
	assert.Equal(t, secondAssetID, *updated.ImageMediaAssetID)
}

func TestProductServiceRejectsDuplicateProductSpecificationTemplateSlug(t *testing.T) {
	_, productService := newTestProductService(t)
	_, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{Name: "首饰", Slug: "jewelry", IsEnabled: true})
	require.NoError(t, err)

	_, err = productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{Name: "另一个类型", Slug: "jewelry", IsEnabled: true})
	assert.ErrorIs(t, err, ErrProductSpecificationTemplateSlugExists)
}

func TestProductServiceDeletesProductSpecificationTemplate(t *testing.T) {
	_, productService := newTestProductService(t)
	created, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name:      "首饰",
		Slug:      "jewelry",
		IsEnabled: true,
		SpecDefinitions: []ProductSpecDefinitionInput{
			{Name: "材质", Slug: "material", FieldType: "text", IsVisible: true},
		},
	})
	require.NoError(t, err)

	require.NoError(t, productService.DeleteProductSpecificationTemplate(created.ID))
	_, err = productService.GetProductSpecificationTemplate(created.ID)
	assert.ErrorIs(t, err, ErrProductSpecificationTemplateNotFound)
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
