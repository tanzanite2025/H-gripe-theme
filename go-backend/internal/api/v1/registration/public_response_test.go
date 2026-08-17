package registration

import (
	domainproduct "commerce-platform/internal/domain/product"
	domainregistration "commerce-platform/internal/domain/registration"
	"commerce-platform/internal/service"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublicRegistrationResponseCanonicalizesMediaWithoutMutatingDomainValue(t *testing.T) {
	resolver := service.NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30)
	registration := domainregistration.ProductRegistration{
		PurchaseProof: "http://media.internal:8080/uploads/proof.jpg",
		Product: &domainproduct.Product{
			Media: []domainproduct.ProductMedia{{
				URL:          "http://media.internal:8080/uploads/product.jpg",
				ThumbnailURL: "http://media.internal:8080/uploads/product-thumb.jpg",
			}},
		},
	}

	publicValue := publicRegistrationFromDomain(registration, resolver)

	require.Equal(t, "https://shop.example.test/uploads/proof.jpg", publicValue.PurchaseProof)
	require.Equal(t, "https://shop.example.test/uploads/product.jpg", publicValue.Product.Media[0].URL)
	require.Equal(t, "https://shop.example.test/uploads/product-thumb.jpg", publicValue.Product.Media[0].ThumbnailURL)
	require.Equal(t, "http://media.internal:8080/uploads/proof.jpg", registration.PurchaseProof)
	require.Equal(t, "http://media.internal:8080/uploads/product.jpg", registration.Product.Media[0].URL)
}

func TestPublicWarrantyClaimResponseCanonicalizesAttachmentsAndNestedRegistration(t *testing.T) {
	resolver := service.NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30)
	claim := domainregistration.WarrantyClaim{
		Images:   `["http://media.internal:8080/uploads/warranty/photo.jpg","https://cdn.example.test/photo.jpg"]`,
		VideoURL: "http://media.internal:8080/uploads/warranty/video.mp4",
		Registration: &domainregistration.ProductRegistration{
			PurchaseProof: "http://media.internal:8080/uploads/proof.jpg",
		},
	}

	publicValue := publicWarrantyClaimFromDomain(claim, resolver)

	require.Equal(
		t,
		`["https://shop.example.test/uploads/warranty/photo.jpg","https://cdn.example.test/photo.jpg"]`,
		publicValue.Images,
	)
	require.Equal(t, "https://shop.example.test/uploads/warranty/video.mp4", publicValue.VideoURL)
	require.Equal(t, "https://shop.example.test/uploads/proof.jpg", publicValue.Registration.PurchaseProof)
	require.Equal(t, "http://media.internal:8080/uploads/warranty/video.mp4", claim.VideoURL)
}
