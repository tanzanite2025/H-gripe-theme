package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/suggestionfeedback"
	"commerce-platform/internal/pkg/storage"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSuggestionAttachmentRequiresEligibleUploaderReceiptAndBindsTrustedMetadataAtomically(t *testing.T) {
	db, svc, uploads, provider := createSuggestionAndPayPalAttachmentRegressionEnvironment(t)
	file := multipartFileHeader(t, "customer.png", "image/png", testOpaquePNG(t, 24, 24))
	url, err := svc.UploadSuggestionAttachmentForEligibleUser(t.Context(), 71, file)
	require.NoError(t, err)
	key := provider.writtenKeys[0]
	event, payload := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, key)
	require.Equal(t, "ready", payload.UploadState)
	require.Equal(t, uint(71), payload.OwnerID)
	require.WithinDuration(t, time.Now().Add(24*time.Hour), *payload.BindingExpiresAt, 5*time.Second)
	public := NewPublicUploadAccessService(nil, nil)
	public.ConfigureSuggestionAndPayPalAttachmentUploads(uploads)
	allowed, err := public.CanServePublicUpload(t.Context(), key)
	require.NoError(t, err)
	require.True(t, allowed)
	for _, forbidden := range []string{url + "?borrow=1", strings.Replace(url, "cdn.example.test", "foreign.example.test", 1), "/uploads/" + key} {
		err := svc.CreateWithAuthorizedSuggestionAttachmentReceipts(t.Context(), &suggestionfeedback.SuggestionFeedback{UserID: 71, Message: "feedback"}, []suggestionfeedback.Attachment{{URL: forbidden}})
		require.Error(t, err)
	}
	require.Error(t, svc.CreateWithAuthorizedSuggestionAttachmentReceipts(t.Context(), &suggestionfeedback.SuggestionFeedback{UserID: 72, Message: "feedback", MemberLevelMet: true}, []suggestionfeedback.Attachment{{URL: url}}))
	item := &suggestionfeedback.SuggestionFeedback{UserID: 71, Message: "feedback"}
	require.NoError(t, svc.CreateWithAuthorizedSuggestionAttachmentReceipts(t.Context(), item, []suggestionfeedback.Attachment{{URL: url, Name: "forged.exe", Size: 1}}))
	var saved []suggestionfeedback.Attachment
	require.NoError(t, json.Unmarshal(item.Attachments, &saved))
	require.Equal(t, []suggestionfeedback.Attachment{{URL: url, Name: file.Filename, Size: file.Size}}, saved)
	require.ErrorIs(t, db.First(&outbox.Event{}, event.ID).Error, gorm.ErrRecordNotFound)
	allowed, err = public.CanServePublicUpload(t.Context(), key)
	require.NoError(t, err)
	require.True(t, allowed)
	require.Error(t, svc.CreateWithAuthorizedSuggestionAttachmentReceipts(t.Context(), &suggestionfeedback.SuggestionFeedback{UserID: 71, Message: "replay"}, saved))
	require.NoError(t, db.Delete(item).Error)
	allowed, err = public.CanServePublicUpload(t.Context(), key)
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestSuggestionAttachmentMembershipCannotBeBypassedByUploadOrClientEligibilityFields(t *testing.T) {
	db, svc, _, provider := createSuggestionAndPayPalAttachmentRegressionEnvironment(t)
	file := multipartFileHeader(t, "customer.png", "image/png", testOpaquePNG(t, 24, 24))
	_, err := svc.UploadSuggestionAttachmentForEligibleUser(t.Context(), 99, file)
	require.ErrorIs(t, err, ErrSuggestionFeedbackAttachmentEligibility)
	require.Empty(t, provider.writtenKeys)
	url, err := svc.UploadSuggestionAttachmentForEligibleUser(t.Context(), 71, file)
	require.NoError(t, err)
	require.NoError(t, db.Model(&loyalty.UserLoyalty{}).Where("user_id = ?", 71).Update("member_level_id", 0).Error)
	err = svc.CreateWithAuthorizedSuggestionAttachmentReceipts(t.Context(), &suggestionfeedback.SuggestionFeedback{UserID: 71, Message: "feedback", MemberLevelMet: true}, []suggestionfeedback.Attachment{{URL: url}})
	require.ErrorIs(t, err, ErrSuggestionFeedbackAttachmentEligibility)
	_, payload := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, provider.writtenKeys[0])
	require.Equal(t, "ready", payload.UploadState)
}

func TestSuggestionAttachmentObjectExtensionFollowsSniffedMimeWhileKeepingOriginalFilename(t *testing.T) {
	db, svc, _, provider := createSuggestionAndPayPalAttachmentRegressionEnvironment(t)
	file := multipartFileHeader(t, "customer-renamed.jpg", "image/jpeg", testOpaquePNG(t, 24, 24))
	_, err := svc.UploadSuggestionAttachmentForEligibleUser(t.Context(), 71, file)
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(provider.writtenKeys[0], ".png"))
	_, payload := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, provider.writtenKeys[0])
	require.Equal(t, "image/png", payload.MimeType)
	require.Equal(t, file.Filename, payload.Name)
}

