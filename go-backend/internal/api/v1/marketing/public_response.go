package marketing

import (
	"commerce-platform/internal/api/v1/publicmedia"
	"commerce-platform/internal/domain/coupon"
)

func publicGiftCardsFromDomain(
	items []coupon.GiftCard,
	resolver publicmedia.Resolver,
) []coupon.GiftCard {
	result := make([]coupon.GiftCard, 0, len(items))
	for _, item := range items {
		item.CoverImage = publicmedia.URL(resolver, item.CoverImage)
		result = append(result, item)
	}
	return result
}
