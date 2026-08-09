package admin

import (
	"encoding/json"
	"strings"
	"testing"

	"tanzanite/internal/domain/currency"
	shippingdomain "tanzanite/internal/domain/shipping"
)

func TestTrackingProviderResponseRedactsSecrets(t *testing.T) {
	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode:  "17TRACK",
		ProviderName:  "17TRACK",
		APIKey:        "real-api-key",
		WebhookSecret: "real-webhook-secret",
	}

	payload, err := json.Marshal(trackingProviderResponse(provider))
	if err != nil {
		t.Fatalf("marshal tracking provider response: %v", err)
	}

	body := string(payload)
	if strings.Contains(body, provider.APIKey) || strings.Contains(body, provider.WebhookSecret) {
		t.Fatalf("expected response to redact secrets, got %s", body)
	}
	if !strings.Contains(body, `"api_key_configured":true`) {
		t.Fatalf("expected api_key_configured flag, got %s", body)
	}
	if !strings.Contains(body, `"webhook_secret_configured":true`) {
		t.Fatalf("expected webhook_secret_configured flag, got %s", body)
	}
}

func TestTrackingCarrierMappingRequestPreservesProviderCarrierCode(t *testing.T) {
	req := shippingTrackingCarrierMappingRequest{
		ProviderID:          1,
		Scope:               "carrier",
		ProviderCarrierCode: "usps-lower-190271",
	}

	mapping := req.toDomain()

	if mapping.ProviderCarrierCode != req.ProviderCarrierCode {
		t.Fatalf("expected provider carrier code to be preserved, got %q", mapping.ProviderCarrierCode)
	}
}

func TestShippingTemplateRequestPersistsDisplayPriceSnapshotsByMoneyField(t *testing.T) {
	req := shippingTemplateRequest{
		Name:       "Display priced shipping",
		Type:       "price",
		DefaultFee: 20,
		DisplayPriceSnapshots: map[string][]currency.DisplayPriceSnapshot{
			"default_fee": {{Amount: 2.8, Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true}},
			"fee":         {{Amount: 99, Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "invalid_scope", Converted: true}},
		},
		Rules: []shippingRuleRequest{
			{
				Region:   "us",
				MinValue: 100,
				Fee:      15,
				DisplayPriceSnapshots: map[string][]currency.DisplayPriceSnapshot{
					"min_value": {{Amount: 14, Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true}},
					"fee":       {{Amount: 2.1, Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true}},
				},
			},
		},
	}

	template := req.toDomain()

	templateSnapshots := currency.ParseDisplayPriceSnapshotMap(template.DisplayPriceData, shippingdomain.ShippingTemplateDisplayPriceFields...)
	if len(templateSnapshots) != 1 || len(templateSnapshots["default_fee"]) != 1 {
		t.Fatalf("expected template default_fee display snapshot only, got %#v", templateSnapshots)
	}
	ruleSnapshots := currency.ParseDisplayPriceSnapshotMap(template.Rules[0].DisplayPriceData, shippingdomain.ShippingRuleDisplayPriceFields...)
	if len(ruleSnapshots) != 2 || len(ruleSnapshots["min_value"]) != 1 || len(ruleSnapshots["fee"]) != 1 {
		t.Fatalf("expected rule min_value and fee display snapshots, got %#v", ruleSnapshots)
	}
}

func TestShippingTemplateRequestDropsNonMoneyRuleThresholdSnapshots(t *testing.T) {
	req := shippingTemplateRequest{
		Name: "Weight shipping",
		Type: "weight",
		Rules: []shippingRuleRequest{
			{
				Region:   "us",
				MinValue: 1,
				Fee:      15,
				DisplayPriceSnapshots: map[string][]currency.DisplayPriceSnapshot{
					"min_value": {{Amount: 14, Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true}},
					"fee":       {{Amount: 2.1, Currency: "USD", QuoteCurrency: "USD", Rate: 0.14, Source: "direct_rate", Converted: true}},
				},
			},
		},
	}

	template := req.toDomain()
	ruleSnapshots := currency.ParseDisplayPriceSnapshotMap(template.Rules[0].DisplayPriceData, shippingdomain.ShippingRuleDisplayPriceFields...)
	if len(ruleSnapshots) != 1 || len(ruleSnapshots["fee"]) != 1 {
		t.Fatalf("expected only fee display snapshot for weight rule, got %#v", ruleSnapshots)
	}
}
