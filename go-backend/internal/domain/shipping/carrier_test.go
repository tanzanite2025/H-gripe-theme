package shipping

import (
	"math/big"
	"testing"
)

func TestCarrierServiceFuelSurchargeRatePreservesDecimalPrecision(t *testing.T) {
	service := CarrierService{FuelSurchargePercentDecimal: "7.125000000000001"}
	rate, err := service.FuelSurchargeRate()
	if err != nil {
		t.Fatalf("parse fuel surcharge: %v", err)
	}
	want, _ := new(big.Rat).SetString("7.125000000000001")
	if rate.Cmp(want) != 0 {
		t.Fatalf("expected exact rate %s, got %s", want.RatString(), rate.RatString())
	}
}

func TestCarrierServiceFuelSurchargeRateRejectsInvalidDecimal(t *testing.T) {
	for _, value := range []string{"5/2", "-1", "100.000000000000001", "101"} {
		service := CarrierService{FuelSurchargePercentDecimal: value}
		if _, err := service.FuelSurchargeRate(); err == nil {
			t.Fatalf("expected fuel surcharge %q to be rejected", value)
		}
	}
}

func TestCarrierServiceFuelSurchargeRateNormalizesEmptyValue(t *testing.T) {
	service := CarrierService{}
	rate, err := service.FuelSurchargeRate()
	if err != nil {
		t.Fatalf("parse empty fuel surcharge: %v", err)
	}
	if rate.Sign() != 0 {
		t.Fatalf("expected empty fuel surcharge to normalize to zero, got %s", rate.RatString())
	}
}
