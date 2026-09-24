package orderevidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/order"
	productrequirement "commerce-platform/internal/domain/productrequirement"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	OrderEvidenceSnapshotSchemaVersion = 1
)

var ErrOrderEvidenceSnapshotImmutable = errors.New("order evidence snapshot is immutable")

// OrderEvidenceSnapshot is the immutable order-time product and requirement
// contract. Operational evidence items and attachments will reference this
// record later; they are intentionally not part of order creation.
type OrderEvidenceSnapshot struct {
	ID                    uint           `gorm:"primaryKey" json:"id"`
	OrderID               uint           `gorm:"not null;uniqueIndex" json:"order_id"`
	SchemaVersion         int            `gorm:"not null" json:"schema_version"`
	ConfirmedAt           time.Time      `gorm:"not null" json:"confirmed_at"`
	Currency              string         `gorm:"size:3;not null" json:"currency"`
	OrderTotalAmountMinor int64          `gorm:"column:order_total_amount_minor;not null" json:"order_total_amount_minor"`
	OrderTotalUSDMinor    int64          `gorm:"column:order_total_usd_minor;not null" json:"order_total_usd_minor"`
	IsHighValue           bool           `gorm:"not null;index" json:"is_high_value"`
	HasSpokeTensionQC     bool           `gorm:"not null;index" json:"has_spoke_tension_qc"`
	SnapshotData          datatypes.JSON `gorm:"column:snapshot_data;type:jsonb;not null" json:"-"`
	SnapshotSHA256        string         `gorm:"column:snapshot_sha256;type:char(64);not null" json:"snapshot_sha256"`
	CreatedAt             time.Time      `json:"created_at"`
}

func (OrderEvidenceSnapshot) TableName() string {
	return "order_evidence_snapshots"
}

// OrderEvidenceSnapshotPayload is the canonical, hashed body of the
// order-time confirmation contract.
type OrderEvidenceSnapshotPayload struct {
	SchemaVersion         int                         `json:"schema_version"`
	ConfirmedAt           time.Time                   `json:"confirmed_at"`
	Currency              string                      `json:"currency"`
	OrderTotalAmountMinor int64                       `json:"order_total_amount_minor"`
	OrderTotalUSDMinor    int64                       `json:"order_total_usd_minor"`
	IsHighValue           bool                        `json:"is_high_value"`
	FXSnapshot            currency.OrderFXSnapshot    `json:"fx_snapshot"`
	Items                 []OrderEvidenceSnapshotItem `json:"items"`
}

type OrderEvidenceSnapshotItem struct {
	OrderItemID                uint                                        `json:"order_item_id"`
	ProductID                  uint                                        `json:"product_id"`
	VariantID                  uint                                        `json:"variant_id"`
	Quantity                   int                                         `json:"quantity"`
	SKU                        string                                      `json:"sku"`
	ProductName                string                                      `json:"product_name"`
	SelectedSpecsJSON          json.RawMessage                             `json:"selected_specs_json"`
	UnitPriceMinor             int64                                       `json:"unit_price_minor"`
	WeightGrams                int                                         `json:"weight_g"`
	ProductRequirementSnapshot productrequirement.SpokeTensionQCResolution `json:"product_requirement_snapshot"`
}

type SnapshotItemInput struct {
	Item               order.OrderItem
	ProductRequirement productrequirement.SpokeTensionQCResolution
}

