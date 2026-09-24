package service

import (
	productdomain "commerce-platform/internal/domain/product"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildProductTemplateSyncDiffReportsMaterializationChanges(t *testing.T) {
	item := &productdomain.Product{
		ID: 7,
		VariantOptionValues: []productdomain.ProductVariantOptionValue{
			{
				SpecDefinitionID:       10,
				SourceTemplateRevision: 2,
				ValueKey:               "hg",
				Label:                  "旧 HG",
				IsEnabled:              true,
				CustomOptionPolicy:     &productdomain.ProductCustomOptionPolicy{PriceDeltaMinor: 0},
			},
			{
				SpecDefinitionID:       10,
				SourceTemplateRevision: 2,
				ValueKey:               "legacy",
				Label:                  "旧值",
				IsEnabled:              true,
				CustomOptionPolicy:     &productdomain.ProductCustomOptionPolicy{PriceDeltaMinor: 100},
			},
		},
	}
	template := &productdomain.ProductSpecificationTemplate{
		ID:       3,
		Revision: 4,
		SpecDefinitions: []productdomain.SpecDefinition{
			{
				ID: 10, Slug: "freehub", Name: "塔基", Role: ProductSpecRoleCustomOption,
				OptionItems: []productdomain.ProductSpecOptionItem{
					{ValueKey: "hg", DefaultLabel: "新 HG", IsEnabledByDefault: false},
					{ValueKey: "xdr", DefaultLabel: "XDR", IsEnabledByDefault: true, DefaultPriceDeltaMinor: int64Ptr(1500)},
				},
			},
		},
	}

	diff := buildProductTemplateSyncDiff(item, template)
	require.True(t, diff.HasChanges)
	require.True(t, diff.RevisionChanged)
	require.Equal(t, 2, diff.MaterializedRevision)
	require.Equal(t, 4, diff.TemplateRevision)
	require.Equal(t, 1, diff.Summary.Added)
	require.Equal(t, 1, diff.Summary.Disabled)
	require.Equal(t, 1, diff.Summary.LabelChanged)
	require.Equal(t, 1, diff.Summary.Removed)
	require.Len(t, diff.Items, 3)
}

func TestBuildProductTemplateSyncDiffNoChangesWhenMaterialized(t *testing.T) {
	price := int64(500)
	item := &productdomain.Product{
		ID: 8,
		VariantOptionValues: []productdomain.ProductVariantOptionValue{{
			SpecDefinitionID:       11,
			SourceTemplateRevision: 3,
			ValueKey:               "black",
			Label:                  "黑色",
			IsEnabled:              true,
			CustomOptionPolicy:     &productdomain.ProductCustomOptionPolicy{PriceDeltaMinor: price},
		}},
	}
	template := &productdomain.ProductSpecificationTemplate{
		ID:       4,
		Revision: 3,
		SpecDefinitions: []productdomain.SpecDefinition{{
			ID: 11, Slug: "finish", Name: "颜色", Role: ProductSpecRoleCustomOption,
			OptionItems: []productdomain.ProductSpecOptionItem{{
				ValueKey: "black", DefaultLabel: "黑色", IsEnabledByDefault: true, DefaultPriceDeltaMinor: &price,
			}},
		}},
	}

	diff := buildProductTemplateSyncDiff(item, template)
	require.False(t, diff.HasChanges)
	require.False(t, diff.RevisionChanged)
	require.Empty(t, diff.Items)
}

func TestSyncProductTemplateMaterializesConfirmedRevisionAndArchivesRemovedValue(t *testing.T) {
	db, productService := newTestProductService(t)
	template, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name:      "Sync Template",
		Slug:      "sync-template",
		IsEnabled: true,
		SpecDefinitions: []ProductSpecDefinitionInput{{
			Name:          "Freehub",
			Slug:          "freehub",
			FieldType:     "select",
			Role:          ProductSpecRoleCustomOption,
			SelectionMode: ProductSpecSelectionSingle,
			OptionItems: []ProductSpecOptionItemInput{
				{ValueKey: "hg", DefaultLabel: "HG"},
				{ValueKey: "xdr", DefaultLabel: "XDR"},
			},
		}},
	})
	require.NoError(t, err)

	created, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductSpecificationTemplateID: &template.ID,
		Name:                           "Sync Product",
		Slug:                           "sync-product",
		Locale:                         "en",
		Status:                         "active",
		Variants: []ProductVariantInput{{
			SKU: "SYNC-001", Currency: "USD", PriceMinor: 10000, Stock: 2,
			IsDefault: true, IsActive: boolPtr(true),
		}},
	})
	require.NoError(t, err)
	require.Len(t, created.VariantOptionValues, 2)

	definition := template.SpecDefinitions[0]
	updatedTemplate, err := productService.UpdateProductSpecificationTemplate(template.ID, ProductSpecificationTemplateInput{
		Name:      template.Name,
		Slug:      template.Slug,
		IsEnabled: true,
		SpecDefinitions: []ProductSpecDefinitionInput{{
			ID:            definition.ID,
			Name:          definition.Name,
			Slug:          definition.Slug,
			FieldType:     definition.FieldType,
			Role:          ProductSpecRoleCustomOption,
			SelectionMode: ProductSpecSelectionSingle,
			OptionItems: []ProductSpecOptionItemInput{{
				ID: definition.OptionItems[0].ID, ValueKey: "hg", DefaultLabel: "HG",
			}},
		}},
	})
	require.NoError(t, err)

	diff, err := productService.PreviewProductTemplateSync(created.ID)
	require.NoError(t, err)
	require.Equal(t, updatedTemplate.Revision, diff.TemplateRevision)
	require.Equal(t, 1, diff.Summary.Removed)
	require.True(t, diff.HasChanges)

	synced, syncedDiff, err := productService.SyncProductTemplate(created.ID, diff.TemplateRevision)
	require.NoError(t, err)
	require.False(t, syncedDiff.HasChanges)
	require.Len(t, synced.VariantOptionValues, 2)

	var archived productdomain.ProductVariantOptionValue
	require.NoError(t, db.Where("product_id = ? AND value_key = ?", created.ID, "xdr").First(&archived).Error)
	require.False(t, archived.IsEnabled)
}

func int64Ptr(value int64) *int64 { return &value }
