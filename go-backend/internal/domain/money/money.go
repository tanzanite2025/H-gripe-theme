// Package money contains the domain value object used for exact monetary
// arithmetic. It is deliberately independent from persistence and transport
// models so those layers can migrate incrementally.
package money

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"commerce-platform/internal/domain/currency"
)

var (
	ErrCurrencyMismatch  = errors.New("money currency mismatch")
	ErrInvalidAllocation = errors.New("money allocation ratios must contain at least one positive value")
	ErrInvalidAmount     = errors.New("invalid major-unit amount")
	ErrExcessPrecision   = errors.New("amount has more fractional digits than the currency supports")
)

// Money stores an amount in the smallest unit of Currency (for example minor
// for USD and whole units for JPY). AmountMinor may be negative for signed
// adjustments; callers that model payments or balances should reject negative
// values at their own domain boundary.
type Money struct {
	amountMinor int64
	currency    currency.Code
}

// New constructs Money only from a validated catalog currency.
func New(amountMinor int64, code string) (Money, error) {
	currencyCode, err := currency.ParseCode(code)
	if err != nil {
		return Money{}, err
	}
	return Money{amountMinor: amountMinor, currency: currencyCode}, nil
}

// MustNew is intended for package-level constants and tests. Runtime input
// should use New and handle its error.
func MustNew(amountMinor int64, code string) Money {
	value, err := New(amountMinor, code)
	if err != nil {
		panic(err)
	}
	return value
}

// FromMajorFloat converts a persistence or transport major-unit value at the
// boundary of the domain model. Rounding is performed exactly once using the
// currency's minor-unit scale; domain arithmetic should use Money thereafter.
func FromMajorFloat(amount float64, code string) (Money, error) {
	currencyCode, err := currency.ParseCode(code)
	if err != nil {
		return Money{}, err
	}
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return Money{}, ErrInvalidAmount
	}
	minorUnits, _ := currency.MinorUnits(currencyCode.String())
	scale := math.Pow10(minorUnits)
	rounded := math.Round(amount * scale)
	int64Limit := math.Exp2(63)
	if rounded >= int64Limit || rounded < -int64Limit {
		return Money{}, errors.New("money amount overflows int64")
	}
	return Money{amountMinor: int64(rounded), currency: currencyCode}, nil
}

// MajorFloat serializes Money to a major-unit number for persistence or API
// transport. It should not be used for arithmetic.
func (m Money) MajorFloat() (float64, error) {
	formatted, err := m.FormatMajor()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(formatted, 64)
}

// ParseMajor parses a human/API amount without passing through float64. The
// parser accepts an optional sign and decimal point, trims surrounding space,
// and rejects non-zero precision beyond the currency's minor-unit scale.
func ParseMajor(value string, code string) (Money, error) {
	currencyCode, err := currency.ParseCode(code)
	if err != nil {
		return Money{}, err
	}
	minorUnits, _ := currency.MinorUnits(currencyCode.String())

	raw := strings.TrimSpace(value)
	if raw == "" {
		return Money{}, ErrInvalidAmount
	}
	sign := ""
	if raw[0] == '+' || raw[0] == '-' {
		sign = raw[:1]
		raw = raw[1:]
	}
	if raw == "" || strings.Count(raw, ".") > 1 {
		return Money{}, ErrInvalidAmount
	}
	major, fraction := raw, ""
	if dot := strings.IndexByte(raw, '.'); dot >= 0 {
		if dot == len(raw)-1 {
			return Money{}, ErrInvalidAmount
		}
		major, fraction = raw[:dot], raw[dot+1:]
	}
	if major == "" || !allDigits(major) || (fraction != "" && !allDigits(fraction)) {
		return Money{}, ErrInvalidAmount
	}
	// Extra trailing zeroes do not represent extra monetary precision. Any
	// remaining fractional digits must still fit the currency scale.
	fraction = strings.TrimRight(fraction, "0")
	if len(fraction) > minorUnits {
		return Money{}, ErrExcessPrecision
	}
	fraction += strings.Repeat("0", minorUnits-len(fraction))

	digits := strings.TrimLeft(major+fraction, "0")
	if digits == "" {
		digits = "0"
	}
	minor := new(big.Int)
	if _, ok := minor.SetString(digits, 10); !ok {
		return Money{}, ErrInvalidAmount
	}
	if sign == "-" {
		minor.Neg(minor)
	}
	if !minor.IsInt64() {
		return Money{}, errors.New("money amount overflows int64")
	}
	return Money{amountMinor: minor.Int64(), currency: currencyCode}, nil
}

