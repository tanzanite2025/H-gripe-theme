package coupon

import (
	"testing"

	domainmoney "commerce-platform/internal/domain/money"
)

func TestCalculateDiscountCapsDiscountAtAmount(t *testing.T) {
	tests := []struct {
		name        string
		coupon      Coupon
		amountMinor int64
		wantMinor   int64
	}{
		{
			name: "fixed discount cannot exceed subtotal",
			coupon: Coupon{
				Type:      "fixed",
				Value:     50,
				MinAmount: 25,
			},
			amountMinor: 3000,
			wantMinor:   3000,
		},
		{
			name: "percentage discount cannot exceed subtotal",
			coupon: Coupon{
				Type:  "percentage",
				Value: 200,
			},
			amountMinor: 3000,
			wantMinor:   3000,
		},
		{
			name: "max discount still applies below subtotal",
			coupon: Coupon{
				Type:        "percentage",
				Value:       80,
				MaxDiscount: 20,
			},
			amountMinor: 10000,
			wantMinor:   2000,
		},
		{
			name:        "percentage discount rounds once in minor units",
			coupon:      Coupon{Type: "percentage", Value: 5.5},
			amountMinor: 333,
			wantMinor:   18,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount := domainmoney.MustNew(tt.amountMinor, "USD")
			got, err := tt.coupon.CalculateDiscountMoney(amount)
			if err != nil {
				t.Fatalf("CalculateDiscountMoney() error = %v", err)
			}
			if got.AmountMinor() != tt.wantMinor {
				t.Fatalf("CalculateDiscountMoney() = %d minor units, want %d", got.AmountMinor(), tt.wantMinor)
			}
		})
	}
}

func TestCalculateDiscountMoneyRejectsCurrencyMismatch(t *testing.T) {
	coupon := Coupon{Type: "fixed", Value: 5, Currency: "EUR"}
	amount := domainmoney.MustNew(1000, "USD")

	if _, err := coupon.CalculateDiscountMoney(amount); err == nil {
		t.Fatal("CalculateDiscountMoney() expected currency mismatch error")
	}
}
