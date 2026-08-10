package admin

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRejectArticleSEORequestFieldsBlocksContentSEOFields(t *testing.T) {
	for _, field := range articleSEORequestFields {
		requestBody := map[string]json.RawMessage{
			"title": json.RawMessage(`"Post"`),
			field:   json.RawMessage(`"Managed by SEO"`),
		}

		blockedField, blocked := rejectArticleSEORequestFields(requestBody)

		require.True(t, blocked)
		require.Equal(t, field, blockedField)
	}
}

func TestRejectArticleSEORequestFieldsAllowsContentFields(t *testing.T) {
	requestBody := map[string]json.RawMessage{
		"title":    json.RawMessage(`"Post"`),
		"slug":     json.RawMessage(`"post"`),
		"content":  json.RawMessage(`"Body"`),
		"excerpt":  json.RawMessage(`"Summary"`),
		"status":   json.RawMessage(`"draft"`),
		"locale":   json.RawMessage(`"en"`),
		"tags":     json.RawMessage(`"news"`),
		"category": json.RawMessage(`"blog"`),
	}

	blockedField, blocked := rejectArticleSEORequestFields(requestBody)

	require.False(t, blocked)
	require.Empty(t, blockedField)
}
