package middleware

import (
	"strings"
	"testing"
)

func TestSanitizeRawQueryRedactsSensitiveValues(t *testing.T) {
	got := sanitizeRawQuery("page=2&access_token=secret-token&redirect=/account&signature=real-signature&code=oauth-code")

	if strings.Contains(got, "secret-token") || strings.Contains(got, "real-signature") || strings.Contains(got, "oauth-code") {
		t.Fatalf("sanitizeRawQuery leaked sensitive values: %s", got)
	}
	for _, want := range []string{
		"page=2",
		"redirect=%2Faccount",
		"access_token=%5BREDACTED%5D",
		"signature=%5BREDACTED%5D",
		"code=%5BREDACTED%5D",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("sanitizeRawQuery = %s, want to contain %s", got, want)
		}
	}
}

func TestSanitizeRawQueryDoesNotLogInvalidQuery(t *testing.T) {
	if got := sanitizeRawQuery("%zz=secret-token"); got != "[invalid_query]" {
		t.Fatalf("sanitizeRawQuery invalid query = %s, want [invalid_query]", got)
	}
}
