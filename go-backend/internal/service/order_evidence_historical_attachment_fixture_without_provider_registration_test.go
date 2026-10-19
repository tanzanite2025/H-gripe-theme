package service

import (
	"testing"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/repository"
)

// Historical fixtures intentionally predate verified uploads. They do not
// exercise or relax the production registration boundary.
func createHistoricalOrderEvidenceAttachmentFixtureWithoutUploading(t *testing.T, repos repository.TxRepositories, input OrderEvidenceAttachmentReferenceInput) (*orderevidence.OrderEvidenceAttachment, error) {
	t.Helper()
	attachment := &orderevidence.OrderEvidenceAttachment{
		EvidenceItemID: input.EvidenceItemID, StorageKey: input.StorageKey,
		OriginalFilename: input.OriginalFilename, MimeType: input.MimeType,
		SizeBytes: input.SizeBytes, SHA256: input.SHA256, UploadedBy: input.UploadedBy,
	}
	return attachment, repos.OrderEvidence.CreateAttachment(attachment)
}
