package warranty

import (
	domainwarranty "commerce-platform/internal/domain/warranty"
	"commerce-platform/internal/service"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublicWarrantyClaimResponseCanonicalizesAttachments(t *testing.T) {
	resolver := service.NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30)
	claim := domainwarranty.WarrantyClaim{
		Images:   `["http://media.internal:8080/uploads/warranty/photo.jpg","https://cdn.example.test/photo.jpg"]`,
		VideoURL: "http://media.internal:8080/uploads/warranty/video.mp4",
	}

	publicValue := publicWarrantyClaimFromDomain(claim, resolver)

	require.Equal(
		t,
		`["https://shop.example.test/uploads/warranty/photo.jpg","https://cdn.example.test/photo.jpg"]`,
		publicValue.Images,
	)
	require.Equal(t, "https://shop.example.test/uploads/warranty/video.mp4", publicValue.VideoURL)
	require.Equal(t, "http://media.internal:8080/uploads/warranty/video.mp4", claim.VideoURL)
}
