package shipping

import "testing"

func TestNormalizeShippingCountryCodesReturnsStableUniqueJSON(t *testing.T) {
	got := NormalizeShippingCountryCodes(` ["us", "CN", "US", ""] `)
	if got != `["CN","US"]` {
		t.Fatalf("NormalizeShippingCountryCodes() = %q, want %q", got, `["CN","US"]`)
	}
}

func TestValidateShippingCountryCodesRejectsMalformedEntries(t *testing.T) {
	if err := ValidateShippingCountryCodes(`["US","CN"]`); err != nil {
		t.Fatalf("valid country scope rejected: %v", err)
	}
	if err := ValidateShippingCountryCodes(`["*"]`); err != nil {
		t.Fatalf("wildcard country scope rejected: %v", err)
	}
	if err := ValidateShippingCountryCodes(`["[US"]`); err == nil {
		t.Fatal("malformed country scope should be rejected")
	}
}

func TestSystemFreeShippingTemplateAllowsOnlySelectedCountries(t *testing.T) {
	template := ShippingTemplate{
		TemplateKind:          ShippingTemplateKindSystemFreeShipping,
		FreeShippingCountries: `["CA","US"]`,
	}

	if !template.AllowsShippingCountry("us") {
		t.Fatal("selected country should be allowed")
	}
	if template.AllowsShippingCountry("DE") {
		t.Fatal("unselected country should be rejected")
	}
	if (ShippingTemplate{TemplateKind: ShippingTemplateKindSystemFreeShipping}).AllowsShippingCountry("US") {
		t.Fatal("empty system country scope should reject every country")
	}
}

func TestSystemFreeShippingIdentityNormalizesKindAndTypeTogether(t *testing.T) {
	template := ShippingTemplate{
		TemplateKind: ShippingTemplateKindSystemFreeShipping,
		Type:         "weight",
	}
	template.NormalizeSystemFreeShippingTemplateIdentity()
	template.ApplySystemFreeShippingTemplatePricingDefaults()

	if template.Type != ShippingTemplateTypeSystemFreeShipping {
		t.Fatalf("system template type = %q, want %q", template.Type, ShippingTemplateTypeSystemFreeShipping)
	}
	if template.TemplateKind != ShippingTemplateKindSystemFreeShipping {
		t.Fatalf("system template kind = %q, want %q", template.TemplateKind, ShippingTemplateKindSystemFreeShipping)
	}
	if !template.IsSystemManaged {
		t.Fatal("system free-shipping template should be system-managed")
	}
	if !template.FreeShipping || template.FreeThresholdMinor != 0 || template.DefaultFeeMinor != 0 {
		t.Fatalf("system free-shipping money fields were not forced to zero/free: %+v", template)
	}
}

func TestSystemManagedCarrierTemplateIsNotSystemFreeShipping(t *testing.T) {
	template := ShippingTemplate{
		TemplateKind:    ShippingTemplateKindCarrier,
		Type:            "weight",
		IsSystemManaged: true,
	}

	if template.IsSystemFreeShippingTemplate() {
		t.Fatal("system-managed carrier template must not be treated as system free shipping")
	}
	if !template.AllowsShippingCountry("US") {
		t.Fatal("system-managed carrier template should keep ordinary country behavior")
	}
}
