package service

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSiteQualityRenderedLinkAuditIssuesCountsAllBrokenLinksBeforeEvidenceLimit(t *testing.T) {
	audit := &siteQualityRenderedLinkAudit{
		Status:     "complete",
		Configured: true,
		Links:      make([]siteQualityRenderedLink, 0, 60),
	}
	for i := 0; i < 60; i++ {
		audit.Links = append(audit.Links, siteQualityRenderedLink{
			Href:       fmt.Sprintf("https://example.com/server-error-%02d", i),
			StatusCode: http.StatusInternalServerError,
			OK:         false,
		})
	}

	issues := siteQualityRenderedLinkAuditIssues(audit)

	require.Len(t, issues, 1)
	issue := issues[0]
	require.NotNil(t, issue.NumericValue)
	require.Equal(t, 60.0, *issue.NumericValue)
	require.Equal(t, "60 broken link(s)", issue.DisplayValue)
	require.Len(t, issue.Links, 12)
	require.Len(t, issue.Runtime, 12)
}

func TestSiteQualityRenderedLinkAuditIssuesRatesCriticalBeforeEvidenceLimit(t *testing.T) {
	audit := &siteQualityRenderedLinkAudit{
		Status:     "complete",
		Configured: true,
		Links:      make([]siteQualityRenderedLink, 0, 16),
	}
	for i := 0; i < 15; i++ {
		audit.Links = append(audit.Links, siteQualityRenderedLink{
			Href:       fmt.Sprintf("https://example.com/server-error-%02d", i),
			StatusCode: http.StatusInternalServerError,
			OK:         false,
		})
	}
	audit.Links = append(audit.Links, siteQualityRenderedLink{
		Href:       "https://example.com/not-found",
		StatusCode: http.StatusNotFound,
		OK:         false,
	})

	issues := siteQualityRenderedLinkAuditIssues(audit)

	require.Len(t, issues, 1)
	issue := issues[0]
	require.Equal(t, "critical", issue.Severity)
	require.Equal(t, "16 broken link(s)", issue.DisplayValue)
	require.Len(t, issue.Runtime, 12)
	for _, evidence := range issue.Runtime {
		require.NotEqual(t, http.StatusNotFound, evidence.StatusCode)
	}
}

func TestSiteQualityInteractionAuditIssuesReportsFailedProbeSeparately(t *testing.T) {
	threshold := 200.0
	audit := &siteQualityInteractionAudit{
		Status:     "complete",
		Configured: true,
		Interactions: []siteQualityInteractionProbe{
			{
				Name:                  "Buy button",
				Selector:              "#buy",
				Action:                "click",
				Status:                "failed",
				ThresholdMilliseconds: &threshold,
				Exceeded:              true,
				Error:                 "interaction selector was not found",
			},
		},
	}

	issues := siteQualityInteractionAuditIssues(audit)

	require.Len(t, issues, 1)
	issue := issues[0]
	require.Equal(t, siteQualityInteractionAuditFailedID, issue.ID)
	require.Equal(t, "Configured user interaction probe failed", issue.Title)
	require.Equal(t, "1 interaction probe(s) failed", issue.DisplayValue)
	require.NotNil(t, issue.NumericValue)
	require.Equal(t, 1.0, *issue.NumericValue)
	require.Nil(t, issue.SavingsMS)
	require.Len(t, issue.Runtime, 1)
	require.Equal(t, "failed", issue.Runtime[0].Status)
	require.Nil(t, issue.Runtime[0].ResponseMS)
	require.Contains(t, issue.Runtime[0].Error, "selector")
}

func TestSiteQualityInteractionAuditIssuesSplitsFailuresFromSlowInteractions(t *testing.T) {
	threshold := 200.0
	response := 360.0
	audit := &siteQualityInteractionAudit{
		Status:     "complete",
		Configured: true,
		Interactions: []siteQualityInteractionProbe{
			{
				Name:                  "Missing button",
				Selector:              "#missing",
				Action:                "click",
				Status:                "failed",
				ThresholdMilliseconds: &threshold,
				Error:                 "interaction selector was not found",
			},
			{
				Name:                  "Buy button",
				Selector:              "#buy",
				Action:                "click",
				Status:                "complete",
				ResponseMilliseconds:  &response,
				ThresholdMilliseconds: &threshold,
				Exceeded:              true,
			},
		},
	}

	issues := siteQualityInteractionAuditIssues(audit)
	byID := siteQualityIssuesByID(issues)

	require.Len(t, issues, 2)
	require.Contains(t, byID, siteQualityInteractionAuditFailedID)
	require.Contains(t, byID, siteQualityInteractionLatencyAuditID)
	require.Len(t, byID[siteQualityInteractionAuditFailedID].Runtime, 1)
	require.Equal(t, "#missing", byID[siteQualityInteractionAuditFailedID].Runtime[0].Selector)
	latencyIssue := byID[siteQualityInteractionLatencyAuditID]
	require.Equal(t, "Slowest interaction 360 ms", latencyIssue.DisplayValue)
	require.Len(t, latencyIssue.Runtime, 1)
	require.Equal(t, "complete", latencyIssue.Runtime[0].Status)
	require.Equal(t, 360.0, *latencyIssue.Runtime[0].ResponseMS)
}

