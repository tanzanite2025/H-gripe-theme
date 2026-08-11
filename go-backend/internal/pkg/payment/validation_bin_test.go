package payment

import "testing"

func TestNormalizeCardBIN(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "six digits", value: "411111", want: "411111"},
		{name: "eight digits", value: "4111-1111", want: "41111111"},
		{name: "empty optional value", value: "", want: ""},
		{name: "full PAN rejected", value: "4111111111111111", wantErr: true},
		{name: "letters rejected", value: "41111A", wantErr: true},
		{name: "too short", value: "41111", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeCardBIN(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NormalizeCardBIN() error = %v, wantErr %t", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("NormalizeCardBIN() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidatePaymentRequestNormalizesCardBIN(t *testing.T) {
	req := &PaymentRequest{
		Amount:   10,
		Currency: "USD",
		OrderID:  "ORD-001",
		CardBIN:  "4111-1111",
		Customer: &Customer{Email: "test@example.com"},
	}
	if err := ValidatePaymentRequest(req); err != nil {
		t.Fatalf("ValidatePaymentRequest() error = %v", err)
	}
	if req.CardBIN != "41111111" {
		t.Fatalf("CardBIN = %q, want normalized 8-digit BIN", req.CardBIN)
	}
}
