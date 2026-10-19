package service

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPayPalInvoiceGenericSnapshotWritesCannotBypassAtomicReceiptBindingOrDisputeOwnership(t *testing.T) {
	for _, entry := range []string{"submission create", "submission canonical", "initial", "export", "wrong dispute"} {
		t.Run(entry, func(t *testing.T) {
			db, _, uploads, provider := createSuggestionAndPayPalAttachmentRegressionEnvironment(t)
			url, err := uploads.uploadSuggestionOrPayPalAttachmentBytesWithDurableReceipt(t.Context(), storage.PayPalDisputeInvoiceAttachmentNamespace, 71, "invoice.pdf", []byte("%PDF-generic-write-boundary"))
			require.NoError(t, err)
			event, payload := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, provider.writtenKeys[0])
			documents := []PayPalDisputeEvidenceDocument{{Type: "commercial_invoice", Name: payload.Name, URL: url, SHA256: payload.ContentSHA256}}
			candidate, err := buildLockedEvidenceSubmissionSnapshot("paypal", 71, "PP-GENERIC", 1, nil, DisputeEvidenceChecklist{}, nil, paypalDisputeEvidenceSnapshotRequest{Documents: documents}, time.Now())
			require.NoError(t, err)
			switch entry {
			case "submission create", "submission canonical":
				repo := repository.NewOrderEvidenceSubmissionSnapshotRepository(db)
				repo.ConfigureImmutableSnapshotMediaReferenceIdentityResolver(uploads)
				if entry == "submission create" {
					err = repo.Create(candidate)
				} else {
					_, err = repo.CreateOrGetLatest(candidate)
				}
			case "initial":
				require.NoError(t, db.AutoMigrate(&orderevidence.OrderEvidenceSnapshot{}))
				repo := repository.NewOrderEvidenceSnapshotRepository(db)
				repo.ConfigureImmutableSnapshotMediaReferenceIdentityResolver(uploads)
				err = repo.Create(&orderevidence.OrderEvidenceSnapshot{OrderID: 1, SchemaVersion: 1, ConfirmedAt: time.Now(), Currency: "USD", SnapshotData: candidate.SnapshotData, SnapshotSHA256: candidate.SnapshotSHA256})
			case "export":
				require.NoError(t, db.AutoMigrate(&orderevidence.OrderEvidenceExportSnapshot{}))
				repo := repository.NewOrderEvidenceExportSnapshotRepository(db)
				repo.ConfigureImmutableSnapshotMediaReferenceIdentityResolver(uploads)
				_, err = repo.CreateOrGetForPackage(&orderevidence.OrderEvidenceExportSnapshot{OrderID: 1, EvidencePackageID: 1, EvidencePackageVersion: 1, Version: 1, SchemaVersion: 1, Status: "locked", LockedAt: time.Now(), SnapshotData: candidate.SnapshotData, SnapshotSHA256: candidate.SnapshotSHA256})
			case "wrong dispute":
				service := NewPaymentService(nil, repository.NewPaymentRepository(db))
				service.ConfigureOrderEvidenceSubmissionSnapshotRepository(repository.NewOrderEvidenceSubmissionSnapshotRepository(db))
				service.ConfigurePayPalInvoiceDurableAttachmentUploads(uploads)
				candidate.DisputeID = 72
				_, err = service.createCanonicalPayPalSubmissionSnapshotAndBindNewInvoiceReceipts(t.Context(), candidate, documents)
			}
			require.ErrorIs(t, err, repository.ErrImmutableOrderEvidenceSnapshotMediaReferenceUnavailable)
			var retained outbox.Event
			require.NoError(t, db.First(&retained, event.ID).Error)
			require.Equal(t, event.EventKey, retained.EventKey)
			require.JSONEq(t, string(event.Payload), string(retained.Payload))
			var count int64
			require.NoError(t, db.Model(&orderevidence.OrderEvidenceSubmissionSnapshot{}).Count(&count).Error)
			require.Zero(t, count)
		})
	}
}

func TestPayPalInvoiceExportCanRetainAuthenticatedSameOrderHistoryWithoutAnotherReceiptExchange(t *testing.T) {
	db, _, uploads, provider := createSuggestionAndPayPalAttachmentRegressionEnvironment(t)
	verifyPayPalInvoiceExportRetainsSameOrderHistoryAfterReceiptConsumption(t, db, uploads, provider)
}

func verifyPayPalInvoiceExportRetainsSameOrderHistoryAfterReceiptConsumption(t *testing.T, db *gorm.DB, uploads *SuggestionAndPayPalAttachmentUploadService, provider *suggestionAndPayPalAttachmentControlledStorage) {
	t.Helper()
	require.NoError(t, db.AutoMigrate(&orderevidence.OrderEvidenceSubmissionSnapshot{}, &orderevidence.OrderEvidenceExportSnapshot{}))
	url, err := uploads.uploadSuggestionOrPayPalAttachmentBytesWithDurableReceipt(t.Context(), storage.PayPalDisputeInvoiceAttachmentNamespace, 71, "invoice.pdf", []byte("%PDF-retained-export"))
	require.NoError(t, err)
	event, payload := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, provider.writtenKeys[0])
	documents := []PayPalDisputeEvidenceDocument{{Type: "commercial_invoice", Name: payload.Name, URL: url, SHA256: payload.ContentSHA256}}
	candidate, err := buildLockedEvidenceSubmissionSnapshot("paypal", 71, "PP-EXPORT-HISTORY", 1, nil, DisputeEvidenceChecklist{}, nil, paypalDisputeEvidenceSnapshotRequest{Documents: documents}, time.Now())
	require.NoError(t, err)
	service := NewPaymentService(nil, repository.NewPaymentRepository(db))
	service.ConfigureOrderEvidenceSubmissionSnapshotRepository(repository.NewOrderEvidenceSubmissionSnapshotRepository(db))
	service.ConfigurePayPalInvoiceDurableAttachmentUploads(uploads)
	_, err = service.createCanonicalPayPalSubmissionSnapshotAndBindNewInvoiceReceipts(t.Context(), candidate, documents)
	require.NoError(t, err)
	data, err := json.Marshal(map[string]string{"invoice": url})
	require.NoError(t, err)
	repo := repository.NewOrderEvidenceExportSnapshotRepository(db)
	repo.ConfigureImmutableSnapshotMediaReferenceIdentityResolver(uploads)
	for _, orderID := range []uint{1, 2} {
		manifest := &orderevidence.OrderEvidenceExportSnapshot{OrderID: orderID, EvidencePackageID: orderID, EvidencePackageVersion: 1, Version: 1, Status: "locked", SchemaVersion: 1, LockedAt: time.Now(), SnapshotData: data, SnapshotSHA256: fmt.Sprintf("%x", sha256.Sum256(data))}
		stored, err := repo.CreateOrGetForPackage(manifest)
		if orderID == 1 {
			require.NoError(t, err)
			require.JSONEq(t, string(data), string(stored.SnapshotData))
			require.Equal(t, manifest.SnapshotSHA256, stored.SnapshotSHA256)
		} else {
			require.ErrorIs(t, err, repository.ErrImmutableOrderEvidenceSnapshotMediaReferenceUnavailable)
		}
	}
	require.True(t, repository.IsRecordNotFound(db.First(&outbox.Event{}, event.ID).Error))
}
