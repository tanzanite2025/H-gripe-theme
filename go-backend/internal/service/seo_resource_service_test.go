package service

import (
	"testing"

	"commerce-platform/internal/domain/product"

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

func TestSEOResourceServiceCanonicalizesProductDiagnosticImages(t *testing.T) {
	service := NewSEOResourceService(nil, nil, nil)
	service.ConfigureMediaService(NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30))

	diagnostics, err := service.ProductDiagnostics(product.Product{
		Name:     "Carbon Wheel",
		Slug:     "carbon-wheel",
		Locale:   "en",
		Status:   "active",
		Price:    199,
		Currency: "USD",
		Stock:    1,
		Media: []product.ProductMedia{{
			MediaType: "image",
			IsVisible: true,
			URL:       "http://media.internal:8080/uploads/products/carbon-wheel.jpg",
		}},
	})

	require.NoError(t, err)
	require.Equal(
		t,
		[]string{"https://shop.example.test/uploads/products/carbon-wheel.jpg"},
		diagnostics.StructuredData.Image,
	)
}
