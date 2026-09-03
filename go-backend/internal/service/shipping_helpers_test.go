package service

import "testing"

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
			if got := shippingRuleMatchesCountry(tt.region, tt.country); got != tt.want {
				t.Fatalf("shippingRuleMatchesCountry(%q, %q) = %v, want %v", tt.region, tt.country, got, tt.want)
			}
		})
	}
}
