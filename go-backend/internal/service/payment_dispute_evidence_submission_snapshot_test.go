package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"commerce-platform/internal/domain/orderevidence"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/invoice"

	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v76"
)

func TestStripeDisputeEvidenceRetryReusesLockedSnapshot(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 61, "stripe-snapshot@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-STRIPE-SNAPSHOT-1", customer.ID)
	disputeRecord := seedStripeDispute(t, db, "dp_snapshot_1", orderRecord.ID, "needs_response")
	seedTrackingEvidence(t, db, orderRecord.ID, "DHL-SNAPSHOT-1")

	fakeSubmitter := &fakeStripeDisputeEvidenceSubmitter{
		status: stripe.DisputeStatusUnderReview,
		err:    errors.New("temporary Stripe outage"),
	}
	paymentService.stripeDisputeEvidenceSubmitter = fakeSubmitter

	_, err := paymentService.SubmitStripeDisputeEvidence(context.Background(), SubmitStripeDisputeEvidenceInput{
		DisputeID:                   disputeRecord.ID,
		APIKey:                      "sk_test_fake",
		Confirm:                     true,
		Submit:                      true,
		AdditionalStatement:         "Original operator statement.",
		ShippingDocumentationFileID: "file-original",
	})
	require.ErrorIs(t, err, fakeSubmitter.err)

	var lockedSnapshot orderevidence.OrderEvidenceSubmissionSnapshot
	require.NoError(t, db.Where(
		"provider = ? AND dispute_id = ?",
		"stripe",
		disputeRecord.ID,
	).First(&lockedSnapshot).Error)
	var lockedPayload struct {
		SchemaVersion   int                            `json:"schema_version"`
		Sources         []OrderEvidenceSourceReference `json:"sources"`
		TrackingContext *OrderEvidenceTrackingContext  `json:"tracking_context"`
	}
	require.NoError(t, json.Unmarshal(lockedSnapshot.SnapshotData, &lockedPayload))
	require.Equal(t, 1, lockedPayload.SchemaVersion)
	require.NotEmpty(t, lockedPayload.Sources)
	require.NotNil(t, lockedPayload.TrackingContext)

	require.NoError(t, db.Save(&orderRecord).Error)
	fakeSubmitter.err = nil

	result, err := paymentService.SubmitStripeDisputeEvidence(context.Background(), SubmitStripeDisputeEvidenceInput{
		DisputeID:                   disputeRecord.ID,
		APIKey:                      "sk_test_fake",
		Confirm:                     true,
		Submit:                      true,
		AdditionalStatement:         "Changed operator statement must be ignored.",
		ShippingDocumentationFileID: "file-changed",
	})
	require.NoError(t, err)
	require.Len(t, fakeSubmitter.paramsHistory, 2)
	require.NotEqual(t, uint(0), result.EvidenceSnapshotID)
	require.Equal(t, "DHL-SNAPSHOT-1", stripe.StringValue(fakeSubmitter.paramsHistory[1].Evidence.ShippingTrackingNumber))
	require.Equal(t, "file-original", stripe.StringValue(fakeSubmitter.paramsHistory[1].Evidence.ShippingDocumentation))
	require.Contains(t, stripe.StringValue(fakeSubmitter.paramsHistory[1].Evidence.UncategorizedText), "Original operator statement.")
	require.NotContains(t, stripe.StringValue(fakeSubmitter.paramsHistory[1].Evidence.UncategorizedText), "Changed operator statement")

	var snapshotCount int64
	require.NoError(t, db.Model(&orderevidence.OrderEvidenceSubmissionSnapshot{}).
		Where("provider = ? AND dispute_id = ?", "stripe", disputeRecord.ID).
		Count(&snapshotCount).Error)
	require.Equal(t, int64(1), snapshotCount)
}

