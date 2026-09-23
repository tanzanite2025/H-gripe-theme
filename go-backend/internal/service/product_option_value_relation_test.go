package service

import (
	"testing"

	productdomain "commerce-platform/internal/domain/product"

	"github.com/stretchr/testify/require"
)

func TestAdminProductOptionValueRelationsPersistNormalizeAndReplace(t *testing.T) {
	db, productService := newTestProductService(t)
	require.NoError(t, db.AutoMigrate(&productdomain.ProductOptionValueRelation{}))

	template := createOptionValueRelationTemplate(t, productService)
	created := createOptionValueRelationProduct(t, productService, template.ID, "relation-primary", "REL-PRIMARY")
	valueIDs := optionValueRelationIDsByKey(t, created)

	lowID, highID := valueIDs["alpha"], valueIDs["gamma"]
	if lowID > highID {
		lowID, highID = highID, lowID
	}
	updated, err := productService.UpdateAdminProduct(created.ID, ProductUpdateInput{
		OptionValueRelations: []ProductOptionValueRelationInput{
			{SourceOptionValueID: highID, TargetOptionValueID: lowID, RelationType: " CONFLICTS "},
			{SourceOptionValueID: lowID, TargetOptionValueID: highID, RelationType: productdomain.OptionValueRelationConflicts},
			{SourceOptionValueID: valueIDs["alpha"], TargetOptionValueID: valueIDs["beta"], RelationType: productdomain.OptionValueRelationRequires},
		},
		UpdateOptionValueRelations: true,
	})
	require.NoError(t, err)
	require.Len(t, updated.OptionValueRelations, 2)
	requireOptionValueRelation(t, updated.OptionValueRelations, productdomain.OptionValueRelationConflicts, lowID, highID)
	requireOptionValueRelation(t, updated.OptionValueRelations, productdomain.OptionValueRelationRequires, valueIDs["alpha"], valueIDs["beta"])

	reloaded, err := productService.GetByID(created.ID)
	require.NoError(t, err)
	require.Len(t, reloaded.OptionValueRelations, 2)
	for _, relation := range reloaded.OptionValueRelations {
		require.NotZero(t, relation.ID)
		require.Equal(t, created.ID, relation.ProductID)
	}
	requireOptionValueRelation(t, reloaded.OptionValueRelations, productdomain.OptionValueRelationConflicts, lowID, highID)
	requireOptionValueRelation(t, reloaded.OptionValueRelations, productdomain.OptionValueRelationRequires, valueIDs["alpha"], valueIDs["beta"])

	var persisted []productdomain.ProductOptionValueRelation
	require.NoError(t, db.Where("product_id = ?", created.ID).Find(&persisted).Error)
	require.Len(t, persisted, 2)

	cleared, err := productService.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateOptionValueRelations: true})
	require.NoError(t, err)
	require.Empty(t, cleared.OptionValueRelations)
	require.NoError(t, db.Where("product_id = ?", created.ID).Find(&persisted).Error)
	require.Empty(t, persisted)
}

