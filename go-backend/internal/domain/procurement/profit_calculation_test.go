package procurement

import (
	"errors"
	"testing"
)

func TestCalculateProfitUsesSalePriceAndCalculatesGrossMargin(t *testing.T) {
	salePrice := 90.0
	purchasePrice := 55.0

	result, err := CalculateProfit(ProfitCalculationInput{
		ProductCode:             "RIM-001",
		ProductName:             "Carbon Rim",
		SellingCurrency:         "USD",
		CostCurrency:            "USD",
		ListPrice:               100,
		SalePrice:               &salePrice,
		PurchasePrice:           &purchasePrice,
		InboundShippingUnitCost: 2,
		PackagingUnitCost:       0.5,
		OtherUnitCost:           0,
	})
	if err != nil {
		t.Fatalf("calculate profit: %v", err)
	}

	if result.Status != ProfitStatusReady {
		t.Fatalf("expected ready status, got %q with warnings %#v", result.Status, result.Warnings)
	}
	if result.EffectiveSellingPrice != 90 {
		t.Fatalf("expected effective selling price 90, got %.2f", result.EffectiveSellingPrice)
	}
	if result.LandedCost == nil || *result.LandedCost != 57.5 {
		t.Fatalf("expected landed cost 57.5, got %#v", result.LandedCost)
	}
	if result.GrossProfit == nil || *result.GrossProfit != 32.5 {
		t.Fatalf("expected gross profit 32.5, got %#v", result.GrossProfit)
	}
	if result.GrossMarginBPS == nil || *result.GrossMarginBPS != 3611 {
		t.Fatalf("expected gross margin 3611 bps, got %#v", result.GrossMarginBPS)
	}
	if result.GrossMarginPercent == nil || *result.GrossMarginPercent != 36.11 {
		t.Fatalf("expected gross margin 36.11%%, got %#v", result.GrossMarginPercent)
	}
	if result.InboundShippingUnitCost != 2 ||
		result.PackagingUnitCost != 0.5 ||
		result.OtherUnitCost != 0 {
		t.Fatalf("expected all additional costs to be retained, got %#v", result)
	}
}

func TestCalculateProfitFallsBackToListPriceWhenSalePriceIsMissing(t *testing.T) {
	purchasePrice := 40.0

	result, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "USD",
		ListPrice:       100,
		PurchasePrice:   &purchasePrice,
	})
	if err != nil {
		t.Fatalf("calculate profit: %v", err)
	}

	if result.Status != ProfitStatusWarning {
		t.Fatalf("expected warning status, got %q", result.Status)
	}
	if result.EffectiveSellingPrice != 100 {
		t.Fatalf("expected list-price fallback, got %.2f", result.EffectiveSellingPrice)
	}
	if len(result.Warnings) != 1 || result.Warnings[0] != ProfitWarningSalePriceMissing {
		t.Fatalf("expected sale price warning, got %#v", result.Warnings)
	}
}

func TestCalculateProfitDistinguishesMissingPurchasePriceFromExplicitZero(t *testing.T) {
	missingResult, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "USD",
		ListPrice:       100,
	})
	if err != nil {
		t.Fatalf("calculate missing purchase price: %v", err)
	}
	if missingResult.Status != ProfitStatusMissingPurchase {
		t.Fatalf("expected missing purchase status, got %q", missingResult.Status)
	}
	if missingResult.LandedCost != nil || missingResult.GrossProfit != nil {
		t.Fatalf("missing purchase price must not produce monetary outputs: %#v", missingResult)
	}

	zero := 0.0
	zeroResult, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "USD",
		ListPrice:       100,
		SalePrice:       float64Pointer(90),
		PurchasePrice:   &zero,
	})
	if err != nil {
		t.Fatalf("calculate explicit zero purchase price: %v", err)
	}
	if zeroResult.Status != ProfitStatusReady {
		t.Fatalf("expected explicit zero purchase price to be valid, got %q", zeroResult.Status)
	}
	if zeroResult.LandedCost == nil || *zeroResult.LandedCost != 0 {
		t.Fatalf("expected zero landed cost, got %#v", zeroResult.LandedCost)
	}
	if zeroResult.GrossProfit == nil || *zeroResult.GrossProfit != 90 {
		t.Fatalf("expected gross profit 90, got %#v", zeroResult.GrossProfit)
	}
}

