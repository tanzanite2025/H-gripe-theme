package service

import (
	"encoding/json"
	"testing"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"

	"github.com/stretchr/testify/require"
)

func TestSiteQualityRuleIdentitySeparatesSystemAndProviderIDs(t *testing.T) {
	linkRule, ok := siteQualityLookupAuditRule(siteQualityLinkTextAuditID)
	require.True(t, ok)
	require.Equal(t, siteQualityLinkDescriptiveTextRuleID, siteQualityRuleIDForAudit(linkRule.ID))
	require.Equal(t, siteQualityLinkTextAuditID, siteQualityProviderAuditIDForAudit(linkRule.ID))

	opportunityRule, ok := siteQualityLookupAuditRule("render-blocking-resources")
	require.True(t, ok)
	require.Equal(t, opportunityRule.ID, siteQualityRuleIDForAudit(opportunityRule.ID))
	require.Equal(t, opportunityRule.ID, siteQualityProviderAuditIDForAudit(opportunityRule.ID))

	customRule, ok := siteQualityLookupAuditRule(siteQualityHeadingMissingH1AuditID)
	require.True(t, ok)
	require.Equal(t, customRule.ID, siteQualityRuleIDForAudit(customRule.ID))
	require.Empty(t, siteQualityProviderAuditIDForAudit(customRule.ID))

	require.Equal(t, siteQualityLinkDescriptiveTextRuleID, sitequalitydomain.SiteQualityRuleIDDescriptiveLinkText)
	require.Equal(t, siteQualityLinkTextAuditID, sitequalitydomain.SiteQualityProviderAuditIDLinkText)
}

func TestSiteQualityRunViewBackfillsRuleIdentityForHistoricalIssues(t *testing.T) {
	run := sitequalitydomain.SiteQualityRun{
		IssuesJSON: `[{"id":"link-text","kind":"links","title":"Links do not have descriptive text"}]`,
	}

	view := siteQualityRunView(run)
	require.Len(t, view.Issues, 1)
	require.Equal(t, siteQualityLinkDescriptiveTextRuleID, view.Issues[0].RuleID)
	require.Equal(t, siteQualityLinkTextAuditID, view.Issues[0].ProviderAuditID)

	encoded, err := json.Marshal(view.Issues[0])
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"rule_id":"link_descriptive_text"`)
}