func TestSuggestionAndPayPalAttachmentFailedWritesKeepOriginalPrecommittedExactKeyForCleanup(t *testing.T) {
	for _, namespace := range []string{storage.SuggestionFeedbackAttachmentNamespace, storage.PayPalDisputeInvoiceAttachmentNamespace} {
		for _, failure := range []string{"lost response", "error with URL", "foreign URL", "wrong key", "wrong digest", "short object", "cancellation", "ready transition failure", "receipt commit failure"} {
			t.Run(namespace+"/"+failure, func(t *testing.T) {
				db, _, uploads, provider := createSuggestionAndPayPalAttachmentRegressionEnvironment(t)
				data := testOpaquePNG(t, 24, 24)
				filename := "proof.png"
				if namespace == storage.PayPalDisputeInvoiceAttachmentNamespace {
					data = []byte("%PDF-synthetic-invoice-bytes")
					filename = "invoice.pdf"
				}
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				switch failure {
				case "lost response":
					empty := ""
					provider.returnedReference = &empty
				case "error with URL":
					provider.writeError = errors.New("provider wrote bytes but response failed")
				case "foreign URL":
					foreign := "https://foreign.example.test/uploads/proof.png"
					provider.returnedReference = &foreign
				case "wrong key":
					wrong := "https://cdn.example.test/uploads/proof.png"
					provider.returnedReference = &wrong
				case "wrong digest":
					provider.replacementBytes = bytes.Repeat([]byte("z"), len(data))
				case "short object":
					provider.replacementBytes = []byte("short")
				case "cancellation":
					provider.afterWrite = cancel
				case "ready transition failure":
					require.NoError(t, db.Callback().Update().Before("gorm:update").Register("reject_ready_attachment_receipt", func(tx *gorm.DB) {
						if tx.Statement.Table == "outbox_events" {
							tx.AddError(errors.New("ready commit failed"))
						}
					}))
				case "receipt commit failure":
					require.NoError(t, db.Callback().Create().Before("gorm:create").Register("reject_initial_attachment_receipt", func(tx *gorm.DB) {
						if tx.Statement.Table == "outbox_events" {
							tx.AddError(errors.New("receipt commit failed"))
						}
					}))
				}
				_, err := uploads.uploadSuggestionOrPayPalAttachmentBytesWithDurableReceipt(ctx, namespace, 71, filename, data)
				require.Error(t, err)
				if failure == "receipt commit failure" {
					require.Empty(t, provider.writtenKeys)
					return
				}
				require.Len(t, provider.writtenKeys, 1)
				key := provider.writtenKeys[0]
				event, payload := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, key)
				require.Equal(t, "uploading", payload.UploadState)
				require.Equal(t, []string{key}, payload.ObjectKeys)
				require.Equal(t, outbox.EventStatusPending, event.Status)
				var count int64
				require.NoError(t, db.Model(&outbox.Event{}).Count(&count).Error)
				require.EqualValues(t, 1, count)
				object, err := provider.Open(t.Context(), key)
				require.NoError(t, err)
				require.NoError(t, object.ReadCloser.Close())
				if failure == "ready transition failure" {
					require.NoError(t, db.Callback().Update().Remove("reject_ready_attachment_receipt"))
				}
				result := processSuggestionAndPayPalAttachmentCleanupForRegression(t, db, uploads, event)
				require.Equal(t, 1, result.Processed)
				require.Equal(t, 0, result.Failed)
				_, err = provider.Open(t.Context(), key)
				require.Error(t, err)
				var completed outbox.Event
				require.NoError(t, db.First(&completed, event.ID).Error)
				require.Equal(t, outbox.EventStatusProcessed, completed.Status)
				require.JSONEq(t, string(event.Payload), string(completed.Payload))
			})
		}
	}
}

