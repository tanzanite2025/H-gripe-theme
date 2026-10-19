package admin

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCarbonRimRequestRejectsUnknownEngineeringFieldsAndPreservesFalse(t *testing.T) {
	var request productCreateRequest
	require.Error(t, json.Unmarshal([]byte(`{"product_type":"carbon_rim","carbon_rim_specification":{"invented_tire_pressure":100}}`), &request))
	require.Error(t, json.Unmarshal([]byte(`{"variants":[{"carbon_rim_specification":{"invented_hole_count":24}}]}`), &request))
	require.NoError(t, json.Unmarshal([]byte(`{"product_type":"carbon_rim","carbon_rim_specification":{"schema_version":1,"tubeless_ready":false},"variants":[{"id":42,"carbon_rim_specification":{"spoke_hole_count":24,"drilling_pattern":"uniform","finish_code":"matte-black","net_weight_grams":430}}]}`), &request))
	require.NotNil(t, request.CarbonRimSpecification.TubelessReady)
	require.False(t, *request.CarbonRimSpecification.TubelessReady)
	inputs := normalizeVariantRequests(request.Variants)
	require.Equal(t, uint(42), *inputs[0].ID)
	require.Equal(t, 24, inputs[0].CarbonRimSpecification.SpokeHoleCount)
}
