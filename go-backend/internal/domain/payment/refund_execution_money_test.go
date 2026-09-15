package payment

import "testing"

func TestPaymentRefundExecutionAmountMoneyPrefersMinorSnapshot(t *testing.T) {
	execution := PaymentRefundExecution{AmountMinor: 1234, Amount: 99, Currency: "USD"}
	amount, err := execution.AmountMoney()
	if err != nil {
		t.Fatalf("amount money: %v", err)
	}
	if amount.AmountMinor() != 1234 {
		t.Fatalf("expected 1234 minor units, got %d", amount.AmountMinor())
	}
}

func TestPaymentRefundExecutionBeforeSaveBackfillsMinorAmount(t *testing.T) {
	execution := PaymentRefundExecution{Amount: 12.34, Currency: "USD"}
	if err := execution.BeforeSave(nil); err != nil {
		t.Fatalf("before save: %v", err)
	}
	if execution.AmountMinor != 1234 {
		t.Fatalf("expected 1234 minor units, got %d", execution.AmountMinor)
	}
}
