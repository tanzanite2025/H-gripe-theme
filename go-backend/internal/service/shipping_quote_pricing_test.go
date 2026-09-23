package service

import (
	"testing"

	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/shipping"
	shippingrating "commerce-platform/internal/domain/shipping/rating"
)

func TestTemplatePricingMoneyPathUsesMinorThresholdsWithoutFloatArithmetic(t *testing.T) {
	template := &shipping.ShippingTemplate{
		Type:            "price",
		Currency:        "USD",
		DefaultFeeMinor: 99,
		Rules: []shipping.ShippingRule{{
			Region:        "US",
			MinValueMinor: 1001,
			FeeMinor:      125,
		}},
	}

	below, free, _, err := calculateTemplateShippingFeeWithDisplayPricesMoney(
		template, "US", 0, 1, domainmoney.MustNew(1000, "USD"), domainmoney.MustNew(1000, "USD"),
	)
	if err != nil {
		t.Fatalf("below-threshold pricing failed: %v", err)
	}
	if free || below.AmountMinor() != 99 {
		t.Fatalf("below-threshold fee = %d/free=%v, want 99/false", below.AmountMinor(), free)
	}

	atThreshold, free, _, err := calculateTemplateShippingFeeWithDisplayPricesMoney(
		template, "US", 0, 1, domainmoney.MustNew(1001, "USD"), domainmoney.MustNew(1001, "USD"),
	)
	if err != nil {
		t.Fatalf("at-threshold pricing failed: %v", err)
	}
	if free || atThreshold.AmountMinor() != 125 {
		t.Fatalf("at-threshold fee = %d/free=%v, want 125/false", atThreshold.AmountMinor(), free)
	}
}

func TestShippingRuleMatchesCountrySupportsRegionMacros(t *testing.T) {
	tests := []struct {
		name    string
		region  string
		country string
		want    bool
	}{
		{name: "EU matches France", region: "EU", country: "FR", want: true},
		{name: "EU matches Italy", region: "EU", country: "IT", want: true},
		{name: "EU matches lowercase country in a region list", region: `["US", "EU"]`, country: " de ", want: true},
		{name: "EU aliases match", region: "european_union", country: "FR", want: true},
		{name: "EEA includes Iceland", region: "EEA", country: "IS", want: true},
		{name: "EEA includes Norway", region: "European Economic Area", country: "NO", want: true},
		{name: "EEA excludes Switzerland", region: "EEA", country: "CH", want: false},
		{name: "EU excludes United Kingdom", region: "EU", country: "GB", want: false},
		{name: "EU excludes Switzerland", region: "EU", country: "CH", want: false},
		{name: "direct country matching remains supported", region: "DE", country: "de", want: true},
		{name: "unknown macro does not match arbitrary countries", region: "NOT_A_REGION_GROUP", country: "JP", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shippingrating.MatchesRegion(tt.region, tt.country); got != tt.want {
				t.Fatalf("MatchesRegion(%q, %q) = %v, want %v", tt.region, tt.country, got, tt.want)
			}
		})
	}
}

func TestCarrierServiceMatchesCountrySupportsRegionMacros(t *testing.T) {
	tests := []struct {
		name      string
		countries string
		country   string
		want      bool
	}{
		{name: "EU matches Germany", countries: "EU", country: "DE", want: true},
		{name: "EU matches France in a region list", countries: `["US", "EU"]`, country: " fr ", want: true},
		{name: "EU alias matches", countries: "european_union", country: "IT", want: true},
		{name: "EEA matches Norway", countries: "EEA", country: "NO", want: true},
		{name: "EEA excludes Switzerland", countries: "EEA", country: "CH", want: false},
		{name: "EU excludes United Kingdom", countries: "EU", country: "GB", want: false},
		{name: "direct country matching remains supported", countries: "DE", country: "de", want: true},
		{name: "unknown macro does not match arbitrary countries", countries: "NOT_A_REGION_GROUP", country: "JP", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shippingrating.MatchesRegion(tt.countries, tt.country); got != tt.want {
				t.Fatalf("MatchesRegion(%q, %q) = %v, want %v", tt.countries, tt.country, got, tt.want)
			}
		})
	}
}

func TestCalculateRuleAdditionalUnitsUsesCarrierWeightSteps(t *testing.T) {
	tests := []struct {
		name       string
		ruleMinKg  float64
		valueKg    float64
		firstGrams int
		stepGrams  int
		want       int
	}{
		{name: "500g step charges 300g excess", ruleMinKg: 0.5, valueKg: 0.8, firstGrams: 500, stepGrams: 500, want: 1},
		{name: "500g step charges exact additional unit", ruleMinKg: 0.5, valueKg: 1.0, firstGrams: 500, stepGrams: 500, want: 1},
		{name: "500g step charges two units", ruleMinKg: 0.5, valueKg: 1.01, firstGrams: 500, stepGrams: 500, want: 2},
		{name: "100g step rounds up", ruleMinKg: 0.1, valueKg: 0.11, firstGrams: 100, stepGrams: 100, want: 1},
		{name: "10g step rounds up", ruleMinKg: 0.01, valueKg: 0.011, firstGrams: 10, stepGrams: 10, want: 1},
		{name: "rule minimum does not replace carrier first weight", ruleMinKg: 1, valueKg: 1.1, firstGrams: 500, stepGrams: 500, want: 2},
		{name: "missing first weight does not invent a first unit", ruleMinKg: 0, valueKg: 0.11, firstGrams: 0, stepGrams: 100, want: 0},
		{name: "at first weight is free", ruleMinKg: 0.5, valueKg: 0.5, firstGrams: 500, stepGrams: 500, want: 0},
	}

	rule := shipping.ShippingRule{AdditionalMinor: 100}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule.MinValue = tt.ruleMinKg
			got := calculateRuleAdditionalUnits(rule, tt.valueKg, shippingWeightBilling{
				firstWeightGrams:      tt.firstGrams,
				additionalWeightGrams: tt.stepGrams,
			})
			if got != tt.want {
				t.Fatalf("calculateRuleAdditionalUnits(%+v, %v) = %d, want %d", rule, tt.valueKg, got, tt.want)
			}
		})
	}
}

func TestCalculateRuleAdditionalUnitsRequiresConfiguredWeightStep(t *testing.T) {
	rule := shipping.ShippingRule{MinValue: 0, AdditionalMinor: 100}
	if got := calculateRuleAdditionalUnits(rule, 1.01); got != 0 {
		t.Fatalf("unconfigured additional units = %d, want 0", got)
	}
}

func TestValidateTemplateWeightBillingRejectsUnconfiguredAdditionalStep(t *testing.T) {
	template := &shipping.ShippingTemplate{
		Type: "weight", Currency: "USD", DefaultFeeMinor: 0,
		Rules: []shipping.ShippingRule{{Region: "US", MinValue: 0, MaxValue: 10, FeeMinor: 500, AdditionalMinor: 100}},
	}
	err := validateTemplateWeightBillingForValue(template, "US", 1.1)
	if err == nil {
		t.Fatal("expected missing carrier weight step error")
	}
}
