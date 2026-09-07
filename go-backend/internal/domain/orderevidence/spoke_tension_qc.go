package orderevidence

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

const SpokeTensionQCPayloadSchemaVersion = 1

type SpokeTensionQCMeasurement struct {
	Position      string  `json:"position"`
	MeasuredValue float64 `json:"measured_value"`
	MinimumValue  float64 `json:"minimum_value"`
	MaximumValue  float64 `json:"maximum_value"`
	Unit          string  `json:"unit"`
}

type SpokeTensionQCRecord struct {
	OrderItemID       uint                        `json:"order_item_id"`
	AssemblyReference string                      `json:"assembly_reference"`
	Measurements      []SpokeTensionQCMeasurement `json:"measurements"`
	MeasurementTool   string                      `json:"measurement_tool"`
	MeasuredAt        time.Time                   `json:"measured_at"`
	Conclusion        string                      `json:"conclusion"`
}

// Validate checks that an optional structured transcription is well-formed.
// It deliberately does not decide whether the recorded tension is acceptable.
func (r SpokeTensionQCRecord) Validate() error {
	if r.OrderItemID == 0 {
		return errors.New("spoke tension QC order_item_id is required")
	}
	if strings.TrimSpace(r.AssemblyReference) == "" {
		return errors.New("spoke tension QC assembly_reference is required")
	}
	if len(r.Measurements) == 0 {
		return errors.New("spoke tension QC measurements are required")
	}
	if strings.TrimSpace(r.MeasurementTool) == "" {
		return errors.New("spoke tension QC measurement_tool is required")
	}
	if r.MeasuredAt.IsZero() {
		return errors.New("spoke tension QC measured_at is required")
	}

	positions := make(map[string]struct{}, len(r.Measurements))
	for index, measurement := range r.Measurements {
		position := strings.TrimSpace(measurement.Position)
		if position == "" {
			return fmt.Errorf("spoke tension QC measurement %d position is required", index)
		}
		if _, exists := positions[position]; exists {
			return fmt.Errorf("spoke tension QC measurement %d position is duplicated", index)
		}
		positions[position] = struct{}{}
		if strings.TrimSpace(measurement.Unit) == "" {
			return fmt.Errorf("spoke tension QC measurement %d unit is required", index)
		}
		if !finite(measurement.MeasuredValue) ||
			!finite(measurement.MinimumValue) ||
			!finite(measurement.MaximumValue) {
			return fmt.Errorf("spoke tension QC measurement %d values must be finite", index)
		}
		if measurement.MeasuredValue < 0 ||
			measurement.MinimumValue < 0 ||
			measurement.MaximumValue < 0 {
			return fmt.Errorf("spoke tension QC measurement %d range must be non-negative", index)
		}
		if measurement.MinimumValue > measurement.MaximumValue {
			return fmt.Errorf("spoke tension QC measurement %d minimum exceeds maximum", index)
		}
	}
	return nil
}

func (r *SpokeTensionQCRecord) Normalize() {
	if r == nil {
		return
	}
	r.AssemblyReference = strings.TrimSpace(r.AssemblyReference)
	r.MeasurementTool = strings.TrimSpace(r.MeasurementTool)
	r.Conclusion = strings.ToLower(strings.TrimSpace(r.Conclusion))
	r.MeasuredAt = r.MeasuredAt.UTC()
	for index := range r.Measurements {
		r.Measurements[index].Position = strings.TrimSpace(r.Measurements[index].Position)
		r.Measurements[index].Unit = strings.TrimSpace(r.Measurements[index].Unit)
	}
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
