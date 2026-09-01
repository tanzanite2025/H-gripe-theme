package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"

	"golang.org/x/net/html"
)

const (
	siteQualityHeadingMissingH1AuditID  = "site-heading-missing-h1"
	siteQualityHeadingMultipleH1AuditID = "site-heading-multiple-h1"
	siteQualityHeadingScanFailedAuditID = "site-heading-rendered-scan-failed"
	siteQualityHeadingHTMLMaxBytes      = 2 << 20
	siteQualityHeadingFetchTimeout      = 8 * time.Second
)

type siteQualityHeadingNode struct {
	Level    int
	Text     string
	Snippet  string
	Selector string
}

func (s *LighthouseRunnerService) siteQualityHeadingOutlineIssues(
	ctx context.Context,
	targetURL string,
) []LighthouseRunnerIssue {
	if s == nil {
		return nil
	}
	headings, err := fetchSiteQualityHeadingOutline(ctx, targetURL)
	if err != nil {
		return nil
	}
	return siteQualityHeadingOutlineIssues(headings)
}

func fetchSiteQualityHeadingOutline(
	ctx context.Context,
	targetURL string,
) ([]siteQualityHeadingNode, error) {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(targetURL))
	if err != nil || parsed == nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, fmt.Errorf("invalid heading outline target URL")
	}

	requestCtx, cancel := context.WithTimeout(ctx, siteQualityHeadingFetchTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "text/html,application/xhtml+xml")
	request.Header.Set("User-Agent", "TANZANITE-SiteQualityHeadings/1.0")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("heading outline fetch returned HTTP %d", response.StatusCode)
	}

	root, err := html.Parse(io.LimitReader(response.Body, siteQualityHeadingHTMLMaxBytes))
	if err != nil {
		return nil, err
	}

	var headings []siteQualityHeadingNode
	walkSiteQualityHeadingHTML(root, nil, &headings, false)
	return headings, nil
}

func siteQualityHeadingOutlineIssues(headings []siteQualityHeadingNode) []LighthouseRunnerIssue {
	h1Headings := make([]siteQualityHeadingNode, 0)
	for _, heading := range headings {
		if heading.Level == 1 {
			h1Headings = append(h1Headings, heading)
		}
	}

	issues := make([]LighthouseRunnerIssue, 0, 2)
	if len(h1Headings) == 0 {
		issues = append(issues, LighthouseRunnerIssue{
			ID:          siteQualityHeadingMissingH1AuditID,
			Kind:        "headings",
			RuleVersion: siteQualityAuditRuleVersion,
			Title:       "Page is missing a primary H1",
			Description: "Every indexable page should expose one H1 before H2/H3 section headings so assistive technology can understand the page topic.",
			Severity:    "critical",
			Headings:    siteQualityHeadingEvidenceFromNodes(siteQualityHeadingContext(headings)),
		})
	}
	if len(h1Headings) > 1 {
		issues = append(issues, LighthouseRunnerIssue{
			ID:          siteQualityHeadingMultipleH1AuditID,
			Kind:        "headings",
			RuleVersion: siteQualityAuditRuleVersion,
			Title:       "Page has multiple H1 headings",
			Description: "Use one primary H1 for the page topic, then H2/H3 headings for sections and cards.",
			Severity:    "medium",
			Headings:    siteQualityHeadingEvidenceFromNodes(h1Headings),
		})
	}
	issues = append(issues, siteQualityHeadingOrderIssues(headings)...)
	return issues
}

