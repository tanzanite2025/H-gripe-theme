package pricing

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidPipeline        = errors.New("invalid pricing pipeline")
	ErrDuplicateDiscountStage = errors.New("duplicate pricing discount stage")
	ErrDiscountStageOrder     = errors.New("pricing discount stages are out of order")
)

// Pipeline is an immutable, ordered pricing calculation. Discount stages may
// be omitted, but when present their canonical order is member then coupon.
type Pipeline struct {
	stages []DiscountInput
}

func NewPipeline(stages ...DiscountInput) (Pipeline, error) {
	result := Pipeline{stages: make([]DiscountInput, len(stages))}
	previousOrder := 0
	seen := make(map[DiscountKind]struct{}, len(stages))
	for i, stage := range stages {
		order, ok := discountStageOrder(stage.Kind)
		if !ok {
			return Pipeline{}, fmt.Errorf("%w: unknown stage %q", ErrInvalidPipeline, stage.Kind)
		}
		if _, exists := seen[stage.Kind]; exists {
			return Pipeline{}, fmt.Errorf("%w: %s", ErrDuplicateDiscountStage, stage.Kind)
		}
		if order < previousOrder {
			return Pipeline{}, fmt.Errorf("%w: %s", ErrDiscountStageOrder, stage.Kind)
		}
		seen[stage.Kind] = struct{}{}
		previousOrder = order
		result.stages[i] = cloneDiscountInput(stage)
	}
	return result, nil
}

// Calculate creates the base-price snapshot and passes it through each
// configured stage. Neither the pipeline nor its inputs are mutated.
func (p Pipeline) Calculate(lines []LineInput) (Snapshot, error) {
	snapshot, err := NewSnapshot(lines)
	if err != nil {
		return Snapshot{}, err
	}
	for _, stage := range p.stages {
		snapshot, err = snapshot.AllocateDiscount(stage)
		if err != nil {
			return Snapshot{}, fmt.Errorf("apply %s pricing stage: %w", stage.Kind, err)
		}
	}
	return snapshot, nil
}

// Stages returns defensive copies for diagnostics and persistence adapters.
func (p Pipeline) Stages() []DiscountInput {
	result := make([]DiscountInput, len(p.stages))
	for i, stage := range p.stages {
		result[i] = cloneDiscountInput(stage)
	}
	return result
}

func cloneDiscountInput(input DiscountInput) DiscountInput {
	result := input
	result.EligibleLineKeys = append([]string(nil), input.EligibleLineKeys...)
	return result
}

func discountStageOrder(kind DiscountKind) (int, bool) {
	switch kind {
	case DiscountKindMember:
		return 1, true
	case DiscountKindCoupon:
		return 2, true
	default:
		return 0, false
	}
}
