package service

import (
	"testing"

	"commerce-platform/internal/domain/ops"
)

func TestRelevantDNSRecordsFiltersNonRoutingRecords(t *testing.T) {
	records := relevantDNSRecords([]cloudflareDNSRecord{
		{Type: "A", Content: "2.25.85.201"},
		{Type: "TXT", Content: "verification-token"},
		{Type: "cname", Content: "proxy.example.com"},
	})

	if len(records) != 2 {
		t.Fatalf("relevantDNSRecords() length = %d, want 2", len(records))
	}
	if records[1].Content != "proxy.example.com" {
		t.Fatalf("relevantDNSRecords() kept unexpected record: %#v", records[1])
	}
}

func TestObservedTargetSortsAndDeduplicatesRecords(t *testing.T) {
	target := observedTarget([]cloudflareDNSRecord{
		{Type: "A", Content: "2.25.85.201"},
		{Type: "CNAME", Content: "edge.example.com"},
		{Type: "AAAA", Content: "2001:db8::1"},
		{Type: "A", Content: "2.25.85.201"},
		{Type: "TXT", Content: "ignored"},
	})

	if target != "2.25.85.201, 2001:db8::1, edge.example.com" {
		t.Fatalf("observedTarget() = %q, want sorted unique targets", target)
	}
}

func TestObservedProxyModeIgnoresNonRoutingRecords(t *testing.T) {
	if got := observedProxyMode([]cloudflareDNSRecord{
		{Type: "A", Content: "2.25.85.201", Proxied: true},
		{Type: "TXT", Content: "ignored", Proxied: false},
	}); got != ops.DomainProxyProxied {
		t.Fatalf("observedProxyMode() = %q, want %q", got, ops.DomainProxyProxied)
	}

	if got := observedProxyMode([]cloudflareDNSRecord{
		{Type: "A", Content: "2.25.85.201", Proxied: true},
		{Type: "CNAME", Content: "edge.example.com", Proxied: false},
	}); got != ops.DomainProxyUnknown {
		t.Fatalf("observedProxyMode() = %q, want %q", got, ops.DomainProxyUnknown)
	}
}

func TestNormalizeObservedTLSMode(t *testing.T) {
	tests := map[string]string{
		"strict":   ops.DomainTLSFullStrict,
		" full ":   ops.DomainTLSFull,
		"Flexible": ops.DomainTLSFlexible,
		"off":      ops.DomainTLSOff,
		"unknown":  ops.DomainTLSUnknown,
		"":         ops.DomainTLSUnknown,
	}

	for input, want := range tests {
		if got := normalizeObservedTLSMode(input); got != want {
			t.Errorf("normalizeObservedTLSMode(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCompareDomainObservedState(t *testing.T) {
	baseDomain := ops.DomainBinding{
		Target:    "2.25.85.201",
		ProxyMode: ops.DomainProxyProxied,
		TLSMode:   ops.DomainTLSFullStrict,
	}
	baseResult := ops.DomainSyncResult{
		DNSRecordCount:    1,
		ObservedTarget:    "2.25.85.201",
		ObservedProxyMode: ops.DomainProxyProxied,
		ObservedTLSMode:   ops.DomainTLSFullStrict,
	}

	status, message := compareDomainObservedState(baseDomain, baseResult)
	if status != ops.DomainObservedMatched || message != "" {
		t.Fatalf("matching state = (%q, %q), want matched with no message", status, message)
	}

	baseResult.ObservedProxyMode = ops.DomainProxyDNSOnly
	status, message = compareDomainObservedState(baseDomain, baseResult)
	if status != ops.DomainObservedDrifted || message == "" {
		t.Fatalf("proxy drift = (%q, %q), want drifted with message", status, message)
	}

	baseResult.ObservedProxyMode = ops.DomainProxyProxied
	baseResult.ObservedTLSMode = ops.DomainTLSFull
	status, message = compareDomainObservedState(baseDomain, baseResult)
	if status != ops.DomainObservedDrifted || message == "" {
		t.Fatalf("TLS drift = (%q, %q), want drifted with message", status, message)
	}

	baseResult.ObservedTLSMode = ops.DomainTLSFullStrict
	baseResult.ObservedTarget = "198.51.100.10"
	status, message = compareDomainObservedState(baseDomain, baseResult)
	if status != ops.DomainObservedDrifted || message == "" {
		t.Fatalf("target drift = (%q, %q), want drifted with message", status, message)
	}
}
