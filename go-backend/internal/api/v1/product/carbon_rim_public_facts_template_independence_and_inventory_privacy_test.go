package product

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	productdomain "commerce-platform/internal/domain/product"
)

func TestCarbonRimPublicFactsHaveNoTemplateOrExactInventoryAndOmitUnknownLimits(t *testing.T) {
	tubeless := false
	item := productdomain.Product{ID: 7, ProductType: productdomain.ProductTypeCarbonRim, Status: "active", Name: "Carbon rim", CarbonRimSpecification: &productdomain.CarbonRimSpecification{SchemaVersion: 1, Model: "R45", BeadSeatDiameterMM: 622, DepthMM: 45, InternalWidthMM: 25, ExternalWidthMM: 32, BeadStructure: "hooked", TubelessReady: &tubeless}, Variants: []productdomain.ProductVariant{{ID: 9, ProductID: 7, SKU: "R45-24", IsActive: true, IsDefault: true, Stock: 123, Currency: "USD", PriceMinor: 19900, OptionValues: `{"rim_configuration":"24H / uniform / matte-black"}`, CarbonRimSpecification: &productdomain.CarbonRimVariantSpecification{SpokeHoleCount: 24, DrillingPattern: "uniform", FinishCode: "matte-black", NetWeightGrams: 430}}}}
	response := PublicProductFromDomainWithLocale(item, "USD", "zh_cn")
	require.Nil(t, response.ProductSpecificationTemplate)
	require.Equal(t, item.CarbonRimSpecification, response.CarbonRimSpecification)
	require.Len(t, response.Variants, 1)
	require.Equal(t, 430.0, response.Variants[0].CarbonRimSpecification.NetWeightGrams)
	values := map[string]string{}
	for _, row := range response.SpecValues {
		values[row.Definition.Slug] = row.Value
	}
	require.Equal(t, "false", values["tubeless_ready"])
	require.Equal(t, "622", values["bead_seat_diameter_mm"])
	require.NotContains(t, values, "max_pressure")
	require.NotContains(t, values, "rim_tape_required")
	require.NotContains(t, values, "wheel_size")
	require.NotContains(t, values, "riding_uses")
	encoded, err := json.Marshal(response)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), `"stock"`)
	require.NotContains(t, string(encoded), `"spec_definition_id"`)
}

func TestCarbonRimPublicWheelSizeMaterialAndRidingUsesMatchEnteredFactsInBothLocales(t *testing.T) {
	spec := &productdomain.CarbonRimSpecification{WheelSize: "700c", BeadSeatDiameterMM: 622, Material: "carbon_fiber", MaterialDetails: "T700 / T800", RidingUses: []string{"road", "gravel"}}
	for _, locale := range []string{"en", "zh_cn"} {
		rows := publicCarbonRimSpecificationValues(spec, locale)
		values := map[string]string{}
		for _, row := range rows {
			values[row.Definition.Slug] = row.Value
		}
		require.Equal(t, "700C", values["wheel_size"])
		require.Equal(t, "622", values["bead_seat_diameter_mm"])
		require.Contains(t, values["material"], "T700 / T800")
		if locale == "en" {
			require.Equal(t, "Road cycling / Gravel cycling", values["riding_uses"])
		} else {
			require.Equal(t, "公路骑行 / Gravel 碎石骑行", values["riding_uses"])
		}
	}
}
