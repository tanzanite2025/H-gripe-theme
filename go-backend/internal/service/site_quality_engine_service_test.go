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

func TestNewSiteQualityEngineServiceDefaultsToSingleRunnerBatch(t *testing.T) {
	t.Parallel()

	engine := NewSiteQualityEngineService(nil, nil, nil, nil, nil, nil, SiteQualityEngineConfig{})

	require.Equal(t, 1, engine.cfg.WorkerBatchLimit)
	require.Equal(t, 1, engine.cfg.ProviderConcurrency)
	require.False(t, engine.cfg.AutoScanEnabled)
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

func TestSiteQualityScopedEvaluationOnlyCleansSelectedAuditGroup(t *testing.T) {
	target := sitequalitydomain.SiteQualityTarget{ID: 1, CanonicalURL: "https://example.com/"}
	job := sitequalitydomain.SiteQualityJob{
		Strategy:              sitequalitydomain.SiteQualityStrategyMobile,
		AuditScope:            sitequalitydomain.SiteQualityAuditScopeHeadings,
		SampleCount:           1,
		RequiredConfirmations: 1,
	}
	decision, detections := evaluateSiteQualityRuns(target, job, []LighthouseRunnerRunView{{
		ID: 12,
		Issues: []LighthouseRunnerIssue{
			{ID: siteQualityHeadingMissingH1AuditID, Kind: "headings", Title: "Missing H1"},
			{ID: siteQualityStructuredDataMissingStructuredDataAuditID, Kind: "schema", Title: "Missing schema"},
			{ID: siteQualityLinkTextAuditID, Kind: "links", Title: "Link text"},
		},
	}})

	require.Len(t, detections, 1)
	require.Equal(t, siteQualityHeadingMissingH1AuditID, detections[0].AuditID)
	require.NotContains(t, decision.Clean, siteQualityStructuredDataMissingStructuredDataAuditID)
	require.NotContains(t, decision.Clean, siteQualityLinkTextAuditID)
	require.Contains(t, decision.Clean, siteQualityHeadingMultipleH1AuditID)
}
