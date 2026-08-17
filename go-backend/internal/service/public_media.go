package service

import (
	"encoding/json"
	"strings"
)

func CanonicalPublicMediaURL(resolver PublicMediaURLResolver, value string) string {
	value = strings.TrimSpace(value)
	if resolver == nil || value == "" {
		return value
	}
	return resolver.CanonicalPublicMediaURL(value)
}

func canonicalPublicMediaURL(resolver PublicMediaURLResolver, value string) string {
	return CanonicalPublicMediaURL(resolver, value)
}

// CanonicalPublicMediaURLsJSON canonicalizes a JSON array of media references.
// Malformed public attachment payloads are hidden rather than echoed back.
func CanonicalPublicMediaURLsJSON(resolver PublicMediaURLResolver, raw string) string {
	if resolver == nil || strings.TrimSpace(raw) == "" {
		return raw
	}

	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return "[]"
	}
	for index := range values {
		values[index] = CanonicalPublicMediaURL(resolver, values[index])
	}

	encoded, err := json.Marshal(values)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}
