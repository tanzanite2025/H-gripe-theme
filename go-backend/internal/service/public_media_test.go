package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanonicalPublicMediaURLsJSONPreservesShapeAndThirdPartyURLs(t *testing.T) {
	resolver := NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30)

	raw := `["http://media.internal:8080/uploads/warranty/photo.jpg","https://cdn.example.test/photo.jpg"]`
	require.Equal(
		t,
		`["https://shop.example.test/uploads/warranty/photo.jpg","https://cdn.example.test/photo.jpg"]`,
		CanonicalPublicMediaURLsJSON(resolver, raw),
	)
	malformed := `{"url":"http://media.internal:8080/uploads/photo.jpg"}`
	require.Equal(t, "[]", CanonicalPublicMediaURLsJSON(resolver, malformed))
}
