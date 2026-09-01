package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	seodomain "commerce-platform/internal/domain/seo"
	sitequalitydomain "commerce-platform/internal/domain/sitequality"
)

const (
	siteQualityStructuredDataScanFailedAuditID            = "site-schema-rendered-scan-failed"
	siteQualityStructuredDataInvalidJSONLDAuditID         = "site-schema-invalid-json-ld"
	siteQualityStructuredDataMissingStructuredDataAuditID = "site-schema-missing-structured-data"
	siteQualityStructuredDataMissingRequiredTypeAuditID   = "site-schema-missing-required-type"
	siteQualityStructuredDataDuplicatePrimaryTypeAuditID  = "site-schema-duplicate-primary-type"
	siteQualityStructuredDataURLMismatchAuditID           = "site-schema-url-mismatch"
	siteQualityStructuredDataBreadcrumbInvalidAuditID     = "site-schema-breadcrumb-invalid"
	siteQualityStructuredDataProductInvalidAuditID        = "site-schema-product-invalid"
	siteQualityStructuredDataFAQInvalidAuditID            = "site-schema-faq-invalid"
	siteQualityStructuredDataFAQContentMismatchAuditID    = "site-schema-faq-content-mismatch"
	siteQualityStructuredDataArticleInvalidAuditID        = "site-schema-article-invalid"
	siteQualityStructuredDataOrganizationInvalidAuditID   = "site-schema-organization-invalid"
	siteQualityStructuredDataWebPageInvalidAuditID        = "site-schema-webpage-invalid"
)

type siteQualityRenderedStructuredDataAudit struct {
	Status    string                             `json:"status"`
	Source    string                             `json:"source"`
	FinalURL  string                             `json:"finalUrl"`
	Error     string                             `json:"error"`
	Page      siteQualityStructuredDataPage      `json:"page"`
	JSONLD    []siteQualityStructuredDataScript  `json:"jsonLd"`
	Microdata []siteQualityStructuredDataSurface `json:"microdata"`
	RDFa      []siteQualityStructuredDataSurface `json:"rdfa"`
}

type siteQualityStructuredDataPage struct {
	Title            string `json:"title"`
	CanonicalURL     string `json:"canonicalUrl"`
	FAQQuestionCount int    `json:"faqQuestionCount"`
	ProductSignal    bool   `json:"productSignal"`
}

type siteQualityStructuredDataScript struct {
	Index      int                             `json:"index"`
	Selector   string                          `json:"selector"`
	Raw        string                          `json:"raw"`
	ParseError string                          `json:"parseError"`
	Nodes      []siteQualityStructuredDataNode `json:"nodes"`
}

type siteQualityStructuredDataNode struct {
	Types     []string        `json:"types"`
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	URL       string          `json:"url"`
	Selector  string          `json:"selector"`
	GraphPath string          `json:"graphPath"`
	Data      json.RawMessage `json:"data"`
}

type siteQualityStructuredDataSurface struct {
	Format   string   `json:"format"`
	Types    []string `json:"types"`
	Type     string   `json:"type"`
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	URL      string   `json:"url"`
	Selector string   `json:"selector"`
	Snippet  string   `json:"snippet"`
}

type siteQualityStructuredDataNodeView struct {
	siteQualityStructuredDataNode
	DataMap map[string]interface{}
	Format  string
}

type siteQualityStructuredDataReport struct {
	Nodes        []siteQualityStructuredDataNodeView
	Microdata    []siteQualityStructuredDataSurface
	RDFa         []siteQualityStructuredDataSurface
	Types        map[string][]siteQualityStructuredDataNodeView
	SurfaceTypes map[string][]siteQualityStructuredDataSurface
}

type siteQualityStructuredDataExpectedGroup struct {
	Label       string
	Types       []string
	Description string
}

type siteQualityStructuredDataPageIntent struct {
	Source     string
	SourceType string
	Title      string
	Locale     string
}

func siteQualityRenderedStructuredDataAuditIssues(
	targetURL string,
	lighthouseFinalURL string,
	audit *siteQualityRenderedStructuredDataAudit,
	intents ...siteQualityStructuredDataPageIntent,
) []LighthouseRunnerIssue {
	if audit == nil {
		return []LighthouseRunnerIssue{
			siteQualityStructuredDataScanFailureIssue("The runner response did not include a browser-rendered structured data snapshot."),
		}
	}
	if strings.TrimSpace(audit.Status) != "complete" {
		reason := strings.TrimSpace(audit.Error)
		if reason == "" {
			reason = "The browser-rendered structured data snapshot did not complete."
		}
		return []LighthouseRunnerIssue{siteQualityStructuredDataScanFailureIssue(reason)}
	}
	if !siteQualityRenderedHeadingFinalURLTrusted(targetURL, lighthouseFinalURL, audit.FinalURL) {
		return []LighthouseRunnerIssue{
			siteQualityStructuredDataScanFailureIssue("The browser-rendered structured data snapshot reached an unexpected final URL."),
		}
	}

	report := siteQualityStructuredDataReportFromAudit(audit)
	issues := make([]LighthouseRunnerIssue, 0, 6)
	invalidJSONLD := siteQualityStructuredDataInvalidJSONLDIssue(audit)
	if invalidJSONLD != nil {
		issues = append(issues, *invalidJSONLD)
	}

	intent := siteQualityStructuredDataPrimaryIntent(intents...)
	expectedGroups := siteQualityStructuredDataExpectedGroups(targetURL, lighthouseFinalURL, audit, intent)
	formatCount := siteQualityStructuredDataFormatCount(report)
	if len(expectedGroups) > 0 && formatCount == 0 && invalidJSONLD == nil {
		issues = append(issues, siteQualityStructuredDataMissingStructuredDataIssue(expectedGroups))
	}
	if formatCount > 0 {
		if missingRequired := siteQualityStructuredDataMissingRequiredTypeIssue(report, expectedGroups); missingRequired != nil {
			issues = append(issues, *missingRequired)
		}
	}

	issues = append(issues, siteQualityStructuredDataProductIssues(targetURL, lighthouseFinalURL, report)...)
	issues = append(issues, siteQualityStructuredDataArticleIssues(targetURL, lighthouseFinalURL, report)...)
	issues = append(issues, siteQualityStructuredDataOrganizationIssues(targetURL, lighthouseFinalURL, report, audit)...)
	issues = append(issues, siteQualityStructuredDataWebPageIssues(targetURL, lighthouseFinalURL, report)...)
	issues = append(issues, siteQualityStructuredDataFAQIssues(targetURL, lighthouseFinalURL, report, audit)...)
	issues = append(issues, siteQualityStructuredDataBreadcrumbIssues(report)...)
	return issues
}

