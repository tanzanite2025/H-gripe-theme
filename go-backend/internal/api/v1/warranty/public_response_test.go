package warranty

import (
	domainwarranty "commerce-platform/internal/domain/warranty"
	"commerce-platform/internal/service"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublicWarrantyClaimResponseUsesAuthorizedAttachmentPathsWithoutChangingStoredReferences(t *testing.T) {
	resolver := service.NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30)
	claim := domainwarranty.WarrantyClaim{
		ID:       17,
		Images:   `["http://media.internal:8080/uploads/warranty/photo.jpg","https://cdn.example.test/photo.jpg"]`,
		VideoURL: "http://media.internal:8080/uploads/warranty/video.mp4",
	}

	publicValue := publicWarrantyClaimFromDomain(claim, resolver)

	require.Equal(
		t,
		`["/warranty/claims/17/attachments/images/0","/warranty/claims/17/attachments/images/1"]`,
		publicValue.Images,
	)
	require.Equal(t, "/warranty/claims/17/attachments/video/0", publicValue.VideoURL)
	require.Equal(t, "http://media.internal:8080/uploads/warranty/video.mp4", claim.VideoURL)
}
