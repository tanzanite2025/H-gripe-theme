package procurement

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"commerce-platform/internal/domain/currency"
)

const (
	ProfitFormulaVersion = "gross-margin-v3-no-customs"

	ProfitStatusReady                = "ready"
	ProfitStatusWarning              = "warning"
	ProfitStatusMissingPurchase      = "missing_purchase_price"
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

// ProfitCalculationInput contains catalog price snapshots and procurement
// inputs. It intentionally uses plain values and does not reference the
// product domain or a product database record.
type ProfitCalculationInput struct {
	ProductCode string
	ProductName string

	SellingCurrency string
	CostCurrency    string

	ListPrice     float64
	SalePrice     *float64
	PurchasePrice *float64

	InboundShippingUnitCost float64
	PackagingUnitCost       float64
	OtherUnitCost           float64
}

type ProfitCalculationResult struct {
	ProductCode string `json:"product_code"`
	ProductName string `json:"product_name"`

	Currency     string `json:"currency"`
	CostCurrency string `json:"cost_currency"`

	Status         string   `json:"status"`
	FormulaVersion string   `json:"formula_version"`
	Warnings       []string `json:"warnings"`

	ListPrice             float64  `json:"list_price"`
	SalePrice             *float64 `json:"sale_price,omitempty"`
	EffectiveSellingPrice float64  `json:"effective_selling_price"`

	PurchasePrice           *float64 `json:"purchase_price,omitempty"`
	InboundShippingUnitCost float64  `json:"inbound_shipping_unit_cost"`
	PackagingUnitCost       float64  `json:"packaging_unit_cost"`
	OtherUnitCost           float64  `json:"other_unit_cost"`

	LandedCost         *float64 `json:"landed_cost,omitempty"`
	GrossProfit        *float64 `json:"gross_profit,omitempty"`
	GrossMarginBPS     *int     `json:"gross_margin_bps,omitempty"`
	GrossMarginPercent *float64 `json:"gross_margin_percent,omitempty"`
}

// CalculateProfit calculates the current estimated gross profit for one SKU.
//
// The calculation is deliberately self-contained. It does not read products,
// product_variants, procurement records, exchange rates, or any other
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

	minorUnits, ok := currency.MinorUnits(sellingCurrency)
	if !ok {
		return result, fmt.Errorf("%w: %s", ErrUnsupportedProfitCurrency, sellingCurrency)
	}

	listPriceMinor, err := majorToMinor(input.ListPrice, minorUnits)
	if err != nil || listPriceMinor <= 0 {
		result.Status = ProfitStatusInvalidSelling
		return result, nil
	}
	result.ListPrice = minorToMajor(listPriceMinor, minorUnits)

	effectiveSellingPriceMinor := listPriceMinor
	if input.SalePrice != nil {
		salePriceMinor, conversionErr := majorToMinor(*input.SalePrice, minorUnits)
		if conversionErr != nil || salePriceMinor <= 0 {
			result.Status = ProfitStatusInvalidSelling
			return result, nil
		}
		normalizedSalePrice := minorToMajor(salePriceMinor, minorUnits)
		result.SalePrice = float64Pointer(normalizedSalePrice)
		effectiveSellingPriceMinor = salePriceMinor
		if salePriceMinor > listPriceMinor {
			addProfitWarning(&result, ProfitWarningSalePriceAboveList)
		}
	} else {
		addProfitWarning(&result, ProfitWarningSalePriceMissing)
	}
	result.EffectiveSellingPrice = minorToMajor(effectiveSellingPriceMinor, minorUnits)

	costs := []float64{
		input.InboundShippingUnitCost,
		input.PackagingUnitCost,
		input.OtherUnitCost,
	}
	costMinorValues := make([]int64, 0, len(costs))
	for _, cost := range costs {
		valueMinor, conversionErr := majorToMinor(cost, minorUnits)
		if conversionErr != nil || valueMinor < 0 {
			result.Status = ProfitStatusInvalidCost
			return result, nil
		}
		costMinorValues = append(costMinorValues, valueMinor)
	}
	result.InboundShippingUnitCost = minorToMajor(costMinorValues[0], minorUnits)
	result.PackagingUnitCost = minorToMajor(costMinorValues[1], minorUnits)
	result.OtherUnitCost = minorToMajor(costMinorValues[2], minorUnits)

	if input.PurchasePrice == nil {
		result.Status = ProfitStatusMissingPurchase
		return result, nil
	}
	purchasePriceMinor, err := majorToMinor(*input.PurchasePrice, minorUnits)
	if err != nil || purchasePriceMinor < 0 {
		result.Status = ProfitStatusInvalidCost
		return result, nil
	}
	normalizedPurchasePrice := minorToMajor(purchasePriceMinor, minorUnits)
	result.PurchasePrice = float64Pointer(normalizedPurchasePrice)

	landedCostMinor := purchasePriceMinor
	for _, costMinor := range costMinorValues {
		if costMinor > math.MaxInt64-landedCostMinor {
			return result, ErrProfitAmountOutOfRange
		}
		landedCostMinor += costMinor
	}
	landedCost := minorToMajor(landedCostMinor, minorUnits)
	result.LandedCost = float64Pointer(landedCost)

	grossProfitMinor := effectiveSellingPriceMinor - landedCostMinor
	grossProfit := minorToMajor(grossProfitMinor, minorUnits)
	result.GrossProfit = float64Pointer(grossProfit)

	marginBPS := int(math.Round(float64(grossProfitMinor) / float64(effectiveSellingPriceMinor) * 10000))
	marginPercent := float64(marginBPS) / 100
	result.GrossMarginBPS = intPointer(marginBPS)
	result.GrossMarginPercent = float64Pointer(marginPercent)

	if grossProfitMinor < 0 {
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

func majorToMinor(value float64, minorUnits int) (int64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0, ErrProfitAmountOutOfRange
	}

	scale := math.Pow10(minorUnits)
	scaled := math.Round(value * scale)
	if math.IsInf(scaled, 0) || scaled > float64(math.MaxInt64) {
		return 0, ErrProfitAmountOutOfRange
	}
	return int64(scaled), nil
}

func minorToMajor(value int64, minorUnits int) float64 {
	return float64(value) / math.Pow10(minorUnits)
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
