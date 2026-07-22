package admin

import (
	"encoding/json"
	"strings"
	"testing"

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
