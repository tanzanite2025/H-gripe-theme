package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSEOResourceServiceValidatesCanonicalURLs(t *testing.T) {
	service := &SEOResourceService{}
	service.ConfigureCanonicalBaseURL("https://store.example.test")

	require.NoError(t, service.validateCanonicalURL("https://store.example.test/blog/release"))
	require.ErrorIs(t, service.validateCanonicalURL("http://store.example.test/blog/release"), ErrInvalidSEOCanonicalURL)
	require.ErrorIs(t, service.validateCanonicalURL("https://other.example.test/blog/release"), ErrInvalidSEOCanonicalURL)
	require.ErrorIs(t, service.validateCanonicalURL("https://store.example.test/blog/release?ref=home"), ErrInvalidSEOCanonicalURL)
	require.NoError(t, service.validateCanonicalURL(""))
}
