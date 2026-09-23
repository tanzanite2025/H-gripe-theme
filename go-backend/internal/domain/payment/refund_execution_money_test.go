package payment

import "testing"

func TestPaymentRefundExecutionAmountMoneyPrefersMinorSnapshot(t *testing.T) {
	execution := PaymentRefundExecution{AmountMinor: 1234, Currency: "USD"}
	amount, err := execution.AmountMoney()
	if err != nil {
		t.Fatalf("amount money: %v", err)
	}
	if amount.AmountMinor() != 1234 {
		t.Fatalf("expected 1234 minor units, got %d", amount.AmountMinor())
	}
}

func TestPaymentRefundExecutionBeforeSaveRequiresMinorAmount(t *testing.T) {
	execution := PaymentRefundExecution{Currency: "USD"}
	if err := execution.BeforeSave(nil); err != nil {
		t.Fatalf("before save: %v", err)
	}
	if execution.AmountMinor != 0 {
		t.Fatalf("legacy major amount must not populate canonical amount, got %d", execution.AmountMinor)
	}
}