// FromMajorRat constructs Money from an exact rational major-unit amount and
// rounds half away from zero to the currency's supported minor-unit scale.
// It is useful for rates and point programs whose intermediate value may be
// finer than the source currency's smallest unit.
func FromMajorRat(value *big.Rat, code string) (Money, error) {
	if value == nil {
		return Money{}, ErrInvalidAmount
	}
	currencyCode, err := currency.ParseCode(code)
	if err != nil {
		return Money{}, err
	}
	minorUnits, _ := currency.MinorUnits(currencyCode.String())
	scale := big.NewInt(pow10Int64(minorUnits))
	numerator := new(big.Int).Mul(value.Num(), scale)
	denominator := new(big.Int).Set(value.Denom())
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, denominator, remainder)
	if remainder.Sign() != 0 {
		absRemainder := new(big.Int).Abs(remainder)
		twiceRemainder := new(big.Int).Lsh(absRemainder, 1)
		if twiceRemainder.Cmp(denominator) >= 0 {
			if numerator.Sign() < 0 {
				quotient.Sub(quotient, big.NewInt(1))
			} else {
				quotient.Add(quotient, big.NewInt(1))
			}
		}
	}
	minor, err := int64FromBig(quotient)
	if err != nil {
		return Money{}, err
	}
	return Money{amountMinor: minor, currency: currencyCode}, nil
}

// FormatMajor returns a fixed-scale decimal string without using floating
// point arithmetic.
func (m Money) FormatMajor() (string, error) {
	if err := m.Validate(); err != nil {
		return "", err
	}
	minorUnits, _ := currency.MinorUnits(m.currency.String())
	value := big.NewInt(m.amountMinor)
	negative := value.Sign() < 0
	if negative {
		value.Neg(value)
	}
	digits := value.Text(10)
	if minorUnits == 0 {
		if negative && digits != "0" {
			return "-" + digits, nil
		}
		return digits, nil
	}
	if len(digits) <= minorUnits {
		digits = strings.Repeat("0", minorUnits+1-len(digits)) + digits
	}
	point := len(digits) - minorUnits
	formatted := digits[:point] + "." + digits[point:]
	if negative && m.amountMinor != 0 {
		formatted = "-" + formatted
	}
	return formatted, nil
}

func (m Money) Validate() error {
	if _, err := currency.ParseCode(m.currency.String()); err != nil {
		return err
	}
	return nil
}

func (m Money) AmountMinor() int64 {
	return m.amountMinor
}

func (m Money) Currency() currency.Code {
	return m.currency
}

func (m Money) Abs() (Money, error) {
	if err := m.Validate(); err != nil {
		return Money{}, err
	}
	if m.amountMinor == math.MinInt64 {
		return Money{}, errors.New("money absolute value overflows int64")
	}
	if m.amountMinor < 0 {
		return Money{amountMinor: -m.amountMinor, currency: m.currency}, nil
	}
	return m, nil
}

func (m Money) Add(other Money) (Money, error) {
	if err := m.sameCurrency(other); err != nil {
		return Money{}, err
	}
	if (other.amountMinor > 0 && m.amountMinor > math.MaxInt64-other.amountMinor) ||
		(other.amountMinor < 0 && m.amountMinor < math.MinInt64-other.amountMinor) {
		return Money{}, errors.New("money addition overflows int64")
	}
	return Money{amountMinor: m.amountMinor + other.amountMinor, currency: m.currency}, nil
}

