package product

import (
	"encoding/json"
	"strings"
	"testing"

	productdomain "commerce-platform/internal/domain/product"

	"github.com/stretchr/testify/require"
)

func TestPublicProductDoesNotExposeRetiredTemplateOptionData(t *testing.T) {
	item := productdomain.Product{
		ID:     7,
		Name:   "Configured product",
		Slug:   "configured-product",
		Status: "active",
		Variants: []productdomain.ProductVariant{{
			ID: 10, ProductID: 7, SKU: "SKU-1", Title: "SKU-1", OptionValues: `{"color":"black"}`, IsActive: true,
		}},
		VariantOptionValues: []productdomain.ProductVariantOptionValue{{
			ID: 51, ProductID: 7, SpecDefinitionID: 8, ValueKey: "black", Label: "Black", SwatchURL: "/uploads/swatches/black.webp", IsEnabled: true,
		}},
		OptionValueRelations: []productdomain.ProductOptionValueRelation{
			{ID: 31, ProductID: 7, SourceOptionValueID: 101, TargetOptionValueID: 202, RelationType: productdomain.OptionValueRelationRequires},
			{ID: 0, ProductID: 7, SourceOptionValueID: 202, TargetOptionValueID: 303, RelationType: productdomain.OptionValueRelationConflicts},
			{ID: 32, ProductID: 7, SourceOptionValueID: 303, TargetOptionValueID: 404, RelationType: "unsupported"},
		},
	}

	payload, err := json.Marshal(PublicProductFromDomain(item))
	require.NoError(t, err)
	for _, retiredField := range []string{"product_specification_template", "spec_values", "variant_option_values", "option_value_relations", "option_values", "spec_definition_id"} {
		require.False(t, strings.Contains(string(payload), `"`+retiredField+`"`), "retired product field %q was exposed: %s", retiredField, payload)
	}
}
