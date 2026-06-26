package payment

import (
	"context"
	"testing"
)

func TestValidatePaymentRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *PaymentRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &PaymentRequest{
				Amount:      99.99,
				Currency:    "USD",
				OrderID:     "ORD-001",
				Description: "Test payment",
				Customer: &Customer{
					Email: "test@example.com",
					Name:  "Test User",
				},
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "invalid amount",
			req: &PaymentRequest{
				Amount:   0,
				Currency: "USD",
				OrderID:  "ORD-001",
				Customer: &Customer{
					Email: "test@example.com",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid currency",
			req: &PaymentRequest{
				Amount:   99.99,
				Currency: "INVALID",
				OrderID:  "ORD-001",
				Customer: &Customer{
					Email: "test@example.com",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			req: &PaymentRequest{
				Amount:   99.99,
				Currency: "USD",
				OrderID:  "ORD-001",
				Customer: &Customer{
					Email: "invalid-email",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePaymentRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePaymentRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRefundAmount(t *testing.T) {
	tests := []struct {
		name           string
		amount         float64
		originalAmount float64
		wantErr        bool
	}{
		{
			name:           "valid partial refund",
			amount:         50.00,
			originalAmount: 99.99,
			wantErr:        false,
		},
		{
			name:           "valid full refund",
			amount:         99.99,
			originalAmount: 99.99,
			wantErr:        false,
		},
		{
			name:           "invalid zero amount",
			amount:         0,
			originalAmount: 99.99,
			wantErr:        true,
		},
		{
			name:           "invalid excessive amount",
			amount:         150.00,
			originalAmount: 99.99,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRefundAmount(tt.amount, tt.originalAmount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRefundAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMockPaymentGateway(t *testing.T) {
	gateway := NewMockPaymentGateway()
	ctx := context.Background()

	// Test CreatePayment
	req := &PaymentRequest{
		Amount:      99.99,
		Currency:    "USD",
		OrderID:     "TEST-001",
		Description: "Test payment",
		Customer: &Customer{
			Email: "test@example.com",
			Name:  "Test User",
		},
	}

	resp, err := gateway.CreatePayment(ctx, req)
	if err != nil {
		t.Errorf("CreatePayment() error = %v", err)
	}

	if resp.Status != "succeeded" {
		t.Errorf("CreatePayment() status = %v, want %v", resp.Status, "succeeded")
	}

	// Test CapturePayment
	captureResp, err := gateway.CapturePayment(ctx, resp.ID)
	if err != nil {
		t.Errorf("CapturePayment() error = %v", err)
	}

	if captureResp.Status != "succeeded" {
		t.Errorf("CapturePayment() status = %v, want %v", captureResp.Status, "succeeded")
	}

	// Test RefundPayment
	refundResp, err := gateway.RefundPayment(ctx, resp.ID, 50.00)
	if err != nil {
		t.Errorf("RefundPayment() error = %v", err)
	}

	if refundResp.Amount != 50.00 {
		t.Errorf("RefundPayment() amount = %v, want %v", refundResp.Amount, 50.00)
	}

	// Test GetPayment
	getResp, err := gateway.GetPayment(ctx, resp.ID)
	if err != nil {
		t.Errorf("GetPayment() error = %v", err)
	}

	if getResp.ID != resp.ID {
		t.Errorf("GetPayment() ID = %v, want %v", getResp.ID, resp.ID)
	}

	// Test VerifyWebhook
	valid, err := gateway.VerifyWebhook([]byte("test payload"), "test signature")
	if err != nil {
		t.Errorf("VerifyWebhook() error = %v", err)
	}

	if !valid {
		t.Error("VerifyWebhook() should return true for mock gateway")
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	config := LoadConfigFromEnv(GatewayStripe)

	if config.Type != GatewayStripe {
		t.Errorf("LoadConfigFromEnv() Type = %v, want %v", config.Type, GatewayStripe)
	}

	// Default environment should be sandbox
	if config.Environment != "sandbox" && config.Environment != "" {
		t.Errorf("LoadConfigFromEnv() Environment = %v, want %v or empty", config.Environment, "sandbox")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Type:          GatewayStripe,
				APIKey:        "test_key",
				SecretKey:     "test_secret",
				WebhookSecret: "test_webhook_secret",
				Environment:   "sandbox",
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "missing type",
			config: &Config{
				APIKey:      "test_key",
				SecretKey:   "test_secret",
				Environment: "sandbox",
			},
			wantErr: true,
		},
		{
			name: "missing API key",
			config: &Config{
				Type:        GatewayStripe,
				SecretKey:   "test_secret",
				Environment: "sandbox",
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			config: &Config{
				Type:        GatewayStripe,
				APIKey:      "test_key",
				SecretKey:   "test_secret",
				Environment: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
