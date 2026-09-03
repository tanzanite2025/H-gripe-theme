package service

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"

	"github.com/stretchr/testify/require"
)

func TestDisputeEvidenceUsesOrderPolicySnapshotAndRefundFacts(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	customer := seedPaymentUser(t, db, 52, "after-sales@example.test")
	orderRecord := seedDisputeEvidenceOrder(t, db, "ORD-AFTER-SALES-1", customer.ID)
	disclosedAt := time.Now().UTC().Add(-96 * time.Hour)
	consentedAt := disclosedAt.Add(2 * time.Minute)
	require.NoError(t, db.Create(&order.PolicyDisclosure{
		OrderID:         orderRecord.ID,
		PolicyKey:       "refund_cancellation_policy",
		Locale:          "en",
		RequestedLocale: "en",
		PolicyVersion:   "sha256:historical-v1",
		PolicyHash:      "historical-v1",
		PolicyJSON:      `{"title":"Historical refund policy"}`,
		PolicyURL:       "/policies/refund-cancellation",
		DisclosedAt:     disclosedAt,
		ConsentedAt:     &consentedAt,
		Source:          "checkout_test",
	}).Error)
	require.NoError(t, db.Create(&paymentdomain.Refund{
		OrderID:             orderRecord.ID,
		TransactionID:       1,
		RefundID:            stringPointer("re_after_sales_1"),
		Amount:              120,
		RequestedAmount:     125,
		CalculationSnapshot: `{"requested_amount":125,"net_amount":120}`,
		Reason:              "customer_return",
		Status:              "completed",
		CreatedAt:           disclosedAt.Add(24 * time.Hour),
		CompletedAt:         timePointer(disclosedAt.Add(48 * time.Hour)),
	}).Error)

	stripeDispute := seedStripeDispute(t, db, "dp_after_sales_1", orderRecord.ID, "needs_response")
	paypalDispute := seedPayPalDispute(t, db, "PP-AFTER-SALES-1", orderRecord.ID, "WAITING_FOR_SELLER_RESPONSE", "REQUIRED_ACTION")

	stripePackage, err := paymentService.BuildStripeDisputeEvidencePackage(stripeDispute.ID)
	require.NoError(t, err)
	require.NotNil(t, stripePackage.PolicyDisclosure)
	require.Equal(t, "sha256:historical-v1", stripePackage.PolicyDisclosure.PolicyVersion)
	require.Len(t, stripePackage.Refunds, 1)
	require.Contains(t, stripePackage.Evidence.UncategorizedText, "historical-v1")
	require.Contains(t, stripePackage.Evidence.UncategorizedText, "re_after_sales_1")
	require.Equal(t, DisputeEvidenceStatusReady, checklistItem(stripePackage.EvidenceChecklist, "policy_disclosure").Status)
	require.Equal(t, DisputeEvidenceStatusReady, checklistItem(stripePackage.EvidenceChecklist, "refund_activity").Status)

	paypalPackage, err := paymentService.BuildPayPalDisputeEvidencePackage(paypalDispute.ID)
	require.NoError(t, err)
	require.NotNil(t, paypalPackage.PolicyDisclosure)
	require.Len(t, paypalPackage.Refunds, 1)
	require.Contains(t, paypalPackage.Evidence.Notes, "historical-v1")
	require.Contains(t, paypalPackage.Evidence.Notes, "re_after_sales_1")
	require.Equal(t, DisputeEvidenceStatusReady, checklistItem(paypalPackage.EvidenceChecklist, "policy_disclosure").Status)
	require.Equal(t, DisputeEvidenceStatusReady, checklistItem(paypalPackage.EvidenceChecklist, "refund_activity").Status)
}

func checklistItem(checklist DisputeEvidenceChecklist, key string) DisputeEvidenceChecklistItem {
	for _, item := range checklist.Items {
		if item.Key == key {
			return item
		}
	}
	return DisputeEvidenceChecklistItem{}
}

func stringPointer(value string) *string {
	return &value
}

func timePointer(value time.Time) *time.Time {
	return &value
}