// MultiplyInt multiplies a monetary amount by an integer quantity while
// retaining its currency and detecting int64 overflow.
func (m Money) MultiplyInt(factor int64) (Money, error) {
	if err := m.Validate(); err != nil {
		return Money{}, err
	}
	product := new(big.Int).Mul(big.NewInt(m.amountMinor), big.NewInt(factor))
	minor, err := int64FromBig(product)
	if err != nil {
		return Money{}, err
	}
	return Money{amountMinor: minor, currency: m.currency}, nil
}

// MultiplyRatio multiplies an amount by an exact integer ratio and rounds
// half away from zero when the result is not an integer minor unit. Keeping
// the ratio in the value object makes conversions such as points/100 explicit
// without exposing division and floating-point rounding to callers.
func (m Money) MultiplyRatio(numerator, denominator int64) (Money, error) {
	if denominator <= 0 {
		return Money{}, errors.New("money ratio denominator must be positive")
	}
	return m.MultiplyRat(new(big.Rat).SetFrac(big.NewInt(numerator), big.NewInt(denominator)))
}

// MultiplyRat multiplies an amount by an exact rational factor and rounds
// half away from zero to the nearest minor unit. The rational is accepted at
// this domain boundary so callers never need to perform floating-point money
// arithmetic themselves.
func (m Money) MultiplyRat(ratio *big.Rat) (Money, error) {
	if err := m.Validate(); err != nil {
		return Money{}, err
	}
	if ratio == nil || ratio.Denom().Sign() <= 0 {
		return Money{}, errors.New("money ratio must be a finite rational")
	}
	product := new(big.Int).Mul(big.NewInt(m.amountMinor), ratio.Num())
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(product, ratio.Denom(), remainder)
	if remainder.Sign() != 0 {
		absRemainder := new(big.Int).Abs(remainder)
		twiceRemainder := new(big.Int).Lsh(absRemainder, 1)
		if twiceRemainder.Cmp(ratio.Denom()) >= 0 {
			if product.Sign() < 0 {
				quotient.Sub(quotient, big.NewInt(1))
			} else {
				quotient.Add(quotient, big.NewInt(1))
			}
		}
	}
	minor, err := int64FromBig(quotient)
	if err != nil {
		return Money{}, err
	}
	return Money{amountMinor: minor, currency: m.currency}, nil
}

// ConvertAtRat converts this amount using an exact captured rational rate. It is used
// when a rate originated from a decimal persistence value and must not be
// converted through another binary floating-point approximation.
func (m Money) ConvertAtRat(rate *big.Rat, targetCurrency string) (Money, error) {
	if err := m.Validate(); err != nil {
		return Money{}, err
	}
	if rate == nil || rate.Sign() <= 0 || rate.Denom().Sign() <= 0 {
		return Money{}, errors.New("money conversion rate must be positive and finite")
	}
	target, err := currency.ParseCode(targetCurrency)
	if err != nil {
		return Money{}, err
	}
	sourceUnits, _ := currency.MinorUnits(m.currency.String())
	targetUnits, _ := currency.MinorUnits(target.String())
	numerator := new(big.Int).Mul(big.NewInt(m.amountMinor), rate.Num())
	numerator.Mul(numerator, big.NewInt(pow10Int64(targetUnits)))
	denominator := new(big.Int).Mul(rate.Denom(), big.NewInt(pow10Int64(sourceUnits)))
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, denominator, remainder)
	if remainder.Sign() != 0 {
		absRemainder := new(big.Int).Abs(remainder)
		twiceRemainder := new(big.Int).Lsh(absRemainder, 1)
		if twiceRemainder.Cmp(denominator) >= 0 {
			if numerator.Sign() < 0 {
				quotient.Sub(quotient, big.NewInt(1))
			} else {
				quotient.Add(quotient, big.NewInt(1))
			}
		}
	}
	minor, err := int64FromBig(quotient)
	if err != nil {
		return Money{}, err
	}
	return Money{amountMinor: minor, currency: target}, nil
}

