package seo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildProductRoute(t *testing.T) {
	require.Equal(t, "/shop/demo-wheel", BuildProductRoute("en", "demo-wheel").Path)
	require.Equal(t, "/zh_cn/shop/demo-wheel", BuildProductRoute("zh-CN", "demo-wheel").Path)
	require.Equal(t, "/products/demo-wheel", BuildLegacyProductRoute("en", "demo-wheel").Path)
}

func TestBuildStaticRoute(t *testing.T) {
	require.Equal(t, "/", BuildStaticRoute("en", "/").Path)
	require.Equal(t, "/de/company/about", BuildStaticRoute("de", "/company/about").Path)
}

func TestBuildArticleRouteUsesTheStorefrontCategory(t *testing.T) {
	require.Equal(t, "/blog/news/release", BuildArticleRoute("en", "release", "featured,news").Path)
	require.Equal(t, "/blog/news/release", BuildArticleRoute("en", "release", "wheelsbuild,news").Path)
	require.Equal(t, "/de/blog/wheelsbuild/build-guide", BuildArticleRoute("de", "build-guide", "wheelsbuild").Path)
	require.Equal(t, "/blog/general-post", BuildArticleRoute("en", "general-post", "").Path)
}
