package service

import (
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
				OptionItems:  []ProductSpecOptionItemInput{{ValueKey: "银"}, {ValueKey: "金"}},
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "首饰", created.Name)
	assert.True(t, created.IsEnabled)
	require.Len(t, created.SpecDefinitions, 1)
	assert.Equal(t, []string{"银", "金"}, []string{created.SpecDefinitions[0].OptionItems[0].ValueKey, created.SpecDefinitions[0].OptionItems[1].ValueKey})
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
				Group:        "Appearance",
				Name:         "Finish",
				Slug:         "finish",
				FieldType:    "select",
				Presentation: "color",
				IsVisible:    true,
				Role:         "variant",
				SortOrder:    10,
			},
		},
	})

	require.NoError(t, err)
	require.Len(t, created.SpecDefinitions, 1)
	assert.Equal(t, "color", created.SpecDefinitions[0].Presentation)
	assert.Empty(t, created.SpecDefinitions[0].OptionItems)
}

func TestProductServicePersistsCustomOptionSelectionContract(t *testing.T) {
	_, productService := newTestProductService(t)
	max := 2
	created, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name:      "Wheelset Options",
		Slug:      "wheelset_options_contract",
		IsEnabled: true,
		SpecDefinitions: []ProductSpecDefinitionInput{{
			Name:          "Freehub",
			Slug:          "freehub",
			FieldType:     "select",
			Role:          "custom_option",
			SelectionMode: "single",
			IsRequired:    true,
			OptionItems:   []ProductSpecOptionItemInput{{ValueKey: "hg"}, {ValueKey: "xdr"}},
		}, {
			Name:          "Accessories",
			Slug:          "accessories",
			FieldType:     "select",
			Role:          "custom_option",
			SelectionMode: "multiple",
			MinSelections: 1,
			MaxSelections: &max,
		}},
	})

	require.NoError(t, err)
	require.Len(t, created.SpecDefinitions, 2)
	assert.Equal(t, "custom_option", created.SpecDefinitions[0].Role)
	assert.Equal(t, "single", created.SpecDefinitions[0].SelectionMode)
	assert.True(t, created.SpecDefinitions[0].IsRequired)
	assert.Equal(t, "custom_option", created.SpecDefinitions[1].Role)
	assert.Equal(t, "multiple", created.SpecDefinitions[1].SelectionMode)
	assert.Equal(t, 1, created.SpecDefinitions[1].MinSelections)
	require.NotNil(t, created.SpecDefinitions[1].MaxSelections)
	assert.Equal(t, 2, *created.SpecDefinitions[1].MaxSelections)
}

func TestProductServicePersistsTemplateOptionItemsAndRevision(t *testing.T) {
	_, productService := newTestProductService(t)
	price := int64(1500)
	created, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name:      "Wheelset option items",
		Slug:      "wheelset_option_items_contract",
		IsEnabled: true,
		SpecDefinitions: []ProductSpecDefinitionInput{{
			Name: "Freehub", Slug: "freehub", FieldType: "select", Role: "custom_option",
			SelectionMode: "single", IsRequired: true,
			OptionItems: []ProductSpecOptionItemInput{{
				ValueKey: "hg", DefaultLabel: "Shimano HG", IsDefault: true,
				DefaultPriceDeltaMinor: &price, DefaultPriceCurrency: "USD",
			}},
		}},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, created.Revision)
	require.Len(t, created.SpecDefinitions, 1)
	require.Len(t, created.SpecDefinitions[0].OptionItems, 1)
	assert.Equal(t, "hg", created.SpecDefinitions[0].OptionItems[0].ValueKey)
	assert.Equal(t, "Shimano HG", created.SpecDefinitions[0].OptionItems[0].DefaultLabel)
	assert.Equal(t, int64(1500), *created.SpecDefinitions[0].OptionItems[0].DefaultPriceDeltaMinor)

	updated, err := productService.UpdateProductSpecificationTemplate(created.ID, ProductSpecificationTemplateInput{
		Name: created.Name, Slug: created.Slug, IsEnabled: true,
		SpecDefinitions: []ProductSpecDefinitionInput{{
			ID:   created.SpecDefinitions[0].ID,
			Name: "Freehub", Slug: "freehub", FieldType: "select", Role: "custom_option",
			SelectionMode: "single", IsRequired: true,
			OptionItems: []ProductSpecOptionItemInput{{ValueKey: "xdr", DefaultLabel: "SRAM XDR", IsDefault: true}},
		}},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, updated.Revision)
	require.Len(t, updated.SpecDefinitions[0].OptionItems, 1)
	assert.Equal(t, "xdr", updated.SpecDefinitions[0].OptionItems[0].ValueKey)
}

