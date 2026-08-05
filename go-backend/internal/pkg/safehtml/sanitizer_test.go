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

func TestSanitizePreservesSafeProductDetailMedia(t *testing.T) {
	got, err := Sanitize(`<figure><img src="/uploads/detail.webp" alt="Angle" onerror="alert(1)"><figcaption>Angle view</figcaption><video src="https://cdn.example.com/demo.webm" poster="/uploads/poster.webp" autoplay onplay="alert(1)"></video></figure>`)
	if err != nil {
		t.Fatalf("Sanitize() error = %v", err)
	}
	for _, expected := range []string{
		`<figure>`,
		`<img src="/uploads/detail.webp" alt="Angle" loading="lazy">`,
		`<figcaption>Angle view</figcaption>`,
		`<video src="https://cdn.example.com/demo.webm" poster="/uploads/poster.webp" controls preload="metadata"></video>`,
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in sanitized html: %s", expected, got)
		}
	}
	for _, forbidden := range []string{"onerror", "onplay", "autoplay"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("unsafe media attribute survived: %s", got)
		}
	}
}

func TestSanitizeDropsUnsafeProductDetailMediaSources(t *testing.T) {
	got, err := Sanitize(`<img src="data:image/svg+xml;base64,aaa"><video src="//evil.example/demo.webm"></video><img src="uploads/not-rooted.webp">`)
	if err != nil {
		t.Fatalf("Sanitize() error = %v", err)
	}
	for _, forbidden := range []string{"<img", "<video", "data:", "evil.example", "not-rooted"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("unsafe media survived: %s", got)
		}
	}
}
