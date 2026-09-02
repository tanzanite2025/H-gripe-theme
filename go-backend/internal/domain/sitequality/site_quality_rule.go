package sitequality

import "strings"

const (
	SiteQualityRuleIDDescriptiveLinkText = "link_descriptive_text"
	SiteQualityProviderAuditIDLinkText   = "link-text"
)

// NormalizeRuleIdentity returns the stable internal rule ID and external
// provider audit ID for legacy or current representations of one rule.
func NormalizeRuleIdentity(ruleID string, auditID string, providerAuditID string) (string, string) {
	normalizedRuleID := RuleIDForAuditID(ruleID)
	normalizedAuditID := strings.TrimSpace(auditID)
	normalizedProviderAuditID := ProviderAuditIDForAuditID(providerAuditID)
	if normalizedAuditID == SiteQualityProviderAuditIDLinkText ||
		normalizedAuditID == "descriptive_link_text" ||
		normalizedProviderAuditID == SiteQualityProviderAuditIDLinkText {
		normalizedRuleID = SiteQualityRuleIDDescriptiveLinkText
	}
	if normalizedRuleID == "" {
		normalizedRuleID = RuleIDForAuditID(auditID)
	}
	if normalizedRuleID == "" {
		normalizedRuleID = RuleIDForAuditID(providerAuditID)
	}

	if normalizedAuditID != "" {
		providerFromAudit := ProviderAuditIDForAuditID(normalizedAuditID)
		if providerFromAudit != "" {
			normalizedProviderAuditID = providerFromAudit
		}
	}
	if normalizedProviderAuditID == "" {
		normalizedProviderAuditID = ProviderAuditIDForAuditID(ruleID)
	}
	return normalizedRuleID, normalizedProviderAuditID
}

// RuleIDForAuditID converts a provider audit ID into the stable system rule
// identity. Unknown IDs are intentionally preserved so synthetic checks and
// future provider rules remain forward-compatible.
func RuleIDForAuditID(value string) string {
	normalized := strings.TrimSpace(value)
	switch normalized {
	case SiteQualityProviderAuditIDLinkText, "descriptive_link_text":
		return SiteQualityRuleIDDescriptiveLinkText
	default:
		return normalized
	}
}

// ProviderAuditIDForAuditID returns the provider identity when the audit is
// backed by a known external rule. Synthetic checks intentionally return
// empty because they have no provider audit ID.
func ProviderAuditIDForAuditID(value string) string {
	normalized := strings.TrimSpace(value)
	switch normalized {
	case SiteQualityProviderAuditIDLinkText:
		return SiteQualityProviderAuditIDLinkText
	case SiteQualityRuleIDDescriptiveLinkText, "descriptive_link_text":
		return SiteQualityProviderAuditIDLinkText
	default:
		if strings.HasPrefix(normalized, "site-heading-") ||
			strings.HasPrefix(normalized, "site-schema-") ||
			strings.HasPrefix(normalized, "site-resource-") ||
			strings.HasPrefix(normalized, "site-link-") ||
			strings.HasPrefix(normalized, "site-interaction-") ||
			strings.HasPrefix(normalized, "site-soft-navigation-") {
			return ""
		}
		return normalized
	}
}