func (m Money) Subtract(other Money) (Money, error) {
	if err := m.sameCurrency(other); err != nil {
		return Money{}, err
	}
	if other.amountMinor == math.MinInt64 {
		if m.amountMinor >= 0 {
			return Money{}, errors.New("money subtraction overflows int64")
		}
		return Money{amountMinor: m.amountMinor + math.MaxInt64 + 1, currency: m.currency}, nil
	}
	return m.Add(Money{amountMinor: -other.amountMinor, currency: other.currency})
}

// Allocate splits an amount using the largest-remainder method. The returned
// values always retain the original currency and sum exactly to the source
// amount. Ties are resolved by input order, making the result deterministic.
func (m Money) Allocate(ratios []int64) ([]Money, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	if len(ratios) == 0 {
		return nil, ErrInvalidAllocation
	}

	total := new(big.Int)
	for _, ratio := range ratios {
		if ratio < 0 {
			return nil, ErrInvalidAllocation
		}
		total.Add(total, big.NewInt(ratio))
	}
	if total.Sign() <= 0 {
		return nil, ErrInvalidAllocation
	}

	negative := m.amountMinor < 0
	absAmount := new(big.Int).SetInt64(m.amountMinor)
	if negative {
		absAmount.Neg(absAmount)
	}

	allocations := make([]Money, len(ratios))
	remainders := make([]*big.Int, len(ratios))
	allocated := new(big.Int)
	for i, ratio := range ratios {
		product := new(big.Int).Mul(absAmount, big.NewInt(ratio))
		quotient := new(big.Int)
		remainder := new(big.Int)
		quotient.QuoRem(product, total, remainder)
		if negative {
			quotient.Neg(quotient)
		}
		minor, err := int64FromBig(quotient)
		if err != nil {
			return nil, err
		}
		allocations[i] = Money{amountMinor: minor, currency: m.currency}
		allocated.Add(allocated, quotient)
		// Remainders are compared by magnitude. The sign is applied only to
		// the final unit, so negative allocations preserve the same rule.
		remainders[i] = remainder
	}

	delta := new(big.Int).Sub(new(big.Int).SetInt64(m.amountMinor), allocated)
	step := int64(1)
	if delta.Sign() < 0 {
		step = -1
		delta.Neg(delta)
	}
	for delta.Sign() > 0 {
		best := -1
		for i, remainder := range remainders {
			if best == -1 || remainder.Cmp(remainders[best]) > 0 {
				best = i
			}
		}
		if best < 0 {
			return nil, errors.New("money allocation failed to distribute remainder")
		}
		if (step > 0 && allocations[best].amountMinor == math.MaxInt64) ||
			(step < 0 && allocations[best].amountMinor == math.MinInt64) {
			return nil, errors.New("money allocation overflows int64")
		}
		allocations[best].amountMinor += step
		// A remainder receives at most one extra unit. Removing it prevents
		// the same entry from receiving another unit before its peers.
		remainders[best] = big.NewInt(-1)
		delta.Sub(delta, big.NewInt(1))
	}
	return allocations, nil
}

func (m Money) sameCurrency(other Money) error {
	if err := m.Validate(); err != nil {
		return err
	}
	if err := other.Validate(); err != nil {
		return err
	}
	if m.currency != other.currency {
		return fmt.Errorf("%w: %s and %s", ErrCurrencyMismatch, m.currency, other.currency)
	}
	return nil
}

func int64FromBig(value *big.Int) (int64, error) {
	if !value.IsInt64() {
		return 0, errors.New("money amount overflows int64")
	}
	return value.Int64(), nil
}

func pow10Int64(power int) int64 {
	result := int64(1)
	for i := 0; i < power; i++ {
		result *= 10
	}
	return result
}

func allDigits(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return value != ""
}
