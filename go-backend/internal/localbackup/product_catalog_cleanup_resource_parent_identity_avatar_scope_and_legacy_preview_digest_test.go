package localbackup

import (
	"encoding/json"
	"testing"

	"commerce-platform/internal/domain/outbox"
	"github.com/stretchr/testify/require"
)

func TestProductCatalogCleanupReceiptsDistinguishDerivativeResourceFromRequiredAssetWithSameNumericID(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	parents := map[string]productCatalogMediaObjectRegistration{"88": {ID: "88"}}
	derivatives := map[string]productCatalogMediaObjectRegistration{"7": {ID: "7", ParentID: "88", Key: "media-derivatives/88/retired.png", URL: "/uploads/media-derivatives/88/retired.png", DeletedAt: "retired"}}
	row := json.RawMessage(`{"event_type":"storage.object_cleanup_requested","event_key":"synthetic-exact-resource","aggregate_type":"media_asset","aggregate_id":"88","status":"processed","payload":{"resource_type":"retired_media_derivative","resource_id":"7","object_keys":["media-derivatives/88/retired.png"]}}`)
	require.NoError(t, cfg.validateProductCatalogRetainedCleanupReceipt(row, map[string]bool{"products/original.png": true}, map[string]bool{"7": true}, parents, derivatives))
	derivatives["7"] = productCatalogMediaObjectRegistration{ID: "7", ParentID: "99", Key: "media-derivatives/88/retired.png", DeletedAt: "retired"}
	require.ErrorContains(t, cfg.validateProductCatalogRetainedCleanupReceipt(row, nil, nil, parents, derivatives), "registration or parent")
}

func TestProductCatalogCleanupChecksAvatarObjectsAndWholeAssetScopeBeyondReceiptKeyList(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	for _, scenario := range []struct {
		name, eventType, aggregateType, payload string
		requiredKeys, requiredIDs               map[string]bool
	}{
		{"avatar", outbox.EventTypeCustomerServiceAvatarCleanup, outbox.AggregateTypeCustomerServiceAgentProfile, `{"url":"/uploads/customer-service/avatars/agent.png"}`, map[string]bool{"customer-service/avatars/agent.png": true}, nil},
		{"whole_asset_private_derivative_namespace", outbox.EventTypeObjectStorageCleanup, outbox.AggregateTypeMediaAsset, `{"resource_type":"media_asset","resource_id":"7","object_keys":["other/old.png"]}`, nil, map[string]bool{"7": true}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			encoded, err := json.Marshal(map[string]any{"event_type": scenario.eventType, "event_key": "synthetic-avatar-or-namespace", "aggregate_type": scenario.aggregateType, "aggregate_id": "7", "status": "processed", "payload": json.RawMessage(scenario.payload)})
			require.NoError(t, err)
			require.ErrorContains(t, cfg.validateProductCatalogRetainedCleanupReceipt(encoded, scenario.requiredKeys, scenario.requiredIDs, nil, nil), "can delete a required product image")
		})
	}
}

func TestProductCatalogOldPreviewJSONWithoutCleanupProtectionRetainsItsOriginalRollbackDigest(t *testing.T) {
	// This is the field inventory serialized before cleanup protection existed.
	legacy := struct {
		Version             int                                   `json:"version"`
		ConfigurationSHA256 string                                `json:"configuration_sha256"`
		Target              productCatalogRecoveryDatabaseBinding `json:"target"`
		Candidate           productCatalogRecoveryDatabaseBinding `json:"candidate"`
		TargetMedia         *moduleRecoveryCompleteMediaInventory `json:"target_media"`
		CandidateMedia      *moduleRecoveryCompleteMediaInventory `json:"candidate_media"`
	}{Version: 1, ConfigurationSHA256: "original-configuration", Target: productCatalogRecoveryDatabaseBinding{OID: 17}, Candidate: productCatalogRecoveryDatabaseBinding{OID: 18}}
	encoded, err := json.Marshal(legacy)
	require.NoError(t, err)
	var decoded productCatalogRecoveryPreviewBinding
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	expected, err := customerServiceRecoveryPlanCanonicalJSONSHA256(legacy)
	require.NoError(t, err)
	actual, err := customerServiceRecoveryPlanCanonicalJSONSHA256(decoded)
	require.NoError(t, err)
	require.Equal(t, expected, actual, "an interrupted older journal must remain eligible for rollback")
}

func TestProductCatalogPreviewWithOriginalCleanupProtectionPreservesLegacyRollbackDigestWithoutStandardTriggerGuardVersion(t *testing.T) {
	// Match the persisted field order from before standard trigger validation.
	type originalCleanupProtection struct {
		Version                 int    `json:"version"`
		RequiredImages          int    `json:"required_images"`
		InspectedRegistrations  int    `json:"inspected_registrations"`
		RetainedCleanupReceipts int    `json:"retained_cleanup_receipts"`
		RequiredPathsSHA256     string `json:"required_paths_sha256"`
		RegistrationRowsSHA256  string `json:"registration_rows_sha256"`
		CleanupRowsSHA256       string `json:"cleanup_rows_sha256"`
	}
	legacy := struct {
		Version                        int                                   `json:"version"`
		ConfigurationSHA256            string                                `json:"configuration_sha256"`
		Target                         productCatalogRecoveryDatabaseBinding `json:"target"`
		Candidate                      productCatalogRecoveryDatabaseBinding `json:"candidate"`
		TargetMedia                    *moduleRecoveryCompleteMediaInventory `json:"target_media"`
		CandidateMedia                 *moduleRecoveryCompleteMediaInventory `json:"candidate_media"`
		RequiredImageCleanupProtection *originalCleanupProtection            `json:"required_image_cleanup_protection,omitempty"`
	}{Version: 1, ConfigurationSHA256: "original-configuration", RequiredImageCleanupProtection: &originalCleanupProtection{Version: 1, RequiredImages: 3, RequiredPathsSHA256: "original-paths", RegistrationRowsSHA256: "original-registrations", CleanupRowsSHA256: "original-receipts"}}
	encoded, err := json.Marshal(legacy)
	require.NoError(t, err)
	var decoded productCatalogRecoveryPreviewBinding
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	require.Zero(t, decoded.RequiredImageCleanupProtection.StandardWriteGuardVersion)
	expected, err := customerServiceRecoveryPlanCanonicalJSONSHA256(legacy)
	require.NoError(t, err)
	actual, err := customerServiceRecoveryPlanCanonicalJSONSHA256(decoded)
	require.NoError(t, err)
	require.Equal(t, expected, actual, "new validation must not strand an interrupted journal from the preceding version")
}
