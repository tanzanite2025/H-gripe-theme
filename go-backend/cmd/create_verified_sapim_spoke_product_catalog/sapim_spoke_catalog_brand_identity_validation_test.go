package main

import (
	"testing"

	"commerce-platform/internal/domain/product"
)

func TestSAPIMSpokeCatalogRequiresCanonicalEnabledBrandIdentity(t *testing.T) {
	canonicalBrand := product.ProductBrand{
		Name:       verifiedSAPIMBrandName,
		Slug:       verifiedSAPIMBrandSlug,
		WebsiteURL: verifiedSAPIMBrandWebsiteURL,
		IsEnabled:  true,
	}
	if err := validateExistingVerifiedSAPIMBrandIdentity(canonicalBrand); err != nil {
		t.Fatalf("canonical SAPIM brand identity was rejected: %v", err)
	}

	for _, scenario := range []struct {
		name  string
		brand product.ProductBrand
	}{
		{
			name: "wrong display name",
			brand: product.ProductBrand{
				Name: "Other brand", Slug: verifiedSAPIMBrandSlug,
				WebsiteURL: verifiedSAPIMBrandWebsiteURL, IsEnabled: true,
			},
		},
		{
			name: "wrong official website",
			brand: product.ProductBrand{
				Name: verifiedSAPIMBrandName, Slug: verifiedSAPIMBrandSlug,
				WebsiteURL: "https://example.com/", IsEnabled: true,
			},
		},
		{
			name: "disabled brand",
			brand: product.ProductBrand{
				Name: verifiedSAPIMBrandName, Slug: verifiedSAPIMBrandSlug,
				WebsiteURL: verifiedSAPIMBrandWebsiteURL, IsEnabled: false,
			},
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			if err := validateExistingVerifiedSAPIMBrandIdentity(scenario.brand); err == nil {
				t.Fatal("mismatched SAPIM brand identity was accepted")
			}
		})
	}
}