func BuildOrderEvidenceSnapshot(
	orderRecord *order.Order,
	items []SnapshotItemInput,
	confirmedAt time.Time,
) (*OrderEvidenceSnapshot, error) {
	if orderRecord == nil {
		return nil, errors.New("order evidence snapshot order is required")
	}
	if orderRecord.ID == 0 {
		return nil, errors.New("order evidence snapshot order_id is required")
	}
	if confirmedAt.IsZero() {
		confirmedAt = time.Now().UTC()
	}
	confirmedAt = confirmedAt.UTC()

	fxSnapshot, err := currency.ParseOrderFXSnapshot(orderRecord.FXSnapshotData)
	if err != nil {
		return nil, fmt.Errorf("parse order FX snapshot: %w", err)
	}
	totalMoney, err := orderRecord.TotalMoney()
	if err != nil {
		return nil, fmt.Errorf("parse order total amount: %w", err)
	}
	highValueEvaluation, err := order.EvaluateHighValueOrder(totalMoney, fxSnapshot)
	if err != nil {
		return nil, fmt.Errorf("evaluate order high-value policy: %w", err)
	}
	if len(items) == 0 {
		return nil, errors.New("order evidence snapshot items are required")
	}

	payloadItems := make([]OrderEvidenceSnapshotItem, 0, len(items))
	hasSpokeTensionQC := false
	for _, input := range items {
		item := input.Item
		if item.ID == 0 {
			return nil, errors.New("order evidence snapshot order_item_id is required")
		}
		if item.ProductID == 0 {
			return nil, fmt.Errorf("order evidence snapshot product_id is required for order item %d", item.ID)
		}
		if item.VariantID == nil || *item.VariantID == 0 {
			return nil, fmt.Errorf("order evidence snapshot variant_id is required for order item %d", item.ID)
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("order evidence snapshot quantity is invalid for order item %d", item.ID)
		}
		if item.WeightGrams <= 0 {
			return nil, fmt.Errorf("order evidence snapshot weight_g is required for order item %d", item.ID)
		}
		selectedSpecs, err := canonicalJSON(item.ConfigurationEvidenceJSON())
		if err != nil {
			return nil, fmt.Errorf("canonicalize selected specs for order item %d: %w", item.ID, err)
		}

		requirement := input.ProductRequirement
		requirement.RequirementType = productrequirement.RequirementTypeSpokeTensionQC
		if requirement.Required {
			hasSpokeTensionQC = true
		}
		if strings.TrimSpace(item.Currency) == "" {
			item.Currency = orderRecord.Currency
		}
		unitPriceMoney, err := item.PriceMoney()
		if err != nil {
			return nil, fmt.Errorf("parse unit price for order item %d: %w", item.ID, err)
		}
		payloadItems = append(payloadItems, OrderEvidenceSnapshotItem{
			OrderItemID:                item.ID,
			ProductID:                  item.ProductID,
			VariantID:                  *item.VariantID,
			Quantity:                   item.Quantity,
			SKU:                        strings.TrimSpace(item.SKU),
			ProductName:                strings.TrimSpace(item.ProductName),
			SelectedSpecsJSON:          selectedSpecs,
			UnitPriceMinor:             unitPriceMoney.AmountMinor(),
			WeightGrams:                item.WeightGrams,
			ProductRequirementSnapshot: requirement,
		})
	}

	payload := OrderEvidenceSnapshotPayload{
		SchemaVersion:         OrderEvidenceSnapshotSchemaVersion,
		ConfirmedAt:           confirmedAt,
		Currency:              currency.NormalizeCode(orderRecord.Currency),
		OrderTotalAmountMinor: totalMoney.AmountMinor(),
		OrderTotalUSDMinor:    highValueEvaluation.OrderTotalUSD.AmountMinor(),
		IsHighValue:           highValueEvaluation.IsHighValue,
		FXSnapshot:            fxSnapshot,
		Items:                 payloadItems,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal order evidence snapshot: %w", err)
	}
	hash := sha256.Sum256(payloadBytes)

	snapshot := &OrderEvidenceSnapshot{
		OrderID:               orderRecord.ID,
		SchemaVersion:         payload.SchemaVersion,
		ConfirmedAt:           payload.ConfirmedAt,
		Currency:              payload.Currency,
		OrderTotalAmountMinor: payload.OrderTotalAmountMinor,
		OrderTotalUSDMinor:    payload.OrderTotalUSDMinor,
		IsHighValue:           payload.IsHighValue,
		HasSpokeTensionQC:     hasSpokeTensionQC,
		SnapshotData:          datatypes.JSON(payloadBytes),
		SnapshotSHA256:        hex.EncodeToString(hash[:]),
	}
	if err := snapshot.Validate(); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func ParseOrderEvidenceSnapshotPayload(snapshot *OrderEvidenceSnapshot) (OrderEvidenceSnapshotPayload, error) {
	if snapshot == nil {
		return OrderEvidenceSnapshotPayload{}, errors.New("order evidence snapshot is required")
	}
	if err := snapshot.Validate(); err != nil {
		return OrderEvidenceSnapshotPayload{}, err
	}
	var payload OrderEvidenceSnapshotPayload
	if err := json.Unmarshal(snapshot.SnapshotData, &payload); err != nil {
		return OrderEvidenceSnapshotPayload{}, fmt.Errorf("decode order evidence snapshot payload: %w", err)
	}
	if payload.SchemaVersion != OrderEvidenceSnapshotSchemaVersion {
		return OrderEvidenceSnapshotPayload{}, fmt.Errorf(
			"unsupported order evidence snapshot payload schema version %d",
			payload.SchemaVersion,
		)
	}
	if len(payload.Items) == 0 {
		return OrderEvidenceSnapshotPayload{}, errors.New("order evidence snapshot payload items are required")
	}
	if payload.Currency != currency.NormalizeCode(snapshot.Currency) {
		return OrderEvidenceSnapshotPayload{}, errors.New("order evidence snapshot currency does not match payload")
	}
	if payload.OrderTotalAmountMinor != snapshot.OrderTotalAmountMinor {
		return OrderEvidenceSnapshotPayload{}, errors.New("order evidence snapshot total amount does not match payload")
	}
	if payload.OrderTotalUSDMinor != snapshot.OrderTotalUSDMinor {
		return OrderEvidenceSnapshotPayload{}, errors.New("order evidence snapshot USD total does not match payload")
	}
	if payload.IsHighValue != snapshot.IsHighValue {
		return OrderEvidenceSnapshotPayload{}, errors.New("order evidence snapshot high-value flag does not match payload")
	}
	if err := payload.FXSnapshot.Validate(payload.Currency); err != nil {
		return OrderEvidenceSnapshotPayload{}, fmt.Errorf("validate order evidence snapshot FX payload: %w", err)
	}
	return payload, nil
}

func (s OrderEvidenceSnapshot) Validate() error {
	if s.OrderID == 0 {
		return errors.New("order evidence snapshot order_id is required")
	}
	if s.SchemaVersion != OrderEvidenceSnapshotSchemaVersion {
		return fmt.Errorf("unsupported order evidence snapshot schema version %d", s.SchemaVersion)
	}
	if s.ConfirmedAt.IsZero() {
		return errors.New("order evidence snapshot confirmed_at is required")
	}
	if !currency.IsCatalogCode(currency.NormalizeCode(s.Currency)) {
		return errors.New("order evidence snapshot currency is invalid")
	}
	if s.OrderTotalAmountMinor < 0 || s.OrderTotalUSDMinor < 0 {
		return errors.New("order evidence snapshot totals must be finite and non-negative")
	}
	if len(s.SnapshotData) == 0 || string(s.SnapshotData) == "{}" {
		return errors.New("order evidence snapshot data is required")
	}
	if len(strings.TrimSpace(s.SnapshotSHA256)) != sha256.Size*2 {
		return errors.New("order evidence snapshot sha256 is invalid")
	}
	return s.VerifyIntegrity()
}

func (s OrderEvidenceSnapshot) VerifyIntegrity() error {
	hash := sha256.Sum256(s.SnapshotData)
	expected := hex.EncodeToString(hash[:])
	if !strings.EqualFold(strings.TrimSpace(s.SnapshotSHA256), expected) {
		return errors.New("order evidence snapshot sha256 does not match snapshot data")
	}
	return nil
}

func (s *OrderEvidenceSnapshot) BeforeCreate(tx *gorm.DB) error {
	if s == nil {
		return errors.New("order evidence snapshot is required")
	}
	return s.Validate()
}

func (s *OrderEvidenceSnapshot) BeforeUpdate(tx *gorm.DB) error {
	return ErrOrderEvidenceSnapshotImmutable
}

func (s *OrderEvidenceSnapshot) BeforeDelete(tx *gorm.DB) error {
	return ErrOrderEvidenceSnapshotImmutable
}

func canonicalJSON(raw string) (json.RawMessage, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = "{}"
	}

	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, errors.New("multiple JSON values are not allowed")
		}
		return nil, err
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(encoded), nil
}
