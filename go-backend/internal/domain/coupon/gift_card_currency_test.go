package coupon

import (
	"testing"

	domainmoney "commerce-platform/internal/domain/money"
)

func TestGiftCardMinorUnitsRespectZeroDecimalCurrencies(t *testing.T) {
	jpy, err := domainmoney.FromMajorFloat(10000, "JPY")
	if err != nil || jpy.AmountMinor() != 10000 {
		t.Fatalf("JPY minor units = %d, error %v; want 10000", jpy.AmountMinor(), err)
	}
	jpyMajor, err := jpy.MajorFloat()
	if err != nil || jpyMajor != 10000 {
		t.Fatalf("JPY major amount = %v, error %v; want 10000", jpyMajor, err)
	}
	usd, err := domainmoney.FromMajorFloat(10.25, "USD")
	if err != nil || usd.AmountMinor() != 1025 {
		t.Fatalf("USD minor units = %d, error %v; want 1025", usd.AmountMinor(), err)
	}
}

func TestGiftCardMoneyCarriesLedgerCurrency(t *testing.T) {
	card := GiftCard{InitialCents: 1234, BalanceCents: 567, Currency: "USD"}
	initial, err := card.InitialMoney()
	if err != nil {
		t.Fatalf("InitialMoney() error = %v", err)
	}
	if initial.AmountMinor() != 1234 || initial.Currency().String() != "USD" {
		t.Fatalf("InitialMoney() = %#v; want USD 1234", initial)
	}
	balance, err := card.BalanceMoney()
	if err != nil {
		t.Fatalf("BalanceMoney() error = %v", err)
	}
	if balance.AmountMinor() != 567 {
		t.Fatalf("BalanceMoney() = %d; want 567", balance.AmountMinor())
	}
}

func TestGiftCardMoneyRejectsUnsupportedCurrency(t *testing.T) {
	_, err := (GiftCard{Currency: "XXX"}).BalanceMoney()
	if err == nil {
		t.Fatal("expected unsupported currency error")
	}
}

func TestGiftCardMoneyUsesMinorUnitsWithoutFloatAdapter(t *testing.T) {
	card := GiftCard{InitialCents: 123, BalanceCents: 45, Currency: "JPY"}
	initial, err := card.InitialMoney()
	if err != nil || initial.AmountMinor() != 123 {
		t.Fatalf("InitialMoney() = %v, error %v; want 123", initial.AmountMinor(), err)
	}
	balance, err := card.BalanceMoney()
	if err != nil || balance.AmountMinor() != 45 {
		t.Fatalf("BalanceMoney() = %v, error %v; want 45", balance.AmountMinor(), err)
	}
}
