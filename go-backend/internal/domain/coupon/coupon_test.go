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
				Type:           "fixed",
				ValueMinor:     5000,
				MinAmountMinor: 2500,
			},
			amountMinor: 3000,
			wantMinor:   3000,
		},
		{
			name: "percentage discount cannot exceed subtotal",
			coupon: Coupon{
				Type:             "percentage",
				ValueRateDecimal: "200",
			},
			amountMinor: 3000,
			wantMinor:   3000,
		},
		{
			name: "max discount still applies below subtotal",
			coupon: Coupon{
				Type:             "percentage",
				ValueRateDecimal: "80",
				MaxDiscountMinor: 2000,
			},
			amountMinor: 10000,
			wantMinor:   2000,
		},
		{
			name:        "percentage discount rounds once in minor units",
			coupon:      Coupon{Type: "percentage", ValueRateDecimal: "5.5"},
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
	coupon := Coupon{Type: "fixed", ValueMinor: 500, Currency: "EUR"}
	amount := domainmoney.MustNew(1000, "USD")

	if _, err := coupon.CalculateDiscountMoney(amount); err == nil {
		t.Fatal("CalculateDiscountMoney() expected currency mismatch error")
	}
}

func TestCalculateDiscountMoneyUsesMinorFixedAmountForZeroDecimalCurrency(t *testing.T) {
	coupon := Coupon{
		Type:           "fixed",
		Currency:       "JPY",
		ValueMinor:     100,
		MinAmountMinor: 1000,
	}
	got, err := coupon.CalculateDiscountMoney(domainmoney.MustNew(2500, "JPY"))
	if err != nil {
		t.Fatalf("CalculateDiscountMoney() error = %v", err)
	}
	if got.AmountMinor() != 100 {
		t.Fatalf("CalculateDiscountMoney() = %d minor units, want 100", got.AmountMinor())
	}
}

func TestCalculateDiscountMoneyUsesCanonicalDecimalPercentage(t *testing.T) {
	c := &Coupon{Type: "percentage", ValueRateDecimal: "12.345678901234567"}
	amount := domainmoney.MustNew(100000, "USD")
	discount, err := c.CalculateDiscountMoney(amount)
	if err != nil {
		t.Fatal(err)
	}
	// 1000.000... minor rounds to exactly 12345 minor units.
	if got, want := discount.AmountMinor(), int64(12346); got != want {
		t.Fatalf("discount minor = %d, want %d", got, want)
	}
}

func TestCalculateDiscountMoneyRequiresCanonicalAmounts(t *testing.T) {
	amount := domainmoney.MustNew(10000, "USD")
	if _, err := (&Coupon{Type: "fixed"}).CalculateDiscountMoney(amount); err == nil {
		t.Fatal("expected fixed coupon without value_minor to be rejected")
	}
	if _, err := (&Coupon{Type: "percentage"}).CalculateDiscountMoney(amount); err == nil {
		t.Fatal("expected percentage coupon without value_rate_decimal to be rejected")
	}
}