func TestCalculateProfitRejectsCurrencyMismatchWithoutCalculating(t *testing.T) {
	purchasePrice := 55.0

	result, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "USD",
		CostCurrency:    "CNY",
		ListPrice:       100,
		PurchasePrice:   &purchasePrice,
	})
	if err != nil {
		t.Fatalf("calculate currency mismatch: %v", err)
	}

	if result.Status != ProfitStatusCurrencyMismatch {
		t.Fatalf("expected currency mismatch, got %q", result.Status)
	}
	if result.GrossProfit != nil || result.GrossMarginBPS != nil {
		t.Fatalf("currency mismatch must not calculate profit: %#v", result)
	}
}

func TestCalculateProfitRejectsNegativeCosts(t *testing.T) {
	purchasePrice := 55.0

	result, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency:   "USD",
		ListPrice:         100,
		SalePrice:         float64Pointer(90),
		PurchasePrice:     &purchasePrice,
		PackagingUnitCost: -1,
	})
	if err != nil {
		t.Fatalf("calculate negative cost: %v", err)
	}
	if result.Status != ProfitStatusInvalidCost {
		t.Fatalf("expected invalid cost, got %q", result.Status)
	}
	if result.GrossProfit != nil {
		t.Fatalf("invalid cost must not calculate profit: %#v", result.GrossProfit)
	}
}

func TestCalculateProfitReportsNegativeGrossProfit(t *testing.T) {
	purchasePrice := 110.0
	salePrice := 90.0

	result, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "USD",
		ListPrice:       100,
		SalePrice:       &salePrice,
		PurchasePrice:   &purchasePrice,
	})
	if err != nil {
		t.Fatalf("calculate negative gross profit: %v", err)
	}

	if result.Status != ProfitStatusWarning {
		t.Fatalf("expected warning status, got %q", result.Status)
	}
	if result.GrossProfit == nil || *result.GrossProfit != -20 {
		t.Fatalf("expected gross profit -20, got %#v", result.GrossProfit)
	}
	if result.GrossMarginBPS == nil || *result.GrossMarginBPS != -2222 {
		t.Fatalf("expected gross margin -2222 bps, got %#v", result.GrossMarginBPS)
	}
	if !containsProfitWarning(result.Warnings, ProfitWarningNegativeGrossProfit) {
		t.Fatalf("expected negative gross profit warning, got %#v", result.Warnings)
	}
}

func TestCalculateProfitReportsSalePriceAboveListPrice(t *testing.T) {
	salePrice := 110.0
	purchasePrice := 50.0

	result, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "USD",
		ListPrice:       100,
		SalePrice:       &salePrice,
		PurchasePrice:   &purchasePrice,
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
	if result.EffectiveSellingPrice != 110 {
		t.Fatalf("expected effective selling price 110, got %.2f", result.EffectiveSellingPrice)
	}
}

func TestCalculateProfitRoundsAccordingToCurrencyMinorUnits(t *testing.T) {
	purchasePrice := 1000.4

	result, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "JPY",
		ListPrice:       1500.6,
		SalePrice:       float64Pointer(1400.6),
		PurchasePrice:   &purchasePrice,
	})
	if err != nil {
		t.Fatalf("calculate JPY profit: %v", err)
	}

	if result.ListPrice != 1501 {
		t.Fatalf("expected JPY list price 1501, got %.2f", result.ListPrice)
	}
	if result.EffectiveSellingPrice != 1401 {
		t.Fatalf("expected JPY effective price 1401, got %.2f", result.EffectiveSellingPrice)
	}
	if result.PurchasePrice == nil || *result.PurchasePrice != 1000 {
		t.Fatalf("expected JPY purchase price 1000, got %#v", result.PurchasePrice)
	}
	if result.GrossProfit == nil || *result.GrossProfit != 401 {
		t.Fatalf("expected JPY gross profit 401, got %#v", result.GrossProfit)
	}
}

func TestCalculateProfitRejectsUnsupportedCurrency(t *testing.T) {
	_, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "XXX",
		ListPrice:       100,
	})
	if !errors.Is(err, ErrUnsupportedProfitCurrency) {
		t.Fatalf("expected unsupported currency error, got %v", err)
	}
}

func TestCalculateProfitRejectsInvalidSellingPrice(t *testing.T) {
	purchasePrice := 1.0

	result, err := CalculateProfit(ProfitCalculationInput{
		SellingCurrency: "USD",
		ListPrice:       0,
		PurchasePrice:   &purchasePrice,
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
