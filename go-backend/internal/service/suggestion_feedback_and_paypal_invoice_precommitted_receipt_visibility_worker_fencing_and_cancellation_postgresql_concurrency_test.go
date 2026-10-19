package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

func createIsolatedPostgreSQLSuggestionAndPayPalAttachmentReceiptEnvironment(t *testing.T) (*gorm.DB, *SuggestionAndPayPalAttachmentUploadService, *suggestionAndPayPalAttachmentControlledStorage) {
	t.Helper()
	base, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	provider := &suggestionAndPayPalAttachmentControlledStorage{StorageService: base.storage.StorageService}
	return base.db, NewSuggestionAndPayPalAttachmentUploadService(repository.NewOutboxRepository(base.db), provider, "https://shop.example.test"), provider
}

func TestPostgreSQLSuggestionAndPayPalInvoiceSnapshotBindingKeepsIntegrityAfterJSONBStorage(t *testing.T) {
	db, uploads, provider := createIsolatedPostgreSQLSuggestionAndPayPalAttachmentReceiptEnvironment(t)
	require.NoError(t, db.AutoMigrate(&orderevidence.OrderEvidenceSubmissionSnapshot{}))
	svc := NewPaymentService(nil, repository.NewPaymentRepository(db))
	svc.ConfigureOrderEvidenceSubmissionSnapshotRepository(repository.NewOrderEvidenceSubmissionSnapshotRepository(db))
	svc.ConfigurePayPalInvoiceDurableAttachmentUploads(uploads)
	url, err := uploads.uploadSuggestionOrPayPalAttachmentBytesWithDurableReceipt(t.Context(), storage.PayPalDisputeInvoiceAttachmentNamespace, 71, "invoice.pdf", []byte("%PDF-jsonb-snapshot"))
	require.NoError(t, err)
	event, payload := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, provider.writtenKeys[0])
	documents := []PayPalDisputeEvidenceDocument{{Type: "commercial_invoice", Name: payload.Name, URL: url, SHA256: payload.ContentSHA256}}
	candidate, err := buildLockedEvidenceSubmissionSnapshot("paypal", 71, "PP-JSONB", 1, nil, DisputeEvidenceChecklist{}, nil, paypalDisputeEvidenceSnapshotRequest{Documents: documents}, time.Now())
	require.NoError(t, err)
	canonical, err := svc.createCanonicalPayPalSubmissionSnapshotAndBindNewInvoiceReceipts(t.Context(), candidate, documents)
	require.NoError(t, err)
	require.NoError(t, canonical.Validate())
	require.ErrorIs(t, db.First(&outbox.Event{}, event.ID).Error, gorm.ErrRecordNotFound)
	allowed, err := uploads.canServeOwnedSuggestionOrPayPalAttachment(t.Context(), provider.writtenKeys[0])
	require.NoError(t, err)
	require.True(t, allowed)
}

func TestPostgreSQLSuggestionAndPayPalAttachmentReceiptIsCommittedBeforeIOAndCancellationKeepsNamespaceUntilProviderReturns(t *testing.T) {
	for _, namespace := range []string{storage.SuggestionFeedbackAttachmentNamespace, storage.PayPalDisputeInvoiceAttachmentNamespace} {
		t.Run(namespace, func(t *testing.T) {
			db, uploads, provider := createIsolatedPostgreSQLSuggestionAndPayPalAttachmentReceiptEnvironment(t)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			entered, released := make(chan struct{}), make(chan struct{})
			var once sync.Once
			release := func() { once.Do(func() { close(released) }) }
			defer release()
			provider.afterWrite = func() { close(entered); <-released }
			captured := capturePrivateOrderBusinessAttachmentWriterPostgreSQLBackendPID(t, db)
			data := testOpaquePNG(t, 24, 24)
			filename := "feedback.png"
			if namespace == storage.PayPalDisputeInvoiceAttachmentNamespace {
				data = []byte("%PDF-isolated-concurrency")
				filename = "invoice.pdf"
			}
			writer := make(chan error, 1)
			go func() {
				_, err := uploads.uploadSuggestionOrPayPalAttachmentBytesWithDurableReceipt(ctx, namespace, 71, filename, data)
				writer <- err
			}()
			select {
			case <-entered:
			case err := <-writer:
				t.Fatalf("writer stopped before provider: %v", err)
			case <-time.After(10 * time.Second):
				t.Fatal("provider was not reached")
			}
			pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(t.Context(), t, captured)
			event, payload := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, provider.writtenKeys[0])
			require.Equal(t, "uploading", payload.UploadState)
			require.Equal(t, []string{provider.writtenKeys[0]}, payload.ObjectKeys)
			claimed, err := repository.NewOutboxRepository(db).ClaimReadyEvents(time.Now().Add(6*time.Minute), "suggestion-paypal-worker", 100, time.Minute)
			require.NoError(t, err)
			require.Empty(t, claimed)
			cancel()
			contender := make(chan error, 1)
			go func() {
				contender <- uploads.outbox.WithSuggestionOrPayPalAttachmentLifecycleLockUntilProviderReturns(t.Context(), func(*gorm.DB) error { return nil })
			}()
			requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t.Context(), t, db, pid)
			release()
			select {
			case err := <-writer:
				require.ErrorIs(t, err, context.Canceled)
			case <-time.After(10 * time.Second):
				t.Fatal("canceled writer did not finish")
			}
			select {
			case err := <-contender:
				require.NoError(t, err)
			case <-time.After(10 * time.Second):
				t.Fatal("namespace contender did not finish")
			}
			retained, current := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, payload.ObjectKeys[0])
			require.Equal(t, event.ID, retained.ID)
			require.Equal(t, "uploading", current.UploadState)
			result := processSuggestionAndPayPalAttachmentCleanupForRegression(t, db, uploads, retained)
			require.Equal(t, 1, result.Processed)
			require.Zero(t, result.Failed)
		})
	}
}