func siteQualityStructuredDataReportFromAudit(
	audit *siteQualityRenderedStructuredDataAudit,
) siteQualityStructuredDataReport {
	report := siteQualityStructuredDataReport{
		Microdata:    audit.Microdata,
		RDFa:         audit.RDFa,
		Types:        make(map[string][]siteQualityStructuredDataNodeView),
		SurfaceTypes: make(map[string][]siteQualityStructuredDataSurface),
	}
	for _, script := range audit.JSONLD {
		for _, node := range script.Nodes {
			node.Selector = strings.TrimSpace(node.Selector)
			if node.Selector == "" {
				node.Selector = strings.TrimSpace(script.Selector)
			}
			view := siteQualityStructuredDataNodeView{
				siteQualityStructuredDataNode: node,
				DataMap:                       siteQualityStructuredDataNodeData(node),
				Format:                        "json-ld",
			}
			report.Nodes = append(report.Nodes, view)
			for _, nodeType := range siteQualityStructuredDataNodeTypes(node) {
				key := siteQualitySchemaTypeKey(nodeType)
				if key != "" {
					report.Types[key] = append(report.Types[key], view)
				}
			}
		}
	}
	for _, surface := range append(append([]siteQualityStructuredDataSurface{}, audit.Microdata...), audit.RDFa...) {
		for _, nodeType := range siteQualityStructuredDataSurfaceTypes(surface) {
			key := siteQualitySchemaTypeKey(nodeType)
			if key != "" {
				report.SurfaceTypes[key] = append(report.SurfaceTypes[key], surface)
			}
		}
	}
	return report
}

func siteQualityStructuredDataInvalidJSONLDIssue(
	audit *siteQualityRenderedStructuredDataAudit,
) *LighthouseRunnerIssue {
	evidence := make([]sitequalitydomain.SiteQualityStructuredDataEvidence, 0)
	for _, script := range audit.JSONLD {
		parseError := strings.TrimSpace(script.ParseError)
		if parseError == "" {
			continue
		}
		evidence = append(evidence, sitequalitydomain.SiteQualityStructuredDataEvidence{
			Format:      "json-ld",
			Type:        "JSON-LD",
			Selector:    strings.TrimSpace(script.Selector),
			Snippet:     strings.TrimSpace(script.Raw),
			Property:    "script",
			Explanation: parseError,
		})
	}
	if len(evidence) == 0 {
		return nil
	}
	return &LighthouseRunnerIssue{
		ID:             siteQualityStructuredDataInvalidJSONLDAuditID,
		Kind:           "schema",
		RuleVersion:    siteQualityAuditRuleVersion,
		Title:          "JSON-LD cannot be parsed",
		Description:    "Every application/ld+json script must parse as valid JSON before search engines can understand the page entities.",
		Severity:       "critical",
		StructuredData: evidence,
	}
}

func siteQualityStructuredDataMissingStructuredDataIssue(
	expectedGroups []siteQualityStructuredDataExpectedGroup,
) LighthouseRunnerIssue {
	labels := make([]string, 0, len(expectedGroups))
	for _, group := range expectedGroups {
		labels = append(labels, group.Label)
	}
	sort.Strings(labels)
	return LighthouseRunnerIssue{
		ID:          siteQualityStructuredDataMissingStructuredDataAuditID,
		Kind:        "schema",
		RuleVersion: siteQualityAuditRuleVersion,
		Title:       "Page has no detectable structured data",
		Description: "This page type should expose structured data, but the rendered DOM did not contain JSON-LD, microdata, or RDFa entities.",
		Severity:    "high",
		StructuredData: []sitequalitydomain.SiteQualityStructuredDataEvidence{
			{
				Property:    "@type",
				Explanation: "Expected " + strings.Join(labels, ", ") + " structured data.",
			},
		},
	}
}