func TestSuggestionAttachmentOwnerFailureRollsBackEveryReceiptExchange(t *testing.T) {
	db, svc, _, provider := createSuggestionAndPayPalAttachmentRegressionEnvironment(t)
	file := multipartFileHeader(t, "customer.png", "image/png", testOpaquePNG(t, 24, 24))
	url, err := svc.UploadSuggestionAttachmentForEligibleUser(t.Context(), 71, file)
	require.NoError(t, err)
	event, _ := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, provider.writtenKeys[0])
	for _, failure := range []string{"owner insertion failure", "second attachment belongs to another user", "duplicate attachment"} {
		t.Run(failure, func(t *testing.T) {
			attachments := []suggestionfeedback.Attachment{{URL: url}}
			switch failure {
			case "owner insertion failure":
				require.NoError(t, db.Callback().Create().Before("gorm:create").Register("reject_feedback_owner", func(tx *gorm.DB) {
					if tx.Statement.Table == "suggestion_feedback" {
						tx.AddError(errors.New("owner insertion failed"))
					}
				}))
				defer db.Callback().Create().Remove("reject_feedback_owner")
			case "second attachment belongs to another user":
				other, err := svc.UploadSuggestionAttachmentForEligibleUser(t.Context(), 72, file)
				require.NoError(t, err)
				attachments = append(attachments, suggestionfeedback.Attachment{URL: other})
			case "duplicate attachment":
				attachments = append(attachments, attachments[0])
			}
			err := svc.CreateWithAuthorizedSuggestionAttachmentReceipts(t.Context(), &suggestionfeedback.SuggestionFeedback{UserID: 71, Message: "feedback"}, attachments)
			require.Error(t, err)
			retained, payload := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, provider.writtenKeys[0])
			require.Equal(t, event.ID, retained.ID)
			require.Equal(t, "ready", payload.UploadState)
			var count int64
			require.NoError(t, db.Model(&suggestionfeedback.SuggestionFeedback{}).Count(&count).Error)
			require.Zero(t, count)
		})
	}
}

func TestSuggestionAndPayPalAttachmentCleanupRetainsSameTaskUntilLastHistoricalReferenceIsRemoved(t *testing.T) {
	for _, source := range []string{"soft deleted feedback", "submission snapshot", "PayPal submission audit"} {
		t.Run(source, func(t *testing.T) {
			db, _, uploads, provider := createSuggestionAndPayPalAttachmentRegressionEnvironment(t)
			namespace := storage.SuggestionFeedbackAttachmentNamespace
			data := testOpaquePNG(t, 24, 24)
			filename := "proof.png"
			if source != "soft deleted feedback" {
				namespace = storage.PayPalDisputeInvoiceAttachmentNamespace
				data = []byte("%PDF-historical")
				filename = "invoice.pdf"
			}
			url, err := uploads.uploadSuggestionOrPayPalAttachmentBytesWithDurableReceipt(t.Context(), namespace, 71, filename, data)
			require.NoError(t, err)
			key := provider.writtenKeys[0]
			event, _ := findSuggestionOrPayPalAttachmentReceiptForRegression(t, db, key)
			table := "suggestion_feedback"
			column := "attachments"
			var rowID uint
			switch source {
			case "soft deleted feedback":
				row := suggestionfeedback.SuggestionFeedback{UserID: 71, Message: "historical", Attachments: suggestionfeedback.JSONFromAttachments([]suggestionfeedback.Attachment{{URL: url}})}
				require.NoError(t, db.Create(&row).Error)
				require.NoError(t, db.Delete(&row).Error)
				rowID = row.ID
			case "submission snapshot":
				snapshot, err := buildLockedEvidenceSubmissionSnapshot("paypal", 71, "PP-HISTORICAL", 1, nil, DisputeEvidenceChecklist{}, nil, paypalDisputeEvidenceSnapshotRequest{Documents: []PayPalDisputeEvidenceDocument{{Type: "commercial_invoice", URL: url}}}, time.Now())
				require.NoError(t, err)
				require.NoError(t, db.Create(snapshot).Error)
				table = "order_evidence_submission_snapshots"
				column = "snapshot_data"
				rowID = snapshot.ID
			case "PayPal submission audit":
				order := seedPaymentOrder(t, db, "ORD-INVOICE-HISTORY", 1, "processing", "paid")
				row := seedPayPalDispute(t, db, "PP-INVOICE-HISTORY", order.ID, "WAITING_FOR_SELLER_RESPONSE", "REQUIRED_ACTION")
				encoded, _ := json.Marshal(map[string]any{"documents": []map[string]string{{"url": url}}})
				require.NoError(t, db.Table("paypal_disputes").Where("id = ?", row.ID).Update("evidence_submission_payload", string(encoded)).Error)
				table = "paypal_disputes"
				column = "evidence_submission_payload"
				rowID = row.ID
			}
			result := processSuggestionAndPayPalAttachmentCleanupForRegression(t, db, uploads, event)
			require.Equal(t, 1, result.Failed)
			var retained outbox.Event
			require.NoError(t, db.First(&retained, event.ID).Error)
			require.Equal(t, event.EventKey, retained.EventKey)
			require.JSONEq(t, string(event.Payload), string(retained.Payload))
			object, err := provider.Open(t.Context(), key)
			require.NoError(t, err)
			require.NoError(t, object.ReadCloser.Close())
			require.NoError(t, db.Table(table).Where("id = ?", rowID).Update(column, "[]").Error)
			result = processSuggestionAndPayPalAttachmentCleanupForRegression(t, db, uploads, retained)
			require.Equal(t, 1, result.Processed)
			require.Zero(t, result.Failed)
			_, err = provider.Open(t.Context(), key)
			require.Error(t, err)
		})
	}
}
