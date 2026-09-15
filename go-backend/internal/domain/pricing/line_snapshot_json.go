package pricing

import (
	"encoding/json"
	"errors"
	"fmt"

	"commerce-platform/internal/domain/money"
)

const LineSnapshotSchemaVersion = 1

var ErrInvalidLineSnapshot = errors.New("invalid pricing line snapshot")

// LineSnapshotPayload is the persistence/transport representation of one
// immutable pricing line. Amounts are always minor-unit integers.
type LineSnapshotPayload struct {
	SchemaVersion      int                         `json:"schema_version"`
	Key                string                      `json:"key"`
	ProductID          uint                        `json:"product_id"`
	VariantID          uint                        `json:"variant_id"`
	Quantity           int                         `json:"quantity"`
	Currency           string                      `json:"currency"`
	UnitPriceMinor     int64                       `json:"unit_price_minor"`
	BaseSubtotalMinor  int64                       `json:"base_subtotal_minor"`
	DiscountTotalMinor int64                       `json:"discount_total_minor"`
	NetSubtotalMinor   int64                       `json:"net_subtotal_minor"`
	TaxMinor           int64                       `json:"tax_minor"`
	Discounts          []DiscountAllocationPayload `json:"discounts"`
}

type DiscountAllocationPayload struct {
	Kind        DiscountKind `json:"kind"`
	Reference   string       `json:"reference"`
	AmountMinor int64        `json:"amount_minor"`
}

// MarshalLineSnapshot serializes one line without exposing private domain
// fields or floating-point amounts.
func MarshalLineSnapshot(line LineSnapshot) ([]byte, error) {
	payload, err := line.Payload()
	if err != nil {
		return nil, err
	}
	return json.Marshal(payload)
}

func (l LineSnapshot) Payload() (LineSnapshotPayload, error) {
	if l.key == "" || l.quantity <= 0 {
		return LineSnapshotPayload{}, ErrInvalidLineSnapshot
	}
	if l.unitPrice.AmountMinor() < 0 || l.baseSubtotal.AmountMinor() < 0 || l.discountTotal.AmountMinor() < 0 || l.netSubtotal.AmountMinor() < 0 {
		return LineSnapshotPayload{}, fmt.Errorf("%w: line amounts must be non-negative", ErrInvalidLineSnapshot)
	}
	if err := l.unitPrice.Validate(); err != nil {
		return LineSnapshotPayload{}, fmt.Errorf("%w: unit price: %v", ErrInvalidLineSnapshot, err)
	}
	if l.baseSubtotal.Currency() != l.unitPrice.Currency() ||
		l.discountTotal.Currency() != l.unitPrice.Currency() ||
		l.netSubtotal.Currency() != l.unitPrice.Currency() {
		return LineSnapshotPayload{}, fmt.Errorf("%w: line currencies do not match", ErrInvalidLineSnapshot)
	}
	expectedBase, err := l.unitPrice.MultiplyInt(int64(l.quantity))
	if err != nil || expectedBase.AmountMinor() != l.baseSubtotal.AmountMinor() {
		return LineSnapshotPayload{}, fmt.Errorf("%w: base subtotal is inconsistent", ErrInvalidLineSnapshot)
	}
	expectedNet, err := l.baseSubtotal.Subtract(l.discountTotal)
	if err != nil || expectedNet.AmountMinor() != l.netSubtotal.AmountMinor() {
		return LineSnapshotPayload{}, fmt.Errorf("%w: net subtotal is inconsistent", ErrInvalidLineSnapshot)
	}

	payload := LineSnapshotPayload{
		SchemaVersion:      LineSnapshotSchemaVersion,
		Key:                l.key,
		ProductID:          l.productID,
		VariantID:          l.variantID,
		Quantity:           l.quantity,
		Currency:           l.unitPrice.Currency().String(),
		UnitPriceMinor:     l.unitPrice.AmountMinor(),
		BaseSubtotalMinor:  l.baseSubtotal.AmountMinor(),
		DiscountTotalMinor: l.discountTotal.AmountMinor(),
		NetSubtotalMinor:   l.netSubtotal.AmountMinor(),
		TaxMinor:           l.tax.AmountMinor(),
		Discounts:          make([]DiscountAllocationPayload, len(l.discounts)),
	}
	var discountTotal int64
	for i, discount := range l.discounts {
		if !validDiscountKind(discount.kind) || discount.amount.AmountMinor() < 0 || discount.amount.Currency() != l.unitPrice.Currency() {
			return LineSnapshotPayload{}, fmt.Errorf("%w: invalid discount allocation", ErrInvalidLineSnapshot)
		}
		if discountTotal > int64(^uint64(0)>>1)-discount.amount.AmountMinor() {
			return LineSnapshotPayload{}, fmt.Errorf("%w: discount total overflows int64", ErrInvalidLineSnapshot)
		}
		discountTotal += discount.amount.AmountMinor()
		payload.Discounts[i] = DiscountAllocationPayload{
			Kind:        discount.kind,
			Reference:   discount.reference,
			AmountMinor: discount.amount.AmountMinor(),
		}
	}
	if discountTotal != l.discountTotal.AmountMinor() {
		return LineSnapshotPayload{}, fmt.Errorf("%w: discount allocations do not sum to discount total", ErrInvalidLineSnapshot)
	}
	if l.discountTotal.AmountMinor() > l.baseSubtotal.AmountMinor() {
		return LineSnapshotPayload{}, fmt.Errorf("%w: discount total exceeds base subtotal", ErrInvalidLineSnapshot)
	}
	if l.tax.AmountMinor() < 0 || l.tax.Currency() != l.unitPrice.Currency() {
		return LineSnapshotPayload{}, fmt.Errorf("%w: invalid line tax", ErrInvalidLineSnapshot)
	}
	return payload, nil
}

