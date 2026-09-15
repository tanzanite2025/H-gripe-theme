package rating

import (
	"errors"
	"testing"
)

func TestMatchesRegion(t *testing.T) {
	tests := []struct {
		region  string
		country string
		want    bool
	}{
		{region: "EU", country: "FR", want: true},
		{region: `["US", "EU"]`, country: "de", want: true},
		{region: "EEA", country: "NO", want: true},
		{region: "EEA", country: "CH", want: false},
		{region: "EU", country: "GB", want: false},
		{region: "*", country: "JP", want: true},
	}
	for _, test := range tests {
		if got := MatchesRegion(test.region, test.country); got != test.want {
			t.Fatalf("MatchesRegion(%q, %q) = %v, want %v", test.region, test.country, got, test.want)
		}
	}
}

func TestAdditionalWeightUnitsUsesConfiguredGramScale(t *testing.T) {
	tests := []struct {
		name       string
		valueGrams int
		first      int
		step       int
		want       int
	}{
		{name: "500g partial continuation", valueGrams: 800, first: 500, step: 500, want: 1},
		{name: "500g exact continuation", valueGrams: 1000, first: 500, step: 500, want: 1},
		{name: "100g continuation", valueGrams: 110, first: 100, step: 100, want: 1},
		{name: "10g continuation", valueGrams: 11, first: 10, step: 10, want: 1},
		{name: "rule ranges do not affect the carrier scale", valueGrams: 1100, first: 500, step: 500, want: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := AdditionalWeightUnits(test.valueGrams, 1, WeightScale{
				FirstWeightGrams: test.first, AdditionalWeightGrams: test.step,
			})
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("units = %d, want %d", got, test.want)
			}
		})
	}
}

func TestAdditionalWeightUnitsRejectsMissingScale(t *testing.T) {
	_, err := AdditionalWeightUnits(1000, 1, WeightScale{})
	if !errors.Is(err, ErrInvalidWeightScale) {
		t.Fatalf("error = %v, want ErrInvalidWeightScale", err)
	}
}
