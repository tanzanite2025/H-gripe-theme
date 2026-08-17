package service

import "strings"

const siteQualityAuditRuleVersion = "2026-08-17"

type siteQualityAuditRule struct {
	ID                    string
	Kind                  string
	DefaultSeverity       string
	MinSavingsMS          float64
	MinSavingsBytes       int64
	FailWhenScoreBelowOne bool
}

// siteQualityActionableAuditRules is intentionally an allowlist. Lighthouse
// metrics describe symptoms, while these opportunities identify a remediable
// cause. New provider audit IDs remain raw evidence until reviewed here.
var siteQualityActionableAuditRules = map[string]siteQualityAuditRule{
	"render-blocking-resources":                           {ID: "render-blocking-resources", Kind: "opportunity", MinSavingsMS: 100},
	"uses-long-cache-ttl":                                 {ID: "uses-long-cache-ttl", Kind: "opportunity", MinSavingsBytes: 10 * 1024},
	"uses-optimized-images":                               {ID: "uses-optimized-images", Kind: "opportunity", MinSavingsBytes: 10 * 1024},
	"modern-image-formats":                                {ID: "modern-image-formats", Kind: "opportunity", MinSavingsBytes: 10 * 1024},
	"uses-text-compression":                               {ID: "uses-text-compression", Kind: "opportunity", MinSavingsBytes: 10 * 1024},
	"unused-css-rules":                                    {ID: "unused-css-rules", Kind: "opportunity", MinSavingsBytes: 10 * 1024},
	"unused-javascript":                                   {ID: "unused-javascript", Kind: "opportunity", MinSavingsBytes: 10 * 1024},
	"unminified-css":                                      {ID: "unminified-css", Kind: "opportunity", MinSavingsBytes: 10 * 1024},
	"unminified-javascript":                               {ID: "unminified-javascript", Kind: "opportunity", MinSavingsBytes: 10 * 1024},
	"efficient-animated-content":                          {ID: "efficient-animated-content", Kind: "opportunity", MinSavingsBytes: 10 * 1024},
	"server-response-time":                                {ID: "server-response-time", Kind: "opportunity", MinSavingsMS: 200},
	"redirects":                                           {ID: "redirects", Kind: "opportunity", MinSavingsMS: 100},
	"uses-rel-preconnect":                                 {ID: "uses-rel-preconnect", Kind: "opportunity", MinSavingsMS: 100},
	"font-display":                                        {ID: "font-display", Kind: "opportunity", MinSavingsMS: 100},
	"uses-rel-preload":                                    {ID: "uses-rel-preload", Kind: "opportunity", MinSavingsMS: 100},
	"heading-order":                                       {ID: "heading-order", Kind: "headings", DefaultSeverity: "high", FailWhenScoreBelowOne: true},
	siteQualityHeadingMissingH1AuditID:                    {ID: siteQualityHeadingMissingH1AuditID, Kind: "headings", DefaultSeverity: "critical", FailWhenScoreBelowOne: true},
	siteQualityHeadingMultipleH1AuditID:                   {ID: siteQualityHeadingMultipleH1AuditID, Kind: "headings", DefaultSeverity: "medium", FailWhenScoreBelowOne: true},
	siteQualityHeadingScanFailedAuditID:                   {ID: siteQualityHeadingScanFailedAuditID, Kind: "headings", DefaultSeverity: "critical", FailWhenScoreBelowOne: true},
	siteQualityStructuredDataScanFailedAuditID:            {ID: siteQualityStructuredDataScanFailedAuditID, Kind: "schema", DefaultSeverity: "critical", FailWhenScoreBelowOne: true},
	siteQualityStructuredDataInvalidJSONLDAuditID:         {ID: siteQualityStructuredDataInvalidJSONLDAuditID, Kind: "schema", DefaultSeverity: "critical", FailWhenScoreBelowOne: true},
	siteQualityStructuredDataMissingStructuredDataAuditID: {ID: siteQualityStructuredDataMissingStructuredDataAuditID, Kind: "schema", DefaultSeverity: "high", FailWhenScoreBelowOne: true},
	siteQualityStructuredDataMissingRequiredTypeAuditID:   {ID: siteQualityStructuredDataMissingRequiredTypeAuditID, Kind: "schema", DefaultSeverity: "high", FailWhenScoreBelowOne: true},
	siteQualityStructuredDataDuplicatePrimaryTypeAuditID:  {ID: siteQualityStructuredDataDuplicatePrimaryTypeAuditID, Kind: "schema", DefaultSeverity: "medium", FailWhenScoreBelowOne: true},
	siteQualityStructuredDataURLMismatchAuditID:           {ID: siteQualityStructuredDataURLMismatchAuditID, Kind: "schema", DefaultSeverity: "high", FailWhenScoreBelowOne: true},
	siteQualityStructuredDataBreadcrumbInvalidAuditID:     {ID: siteQualityStructuredDataBreadcrumbInvalidAuditID, Kind: "schema", DefaultSeverity: "high", FailWhenScoreBelowOne: true},
	siteQualityStructuredDataProductInvalidAuditID:        {ID: siteQualityStructuredDataProductInvalidAuditID, Kind: "schema", DefaultSeverity: "critical", FailWhenScoreBelowOne: true},
	siteQualityStructuredDataFAQInvalidAuditID:            {ID: siteQualityStructuredDataFAQInvalidAuditID, Kind: "schema", DefaultSeverity: "high", FailWhenScoreBelowOne: true},
	siteQualityStructuredDataFAQContentMismatchAuditID:    {ID: siteQualityStructuredDataFAQContentMismatchAuditID, Kind: "schema", DefaultSeverity: "high", FailWhenScoreBelowOne: true},
	siteQualityStructuredDataArticleInvalidAuditID:        {ID: siteQualityStructuredDataArticleInvalidAuditID, Kind: "schema", DefaultSeverity: "high", FailWhenScoreBelowOne: true},
	siteQualityStructuredDataOrganizationInvalidAuditID:   {ID: siteQualityStructuredDataOrganizationInvalidAuditID, Kind: "schema", DefaultSeverity: "high", FailWhenScoreBelowOne: true},
	siteQualityStructuredDataWebPageInvalidAuditID:        {ID: siteQualityStructuredDataWebPageInvalidAuditID, Kind: "schema", DefaultSeverity: "medium", FailWhenScoreBelowOne: true},
}

func siteQualityLookupAuditRule(auditID string) (siteQualityAuditRule, bool) {
	rule, ok := siteQualityActionableAuditRules[strings.TrimSpace(auditID)]
	return rule, ok
}

func siteQualityAuditMeetsRule(rule siteQualityAuditRule, audit siteQualityAPIAudit) bool {
	if audit.ScoreDisplayMode == "notApplicable" {
		return false
	}
	if rule.FailWhenScoreBelowOne && audit.Score != nil && *audit.Score < 1 {
		return true
	}
	if audit.Details.OverallSavingsMS != nil && *audit.Details.OverallSavingsMS >= rule.MinSavingsMS && rule.MinSavingsMS > 0 {
		return true
	}
	if audit.Details.OverallSavingsBytes != nil && *audit.Details.OverallSavingsBytes >= rule.MinSavingsBytes && rule.MinSavingsBytes > 0 {
		return true
	}
	return false
}
