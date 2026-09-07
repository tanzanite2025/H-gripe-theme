package orderevidence

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderEvidenceAttachmentRejectsNonHexSHA256(t *testing.T) {
	attachment := OrderEvidenceAttachment{
		EvidenceItemID:   7,
		StorageKey:       "order-evidence/22/photo.jpg",
		OriginalFilename: "photo.jpg",
		MimeType:         "image/jpeg",
		SizeBytes:        123,
		SHA256:           "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
	}

	require.ErrorContains(t, attachment.Validate(), "sha256 is invalid")
}
