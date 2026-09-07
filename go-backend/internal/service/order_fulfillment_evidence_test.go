package service

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateOrderFulfillmentEvidenceExcludesPODButRequiresEveryOrderLine(t *testing.T) {
	firstOrderItemID := uint(11)
	secondOrderItemID := uint(12)
	items := []orderevidence.OrderEvidenceItem{
		{
			ID:          1,
			OrderItemID: &firstOrderItemID,
			ItemType:    orderevidence.EvidenceItemTypeConfigurationConfirmation,
			Status:      orderevidence.EvidenceItemStatusComplete,
		},
		{
			ID:          2,
			OrderItemID: &firstOrderItemID,
			ItemType:    orderevidence.EvidenceItemTypeProductIdentity,
			Status:      orderevidence.EvidenceItemStatusComplete,
			Attachments: []orderevidence.OrderEvidenceAttachment{{ID: 21}},
		},
		{
			ID:          3,
			ItemType:    orderevidence.EvidenceItemTypeOutboundWeightPackaging,
			Status:      orderevidence.EvidenceItemStatusComplete,
			Attachments: []orderevidence.OrderEvidenceAttachment{{ID: 31}},
		},
		{
			ID:       4,
			ItemType: orderevidence.EvidenceItemTypeSignedPOD,
			Status:   orderevidence.EvidenceItemStatusMissing,
		},
	}

	check := evaluateOrderFulfillmentEvidence(items, []order.OrderItem{
		{ID: firstOrderItemID},
		{ID: secondOrderItemID},
	}, nil)

	assert.False(t, check.Ready)
	assert.Equal(t, 5, check.Total)
	assert.Equal(t, 3, check.Complete)
	assert.Equal(t, 2, check.Pending)
	assert.Len(t, check.Missing, 2)
	assert.Contains(t, check.Missing[0].Reason, "missing")
	assert.Contains(t, check.Missing[1].Reason, "missing")
}

func TestEvaluateOrderFulfillmentEvidenceRejectsWaivedPhysicalEvidence(t *testing.T) {
	orderItemID := uint(11)
	check := evaluateOrderFulfillmentEvidence(
		[]orderevidence.OrderEvidenceItem{
			{
				ID:          1,
				OrderItemID: &orderItemID,
				ItemType:    orderevidence.EvidenceItemTypeConfigurationConfirmation,
				Status:      orderevidence.EvidenceItemStatusComplete,
			},
			{
				ID:          2,
				OrderItemID: &orderItemID,
				ItemType:    orderevidence.EvidenceItemTypeProductIdentity,
				Status:      orderevidence.EvidenceItemStatusWaived,
			},
			{
				ID:       3,
				ItemType: orderevidence.EvidenceItemTypeOutboundWeightPackaging,
				Status:   orderevidence.EvidenceItemStatusComplete,
				Attachments: []orderevidence.OrderEvidenceAttachment{
					{ID: 31},
				},
			},
		},
		[]order.OrderItem{{ID: orderItemID}},
		nil,
	)

	assert.False(t, check.Ready)
	assert.Len(t, check.Missing, 1)
	assert.Equal(t, orderevidence.EvidenceItemStatusWaived, check.Missing[0].Status)
}

