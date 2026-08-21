package setting

import "testing"

func TestPaymentProviderInstallmentsSettingsRoundTrip(t *testing.T) {
	request := PaymentProviderInstallmentsUpdateRequest{
		Enabled:            true,
		PaymentMethodTypes: []string{"klarna", "card", "klana", "card"},
		Countries:          []string{"us", "gb", "US"},
		Currencies:         []string{"usd", "eur", "usd"},
		MinAmount:          100,
		MaxAmount:          5000,
		Notes:              "  test note  ",
	}
	settings := request.Settings("stripe")
	if settings.Provider != "stripe" {
		t.Fatalf("expected provider stripe, got %s", settings.Provider)
	}
	if len(settings.PaymentMethodTypes) != 2 || settings.PaymentMethodTypes[0] != "card" || settings.PaymentMethodTypes[1] != "klarna" {
		t.Fatalf("unexpected payment method types: %#v", settings.PaymentMethodTypes)
	}
	if len(settings.Countries) != 2 || settings.Countries[0] != "US" || settings.Countries[1] != "GB" {
		t.Fatalf("unexpected countries: %#v", settings.Countries)
	}
	if len(settings.Currencies) != 2 || settings.Currencies[0] != "USD" || settings.Currencies[1] != "EUR" {
		t.Fatalf("unexpected currencies: %#v", settings.Currencies)
	}
	if settings.MinAmount != 100 || settings.MaxAmount != 5000 {
		t.Fatalf("unexpected amount thresholds: min=%v max=%v", settings.MinAmount, settings.MaxAmount)
	}
	if settings.Notes != "test note" {
		t.Fatalf("unexpected notes: %q", settings.Notes)
	}

	value, err := settings.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}
	parsed, err := PaymentProviderInstallmentsSettingsFromValue("stripe", value)
	if err != nil {
		t.Fatalf("PaymentProviderInstallmentsSettingsFromValue() error = %v", err)
	}
	if !parsed.Enabled || len(parsed.PaymentMethodTypes) != 2 || parsed.MinAmount != 100 || parsed.MaxAmount != 5000 {
		t.Fatalf("unexpected parsed settings: %#v", parsed)
	}
}