func siteQualityStructuredDataMissingRequiredTypeIssue(
	report siteQualityStructuredDataReport,
	expectedGroups []siteQualityStructuredDataExpectedGroup,
) *LighthouseRunnerIssue {
	if len(expectedGroups) == 0 {
		return nil
	}
	missing := make([]siteQualityStructuredDataExpectedGroup, 0)
	for _, group := range expectedGroups {
		if !siteQualityStructuredDataReportHasAnyType(report, group.Types) {
			missing = append(missing, group)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	foundTypes := siteQualityStructuredDataFoundTypes(report)
	expectedLabels := make([]string, 0, len(missing))
	for _, group := range missing {
		expectedLabels = append(expectedLabels, group.Label)
	}
	sort.Strings(expectedLabels)
	return &LighthouseRunnerIssue{
		ID:          siteQualityStructuredDataMissingRequiredTypeAuditID,
		Kind:        "schema",
		RuleVersion: siteQualityAuditRuleVersion,
		Title:       "Structured data is missing a required page type",
		Description: "The page route implies " + strings.Join(expectedLabels, ", ") + " structured data, but the rendered structured data did not expose that type.",
		Severity:    "high",
		StructuredData: []sitequalitydomain.SiteQualityStructuredDataEvidence{
			{
				Format:      "structured-data",
				Property:    "@type",
				Type:        strings.Join(foundTypes, ", "),
				Explanation: "Expected one of: " + strings.Join(expectedLabels, ", "),
			},
		},
	}
}

func siteQualityStructuredDataProductIssues(
	targetURL string,
	lighthouseFinalURL string,
	report siteQualityStructuredDataReport,
) []LighthouseRunnerIssue {
	productNodes := siteQualityStructuredDataPrimaryNodesWithTypes(report, []string{"Product", "ProductGroup"})
	if len(productNodes) == 0 {
		return nil
	}

	issues := make([]LighthouseRunnerIssue, 0, 3)
	uniqueProductNodes := siteQualityStructuredDataUniquePrimaryEntityNodes(productNodes)
	if len(uniqueProductNodes) > 1 {
		evidence := make([]sitequalitydomain.SiteQualityStructuredDataEvidence, 0, len(uniqueProductNodes))
		for _, node := range uniqueProductNodes {
			evidence = append(evidence, siteQualityStructuredDataEvidenceFromNode(node, "Multiple top-level product entities compete as the primary page entity.", "@type"))
		}
		issues = append(issues, LighthouseRunnerIssue{
			ID:             siteQualityStructuredDataDuplicatePrimaryTypeAuditID,
			Kind:           "schema",
			RuleVersion:    siteQualityAuditRuleVersion,
			Title:          "Page has duplicate primary product schema",
			Description:    "A product detail page should expose one primary Product or ProductGroup entity. Variant products should be nested under ProductGroup.hasVariant.",
			Severity:       "medium",
			StructuredData: evidence,
		})
	}

	invalidEvidence := make([]sitequalitydomain.SiteQualityStructuredDataEvidence, 0)
	urlMismatchEvidence := make([]sitequalitydomain.SiteQualityStructuredDataEvidence, 0)
	for _, node := range productNodes {
		nodeType := siteQualityStructuredDataPrimaryType(node)
		missing := siteQualityStructuredDataProductMissingFields(node)
		for _, property := range missing {
			invalidEvidence = append(invalidEvidence, siteQualityStructuredDataEvidenceFromNode(
				node,
				fmt.Sprintf("%s is required for the primary %s entity.", property, nodeType),
				property,
			))
		}
		schemaURL := siteQualityStructuredDataStringField(node.DataMap, "url")
		if schemaURL != "" && !siteQualityStructuredDataURLMatchesPage(schemaURL, targetURL, lighthouseFinalURL) {
			urlMismatchEvidence = append(urlMismatchEvidence, siteQualityStructuredDataEvidenceFromNode(
				node,
				"Schema URL should resolve to the audited page canonical URL.",
				"url",
			))
		}
	}
	if len(invalidEvidence) > 0 {
		issues = append(issues, LighthouseRunnerIssue{
			ID:             siteQualityStructuredDataProductInvalidAuditID,
			Kind:           "schema",
			RuleVersion:    siteQualityAuditRuleVersion,
			Title:          "Product structured data is incomplete",
			Description:    "Product and ProductGroup entities need enough information for search engines and shopping surfaces to identify the product, images, variants, and offer availability.",
			Severity:       "critical",
			StructuredData: invalidEvidence,
		})
	}
	if len(urlMismatchEvidence) > 0 {
		issues = append(issues, LighthouseRunnerIssue{
			ID:             siteQualityStructuredDataURLMismatchAuditID,
			Kind:           "schema",
			RuleVersion:    siteQualityAuditRuleVersion,
			Title:          "Structured data URL does not match the audited page",
			Description:    "Primary page entities should point at the canonical URL for the page being audited, not another storefront route.",
			Severity:       "high",
			StructuredData: urlMismatchEvidence,
		})
	}
	return issues
}

func siteQualityStructuredDataArticleIssues(
	targetURL string,
	lighthouseFinalURL string,
	report siteQualityStructuredDataReport,
) []LighthouseRunnerIssue {
	articleNodes := siteQualityStructuredDataPrimaryNodesWithTypes(report, []string{"Article", "BlogPosting", "NewsArticle"})
	if len(articleNodes) == 0 {
		return nil
	}

	invalidEvidence := make([]sitequalitydomain.SiteQualityStructuredDataEvidence, 0)
	urlMismatchEvidence := make([]sitequalitydomain.SiteQualityStructuredDataEvidence, 0)
	for _, node := range articleNodes {
		for _, property := range siteQualityStructuredDataArticleMissingFields(node) {
			invalidEvidence = append(invalidEvidence, siteQualityStructuredDataEvidenceFromNode(
				node,
				fmt.Sprintf("%s is required for the primary article entity.", property),
				property,
			))
		}
		schemaURL := siteQualityStructuredDataPrimaryEntityURL(node)
		if schemaURL != "" && !siteQualityStructuredDataURLMatchesPage(schemaURL, targetURL, lighthouseFinalURL) {
			urlMismatchEvidence = append(urlMismatchEvidence, siteQualityStructuredDataEvidenceFromNode(
				node,
				"Article URL should resolve to the audited page canonical URL.",
				"url",
			))
		}
	}

	issues := make([]LighthouseRunnerIssue, 0, 2)
	if len(invalidEvidence) > 0 {
		issues = append(issues, LighthouseRunnerIssue{
			ID:             siteQualityStructuredDataArticleInvalidAuditID,
			Kind:           "schema",
			RuleVersion:    siteQualityAuditRuleVersion,
			Title:          "Article structured data is incomplete",
			Description:    "Article, BlogPosting, and NewsArticle entities must expose headline, image, publication timing, authorship or publisher, and a page URL.",
			Severity:       "high",
			StructuredData: invalidEvidence,
		})
	}
	if len(urlMismatchEvidence) > 0 {
		issues = append(issues, LighthouseRunnerIssue{
			ID:             siteQualityStructuredDataURLMismatchAuditID,
			Kind:           "schema",
			RuleVersion:    siteQualityAuditRuleVersion,
			Title:          "Structured data URL does not match the audited page",
			Description:    "Primary page entities should point at the canonical URL for the page being audited, not another storefront route.",
			Severity:       "high",
			StructuredData: urlMismatchEvidence,
		})
	}
	return issues
}

func siteQualityStructuredDataOrganizationIssues(
	targetURL string,
	lighthouseFinalURL string,
	report siteQualityStructuredDataReport,
	audit *siteQualityRenderedStructuredDataAudit,
) []LighthouseRunnerIssue {
	orgNodes := siteQualityStructuredDataPrimaryNodesWithTypes(report, []string{"Organization", "LocalBusiness"})
	if len(orgNodes) == 0 {
		return nil
	}

	pageURL := lighthouseFinalURL
	if strings.TrimSpace(pageURL) == "" {
		pageURL = targetURL
	}
	requireIdentityLogo := siteQualityStructuredDataShouldRequireOrganizationLogo(targetURL, lighthouseFinalURL, audit)
	invalidEvidence := make([]sitequalitydomain.SiteQualityStructuredDataEvidence, 0)
	urlMismatchEvidence := make([]sitequalitydomain.SiteQualityStructuredDataEvidence, 0)
	for _, node := range orgNodes {
		for _, property := range siteQualityStructuredDataOrganizationMissingFields(node, requireIdentityLogo) {
			invalidEvidence = append(invalidEvidence, siteQualityStructuredDataEvidenceFromNode(
				node,
				fmt.Sprintf("%s is required for the %s entity.", property, siteQualityStructuredDataPrimaryType(node)),
				property,
			))
		}
		schemaURL := siteQualityStructuredDataPrimaryEntityURL(node)
		if schemaURL != "" && !siteQualityStructuredDataURLMatchesOrigin(schemaURL, pageURL) {
			urlMismatchEvidence = append(urlMismatchEvidence, siteQualityStructuredDataEvidenceFromNode(
				node,
				"Organization and LocalBusiness URLs should stay on the audited storefront origin.",
				"url",
			))
		}
	}

	issues := make([]LighthouseRunnerIssue, 0, 2)
	if len(invalidEvidence) > 0 {
		issues = append(issues, LighthouseRunnerIssue{
			ID:             siteQualityStructuredDataOrganizationInvalidAuditID,
			Kind:           "schema",
			RuleVersion:    siteQualityAuditRuleVersion,
			Title:          "Organization structured data is incomplete",
			Description:    "Organization and LocalBusiness entities need name, URL, and enough logo or contact identity for search engines to connect the storefront brand to the rendered page.",
			Severity:       "high",
			StructuredData: invalidEvidence,
		})
	}
	if len(urlMismatchEvidence) > 0 {
		issues = append(issues, LighthouseRunnerIssue{
			ID:             siteQualityStructuredDataURLMismatchAuditID,
			Kind:           "schema",
			RuleVersion:    siteQualityAuditRuleVersion,
			Title:          "Structured data URL does not match the storefront origin",
			Description:    "Organization and LocalBusiness schema should point to the audited storefront origin, not an unrelated domain.",
			Severity:       "high",
			StructuredData: urlMismatchEvidence,
		})
	}
	return issues
}

func siteQualityStructuredDataWebPageIssues(
	targetURL string,
	lighthouseFinalURL string,
	report siteQualityStructuredDataReport,
) []LighthouseRunnerIssue {
	webPageNodes := siteQualityStructuredDataPrimaryNodesWithTypes(report, []string{"WebPage", "CollectionPage", "AboutPage", "ContactPage"})
	if len(webPageNodes) == 0 {
		return nil
	}

	invalidEvidence := make([]sitequalitydomain.SiteQualityStructuredDataEvidence, 0)
	urlMismatchEvidence := make([]sitequalitydomain.SiteQualityStructuredDataEvidence, 0)
	for _, node := range webPageNodes {
		if siteQualityStructuredDataFirstStringField(node.DataMap, "name", "headline") == "" {
			invalidEvidence = append(invalidEvidence, siteQualityStructuredDataEvidenceFromNode(node, "WebPage schema should expose name or headline.", "name"))
		}
		schemaURL := siteQualityStructuredDataPrimaryEntityURL(node)
		if schemaURL == "" {
			invalidEvidence = append(invalidEvidence, siteQualityStructuredDataEvidenceFromNode(node, "WebPage schema should expose url, @id, or mainEntityOfPage.", "url"))
			continue
		}
		if !siteQualityStructuredDataURLMatchesPage(schemaURL, targetURL, lighthouseFinalURL) {
			urlMismatchEvidence = append(urlMismatchEvidence, siteQualityStructuredDataEvidenceFromNode(
				node,
				"WebPage URL should resolve to the audited page canonical URL.",
				"url",
			))
		}
	}

	issues := make([]LighthouseRunnerIssue, 0, 2)
	if len(invalidEvidence) > 0 {
		issues = append(issues, LighthouseRunnerIssue{
			ID:             siteQualityStructuredDataWebPageInvalidAuditID,
			Kind:           "schema",
			RuleVersion:    siteQualityAuditRuleVersion,
			Title:          "WebPage structured data is incomplete",
			Description:    "When a page emits WebPage-like schema, it must identify the rendered page with a name and URL.",
			Severity:       "medium",
			StructuredData: invalidEvidence,
		})
	}
	if len(urlMismatchEvidence) > 0 {
		issues = append(issues, LighthouseRunnerIssue{
			ID:             siteQualityStructuredDataURLMismatchAuditID,
			Kind:           "schema",
			RuleVersion:    siteQualityAuditRuleVersion,
			Title:          "Structured data URL does not match the audited page",
			Description:    "Primary page entities should point at the canonical URL for the page being audited, not another storefront route.",
			Severity:       "high",
			StructuredData: urlMismatchEvidence,
		})
	}
	return issues
}

func siteQualityStructuredDataFAQIssues(
	targetURL string,
	lighthouseFinalURL string,
	report siteQualityStructuredDataReport,
	audit *siteQualityRenderedStructuredDataAudit,
) []LighthouseRunnerIssue {
	faqNodes := siteQualityStructuredDataPrimaryNodesWithTypes(report, []string{"FAQPage"})
	if len(faqNodes) == 0 {
		return nil
	}
	evidence := make([]sitequalitydomain.SiteQualityStructuredDataEvidence, 0)
	mismatchEvidence := make([]sitequalitydomain.SiteQualityStructuredDataEvidence, 0)
	for _, node := range faqNodes {
		mainEntities := siteQualityStructuredDataArrayOrSingle(node.DataMap["mainEntity"])
		if len(mainEntities) == 0 {
			evidence = append(evidence, siteQualityStructuredDataEvidenceFromNode(node, "FAQPage.mainEntity must contain at least one Question.", "mainEntity"))
			continue
		}
		for index, entity := range mainEntities {
			question, ok := entity.(map[string]interface{})
			if !ok {
				evidence = append(evidence, siteQualityStructuredDataEvidenceFromNode(node, fmt.Sprintf("FAQPage.mainEntity[%d] is not an object.", index), "mainEntity"))
				continue
			}
			if !siteQualityStructuredDataValueHasType(question["@type"], "Question") {
				evidence = append(evidence, siteQualityStructuredDataEvidenceFromNode(node, fmt.Sprintf("FAQPage.mainEntity[%d] must be a Question.", index), "mainEntity.@type"))
			}
			if siteQualityStructuredDataStringField(question, "name") == "" {
				evidence = append(evidence, siteQualityStructuredDataEvidenceFromNode(node, fmt.Sprintf("Question %d is missing name.", index+1), "mainEntity.name"))
			}
			if !siteQualityStructuredDataAcceptedAnswerValid(question["acceptedAnswer"]) {
				evidence = append(evidence, siteQualityStructuredDataEvidenceFromNode(node, fmt.Sprintf("Question %d is missing acceptedAnswer.text.", index+1), "mainEntity.acceptedAnswer"))
			}
		}
		if renderedCount := siteQualityStructuredDataExpectedFAQQuestionCount(targetURL, lighthouseFinalURL, audit); renderedCount >= 2 && len(mainEntities) > 0 && len(mainEntities) < renderedCount {
			mismatchEvidence = append(mismatchEvidence, siteQualityStructuredDataEvidenceFromNode(
				node,
				fmt.Sprintf("Rendered FAQ-like questions: %d; FAQPage.mainEntity questions: %d.", renderedCount, len(mainEntities)),
				"mainEntity",
			))
		}
	}
	issues := make([]LighthouseRunnerIssue, 0, 2)
	if len(evidence) > 0 {
		issues = append(issues, LighthouseRunnerIssue{
			ID:             siteQualityStructuredDataFAQInvalidAuditID,
			Kind:           "schema",
			RuleVersion:    siteQualityAuditRuleVersion,
			Title:          "FAQPage structured data is incomplete",
			Description:    "FAQPage structured data must expose Question entries with accepted answers so crawlers can map rendered FAQ content to schema entities.",
			Severity:       "high",
			StructuredData: evidence,
		})
	}
	if len(mismatchEvidence) > 0 {
		issues = append(issues, LighthouseRunnerIssue{
			ID:             siteQualityStructuredDataFAQContentMismatchAuditID,
			Kind:           "schema",
			RuleVersion:    siteQualityAuditRuleVersion,
			Title:          "FAQPage structured data does not cover rendered FAQ content",
			Description:    "The rendered FAQ page appears to expose more question rows than FAQPage.mainEntity contains, so search engines may miss visible answers.",
			Severity:       "high",
			StructuredData: mismatchEvidence,
		})
	}
	return issues
}

func siteQualityStructuredDataBreadcrumbIssues(report siteQualityStructuredDataReport) []LighthouseRunnerIssue {
	breadcrumbNodes := siteQualityStructuredDataPrimaryNodesWithTypes(report, []string{"BreadcrumbList"})
	if len(breadcrumbNodes) == 0 {
		return nil
	}
	evidence := make([]sitequalitydomain.SiteQualityStructuredDataEvidence, 0)
	for _, node := range breadcrumbNodes {
		items := siteQualityStructuredDataArrayOrSingle(node.DataMap["itemListElement"])
		if len(items) == 0 {
			evidence = append(evidence, siteQualityStructuredDataEvidenceFromNode(node, "BreadcrumbList.itemListElement must contain at least one ListItem.", "itemListElement"))
			continue
		}
		for index, item := range items {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				evidence = append(evidence, siteQualityStructuredDataEvidenceFromNode(node, fmt.Sprintf("Breadcrumb item %d is not an object.", index+1), "itemListElement"))
				continue
			}
			position := siteQualityStructuredDataNumberField(itemMap, "position")
			if int(position) != index+1 {
				evidence = append(evidence, siteQualityStructuredDataEvidenceFromNode(node, fmt.Sprintf("Breadcrumb item %d has position %.0f; expected %d.", index+1, position, index+1), "itemListElement.position"))
			}
			if siteQualityStructuredDataStringField(itemMap, "name") == "" {
				evidence = append(evidence, siteQualityStructuredDataEvidenceFromNode(node, fmt.Sprintf("Breadcrumb item %d is missing name.", index+1), "itemListElement.name"))
			}
			if !siteQualityStructuredDataBreadcrumbItemHasURL(itemMap["item"]) {
				evidence = append(evidence, siteQualityStructuredDataEvidenceFromNode(node, fmt.Sprintf("Breadcrumb item %d is missing item URL.", index+1), "itemListElement.item"))
			}
		}
	}
	if len(evidence) == 0 {
		return nil
	}
	return []LighthouseRunnerIssue{
		{
			ID:             siteQualityStructuredDataBreadcrumbInvalidAuditID,
			Kind:           "schema",
			RuleVersion:    siteQualityAuditRuleVersion,
			Title:          "Breadcrumb structured data is incomplete",
			Description:    "BreadcrumbList entries must have sequential positions, names, and item URLs.",
			Severity:       "high",
			StructuredData: evidence,
		},
	}
}

func siteQualityStructuredDataProductMissingFields(
	node siteQualityStructuredDataNodeView,
) []string {
	missing := make([]string, 0)
	if siteQualityStructuredDataStringField(node.DataMap, "name") == "" {
		missing = append(missing, "name")
	}
	if siteQualityStructuredDataStringField(node.DataMap, "url") == "" {
		missing = append(missing, "url")
	}
	if !siteQualityStructuredDataHasValue(node.DataMap["image"]) {
		missing = append(missing, "image")
	}
	nodeType := siteQualityStructuredDataPrimaryType(node)
	if nodeType == "ProductGroup" {
		if len(siteQualityStructuredDataArrayOrSingle(node.DataMap["hasVariant"])) < 2 {
			missing = append(missing, "hasVariant")
		}
		return missing
	}
	if !siteQualityStructuredDataOfferValid(node.DataMap["offers"]) {
		missing = append(missing, "offers")
	}
	return missing
}

func siteQualityStructuredDataArticleMissingFields(
	node siteQualityStructuredDataNodeView,
) []string {
	missing := make([]string, 0)
	if siteQualityStructuredDataFirstStringField(node.DataMap, "headline", "name") == "" {
		missing = append(missing, "headline")
	}
	if !siteQualityStructuredDataImageValid(node.DataMap["image"]) {
		missing = append(missing, "image")
	}
	if siteQualityStructuredDataFirstStringField(node.DataMap, "datePublished", "dateModified") == "" {
		missing = append(missing, "datePublished")
	}
	if !siteQualityStructuredDataEntityReferenceValid(node.DataMap["author"]) &&
		!siteQualityStructuredDataEntityReferenceValid(node.DataMap["publisher"]) {
		missing = append(missing, "author")
	}
	if siteQualityStructuredDataPrimaryEntityURL(node) == "" {
		missing = append(missing, "url")
	}
	return missing
}

func siteQualityStructuredDataOrganizationMissingFields(
	node siteQualityStructuredDataNodeView,
	requireLogo bool,
) []string {
	missing := make([]string, 0)
	if siteQualityStructuredDataStringField(node.DataMap, "name") == "" {
		missing = append(missing, "name")
	}
	if siteQualityStructuredDataPrimaryEntityURL(node) == "" {
		missing = append(missing, "url")
	}
	nodeType := siteQualityStructuredDataPrimaryType(node)
	if nodeType == "LocalBusiness" {
		if !siteQualityStructuredDataLocalBusinessContactValid(node.DataMap) {
			missing = append(missing, "address")
		}
		return missing
	}
	if requireLogo && !siteQualityStructuredDataImageValid(node.DataMap["logo"]) {
		missing = append(missing, "logo")
	}
	return missing
}

func siteQualityStructuredDataShouldRequireOrganizationLogo(
	targetURL string,
	lighthouseFinalURL string,
	audit *siteQualityRenderedStructuredDataAudit,
) bool {
	path := siteQualityStructuredDataComparablePath(targetURL)
	if finalPath := siteQualityStructuredDataComparablePath(lighthouseFinalURL); finalPath != "" {
		path = finalPath
	}
	if audit != nil {
		if canonicalPath := siteQualityStructuredDataComparablePath(audit.Page.CanonicalURL); canonicalPath != "" {
			path = canonicalPath
		}
	}
	return path == "/"
}

func siteQualityStructuredDataExpectedFAQQuestionCount(
	targetURL string,
	lighthouseFinalURL string,
	audit *siteQualityRenderedStructuredDataAudit,
) int {
	if audit == nil || audit.Page.FAQQuestionCount < 2 {
		return 0
	}
	for _, rawURL := range []string{targetURL, lighthouseFinalURL, audit.Page.CanonicalURL} {
		if siteQualityStructuredDataPathLooksLikeFAQ(siteQualityStructuredDataComparablePath(rawURL)) {
			return audit.Page.FAQQuestionCount
		}
	}
	return 0
}

func siteQualityStructuredDataPrimaryIntent(
	intents ...siteQualityStructuredDataPageIntent,
) siteQualityStructuredDataPageIntent {
	if len(intents) == 0 {
		return siteQualityStructuredDataPageIntent{}
	}
	intent := intents[0]
	intent.Source = strings.ToLower(strings.TrimSpace(intent.Source))
	intent.SourceType = strings.ToLower(strings.TrimSpace(intent.SourceType))
	intent.Title = strings.TrimSpace(intent.Title)
	intent.Locale = strings.TrimSpace(intent.Locale)
	return intent
}

func siteQualityStructuredDataExpectedGroups(
	targetURL string,
	lighthouseFinalURL string,
	audit *siteQualityRenderedStructuredDataAudit,
	intent siteQualityStructuredDataPageIntent,
) []siteQualityStructuredDataExpectedGroup {
	path := siteQualityStructuredDataComparablePath(targetURL)
	if finalPath := siteQualityStructuredDataComparablePath(lighthouseFinalURL); finalPath != "" {
		path = finalPath
	}
	if canonicalPath := siteQualityStructuredDataComparablePath(audit.Page.CanonicalURL); canonicalPath != "" {
		path = canonicalPath
	}

	groups := make([]siteQualityStructuredDataExpectedGroup, 0, 2)
	sourceType := strings.ToLower(strings.TrimSpace(intent.SourceType))
	switch sourceType {
	case seodomain.RouteSourceProduct:
		groups = append(groups, siteQualityStructuredDataExpectedGroup{
			Label:       "Product or ProductGroup",
			Types:       []string{"Product", "ProductGroup"},
			Description: "route catalog product detail page",
		})
	case seodomain.RouteSourceBlog:
		groups = append(groups, siteQualityStructuredDataExpectedGroup{
			Label:       "Article, BlogPosting, or NewsArticle",
			Types:       []string{"Article", "BlogPosting", "NewsArticle"},
			Description: "route catalog article page",
		})
	}

	switch {
	case path == "/":
		groups = appendSiteQualityStructuredDataExpectedGroup(groups, siteQualityStructuredDataExpectedGroup{
			Label:       "Organization",
			Types:       []string{"Organization", "LocalBusiness"},
			Description: "home page site identity",
		})
	case strings.HasPrefix(path, "/products/") && path != "/products/":
		groups = appendSiteQualityStructuredDataExpectedGroup(groups, siteQualityStructuredDataExpectedGroup{
			Label:       "Product or ProductGroup",
			Types:       []string{"Product", "ProductGroup"},
			Description: "product detail page",
		})
	case strings.HasPrefix(path, "/resources/blog/") &&
		path != "/resources/blog/" &&
		!siteQualityStructuredDataPathLooksLikeBlogListing(path):
		groups = appendSiteQualityStructuredDataExpectedGroup(groups, siteQualityStructuredDataExpectedGroup{
			Label:       "Article, BlogPosting, or NewsArticle",
			Types:       []string{"Article", "BlogPosting", "NewsArticle"},
			Description: "article page",
		})
	case siteQualityStructuredDataPathLooksLikeFAQ(path):
		groups = appendSiteQualityStructuredDataExpectedGroup(groups, siteQualityStructuredDataExpectedGroup{
			Label:       "FAQPage",
			Types:       []string{"FAQPage"},
			Description: "FAQ page",
		})
	case strings.Contains(path, "/contact"):
		groups = appendSiteQualityStructuredDataExpectedGroup(groups, siteQualityStructuredDataExpectedGroup{
			Label:       "LocalBusiness",
			Types:       []string{"LocalBusiness", "Organization"},
			Description: "contact page",
		})
	}
	return groups
}

func appendSiteQualityStructuredDataExpectedGroup(
	groups []siteQualityStructuredDataExpectedGroup,
	next siteQualityStructuredDataExpectedGroup,
) []siteQualityStructuredDataExpectedGroup {
	nextKeys := make(map[string]struct{}, len(next.Types))
	for _, value := range next.Types {
		key := siteQualitySchemaTypeKey(value)
		if key != "" {
			nextKeys[key] = struct{}{}
		}
	}
	for _, group := range groups {
		for _, value := range group.Types {
			if _, exists := nextKeys[siteQualitySchemaTypeKey(value)]; exists {
				return groups
			}
		}
	}
	return append(groups, next)
}

func siteQualityStructuredDataPathLooksLikeFAQ(path string) bool {
	path = strings.ToLower(strings.TrimSpace(path))
	return strings.Contains(path, "/faq") || strings.Contains(path, "/faqs")
}

func siteQualityStructuredDataPathLooksLikeBlogListing(path string) bool {
	path = strings.Trim(strings.ToLower(strings.TrimSpace(path)), "/")
	parts := strings.Split(path, "/")
	return len(parts) == 3 &&
		parts[0] == "resources" &&
		parts[1] == "blog" &&
		(parts[2] == "news" || parts[2] == "wheelsbuild")
}

func siteQualityStructuredDataComparablePath(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed == nil {
		return ""
	}
	path := parsed.EscapedPath()
	if path == "" {
		path = parsed.Path
	}
	if path == "" {
		path = "/"
	}
	path = "/" + strings.Trim(strings.ToLower(path), "/")
	if path == "/" {
		return path
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 1 && siteQualityStructuredDataLooksLikeLocale(parts[0]) {
		path = "/" + strings.Join(parts[1:], "/")
	}
	if path != "/" {
		path = strings.TrimRight(path, "/")
	}
	if path == "" {
		return "/"
	}
	return path
}

func siteQualityStructuredDataLooksLikeLocale(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) == 2 {
		return true
	}
	if len(value) >= 4 && len(value) <= 8 && (strings.Contains(value, "-") || strings.Contains(value, "_")) {
		return true
	}
	return false
}

func siteQualityStructuredDataPrimaryNodesWithTypes(
	report siteQualityStructuredDataReport,
	types []string,
) []siteQualityStructuredDataNodeView {
	result := make([]siteQualityStructuredDataNodeView, 0)
	seen := make(map[string]struct{})
	for _, typeName := range types {
		for _, node := range report.Types[siteQualitySchemaTypeKey(typeName)] {
			if !siteQualityStructuredDataPrimaryNode(node) {
				continue
			}
			key := strings.Join([]string{node.GraphPath, node.Selector, node.ID, node.Name, siteQualityStructuredDataPrimaryType(node)}, "\x00")
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, node)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].GraphPath < result[j].GraphPath
	})
	return result
}

