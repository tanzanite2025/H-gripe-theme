package productsuppliercost

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
)

const (
	ProfitFormulaVersion = "gross-margin-v3-no-customs"

	ProfitStatusReady                = "ready"
	ProfitStatusWarning              = "warning"
	ProfitStatusMissingUnitCost      = "missing_unit_cost"
	ProfitStatusCurrencyMismatch     = "currency_mismatch"
	ProfitStatusInvalidSelling       = "invalid_selling_price"
	ProfitStatusInvalidCost          = "invalid_cost"
	ProfitWarningSalePriceMissing    = "sale_price_missing"
	ProfitWarningSalePriceAboveList  = "sale_price_above_list_price"
	ProfitWarningNegativeGrossProfit = "negative_gross_profit"
)

var (
	ErrUnsupportedProfitCurrency = errors.New("unsupported profitability currency")
	ErrProfitAmountOutOfRange    = errors.New("profitability amount is out of range")
)

// ProfitCalculationInput contains canonical minor-unit catalog price
// snapshots and supplier-cost inputs. It intentionally does not reference
// the product domain or a product database record.
type ProfitCalculationInput struct {
	ProductCode string
	ProductName string

	SellingCurrency string
	CostCurrency    string

	ListPriceMinor               int64
	SalePriceMinor               *int64
	UnitCostMinor                *int64
	InboundShippingUnitCostMinor int64
	PackagingUnitCostMinor       int64
	OtherUnitCostMinor           int64
}

type ProfitCalculationResult struct {
	ProductCode string `json:"product_code"`
	ProductName string `json:"product_name"`

	Currency     string `json:"currency"`
	CostCurrency string `json:"cost_currency"`

	Status         string   `json:"status"`
	FormulaVersion string   `json:"formula_version"`
	Warnings       []string `json:"warnings"`

	ListPriceMinor               int64    `json:"list_price_minor"`
	SalePriceMinor               *int64   `json:"sale_price_minor,omitempty"`
	EffectiveSellingPriceMinor   int64    `json:"effective_selling_price_minor"`
	UnitCostMinor                *int64   `json:"unit_cost_minor,omitempty"`
	InboundShippingUnitCostMinor int64    `json:"inbound_shipping_unit_cost_minor"`
	PackagingUnitCostMinor       int64    `json:"packaging_unit_cost_minor"`
	OtherUnitCostMinor           int64    `json:"other_unit_cost_minor"`
	LandedCostMinor              *int64   `json:"landed_cost_minor,omitempty"`
	GrossProfitMinor             *int64   `json:"gross_profit_minor,omitempty"`
	GrossMarginBPS               *int     `json:"gross_margin_bps,omitempty"`
	GrossMarginPercent           *float64 `json:"gross_margin_percent,omitempty"`

}

