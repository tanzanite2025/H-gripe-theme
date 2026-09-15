package service

import (
	productdomain "commerce-platform/internal/domain/product"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveProductConfigurationCanonicalizesAndPricesServerOptions(t *testing.T) {
	max := 2
	item := &productdomain.Product{
		ProductSpecificationTemplate: &productdomain.ProductSpecificationTemplate{SpecDefinitions: []productdomain.SpecDefinition{{
			ID: 10, Slug: "accessories", Role: ProductSpecRoleCustomOption,
			SelectionMode: ProductSpecSelectionMultiple, MinSelections: 1, MaxSelections: &max,
		}}},
		VariantOptionValues: []productdomain.ProductVariantOptionValue{
			{ProductID: 1, SpecDefinitionID: 10, ValueKey: "spare_spokes", IsEnabled: true, CustomOptionPolicy: &productdomain.ProductCustomOptionPolicy{PriceDeltaMinor: 1500, InventoryPolicy: ProductCustomOptionInventoryNone}},
			{ProductID: 1, SpecDefinitionID: 10, ValueKey: "tape", IsEnabled: true, CustomOptionPolicy: &productdomain.ProductCustomOptionPolicy{PriceDeltaMinor: 0, InventoryPolicy: ProductCustomOptionInventoryNone}},
		},
	}
	result, err := ResolveProductConfiguration(item, &productdomain.ProductVariant{Currency: "USD"}, []SelectedOption{{GroupSlug: "accessories", ValueKeys: []string{"spare_spokes", "tape"}}})
	require.NoError(t, err)
	require.Equal(t, int64(1500), result.Delta.AmountMinor())
	require.Equal(t, "cc5102b9f28edcfc77c7364bdc51b2d1cee9d85f6425b83ba94e772f5e69a56e", result.Hash)
	require.JSONEq(t, `{"schema_version":1,"selections":[{"group_slug":"accessories","value_keys":["spare_spokes","tape"]}]}`, string(result.Data))
}

func TestResolveProductConfigurationRejectsInvalidSelections(t *testing.T) {
	item := &productdomain.Product{
		ProductSpecificationTemplate: &productdomain.ProductSpecificationTemplate{SpecDefinitions: []productdomain.SpecDefinition{{
			ID: 10, Slug: "freehub", Role: ProductSpecRoleCustomOption, SelectionMode: ProductSpecSelectionSingle, IsRequired: true,
		}}},
		VariantOptionValues: []productdomain.ProductVariantOptionValue{{
			SpecDefinitionID: 10, ValueKey: "hg", IsEnabled: true,
			CustomOptionPolicy: &productdomain.ProductCustomOptionPolicy{InventoryPolicy: ProductCustomOptionInventoryNone},
		}},
	}
	_, err := ResolveProductConfiguration(item, &productdomain.ProductVariant{Currency: "USD"}, []SelectedOption{{GroupSlug: "freehub", ValueKeys: []string{"missing"}}})
	require.ErrorIs(t, err, ErrProductConfigurationConflict)

	_, err = ResolveProductConfiguration(item, &productdomain.ProductVariant{Currency: "USD"}, nil)
	require.ErrorIs(t, err, ErrProductConfigurationRequired)
}

func TestResolveProductConfigurationAppliesVariantRules(t *testing.T) {
	max := 2
	override := int64(2400)
	item := &productdomain.Product{
		ID: 9,
		ProductSpecificationTemplate: &productdomain.ProductSpecificationTemplate{SpecDefinitions: []productdomain.SpecDefinition{{
			ID: 10, Slug: "accessories", Name: "Accessories", Role: ProductSpecRoleCustomOption,
			SelectionMode: ProductSpecSelectionMultiple, MinSelections: 1,
		}}},
		VariantOptionValues: []productdomain.ProductVariantOptionValue{
			{ID: 101, ProductID: 9, SpecDefinitionID: 10, ValueKey: "bag", Label: "Bag", IsEnabled: true, CustomOptionPolicy: &productdomain.ProductCustomOptionPolicy{PriceDeltaMinor: 1200, InventoryPolicy: ProductCustomOptionInventoryNone}},
			{ID: 102, ProductID: 9, SpecDefinitionID: 10, ValueKey: "tape", Label: "Tape", IsEnabled: true, CustomOptionPolicy: &productdomain.ProductCustomOptionPolicy{PriceDeltaMinor: 300, InventoryPolicy: ProductCustomOptionInventoryNone}},
		},
	}
	variant := &productdomain.ProductVariant{
		ID: 20, Currency: "USD",
		OptionGroupRules: []productdomain.ProductOptionGroupVariantRule{{SpecDefinitionID: 10, IsApplicable: true, MinSelectionsOverride: &max}},
		OptionValueRules: []productdomain.ProductOptionValueVariantRule{{ProductVariantOptionValueID: 101, IsEnabled: true, PriceDeltaMinorOverride: &override}},
	}
	result, err := ResolveProductConfiguration(item, variant, []SelectedOption{{GroupSlug: "accessories", ValueKeys: []string{"bag", "tape"}}})
	require.NoError(t, err)
	require.Equal(t, int64(2700), result.Delta.AmountMinor())

	variant.OptionValueRules[0].IsEnabled = false
	_, err = ResolveProductConfiguration(item, variant, []SelectedOption{{GroupSlug: "accessories", ValueKeys: []string{"bag", "tape"}}})
	require.ErrorIs(t, err, ErrProductConfigurationConflict)
}

func TestResolveProductConfigurationSkipsDisabledVariantOptionGroup(t *testing.T) {
	item := &productdomain.Product{
		ProductSpecificationTemplate: &productdomain.ProductSpecificationTemplate{SpecDefinitions: []productdomain.SpecDefinition{{
			ID: 10, Slug: "engraving", Role: ProductSpecRoleCustomOption, IsRequired: true,
		}}},
	}
	variant := &productdomain.ProductVariant{Currency: "USD", OptionGroupRules: []productdomain.ProductOptionGroupVariantRule{{SpecDefinitionID: 10, IsApplicable: false}}}
	result, err := ResolveProductConfiguration(item, variant, nil)
	require.NoError(t, err)
	require.Equal(t, int64(0), result.Delta.AmountMinor())
}

func TestAdminProductVariantOptionRulesPersistAcrossCreateAndUpdateReload(t *testing.T) {
	db, productService := newTestProductService(t)
	max := 1
	template, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name:      "Rule Matrix Template",
		Slug:      "rule_matrix_template",
		IsEnabled: true,
		SpecDefinitions: []ProductSpecDefinitionInput{{
			Name:          "Accessories",
			Slug:          "accessories",
			FieldType:     "select",
			Role:          ProductSpecRoleCustomOption,
			SelectionMode: ProductSpecSelectionMultiple,
			MinSelections: 0,
			MaxSelections: &max,
			OptionItems: []ProductSpecOptionItemInput{
				{ValueKey: "tape", DefaultLabel: "Tape"},
				{ValueKey: "valves", DefaultLabel: "Valves"},
			},
		}},
	})
	require.NoError(t, err)
	require.Len(t, template.SpecDefinitions, 1)
	definitionID := template.SpecDefinitions[0].ID

	created, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductSpecificationTemplateID: &template.ID,
		Name:                           "Rule Matrix Product",
		Slug:                           "rule-matrix-product",
		Locale:                         "en",
		Status:                         "active",
		Variants: []ProductVariantInput{{
			SKU:        "RULE-MATRIX-001",
			PriceMinor: 10000,
			Currency:   "USD",
			Stock:      4,
			IsDefault:  true,
			IsActive:   boolPtr(true),
			OptionGroupRules: []productdomain.ProductOptionGroupVariantRule{{
				SpecDefinitionID:      definitionID,
				IsApplicable:          true,
				MaxSelectionsOverride: &max,
			}},
		}},
	})
	require.NoError(t, err)
	require.Len(t, created.Variants, 1)
	require.Len(t, created.VariantOptionValues, 2)
	require.Len(t, created.Variants[0].OptionGroupRules, 1)

	valvesID := created.VariantOptionValues[1].ID
	override := int64(250)
	active := true
	updated, err := productService.UpdateAdminProduct(created.ID, ProductUpdateInput{
		Variants: []ProductVariantInput{{
			ID:         &created.Variants[0].ID,
			SKU:        created.Variants[0].SKU,
			PriceMinor: created.Variants[0].PriceMinor,
			Currency:   created.Variants[0].Currency,
			Stock:      created.Variants[0].Stock,
			IsDefault:  true,
			IsActive:   &active,
			OptionGroupRules: []productdomain.ProductOptionGroupVariantRule{{
				SpecDefinitionID:      definitionID,
				IsApplicable:          false,
				MaxSelectionsOverride: &max,
			}},
			OptionValueRules: []productdomain.ProductOptionValueVariantRule{{
				ProductVariantOptionValueID: valvesID,
				IsEnabled:                   false,
				PriceDeltaMinorOverride:     &override,
				UnavailableReason:           "not bundled with this variant",
			}},
		}},
		UpdateVariants: true,
	})
	require.NoError(t, err)
	require.Len(t, updated.Variants, 1)
	require.Len(t, updated.Variants[0].OptionGroupRules, 1)
	require.Len(t, updated.Variants[0].OptionValueRules, 1)
	require.False(t, updated.Variants[0].OptionGroupRules[0].IsApplicable)
	require.False(t, updated.Variants[0].OptionValueRules[0].IsEnabled)
	require.Equal(t, "not bundled with this variant", updated.Variants[0].OptionValueRules[0].UnavailableReason)

	reloaded, err := productService.GetByID(created.ID)
	require.NoError(t, err)
	require.Len(t, reloaded.Variants[0].OptionGroupRules, 1)
	require.Len(t, reloaded.Variants[0].OptionValueRules, 1)
	require.False(t, reloaded.Variants[0].OptionGroupRules[0].IsApplicable)
	require.Equal(t, valvesID, reloaded.Variants[0].OptionValueRules[0].ProductVariantOptionValueID)
	require.Equal(t, override, *reloaded.Variants[0].OptionValueRules[0].PriceDeltaMinorOverride)
	require.False(t, reloaded.Variants[0].OptionValueRules[0].IsEnabled)

	var persisted productdomain.ProductOptionValueVariantRule
	require.NoError(t, db.Where("variant_id = ? AND product_variant_option_value_id = ?", created.Variants[0].ID, valvesID).First(&persisted).Error)
	require.Equal(t, "not bundled with this variant", persisted.UnavailableReason)
}