func siteQualityStructuredDataUniquePrimaryEntityNodes(
	nodes []siteQualityStructuredDataNodeView,
) []siteQualityStructuredDataNodeView {
	result := make([]siteQualityStructuredDataNodeView, 0, len(nodes))
	seen := make(map[string]struct{})
	for _, node := range nodes {
		key := siteQualityStructuredDataPrimaryEntityKey(node)
		if key == "" {
			key = strings.Join([]string{node.GraphPath, node.Selector, node.ID, node.Name, siteQualityStructuredDataPrimaryType(node)}, "\x00")
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, node)
	}
	return result
}

func siteQualityStructuredDataPrimaryEntityKey(node siteQualityStructuredDataNodeView) string {
	nodeType := siteQualitySchemaTypeKey(siteQualityStructuredDataPrimaryType(node))
	for _, value := range []string{
		node.ID,
		siteQualityStructuredDataPrimaryEntityURL(node),
		node.URL,
	} {
		normalized := strings.ToLower(siteQualityStructuredDataNormalizeURL(value))
		if normalized != "" {
			return nodeType + "\x00" + normalized
		}
	}
	name := strings.ToLower(strings.TrimSpace(node.Name))
	if name == "" {
		name = strings.ToLower(siteQualityStructuredDataFirstStringField(node.DataMap, "name", "headline"))
	}
	if name == "" {
		return ""
	}
	return nodeType + "\x00" + name
}

