package orderevidence

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOutboundWeightPackagingRecordRequiresStableFields(t *testing.T) {
	record, err := ParseOutboundWeightPackagingRecord([]byte(`{
		"schema_version": 1,
		"gross_weight_g": 12640,
		"package_count": 2,
		"packaging_method": "double wall carton",
		"packaging_note": "corner protection installed"
	}`))

	require.NoError(t, err)
	assert.Equal(t, 12640, record.GrossWeightGrams)
	assert.Equal(t, 2, record.PackageCount)
	assert.Equal(t, "double wall carton", record.PackagingMethod)
}

func TestParseOutboundWeightPackagingRecordRejectsDriftAndInvalidValues(t *testing.T) {
	_, err := ParseOutboundWeightPackagingRecord([]byte(`{
		"schema_version": 1,
		"gross_weight_g": 0,
		"package_count": 1,
		"packaging_method": "box"
	}`))
	require.ErrorContains(t, err, "gross_weight_g")

	_, err = ParseOutboundWeightPackagingRecord([]byte(`{
		"schema_version": 1,
		"gross_weight_g": 1000,
		"package_count": 1,
		"packaging_method": "box",
		"weight": 1000
	}`))
	require.ErrorContains(t, err, "unknown field")
}

func TestParseSignedPODRecordRequiresTrackingAndManualDeliveryTime(t *testing.T) {
	deliveredAt := time.Date(2026, 9, 5, 10, 30, 0, 0, time.UTC)
	record, err := ParseSignedPODRecord([]byte(`{
		"schema_version": 1,
		"tracking_number": "TRACK-100",
		"delivered_at": "2026-09-05T10:30:00Z",
		"recipient_name": "A. Customer",
		"proof_reference": "POD-100"
	}`))

	require.NoError(t, err)
	assert.Equal(t, "TRACK-100", record.TrackingNumber)
	assert.Equal(t, deliveredAt, record.DeliveredAt)

	_, err = ParseSignedPODRecord([]byte(`{
		"schema_version": 1,
		"delivered_at": "2026-09-05T10:30:00Z"
	}`))
	require.ErrorContains(t, err, "tracking_number")
}
