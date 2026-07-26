package antifraud

import (
	"testing"
	"time"

	"tanzanite/internal/pkg/config"
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
