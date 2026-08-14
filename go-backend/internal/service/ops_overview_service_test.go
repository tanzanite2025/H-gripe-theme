package service

import (
	"testing"

	"commerce-platform/internal/domain/ops"
)

func TestSummarizeDomainsKeepsUnknownSeparateFromHealthy(t *testing.T) {
	summary := summarizeDomains([]ops.DomainBinding{
		{Enabled: true, Status: ops.DomainStatusActive, ObservedStatus: ops.DomainObservedUnknown},
		{Enabled: true, Status: ops.DomainStatusActive, ObservedStatus: ops.DomainObservedMatched},
		{Enabled: false, Status: ops.DomainStatusDrifted, ObservedStatus: ops.DomainObservedDrifted},
	})

	if summary.Total != 3 {
		t.Fatalf("Total = %d, want 3", summary.Total)
	}
	if summary.Enabled != 2 {
		t.Fatalf("Enabled = %d, want 2", summary.Enabled)
	}
	if summary.Unknown != 1 {
		t.Fatalf("Unknown = %d, want 1", summary.Unknown)
	}
	if summary.Healthy != 1 {
		t.Fatalf("Healthy = %d, want 1", summary.Healthy)
	}
	if summary.Attention != 2 {
		t.Fatalf("Attention = %d, want 2", summary.Attention)
	}
}

func TestIsDomainAttentionTreatsUnknownAsAttention(t *testing.T) {
	if !isDomainAttention(ops.DomainBinding{
		Status:         ops.DomainStatusActive,
		ObservedStatus: ops.DomainObservedUnknown,
	}) {
		t.Fatal("unknown observed status should require attention")
	}
	if isDomainAttention(ops.DomainBinding{
		Status:         ops.DomainStatusActive,
		ObservedStatus: ops.DomainObservedMatched,
	}) {
		t.Fatal("matched active domain should not require attention")
	}
}