// CalculateProfit calculates the current estimated gross profit for one SKU.
//
// The calculation is deliberately self-contained. It does not read products,
// product_variants, supplier-cost records, exchange rates, or any other
// repository. Callers may use the result for a preview or persist it as a
// separate snapshot after validation.
func CalculateProfit(input ProfitCalculationInput) (ProfitCalculationResult, error) {
	result := ProfitCalculationResult{
		ProductCode:    strings.TrimSpace(input.ProductCode),
		ProductName:    strings.TrimSpace(input.ProductName),
		Status:         ProfitStatusReady,
		FormulaVersion: ProfitFormulaVersion,
		Warnings:       []string{},
	}

	sellingCurrency, err := normalizeProfitCurrency(input.SellingCurrency)
	if err != nil {
		return result, err
	}
	costCurrency := input.CostCurrency
	if strings.TrimSpace(costCurrency) == "" {
		costCurrency = sellingCurrency
	}
	costCurrency, err = normalizeProfitCurrency(costCurrency)
	if err != nil {
		return result, err
	}
	result.Currency = sellingCurrency
	result.CostCurrency = costCurrency

	if sellingCurrency != costCurrency {
		result.Status = ProfitStatusCurrencyMismatch
		return result, nil
	}

	listPrice, err := domainmoney.New(input.ListPriceMinor, sellingCurrency)
	if err != nil || listPrice.AmountMinor() <= 0 {
		result.Status = ProfitStatusInvalidSelling
		return result, nil
	}
	result.ListPriceMinor = listPrice.AmountMinor()

	effectiveSellingPrice := listPrice
	if input.SalePriceMinor != nil {
		salePrice, conversionErr := domainmoney.New(*input.SalePriceMinor, sellingCurrency)
		if conversionErr != nil || salePrice.AmountMinor() <= 0 {
			result.Status = ProfitStatusInvalidSelling
			return result, nil
		}
		normalizedSalePrice := salePrice.AmountMinor()
		result.SalePriceMinor = &normalizedSalePrice
		effectiveSellingPrice = salePrice
		if salePrice.AmountMinor() > listPrice.AmountMinor() {
			addProfitWarning(&result, ProfitWarningSalePriceAboveList)
		}
	} else {
		addProfitWarning(&result, ProfitWarningSalePriceMissing)
	}
	result.EffectiveSellingPriceMinor = effectiveSellingPrice.AmountMinor()

	costs := []int64{
		input.InboundShippingUnitCostMinor,
		input.PackagingUnitCostMinor,
		input.OtherUnitCostMinor,
	}
	costValues := make([]domainmoney.Money, 0, len(costs))
	for _, cost := range costs {
		valueMoney, conversionErr := domainmoney.New(cost, sellingCurrency)
		if conversionErr != nil || valueMoney.AmountMinor() < 0 {
			result.Status = ProfitStatusInvalidCost
			return result, nil
		}
		costValues = append(costValues, valueMoney)
	}
	result.InboundShippingUnitCostMinor = costValues[0].AmountMinor()
	result.PackagingUnitCostMinor = costValues[1].AmountMinor()
	result.OtherUnitCostMinor = costValues[2].AmountMinor()

	if input.UnitCostMinor == nil {
		result.Status = ProfitStatusMissingUnitCost
		return result, nil
	}
	unitCost, err := domainmoney.New(*input.UnitCostMinor, sellingCurrency)
	if err != nil || unitCost.AmountMinor() < 0 {
		result.Status = ProfitStatusInvalidCost
		return result, nil
	}
	normalizedUnitCost := unitCost.AmountMinor()
	result.UnitCostMinor = &normalizedUnitCost

	landedCost := unitCost
	for _, cost := range costValues {
		landedCost, err = landedCost.Add(cost)
		if err != nil {
			return result, ErrProfitAmountOutOfRange
		}
	}
	landedCostAmount := landedCost.AmountMinor()
	result.LandedCostMinor = &landedCostAmount

	grossProfit, err := effectiveSellingPrice.Subtract(landedCost)
	if err != nil {
		return result, ErrProfitAmountOutOfRange
	}
	grossProfitAmount := grossProfit.AmountMinor()
	result.GrossProfitMinor = &grossProfitAmount

	marginBPS, err := roundedProfitRatio(grossProfit.AmountMinor(), effectiveSellingPrice.AmountMinor(), 10000)
	if err != nil {
		return result, ErrProfitAmountOutOfRange
	}
	marginPercent := float64(marginBPS) / 100
	result.GrossMarginBPS = intPointer(marginBPS)
	result.GrossMarginPercent = float64Pointer(marginPercent)

	if grossProfit.AmountMinor() < 0 {
		addProfitWarning(&result, ProfitWarningNegativeGrossProfit)
	}
	if len(result.Warnings) > 0 {
		result.Status = ProfitStatusWarning
	}

	return result, nil
}

func normalizeProfitCurrency(value string) (string, error) {
	code := currency.NormalizeCode(value)
	if code == "" {
		code = DefaultCurrency
	}
	if !currency.IsCatalogCode(code) {
		return "", fmt.Errorf("%w: %s", ErrUnsupportedProfitCurrency, code)
	}
	return code, nil
}

func roundedProfitRatio(numerator, denominator, scale int64) (int, error) {
	if denominator <= 0 || scale <= 0 {
		return 0, ErrProfitAmountOutOfRange
	}
	n := new(big.Int).Mul(big.NewInt(numerator), big.NewInt(scale))
	d := big.NewInt(denominator)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(n, d, remainder)
	if remainder.Sign() != 0 {
		absRemainder := new(big.Int).Abs(remainder)
		twiceRemainder := new(big.Int).Lsh(absRemainder, 1)
		if twiceRemainder.Cmp(d) >= 0 {
			if n.Sign() < 0 {
				quotient.Sub(quotient, big.NewInt(1))
			} else {
				quotient.Add(quotient, big.NewInt(1))
			}
		}
	}
	if !quotient.IsInt64() {
		return 0, ErrProfitAmountOutOfRange
	}
	return int(quotient.Int64()), nil
}

func addProfitWarning(result *ProfitCalculationResult, warning string) {
	for _, existing := range result.Warnings {
		if existing == warning {
			return
		}
	}
	result.Warnings = append(result.Warnings, warning)
}

func float64Pointer(value float64) *float64 {
	return &value
}

func intPointer(value int) *int {
	return &value
}
