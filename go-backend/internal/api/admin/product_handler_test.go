package admin

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRejectProductSEORequestFieldsBlocksCatalogSEOFields(t *testing.T) {
	for _, field := range productSEORequestFields {
		requestBody := map[string]json.RawMessage{
			"name": json.RawMessage(`"Product"`),
			field:  json.RawMessage(`"Managed by SEO"`),
		}

		blockedField, blocked := rejectProductSEORequestFields(requestBody)

		require.True(t, blocked)
		require.Equal(t, field, blockedField)
	}
}

func TestRejectProductSEORequestFieldsAllowsCatalogFields(t *testing.T) {
	requestBody := map[string]json.RawMessage{
		"name":              json.RawMessage(`"Product"`),
		"slug":              json.RawMessage(`"product"`),
		"short_description": json.RawMessage(`"Summary"`),
		"description":       json.RawMessage(`"Body"`),
		"status":            json.RawMessage(`"active"`),
		"locale":            json.RawMessage(`"en"`),
		"variants":          json.RawMessage(`[]`),
		"media":             json.RawMessage(`[]`),
	}

	blockedField, blocked := rejectProductSEORequestFields(requestBody)

	require.False(t, blocked)
	require.Empty(t, blockedField)
}

func TestRetiredProductSpecificationFieldsAreRejectedEvenWhenEmpty(t *testing.T) {
	for _, body := range []string{
		`{"name":"Product","product_specification_template_id":null}`,
		`{"name":"Product","specs":{}}`,
		`{"name":"Product","variant_option_values":[]}`,
		`{"name":"Product","option_value_relations":[]}`,
		`{"name":"Product","variants":[{"option_values":{}}]}`,
		`{"name":"Product","media":[{"variant_option_value_id":null}]}`,
	} {
		var raw map[string]json.RawMessage
		require.NoError(t, json.Unmarshal([]byte(body), &raw))
		require.NotEmpty(t, retiredProductSpecificationRequestField(raw), body)
	}
}