func siteQualityRenderedHeadingAuditIssues(
	targetURL string,
	lighthouseFinalURL string,
	audit *siteQualityRenderedHeadingAudit,
) []LighthouseRunnerIssue {
	if audit == nil {
		return []LighthouseRunnerIssue{
			siteQualityHeadingScanFailureIssue("The runner response did not include a browser-rendered heading snapshot."),
		}
	}
	if strings.TrimSpace(audit.Status) != "complete" {
		reason := strings.TrimSpace(audit.Error)
		if reason == "" {
			reason = "The browser-rendered heading snapshot did not complete."
		}
		return []LighthouseRunnerIssue{siteQualityHeadingScanFailureIssue(reason)}
	}
	if !siteQualityRenderedHeadingFinalURLTrusted(targetURL, lighthouseFinalURL, audit.FinalURL) {
		return []LighthouseRunnerIssue{
			siteQualityHeadingScanFailureIssue("The browser-rendered heading snapshot reached an unexpected final URL."),
		}
	}
	headings := siteQualityHeadingNodesFromEvidence(audit.Headings)
	return siteQualityHeadingOutlineIssues(headings)
}

func siteQualityRenderedHeadingFinalURLTrusted(
	targetURL string,
	lighthouseFinalURL string,
	renderedFinalURL string,
) bool {
	rendered, err := url.Parse(strings.TrimSpace(renderedFinalURL))
	if err != nil || rendered == nil || rendered.Scheme == "" || rendered.Host == "" {
		return false
	}
	if !sameSiteQualityOrigin(rendered, targetURL) {
		return false
	}
	lighthouseFinal, err := url.Parse(strings.TrimSpace(lighthouseFinalURL))
	if err == nil && lighthouseFinal != nil && lighthouseFinal.Scheme != "" && lighthouseFinal.Host != "" {
		return sameSiteQualityOrigin(rendered, lighthouseFinal.String())
	}
	return true
}

func siteQualityHeadingScanFailureIssue(reason string) LighthouseRunnerIssue {
	description := strings.TrimSpace(reason)
	if description == "" {
		description = "The browser-rendered heading snapshot did not complete."
	}
	return LighthouseRunnerIssue{
		ID:          siteQualityHeadingScanFailedAuditID,
		Kind:        "headings",
		RuleVersion: siteQualityAuditRuleVersion,
		Title:       "Rendered heading audit did not complete",
		Description: description + " This release cannot be treated as heading-clean until the runner verifies the final DOM.",
		Severity:    "critical",
	}
}

func siteQualityHeadingOrderIssues(headings []siteQualityHeadingNode) []LighthouseRunnerIssue {
	offenders := make([]sitequalitydomain.SiteQualityHeadingEvidence, 0)
	var previous siteQualityHeadingNode
	seen := make(map[string]struct{})
	for _, heading := range headings {
		if heading.Level < 1 || heading.Level > 6 {
			continue
		}
		if previous.Level == 0 {
			if heading.Level != 1 {
				offenders = appendHeadingOrderEvidence(
					offenders,
					seen,
					heading,
					fmt.Sprintf("First visible heading is H%d. The first visible heading should be H1.", heading.Level),
				)
			}
			previous = heading
			continue
		}
		if heading.Level > previous.Level+1 {
			offenders = appendHeadingOrderEvidence(
				offenders,
				seen,
				previous,
				fmt.Sprintf("Previous visible heading before skipped level H%d.", heading.Level-1),
			)
			offenders = appendHeadingOrderEvidence(
				offenders,
				seen,
				heading,
				fmt.Sprintf("H%d follows H%d. Insert an H%d or lower this heading level.", heading.Level, previous.Level, previous.Level+1),
			)
		}
		previous = heading
	}
	if len(offenders) == 0 {
		return nil
	}
	return []LighthouseRunnerIssue{
		{
			ID:          "heading-order",
			Kind:        "headings",
			RuleVersion: siteQualityAuditRuleVersion,
			Title:       "Heading levels are not sequential",
			Description: "Visible headings must start with H1 and should not skip levels, such as H1 directly followed by H3.",
			Severity:    "high",
			Headings:    offenders,
		},
	}
}

