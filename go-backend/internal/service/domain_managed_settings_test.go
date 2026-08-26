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

func TestDomainManagedSettingsIncludesWebsiteNameKeys(t *testing.T) {
	if !IsDomainManagedSettingKey("website_name_title") {
		t.Fatalf("expected website_name_title to be domain managed")
	}
	if !IsDomainManagedSettingKey("WEBSITE_NAME_BODY") {
		t.Fatalf("expected website_name_body to be matched case-insensitively")
	}
	if IsDomainManagedSettingKey("website_name") {
		t.Fatalf("unexpected bare website_name key to be treated as a setting key")
	}
}
