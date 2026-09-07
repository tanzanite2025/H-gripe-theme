package orderevidence

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeEvidenceStorageKeyRejectsURLsAndTraversal(t *testing.T) {
	normalized, err := NormalizeEvidenceStorageKey(`/order-evidence/22/photo.jpg/`)
	require.NoError(t, err)
	assert.Equal(t, "order-evidence/22/photo.jpg", normalized)

	for _, value := range []string{
		"https://cdn.example.com/order-evidence/22/photo.jpg",
		"order-evidence/22/../other.jpg",
		"order-evidence/22/?download=1",
		"",
	} {
		_, err := NormalizeEvidenceStorageKey(value)
		require.Error(t, err, value)
	}
}

func TestOrderEvidenceAttachmentValidateRequiresObjectMetadata(t *testing.T) {
	attachment := OrderEvidenceAttachment{
		EvidenceItemID:   7,
		StorageKey:       "order-evidence/22/photo.jpg",
		OriginalFilename: "photo.jpg",
		MimeType:         "image/jpeg",
		SizeBytes:        123,
		SHA256:           "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	require.NoError(t, attachment.BeforeCreate(nil))

	attachment.MimeType = "not-a-mime"
	require.ErrorContains(t, attachment.Validate(), "mime_type is invalid")
}
