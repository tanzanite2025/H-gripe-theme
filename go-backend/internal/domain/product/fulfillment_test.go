package product

import "testing"

func TestNormalizeFulfillmentModeDefaultsToStock(t *testing.T) {
	if got := NormalizeFulfillmentMode(""); got != FulfillmentModeStock {
		t.Fatalf("NormalizeFulfillmentMode(\"\") = %q, want %q", got, FulfillmentModeStock)
	}
	if got := NormalizeFulfillmentMode(" MADE_TO_ORDER "); got != FulfillmentModeMadeToOrder {
		t.Fatalf("NormalizeFulfillmentMode() = %q, want %q", got, FulfillmentModeMadeToOrder)
	}
}

func TestIsValidFulfillmentMode(t *testing.T) {
	if !IsValidFulfillmentMode(FulfillmentModeStock) || !IsValidFulfillmentMode(FulfillmentModeMadeToOrder) {
		t.Fatal("expected supported fulfillment modes to be valid")
	}
	if IsValidFulfillmentMode("unsupported") {
		t.Fatal("unsupported fulfillment mode should be invalid")
	}
}
