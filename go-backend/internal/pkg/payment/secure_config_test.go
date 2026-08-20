package payment

import "testing"

func TestSecureGatewayConfigEncryptDecrypt(t *testing.T) {
	masterKey := "test-master-key-with-enough-length-123"
	encrypted, err := EncryptSecureGatewayConfig(SecureGatewayConfig{
		Provider:    GatewayStripe,
		Environment: "production",
		Credentials: map[string]string{
			"api_key":        "sk_live_123",
			"webhook_secret": "whsec_123",
		},
	}, masterKey)
	if err != nil {
		t.Fatalf("EncryptSecureGatewayConfig() error = %v", err)
	}
	if encrypted == "" || encrypted == "sk_live_123" {
		t.Fatalf("encrypted value should not expose plaintext")
	}

	decrypted, err := DecryptSecureGatewayConfig(encrypted, masterKey)
	if err != nil {
		t.Fatalf("DecryptSecureGatewayConfig() error = %v", err)
	}
	if decrypted.Provider != GatewayStripe || decrypted.Environment != "production" {
		t.Fatalf("unexpected decrypted config: %#v", decrypted)
	}
	if decrypted.Credentials["api_key"] != "sk_live_123" || decrypted.Credentials["webhook_secret"] != "whsec_123" {
		t.Fatalf("credentials did not round trip")
	}
}

func TestSecureGatewayConfigRequiresMasterKey(t *testing.T) {
	if _, err := EncryptSecureGatewayConfig(SecureGatewayConfig{Provider: GatewayStripe}, ""); err == nil {
		t.Fatalf("expected encrypt to require master key")
	}
	if _, err := DecryptSecureGatewayConfig("v1:test", ""); err == nil {
		t.Fatalf("expected decrypt to require master key")
	}
}

func TestDecodeStoredSecureGatewayConfigRequiresMatchingProvider(t *testing.T) {
	masterKey := "test-master-key-with-enough-length-456"
	t.Setenv(PaymentConfigMasterKeyEnv, masterKey)
	encrypted, err := EncryptSecureGatewayConfig(SecureGatewayConfig{
		Provider:    GatewayStripe,
		Environment: "production",
		Credentials: map[string]string{"api_key": "sk_live_123"},
	}, masterKey)
	if err != nil {
		t.Fatalf("EncryptSecureGatewayConfig() error = %v", err)
	}

	if _, err := DecodeStoredSecureGatewayConfig(encrypted, GatewayPayPal); err == nil {
		t.Fatalf("expected DecodeStoredSecureGatewayConfig to reject provider mismatch")
	}
}

func TestApplySecureGatewayStatusesOverridesStripeRuntimeSource(t *testing.T) {
	readiness := BuildRuntimeReadiness("https://shop.example.com")
	ApplySecureGatewayStatuses(&readiness, []SecureGatewayConfigStatus{
		{
			Provider:         GatewayStripe,
			Configured:       true,
			Readable:         true,
			RuntimeSource:    "admin-encrypted",
			ConfiguredFields: []string{"api_key", "webhook_secret"},
		},
	})

	status := findRuntimeStatus(t, readiness, GatewayStripe)
	if readiness.RuntimeSource != "mixed" {
		t.Fatalf("expected mixed runtime source, got %s", readiness.RuntimeSource)
	}
	if status.RuntimeSource != "admin-encrypted" || !status.ProductionReady {
		t.Fatalf("expected Stripe to use admin encrypted runtime and be ready: %#v", status)
	}
}

func TestNormalizeThreeDSecureModeSupportsChallenge(t *testing.T) {
	if got := NormalizeThreeDSecureMode("challenge"); got != "challenge" {
		t.Fatalf("NormalizeThreeDSecureMode(challenge) = %q, want challenge", got)
	}
}

func TestGatewayConfigFromSecureConfigReadsStripePaymentMethodTypesFromEnv(t *testing.T) {
	t.Setenv("STRIPE_PAYMENT_METHOD_TYPES", "card,klarna")
	config := GatewayConfigFromSecureConfig(SecureGatewayConfig{
		Provider:    GatewayStripe,
		Environment: "production",
		Credentials: map[string]string{
			"api_key":        "sk_live_123",
			"webhook_secret": "whsec_123",
		},
	})
	if len(config.PaymentMethodTypes) != 2 || config.PaymentMethodTypes[0] != "card" || config.PaymentMethodTypes[1] != "klarna" {
		t.Fatalf("unexpected payment method types: %#v", config.PaymentMethodTypes)
	}
}
