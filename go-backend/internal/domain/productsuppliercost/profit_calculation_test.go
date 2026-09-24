package productsuppliercost

import (
	"errors"
	"testing"
)

func TestCalculateProfitUsesSalePriceAndCalculatesGrossMargin(t *testing.T) {
	salePrice := int64(9000)
	unitCost := int64(5500)

	result, err := CalculateProfit(ProfitCalculationInput{
		ProductCode:                  "RIM-001",
		ProductName:                  "Carbon Rim",
		SellingCurrency:              "USD",
		CostCurrency:                 "USD",
		ListPriceMinor:               10000,
		SalePriceMinor:               &salePrice,
		UnitCostMinor:                &unitCost,
		InboundShippingUnitCostMinor: 200,
		PackagingUnitCostMinor:       50,
		OtherUnitCostMinor:           0,
	})
	if err != nil {
		t.Fatalf("calculate profit: %v", err)
	}

	if result.Status != ProfitStatusReady {
		t.Fatalf("expected ready status, got %q with warnings %#v", result.Status, result.Warnings)
	}
	if result.EffectiveSellingPriceMinor != 9000 {
		t.Fatalf("expected effective selling price 9000, got %d", result.EffectiveSellingPriceMinor)
	}
	if result.LandedCostMinor == nil || *result.LandedCostMinor != 5750 {
		t.Fatalf("expected landed cost 5750, got %#v", result.LandedCostMinor)
	}
	if result.GrossProfitMinor == nil || *result.GrossProfitMinor != 3250 {
		t.Fatalf("expected gross profit 3250, got %#v", result.GrossProfitMinor)
	}
	if result.GrossMarginBPS == nil || *result.GrossMarginBPS != 3611 {
		t.Fatalf("expected gross margin 3611 bps, got %#v", result.GrossMarginBPS)
	}
	if result.GrossMarginPercent == nil || *result.GrossMarginPercent != 36.11 {
		t.Fatalf("expected gross margin 36.11%%, got %#v", result.GrossMarginPercent)
	}
	if result.InboundShippingUnitCostMinor != 200 ||
		result.PackagingUnitCostMinor != 50 ||
		result.OtherUnitCostMinor != 0 {
		t.Fatalf("expected all additional costs to be retained, got %#v", result)
	}
}

func TestCalculateProfitFallsBackToListPriceWhenSalePriceIsMissing(t *testing.T) {
	unitCost := int64(4000)

	result, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "USD",
		ListPriceMinor:  10000,
		UnitCostMinor:   &unitCost,
	})
	if err != nil {
		t.Fatalf("calculate profit: %v", err)
	}

	if result.Status != ProfitStatusWarning {
		t.Fatalf("expected warning status, got %q", result.Status)
	}
	if result.EffectiveSellingPriceMinor != 10000 {
		t.Fatalf("expected list-price fallback, got %d", result.EffectiveSellingPriceMinor)
	}
	if len(result.Warnings) != 1 || result.Warnings[0] != ProfitWarningSalePriceMissing {
		t.Fatalf("expected sale price warning, got %#v", result.Warnings)
	}
}

func TestCalculateProfitDistinguishesMissingUnitCostFromExplicitZero(t *testing.T) {
	missingResult, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "USD",
		ListPriceMinor:  10000,
	})
	if err != nil {
		t.Fatalf("calculate missing unit cost: %v", err)
	}
	if missingResult.Status != ProfitStatusMissingUnitCost {
		t.Fatalf("expected missing unit cost status, got %q", missingResult.Status)
	}
	if missingResult.LandedCostMinor != nil || missingResult.GrossProfitMinor != nil {
		t.Fatalf("missing unit cost must not produce monetary outputs: %#v", missingResult)
	}

	zero := int64(0)
	zeroResult, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "USD",
		ListPriceMinor:  10000,
		SalePriceMinor:  func() *int64 { value := int64(9000); return &value }(),
		UnitCostMinor:   &zero,
	})
	if err != nil {
		t.Fatalf("calculate explicit zero unit cost: %v", err)
	}
	if zeroResult.Status != ProfitStatusReady {
		t.Fatalf("expected explicit zero unit cost to be valid, got %q", zeroResult.Status)
	}
	if zeroResult.LandedCostMinor == nil || *zeroResult.LandedCostMinor != 0 {
		t.Fatalf("expected zero landed cost, got %#v", zeroResult.LandedCostMinor)
	}
	if zeroResult.GrossProfitMinor == nil || *zeroResult.GrossProfitMinor != 9000 {
		t.Fatalf("expected gross profit 9000, got %#v", zeroResult.GrossProfitMinor)
	}
}

func TestCalculateProfitRejectsCurrencyMismatchWithoutCalculating(t *testing.T) {
	unitCost := int64(5500)

	result, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "USD",
		CostCurrency:    "CNY",
		ListPriceMinor:  10000,
		UnitCostMinor:   &unitCost,
	})
	if err != nil {
		t.Fatalf("calculate currency mismatch: %v", err)
	}

	if result.Status != ProfitStatusCurrencyMismatch {
		t.Fatalf("expected currency mismatch, got %q", result.Status)
	}
	if result.GrossProfitMinor != nil || result.GrossMarginBPS != nil {
		t.Fatalf("currency mismatch must not calculate profit: %#v", result)
	}
}

