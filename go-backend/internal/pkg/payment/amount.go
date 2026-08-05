package payment

import (
	"fmt"
	"math"
	"strconv"

	"tanzanite/internal/domain/currency"
)

func MajorToMinorAmount(amount float64, code string) (int64, error) {
	minorUnits, ok := currency.MinorUnits(code)
	if !ok {
		return 0, fmt.Errorf("unsupported currency %s", currency.NormalizeCode(code))
	}
	scale := math.Pow10(minorUnits)
	return int64(math.Round(amount * scale)), nil
}

func MinorToMajorAmount(amount int64, code string) (float64, error) {
	minorUnits, ok := currency.MinorUnits(code)
	if !ok {
		return 0, fmt.Errorf("unsupported currency %s", currency.NormalizeCode(code))
	}
	scale := math.Pow10(minorUnits)
	return float64(amount) / scale, nil
}

func FormatMajorAmount(amount float64, code string) (string, error) {
	minorUnits, ok := currency.MinorUnits(code)
	if !ok {
		return "", fmt.Errorf("unsupported currency %s", currency.NormalizeCode(code))
	}
	return strconv.FormatFloat(amount, 'f', minorUnits, 64), nil
}
