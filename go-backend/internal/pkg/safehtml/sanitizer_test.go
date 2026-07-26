package safehtml

import (
	"strings"
	"testing"
)

func TestSanitizeRemovesDangerousHTML(t *testing.T) {
	got, err := Sanitize(`<p onclick="alert(1)">Hello<script>alert(1)</script><img src=x onerror=alert(1)></p>`)
	if err != nil {
		t.Fatalf("Sanitize() error = %v", err)
	}
	for _, forbidden := range []string{"onclick", "script", "onerror", "<img"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("sanitized html still contains %q: %s", forbidden, got)
		}
	}
	if got != "<p>Hello</p>" {
		t.Fatalf("unexpected sanitized html: %s", got)
	}
}

func TestSanitizeFiltersUnsafeLinks(t *testing.T) {
	got, err := Sanitize(`<a href="javascript:alert(1)">bad</a><a href="/support">ok</a>`)
	if err != nil {
		t.Fatalf("Sanitize() error = %v", err)
	}
	if strings.Contains(got, "javascript") {
		t.Fatalf("unsafe link survived: %s", got)
	}
	if !strings.Contains(got, `href="/support"`) {
		t.Fatalf("safe relative link was not preserved: %s", got)
	}
}
