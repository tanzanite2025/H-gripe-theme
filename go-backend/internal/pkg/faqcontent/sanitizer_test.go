package faqcontent

import "testing"

func TestSanitizeAnswerAcceptsRichTextFragment(t *testing.T) {
	input := `<p><strong>load_kg</strong> 通常按单条外胎的目录承载能力理解。</p>`

	got, err := SanitizeAnswer(input)
	if err != nil {
		t.Fatalf("SanitizeAnswer() error = %v", err)
	}
	if got != input {
		t.Fatalf("SanitizeAnswer() = %q, want %q", got, input)
	}
	if !HasVisibleText(got) {
		t.Fatal("HasVisibleText() = false for visible FAQ answer text")
	}
}

func TestSanitizeAnswerStripsUnsupportedMarkup(t *testing.T) {
	got, err := SanitizeAnswer(`<p>安全内容</p><script>alert('xss')</script><img src="bad">`)
	if err != nil {
		t.Fatalf("SanitizeAnswer() error = %v", err)
	}
	if got != `<p>安全内容</p>` {
		t.Fatalf("SanitizeAnswer() = %q, want sanitized content", got)
	}
}
