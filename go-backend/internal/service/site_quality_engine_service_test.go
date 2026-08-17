package service

import (
	"testing"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"
	"github.com/stretchr/testify/require"
)

func TestSiteQualityOperationalStatusRequiresManualJobWorker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		summary *SiteQualityOperationalSummary
		want    string
	}{
		{
			name: "runner is not configured",
			summary: &SiteQualityOperationalSummary{
				RunnerConfigured: false,
				WorkerEnabled:    false,
			},
			want: "not_configured",
		},
		{
			name: "runner is ready but job worker is disabled",
			summary: &SiteQualityOperationalSummary{
				RunnerConfigured: true,
				WorkerEnabled:    false,
			},
			want: "degraded",
		},
		{
			name: "runner and job worker are ready",
			summary: &SiteQualityOperationalSummary{
				RunnerConfigured: true,
				WorkerEnabled:    true,
			},
			want: "healthy",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, test.want, siteQualityOperationalStatus(test.summary, nil))
		})
	}
}

func TestSiteQualityRecheckDecisionKeepsOnlyBoundFinding(t *testing.T) {
	decision, detections := restrictSiteQualityDecisionToFinding(
		siteQualityEvaluationDecision{
			Confirmed: []siteQualityDecision{
				{AuditID: "audit-one"},
				{AuditID: "audit-two"},
			},
			Clean:    []string{"audit-three", "audit-four"},
			Observed: []string{"audit-one", "audit-two"},
			Runs:     []uint{10, 11, 12},
		},
		[]sitequalitydomain.SiteQualityFindingDetection{
			{AuditID: "audit-one"},
			{AuditID: "audit-two"},
		},
		"audit-one",
	)

	require.Len(t, decision.Confirmed, 1)
	require.Equal(t, "audit-one", decision.Confirmed[0].AuditID)
	require.Empty(t, decision.Clean)
	require.Equal(t, []string{"audit-one"}, decision.Observed)
	require.Len(t, detections, 1)
	require.Equal(t, "audit-one", detections[0].AuditID)
}
