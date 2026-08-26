package warranty

import (
	"commerce-platform/internal/api/v1/publicmedia"
	domainwarranty "commerce-platform/internal/domain/warranty"
	"commerce-platform/internal/service"
)

func publicWarrantyClaimFromDomain(
	item domainwarranty.WarrantyClaim,
	resolver publicmedia.Resolver,
) domainwarranty.WarrantyClaim {
	result := item
	result.Images = service.CanonicalPublicMediaURLsJSON(resolver, item.Images)
	result.VideoURL = publicmedia.URL(resolver, item.VideoURL)

	return result
}

func publicWarrantyClaimsFromDomain(
	items []domainwarranty.WarrantyClaim,
	resolver publicmedia.Resolver,
) []domainwarranty.WarrantyClaim {
	result := make([]domainwarranty.WarrantyClaim, 0, len(items))
	for _, item := range items {
		result = append(result, publicWarrantyClaimFromDomain(item, resolver))
	}
	return result
}
