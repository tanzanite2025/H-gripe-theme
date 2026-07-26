package registration

import (
	"encoding/json"
	"strings"
	domainregistration "tanzanite/internal/domain/registration"
	"testing"
	"time"
)

func TestPublicRegistrationVerificationResponseOmitsSensitiveFields(t *testing.T) {
	payload, err := json.Marshal(publicRegistrationVerificationResponse(&domainregistration.ProductRegistration{
		ID:              1,
		UserID:          42,
		ProductID:       9,
		SerialNumber:    "TZ-123",
		PurchaseDate:    time.Now(),
		PurchaseProof:   "https://example.test/proof.jpg",
		Retailer:        "Hidden Retailer",
		Notes:           "private note",
		WarrantyPeriod:  24,
		WarrantyExpires: time.Now().AddDate(1, 0, 0),
		Status:          "active",
	}))
	if err != nil {
		t.Fatalf("marshal public registration response: %v", err)
	}

	body := string(payload)
	for _, forbidden := range []string{"user_id", "product_id", "purchase_proof", "retailer", "notes", "Hidden Retailer", "private note"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("public registration response leaked %q: %s", forbidden, body)
		}
	}
}
