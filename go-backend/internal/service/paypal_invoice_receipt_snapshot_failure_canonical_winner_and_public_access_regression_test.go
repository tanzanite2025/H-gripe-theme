package service

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/suggestionfeedback"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPayPalInvoiceReceiptSnapshotBindingRollbackAndConcurrentCanonicalWinnerPreserveUnusedObjects(t *testing.T) {
	for _, scenario := range []string{"snapshot failure", "another snapshot won", "atomic success"} {
		t.Run(scenario, func(t *testing.T) {
			db, _, uploads, provider := createSuggestionAndPayPalAttachmentRegressionEnvironment(t)
			svc := NewPaymentService(nil, repository.NewPaymentRepository(db))
			svc.ConfigureOrderEvidenceSubmissionSnapshotRepository(repository.NewOrderEvidenceSubmissionSnapshotRepository(db))
			svc.ConfigurePayPalInvoiceDurableAttachmentUploads(uploads)
			data := []byte("%PDF-invoice-receipt-snapshot")
			url, err := uploads.uploadSuggestionOrPayPalAttachmentBytesWithDurableReceipt(t.Context(), storage.PayPalDisputeInvoiceAttachmentNamespace, 71, "invoice.pdf", data)
			require.NoError(t, err)
			key := provider.writtenKeys[0]
			event, payload := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, key)
			public := NewPublicUploadAccessService(nil, nil)
			public.ConfigureSuggestionAndPayPalAttachmentUploads(uploads)
			allowed, err := public.CanServePublicUpload(t.Context(), key)
			require.NoError(t, err)
			require.False(t, allowed)
			documents := []PayPalDisputeEvidenceDocument{{Type: "commercial_invoice", Name: payload.Name, URL: url, SHA256: payload.ContentSHA256}}
			candidate, err := buildLockedEvidenceSubmissionSnapshot("paypal", 71, "PP-CANONICAL", 1, nil, DisputeEvidenceChecklist{}, nil, paypalDisputeEvidenceSnapshotRequest{Documents: documents}, time.Now())
			require.NoError(t, err)
			if scenario == "snapshot failure" {
				require.NoError(t, db.Callback().Create().Before("gorm:create").Register("reject_invoice_snapshot_owner", func(tx *gorm.DB) {
					if tx.Statement.Table == "order_evidence_submission_snapshots" {
						tx.AddError(errors.New("snapshot insertion failed"))
					}
				}))
			}
			if scenario == "another snapshot won" {
				winner, err := buildLockedEvidenceSubmissionSnapshot("paypal", 71, "PP-CANONICAL", 1, nil, DisputeEvidenceChecklist{}, nil, paypalDisputeEvidenceSnapshotRequest{Notes: "already chosen"}, time.Now())
				require.NoError(t, err)
				require.NoError(t, db.Create(winner).Error)
			}
			canonical, err := svc.createCanonicalPayPalSubmissionSnapshotAndBindNewInvoiceReceipts(t.Context(), candidate, documents)
			if scenario == "snapshot failure" {
				require.Error(t, err)
				require.NoError(t, db.Callback().Create().Remove("reject_invoice_snapshot_owner"))
			} else {
				require.NoError(t, err)
				require.NotNil(t, canonical)
			}
			if scenario != "atomic success" {
				retained, current := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, key)
				require.Equal(t, event.ID, retained.ID)
				require.Equal(t, "ready", current.UploadState)
				result := processSuggestionAndPayPalAttachmentCleanupForRegression(t, db, uploads, retained)
				require.Equal(t, 1, result.Processed)
				require.Zero(t, result.Failed)
				_, err = provider.Open(t.Context(), key)
				require.Error(t, err)
			} else {
				require.Equal(t, candidate.SnapshotSHA256, canonical.SnapshotSHA256)
				require.ErrorIs(t, db.First(&outbox.Event{}, event.ID).Error, gorm.ErrRecordNotFound)
				allowed, err = public.CanServePublicUpload(t.Context(), key)
				require.NoError(t, err)
				require.True(t, allowed)
			}
		})
	}
}

func TestSuggestionAttachmentCannotBindExpiredClaimedRetriedOrIncompleteReceipts(t *testing.T) {
	for _, state := range []string{"expired", "processing", "attempted", "manually reset", "processed", "uploading", "missing"} {
		t.Run(state, func(t *testing.T) {
			db, svc, _, provider := createSuggestionAndPayPalAttachmentRegressionEnvironment(t)
			file := multipartFileHeader(t, "proof.png", "image/png", testOpaquePNG(t, 24, 24))
			url, err := svc.UploadSuggestionAttachmentForEligibleUser(t.Context(), 71, file)
			require.NoError(t, err)
			event, payload := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, provider.writtenKeys[0])
			updates := map[string]any{}
			switch state {
			case "expired":
				updates["available_at"] = time.Now().Add(-time.Hour)
			case "processing":
				updates["status"] = outbox.EventStatusProcessing
			case "attempted":
				updates["attempts"] = 1
			case "manually reset":
				attempt := time.Now()
				updates["attempts"] = 0
				updates["last_attempt_at"] = &attempt
			case "processed":
				updates["status"] = outbox.EventStatusProcessed
			case "uploading":
				payload.UploadState = "uploading"
				encoded, err := json.Marshal(payload)
				require.NoError(t, err)
				updates["payload"] = encoded
			case "missing":
				require.NoError(t, db.Delete(&event).Error)
			}
			if len(updates) > 0 {
				require.NoError(t, db.Model(&event).Updates(updates).Error)
			}
			err = svc.CreateWithAuthorizedSuggestionAttachmentReceipts(t.Context(), &suggestionfeedback.SuggestionFeedback{UserID: 71, Message: "feedback"}, []suggestionfeedback.Attachment{{URL: url}})
			require.Error(t, err)
		})
	}
}
