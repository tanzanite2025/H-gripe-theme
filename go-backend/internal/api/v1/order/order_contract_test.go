package order

import (
	"encoding/json"
	"testing"
	"time"

	orderdomain "commerce-platform/internal/domain/order"
)

func TestPublicOrderResponseOmitsInternalDatabaseIDs(t *testing.T) {
	variantID := uint(17)
	item := orderdomain.Order{
		ID:          42,
		OrderNumber: "TZ-2026-ABCDEFGHIJKLMNOPQRST",
		UserID:      99,
		Items: []orderdomain.OrderItem{
			{
				ID:        84,
				OrderID:   42,
				ProductID: 5,
				VariantID: &variantID,
			},
		},
		CreatedAt: time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC),
	}

	payload, err := json.Marshal(publicOrderResponse(item))
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"id", "user_id", "order_id"} {
		if _, exists := decoded[forbidden]; exists {
			t.Fatalf("public order response leaked %s: %s", forbidden, payload)
		}
	}
	var orderNumber string
	if err := json.Unmarshal(decoded["order_number"], &orderNumber); err != nil {
		t.Fatal(err)
	}
	if orderNumber != "TZ-2026-ABCDEFGHIJKLMNOPQRST" {
		t.Fatalf("public order response order_number = %q", orderNumber)
	}
	var publicItems []map[string]json.RawMessage
	if err := json.Unmarshal(decoded["items"], &publicItems); err != nil {
		t.Fatal(err)
	}
	for _, publicItem := range publicItems {
		for _, forbidden := range []string{"id", "order_id"} {
			if _, exists := publicItem[forbidden]; exists {
				t.Fatalf("public order item leaked %s: %s", forbidden, payload)
			}
		}
	}
}
