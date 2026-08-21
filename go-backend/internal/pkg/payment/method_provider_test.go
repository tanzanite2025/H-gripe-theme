package payment

import "testing"

func TestProviderForPaymentMethod(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{name: "card maps to stripe", code: "card", want: string(GatewayStripe)},
		{name: "credit card maps to stripe", code: "credit_card", want: string(GatewayStripe)},
		{name: "stripe maps to stripe", code: "stripe", want: string(GatewayStripe)},
		{name: "paypal maps to paypal", code: "paypal", want: string(GatewayPayPal)},
		{name: "alipay maps to alipay", code: "alipay", want: string(GatewayAlipay)},
		{name: "wechat pay alias maps to wechat", code: "wechat_pay", want: string(GatewayWechat)},
		{name: "unknown stays empty", code: "bank", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProviderForPaymentMethod(tt.code); got != tt.want {
				t.Fatalf("ProviderForPaymentMethod(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestGatewayCurrencyCapabilities(t *testing.T) {
	tests := []struct {
		name     string
		provider GatewayType
		currency string
		want     bool
	}{
		{name: "stripe accepts catalog currency", provider: GatewayStripe, currency: "usd", want: true},
		{name: "paypal accepts configured currency", provider: GatewayPayPal, currency: "EUR", want: true},
		{name: "paypal rejects unsupported catalog currency", provider: GatewayPayPal, currency: "AED", want: false},
		{name: "alipay is cny only", provider: GatewayAlipay, currency: "CNY", want: true},
		{name: "alipay rejects usd", provider: GatewayAlipay, currency: "USD", want: false},
		{name: "wechat is cny only", provider: GatewayWechat, currency: "CNY", want: true},
		{name: "unknown provider rejects currency", provider: GatewayType("bank"), currency: "USD", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GatewaySupportsCurrency(tt.provider, tt.currency); got != tt.want {
				t.Fatalf("GatewaySupportsCurrency(%q, %q) = %v, want %v", tt.provider, tt.currency, got, tt.want)
			}
			err := ValidateGatewayCurrency(tt.provider, tt.currency)
			if (err == nil) != tt.want {
				t.Fatalf("ValidateGatewayCurrency(%q, %q) error = %v, want supported %v", tt.provider, tt.currency, err, tt.want)
			}
		})
	}
}

func TestFormatMajorAmountUsesCurrencyMinorUnits(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		currency string
		want     string
	}{
		{name: "usd uses two decimals", amount: 123.45, currency: "USD", want: "123.45"},
		{name: "eur uses two decimals", amount: 123.4, currency: "EUR", want: "123.40"},
		{name: "jpy uses zero decimals", amount: 123, currency: "JPY", want: "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := FormatMajorAmount(tt.amount, tt.currency)
			if err != nil {
				t.Fatalf("FormatMajorAmount(%s) error = %v", tt.currency, err)
			}
			if value != tt.want {
				t.Fatalf("FormatMajorAmount(%s) = %q, want %q", tt.currency, value, tt.want)
			}
		})
	}
}
