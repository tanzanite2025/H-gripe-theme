package admin

import (
	"encoding/json"
	"testing"

	"commerce-platform/internal/service"

	"github.com/stretchr/testify/require"
)

func TestYanwenPublishedCollectionReferenceDTOContainsOnlyCrossDomainFields(t *testing.T) {
	reference := newYanwenPublishedCollectionReferenceDTO(service.YanwenPublishedCollectionReference{
		ID:          12,
		ProductCode: "481",
		DisplayName: "燕文专线",
		Countries:   `["US","CA"]`,
		Enabled:     true,
	})

	payload, err := json.Marshal(reference)
	require.NoError(t, err)
	require.JSONEq(t, `{"id":12,"product_code":"481","display_name":"燕文专线","countries":"[\"US\",\"CA\"]","enabled":true}`, string(payload))
	require.NotContains(t, string(payload), "require_ioss")
	require.NotContains(t, string(payload), "max_weight_grams")
	require.NotContains(t, string(payload), "notes")
}
