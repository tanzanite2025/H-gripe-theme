package registration

import (
	"commerce-platform/internal/api/v1/publicmedia"
	domainproduct "commerce-platform/internal/domain/product"
	domainregistration "commerce-platform/internal/domain/registration"
	"commerce-platform/internal/service"
)

func publicRegistrationFromDomain(
	item domainregistration.ProductRegistration,
	resolver publicmedia.Resolver,
) domainregistration.ProductRegistration {
	result := item
	result.PurchaseProof = publicmedia.URL(resolver, item.PurchaseProof)

	if item.Product != nil {
		product := *item.Product
		product.Media = append([]domainproduct.ProductMedia(nil), item.Product.Media...)
		for index := range product.Media {
			product.Media[index].URL = publicmedia.URL(resolver, product.Media[index].URL)
			product.Media[index].ThumbnailURL = publicmedia.URL(resolver, product.Media[index].ThumbnailURL)
			product.Media[index].PosterURL = publicmedia.URL(resolver, product.Media[index].PosterURL)
		}
		result.Product = &product
	}

	return result
}

func publicRegistrationsFromDomain(
	items []domainregistration.ProductRegistration,
	resolver publicmedia.Resolver,
) []domainregistration.ProductRegistration {
	result := make([]domainregistration.ProductRegistration, 0, len(items))
	for _, item := range items {
		result = append(result, publicRegistrationFromDomain(item, resolver))
	}
	return result
}

func publicWarrantyClaimFromDomain(
	item domainregistration.WarrantyClaim,
	resolver publicmedia.Resolver,
) domainregistration.WarrantyClaim {
	result := item
	result.Images = service.CanonicalPublicMediaURLsJSON(resolver, item.Images)
	result.VideoURL = publicmedia.URL(resolver, item.VideoURL)

	if item.Registration != nil {
		registration := publicRegistrationFromDomain(*item.Registration, resolver)
		result.Registration = &registration
	}

	return result
}

func publicWarrantyClaimsFromDomain(
	items []domainregistration.WarrantyClaim,
	resolver publicmedia.Resolver,
) []domainregistration.WarrantyClaim {
	result := make([]domainregistration.WarrantyClaim, 0, len(items))
	for _, item := range items {
		result = append(result, publicWarrantyClaimFromDomain(item, resolver))
	}
	return result
}
