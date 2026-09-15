// Package pricing defines the immutable calculation snapshot shared by
// checkout, order creation, and refunds.
package pricing

import (
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/money"
)

const SnapshotVersion = 1

type DiscountKind string

const (
	DiscountKindMember DiscountKind = "member"
	DiscountKindCoupon DiscountKind = "coupon"
	DiscountKindPoints DiscountKind = "points"
)

var (
	ErrInvalidLine             = errors.New("invalid pricing line")
	ErrDuplicateLineKey        = errors.New("duplicate pricing line key")
	ErrInvalidDiscount         = errors.New("invalid pricing discount")
	ErrDiscountExceedsEligible = errors.New("discount exceeds eligible line net amount")
)

// LineInput is a command DTO. Snapshot construction copies its values into
// immutable line snapshots.
type LineInput struct {
	Key       string
	ProductID uint
	VariantID uint
	Quantity  int
	UnitPrice money.Money
}

// DiscountInput describes one pricing-pipeline discount stage. An empty list
// of EligibleLineKeys applies to every line. Unknown keys are rejected so a
// stale promotion rule cannot silently broaden or narrow its scope.
type DiscountInput struct {
	Kind             DiscountKind
	Reference        string
	Amount           money.Money
	EligibleLineKeys []string
}

type DiscountAllocation struct {
	kind      DiscountKind
	reference string
	amount    money.Money
}

func (a DiscountAllocation) Kind() DiscountKind  { return a.kind }
func (a DiscountAllocation) Reference() string   { return a.reference }
func (a DiscountAllocation) Amount() money.Money { return a.amount }

type LineSnapshot struct {
	key           string
	productID     uint
	variantID     uint
	quantity      int
	unitPrice     money.Money
	baseSubtotal  money.Money
	discountTotal money.Money
	netSubtotal   money.Money
	tax           money.Money
	discounts     []DiscountAllocation
}

func (l LineSnapshot) Key() string                { return l.key }
func (l LineSnapshot) ProductID() uint            { return l.productID }
func (l LineSnapshot) VariantID() uint            { return l.variantID }
func (l LineSnapshot) Quantity() int              { return l.quantity }
func (l LineSnapshot) UnitPrice() money.Money     { return l.unitPrice }
func (l LineSnapshot) BaseSubtotal() money.Money  { return l.baseSubtotal }
func (l LineSnapshot) DiscountTotal() money.Money { return l.discountTotal }
func (l LineSnapshot) NetSubtotal() money.Money   { return l.netSubtotal }
func (l LineSnapshot) Tax() money.Money           { return l.tax }
func (l LineSnapshot) DiscountAllocations() []DiscountAllocation {
	return append([]DiscountAllocation(nil), l.discounts...)
}

type Snapshot struct {
	version       int
	currency      currency.Code
	lines         []LineSnapshot
	baseTotal     money.Money
	discountTotal money.Money
	netTotal      money.Money
	taxTotal      money.Money
}

func NewSnapshot(inputs []LineInput) (Snapshot, error) {
	if len(inputs) == 0 {
		return Snapshot{}, fmt.Errorf("%w: at least one line is required", ErrInvalidLine)
	}
	firstCurrency := inputs[0].UnitPrice.Currency().String()
	zero, err := money.New(0, firstCurrency)
	if err != nil {
		return Snapshot{}, fmt.Errorf("%w: %v", ErrInvalidLine, err)
	}

	result := Snapshot{
		version:       SnapshotVersion,
		currency:      zero.Currency(),
		lines:         make([]LineSnapshot, 0, len(inputs)),
		baseTotal:     zero,
		discountTotal: zero,
		netTotal:      zero,
	}
	seen := make(map[string]struct{}, len(inputs))
	for _, input := range inputs {
		key := strings.TrimSpace(input.Key)
		if key == "" || input.Quantity <= 0 || input.UnitPrice.AmountMinor() < 0 {
			return Snapshot{}, ErrInvalidLine
		}
		if _, exists := seen[key]; exists {
			return Snapshot{}, fmt.Errorf("%w: %s", ErrDuplicateLineKey, key)
		}
		seen[key] = struct{}{}
		if input.UnitPrice.Currency() != result.currency {
			return Snapshot{}, fmt.Errorf("%w: line %s uses %s", money.ErrCurrencyMismatch, key, input.UnitPrice.Currency())
		}
		baseSubtotal, err := input.UnitPrice.MultiplyInt(int64(input.Quantity))
		if err != nil {
			return Snapshot{}, fmt.Errorf("%w: line %s subtotal: %v", ErrInvalidLine, key, err)
		}
		result.baseTotal, err = result.baseTotal.Add(baseSubtotal)
		if err != nil {
			return Snapshot{}, fmt.Errorf("%w: base total: %v", ErrInvalidLine, err)
		}
		result.lines = append(result.lines, LineSnapshot{
			key:           key,
			productID:     input.ProductID,
			variantID:     input.VariantID,
			quantity:      input.Quantity,
			unitPrice:     input.UnitPrice,
			baseSubtotal:  baseSubtotal,
			discountTotal: zero,
			netSubtotal:   baseSubtotal,
			tax:           zero,
			discounts:     []DiscountAllocation{},
		})
	}
	result.netTotal = result.baseTotal
	return result, nil
}

func (s Snapshot) Version() int               { return s.version }
func (s Snapshot) Currency() currency.Code    { return s.currency }
func (s Snapshot) BaseTotal() money.Money     { return s.baseTotal }
func (s Snapshot) DiscountTotal() money.Money { return s.discountTotal }
func (s Snapshot) NetTotal() money.Money      { return s.netTotal }
func (s Snapshot) TaxTotal() money.Money      { return s.taxTotal }
func (s Snapshot) Lines() []LineSnapshot      { return cloneLines(s.lines) }

