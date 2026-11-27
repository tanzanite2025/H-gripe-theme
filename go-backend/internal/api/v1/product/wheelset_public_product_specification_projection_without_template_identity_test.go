package product

import (
	"encoding/json"
	"strings"
	"testing"

	productdomain "commerce-platform/internal/domain/product"
)

func TestWheelsetPublicResponseUsesIndependentProductFactsWithoutTemplateValues(t *testing.T) {
	rimDepth := 45.0
	item := productdomain.Product{
		ProductType: productdomain.ProductTypeWheelset,
		WheelsetSpecification: &productdomain.WheelsetSpecification{
			SchemaVersion: 1,
			FrontWheel: productdomain.WheelsetProductWheelSpecification{
				WheelSize: "700C", SpokeHoleCounts: []int{24, 28}, LacingPatterns: []string{"2x", "3x"},
				RimDepthMM: &rimDepth, Material: "Carbon fiber",
			},
			RearWheel: productdomain.WheelsetProductWheelSpecification{
				WheelSize: "650B", SpokeHoleCounts: []int{28, 32}, LacingPatterns: []string{"3x"},
				Material: "Carbon fiber", FreehubBody: "XDR",
			},
		},
		SpecValues: []productdomain.ProductSpecValue{{
			Value: "legacy template value",
			SpecDefinition: &productdomain.SpecDefinition{
				Name: "Legacy template value", Slug: "legacy-template-value", IsVisible: true,
			},
		}},
		ProductSpecificationTemplate: &productdomain.ProductSpecificationTemplate{
			Name: "Retired wheelset template", Slug: "wheelset",
		},
	}

	response := PublicProductFromDomainWithLocale(item, "", "zh_cn")
	if response.WheelsetSpecification != item.WheelsetSpecification {
		t.Fatal("public response did not expose the independent wheelset specification")
	}
	values := make(map[string]string, len(response.SpecValues))
	for _, value := range response.SpecValues {
		if value.Definition == nil {
			continue
		}
		if value.Definition.Slug == "legacy-template-value" {
			t.Fatal("wheelset response included a legacy template value")
		}
		values[value.Definition.Slug] = value.Value
	}
	for slug, expected := range map[string]string{
		"front-wheel-size":       "700C",
		"rear-wheel-size":        "650B",
		"front-spoke-hole-count": "24, 28",
		"rear-spoke-hole-count":  "28, 32",
		"front-lacing-pattern":   "2x, 3x",
		"rear-lacing-pattern":    "3x",
		"front-rim-depth":        "45",
		"rear-freehub-body":      "XDR",
	} {
		if values[slug] != expected {
			t.Fatalf("public specification %q = %q, want %q", slug, values[slug], expected)
		}
	}
	if _, exists := values["rear-rim-depth"]; exists {
		t.Fatal("an empty rear rim depth was projected as a product fact")
	}
	if len(response.SpecValues) == 0 || response.SpecValues[0].Definition.Group != "轮组规格" {
		t.Fatal("public wheelset specification rows were not localized")
	}

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal wheelset public response: %v", err)
	}
	if !strings.Contains(string(payload), `"wheelset_specification"`) {
		t.Fatalf("public response omitted wheelset_specification: %s", payload)
	}
	if strings.Contains(string(payload), "legacy template value") {
		t.Fatalf("public response leaked legacy template data: %s", payload)
	}
}
