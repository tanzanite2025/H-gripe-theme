package payment

import "testing"

func TestLoadConfigFromEnvReadsStripePaymentMethodTypes(t *testing.T) {
	t.Setenv("STRIPE_API_KEY", "sk_test_123")
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_123")
	t.Setenv("STRIPE_PUBLISHABLE_KEY", "pk_test_123")
	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test")
	t.Setenv("STRIPE_PAYMENT_METHOD_TYPES", "card,klarna, card ,afterpay_clearpay")

	config := LoadConfigFromEnv(GatewayStripe)
	if len(config.PaymentMethodTypes) != 3 {
		t.Fatalf("unexpected payment method types: %#v", config.PaymentMethodTypes)
	}
	if config.PaymentMethodTypes[0] != "card" || config.PaymentMethodTypes[1] != "klarna" || config.PaymentMethodTypes[2] != "afterpay_clearpay" {
		t.Fatalf("unexpected normalized payment method types: %#v", config.PaymentMethodTypes)
	}
}