func appendHeadingOrderEvidence(
	evidence []sitequalitydomain.SiteQualityHeadingEvidence,
	seen map[string]struct{},
	node siteQualityHeadingNode,
	explanation string,
) []sitequalitydomain.SiteQualityHeadingEvidence {
	key := strings.Join([]string{
		strconv.Itoa(node.Level),
		node.Selector,
		node.Snippet,
		node.Text,
		explanation,
	}, "\x00")
	if _, exists := seen[key]; exists {
		return evidence
	}
	seen[key] = struct{}{}
	evidence = append(evidence, sitequalitydomain.SiteQualityHeadingEvidence{
		Level:       node.Level,
		Text:        node.Text,
		Snippet:     node.Snippet,
		Selector:    node.Selector,
		Explanation: explanation,
	})
	return evidence
}

func siteQualityHeadingNodesFromEvidence(
	evidence []sitequalitydomain.SiteQualityHeadingEvidence,
) []siteQualityHeadingNode {
	headings := make([]siteQualityHeadingNode, 0, len(evidence))
	for _, heading := range evidence {
		if heading.Level < 1 || heading.Level > 6 {
			continue
		}
		text := siteQualityNormalizeHeadingText(heading.Text)
		if text == "" {
			text = siteQualityNormalizeHeadingText(heading.Snippet)
		}
		if text == "" {
			continue
		}
		headings = append(headings, siteQualityHeadingNode{
			Level:    heading.Level,
			Text:     text,
			Snippet:  strings.TrimSpace(heading.Snippet),
			Selector: strings.TrimSpace(heading.Selector),
		})
	}
	return headings
}

