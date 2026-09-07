package orderevidence

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const SignedPODSchemaVersion = 1

// SignedPODRecord is the human-entered delivery proof reference. A carrier
// delivery event is only read-only context and never completes this record.
type SignedPODRecord struct {
	SchemaVersion  int       `json:"schema_version"`
	TrackingNumber string    `json:"tracking_number"`
	DeliveredAt    time.Time `json:"delivered_at"`
	RecipientName  string    `json:"recipient_name,omitempty"`
	ProofReference string    `json:"proof_reference,omitempty"`
	DeliveryNote   string    `json:"delivery_note,omitempty"`
}

func (r *SignedPODRecord) Normalize() {
	if r == nil {
		return
	}
	r.TrackingNumber = strings.TrimSpace(r.TrackingNumber)
	r.RecipientName = strings.TrimSpace(r.RecipientName)
	r.ProofReference = strings.TrimSpace(r.ProofReference)
	r.DeliveryNote = strings.TrimSpace(r.DeliveryNote)
	if !r.DeliveredAt.IsZero() {
		r.DeliveredAt = r.DeliveredAt.UTC()
	}
}

func (r SignedPODRecord) Validate() error {
	if r.SchemaVersion != SignedPODSchemaVersion {
		return fmt.Errorf("unsupported signed POD schema version %d", r.SchemaVersion)
	}
	if r.TrackingNumber == "" {
		return errors.New("signed POD tracking_number is required")
	}
	if r.DeliveredAt.IsZero() {
		return errors.New("signed POD delivered_at is required")
	}
	return nil
}

func ParseSignedPODRecord(data []byte) (SignedPODRecord, error) {
	var record SignedPODRecord
	if err := decodeEvidenceRecord(data, &record); err != nil {
		return record, fmt.Errorf("decode signed POD record: %w", err)
	}
	record.Normalize()
	if err := record.Validate(); err != nil {
		return record, err
	}
	return record, nil
}
