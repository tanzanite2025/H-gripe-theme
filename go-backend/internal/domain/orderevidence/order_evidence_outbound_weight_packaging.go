package orderevidence

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const OutboundWeightPackagingSchemaVersion = 1

// OutboundWeightPackagingRecord is the human-entered outbound measurement
// record. Photos and scans remain separate immutable attachments.
type OutboundWeightPackagingRecord struct {
	SchemaVersion    int    `json:"schema_version"`
	GrossWeightGrams int    `json:"gross_weight_g"`
	PackageCount     int    `json:"package_count"`
	PackagingMethod  string `json:"packaging_method"`
	PackagingNote    string `json:"packaging_note,omitempty"`
}

func (r *OutboundWeightPackagingRecord) Normalize() {
	if r == nil {
		return
	}
	r.PackagingMethod = strings.TrimSpace(r.PackagingMethod)
	r.PackagingNote = strings.TrimSpace(r.PackagingNote)
}

func (r OutboundWeightPackagingRecord) Validate() error {
	if r.SchemaVersion != OutboundWeightPackagingSchemaVersion {
		return fmt.Errorf(
			"unsupported outbound weight packaging schema version %d",
			r.SchemaVersion,
		)
	}
	if r.GrossWeightGrams <= 0 {
		return errors.New("outbound weight packaging gross_weight_g must be greater than zero")
	}
	if r.PackageCount <= 0 {
		return errors.New("outbound weight packaging package_count must be greater than zero")
	}
	if r.PackageCount > 100 {
		return errors.New("outbound weight packaging package_count is too large")
	}
	if r.PackagingMethod == "" {
		return errors.New("outbound weight packaging packaging_method is required")
	}
	return nil
}

func ParseOutboundWeightPackagingRecord(data []byte) (OutboundWeightPackagingRecord, error) {
	var record OutboundWeightPackagingRecord
	if err := decodeEvidenceRecord(data, &record); err != nil {
		return record, fmt.Errorf("decode outbound weight packaging record: %w", err)
	}
	record.Normalize()
	if err := record.Validate(); err != nil {
		return record, err
	}
	return record, nil
}

func decodeEvidenceRecord(data []byte, target any) error {
	if len(bytes.TrimSpace(data)) == 0 {
		return errors.New("structured evidence record is required")
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}

	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return errors.New("structured evidence record must contain one JSON value")
		}
		return err
	}
	return nil
}
