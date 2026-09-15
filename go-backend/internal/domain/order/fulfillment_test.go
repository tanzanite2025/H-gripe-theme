package order

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/money"
)

func TestResolveFulfillmentModeDefaultsLegacyItemsToStock(t *testing.T) {
	mode := ResolveFulfillmentMode([]OrderItem{{FulfillmentMode: ""}})
	if mode != FulfillmentModeStock {
		t.Fatalf("ResolveFulfillmentMode() = %q, want %q", mode, FulfillmentModeStock)
	}
}

func TestResolveFulfillmentModeDetectsMixedOrders(t *testing.T) {
	mode := ResolveFulfillmentMode([]OrderItem{
		{FulfillmentMode: FulfillmentModeStock},
		{FulfillmentMode: FulfillmentModeMadeToOrder},
	})
	if mode != FulfillmentModeMixed {
		t.Fatalf("ResolveFulfillmentMode() = %q, want %q", mode, FulfillmentModeMixed)
	}
}

func TestDefaultProductionStatus(t *testing.T) {
	if got := DefaultProductionStatus(FulfillmentModeStock); got != ProductionStatusNotApplicable {
		t.Fatalf("stock production status = %q", got)
	}
	if got := DefaultProductionStatus(FulfillmentModeMadeToOrder); got != ProductionStatusNotStarted {
		t.Fatalf("made-to-order production status = %q", got)
	}
}

func TestResolveSignatureRequiredUsesUSDOrderFXSnapshot(t *testing.T) {
	tests := []struct {
		name         string
		totalAmount  float64
		fxSnapshot   currency.OrderFXSnapshot
		wantRequired bool
	}{
		{
			name:         "USD just below threshold",
			totalAmount:  749.99,
			fxSnapshot:   testOrderFXSnapshot("USD", "USD", 1),
			wantRequired: false,
		},
		{
			name:         "USD at threshold",
			totalAmount:  HighValueSignatureThresholdUSD,
			fxSnapshot:   testOrderFXSnapshot("USD", "USD", 1),
			wantRequired: true,
		},
		{
			name:         "USD above threshold",
			totalAmount:  750.01,
			fxSnapshot:   testOrderFXSnapshot("USD", "USD", 1),
			wantRequired: true,
		},
		{
			name:         "foreign currency at converted threshold",
			totalAmount:  675,
			fxSnapshot:   testOrderFXSnapshot("USD", "EUR", 0.9),
			wantRequired: true,
		},
		{
			name:         "foreign currency just below converted threshold",
			totalAmount:  674.99,
			fxSnapshot:   testOrderFXSnapshot("USD", "EUR", 0.9),
			wantRequired: false,
		},
		{
			name:         "foreign currency above converted threshold",
			totalAmount:  675.01,
			fxSnapshot:   testOrderFXSnapshot("USD", "EUR", 0.9),
			wantRequired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalMoney, err := money.FromMajorFloat(tt.totalAmount, tt.fxSnapshot.OrderCurrency)
			if err != nil {
				t.Fatalf("parse test amount: %v", err)
			}
			if got := ResolveSignatureRequired(totalMoney, tt.fxSnapshot); got != tt.wantRequired {
				t.Fatalf("ResolveSignatureRequired(%v, %#v) = %v, want %v",
					tt.totalAmount,
					tt.fxSnapshot,
					got,
					tt.wantRequired,
				)
			}
		})
	}
}

func TestResolveSignatureRequiredRejectsNonUSDOrInvalidSnapshots(t *testing.T) {
	nonUSDBase := currency.OrderFXSnapshot{
		Version:         currency.OrderFXSnapshotVersion,
		BaseCurrency:    "CNY",
		OrderCurrency:   "CNY",
		BaseToOrderRate: 1,
		Source:          "test",
		CapturedAt:      time.Now().UTC(),
	}
	if ResolveSignatureRequired(money.MustNew(75000, "CNY"), nonUSDBase) {
		t.Fatal("non-USD policy snapshots must not be treated as USD thresholds")
	}
	if ResolveSignatureRequired(money.MustNew(100000, "USD"), currency.OrderFXSnapshot{}) {
		t.Fatal("invalid FX snapshots must not require signature")
	}
}

func testOrderFXSnapshot(baseCurrency, orderCurrency string, baseToOrderRate float64) currency.OrderFXSnapshot {
	return currency.OrderFXSnapshot{
		Version:         currency.OrderFXSnapshotVersion,
		BaseCurrency:    baseCurrency,
		OrderCurrency:   orderCurrency,
		BaseToOrderRate: baseToOrderRate,
		Source:          "test",
		CapturedAt:      time.Now().UTC(),
	}
}
