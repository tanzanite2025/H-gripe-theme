package rating

import (
	"errors"
	"math"
)

var ErrInvalidWeightScale = errors.New("shipping weight scale is invalid")

type WeightScale struct {
	FirstWeightGrams      int
	AdditionalWeightGrams int
}

func ValidateWeightScale(additionalFee float64, scale WeightScale) error {
	if additionalFee <= 0 {
		return nil
	}
	if scale.FirstWeightGrams <= 0 || scale.AdditionalWeightGrams <= 0 {
		return ErrInvalidWeightScale
	}
	return nil
}

// AdditionalWeightUnits returns the number of configured continuation steps
// above the carrier's first included weight. Rule ranges select a price; they
// do not redefine the carrier's weight scale.
func AdditionalWeightUnits(valueGrams int, additionalFee float64, scale WeightScale) (int, error) {
	if additionalFee <= 0 {
		return 0, nil
	}
	if err := ValidateWeightScale(additionalFee, scale); err != nil {
		return 0, err
	}
	if valueGrams <= scale.FirstWeightGrams {
		return 0, nil
	}
	excessGrams := valueGrams - scale.FirstWeightGrams
	return (excessGrams + scale.AdditionalWeightGrams - 1) / scale.AdditionalWeightGrams, nil
}

func AdditionalWholeUnits(ruleMinimum float64, value float64, additionalFee float64) int {
	if additionalFee <= 0 || value <= ruleMinimum {
		return 0
	}
	units := int(math.Ceil(value-ruleMinimum)) - 1
	if units < 0 {
		return 0
	}
	return units
}

func BillableWeightGrams(chargeWeightGrams int, minimumWeightGrams int, scale WeightScale) int {
	billable := max(chargeWeightGrams, minimumWeightGrams)
	if billable <= 0 {
		return 0
	}
	if scale.FirstWeightGrams <= 0 {
		return billable
	}
	if billable <= scale.FirstWeightGrams {
		return scale.FirstWeightGrams
	}
	if scale.AdditionalWeightGrams <= 0 {
		return billable
	}
	excess := billable - scale.FirstWeightGrams
	units := int(math.Ceil(float64(excess) / float64(scale.AdditionalWeightGrams)))
	return scale.FirstWeightGrams + units*scale.AdditionalWeightGrams
}

func VolumetricWeightGrams(lengthCm float64, widthCm float64, heightCm float64, divisor int) (int, bool) {
	if divisor <= 0 || lengthCm <= 0 || widthCm <= 0 || heightCm <= 0 {
		return 0, false
	}
	weightKg := lengthCm * widthCm * heightCm / float64(divisor)
	if weightKg <= 0 {
		return 0, false
	}
	return int(math.Ceil(weightKg * 1000)), true
}