func TestOrderEvidenceServiceCheckFulfillmentReadinessAllowsMissingPOD(t *testing.T) {
	db := newOrderEvidenceServiceTestDB(t)
	orderRecord := order.Order{
		OrderNumber:   "ORD-DISPATCH-EVIDENCE-GATE",
		UserID:        42,
		Status:        "processing",
		PaymentStatus: "paid",
		TotalAmount:   100,
		Currency:      "USD",
	}
	if err := db.Create(&orderRecord).Error; err != nil {
		t.Fatal(err)
	}
	orderItemID := uint(11)
	orderItem := order.OrderItem{
		ID:          orderItemID,
		OrderID:     orderRecord.ID,
		ProductID:   1,
		VariantID:   uintPtrForEvidenceTest(2),
		ProductName: "Dispatch item",
		SKU:         "DISPATCH-ITEM",
		Quantity:    1,
		Price:       100,
		Subtotal:    100,
		Total:       100,
		WeightGrams: 1000,
	}
	if err := db.Create(&orderItem).Error; err != nil {
		t.Fatal(err)
	}
	orderRecord.Items = []order.OrderItem{orderItem}

	snapshot := orderevidence.OrderEvidenceSnapshot{
		OrderID:          orderRecord.ID,
		SchemaVersion:    orderevidence.OrderEvidenceSnapshotSchemaVersion,
		ConfirmedAt:      time.Now().UTC(),
		Currency:         "USD",
		OrderTotalAmount: 100,
		OrderTotalUSD:    100,
		SnapshotData:     []byte(`{"schema_version":1,"items":[{"order_item_id":11}]}`),
		SnapshotSHA256:   evidenceHashForTest(`{"schema_version":1,"items":[{"order_item_id":11}]}`),
	}
	if err := db.Create(&snapshot).Error; err != nil {
		t.Fatal(err)
	}

	evidenceRepo := repository.NewOrderEvidenceRepository(db)
	pkg := orderevidence.OrderEvidencePackage{
		OrderID:               orderRecord.ID,
		SnapshotID:            snapshot.ID,
		PackageVersion:        1,
		Status:                orderevidence.PackageStatusIncomplete,
		OrderTotalUSDSnapshot: 100,
		SchemaVersion:         orderevidence.OrderEvidencePackageSchemaVersion,
	}
	if err := db.Create(&pkg).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	configData := []byte(`{"order_item_id":11}`)
	identityData := []byte(`{"capture_note":"dispatch"}`)
	outboundData := []byte(`{"schema_version":1,"gross_weight_g":1200,"package_count":1,"packaging_method":"carton"}`)
	items := []orderevidence.OrderEvidenceItem{
		{
			PackageID:      pkg.ID,
			OrderID:        orderRecord.ID,
			OrderItemID:    &orderItemID,
			SnapshotID:     &snapshot.ID,
			ItemType:       orderevidence.EvidenceItemTypeConfigurationConfirmation,
			Status:         orderevidence.EvidenceItemStatusComplete,
			RequiredReason: orderevidence.EvidenceRequiredReasonBase,
			DataJSON:       configData,
			CapturedAt:     &now,
			ContentSHA256:  evidenceHashForTest(string(configData)),
		},
		{
			PackageID:      pkg.ID,
			OrderID:        orderRecord.ID,
			OrderItemID:    &orderItemID,
			ItemType:       orderevidence.EvidenceItemTypeProductIdentity,
			Status:         orderevidence.EvidenceItemStatusComplete,
			RequiredReason: orderevidence.EvidenceRequiredReasonBase,
			DataJSON:       identityData,
			CapturedAt:     &now,
			ContentSHA256:  evidenceHashForTest(string(identityData)),
			Attachments:    []orderevidence.OrderEvidenceAttachment{{StorageKey: "order-evidence/test/identity.jpg", OriginalFilename: "identity.jpg", MimeType: "image/jpeg", SizeBytes: 1, SHA256: evidenceHashForTest("identity"), UploadedBy: 7}},
		},
		{
			PackageID:      pkg.ID,
			OrderID:        orderRecord.ID,
			ItemType:       orderevidence.EvidenceItemTypeOutboundWeightPackaging,
			Status:         orderevidence.EvidenceItemStatusComplete,
			RequiredReason: orderevidence.EvidenceRequiredReasonBase,
			DataJSON:       outboundData,
			CapturedAt:     &now,
			ContentSHA256:  evidenceHashForTest(string(outboundData)),
			Attachments:    []orderevidence.OrderEvidenceAttachment{{StorageKey: "order-evidence/test/outbound.jpg", OriginalFilename: "outbound.jpg", MimeType: "image/jpeg", SizeBytes: 1, SHA256: evidenceHashForTest("outbound"), UploadedBy: 7}},
		},
		{
			PackageID:      pkg.ID,
			OrderID:        orderRecord.ID,
			ItemType:       orderevidence.EvidenceItemTypeSignedPOD,
			Status:         orderevidence.EvidenceItemStatusMissing,
			RequiredReason: orderevidence.EvidenceRequiredReasonBase,
		},
	}
	if err := db.Create(&items).Error; err != nil {
		t.Fatal(err)
	}

	check, err := NewOrderEvidenceService().CheckFulfillmentReadiness(
		repository.TxRepositories{OrderEvidence: evidenceRepo},
		&orderRecord,
	)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, check.Ready)
	assert.Equal(t, 3, check.Total)
	assert.Equal(t, 3, check.Complete)
}

func evidenceHashForTest(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}
