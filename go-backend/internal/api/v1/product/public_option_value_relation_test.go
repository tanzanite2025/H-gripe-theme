package product

import (
	"encoding/json"
	"strings"
	"testing"

	productdomain "commerce-platform/internal/domain/product"

	"github.com/stretchr/testify/require"
)

func TestPublicProductExposesPersistedOptionValueRelations(t *testing.T) {
	item := productdomain.Product{
		ID:     7,
		Name:   "Configured product",
		Slug:   "configured-product",
		Status: "active",
		OptionValueRelations: []productdomain.ProductOptionValueRelation{
			{ID: 31, ProductID: 7, SourceOptionValueID: 101, TargetOptionValueID: 202, RelationType: productdomain.OptionValueRelationRequires},
			{ID: 0, ProductID: 7, SourceOptionValueID: 202, TargetOptionValueID: 303, RelationType: productdomain.OptionValueRelationConflicts},
			{ID: 32, ProductID: 7, SourceOptionValueID: 303, TargetOptionValueID: 404, RelationType: "unsupported"},
		},
	}

	publicProduct := PublicProductFromDomain(item)
	require.Equal(t, []PublicProductOptionValueRelation{{
		ID: 31, SourceOptionValueID: 101, TargetOptionValueID: 202, RelationType: productdomain.OptionValueRelationRequires,
	}}, publicProduct.OptionValueRelations)

	payload, err := json.Marshal(publicProduct)
	require.NoError(t, err)
	require.Contains(t, string(payload), `"option_value_relations":[{"id":31,"source_option_value_id":101,"target_option_value_id":202,"relation_type":"requires"}]`)
	require.False(t, strings.Contains(string(payload), `"product_id"`), "public relation must not expose its storage owner")
}