func TestCalculateProfitRejectsNegativeCosts(t *testing.T) {
	unitCost := int64(5500)

	result, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency:        "USD",
		ListPriceMinor:         10000,
		SalePriceMinor:         func() *int64 { value := int64(9000); return &value }(),
		UnitCostMinor:          &unitCost,
		PackagingUnitCostMinor: -1,
	})
	if err != nil {
		t.Fatalf("calculate negative cost: %v", err)
	}
	if result.Status != ProfitStatusInvalidCost {
		t.Fatalf("expected invalid cost, got %q", result.Status)
	}
	if result.GrossProfitMinor != nil {
		t.Fatalf("invalid cost must not calculate profit: %#v", result.GrossProfitMinor)
	}
}

func TestCalculateProfitReportsNegativeGrossProfit(t *testing.T) {
	unitCost := int64(11000)
	salePrice := int64(9000)

	result, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "USD",
		ListPriceMinor:  10000,
		SalePriceMinor:  &salePrice,
		UnitCostMinor:   &unitCost,
	})
	if err != nil {
		t.Fatalf("calculate negative gross profit: %v", err)
	}

	if result.Status != ProfitStatusWarning {
		t.Fatalf("expected warning status, got %q", result.Status)
	}
	if result.GrossProfitMinor == nil || *result.GrossProfitMinor != -2000 {
		t.Fatalf("expected gross profit -2000, got %#v", result.GrossProfitMinor)
	}
	if result.GrossMarginBPS == nil || *result.GrossMarginBPS != -2222 {
		t.Fatalf("expected gross margin -2222 bps, got %#v", result.GrossMarginBPS)
	}
	if !containsProfitWarning(result.Warnings, ProfitWarningNegativeGrossProfit) {
		t.Fatalf("expected negative gross profit warning, got %#v", result.Warnings)
	}
}

func TestCalculateProfitReportsSalePriceAboveListPrice(t *testing.T) {
	salePrice := int64(11000)
	unitCost := int64(5000)

	result, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "USD",
		ListPriceMinor:  10000,
		SalePriceMinor:  &salePrice,
		UnitCostMinor:   &unitCost,
	})
	if err != nil {
		t.Fatalf("calculate sale price above list price: %v", err)
	}

	if result.Status != ProfitStatusWarning {
		t.Fatalf("expected warning status, got %q", result.Status)
	}
	if !containsProfitWarning(result.Warnings, ProfitWarningSalePriceAboveList) {
		t.Fatalf("expected sale price warning, got %#v", result.Warnings)
	}
	if result.EffectiveSellingPriceMinor != 11000 {
		t.Fatalf("expected effective selling price 11000, got %d", result.EffectiveSellingPriceMinor)
	}
}

func TestCalculateProfitRoundsAccordingToCurrencyMinorUnits(t *testing.T) {
	unitCost := int64(1000)

	result, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "JPY",
		ListPriceMinor:  1501,
		SalePriceMinor:  func() *int64 { value := int64(1401); return &value }(),
		UnitCostMinor:   &unitCost,
	})
	if err != nil {
		t.Fatalf("calculate JPY profit: %v", err)
	}

	if result.ListPriceMinor != 1501 {
		t.Fatalf("expected JPY list price 1501, got %d", result.ListPriceMinor)
	}
	if result.EffectiveSellingPriceMinor != 1401 {
		t.Fatalf("expected JPY effective price 1401, got %d", result.EffectiveSellingPriceMinor)
	}
	if result.UnitCostMinor == nil || *result.UnitCostMinor != 1000 {
		t.Fatalf("expected JPY unit cost 1000, got %#v", result.UnitCostMinor)
	}
	if result.GrossProfitMinor == nil || *result.GrossProfitMinor != 401 {
		t.Fatalf("expected JPY gross profit 401, got %#v", result.GrossProfitMinor)
	}
}

func TestCalculateProfitRejectsUnsupportedCurrency(t *testing.T) {
	_, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "XXX",
		ListPriceMinor:  10000,
	})
	if !errors.Is(err, ErrUnsupportedProfitCurrency) {
		t.Fatalf("expected unsupported currency error, got %v", err)
	}
}

func TestCalculateProfitRejectsInvalidSellingPrice(t *testing.T) {
	unitCost := int64(100)

	result, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "USD",
		ListPriceMinor:  0,
		UnitCostMinor:   &unitCost,
	})
	if err != nil {
		t.Fatalf("calculate invalid selling price: %v", err)
	}
	if result.Status != ProfitStatusInvalidSelling {
		t.Fatalf("expected invalid selling price, got %q", result.Status)
	}
}

func containsProfitWarning(warnings []string, expected string) bool {
	for _, warning := range warnings {
		if warning == expected {
			return true
		}
	}
	return false
}
