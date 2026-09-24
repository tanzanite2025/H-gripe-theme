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

func TestResolveProductConfigurationEnforcesOptionValueRelations(t *testing.T) {
	item := &productdomain.Product{
		ID: 1,
		ProductSpecificationTemplate: &productdomain.ProductSpecificationTemplate{SpecDefinitions: []productdomain.SpecDefinition{
			{ID: 10, Slug: "base", Role: ProductSpecRoleCustomOption, SelectionMode: ProductSpecSelectionSingle},
			{ID: 11, Slug: "addon", Role: ProductSpecRoleCustomOption, SelectionMode: ProductSpecSelectionSingle},
		}},
		VariantOptionValues: []productdomain.ProductVariantOptionValue{
			{ID: 101, SpecDefinitionID: 10, ValueKey: "premium", IsEnabled: true, CustomOptionPolicy: &productdomain.ProductCustomOptionPolicy{}},
			{ID: 102, SpecDefinitionID: 11, ValueKey: "support", IsEnabled: true, CustomOptionPolicy: &productdomain.ProductCustomOptionPolicy{}},
		},
		OptionValueRelations: []productdomain.ProductOptionValueRelation{{SourceOptionValueID: 101, TargetOptionValueID: 102, RelationType: productdomain.OptionValueRelationRequires}},
	}
	_, err := ResolveProductConfiguration(item, &productdomain.ProductVariant{Currency: "USD"}, []SelectedOption{{GroupSlug: "base", ValueKeys: []string{"premium"}}})
	require.ErrorIs(t, err, ErrProductConfigurationDependencyConflict)
	require.ErrorIs(t, err, ErrProductConfigurationConflict)

	result, err := ResolveProductConfiguration(item, &productdomain.ProductVariant{Currency: "USD"}, []SelectedOption{{GroupSlug: "base", ValueKeys: []string{"premium"}}, {GroupSlug: "addon", ValueKeys: []string{"support"}}})
	require.NoError(t, err)
	require.Len(t, result.Snapshot.OptionRelations, 1)

	item.OptionValueRelations[0].RelationType = productdomain.OptionValueRelationConflicts
	_, err = ResolveProductConfiguration(item, &productdomain.ProductVariant{Currency: "USD"}, []SelectedOption{{GroupSlug: "base", ValueKeys: []string{"premium"}}, {GroupSlug: "addon", ValueKeys: []string{"support"}}})
	require.ErrorIs(t, err, ErrProductConfigurationDependencyConflict)
	require.ErrorIs(t, err, ErrProductConfigurationConflict)
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

func TestResolveProductConfigurationBuildsComponentInventoryAllocations(t *testing.T) {
	componentVariantID := uint(900)
	item := &productdomain.Product{
		ID: 1,
		ProductSpecificationTemplate: &productdomain.ProductSpecificationTemplate{SpecDefinitions: []productdomain.SpecDefinition{{
			ID: 10, Slug: "accessories", Role: ProductSpecRoleCustomOption,
			SelectionMode: ProductSpecSelectionMultiple, MinSelections: 1,
		}}},
		VariantOptionValues: []productdomain.ProductVariantOptionValue{
			{ProductID: 1, SpecDefinitionID: 10, ValueKey: "valves", IsEnabled: true, CustomOptionPolicy: &productdomain.ProductCustomOptionPolicy{InventoryPolicy: ProductCustomOptionInventoryComponent, ComponentVariantID: &componentVariantID, ComponentQuantity: 2}},
			{ProductID: 1, SpecDefinitionID: 10, ValueKey: "spokes", IsEnabled: true, CustomOptionPolicy: &productdomain.ProductCustomOptionPolicy{InventoryPolicy: ProductCustomOptionInventoryComponent, ComponentVariantID: &componentVariantID, ComponentQuantity: 1}},
		},
	}
	result, err := ResolveProductConfiguration(item, &productdomain.ProductVariant{ID: 42, Currency: "USD"}, []SelectedOption{{GroupSlug: "accessories", ValueKeys: []string{"spokes", "valves"}}})
	require.NoError(t, err)
	require.Equal(t, []ProductConfigurationInventoryAllocation{{VariantID: componentVariantID, Quantity: 3, GroupSlug: "accessories", ValueKey: "spokes"}}, result.InventoryAllocations)
	require.Equal(t, result.InventoryAllocations, result.Snapshot.InventoryAllocations)
	require.Equal(t, componentVariantID, *result.Snapshot.Selections[0].Values[1].ComponentVariantID)
	require.Equal(t, 2, result.Snapshot.Selections[0].Values[1].ComponentQuantity)
}

func TestResolveProductConfigurationRejectsInvalidComponentPolicy(t *testing.T) {
	item := &productdomain.Product{
		ProductSpecificationTemplate: &productdomain.ProductSpecificationTemplate{SpecDefinitions: []productdomain.SpecDefinition{{
			ID: 10, Slug: "accessories", Role: ProductSpecRoleCustomOption, SelectionMode: ProductSpecSelectionSingle,
		}}},
		VariantOptionValues: []productdomain.ProductVariantOptionValue{{
			SpecDefinitionID: 10, ValueKey: "kit", IsEnabled: true,
			CustomOptionPolicy: &productdomain.ProductCustomOptionPolicy{InventoryPolicy: ProductCustomOptionInventoryComponent},
		}},
	}
	_, err := ResolveProductConfiguration(item, &productdomain.ProductVariant{ID: 42, Currency: "USD"}, []SelectedOption{{GroupSlug: "accessories", ValueKeys: []string{"kit"}}})
	require.ErrorIs(t, err, ErrProductConfigurationConflict)
}

func TestResolveProductConfigurationCapturesWeightDeltas(t *testing.T) {
	item := &productdomain.Product{
		ProductSpecificationTemplate: &productdomain.ProductSpecificationTemplate{SpecDefinitions: []productdomain.SpecDefinition{{
			ID: 10, Slug: "service", Role: ProductSpecRoleCustomOption, SelectionMode: ProductSpecSelectionSingle,
		}}},
		VariantOptionValues: []productdomain.ProductVariantOptionValue{{
			SpecDefinitionID: 10, ValueKey: "reinforced", IsEnabled: true,
			CustomOptionPolicy: &productdomain.ProductCustomOptionPolicy{InventoryPolicy: ProductCustomOptionInventoryNone, WeightDeltaGrams: 120, PackagingWeightDeltaGrams: 80},
		}},
	}
	result, err := ResolveProductConfiguration(item, &productdomain.ProductVariant{ID: 42, Currency: "USD"}, []SelectedOption{{GroupSlug: "service", ValueKeys: []string{"reinforced"}}})
	require.NoError(t, err)
	require.Equal(t, 120, result.WeightDeltaGrams)
	require.Equal(t, 80, result.PackagingWeightDeltaGrams)
	require.Equal(t, 120, result.Snapshot.WeightDeltaGrams)
	require.Equal(t, 80, result.Snapshot.PackagingWeightDeltaGrams)
}

func TestResolveProductConfigurationAggregatesProductionAndPolicies(t *testing.T) {
	item := &productdomain.Product{
		ProductSpecificationTemplate: &productdomain.ProductSpecificationTemplate{SpecDefinitions: []productdomain.SpecDefinition{{
			ID: 10, Slug: "custom", Role: ProductSpecRoleCustomOption, SelectionMode: ProductSpecSelectionMultiple,
		}}},
		VariantOptionValues: []productdomain.ProductVariantOptionValue{
			{SpecDefinitionID: 10, ValueKey: "made", IsEnabled: true, CustomOptionPolicy: &productdomain.ProductCustomOptionPolicy{ProductionLeadTimeDays: 3, InventoryPolicy: ProductCustomOptionInventoryNone}},
			{SpecDefinitionID: 10, ValueKey: "restricted", IsEnabled: true, CustomOptionPolicy: &productdomain.ProductCustomOptionPolicy{RequiresProduction: true, CancellationPolicy: ProductCustomOptionCancellationNever, ReturnPolicy: ProductCustomOptionReturnNotAllowed, InventoryPolicy: ProductCustomOptionInventoryNone}},
		},
	}
	result, err := ResolveProductConfiguration(item, &productdomain.ProductVariant{Currency: "USD"}, []SelectedOption{{GroupSlug: "custom", ValueKeys: []string{"made", "restricted"}}})
	require.NoError(t, err)
	require.Equal(t, 3, result.ProductionLeadTimeDays)
	require.True(t, result.RequiresProduction)
	require.Equal(t, ProductCustomOptionCancellationNever, result.CancellationPolicy)
	require.Equal(t, ProductCustomOptionReturnNotAllowed, result.ReturnPolicy)
	require.Equal(t, ProductCustomOptionCancellationBeforeProduction, result.Snapshot.Selections[0].Values[0].CancellationPolicy)
}

func TestResolveProductConfigurationRejectsInvalidProductionPolicy(t *testing.T) {
	item := &productdomain.Product{
		ProductSpecificationTemplate: &productdomain.ProductSpecificationTemplate{SpecDefinitions: []productdomain.SpecDefinition{{ID: 10, Slug: "custom", Role: ProductSpecRoleCustomOption, SelectionMode: ProductSpecSelectionSingle}}},
		VariantOptionValues:          []productdomain.ProductVariantOptionValue{{SpecDefinitionID: 10, ValueKey: "bad", IsEnabled: true, CustomOptionPolicy: &productdomain.ProductCustomOptionPolicy{CancellationPolicy: "invalid", InventoryPolicy: ProductCustomOptionInventoryNone}}},
	}
	_, err := ResolveProductConfiguration(item, &productdomain.ProductVariant{Currency: "USD"}, []SelectedOption{{GroupSlug: "custom", ValueKeys: []string{"bad"}}})
	require.ErrorIs(t, err, ErrProductConfigurationConflict)
}

func TestAdminProductPersistsComponentInventoryPolicy(t *testing.T) {
	db, productService := newTestProductService(t)
	component, err := productService.CreateAdminProduct(ProductCreateInput{
		Name: "Valve Component", Slug: "valve-component", Locale: "en", Status: "active", Currency: "USD",
		Variants: []ProductVariantInput{{SKU: "VALVE-COMP-01", PriceMinor: 500, Currency: "USD", Stock: 10, IsActive: boolPtr(true), IsDefault: true}},
	})
	require.NoError(t, err)
	template, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name: "Component Option Template", Slug: "component-option-template", IsEnabled: true,
		SpecDefinitions: []ProductSpecDefinitionInput{{Name: "Valve", Slug: "valve", FieldType: "select", Role: ProductSpecRoleCustomOption, SelectionMode: ProductSpecSelectionSingle, OptionItems: []ProductSpecOptionItemInput{{ValueKey: "standard", DefaultLabel: "Standard"}}}},
	})
	require.NoError(t, err)
	definitionID := template.SpecDefinitions[0].ID
	created, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductSpecificationTemplateID: &template.ID,
		Name:                           "Configured Wheel", Slug: "configured-wheel", Locale: "en", Status: "active", Currency: "USD",
		Variants:            []ProductVariantInput{{SKU: "CONFIG-WHEEL-01", PriceMinor: 10000, Currency: "USD", Stock: 3, IsActive: boolPtr(true), IsDefault: true}},
		VariantOptionValues: []ProductVariantOptionValueInput{{SpecDefinitionID: definitionID, ValueKey: "standard", Label: "Standard", IsEnabled: boolPtr(true), InventoryPolicy: ProductCustomOptionInventoryComponent, ComponentVariantID: &component.Variants[0].ID, ComponentQuantity: 2}},
	})
	require.NoError(t, err)
	require.Len(t, created.VariantOptionValues, 1)
	policy := created.VariantOptionValues[0].CustomOptionPolicy
	require.NotNil(t, policy)
	require.Equal(t, ProductCustomOptionInventoryComponent, policy.InventoryPolicy)
	require.Equal(t, component.Variants[0].ID, *policy.ComponentVariantID)
	require.Equal(t, 2, policy.ComponentQuantity)

	var persisted productdomain.ProductCustomOptionPolicy
	require.NoError(t, db.Where("product_variant_option_value_id = ?", created.VariantOptionValues[0].ID).First(&persisted).Error)
	require.Equal(t, component.Variants[0].ID, *persisted.ComponentVariantID)
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