func removeSiteQualityRenderedHeadingManagedIssues(
	issues []LighthouseRunnerIssue,
) []LighthouseRunnerIssue {
	filtered := issues[:0]
	for _, issue := range issues {
		switch issue.ID {
		case "heading-order",
			siteQualityHeadingMissingH1AuditID,
			siteQualityHeadingMultipleH1AuditID,
			siteQualityHeadingScanFailedAuditID:
			continue
		default:
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

func siteQualityHeadingContext(headings []siteQualityHeadingNode) []siteQualityHeadingNode {
	if len(headings) <= 8 {
		return headings
	}
	return headings[:8]
}

func siteQualityHeadingEvidenceFromNodes(
	nodes []siteQualityHeadingNode,
) []sitequalitydomain.SiteQualityHeadingEvidence {
	evidence := make([]sitequalitydomain.SiteQualityHeadingEvidence, 0, len(nodes))
	for _, node := range nodes {
		evidence = append(evidence, sitequalitydomain.SiteQualityHeadingEvidence{
			Level:    node.Level,
			Text:     node.Text,
			Snippet:  node.Snippet,
			Selector: node.Selector,
		})
	}
	return evidence
}

func walkSiteQualityHeadingHTML(
	node *html.Node,
	path []string,
	headings *[]siteQualityHeadingNode,
	hidden bool,
) {
	if node == nil || hidden {
		return
	}
	if node.Type == html.ElementNode {
		tag := strings.ToLower(node.Data)
		if tag == "script" || tag == "style" || tag == "noscript" || tag == "template" || siteQualityHeadingNodeHidden(node) {
			return
		}
		nextPath := append(path, siteQualityNodeSelectorPart(node))
		if level := siteQualityHeadingTagLevel(tag); level > 0 {
			text := siteQualityNormalizeHeadingText(siteQualityNodeText(node))
			if text != "" {
				*headings = append(*headings, siteQualityHeadingNode{
					Level:    level,
					Text:     text,
					Snippet:  siteQualityHeadingSnippet(tag, node, text),
					Selector: strings.Join(nextPath, " > "),
				})
			}
		}
		if tag == "details" && !siteQualityHeadingNodeHasAttr(node, "open") {
			summary := siteQualityHeadingSummaryChild(node)
			if summary != nil {
				walkSiteQualityHeadingHTML(summary, nextPath, headings, false)
			}
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if child == summary {
					continue
				}
				walkSiteQualityHeadingHTML(child, nextPath, headings, true)
			}
			return
		}
		path = nextPath
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walkSiteQualityHeadingHTML(child, path, headings, false)
	}
}

func siteQualityHeadingNodeHasAttr(node *html.Node, key string) bool {
	if node == nil {
		return false
	}
	for _, attr := range node.Attr {
		if strings.EqualFold(strings.TrimSpace(attr.Key), key) {
			return true
		}
	}
	return false
}

func siteQualityHeadingSummaryChild(node *html.Node) *html.Node {
	if node == nil {
		return nil
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && strings.EqualFold(child.Data, "summary") {
			return child
		}
	}
	return nil
}

func siteQualityHeadingNodeHidden(node *html.Node) bool {
	for _, attr := range node.Attr {
		key := strings.ToLower(strings.TrimSpace(attr.Key))
		value := strings.ToLower(strings.TrimSpace(attr.Val))
		if key == "hidden" || key == "inert" || (key == "aria-hidden" && value == "true") {
			return true
		}
		if key == "style" && (strings.Contains(value, "display:none") ||
			strings.Contains(value, "display: none") ||
			strings.Contains(value, "visibility:hidden") ||
			strings.Contains(value, "visibility: hidden") ||
			strings.Contains(value, "visibility:collapse") ||
			strings.Contains(value, "visibility: collapse") ||
			strings.Contains(value, "content-visibility:hidden") ||
			strings.Contains(value, "content-visibility: hidden")) {
			return true
		}
	}
	return false
}

func siteQualityHeadingTagLevel(tag string) int {
	if len(tag) != 2 || tag[0] != 'h' {
		return 0
	}
	level := int(tag[1] - '0')
	if level < 1 || level > 6 {
		return 0
	}
	return level
}

func siteQualityNodeText(node *html.Node) string {
	if node == nil {
		return ""
	}
	if node.Type == html.TextNode {
		return node.Data
	}
	var builder strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		builder.WriteString(" ")
		builder.WriteString(siteQualityNodeText(child))
	}
	return builder.String()
}

func siteQualityNormalizeHeadingText(value string) string {
	normalized := strings.Join(strings.Fields(value), " ")
	if len(normalized) > 180 {
		return normalized[:177] + "..."
	}
	return normalized
}

func siteQualityHeadingSnippet(tag string, node *html.Node, text string) string {
	attrs := make([]string, 0, 2)
	for _, attr := range node.Attr {
		key := strings.ToLower(strings.TrimSpace(attr.Key))
		if key != "id" && key != "class" {
			continue
		}
		value := strings.TrimSpace(attr.Val)
		if value == "" {
			continue
		}
		attrs = append(attrs, fmt.Sprintf(`%s="%s"`, key, value))
	}
	openTag := "<" + tag
	if len(attrs) > 0 {
		openTag += " " + strings.Join(attrs, " ")
	}
	openTag += ">"
	return openTag + text + "</" + tag + ">"
}

func siteQualityNodeSelectorPart(node *html.Node) string {
	if node == nil || node.Type != html.ElementNode {
		return ""
	}
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, "id") && strings.TrimSpace(attr.Val) != "" {
			return strings.ToLower(node.Data) + "#" + siteQualityCSSIdentifier(attr.Val)
		}
	}
	return strings.ToLower(node.Data) + ":nth-of-type(" + strconv.Itoa(siteQualityNodeTypeIndex(node)) + ")"
}

func siteQualityNodeTypeIndex(node *html.Node) int {
	if node == nil || node.Parent == nil {
		return 1
	}
	index := 0
	for sibling := node.Parent.FirstChild; sibling != nil; sibling = sibling.NextSibling {
		if sibling.Type == html.ElementNode && strings.EqualFold(sibling.Data, node.Data) {
			index++
		}
		if sibling == node {
			if index == 0 {
				return 1
			}
			return index
		}
	}
	return 1
}

func siteQualityCSSIdentifier(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, value)
	if value == "" {
		return "unknown"
	}
	return value
}
