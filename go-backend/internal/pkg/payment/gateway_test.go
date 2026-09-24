package payment

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stripe/stripe-go/v76"
	stripeclient "github.com/stripe/stripe-go/v76/client"
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
				AmountMinor: 9999,
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
				AmountMinor: 0,
				Currency:    "USD",
				OrderID:     "ORD-001",
				Customer: &Customer{
					Email: "test@example.com",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid currency",
			req: &PaymentRequest{
				AmountMinor: 9999,
				Currency:    "INVALID",
				OrderID:     "ORD-001",
				Customer: &Customer{
					Email: "test@example.com",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			req: &PaymentRequest{
				AmountMinor: 9999,
				Currency:    "USD",
				OrderID:     "ORD-001",
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

func TestValidatePaymentRequestUsesExactMinorAmount(t *testing.T) {
	req := &PaymentRequest{
		AmountMinor: 123,
		Currency:    "JPY",
		OrderID:     "ORD-JPY-001",
		Customer:    &Customer{Email: "test@example.com"},
	}
	if err := ValidatePaymentRequest(req); err != nil {
		t.Fatalf("ValidatePaymentRequest() error = %v", err)
	}
	if req.AmountMinor != 123 {
		t.Fatalf("request amount minor = %v, want exact JPY amount 123", req.AmountMinor)
	}
	money, err := PaymentRequestMoney(req)
	if err != nil || money.AmountMinor() != 123 || money.Currency().String() != "JPY" {
		t.Fatalf("PaymentRequestMoney() = %+v, err=%v; want 123 JPY", money, err)
	}
}

func TestValidateRefundAmount(t *testing.T) {
	tests := []struct {
		name           string
		amount         int64
		originalAmount int64
		wantErr        bool
	}{
		{
			name:           "valid partial refund",
			amount:         5000,
			originalAmount: 9999,
			wantErr:        false,
		},
		{
			name:           "valid full refund",
			amount:         9999,
			originalAmount: 9999,
			wantErr:        false,
		},
		{
			name:           "invalid zero amount",
			amount:         0,
			originalAmount: 9999,
			wantErr:        true,
		},
		{
			name:           "invalid excessive amount",
			amount:         15000,
			originalAmount: 9999,
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
		AmountMinor: 9999,
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
	refundResp, err := gateway.RefundPayment(ctx, resp.ID, 5000)
	if err != nil {
		t.Errorf("RefundPayment() error = %v", err)
	}

	if refundResp.AmountMinor != 5000 {
		t.Errorf("RefundPayment() amount minor = %v, want %v", refundResp.AmountMinor, 5000)
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

func TestNewStripeGatewayInitializesClientWithoutTouchingGlobalKey(t *testing.T) {
	originalKey := stripe.Key
	defer func() {
		stripe.Key = originalKey
	}()

	stripe.Key = "sentinel_key"

	gateway, err := NewStripeGateway(&Config{
		Type:        GatewayStripe,
		APIKey:      "sk_test_123",
		Environment: "sandbox",
	})
	if err != nil {
		t.Fatalf("NewStripeGateway() error = %v", err)
	}
	if stripe.Key != "sentinel_key" {
		t.Fatalf("NewStripeGateway() mutated stripe.Key = %q, want %q", stripe.Key, "sentinel_key")
	}

	impl, ok := gateway.(*stripeGatewayImpl)
	if !ok {
		t.Fatalf("NewStripeGateway() type = %T, want *stripeGatewayImpl", gateway)
	}
	if impl.client == nil || impl.client.PaymentIntents == nil || impl.client.Refunds == nil {
		t.Fatal("NewStripeGateway() did not initialize Stripe client subclients")
	}
}

func TestStripePaymentIntentLiabilityShiftedFromExpandedCharge(t *testing.T) {
	tests := []struct {
		name   string
		result stripe.ChargePaymentMethodDetailsCardThreeDSecureResult
		want   *bool
	}{
		{
			name:   "authenticated shifts liability",
			result: stripe.ChargePaymentMethodDetailsCardThreeDSecureResultAuthenticated,
			want:   stripe.Bool(true),
		},
		{
			name:   "attempt acknowledged shifts liability",
			result: stripe.ChargePaymentMethodDetailsCardThreeDSecureResultAttemptAcknowledged,
			want:   stripe.Bool(true),
		},
		{
			name:   "failed authentication does not shift liability",
			result: stripe.ChargePaymentMethodDetailsCardThreeDSecureResultFailed,
			want:   stripe.Bool(false),
		},
		{
			name:   "exempted remains unknown",
			result: stripe.ChargePaymentMethodDetailsCardThreeDSecureResultExempted,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			intent := &stripe.PaymentIntent{
				LatestCharge: &stripe.Charge{
					PaymentMethodDetails: &stripe.ChargePaymentMethodDetails{
						Card: &stripe.ChargePaymentMethodDetailsCard{
							ThreeDSecure: &stripe.ChargePaymentMethodDetailsCardThreeDSecure{
								Result: test.result,
							},
						},
					},
				},
			}
			got := stripePaymentIntentLiabilityShifted(intent)
			if test.want == nil {
				if got != nil {
					t.Fatalf("expected unknown liability shift, got %t", *got)
				}
				return
			}
			if got == nil || *got != *test.want {
				t.Fatalf("liability shift = %v, want %t", got, *test.want)
			}
		})
	}
}

func TestStripeGetPaymentExpandsLatestChargeAndPersistsLiabilityShift(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "pi_liability_expanded",
			"amount": 120000,
			"amount_received": 120000,
			"currency": "usd",
			"status": "succeeded",
			"created": 1710000000,
			"latest_charge": {
				"id": "ch_liability_expanded",
				"payment_method_details": {
					"card": {
						"three_d_secure": {"result": "authenticated"}
					}
				}
			}
		}`))
	}))
	defer server.Close()

	backend := stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
		URL:        stripe.String(server.URL),
		HTTPClient: server.Client(),
	})
	client := stripeclient.New("sk_test_liability_expanded", &stripe.Backends{
		API:     backend,
		Connect: backend,
		Uploads: backend,
	})
	gateway := &stripeGatewayImpl{
		client: client,
		config: &Config{Type: GatewayStripe, APIKey: "sk_test_liability_expanded"},
	}

	response, err := gateway.GetPayment(context.Background(), "pi_liability_expanded")
	if err != nil {
		t.Fatalf("GetPayment() error = %v", err)
	}
	if response.LiabilityShifted == nil || !*response.LiabilityShifted {
		t.Fatalf("LiabilityShifted = %v, want true", response.LiabilityShifted)
	}
	if !strings.Contains(receivedPath, "expand[0]=latest_charge") {
		t.Fatalf("request URI = %q, want latest_charge expansion", receivedPath)
	}
}

func TestStripeGetPaymentRefetchesChargeWhenLatestChargeIsIDOnly(t *testing.T) {
	var receivedPaths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPaths = append(receivedPaths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if strings.HasPrefix(r.URL.Path, "/v1/charges/") {
			_, _ = w.Write([]byte(`{
				"id": "ch_liability_id_only",
				"payment_method_details": {
					"card": {"three_d_secure": {"liability_shifted": true}}
				}
			}`))
			return
		}
		_, _ = w.Write([]byte(`{
			"id": "pi_liability_id_only",
			"amount": 120000,
			"currency": "usd",
			"status": "succeeded",
			"latest_charge": "ch_liability_id_only"
		}`))
	}))
	defer server.Close()

	backend := stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
		URL:        stripe.String(server.URL),
		HTTPClient: server.Client(),
	})
	client := stripeclient.New("sk_test_liability_id_only", &stripe.Backends{
		API:     backend,
		Connect: backend,
		Uploads: backend,
	})
	gateway := &stripeGatewayImpl{
		client: client,
		config: &Config{Type: GatewayStripe, APIKey: "sk_test_liability_id_only"},
	}

	response, err := gateway.GetPayment(context.Background(), "pi_liability_id_only")
	if err != nil {
		t.Fatalf("GetPayment() error = %v", err)
	}
	if response.LiabilityShifted == nil || !*response.LiabilityShifted {
		t.Fatalf("LiabilityShifted = %v, want true", response.LiabilityShifted)
	}
	if len(receivedPaths) != 2 || receivedPaths[1] != "/v1/charges/ch_liability_id_only" {
		t.Fatalf("request paths = %#v, want PaymentIntent then Charge lookup", receivedPaths)
	}
}
