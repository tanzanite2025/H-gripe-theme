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

func TestBuildRuntimeReadinessSupportsProviderWebhooksButRequiresCredentials(t *testing.T) {
	readiness := BuildRuntimeReadiness("")

	for _, gatewayType := range []GatewayType{GatewayPayPal, GatewayAlipay, GatewayWechat} {
		status := findRuntimeStatus(t, readiness, gatewayType)
		if status.ProductionReady {
			t.Fatalf("%s must not be production ready without credentials", gatewayType)
		}
		if !status.WebhookSupported {
			t.Fatalf("%s should advertise supported webhook verification", gatewayType)
		}
		if len(status.Missing) == 0 {
			t.Fatalf("%s should expose missing credential fields", gatewayType)
		}
	}
}

func TestBuildRuntimeReadinessMarksPayPalReadyWhenEnvConfigured(t *testing.T) {
	t.Setenv("PAYPAL_CLIENT_ID", "paypal-client")
	t.Setenv("PAYPAL_SECRET", "paypal-secret")
	t.Setenv("PAYPAL_WEBHOOK_ID", "paypal-webhook")

	readiness := BuildRuntimeReadiness("https://shop.example.com")
	status := findRuntimeStatus(t, readiness, GatewayPayPal)

	if !status.ProductionReady {
		t.Fatalf("expected PayPal to be production ready, missing=%v blockers=%v", status.Missing, status.Blockers)
	}
	if !status.Configured || !status.WebhookConfigured || !status.WebhookSupported {
		t.Fatalf("expected PayPal credentials and webhook to be configured: %#v", status)
	}
}

func TestBuildRuntimeReadinessMarksAlipayReadyWhenEnvConfigured(t *testing.T) {
	t.Setenv("ALIPAY_APP_ID", "alipay-app")
	t.Setenv("ALIPAY_PRIVATE_KEY", "alipay-private-key")
	t.Setenv("ALIPAY_PUBLIC_KEY", "alipay-public-key")

	readiness := BuildRuntimeReadiness("https://shop.example.com")
	status := findRuntimeStatus(t, readiness, GatewayAlipay)

	if !status.ProductionReady {
		t.Fatalf("expected Alipay to be production ready, missing=%v blockers=%v", status.Missing, status.Blockers)
	}
	if !status.Configured || !status.WebhookConfigured || !status.WebhookSupported {
		t.Fatalf("expected Alipay credentials and webhook to be configured: %#v", status)
	}
}

func TestBuildRuntimeReadinessMarksWechatReadyWithPlatformPublicKeyWhenEnvConfigured(t *testing.T) {
	t.Setenv("WECHAT_MCH_ID", "wechat-mch")
	t.Setenv("WECHAT_APP_ID", "wechat-app")
	t.Setenv("WECHAT_PRIVATE_KEY_PATH", "merchant-private-key.pem")
	t.Setenv("WECHAT_MERCHANT_SERIAL", "merchant-serial")
	t.Setenv("WECHAT_API_V3_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("WECHAT_PAY_PLATFORM_PUBLIC_KEY", "wechat-platform-public-key")
	t.Setenv("WECHAT_PAY_PLATFORM_PUBLIC_KEY_ID", "PUB_KEY_ID")

	readiness := BuildRuntimeReadiness("https://shop.example.com")
	status := findRuntimeStatus(t, readiness, GatewayWechat)

	if !status.ProductionReady {
		t.Fatalf("expected WeChat to be production ready, missing=%v blockers=%v", status.Missing, status.Blockers)
	}
	if !status.Configured || !status.WebhookConfigured || !status.WebhookSupported {
		t.Fatalf("expected WeChat credentials and webhook to be configured: %#v", status)
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
