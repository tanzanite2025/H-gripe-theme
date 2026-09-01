package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSiteQualityStructuredDataNormalizesSchemaTypes(t *testing.T) {
	require.Equal(t, []string{"Product", "Thing"}, siteQualityNormalizeSchemaTypes([]string{
		"https://schema.org/Product/",
		"product",
		"https://schema.org/Thing#",
		"THING",
	}))
}

func TestSiteQualityStructuredDataNormalizesAbsoluteURLs(t *testing.T) {
	normalized := siteQualityStructuredDataNormalizeURL("HTTPS://Example.com:443/products/carbon-rim/?utm_source=nav#hero")
	require.Equal(t, "https://example.com/products/carbon-rim", normalized)

	node := siteQualityStructuredDataNodeView{
		siteQualityStructuredDataNode: siteQualityStructuredDataNode{
			URL: "HTTPS://Example.com:443/products/carbon-rim/?utm_source=nav#hero",
		},
	}
	require.Equal(t, "https://example.com/products/carbon-rim", siteQualityStructuredDataPrimaryEntityURL(node))
	require.Equal(t, "https://example.com/products/carbon-rim", siteQualityStructuredDataURLValue("HTTPS://Example.com:443/products/carbon-rim/?utm_source=nav#hero"))
}

func TestSiteQualityStructuredDataURLMatchesCanonicalizedPage(t *testing.T) {
	require.True(t, siteQualityStructuredDataURLMatchesPage(
		"HTTPS://Example.com:443/products/carbon-rim/?utm_source=nav#hero",
		"https://example.com/products/carbon-rim/",
		"https://example.com/products/carbon-rim",
	))
	require.True(t, siteQualityStructuredDataURLMatchesOrigin(
		"HTTPS://Example.com:443/products/carbon-rim/?utm_source=nav#hero",
		"https://example.com/products/carbon-rim/",
	))
	require.False(t, siteQualityStructuredDataURLMatchesPage(
		"https://example.com/products/other-rim",
		"https://example.com/products/carbon-rim",
		"https://example.com/products/carbon-rim",
	))
}