func siteQualityStructuredDataPrimaryNode(node siteQualityStructuredDataNodeView) bool {
	path := strings.TrimSpace(node.GraphPath)
	if path == "" {
		return true
	}
	for _, nested := range []string{
		".mainEntity",
		".itemListElement",
		".hasVariant",
		".item",
		".offers",
		".brand",
		".aggregateRating",
		".acceptedAnswer",
		".address",
		".geo",
	} {
		if strings.Contains(path, nested) {
			return false
		}
	}
	return true
}

func siteQualityStructuredDataReportHasAnyType(
	report siteQualityStructuredDataReport,
	types []string,
) bool {
	for _, typeName := range types {
		key := siteQualitySchemaTypeKey(typeName)
		if key == "" {
			continue
		}
		if len(report.Types[key]) > 0 || len(report.SurfaceTypes[key]) > 0 {
			return true
		}
	}
	return false
}

func siteQualityStructuredDataFoundTypes(report siteQualityStructuredDataReport) []string {
	types := make([]string, 0, len(report.Types)+len(report.SurfaceTypes))
	seen := make(map[string]struct{})
	for _, bucket := range report.Types {
		for _, node := range bucket {
			for _, typeName := range siteQualityStructuredDataNodeTypes(node.siteQualityStructuredDataNode) {
				key := siteQualitySchemaTypeKey(typeName)
				if key == "" {
					continue
				}
				if _, exists := seen[key]; exists {
					continue
				}
				seen[key] = struct{}{}
				types = append(types, siteQualitySchemaTypeLabel(typeName))
			}
		}
	}
	for _, bucket := range report.SurfaceTypes {
		for _, surface := range bucket {
			for _, typeName := range siteQualityStructuredDataSurfaceTypes(surface) {
				key := siteQualitySchemaTypeKey(typeName)
				if key == "" {
					continue
				}
				if _, exists := seen[key]; exists {
					continue
				}
				seen[key] = struct{}{}
				types = append(types, siteQualitySchemaTypeLabel(typeName))
			}
		}
	}
	sort.Strings(types)
	if len(types) == 0 {
		return []string{"none"}
	}
	return types
}

