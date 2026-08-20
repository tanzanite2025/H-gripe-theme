package refundreturn

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeRejectsProtocolRelativeURLs(t *testing.T) {
	policy := DefaultPolicy()
	policy.ContactURL = "//example.com/contact"
	_, err := Normalize(policy)
	require.Error(t, err)
	require.Contains(t, err.Error(), "contact URL")

	policy = DefaultPolicy()
	policy.Sections[0].Image = &Image{URL: "//example.com/return.jpg"}
	_, err = Normalize(policy)
	require.Error(t, err)
	require.Contains(t, err.Error(), "image URL")
}

func TestNormalizeAllowsFirstPartyAndAbsoluteURLs(t *testing.T) {
	policy := DefaultPolicy()
	policy.ContactURL = "mailto:support@example.com"
	policy.Sections[0].Image = &Image{URL: "/uploads/policy/return.jpg"}
	_, err := Normalize(policy)
	require.NoError(t, err)

	policy.ContactURL = "https://example.com/contact"
	policy.Sections[0].Image.URL = "https://example.com/return.jpg"
	normalized, err := Normalize(policy)
	require.NoError(t, err)
	require.False(t, strings.HasPrefix(normalized.Sections[0].Image.URL, "//"))
}

func TestNormalizeRemovesEmptyImageBeforeSectionContentValidation(t *testing.T) {
	policy := Policy{
		Title: "Returns",
		Sections: []Section{{
			ID:    "empty-image",
			Title: "Empty image",
			Image: &Image{
				Alt:     "No actual image",
				Caption: "This object has no URL.",
			},
		}},
	}

	_, err := Normalize(policy)
	require.Error(t, err)
	require.Contains(t, err.Error(), "must contain text, bullets, or an image")
}

func TestNormalizeRejectsFrontendAnchorCollisions(t *testing.T) {
	policy := Policy{
		Title: "Returns",
		Sections: []Section{
			{ID: "Returns!", Title: "First", Body: "First body"},
			{ID: "returns", Title: "Second", Body: "Second body"},
		},
	}

	_, err := Normalize(policy)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicated page anchor")
}

func TestNormalizeRejectsFrontendFallbackAnchorCollisions(t *testing.T) {
	policy := Policy{
		Title: "Returns",
		Sections: []Section{
			{ID: "!!!", Title: "First", Body: "First body"},
			{ID: "section-1", Title: "Second", Body: "Second body"},
		},
	}

	_, err := Normalize(policy)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicated page anchor")
}