func TestStripeDisputeEvidenceDraftDoesNotLockSnapshotAndFinalSubmitUsesCurrentEvidence(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 63, "stripe-stage@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-STRIPE-STAGE-1", customer.ID)
	disputeRecord := seedStripeDispute(t, db, "dp_stage_1", orderRecord.ID, "needs_response")
	seedTrackingEvidence(t, db, orderRecord.ID, "DHL-STAGE-1")

	fakeSubmitter := &fakeStripeDisputeEvidenceSubmitter{
		status: stripe.DisputeStatusUnderReview,
	}
	paymentService.stripeDisputeEvidenceSubmitter = fakeSubmitter

	staged, err := paymentService.SubmitStripeDisputeEvidence(context.Background(), SubmitStripeDisputeEvidenceInput{
		DisputeID:                   disputeRecord.ID,
		APIKey:                      "sk_test_fake",
		Confirm:                     true,
		Submit:                      false,
		AdditionalStatement:         "Stage this evidence.",
		ShippingDocumentationFileID: "file-stage",
	})
	require.NoError(t, err)
	require.True(t, staged.Staged)
	require.Nil(t, staged.SubmittedAt)
	require.Zero(t, staged.EvidenceSnapshotID)
	require.NotNil(t, fakeSubmitter.params.Submit)
	require.False(t, *fakeSubmitter.params.Submit)
	var snapshotCount int64
	require.NoError(t, db.Model(&orderevidence.OrderEvidenceSubmissionSnapshot{}).
		Where("provider = ? AND dispute_id = ?", "stripe", disputeRecord.ID).
		Count(&snapshotCount).Error)
	require.Equal(t, int64(0), snapshotCount)

	require.NoError(t, db.Save(&orderRecord).Error)
	require.NoError(t, db.Model(&shippingdomain.TrackingShipment{}).
		Where("order_id = ?", orderRecord.ID).
		Update("tracking_number", "DHL-STAGE-CHANGED").Error)
	require.NoError(t, db.Model(&shippingdomain.TrackingEvent{}).
		Where("order_id = ?", orderRecord.ID).
		Update("tracking_number", "DHL-STAGE-CHANGED").Error)
	submitted, err := paymentService.SubmitStripeDisputeEvidence(context.Background(), SubmitStripeDisputeEvidenceInput{
		DisputeID:                   disputeRecord.ID,
		APIKey:                      "sk_test_fake",
		Confirm:                     true,
		Submit:                      true,
		AdditionalStatement:         "Final statement after review.",
		ShippingDocumentationFileID: "file-final",
	})
	require.NoError(t, err)
	require.False(t, submitted.Staged)
	require.NotNil(t, submitted.SubmittedAt)
	require.NotZero(t, submitted.EvidenceSnapshotID)
	require.Len(t, fakeSubmitter.paramsHistory, 2)
	require.NotNil(t, fakeSubmitter.paramsHistory[1].Submit)
	require.True(t, *fakeSubmitter.paramsHistory[1].Submit)
	require.Equal(t, "file-final", stripe.StringValue(fakeSubmitter.paramsHistory[1].Evidence.ShippingDocumentation))
	require.Equal(t, "DHL-STAGE-CHANGED", stripe.StringValue(fakeSubmitter.paramsHistory[1].Evidence.ShippingTrackingNumber))
	require.Contains(t, stripe.StringValue(fakeSubmitter.paramsHistory[1].Evidence.UncategorizedText), "Final statement after review.")
	require.NotContains(t, stripe.StringValue(fakeSubmitter.paramsHistory[1].Evidence.UncategorizedText), "Stage this evidence.")

	require.NoError(t, db.Model(&orderevidence.OrderEvidenceSubmissionSnapshot{}).
		Where("provider = ? AND dispute_id = ?", "stripe", disputeRecord.ID).
		Count(&snapshotCount).Error)
	require.Equal(t, int64(1), snapshotCount)
}