func TestProductServiceMaterializesTemplateOptionItemsOnProductCreate(t *testing.T) {
	_, productService := newTestProductService(t)
	priceDelta := int64(1500)
	template, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name: "Materialized Wheelset Options", Slug: "materialized_wheelset_options", IsEnabled: true,
		SpecDefinitions: []ProductSpecDefinitionInput{{
			Name: "Freehub", Slug: "freehub", FieldType: "select", Role: "custom_option",
			SelectionMode: "single", IsRequired: true,
			OptionItems: []ProductSpecOptionItemInput{{
				ValueKey: "xdr", DefaultLabel: "SRAM XDR", IsDefault: true,
				DefaultPriceDeltaMinor: &priceDelta, DefaultPriceCurrency: "USD",
			}},
		}},
	})
	require.NoError(t, err)
	require.Len(t, template.SpecDefinitions, 1)

	created, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductSpecificationTemplateID: &template.ID,
		Name:                           "Materialized Wheelset", Slug: "materialized-wheelset", Status: "active", Locale: "en",
		Variants: []ProductVariantInput{{SKU: "MAT-WHEELSET-001", PriceMinor: 499, Stock: 3, IsDefault: true, IsActive: boolPtr(true)}},
	})
	require.NoError(t, err)
	require.Len(t, created.VariantOptionValues, 1)
	materialized := created.VariantOptionValues[0]
	assert.Equal(t, "xdr", materialized.ValueKey)
	assert.Equal(t, template.Revision, materialized.SourceTemplateRevision)
	require.NotNil(t, materialized.TemplateOptionItemID)
	assert.Equal(t, template.SpecDefinitions[0].OptionItems[0].ID, *materialized.TemplateOptionItemID)
	require.NotNil(t, materialized.CustomOptionPolicy)
	assert.Equal(t, priceDelta, materialized.CustomOptionPolicy.PriceDeltaMinor)
	assert.True(t, materialized.CustomOptionPolicy.IsDefault)
	assert.Equal(t, ProductCustomOptionInventoryNone, materialized.CustomOptionPolicy.InventoryPolicy)
}

func TestProductServiceRejectsInvalidCustomOptionSelectionContract(t *testing.T) {
	_, productService := newTestProductService(t)
	tests := []ProductSpecDefinitionInput{
		{Name: "Bad type", Slug: "bad_type", FieldType: "text", Role: "custom_option"},
		{Name: "Bad role", Slug: "bad_role", FieldType: "select", Role: "unknown"},
		{Name: "Bad bounds", Slug: "bad_bounds", FieldType: "select", Role: "custom_option", SelectionMode: "multiple", MinSelections: 2, MaxSelections: intPtrForProductSpecTest(1)},
		{Name: "Single max", Slug: "single_max", FieldType: "select", Role: "custom_option", SelectionMode: "single", MaxSelections: intPtrForProductSpecTest(2)},
	}
	for _, definition := range tests {
		_, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
			Name:            "Invalid " + definition.Name,
			Slug:            "invalid_" + definition.Slug,
			IsEnabled:       true,
			SpecDefinitions: []ProductSpecDefinitionInput{definition},
		})
		assert.ErrorIs(t, err, ErrProductSpecInvalid, definition.Name)
	}
}

func intPtrForProductSpecTest(value int) *int {
	return &value
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
	assert.Empty(t, created.SpecDefinitions[0].OptionItems)
}

func TestProductServiceListsPublicProductSpecificationTemplatesWithoutSpecifications(t *testing.T) {
	_, productService := newTestProductService(t)

	created, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name:      "Wheelset",
		Slug:      "wheelset_public_index",
		IsEnabled: true,
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
	assert.Equal(t, "Wheelset", publicTypes[0].Name)
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