func siteQualityStructuredDataFormatCount(report siteQualityStructuredDataReport) int {
	return len(report.Nodes) + len(report.Microdata) + len(report.RDFa)
}

func siteQualityStructuredDataNodeTypes(node siteQualityStructuredDataNode) []string {
	values := append([]string{}, node.Types...)
	if strings.TrimSpace(node.Type) != "" {
		values = append(values, node.Type)
	}
	return siteQualityNormalizeSchemaTypes(values)
}

func siteQualityStructuredDataSurfaceTypes(surface siteQualityStructuredDataSurface) []string {
	values := append([]string{}, surface.Types...)
	if strings.TrimSpace(surface.Type) != "" {
		values = append(values, surface.Type)
	}
	return siteQualityNormalizeSchemaTypes(values)
}

func siteQualityNormalizeSchemaTypes(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{})
	for _, value := range values {
		label := siteQualitySchemaTypeLabel(value)
		key := siteQualitySchemaTypeKey(label)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, label)
	}
	return result
}

func siteQualitySchemaTypeKey(value string) string {
	return strings.ToLower(siteQualitySchemaTypeLabel(value))
}

func siteQualitySchemaTypeLabel(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return ""
	}
	normalized = strings.TrimRight(normalized, "/#")
	if idx := strings.LastIndex(normalized, "#"); idx >= 0 {
		normalized = normalized[idx+1:]
	}
	if idx := strings.LastIndex(normalized, "/"); idx >= 0 {
		normalized = normalized[idx+1:]
	}
	return strings.TrimSpace(normalized)
}

