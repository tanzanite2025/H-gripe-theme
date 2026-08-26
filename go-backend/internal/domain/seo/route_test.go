package seo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildProductRoute(t *testing.T) {
	require.Equal(t, "/products/demo-wheel", BuildProductRoute("en", "demo-wheel").Path)
	require.Equal(t, "/zh_cn/products/demo-wheel", BuildProductRoute("zh-CN", "demo-wheel").Path)
}

func TestBuildCategoryRoute(t *testing.T) {
	require.Equal(t, "/shop/road-wheels/climbing", BuildCategoryRoute("en", "road-wheels", "climbing").Path)
	require.Equal(t, "/zh_cn/shop/road-wheels/climbing", BuildCategoryRoute("zh-CN", "road-wheels", "climbing").Path)
	require.Equal(t, "", BuildCategoryRoute("en").Path)
}

func TestBuildStaticRoute(t *testing.T) {
	require.Equal(t, "/", BuildStaticRoute("en", "/").Path)
	require.Equal(t, "/de/company/about", BuildStaticRoute("de", "/company/about").Path)
}

func TestBuildArticleRouteUsesTheStorefrontCategory(t *testing.T) {
	require.Equal(t, "/resources/blog/news/release", BuildArticleRoute("en", "release", "featured,news").Path)
	require.Equal(t, "/resources/blog/news/release", BuildArticleRoute("en", "release", "wheelsbuild,news").Path)
	require.Equal(t, "/de/resources/blog/wheelsbuild/build-guide", BuildArticleRoute("de", "build-guide", "wheelsbuild").Path)
	require.Equal(t, "/resources/blog/general-post", BuildArticleRoute("en", "general-post", "").Path)
}

func TestRoutePrefixContractsKeepProductsAndCategoriesSeparate(t *testing.T) {
	require.True(t, IsProductRoute("en", "/products/demo-wheel", "demo-wheel"))
	require.True(t, IsProductRoute("zh-CN", "/zh_cn/products/demo-wheel", "demo-wheel"))
	require.False(t, IsProductRoute("en", "/shop/demo-wheel", "demo-wheel"))
	require.False(t, IsProductRoute("en", "/products/demo-wheel/extra", "demo-wheel"))

	require.True(t, IsProductRoutePath("en", "/products/demo-wheel"))
	require.True(t, IsProductRoutePath("zh-CN", "/zh_cn/products/demo-wheel"))
	require.False(t, IsProductRoutePath("en", "/shop/demo-wheel"))
	require.False(t, IsProductRoutePath("en", "/products/demo-wheel/extra"))

	require.True(t, IsCategoryRoutePath("en", "/shop/road-wheels"))
	require.True(t, IsCategoryRoutePath("zh-CN", "/zh_cn/shop/road-wheels/climbing"))
	require.False(t, IsCategoryRoutePath("en", "/shop"))
	require.False(t, IsCategoryRoutePath("en", "/products/demo-wheel"))
}