// ParseLineSnapshot validates persisted data before reconstructing the domain
// value. Invalid or tampered rows are rejected instead of being reinterpreted.
func ParseLineSnapshot(raw []byte) (LineSnapshot, error) {
	if len(raw) == 0 {
		return LineSnapshot{}, ErrInvalidLineSnapshot
	}
	var payload LineSnapshotPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return LineSnapshot{}, fmt.Errorf("%w: decode JSON: %v", ErrInvalidLineSnapshot, err)
	}
	if payload.SchemaVersion != LineSnapshotSchemaVersion || payload.Key == "" || payload.Quantity <= 0 {
		return LineSnapshot{}, ErrInvalidLineSnapshot
	}
	if payload.UnitPriceMinor < 0 || payload.BaseSubtotalMinor < 0 || payload.DiscountTotalMinor < 0 || payload.NetSubtotalMinor < 0 {
		return LineSnapshot{}, fmt.Errorf("%w: line amounts must be non-negative", ErrInvalidLineSnapshot)
	}
	if payload.DiscountTotalMinor > payload.BaseSubtotalMinor {
		return LineSnapshot{}, fmt.Errorf("%w: discount total exceeds base subtotal", ErrInvalidLineSnapshot)
	}
	unitPrice, err := money.New(payload.UnitPriceMinor, payload.Currency)
	if err != nil {
		return LineSnapshot{}, fmt.Errorf("%w: unit price: %v", ErrInvalidLineSnapshot, err)
	}
	baseSubtotal, err := money.New(payload.BaseSubtotalMinor, payload.Currency)
	if err != nil {
		return LineSnapshot{}, fmt.Errorf("%w: base subtotal: %v", ErrInvalidLineSnapshot, err)
	}
	discountTotal, err := money.New(payload.DiscountTotalMinor, payload.Currency)
	if err != nil {
		return LineSnapshot{}, fmt.Errorf("%w: discount total: %v", ErrInvalidLineSnapshot, err)
	}
	netSubtotal, err := money.New(payload.NetSubtotalMinor, payload.Currency)
	if err != nil {
		return LineSnapshot{}, fmt.Errorf("%w: net subtotal: %v", ErrInvalidLineSnapshot, err)
	}
	tax, err := money.New(payload.TaxMinor, payload.Currency)
	if err != nil {
		return LineSnapshot{}, fmt.Errorf("%w: tax: %v", ErrInvalidLineSnapshot, err)
	}
	line := LineSnapshot{
		key:           payload.Key,
		productID:     payload.ProductID,
		variantID:     payload.VariantID,
		quantity:      payload.Quantity,
		unitPrice:     unitPrice,
		baseSubtotal:  baseSubtotal,
		discountTotal: discountTotal,
		netSubtotal:   netSubtotal,
		tax:           tax,
		discounts:     make([]DiscountAllocation, len(payload.Discounts)),
	}
	for i, discount := range payload.Discounts {
		amount, amountErr := money.New(discount.AmountMinor, payload.Currency)
		if amountErr != nil {
			return LineSnapshot{}, fmt.Errorf("%w: discount allocation: %v", ErrInvalidLineSnapshot, amountErr)
		}
		line.discounts[i] = DiscountAllocation{
			kind:      discount.Kind,
			reference: discount.Reference,
			amount:    amount,
		}
	}
	if _, err := line.Payload(); err != nil {
		return LineSnapshot{}, err
	}
	return line, nil
}
