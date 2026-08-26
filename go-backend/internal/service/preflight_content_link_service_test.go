package service

import (
	"strings"
	"testing"

	preflightdomain "commerce-platform/internal/domain/preflight"
)

func TestContentLinkDetectionFlagsGenericAnchorsOnly(t *testing.T) {
	svc := &PreflightContentLinkService{}
	detections, err := svc.detectContentLinks(`
		<html>
			<body>
				<h2>Wheel guides</h2>
				<a href="/blog">Read more</a>
				<button>Read more</button>
				<a>Learn more</a>
				<a href="/guides">Wheel setup guide</a>
			</body>
		</html>
	`, "https://example.test/", "https://example.test/", nil)
	if err != nil {
		t.Fatalf("detect content links: %v", err)
	}
	if len(detections) != 1 {
		t.Fatalf("expected one generic anchor finding, got %d", len(detections))
	}
	if detections[0].LinkText != "Read more" {
		t.Fatalf("unexpected link text %q", detections[0].LinkText)
	}
	if !strings.Contains(detections[0].SuggestedText, "Wheel guides") {
		t.Fatalf("expected heading-based suggestion, got %q", detections[0].SuggestedText)
	}
}

func TestContentLinkDetectionMatchesOfficialLinkTextBoundaries(t *testing.T) {
	svc := &PreflightContentLinkService{}
	detections, err := svc.detectContentLinks(`
		<html>
			<body>
				<a href="/guides">Click this</a>
				<a href="/guides" rel="nofollow">Go</a>
				<a href="#details">Here</a>
				<a href="mailto:test@example.test">More</a>
				<a href="/guides">View more</a>
			</body>
		</html>
	`, "https://example.test/", "https://example.test/", nil)
	if err != nil {
		t.Fatalf("detect content links: %v", err)
	}
	if len(detections) != 1 {
		t.Fatalf("expected one official link-text finding, got %d", len(detections))
	}
	if detections[0].LinkText != "Click this" {
		t.Fatalf("unexpected link text %q", detections[0].LinkText)
	}
	if detections[0].RuleID != preflightdomain.ContentLinkRuleID {
		t.Fatalf("unexpected rule ID %q", detections[0].RuleID)
	}
	if detections[0].ProviderAuditID != preflightdomain.ContentLinkProviderAuditID {
		t.Fatalf("unexpected provider audit ID %q", detections[0].ProviderAuditID)
	}
	if !strings.HasPrefix(
		detections[0].IssueKey,
		"content-link:"+preflightdomain.ContentLinkRuleID+":",
	) {
		t.Fatalf("content link issue key is not rule-scoped: %q", detections[0].IssueKey)
	}
}

func TestReplaceFirstContentLinkSuggestionMatchesSourceAnchor(t *testing.T) {
	issue := &preflightdomain.ContentLinkIssue{
		TargetURL:     "https://example.test/blog/example",
		LinkURL:       "https://example.test/guides",
		LinkText:      "Learn more",
		SuggestedText: "Learn about wheel setup guides",
	}
	next, changed, err := replaceFirstContentLinkSuggestion(
		`<p><a href="/guides"><span>Learn more</span></a></p><p><a href="/other">Learn more</a></p>`,
		issue,
	)
	if err != nil {
		t.Fatalf("replace suggestion: %v", err)
	}
	if !changed {
		t.Fatal("expected content to change")
	}
	if !strings.Contains(next, ">Learn about wheel setup guides</a>") {
		t.Fatalf("expected suggested text in content, got %s", next)
	}
	if !strings.Contains(next, `href="/other">Learn more</a>`) {
		t.Fatalf("expected non-matching link to remain unchanged, got %s", next)
	}
}