func siteQualityStructuredDataPrimaryType(node siteQualityStructuredDataNodeView) string {
	for _, value := range siteQualityStructuredDataNodeTypes(node.siteQualityStructuredDataNode) {
		if value != "" {
			return value
		}
	}
	return strings.TrimSpace(node.Type)
}

func siteQualityStructuredDataPrimaryEntityURL(node siteQualityStructuredDataNodeView) string {
	if node.URL != "" {
		return siteQualityStructuredDataNormalizeURL(node.URL)
	}
	for _, key := range []string{"url", "@id", "mainEntityOfPage"} {
		if value := siteQualityStructuredDataURLValue(node.DataMap[key]); value != "" {
			return value
		}
	}
	return ""
}

func siteQualityStructuredDataNodeData(node siteQualityStructuredDataNode) map[string]interface{} {
	if len(node.Data) == 0 {
		return nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil
	}
	return data
}

func siteQualityStructuredDataFirstStringField(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value := siteQualityStructuredDataStringField(data, key); value != "" {
			return value
		}
	}
	return ""
}

func siteQualityStructuredDataEvidenceFromNode(
	node siteQualityStructuredDataNodeView,
	explanation string,
	property string,
) sitequalitydomain.SiteQualityStructuredDataEvidence {
	return sitequalitydomain.SiteQualityStructuredDataEvidence{
		Format:      node.Format,
		Type:        siteQualityStructuredDataPrimaryType(node),
		ID:          strings.TrimSpace(node.ID),
		Name:        strings.TrimSpace(node.Name),
		URL:         siteQualityStructuredDataNormalizeURL(node.URL),
		Selector:    strings.TrimSpace(node.Selector),
		Property:    strings.TrimSpace(property),
		Explanation: strings.TrimSpace(explanation),
	}
}

func siteQualityStructuredDataStringField(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	return siteQualityStructuredDataStringValue(data[key])
}

func siteQualityStructuredDataStringValue(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	case float64:
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	case bool:
		return fmt.Sprintf("%t", typed)
	default:
		return ""
	}
}

func siteQualityStructuredDataURLValue(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return siteQualityStructuredDataNormalizeURL(typed)
	case map[string]interface{}:
		for _, key := range []string{"url", "@id", "contentUrl", "image"} {
			if nested := siteQualityStructuredDataURLValue(typed[key]); nested != "" {
				return nested
			}
		}
	case []interface{}:
		for _, item := range typed {
			if nested := siteQualityStructuredDataURLValue(item); nested != "" {
				return nested
			}
		}
	}
	return ""
}

func siteQualityStructuredDataNormalizeURL(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return ""
	}
	if canonical, err := canonicalizeAbsoluteSiteQualityURL(normalized); err == nil {
		return canonical
	}
	return normalized
}

func siteQualityStructuredDataImageValid(value interface{}) bool {
	for _, item := range siteQualityStructuredDataArrayOrSingle(value) {
		switch typed := item.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return true
			}
		case map[string]interface{}:
			if siteQualityStructuredDataURLValue(typed) != "" ||
				siteQualityStructuredDataStringField(typed, "caption") != "" {
				return true
			}
		}
	}
	return false
}

func siteQualityStructuredDataEntityReferenceValid(value interface{}) bool {
	for _, item := range siteQualityStructuredDataArrayOrSingle(value) {
		switch typed := item.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return true
			}
		case map[string]interface{}:
			if siteQualityStructuredDataStringField(typed, "name") != "" ||
				siteQualityStructuredDataStringField(typed, "@id") != "" ||
				siteQualityStructuredDataStringField(typed, "url") != "" {
				return true
			}
		}
	}
	return false
}