func TestAdminProductOptionValueRelationsRejectInvalidGraph(t *testing.T) {
	db, productService := newTestProductService(t)
	require.NoError(t, db.AutoMigrate(&productdomain.ProductOptionValueRelation{}))

	template := createOptionValueRelationTemplate(t, productService)
	first := createOptionValueRelationProduct(t, productService, template.ID, "relation-first", "REL-FIRST")
	second := createOptionValueRelationProduct(t, productService, template.ID, "relation-second", "REL-SECOND")
	firstIDs := optionValueRelationIDsByKey(t, first)
	secondIDs := optionValueRelationIDsByKey(t, second)

	tests := []struct {
		name      string
		relations []ProductOptionValueRelationInput
	}{
		{
			name: "self relation",
			relations: []ProductOptionValueRelationInput{{
				SourceOptionValueID: firstIDs["alpha"], TargetOptionValueID: firstIDs["alpha"], RelationType: productdomain.OptionValueRelationRequires,
			}},
		},
		{
			name: "cross product relation",
			relations: []ProductOptionValueRelationInput{{
				SourceOptionValueID: firstIDs["alpha"], TargetOptionValueID: secondIDs["beta"], RelationType: productdomain.OptionValueRelationRequires,
			}},
		},
		{
			name: "requires cycle",
			relations: []ProductOptionValueRelationInput{
				{SourceOptionValueID: firstIDs["alpha"], TargetOptionValueID: firstIDs["beta"], RelationType: productdomain.OptionValueRelationRequires},
				{SourceOptionValueID: firstIDs["beta"], TargetOptionValueID: firstIDs["gamma"], RelationType: productdomain.OptionValueRelationRequires},
				{SourceOptionValueID: firstIDs["gamma"], TargetOptionValueID: firstIDs["alpha"], RelationType: productdomain.OptionValueRelationRequires},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := productService.UpdateAdminProduct(first.ID, ProductUpdateInput{
				OptionValueRelations:       tt.relations,
				UpdateOptionValueRelations: true,
			})
			require.ErrorIs(t, err, ErrProductOptionRelationInvalid)

			var count int64
			require.NoError(t, db.Model(&productdomain.ProductOptionValueRelation{}).Where("product_id = ?", first.ID).Count(&count).Error)
			require.Zero(t, count)
		})
	}
}

func createOptionValueRelationTemplate(t *testing.T, productService *ProductService) *productdomain.ProductSpecificationTemplate {
	t.Helper()
	template, err := productService.CreateProductSpecificationTemplate(ProductSpecificationTemplateInput{
		Name: "Option relation template", Slug: "option_relation_template", IsEnabled: true,
		SpecDefinitions: []ProductSpecDefinitionInput{{
			Name: "Extras", Slug: "extras", FieldType: "select", Role: ProductSpecRoleCustomOption, SelectionMode: ProductSpecSelectionMultiple,
			OptionItems: []ProductSpecOptionItemInput{
				{ValueKey: "alpha", DefaultLabel: "Alpha"},
				{ValueKey: "beta", DefaultLabel: "Beta"},
				{ValueKey: "gamma", DefaultLabel: "Gamma"},
			},
		}},
	})
	require.NoError(t, err)
	return template
}

func createOptionValueRelationProduct(t *testing.T, productService *ProductService, templateID uint, slug, sku string) *productdomain.Product {
	t.Helper()
	active := true
	created, err := productService.CreateAdminProduct(ProductCreateInput{
		ProductSpecificationTemplateID: &templateID,
		Name:                           slug,
		Slug:                           slug,
		Locale:                         "en",
		Status:                         "active",
		Currency:                       "USD",
		Variants: []ProductVariantInput{{
			SKU: sku, PriceMinor: 10000, Currency: "USD", Stock: 5, IsDefault: true, IsActive: &active,
		}},
	})
	require.NoError(t, err)
	require.Len(t, created.VariantOptionValues, 3)
	return created
}

func optionValueRelationIDsByKey(t *testing.T, item *productdomain.Product) map[string]uint {
	t.Helper()
	result := make(map[string]uint, len(item.VariantOptionValues))
	for _, value := range item.VariantOptionValues {
		require.NotZero(t, value.ID)
		result[value.ValueKey] = value.ID
	}
	for _, key := range []string{"alpha", "beta", "gamma"} {
		require.NotZero(t, result[key])
	}
	return result
}

func requireOptionValueRelation(t *testing.T, relations []productdomain.ProductOptionValueRelation, relationType string, sourceID, targetID uint) {
	t.Helper()
	for _, relation := range relations {
		if relation.RelationType == relationType && relation.SourceOptionValueID == sourceID && relation.TargetOptionValueID == targetID {
			return
		}
	}
	t.Fatalf("missing %s relation %d -> %d in %#v", relationType, sourceID, targetID, relations)
}
