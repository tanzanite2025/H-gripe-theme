package orderevidence

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSpokeTensionQCRecordValidatesStructuredMeasurement(t *testing.T) {
	record := SpokeTensionQCRecord{
		OrderItemID:       12,
		AssemblyReference: "wheelset-001",
		Measurements: []SpokeTensionQCMeasurement{{
			Position:      "drive-side",
			MeasuredValue: 118,
			MinimumValue:  105,
			MaximumValue:  130,
			Unit:          "kgf",
		}},
		MeasurementTool: "TM-01",
		MeasuredAt:      time.Date(2026, 9, 5, 8, 0, 0, 0, time.UTC),
		Conclusion:      "pass",
	}

	require.NoError(t, record.Validate())
}

func TestSpokeTensionQCRecordRejectsMalformedMeasurementRange(t *testing.T) {
	record := SpokeTensionQCRecord{
		OrderItemID:       12,
		AssemblyReference: "wheelset-001",
		Measurements: []SpokeTensionQCMeasurement{{
			Position:      "drive-side",
			MeasuredValue: 118,
			MinimumValue:  130,
			MaximumValue:  105,
			Unit:          "kgf",
		}},
		MeasurementTool: "TM-01",
		MeasuredAt:      time.Now().UTC(),
		Conclusion:      "manual review required",
	}

	require.Error(t, record.Validate())
}

func TestSpokeTensionQCRecordAcceptsManualConclusionAndOutOfRangeValue(t *testing.T) {
	record := SpokeTensionQCRecord{
		OrderItemID:       12,
		AssemblyReference: "wheelset-001",
		Measurements: []SpokeTensionQCMeasurement{{
			Position:      "drive-side",
			MeasuredValue: 140,
			MinimumValue:  105,
			MaximumValue:  130,
			Unit:          "kgf",
		}},
		MeasurementTool: "TM-01",
		MeasuredAt:      time.Now().UTC(),
		Conclusion:      "fail",
	}

	require.NoError(t, record.Validate())
}