func TestSiteQualitySoftNavigationAuditIssuesReportsFailedProbeSeparately(t *testing.T) {
	duration := 0.0
	threshold := 2000.0
	audit := &siteQualitySoftNavigationAudit{
		Status:     "complete",
		Configured: true,
		Navigations: []siteQualitySoftNavigationResult{
			{
				Selector:              "nav a",
				Text:                  "Products",
				Status:                "failed",
				DurationMilliseconds:  &duration,
				ThresholdMilliseconds: &threshold,
				Exceeded:              true,
				Error:                 "route did not change after clicking navigation target",
			},
		},
	}

	issues := siteQualitySoftNavigationAuditIssues(audit)

	require.Len(t, issues, 1)
	issue := issues[0]
	require.Equal(t, siteQualitySoftNavigationAuditFailedID, issue.ID)
	require.Equal(t, "Configured soft navigation probe failed", issue.Title)
	require.Equal(t, "1 soft navigation probe(s) failed", issue.DisplayValue)
	require.NotNil(t, issue.NumericValue)
	require.Equal(t, 1.0, *issue.NumericValue)
	require.Nil(t, issue.SavingsMS)
	require.Len(t, issue.Runtime, 1)
	require.Equal(t, "failed", issue.Runtime[0].Status)
	require.Equal(t, "route did not change after clicking navigation target", issue.Runtime[0].Error)
}

func TestSiteQualitySoftNavigationAuditIssuesSplitsFailuresFromRegressions(t *testing.T) {
	duration := 920.0
	threshold := 500.0
	audit := &siteQualitySoftNavigationAudit{
		Status:     "complete",
		Configured: true,
		Navigations: []siteQualitySoftNavigationResult{
			{
				Selector:              "nav a.missing",
				Text:                  "Missing",
				Status:                "failed",
				ThresholdMilliseconds: &threshold,
				Error:                 "navigation target was detached",
			},
			{
				Selector:              "nav a.products",
				Text:                  "Products",
				Status:                "complete",
				Mode:                  "hard-navigation",
				DurationMilliseconds:  &duration,
				ThresholdMilliseconds: &threshold,
				Exceeded:              true,
			},
		},
	}

	issues := siteQualitySoftNavigationAuditIssues(audit)
	byID := siteQualityIssuesByID(issues)

	require.Len(t, issues, 2)
	require.Contains(t, byID, siteQualitySoftNavigationAuditFailedID)
	require.Contains(t, byID, siteQualitySoftNavigationRegressionID)
	require.Len(t, byID[siteQualitySoftNavigationAuditFailedID].Runtime, 1)
	require.Equal(t, "failed", byID[siteQualitySoftNavigationAuditFailedID].Runtime[0].Status)
	regressionIssue := byID[siteQualitySoftNavigationRegressionID]
	require.Equal(t, "Slowest route transition 920 ms", regressionIssue.DisplayValue)
	require.Len(t, regressionIssue.Runtime, 1)
	require.Equal(t, "complete", regressionIssue.Runtime[0].Status)
	require.Equal(t, "hard-navigation", regressionIssue.Runtime[0].Mode)
}

func siteQualityIssuesByID(issues []LighthouseRunnerIssue) map[string]LighthouseRunnerIssue {
	byID := make(map[string]LighthouseRunnerIssue, len(issues))
	for _, issue := range issues {
		byID[issue.ID] = issue
	}
	return byID
}

func TestSiteQualityRenderedLinkAuditIssuesIgnoreDeadlineSkippedLinks(t *testing.T) {
	audit := &siteQualityRenderedLinkAudit{
		Status:     "timeout_skipped",
		Configured: true,
		Links: []siteQualityRenderedLink{
			{
				Href:           "https://example.com/slow",
				StatusCode:     0,
				OK:             false,
				TimeoutSkipped: true,
				Error:          "link probe stopped because the rendered audit deadline was reached",
			},
			{
				Href:       "https://example.com/missing",
				Text:       "Missing",
				StatusCode: http.StatusNotFound,
				OK:         false,
				Error:      "not found",
			},
		},
	}

	issues := siteQualityRenderedLinkAuditIssues(audit)

	require.Len(t, issues, 1)
	require.Equal(t, siteQualityBrokenLinkAuditID, issues[0].ID)
	require.Len(t, issues[0].Links, 1)
	require.Equal(t, "https://example.com/missing", issues[0].Links[0].Href)
}

func TestSiteQualityRuntimeAuditIssuesDoNotTreatDeadlineSkipAsScanFailure(t *testing.T) {
	tests := []struct {
		name   string
		issues []LighthouseRunnerIssue
	}{
		{
			name: "links",
			issues: siteQualityRenderedLinkAuditIssues(&siteQualityRenderedLinkAudit{
				Status:     "timeout_skipped",
				Configured: true,
				Links: []siteQualityRenderedLink{
					{
						Href:           "https://example.com/slow",
						StatusCode:     0,
						OK:             false,
						TimeoutSkipped: true,
						Error:          "deadline reached",
					},
				},
			}),
		},
		{
			name: "interactions",
			issues: siteQualityInteractionAuditIssues(&siteQualityInteractionAudit{
				Status:     "timeout_skipped",
				Configured: true,
				Interactions: []siteQualityInteractionProbe{
					{Selector: "#buy", Status: "timeout_skipped", Error: "deadline reached"},
				},
			}),
		},
		{
			name: "soft navigations",
			issues: siteQualitySoftNavigationAuditIssues(&siteQualitySoftNavigationAudit{
				Status:     "timeout_skipped",
				Configured: true,
				Navigations: []siteQualitySoftNavigationResult{
					{Selector: "a#products", Status: "timeout_skipped", Error: "deadline reached"},
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, issue := range tt.issues {
				require.NotContains(t, []string{
					siteQualityLinkAuditFailedAuditID,
					siteQualityInteractionAuditFailedID,
					siteQualitySoftNavigationAuditFailedID,
				}, issue.ID)
			}
			require.Empty(t, tt.issues)
		})
	}
}
