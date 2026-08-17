package service

import (
	"testing"

	seodomain "commerce-platform/internal/domain/seo"
	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"
)

func TestDeriveStorefrontURLIssueDefinitions(t *testing.T) {
	tests := []struct {
		name     string
		entry    seodomain.StorefrontRouteCatalogEntry
		expected string
	}{
		{
			name: "duplicate route becomes collision",
			entry: seodomain.StorefrontRouteCatalogEntry{
				EntryStatus: seodomain.RouteEntryStatusDuplicate,
			},
			expected: urlmanagementdomain.URLIssueTypePathCollision,
		},
		{
			name: "stale route is tracked",
			entry: seodomain.StorefrontRouteCatalogEntry{
				EntryStatus: seodomain.RouteEntryStatusStale,
			},
			expected: urlmanagementdomain.URLIssueTypeStaleRoute,
		},
		{
			name: "alias redirect chain remains actionable",
			entry: seodomain.StorefrontRouteCatalogEntry{
				IsAlias:         true,
				LastCheckStatus: seodomain.RouteCheckStatusRedirectChain,
			},
			expected: urlmanagementdomain.URLIssueTypeRedirectChain,
		},
		{
			name: "canonical route redirect is a status mismatch",
			entry: seodomain.StorefrontRouteCatalogEntry{
				LastCheckStatus: seodomain.RouteCheckStatusRedirect,
			},
			expected: urlmanagementdomain.URLIssueTypeRedirectStatusMismatch,
		},
		{
			name: "healthy alias redirect is not an issue",
			entry: seodomain.StorefrontRouteCatalogEntry{
				IsAlias:         true,
				LastCheckStatus: seodomain.RouteCheckStatusRedirect,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			definitions := deriveStorefrontURLIssueDefinitions(tt.entry)
			if tt.expected == "" {
				if len(definitions) != 0 {
					t.Fatalf("expected no issue definitions, got %#v", definitions)
				}
				return
			}
			for _, definition := range definitions {
				if definition.issueType == tt.expected {
					return
				}
			}
			t.Fatalf("expected issue %q, got %#v", tt.expected, definitions)
		})
	}
}
