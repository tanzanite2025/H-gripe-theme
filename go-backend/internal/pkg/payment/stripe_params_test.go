package payment

import "testing"

func TestStripeShippingDetailsParamsPreservesCompleteAddress(t *testing.T) {
	params := stripeShippingDetailsParams(&ShippingAddress{
		Name:       "Ada Lovelace",
		Line1:      "1 Analytical Engine Way",
		Line2:      "Suite 2",
		City:       "London",
		State:      "Greater London",
		PostalCode: "EC1A 1BB",
		Country:    "gb",
		Phone:      "+44 20 0000 0000",
	})

	if params == nil || params.Address == nil {
		t.Fatal("expected Stripe shipping details and address")
	}
	if *params.Name != "Ada Lovelace" || *params.Address.Line1 != "1 Analytical Engine Way" ||
		*params.Address.Line2 != "Suite 2" || *params.Address.City != "London" ||
		*params.Address.State != "Greater London" || *params.Address.PostalCode != "EC1A 1BB" ||
		*params.Address.Country != "GB" || *params.Phone != "+44 20 0000 0000" {
		t.Fatalf("shipping details did not preserve the complete address: %#v", params)
	}
}

func TestStripeShippingDetailsParamsHandlesNilAddress(t *testing.T) {
	if params := stripeShippingDetailsParams(nil); params != nil {
		t.Fatalf("nil address should not create Stripe shipping params: %#v", params)
	}
}