func TestDisputeEvidenceRetryRejectsTamperedSnapshot(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 64, "tampered-snapshot@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-TAMPERED-SNAPSHOT-1", customer.ID)
	disputeRecord := seedStripeDispute(t, db, "dp_tampered_snapshot_1", orderRecord.ID, "needs_response")
	seedTrackingEvidence(t, db, orderRecord.ID, "DHL-TAMPERED-1")

	fakeSubmitter := &fakeStripeDisputeEvidenceSubmitter{
		status: stripe.DisputeStatusUnderReview,
	}
	paymentService.stripeDisputeEvidenceSubmitter = fakeSubmitter

	_, err := paymentService.SubmitStripeDisputeEvidence(context.Background(), SubmitStripeDisputeEvidenceInput{
		DisputeID:                   disputeRecord.ID,
		APIKey:                      "sk_test_fake",
		Confirm:                     true,
		Submit:                      true,
		ShippingDocumentationFileID: "file-original",
	})
	require.NoError(t, err)
	require.Len(t, fakeSubmitter.paramsHistory, 1)

	require.NoError(t, db.Exec(
		"UPDATE order_evidence_submission_snapshots SET snapshot_data = ? WHERE provider = ? AND dispute_id = ?",
		`{"tampered":true}`,
		"stripe",
		disputeRecord.ID,
	).Error)

	_, err = paymentService.SubmitStripeDisputeEvidence(context.Background(), SubmitStripeDisputeEvidenceInput{
		DisputeID: disputeRecord.ID,
		APIKey:    "sk_test_fake",
		Confirm:   true,
		Submit:    true,
	})
	require.ErrorIs(t, err, ErrOrderEvidenceSubmissionSnapshotInvalid)
	require.Len(t, fakeSubmitter.paramsHistory, 1)
}

func TestPayPalDisputeEvidenceRetryReusesSnapshotAndDoesNotReuploadInvoice(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 62, "paypal-snapshot@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-PAYPAL-SNAPSHOT-1", customer.ID)
	disputeRecord := seedPayPalDispute(t, db, "PP-D-SNAPSHOT-1", orderRecord.ID, "WAITING_FOR_SELLER_RESPONSE", "REQUIRED_ACTION")
	seedTrackingEvidence(t, db, orderRecord.ID, "DHL-PAYPAL-SNAPSHOT-1")

	fakeSubmitter := &fakePayPalDisputeEvidenceSubmitter{
		err: errors.New("temporary PayPal outage"),
	}
	fakeStorage := &fakePayPalDisputeDocumentStorage{
		url: "https://cdn.example.test/evidence/paypal-snapshot.pdf",
	}
	paymentService.ConfigurePayPalDisputeEvidenceSubmitter(fakeSubmitter)
	paymentService.ConfigurePayPalDisputeEvidenceDocumentStorage(fakeStorage)
	paymentService.ConfigurePayPalDisputeInvoiceOptions(PayPalDisputeInvoiceOptions{
		Seller: invoice.SellerProfile{
			Name:    "Commerce Platform Factory",
			Address: "100 Factory Road\nAustin, TX 78701\nUS",
		},
		AutoAttachPDF: true,
	})

	_, err := paymentService.SubmitPayPalDisputeEvidence(context.Background(), SubmitPayPalDisputeEvidenceInput{
		DisputeID:           disputeRecord.ID,
		ClientID:            "paypal-client",
		SecretKey:           "paypal-secret",
		Environment:         "sandbox",
		AdditionalStatement: "Original PayPal statement.",
	})
	require.ErrorIs(t, err, fakeSubmitter.err)
	require.Equal(t, 1, fakeStorage.uploads)

	require.NoError(t, db.Save(&orderRecord).Error)
	fakeSubmitter.err = nil

	result, err := paymentService.SubmitPayPalDisputeEvidence(context.Background(), SubmitPayPalDisputeEvidenceInput{
		DisputeID:           disputeRecord.ID,
		ClientID:            "paypal-client",
		SecretKey:           "paypal-secret",
		Environment:         "sandbox",
		AdditionalStatement: "Changed PayPal statement must be ignored.",
	})
	require.NoError(t, err)
	require.Equal(t, 1, fakeStorage.uploads)
	require.Len(t, fakeSubmitter.paramsHistory, 2)
	require.NotEqual(t, uint(0), result.EvidenceSnapshotID)
	require.Equal(t, "DHL-PAYPAL-SNAPSHOT-1", fakeSubmitter.paramsHistory[1].Evidences.EvidenceInfo.TrackingInfo[0].TrackingNumber)
	require.Contains(t, fakeSubmitter.paramsHistory[1].Evidences.Notes, "Original PayPal statement.")
	require.NotContains(t, fakeSubmitter.paramsHistory[1].Evidences.Notes, "Changed PayPal statement")
	require.Equal(t, "https://cdn.example.test/evidence/paypal-snapshot.pdf", fakeSubmitter.paramsHistory[1].Evidences.Documents[0].URL)
}
