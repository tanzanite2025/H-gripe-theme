package marketing

import (
	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/service"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublicGiftCardsCanonicalizeCoverImageWithoutMutatingDomainValue(t *testing.T) {
	resolver := service.NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30)
	giftCards := []coupon.GiftCard{{
		CoverImage: "http://media.internal:8080/uploads/gift-card/cover.jpg",
	}}

	publicValues := publicGiftCardsFromDomain(giftCards, resolver)

	require.Equal(t, "https://shop.example.test/uploads/gift-card/cover.jpg", publicValues[0].CoverImage)
	require.Equal(t, "http://media.internal:8080/uploads/gift-card/cover.jpg", giftCards[0].CoverImage)
}
