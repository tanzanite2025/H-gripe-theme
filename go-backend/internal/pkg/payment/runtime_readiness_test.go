package payment

import "testing"

func TestBuildRuntimeReadinessMarksStripeReadyWhenEnvConfigured(t *testing.T) {
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_ready")
	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_ready")

	readiness := BuildRuntimeReadiness("https://shop.example.com")
	stripeStatus := findRuntimeStatus(t, readiness, GatewayStripe)

	if !stripeStatus.ProductionReady {
		t.Fatalf("expected Stripe to be production ready, missing=%v blockers=%v", stripeStatus.Missing, stripeStatus.Blockers)
	}
	if !stripeStatus.Configured || !stripeStatus.WebhookConfigured {
		t.Fatalf("expected Stripe credentials and webhook to be configured")
	}
	if stripeStatus.CallbackURL != "https://shop.example.com/api/v1/payment/webhook/stripe" {
		t.Fatalf("unexpected callback URL: %s", stripeStatus.CallbackURL)
	}
}

func TestBuildRuntimeReadinessLocksProvidersWithoutSafeWebhookVerification(t *testing.T) {
	readiness := BuildRuntimeReadiness("")

	for _, gatewayType := range []GatewayType{GatewayPayPal, GatewayAlipay, GatewayWechat} {
		status := findRuntimeStatus(t, readiness, gatewayType)
		if status.ProductionReady {
			t.Fatalf("%s must not be production ready before official webhook verification is implemented", gatewayType)
		}
		if status.WebhookSupported {
			t.Fatalf("%s must not advertise supported webhooks yet", gatewayType)
		}
		if len(status.Blockers) == 0 {
			t.Fatalf("%s should expose a production blocker", gatewayType)
		}
	}
}

func findRuntimeStatus(t *testing.T, readiness RuntimeReadiness, gatewayType GatewayType) GatewayRuntimeStatus {
	t.Helper()
	for _, status := range readiness.Gateways {
		if status.Provider == gatewayType {
			return status
		}
	}
	t.Fatalf("missing runtime status for %s", gatewayType)
	return GatewayRuntimeStatus{}
}
