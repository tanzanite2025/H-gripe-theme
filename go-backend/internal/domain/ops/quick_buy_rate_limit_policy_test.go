package ops

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeQuickBuyRateLimitPolicyJSONUsesDefaults(t *testing.T) {
	encoded, policy, err := NormalizeQuickBuyRateLimitPolicyJSON("")
	if err != nil {
		t.Fatalf("NormalizeQuickBuyRateLimitPolicyJSON() error = %v", err)
	}
	if !policy.Enabled || policy.IPRequestsPerMinute != 120 || policy.SessionBurst != 20 {
		t.Fatalf("policy = %#v, want defaults", policy)
	}
	if !json.Valid([]byte(encoded)) {
		t.Fatalf("encoded policy is not valid JSON: %s", encoded)
	}
}

func TestNormalizeQuickBuyRateLimitPolicyJSONRejectsInvalidEnabledLimits(t *testing.T) {
	_, _, err := NormalizeQuickBuyRateLimitPolicyJSON(`{
		"enabled": true,
		"ip_requests_per_minute": 120,
		"ip_burst": 0,
		"session_requests_per_minute": 60,
		"session_burst": 20
	}`)
	if err == nil {
		t.Fatal("NormalizeQuickBuyRateLimitPolicyJSON should reject invalid limits")
	}
	if !strings.Contains(err.Error(), "positive") {
		t.Fatalf("error = %v, want positive-limit diagnostic", err)
	}
}

func TestNormalizeQuickBuyRateLimitPolicyJSONAllowsDisabledPolicy(t *testing.T) {
	_, policy, err := NormalizeQuickBuyRateLimitPolicyJSON(`{"enabled": false}`)
	if err != nil {
		t.Fatalf("NormalizeQuickBuyRateLimitPolicyJSON() error = %v", err)
	}
	if policy.Enabled {
		t.Fatalf("policy = %#v, want disabled", policy)
	}
}
