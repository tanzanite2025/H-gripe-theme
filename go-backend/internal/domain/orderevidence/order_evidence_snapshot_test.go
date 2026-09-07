package orderevidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/order"
	productrequirement "commerce-platform/internal/domain/productrequirement"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestBuildOrderEvidenceSnapshotFreezesOrderAndRequirementFacts(t *testing.T) {
	variantID := uint(22)
	orderRecord := testEvidenceOrder(750)
	orderRecord.Items = []order.OrderItem{{
		ID:          501,
		OrderID:     orderRecord.ID,
		ProductID:   10,
		VariantID:   &variantID,
		ProductName: "Wheelset Configuration",
		SKU:         "WHEEL-22H",
		Quantity:    1,
		Price:       750,
		WeightGrams: 9200,
		Attributes:  ` { "spoke_holes":  "22", "color": "black" } `,
	}}
	confirmedAt := time.Date(2026, 9, 4, 12, 30, 0, 0, time.UTC)
	ruleID := uint(901)

	snapshot, err := BuildOrderEvidenceSnapshot(orderRecord, []SnapshotItemInput{{
		Item: orderRecord.Items[0],
		ProductRequirement: productrequirement.SpokeTensionQCResolution{
			Required:        true,
			Matched:         true,
			RuleID:          &ruleID,
			RuleVersion:     "wheel-v4",
			Reason:          "configured assembly product",
			Source:          productrequirement.ResolutionSourceProduct,
			RequirementType: productrequirement.RequirementTypeSpokeTensionQC,
		},
	}}, confirmedAt)

	require.NoError(t, err)
	require.NotNil(t, snapshot)
	assert.Equal(t, orderRecord.ID, snapshot.OrderID)
	assert.Equal(t, OrderEvidenceSnapshotSchemaVersion, snapshot.SchemaVersion)
	assert.Equal(t, confirmedAt, snapshot.ConfirmedAt)
	assert.InDelta(t, 750, snapshot.OrderTotalUSD, 0.0001)
	assert.True(t, snapshot.IsHighValue)
	assert.True(t, snapshot.HasSpokeTensionQC)
	require.NoError(t, snapshot.VerifyIntegrity())

	var payload OrderEvidenceSnapshotPayload
	require.NoError(t, json.Unmarshal(snapshot.SnapshotData, &payload))
	require.Len(t, payload.Items, 1)
	assert.JSONEq(t, `{"color":"black","spoke_holes":"22"}`, string(payload.Items[0].SelectedSpecsJSON))
	assert.Equal(t, "wheel-v4", payload.Items[0].ProductRequirementSnapshot.RuleVersion)
	assert.Equal(t, 9200, payload.Items[0].WeightGrams)

	hash := sha256.Sum256(snapshot.SnapshotData)
	assert.Equal(t, hex.EncodeToString(hash[:]), snapshot.SnapshotSHA256)
}

func TestBuildOrderEvidenceSnapshotUsesOnlyFinalOrderTotalForHighValue(t *testing.T) {
	variantID := uint(22)
	orderRecord := testEvidenceOrder(750)
	item := order.OrderItem{
		ID:          502,
		OrderID:     orderRecord.ID,
		ProductID:   11,
		VariantID:   &variantID,
		ProductName: "Ordinary Accessory",
		SKU:         "ACCESSORY",
		Quantity:    1,
		Price:       5,
		WeightGrams: 100,
		Attributes:  "{}",
	}

	snapshot, err := BuildOrderEvidenceSnapshot(orderRecord, []SnapshotItemInput{{
		Item: item,
	}}, time.Time{})
	require.NoError(t, err)
	assert.True(t, snapshot.IsHighValue)
	assert.False(t, snapshot.HasSpokeTensionQC)

	orderRecord.TotalAmount = 749.99
	belowThreshold, err := BuildOrderEvidenceSnapshot(orderRecord, []SnapshotItemInput{{
		Item: item,
	}}, time.Time{})
	require.NoError(t, err)
	assert.False(t, belowThreshold.IsHighValue)
}

func TestBuildOrderEvidenceSnapshotRejectsMissingFacts(t *testing.T) {
	variantID := uint(22)
	orderRecord := testEvidenceOrder(100)
	item := order.OrderItem{
		ID:         503,
		OrderID:    orderRecord.ID,
		ProductID:  10,
		VariantID:  &variantID,
		Quantity:   1,
		Attributes: datatypes.JSON([]byte("{}")).String(),
	}

	_, err := BuildOrderEvidenceSnapshot(orderRecord, []SnapshotItemInput{{Item: item}}, time.Time{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "weight_g is required")
}

func testEvidenceOrder(total float64) *order.Order {
	return &order.Order{
		ID:          1001,
		OrderNumber: "TZ-2026-EVIDENCE-TEST",
		TotalAmount: total,
		Currency:    "USD",
		FXSnapshotData: currency.OrderFXSnapshotJSON(currency.OrderFXSnapshot{
			Version:         currency.OrderFXSnapshotVersion,
			BaseCurrency:    "USD",
			OrderCurrency:   "USD",
			BaseToOrderRate: 1,
			Source:          "test",
			CapturedAt:      time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC),
		}),
	}
}
