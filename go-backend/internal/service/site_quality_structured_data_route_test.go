package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSiteQualityStructuredDataRecognizesResourceBlogRoutes(t *testing.T) {
	require.True(t, siteQualityStructuredDataPathLooksLikeBlogListing("/resources/blog/news"))
	require.True(t, siteQualityStructuredDataPathLooksLikeBlogListing("/resources/blog/wheelsbuild"))
	require.False(t, siteQualityStructuredDataPathLooksLikeBlogListing("/resources/blog/news/release"))
	require.False(t, siteQualityStructuredDataPathLooksLikeBlogListing("/blog/news"))
}

func TestSiteQualityStructuredDataComparablePathRemovesLocaleFromResourceBlogRoute(t *testing.T) {
	require.Equal(
		t,
		"/resources/blog/news/release",
		siteQualityStructuredDataComparablePath("https://example.com/de/resources/blog/news/release"),
	)
}

func TestSiteQualityStructuredDataProductIntentUsesProductsNamespace(t *testing.T) {
	productGroups := siteQualityStructuredDataExpectedGroups(
		"https://example.com/products/carbon-rim",
		"https://example.com/products/carbon-rim",
		&siteQualityRenderedStructuredDataAudit{
			Page: siteQualityStructuredDataPage{
				CanonicalURL: "https://example.com/products/carbon-rim",
			},
		},
		siteQualityStructuredDataPageIntent{},
	)
	require.Len(t, productGroups, 1)
	require.Equal(t, "Product or ProductGroup", productGroups[0].Label)

	categoryGroups := siteQualityStructuredDataExpectedGroups(
		"https://example.com/shop/road-wheels",
		"https://example.com/shop/road-wheels",
		&siteQualityRenderedStructuredDataAudit{
			Page: siteQualityStructuredDataPage{
				CanonicalURL: "https://example.com/shop/road-wheels",
			},
		},
		siteQualityStructuredDataPageIntent{},
	)
	require.Empty(t, categoryGroups)
}
