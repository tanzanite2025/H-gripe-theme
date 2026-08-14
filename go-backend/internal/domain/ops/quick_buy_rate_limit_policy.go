package ops

import (
	"encoding/json"
	"fmt"
	"strings"
)

type QuickBuyRateLimitPolicy struct {
	Enabled                  bool `json:"enabled"`
	IPRequestsPerMinute      int  `json:"ip_requests_per_minute"`
	IPBurst                  int  `json:"ip_burst"`
	SessionRequestsPerMinute int  `json:"session_requests_per_minute"`
	SessionBurst             int  `json:"session_burst"`
	EdgeIPRequestsPerMinute  int  `json:"edge_ip_requests_per_minute"`
	EdgeIPBurst              int  `json:"edge_ip_burst"`
	CaddyRateLimitEnabled    bool `json:"caddy_rate_limit_enabled"`
}

func DefaultQuickBuyRateLimitPolicy() QuickBuyRateLimitPolicy {
	return QuickBuyRateLimitPolicy{
		Enabled:                  true,
		IPRequestsPerMinute:      120,
		IPBurst:                  40,
		SessionRequestsPerMinute: 60,
		SessionBurst:             20,
		EdgeIPRequestsPerMinute:  240,
		EdgeIPBurst:              80,
		CaddyRateLimitEnabled:    false,
	}
}

func NormalizeQuickBuyRateLimitPolicyJSON(raw string) (string, QuickBuyRateLimitPolicy, error) {
	policy, err := ParseQuickBuyRateLimitPolicy(raw)
	if err != nil {
		return "", QuickBuyRateLimitPolicy{}, err
	}
	encoded, err := json.Marshal(policy)
	if err != nil {
		return "", QuickBuyRateLimitPolicy{}, err
	}
	return string(encoded), policy, nil
}

func ParseQuickBuyRateLimitPolicy(raw string) (QuickBuyRateLimitPolicy, error) {
	policy := DefaultQuickBuyRateLimitPolicy()
	raw = strings.TrimSpace(raw)
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &policy); err != nil {
			return QuickBuyRateLimitPolicy{}, fmt.Errorf("quick buy rate limit policy must be valid JSON: %w", err)
		}
	}
	if err := ValidateQuickBuyRateLimitPolicy(policy); err != nil {
		return QuickBuyRateLimitPolicy{}, err
	}
	return policy, nil
}

func ValidateQuickBuyRateLimitPolicy(policy QuickBuyRateLimitPolicy) error {
	if !policy.Enabled {
		return nil
	}
	if policy.IPRequestsPerMinute <= 0 ||
		policy.IPBurst <= 0 ||
		policy.SessionRequestsPerMinute <= 0 ||
		policy.SessionBurst <= 0 {
		return fmt.Errorf("quick buy rate limit policy requires positive backend IP and session limits")
	}
	if policy.CaddyRateLimitEnabled {
		if policy.EdgeIPRequestsPerMinute <= 0 || policy.EdgeIPBurst <= 0 {
			return fmt.Errorf("quick buy Caddy rate limit policy requires positive edge IP limits")
		}
	}
	return nil
}