// AllocateTax returns a new snapshot with order tax allocated to each
// discounted line by its net subtotal. Remainder minor units are distributed
// deterministically by Money.Allocate.
func (s Snapshot) AllocateTax(total money.Money) (Snapshot, error) {
	if err := s.validate(); err != nil {
		return Snapshot{}, err
	}
	if err := total.Validate(); err != nil || total.Currency() != s.currency || total.AmountMinor() < 0 {
		return Snapshot{}, ErrInvalidLine
	}
	result := s.clone()
	if total.AmountMinor() == 0 {
		result.taxTotal = total
		return result, nil
	}
	ratios := make([]int64, len(s.lines))
	var netTotal int64
	for i, line := range s.lines {
		ratios[i] = line.netSubtotal.AmountMinor()
		if ratios[i] > 0 && netTotal > int64(^uint64(0)>>1)-ratios[i] {
			return Snapshot{}, ErrInvalidLine
		}
		netTotal += ratios[i]
	}
	if netTotal <= 0 {
		return Snapshot{}, ErrInvalidLine
	}
	allocations, err := total.Allocate(ratios)
	if err != nil {
		return Snapshot{}, err
	}
	for i := range result.lines {
		result.lines[i].tax = allocations[i]
	}
	result.taxTotal = total
	return result, nil
}

// AllocateDiscount returns a new snapshot and never mutates the receiver.
func (s Snapshot) AllocateDiscount(input DiscountInput) (Snapshot, error) {
	if err := s.validate(); err != nil {
		return Snapshot{}, err
	}
	if !validDiscountKind(input.Kind) || input.Amount.AmountMinor() < 0 || input.Amount.Currency() != s.currency {
		return Snapshot{}, ErrInvalidDiscount
	}
	eligible, err := s.eligibleLineIndexes(input.EligibleLineKeys)
	if err != nil {
		return Snapshot{}, err
	}
	if input.Amount.AmountMinor() == 0 {
		return s.clone(), nil
	}

	ratios := make([]int64, len(eligible))
	eligibleTotal, _ := money.New(0, s.currency.String())
	for i, lineIndex := range eligible {
		ratios[i] = s.lines[lineIndex].netSubtotal.AmountMinor()
		eligibleTotal, err = eligibleTotal.Add(s.lines[lineIndex].netSubtotal)
		if err != nil {
			return Snapshot{}, err
		}
	}
	if input.Amount.AmountMinor() > eligibleTotal.AmountMinor() {
		return Snapshot{}, ErrDiscountExceedsEligible
	}
	allocations, err := input.Amount.Allocate(ratios)
	if err != nil {
		return Snapshot{}, fmt.Errorf("%w: %v", ErrInvalidDiscount, err)
	}

	result := s.clone()
	for i, lineIndex := range eligible {
		allocation := allocations[i]
		line := &result.lines[lineIndex]
		line.discountTotal, err = line.discountTotal.Add(allocation)
		if err != nil {
			return Snapshot{}, err
		}
		line.netSubtotal, err = line.netSubtotal.Subtract(allocation)
		if err != nil || line.netSubtotal.AmountMinor() < 0 {
			return Snapshot{}, ErrDiscountExceedsEligible
		}
		if allocation.AmountMinor() > 0 {
			line.discounts = append(line.discounts, DiscountAllocation{
				kind:      input.Kind,
				reference: strings.TrimSpace(input.Reference),
				amount:    allocation,
			})
		}
	}
	result.discountTotal, err = result.discountTotal.Add(input.Amount)
	if err != nil {
		return Snapshot{}, err
	}
	result.netTotal, err = result.baseTotal.Subtract(result.discountTotal)
	if err != nil {
		return Snapshot{}, err
	}
	return result, nil
}

func (s Snapshot) eligibleLineIndexes(keys []string) ([]int, error) {
	if len(keys) == 0 {
		indexes := make([]int, len(s.lines))
		for i := range s.lines {
			indexes[i] = i
		}
		return indexes, nil
	}
	wanted := make(map[string]struct{}, len(keys))
	for _, raw := range keys {
		key := strings.TrimSpace(raw)
		if key == "" {
			return nil, ErrInvalidDiscount
		}
		wanted[key] = struct{}{}
	}
	indexes := make([]int, 0, len(wanted))
	for i, line := range s.lines {
		if _, ok := wanted[line.key]; ok {
			indexes = append(indexes, i)
			delete(wanted, line.key)
		}
	}
	if len(wanted) > 0 {
		return nil, fmt.Errorf("%w: eligible line key does not exist", ErrInvalidDiscount)
	}
	return indexes, nil
}

func (s Snapshot) validate() error {
	if s.version != SnapshotVersion || len(s.lines) == 0 || s.currency.String() == "" {
		return errors.New("invalid pricing snapshot")
	}
	return nil
}

func (s Snapshot) clone() Snapshot {
	result := s
	result.lines = cloneLines(s.lines)
	return result
}

func cloneLines(lines []LineSnapshot) []LineSnapshot {
	result := make([]LineSnapshot, len(lines))
	copy(result, lines)
	for i := range result {
		result[i].discounts = append([]DiscountAllocation(nil), lines[i].discounts...)
	}
	return result
}

func validDiscountKind(kind DiscountKind) bool {
	switch kind {
	case DiscountKindMember, DiscountKindCoupon, DiscountKindPoints:
		return true
	default:
		return false
	}
}