func siteQualityStructuredDataLocalBusinessContactValid(data map[string]interface{}) bool {
	for _, key := range []string{"address", "telephone", "email", "contactPoint", "geo"} {
		if siteQualityStructuredDataHasValue(data[key]) {
			return true
		}
	}
	return false
}

func siteQualityStructuredDataNumberField(data map[string]interface{}, key string) float64 {
	if data == nil {
		return 0
	}
	switch typed := data[key].(type) {
	case float64:
		return typed
	case int:
		return float64(typed)
	case json.Number:
		value, _ := typed.Float64()
		return value
	case string:
		var value float64
		if _, err := fmt.Sscanf(strings.TrimSpace(typed), "%f", &value); err == nil {
			return value
		}
	}
	return 0
}

func siteQualityStructuredDataHasValue(value interface{}) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(typed) != ""
	case []interface{}:
		return len(typed) > 0
	case map[string]interface{}:
		return len(typed) > 0
	default:
		return true
	}
}

func siteQualityStructuredDataArrayOrSingle(value interface{}) []interface{} {
	switch typed := value.(type) {
	case nil:
		return nil
	case []interface{}:
		return typed
	default:
		return []interface{}{typed}
	}
}

func siteQualityStructuredDataOfferValid(value interface{}) bool {
	for _, offer := range siteQualityStructuredDataArrayOrSingle(value) {
		offerMap, ok := offer.(map[string]interface{})
		if !ok {
			continue
		}
		if siteQualityStructuredDataHasValue(offerMap["price"]) &&
			siteQualityStructuredDataStringField(offerMap, "priceCurrency") != "" &&
			siteQualityStructuredDataStringField(offerMap, "availability") != "" {
			return true
		}
	}
	return false
}

func siteQualityStructuredDataAcceptedAnswerValid(value interface{}) bool {
	for _, answer := range siteQualityStructuredDataArrayOrSingle(value) {
		answerMap, ok := answer.(map[string]interface{})
		if !ok {
			continue
		}
		if !siteQualityStructuredDataValueHasType(answerMap["@type"], "Answer") {
			continue
		}
		if siteQualityStructuredDataStringField(answerMap, "text") != "" {
			return true
		}
	}
	return false
}

func siteQualityStructuredDataValueHasType(value interface{}, expected string) bool {
	expectedKey := siteQualitySchemaTypeKey(expected)
	for _, item := range siteQualityStructuredDataArrayOrSingle(value) {
		switch typed := item.(type) {
		case string:
			if siteQualitySchemaTypeKey(typed) == expectedKey {
				return true
			}
		case []interface{}:
			if siteQualityStructuredDataValueHasType(typed, expected) {
				return true
			}
		}
	}
	return false
}

func siteQualityStructuredDataBreadcrumbItemHasURL(value interface{}) bool {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) != ""
	case map[string]interface{}:
		return siteQualityStructuredDataStringField(typed, "@id") != "" ||
			siteQualityStructuredDataStringField(typed, "url") != ""
	default:
		return false
	}
}

func siteQualityStructuredDataURLMatchesPage(
	rawSchemaURL string,
	targetURL string,
	lighthouseFinalURL string,
) bool {
	pageURL := strings.TrimSpace(lighthouseFinalURL)
	if pageURL == "" {
		pageURL = targetURL
	}
	pageCanonical, err := canonicalizeAbsoluteSiteQualityURL(pageURL)
	if err != nil {
		return true
	}
	schemaURL, err := siteQualityStructuredDataResolveURL(rawSchemaURL, pageURL)
	if err != nil {
		return true
	}
	schemaCanonical, err := canonicalizeAbsoluteSiteQualityURL(schemaURL)
	if err != nil {
		return true
	}
	return strings.EqualFold(schemaCanonical, pageCanonical)
}

func siteQualityStructuredDataURLMatchesOrigin(
	rawSchemaURL string,
	pageURL string,
) bool {
	schemaURL, err := siteQualityStructuredDataResolveURL(rawSchemaURL, pageURL)
	if err != nil {
		return true
	}
	schema, err := url.Parse(schemaURL)
	if err != nil || schema == nil || schema.Host == "" {
		return true
	}
	page, err := url.Parse(strings.TrimSpace(pageURL))
	if err != nil || page == nil || page.Host == "" {
		return true
	}
	return strings.EqualFold(schema.Scheme, page.Scheme) &&
		strings.EqualFold(schema.Hostname(), page.Hostname()) &&
		siteQualityURLPort(schema) == siteQualityURLPort(page)
}

func siteQualityStructuredDataResolveURL(rawValue string, baseValue string) (string, error) {
	value := strings.TrimSpace(rawValue)
	if value == "" {
		return "", fmt.Errorf("schema URL is empty")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", err
	}
	if parsed.IsAbs() {
		return parsed.String(), nil
	}
	base, err := url.Parse(strings.TrimSpace(baseValue))
	if err != nil || base == nil || base.Scheme == "" || base.Host == "" {
		return "", fmt.Errorf("page URL is not absolute")
	}
	return base.ResolveReference(parsed).String(), nil
}

func siteQualityStructuredDataScanFailureIssue(reason string) LighthouseRunnerIssue {
	description := strings.TrimSpace(reason)
	if description == "" {
		description = "The browser-rendered structured data snapshot did not complete."
	}
	return LighthouseRunnerIssue{
		ID:          siteQualityStructuredDataScanFailedAuditID,
		Kind:        "schema",
		RuleVersion: siteQualityAuditRuleVersion,
		Title:       "Rendered structured data audit did not complete",
		Description: description + " This release cannot be treated as schema-clean until the runner verifies the final DOM.",
		Severity:    "critical",
	}
}

func removeSiteQualityRenderedStructuredDataManagedIssues(
	issues []LighthouseRunnerIssue,
) []LighthouseRunnerIssue {
	filtered := issues[:0]
	for _, issue := range issues {
		switch issue.ID {
		case siteQualityStructuredDataScanFailedAuditID,
			siteQualityStructuredDataInvalidJSONLDAuditID,
			siteQualityStructuredDataMissingStructuredDataAuditID,
			siteQualityStructuredDataMissingRequiredTypeAuditID,
			siteQualityStructuredDataDuplicatePrimaryTypeAuditID,
			siteQualityStructuredDataURLMismatchAuditID,
			siteQualityStructuredDataBreadcrumbInvalidAuditID,
			siteQualityStructuredDataProductInvalidAuditID,
			siteQualityStructuredDataFAQInvalidAuditID,
			siteQualityStructuredDataFAQContentMismatchAuditID,
			siteQualityStructuredDataArticleInvalidAuditID,
			siteQualityStructuredDataOrganizationInvalidAuditID,
			siteQualityStructuredDataWebPageInvalidAuditID:
			continue
		default:
			filtered = append(filtered, issue)
		}
	}
	return filtered
}
