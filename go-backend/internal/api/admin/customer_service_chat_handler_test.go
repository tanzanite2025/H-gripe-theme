package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeAdminCustomerServiceMessageTypeAllowsVideo(t *testing.T) {
	require.Equal(t, "video", normalizeAdminCustomerServiceMessageType(" video "))
	require.Equal(t, "text", normalizeAdminCustomerServiceMessageType("unsupported"))
}

func TestMarshalAdminCustomerServiceMessageMetadataIgnoresNull(t *testing.T) {
	payload, err := marshalAdminCustomerServiceMessageMetadata(map[string]any{
		"url":       "https://example.test/products/demo",
		"source":    "admin",
		"thumbnail": nil,
	})
	require.NoError(t, err)
	assert.Contains(t, payload, `"url":"https://example.test/products/demo"`)
	assert.Contains(t, payload, `"source":"admin"`)

	empty, err := marshalAdminCustomerServiceMessageMetadata(nil)
	require.NoError(t, err)
	assert.Empty(t, empty)
}
