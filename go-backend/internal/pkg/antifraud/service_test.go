package antifraud

import (
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/pkg/config"
)

func TestFailureKeysIncludeUserParentForSessionKey(t *testing.T) {
	got := failureKeys("user:42:session:abc")
	want := []string{"user:42:session:abc", "user:42"}

	if len(got) != len(want) {
		t.Fatalf("failureKeys length = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("failureKeys[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}

func TestFailureKeysKeepPlainKeySingle(t *testing.T) {
	got := failureKeys("user:42")
	if len(got) != 1 || got[0] != "user:42" {
		t.Fatalf("failureKeys = %#v, want [user:42]", got)
	}
}

func TestEvaluateSignalsScoresCardingSignalsAndAppliesDelay(t *testing.T) {
	service := &Service{
		config: config.PaymentRiskConfig{
			FailureThreshold:     2,
			DelaySeconds:         2,
			HighRiskScore:        60,
			FailureWindowSeconds: 600,
		},
	}

	decision := service.evaluateSignals(2, Signals{
		IPCountry:      "US",
		BillingCountry: "DE",
		VPNDetected:    true,
		UserAgent:      "",
	})

	if !decision.HighRisk {
		t.Fatal("expected decision to be high risk")
	}
	if decision.Score != 110 {
		t.Fatalf("score = %d, want 110", decision.Score)
	}
	if decision.Delay != 2*time.Second {
		t.Fatalf("delay = %s, want 2s", decision.Delay)
	}
	if len(decision.Reasons) != 4 {
		t.Fatalf("reasons = %#v, want four reasons", decision.Reasons)
	}
}

func TestAttemptIdentityFailureKeysHashSensitiveValues(t *testing.T) {
	identity := AttemptIdentity{
		Provider:    "Stripe",
		UserID:      "42",
		SessionID:   "session-secret",
		AnonymousID: "visitor-secret",
		IPAddress:   "203.0.113.9",
		UserAgent:   "Mozilla/5.0 Card Test",
	}

	keys := identity.FailureKeys()
	if len(keys) != 5 {
		t.Fatalf("FailureKeys length = %d, want 5: %#v", len(keys), keys)
	}

	joined := strings.Join(keys, "|")
	for _, raw := range []string{"session-secret", "visitor-secret", "203.0.113.9", "Mozilla/5.0 Card Test"} {
		if strings.Contains(joined, strings.ToLower(raw)) || strings.Contains(joined, raw) {
			t.Fatalf("FailureKeys leaked raw value %q in %#v", raw, keys)
		}
	}
	for _, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) != 3 {
			t.Fatalf("FailureKeys key %q should use provider:dimension:digest format", key)
		}
		if len(parts[2]) != 64 || parts[2] == "42" {
			t.Fatalf("FailureKeys key %q should store a sha256 digest suffix", key)
		}
	}
	for _, prefix := range []string{"stripe:user:", "stripe:session:", "stripe:anonymous:", "stripe:ip:", "stripe:ipua:"} {
		if !hasKeyPrefix(keys, prefix) {
			t.Fatalf("FailureKeys missing prefix %q in %#v", prefix, keys)
		}
	}
}

func TestEvaluateSignalsChallengesAfterFailureThreshold(t *testing.T) {
	service := &Service{
		config: config.PaymentRiskConfig{
			FailureThreshold:     3,
			DelaySeconds:         2,
			HighRiskScore:        60,
			FailureWindowSeconds: 600,
		},
	}

	decision := service.evaluateSignals(3, Signals{UserAgent: "Mozilla/5.0"})

	if decision.Action != DecisionActionChallenge {
		t.Fatalf("action = %q, want %q", decision.Action, DecisionActionChallenge)
	}
	if !decision.ChallengeRequired {
		t.Fatal("expected challenge to be required")
	}
	if decision.ChallengeReason != "repeated_payment_failures" {
		t.Fatalf("challenge reason = %q, want repeated_payment_failures", decision.ChallengeReason)
	}
	if decision.Delay != 2*time.Second {
		t.Fatalf("delay = %s, want 2s", decision.Delay)
	}
}

func TestPaymentIntentBindingKeyDoesNotExposeRawProviderID(t *testing.T) {
	key := paymentIntentBindingKey("pi_123_secret")
	if strings.Contains(key, "pi_123_secret") {
		t.Fatalf("paymentIntentBindingKey leaked raw payment intent id: %q", key)
	}
	if !strings.HasPrefix(key, "commerce_platform:payment-risk:payment-intent:") {
		t.Fatalf("paymentIntentBindingKey = %q, want payment-intent namespace", key)
	}
}

func hasKeyPrefix(keys []string, prefix string) bool {
	for _, key := range keys {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}
