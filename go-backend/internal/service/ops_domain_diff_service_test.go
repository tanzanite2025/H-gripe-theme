package service

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/ops"
)

func TestBuildOpsDomainDiffReportsMatchedState(t *testing.T) {
	observedAt := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	diff := BuildOpsDomainDiff(ops.DomainBinding{
		ID:             9,
		Domain:         "learn.gripe",
		Environment:    ops.DomainEnvironmentProduction,
		Zone:           "learn.gripe",
		Target:         "2.25.85.201",
		ProxyMode:      ops.DomainProxyProxied,
		TLSMode:        ops.DomainTLSFullStrict,
		ObservedStatus: ops.DomainObservedMatched,
		ObservedTarget: "2.25.85.201",
		ObservedProxy:  ops.DomainProxyProxied,
		ObservedTLS:    ops.DomainTLSFullStrict,
		ObservedSource: "cloudflare:prod",
		LastObservedAt: &observedAt,
	})

	if diff.Status != ops.DomainDiffStatusMatched {
		t.Fatalf("diff status = %q, want matched", diff.Status)
	}
	if diff.ObservedSource != "cloudflare:prod" || diff.LastObservedAt == nil {
		t.Fatalf("diff observed metadata = (%q, %#v), want source and timestamp", diff.ObservedSource, diff.LastObservedAt)
	}
	for _, item := range diff.Items {
		if item.Status != ops.DomainDiffStatusMatched {
			t.Fatalf("diff item %s status = %q, want matched", item.Key, item.Status)
		}
	}
}

func TestBuildOpsDomainDiffAcceptsExpectedTargetAmongMultipleObservedRecords(t *testing.T) {
	diff := BuildOpsDomainDiff(ops.DomainBinding{
		Domain:         "learn.gripe",
		Environment:    ops.DomainEnvironmentProduction,
		Zone:           "learn.gripe",
		Target:         "2.25.85.201",
		ProxyMode:      ops.DomainProxyDNSOnly,
		TLSMode:        ops.DomainTLSFull,
		ObservedStatus: ops.DomainObservedDrifted,
		ObservedTarget: "198.51.100.1, 2.25.85.201",
		ObservedProxy:  ops.DomainProxyProxied,
		ObservedTLS:    ops.DomainTLSFull,
	})

	target := findDomainDiffItem(diff.Items, "target")
	if target == nil {
		t.Fatal("target diff item is missing")
	}
	if target.Status != ops.DomainDiffStatusMatched {
		t.Fatalf("target diff status = %q, want matched", target.Status)
	}
	proxy := findDomainDiffItem(diff.Items, "proxy_mode")
	if proxy == nil || proxy.Status != ops.DomainDiffStatusDrifted {
		t.Fatalf("proxy diff item = %#v, want drifted", proxy)
	}
}

func TestBuildOpsDomainDiffReportsUnknownWhenNotSynced(t *testing.T) {
	diff := BuildOpsDomainDiff(ops.DomainBinding{
		Domain:         "admin.learn.gripe",
		Environment:    ops.DomainEnvironmentProduction,
		Zone:           "learn.gripe",
		Target:         "2.25.85.201",
		ProxyMode:      ops.DomainProxyProxied,
		TLSMode:        ops.DomainTLSFullStrict,
		ObservedStatus: ops.DomainObservedUnknown,
	})

	if diff.Status != ops.DomainDiffStatusUnknown {
		t.Fatalf("diff status = %q, want unknown", diff.Status)
	}
	for _, item := range diff.Items {
		if item.Status != ops.DomainDiffStatusUnknown && item.Key != "zone" {
			t.Fatalf("diff item %s status = %q, want unknown", item.Key, item.Status)
		}
	}
}

func findDomainDiffItem(items []ops.DomainDiffItem, key string) *ops.DomainDiffItem {
	for index := range items {
		if items[index].Key == key {
			return &items[index]
		}
	}
	return nil
}
