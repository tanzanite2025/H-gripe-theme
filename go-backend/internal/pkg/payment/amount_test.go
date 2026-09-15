package payment

import (
	"testing"

	domainmoney "commerce-platform/internal/domain/money"
)

func TestPaymentMajorStringRoundsBeforeFormatting(t *testing.T) {
	got, err := paymentMajorString(199.99999999999997, "USD")
	if err != nil || got != "200.00" {
		t.Fatalf("formatted amount = %q, err=%v; want 200.00", got, err)
	}
}

func TestPaymentMajorStringUsesZeroDecimalCurrency(t *testing.T) {
	got, err := paymentMajorString(1000.4, "JPY")
	if err != nil || got != "1000" {
		t.Fatalf("formatted amount = %q, err=%v; want 1000", got, err)
	}
}

func TestMoneyParserDoesNotUseBinaryFloatingPoint(t *testing.T) {
	money, err := domainmoney.ParseMajor("199.99", "USD")
	if err != nil || money.AmountMinor() != 19999 {
		t.Fatalf("ParseMajor() = %d, err=%v; want 19999", money.AmountMinor(), err)
	}

	money, err = domainmoney.ParseMajor("123", "JPY")
	if err != nil || money.AmountMinor() != 123 {
		t.Fatalf("ParseMajor(JPY) = %d, err=%v; want 123", money.AmountMinor(), err)
	}
}

func TestMoneyParserRejectsExcessPrecision(t *testing.T) {
	if _, err := domainmoney.ParseMajor("1.001", "USD"); err == nil {
		t.Fatal("expected excess precision error")
	}
}

func TestMoneyFormatterUsesCurrencyScale(t *testing.T) {
	money, err := domainmoney.New(19999, "USD")
	formatted, formatErr := money.FormatMajor()
	if err != nil || formatErr != nil || formatted != "199.99" {
		t.Fatalf("Money.FormatMajor() = %q, err=%v; want 199.99", formatted, formatErr)
	}

	money, err = domainmoney.New(123, "JPY")
	formatted, formatErr = money.FormatMajor()
	if err != nil || formatErr != nil || formatted != "123" {
		t.Fatalf("Money.FormatMajor(JPY) = %q, err=%v; want 123", formatted, formatErr)
	}
}
