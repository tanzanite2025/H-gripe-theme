package ugc

import (
	"errors"
	"strings"
	"testing"
)

func TestPlainTextRemovesDangerousMarkupAndNormalizesWhitespace(t *testing.T) {
	got, err := PlainText(`<p onclick="alert(1)">Great   product</p><script>alert("x")</script><a href="javascript:alert(1)">bad link</a>`, 500)
	if err != nil {
		t.Fatalf("PlainText() error = %v", err)
	}
	if got != "Great product bad link" {
		t.Fatalf("PlainText() = %q", got)
	}
	for _, forbidden := range []string{"alert", "script", "onclick", "<"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("PlainText() retained %q in %q", forbidden, got)
		}
	}
}

func TestPlainTextRejectsOverlongValues(t *testing.T) {
	_, err := PlainText("12345", 4)
	if !errors.Is(err, ErrTextTooLong) {
		t.Fatalf("PlainText() error = %v, want ErrTextTooLong", err)
	}
}
