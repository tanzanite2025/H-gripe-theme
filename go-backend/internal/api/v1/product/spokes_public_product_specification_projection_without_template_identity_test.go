package product

import (
	"encoding/json"
	"strings"
	"testing"

	productdomain "commerce-platform/internal/domain/product"
)

func TestSpokePublicResponseExposesDedicatedFactsWithoutGenericTemplateData(t *testing.T) {
	middleWidth := 2.2
	middleThickness := 0.9
	item := productdomain.Product{
		ProductType:                   productdomain.ProductTypeSpoke,
		Name:                          "SAPIM CX-Ray",
		Slug:                          "sapim-cx-ray",
		Status:                        "inactive",
		MerchantProductIdentifier:     "PRIVATE-MERCHANT-ID",
		ManufacturerProductIdentifier: "PRIVATE-MANUFACTURER-ID",
		SpokeSpecification: &productdomain.SpokesProductSpecification{
			SchemaVersion:      1,
			ModelName:          "CX-Ray",
			ManufacturerFamily: "Aero spokes",
			GaugeOptions: []productdomain.SpokeGaugeOption{{
				Gauge:             "14G",
				MiddleWidthMM:     &middleWidth,
				MiddleThicknessMM: &middleThickness,
			}},
			MiddleStrength: &productdomain.SpokeMiddleStrengthRange{MinimumNPerMM2: 1600, MaximumNPerMM2: 1600},
			ColorOptions:   []string{"Black oxidized", "Silver"},
		},
		SpecValues: []productdomain.ProductSpecValue{{
			Value: "must not be projected",
			SpecDefinition: &productdomain.SpecDefinition{
				Name: "Generic template value", Slug: "generic-template-value", IsVisible: true,
			},
		}},
		ProductSpecificationTemplate: &productdomain.ProductSpecificationTemplate{
			Name: "Generic spoke template", Slug: "generic-spoke-template",
		},
	}

	response := PublicProductFromDomainWithLocale(item, "", "zh_cn")
	if response.SpokeSpecification != item.SpokeSpecification {
		t.Fatal("public response did not expose the dedicated spoke specification")
	}
	if response.ProductSpecificationTemplate != nil {
		t.Fatal("spoke response exposed a generic product specification template")
	}
	values := make(map[string]string, len(response.SpecValues))
	colorOptionCount := 0
	for _, value := range response.SpecValues {
		if value.Definition != nil {
			values[value.Definition.Slug] = value.Value
			if value.Definition.Slug == "color-option" {
				colorOptionCount++
			}
		}
	}
	if colorOptionCount != 2 {
		t.Fatalf("public spoke response color option count = %d, want 2", colorOptionCount)
	}
	for slug, expected := range map[string]string{
		"model_name":                  "CX-Ray",
		"manufacturer_family":         "Aero spokes",
		"gauge-1-middle-width-mm":     "2.2",
		"gauge-1-middle-thickness-mm": "0.9",
		"middle_strength_n_per_mm2":   "1600",
		"color-option":                "Silver",
	} {
		if values[slug] != expected {
			t.Fatalf("public spoke specification %q = %q, want %q", slug, values[slug], expected)
		}
	}
	for _, value := range response.SpecValues {
		if value.Definition != nil && value.Definition.Slug == "middle_strength_n_per_mm2" && value.Definition.Unit != productdomain.SpokeMiddleStrengthUnitLabel {
			t.Fatalf("Middle strength unit = %q, want %q", value.Definition.Unit, productdomain.SpokeMiddleStrengthUnitLabel)
		}
	}

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal public spoke response: %v", err)
	}
	serialized := string(payload)
	if !strings.Contains(serialized, `"spoke_specification"`) {
		t.Fatalf("public response omitted spoke_specification: %s", serialized)
	}
	for _, privateOrGenericValue := range []string{"PRIVATE-MERCHANT-ID", "PRIVATE-MANUFACTURER-ID", "must not be projected", "generic-spoke-template"} {
		if strings.Contains(serialized, privateOrGenericValue) {
			t.Fatalf("public spoke response leaked %q: %s", privateOrGenericValue, serialized)
		}
	}
}

func TestSpokePublicResponseProjectsOnlyVerifiedManufacturerAndStandardProductIdentifiers(t *testing.T) {
	item := productdomain.Product{
		ProductType:                   productdomain.ProductTypeSpoke,
		ManufacturerProductIdentifier: "PRIVATE-NOT-MPN",
		SpokeSpecification: &productdomain.SpokesProductSpecification{
			SchemaVersion: 1,
			ModelName:     "CX-Ray",
		},
	}
	if response := PublicProductFromDomain(item); response.MPN != "" {
		t.Fatalf("model name or internal manufacturer identifier was exposed as MPN: %q", response.MPN)
	}

	item.SpokeSpecification.ManufacturerPartNumber = "SAPIM-PART-123"
	item.StandardProductIdentifier = "4006381333931"
	response := PublicProductFromDomain(item)
	if response.MPN != "SAPIM-PART-123" {
		t.Fatalf("verified manufacturer part number = %q", response.MPN)
	}
	if response.GTIN != "4006381333931" {
		t.Fatalf("verified standard product identifier = %q", response.GTIN)
	}
	item.StandardProductIdentifier = "4006381333932"
	if response := PublicProductFromDomain(item); response.GTIN != "" {
		t.Fatalf("invalid standard product identifier exposed as GTIN: %q", response.GTIN)
	}
}
