package admin

import (
	orderdomain "commerce-platform/internal/domain/order"
	"testing"
)

func TestOrderCustomsExportRow(t *testing.T) {
	declaredValue := 18.5
	record := orderdomain.Order{
		OrderNumber: "TZ-2026-ORDER",
		ShippingAddress: orderdomain.Address{
			FirstName:  "Ada",
			LastName:   "Lovelace",
			Country:    "US",
			PostalCode: "10001",
		},
	}
	item := orderdomain.OrderItem{
		ID:                     7,
		ProductName:            "Wheel rim",
		SKU:                    "RIM-001",
		Quantity:               2,
		HSCode:                 "871499",
		CNCode:                 "87149990",
		CountryOfOrigin:        "CN",
		CustomsDescription:     "Bicycle parts",
		DeclaredValue:          &declaredValue,
		DeclaredValueConfirmed: true,
	}

	row := orderCustomsExportRow(record, item)
	if got, want := row[0], "TZ-2026-ORDER"; got != want {
		t.Fatalf("order number = %q, want %q", got, want)
	}
	if got, want := row[1], "Ada Lovelace"; got != want {
		t.Fatalf("recipient = %q, want %q", got, want)
	}
	if got, want := row[14], "2"; got != want {
		t.Fatalf("quantity = %q, want %q", got, want)
	}
	if got, want := row[15], "871499"; got != want {
		t.Fatalf("HS code = %q, want %q", got, want)
	}
	if got, want := row[19], "18.50"; got != want {
		t.Fatalf("declared value = %q, want %q", got, want)
	}
	if got, want := row[20], "confirmed"; got != want {
		t.Fatalf("declared value status = %q, want %q", got, want)
	}
}
