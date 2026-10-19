package orderevidence

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	"github.com/stretchr/testify/require"
)

func TestImmutableSnapshotLegacyJSONBSerializationRestoresOnlyAuthenticatedCompleteContent(t *testing.T) {
	initial := OrderEvidenceSnapshotPayload{SchemaVersion: 1, ConfirmedAt: time.Now().UTC(), Currency: "USD", Items: []OrderEvidenceSnapshotItem{{OrderItemID: 9007199254740993, SelectedSpecsJSON: json.RawMessage(`{"long_key":1,"z":2}`)}}}
	stripe := StripeDisputeEvidenceSnapshotPayload{DisputeEvidenceSnapshotEnvelope: DisputeEvidenceSnapshotEnvelope{SchemaVersion: 1, Provider: "stripe", OrderID: 9007199254740993}, Request: StripeDisputeEvidenceSnapshotRequest{AdditionalStatement: "original <proof> & statement"}}
	paypal := PayPalDisputeEvidenceSnapshotPayload{DisputeEvidenceSnapshotEnvelope: DisputeEvidenceSnapshotEnvelope{SchemaVersion: 1, Provider: "paypal", OrderID: 9007199254740993}, Request: PayPalDisputeEvidenceSnapshotRequest{Notes: "original <proof> & statement", Documents: []PayPalDisputeEvidenceDocument{{URL: "https://cdn.example.test/invoice.pdf", Name: "invoice"}}}}
	export := OrderEvidenceExportSnapshotPayload{SchemaVersion: 1, ExportType: "order_evidence_manifest", Order: &order.Order{ID: 9007199254740993}, Package: &OrderEvidenceExportPackage{Items: []OrderEvidenceExportItem{{DataJSON: json.RawMessage(`{"z":2,"long_key":1}`)}}}}
	for _, scenario := range []struct {
		name    string
		payload any
		fresh   func() any
	}{
		{"initial", initial, func() any { return &OrderEvidenceSnapshotPayload{} }},
		{"stripe", stripe, func() any { return &StripeDisputeEvidenceSnapshotPayload{} }},
		{"paypal", paypal, func() any { return &PayPalDisputeEvidenceSnapshotPayload{} }},
		{"export", export, func() any { return &OrderEvidenceExportSnapshotPayload{} }},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			original, err := json.Marshal(scenario.payload)
			require.NoError(t, err)
			digest := fmt.Sprintf("%x", sha256.Sum256(original))
			jsonb, err := serializeLegacyImmutableSnapshotRawJSONWithKnownKeyOrder(original, true)
			require.NoError(t, err)
			require.NotEqual(t, string(original), string(jsonb))
			require.True(t, immutableSnapshotDigestMatchesDocumentOrVerifiedLegacySerialization(jsonb, digest, scenario.fresh()))
			for _, mutation := range []string{
				strings.Replace(string(jsonb), "9007199254740993", "9007199254740992", 1),
				`{"unrecognized_field":"https://cdn.example.test/unowned.png",` + string(jsonb)[1:],
				`{"SCHEMA_VERSION":999,` + string(jsonb)[1:],
			} {
				require.False(t, immutableSnapshotDigestMatchesDocumentOrVerifiedLegacySerialization([]byte(mutation), digest, scenario.fresh()))
			}
		})
	}
}

func TestImmutableSnapshotCanonicalDigestPreservesExactDecimalAndLargeIntegerAcrossJSONBNotation(t *testing.T) {
	first, err := CanonicalImmutableSnapshotJSONForStorageIndependentIntegrity([]byte(`{"id":9007199254740993,"raw":{"fraction":1.2300e-4,"zero":-0.0,"large":1e21}}`))
	require.NoError(t, err)
	second, err := CanonicalImmutableSnapshotJSONForStorageIndependentIntegrity([]byte(`{"raw":{"large":1000000000000000000000,"zero":0,"fraction":0.000123},"id":9007199254740993}`))
	require.NoError(t, err)
	require.Equal(t, string(first), string(second))
	require.Contains(t, string(first), "9007199254740993")
	digest := fmt.Sprintf("%x", sha256.Sum256(first))
	require.True(t, immutableSnapshotDigestMatchesDocumentOrVerifiedLegacySerialization(second, digest, nil))
	require.False(t, immutableSnapshotDigestMatchesDocumentOrVerifiedLegacySerialization([]byte(strings.Replace(string(second), "9007199254740993", "9007199254740992", 1)), digest, nil))
}
