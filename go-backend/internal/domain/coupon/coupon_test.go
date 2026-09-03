package coupon

import "testing"

func TestCalculateDiscountCapsDiscountAtAmount(t *testing.T) {
	tests := []struct {
		name   string
		coupon Coupon
		amount float64
		want   float64
	}{
		{
			name: "fixed discount cannot exceed subtotal",
			coupon: Coupon{
				Type:      "fixed",
				Value:     50,
				MinAmount: 25,
			},
			amount: 30,
			want:   30,
		},
		{
			name: "percentage discount cannot exceed subtotal",
			coupon: Coupon{
				Type:  "percentage",
				Value: 200,
			},
			amount: 30,
			want:   30,
		},
		{
			name: "max discount still applies below subtotal",
			coupon: Coupon{
				Type:        "percentage",
				Value:       80,
				MaxDiscount: 20,
			},
			amount: 100,
			want:   20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.coupon.CalculateDiscount(tt.amount); got != tt.want {
				t.Fatalf("CalculateDiscount() = %v, want %v", got, tt.want)
			}
		})
	}
}
