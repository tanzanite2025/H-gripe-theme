package orderevidence

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestOrderEvidenceItemValidateSeparatesLineAndOrderScopes(t *testing.T) {
	lineID := uint(77)
	item := OrderEvidenceItem{
		PackageID:      1,
		OrderID:        2,
		OrderItemID:    &lineID,
		ItemType:       EvidenceItemTypeProductIdentity,
		Status:         EvidenceItemStatusMissing,
		RequiredReason: EvidenceRequiredReasonBase,
	}
	require.NoError(t, item.BeforeCreate(nil))

	item.ItemType = EvidenceItemTypeOutboundWeightPackaging
	require.ErrorContains(t, item.Validate(), "cannot have order_item_id")

	item.OrderItemID = nil
	require.NoError(t, item.Validate())
}

func TestOrderEvidenceItemCompleteRequiresMatchingStructuredDataHash(t *testing.T) {
	data := datatypes.JSON([]byte(`{"serial_number":"SN-100"}`))
	hash := sha256.Sum256(data)
	capturedAt := time.Now().UTC()
	item := OrderEvidenceItem{
		PackageID:      1,
		OrderID:        2,
		OrderItemID:    uintPtr(77),
		ItemType:       EvidenceItemTypeProductIdentity,
		Status:         EvidenceItemStatusComplete,
		RequiredReason: EvidenceRequiredReasonBase,
		DataJSON:       data,
		CapturedAt:     &capturedAt,
		ContentSHA256:  hex.EncodeToString(hash[:]),
	}
	require.NoError(t, item.Validate())

	item.ContentSHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	require.ErrorContains(t, item.Validate(), "does not match data_json")
}

func TestOrderEvidenceItemSnapshotReferenceOnlyBelongsToConfigurationConfirmation(t *testing.T) {
	item := OrderEvidenceItem{
		PackageID:      1,
		OrderID:        2,
		OrderItemID:    uintPtr(77),
		SnapshotID:     uintPtr(99),
		ItemType:       EvidenceItemTypeProductIdentity,
		Status:         EvidenceItemStatusComplete,
		RequiredReason: EvidenceRequiredReasonBase,
	}
	require.ErrorContains(t, item.Validate(), "snapshot-backed")

	item.ItemType = EvidenceItemTypeConfigurationConfirmation
	item.DataJSON = datatypes.JSON([]byte(`{"snapshot_id":99}`))
	hash := sha256.Sum256(item.DataJSON)
	item.ContentSHA256 = hex.EncodeToString(hash[:])
	capturedAt := time.Now().UTC()
	item.CapturedAt = &capturedAt
	require.NoError(t, item.Validate())
	assert.Equal(t, EvidenceItemTypeConfigurationConfirmation, item.ItemType)
}

func uintPtr(value uint) *uint {
	return &value
}
