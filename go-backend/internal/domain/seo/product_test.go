package seo

import (
	"testing"

	productdomain "tanzanite/internal/domain/product"

	"github.com/stretchr/testify/require"
)

func TestBuildProductSEOReadinessUsesSourceDataAndFallbacks(t *testing.T) {
	price := 1299.0
	item := productdomain.Product{
		ID:        42,
		Name:      "C50 Disc Carbon Wheelset",
		Slug:      "c50-disc-carbon-wheelset",
		ShortDesc: "A fast carbon wheelset.",
		SKU:       "C50",
		Currency:  "USD",
		Price:     price,
		Stock:     1,
		Status:    "active",
		Media: []productdomain.ProductMedia{{
			URL:       "/media/c50.jpg",
			MediaType: "image",
			IsVisible: true,
			IsPrimary: true,
		}},
	}

	diagnostics := BuildProductSEOReadiness(item, "HACK-GRIPE", "/shop/c50-disc-carbon-wheelset")

	require.True(t, diagnostics.Ready)
	require.Equal(t, "HACK-GRIPE", diagnostics.Brand)
	require.Equal(t, "C50", diagnostics.SKU)
	require.Equal(t, "product_name", diagnostics.MetaTitle.Source)
	require.True(t, diagnostics.MetaTitle.FallbackActive)
	require.Equal(t, "short_description", diagnostics.MetaDescription.Source)
	require.True(t, diagnostics.HasOffer)
	require.Equal(t, "Product", diagnostics.StructuredDataType)
	require.NotNil(t, diagnostics.StructuredData.Offers)
	require.Equal(t, "https://schema.org/InStock", diagnostics.StructuredData.Offers.Availability)
}

func TestBuildProductSEOReadinessUsesProductGroupForActiveVariants(t *testing.T) {
	item := productdomain.Product{
		ID:       7,
		Name:     "Variant Wheelset",
		Slug:     "variant-wheelset",
		Currency: "USD",
		Price:    100,
		Status:   "active",
		Variants: []productdomain.ProductVariant{
			{ID: 10, SKU: "VAR-10", Title: "50 mm", Currency: "USD", Price: 100, Stock: 2, IsActive: true},
			{ID: 11, SKU: "VAR-11", Title: "60 mm", Currency: "USD", Price: 120, Stock: 0, IsActive: true},
		},
		Media: []productdomain.ProductMedia{{
			URL:       "https://cdn.example.test/variant.jpg",
			MediaType: "image",
			IsVisible: true,
		}},
	}

	diagnostics := BuildProductSEOReadiness(item, "", "/shop/variant-wheelset")

	require.True(t, diagnostics.Ready)
	require.Equal(t, "ProductGroup", diagnostics.StructuredDataType)
	require.Len(t, diagnostics.StructuredData.HasVariant, 2)
	require.Equal(t, "/shop/variant-wheelset?variant=10", diagnostics.StructuredData.HasVariant[0].URL)
	require.Contains(t, diagnostics.Warnings, "brand")
	require.Contains(t, diagnostics.Missing, "brand")
}

func TestBuildProductSEOReadinessBlocksInactiveProductAndMissingImage(t *testing.T) {
	item := productdomain.Product{
		Name:     "Unpublished",
		Currency: "USD",
		Price:    10,
		Status:   "inactive",
	}

	diagnostics := BuildProductSEOReadiness(item, "Brand", "/shop/unpublished")

	require.False(t, diagnostics.Ready)
	require.Contains(t, diagnostics.BlockingIssues, "product_not_public")
	require.Contains(t, diagnostics.BlockingIssues, "image")
}
