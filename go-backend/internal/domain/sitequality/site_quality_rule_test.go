package sitequality

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeRuleIdentityPrefersKnownLinkTextIdentity(t *testing.T) {
	ruleID, providerAuditID := NormalizeRuleIdentity(
		"legacy-link-rule",
		SiteQualityProviderAuditIDLinkText,
		"link_descriptive_text",
	)

	require.Equal(t, SiteQualityRuleIDDescriptiveLinkText, ruleID)
	require.Equal(t, SiteQualityProviderAuditIDLinkText, providerAuditID)
}