func TestPostgreSQLSuggestionAndPayPalCleanupClaimOrCompletionFencesWriterWaitingForNamespace(t *testing.T) {
	for _, namespace := range []string{storage.SuggestionFeedbackAttachmentNamespace, storage.PayPalDisputeInvoiceAttachmentNamespace} {
		for _, complete := range []bool{false, true} {
			scenario := "claimed"
			if complete {
				scenario = "processed"
			}
			t.Run(namespace+"/"+scenario, func(t *testing.T) {
				db, uploads, provider := createIsolatedPostgreSQLSuggestionAndPayPalAttachmentReceiptEnvironment(t)
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				locked, released := make(chan struct{}), make(chan struct{})
				var once sync.Once
				release := func() { once.Do(func() { close(released) }) }
				defer release()
				blocker := make(chan error, 1)
				go func() {
					blocker <- uploads.outbox.WithSuggestionOrPayPalAttachmentLifecycleLockUntilProviderReturns(ctx, func(*gorm.DB) error {
						close(locked)
						select {
						case <-released:
							return nil
						case <-ctx.Done():
							return ctx.Err()
						}
					})
				}()
				select {
				case <-locked:
				case <-ctx.Done():
					t.Fatal("namespace blocker did not start")
				}
				writer := make(chan error, 1)
				data := testOpaquePNG(t, 24, 24)
				filename := "feedback.png"
				if namespace == storage.PayPalDisputeInvoiceAttachmentNamespace {
					data = []byte("%PDF-fenced-invoice")
					filename = "invoice.pdf"
				}
				go func() {
					_, err := uploads.uploadSuggestionOrPayPalAttachmentBytesWithDurableReceipt(ctx, namespace, 71, filename, data)
					writer <- err
				}()
				var receipt outbox.Event
				require.Eventually(t, func() bool {
					return db.Where("event_key LIKE ?", outbox.SuggestionAndPayPalAttachmentReceiptEventKeyPrefix+"%").First(&receipt).Error == nil
				}, 5*time.Second, 20*time.Millisecond)
				repo := repository.NewOutboxRepository(db)
				claimed, err := repo.ClaimReadyEvents(time.Now().Add(6*time.Minute), "suggestion-paypal-fence", 100, time.Minute)
				require.NoError(t, err)
				require.Len(t, claimed, 1)
				if complete {
					require.NoError(t, repo.MarkProcessedByWorker(receipt.ID, "suggestion-paypal-fence", time.Now()))
				}
				release()
				require.NoError(t, <-blocker)
				select {
				case err := <-writer:
					require.ErrorIs(t, err, repository.ErrSuggestionOrPayPalAttachmentReceiptUnavailable)
				case <-ctx.Done():
					t.Fatal("fenced writer did not finish")
				}
				require.Empty(t, provider.writtenKeys)
				if !complete {
					handler := NewObjectStorageCleanupOutboxHandler(nil, nil, nil, nil)
					handler.ConfigureSuggestionAndPayPalAttachmentUploads(uploads)
					require.NoError(t, handler.Handle(t.Context(), claimed[0]))
					require.NoError(t, repo.MarkProcessedByWorker(receipt.ID, "suggestion-paypal-fence", time.Now()))
				}
			})
		}
	}
}
