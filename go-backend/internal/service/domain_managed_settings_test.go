package service

import "testing"

func TestDomainManagedSettingsIncludesPaymentInstallments(t *testing.T) {
	if !IsDomainManagedSettingGroup("payment_installments") {
		t.Fatalf("expected payment_installments group to be domain managed")
	}
	if !IsDomainManagedSettingKey("payment_installments_stripe") {
		t.Fatalf("expected payment_installments_stripe key to be domain managed")
	}
	if IsDomainManagedSettingKey("payment_installments") {
		t.Fatalf("unexpected bare group key to be treated as a setting key")
	}
}
