package websitecontent

import (
	"strings"
	"testing"
)

func TestNormalizeWebsiteBodyPreservesLegacyPlainTextLayout(t *testing.T) {
	got, err := NormalizeWebsiteBody("第一行\n第二行\n\n新的段落")
	if err != nil {
		t.Fatalf("NormalizeWebsiteNameBody() error = %v", err)
	}

	want := "<p>第一行<br>第二行</p><p>新的段落</p>"
	if got != want {
		t.Fatalf("unexpected normalized body: %s", got)
	}
}

func TestNormalizeWebsiteBodySanitizesRichText(t *testing.T) {
	got, err := NormalizeWebsiteBody(`<div><strong>重点</strong><script>alert(1)</script></div><ul><li>项目</li></ul><a href="javascript:alert(1)">危险链接</a>`)
	if err != nil {
		t.Fatalf("NormalizeWebsiteNameBody() error = %v", err)
	}

	for _, forbidden := range []string{"script", "javascript"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("unsafe content survived: %s", got)
		}
	}
	for _, expected := range []string{"<p><strong>重点</strong></p>", "<ul><li>项目</li></ul>", "危险链接"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in normalized body: %s", expected, got)
		}
	}
}
