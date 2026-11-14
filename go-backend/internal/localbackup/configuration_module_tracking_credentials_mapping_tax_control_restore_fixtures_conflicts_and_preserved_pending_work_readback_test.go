package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/taxrate"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

const configurationTrackingOriginalSyntheticAPIKey = "synthetic-original-tracking-api-key-must-not-be-task-data"
const configurationTrackingOriginalSyntheticWebhookSecret = "synthetic-original-webhook-secret-must-not-be-task-data"
const configurationTaxCollectionOriginalSyntheticError = "synthetic-original-tax-source-error-must-not-be-task-data"

func TestConfigurationTrackingProviderMappingAndTaxCollectionRecoveryPreservesOriginalCredentialsSwitchesAndDates(t *testing.T) {
	provider := `{"id":1,"provider_code":"17TRACK","provider_name":"Original provider","environment":"production","base_url":"https://synthetic.invalid/tracking","api_key":"original-api-key","webhook_secret":"original-webhook-secret","webhook_enabled":true,"auto_register":true,"polling_enabled":true,"polling_interval_minutes":37,"request_timeout_seconds":19,"sort_order":7,"enabled":true,"deleted_at":null}`
	mapping := `{"id":2,"provider_id":1,"scope":"carrier","carrier_id":3,"carrier_service_id":null,"provider_carrier_code":"ORIGINAL-CARRIER","enabled":true,"priority":17}`
	control := `{"id":1,"provider_code":"vatcomply","enabled":false,"refresh_interval_hours":731,"last_checked_at":"2025-01-02T03:04:05Z","last_successful_sync_at":"2025-01-01T03:04:05Z","last_sync_error":"original failure"}`
	for _, sample := range []struct {
		name, table, original, column, replacement string
		valid                                      bool
	}{
		{"original_tracking_provider", "shipping_tracking_providers", provider, "", "", true},
		{"original_mapping", "shipping_tracking_carrier_mappings", mapping, "", "", true},
		{"original_tax_control", "tax_rate_source_configs", control, "", "", true},
		{"original_zero_timeout_fallback", "shipping_tracking_providers", provider, "request_timeout_seconds", "0", true},
		{"original_disabled_future_provider", "shipping_tracking_providers", strings.Replace(provider, `"enabled":true`, `"enabled":false`, 1), "provider_code", `"FUTURE"`, true},
		{"missing_provider_name", "shipping_tracking_providers", provider, "provider_name", `" "`, false},
		{"invalid_environment", "shipping_tracking_providers", provider, "environment", `"test"`, false},
		{"negative_interval", "shipping_tracking_providers", provider, "polling_interval_minutes", "-1", false},
		{"fractional_timeout", "shipping_tracking_providers", provider, "request_timeout_seconds", "1.5", false},
		{"negative_sort", "shipping_tracking_providers", provider, "sort_order", "-1", false},
		{"missing_original_switch", "shipping_tracking_providers", provider, "auto_register", "null", false},
		{"unsupported_active_provider", "shipping_tracking_providers", provider, "provider_code", `"FUTURE"`, false},
		{"missing_original_api_key", "shipping_tracking_providers", provider, "api_key", `" "`, false},
		{"non_http_endpoint", "shipping_tracking_providers", provider, "base_url", `"file:///tmp/invalid"`, false},
		{"endpoint_credentials", "shipping_tracking_providers", provider, "base_url", `"https://name:secret@synthetic.invalid"`, false},
		{"missing_webhook_secret", "shipping_tracking_providers", provider, "webhook_secret", `""`, false},
		{"invalid_mapping_scope", "shipping_tracking_carrier_mappings", mapping, "scope", `"other"`, false},
		{"missing_mapping_parent", "shipping_tracking_carrier_mappings", mapping, "provider_id", "0", false},
		{"dual_mapping_targets", "shipping_tracking_carrier_mappings", mapping, "carrier_service_id", "4", false},
		{"missing_local_target", "shipping_tracking_carrier_mappings", mapping, "carrier_id", "null", false},
		{"negative_mapping_priority", "shipping_tracking_carrier_mappings", mapping, "priority", "-1", false},
		{"empty_provider_carrier_code", "shipping_tracking_carrier_mappings", mapping, "provider_carrier_code", `" "`, false},
		{"tax_control_wrong_identity", "tax_rate_source_configs", control, "id", "2", false},
		{"tax_control_wrong_provider", "tax_rate_source_configs", control, "provider_code", `"other"`, false},
		{"tax_control_interval_too_short", "tax_rate_source_configs", control, "refresh_interval_hours", "23", false},
		{"tax_control_interval_too_long", "tax_rate_source_configs", control, "refresh_interval_hours", "8761", false},
		{"tax_control_missing_enabled", "tax_rate_source_configs", control, "enabled", "null", false},
		{"tax_control_invalid_original_time", "tax_rate_source_configs", control, "last_checked_at", `"invalid"`, false},
	} {
		t.Run(sample.name, func(t *testing.T) {
			var fields map[string]json.RawMessage
			require.NoError(t, json.Unmarshal([]byte(sample.original), &fields))
			if sample.column != "" {
				fields[sample.column] = json.RawMessage(sample.replacement)
			}
			row, err := json.Marshal(fields)
			require.NoError(t, err)
			before := string(row)
			err = validateConfigurationModuleMissingTrackingProviderMappingAndTaxCollectionControlDomainRows(map[string][]json.RawMessage{sample.table: {row}})
			require.Equal(t, sample.valid, err == nil, "%v", err)
			require.Equal(t, before, string(row))
			if err != nil {
				require.NotContains(t, err.Error(), "original-api-key")
				require.NotContains(t, err.Error(), "original-webhook-secret")
				require.NotContains(t, err.Error(), "original failure")
			}
		})
	}
	for _, table := range strings.Fields("settings shipping_tracking_shipments tracking_events outbox_events orders social_oauth_connections") {
		require.False(t, configurationModuleHasMissingSiteContentWriteContract(selectedBusinessModuleCompleteSnapshot{Modules: []string{"configuration"}}, table))
	}
}

func configurationTrackingProviderMappingAndTaxCollectionControlRestoreIntegrationScenarios() []string {
	return []string{
		"tracking_partial_carrier_mapping_restore", "tracking_partial_service_mapping_restore", "tracking_partial_disabled_provider_and_mapping_restore", "tracking_partial_deleted_provider_restore", "collection_partial_enabled_tax_control_restore",
		"tracking_invalid_provider_name", "tracking_invalid_environment", "tracking_invalid_interval", "tracking_unsupported_active_provider", "tracking_missing_api_key", "tracking_invalid_endpoint", "tracking_missing_webhook_secret",
		"tracking_invalid_mapping_scope", "tracking_dual_mapping_targets", "tracking_negative_mapping_priority", "tracking_empty_provider_carrier_code", "tracking_partial_mapping_invalid_existing_provider", "tracking_partial_mapping_disabled_existing_service",
		"tracking_disabled_local_carrier", "tracking_deleted_provider", "tracking_provider_case_identity_conflict", "tracking_mapping_natural_identity_conflict", "tracking_changed_existing_credentials", "collection_changed_existing_tax_control", "legacy_v10_policy_requires_rehearsal",
	}
}

func createConfigurationTrackingProviderMappingTaxCollectionFixtureAndPreservedPendingWork(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `
INSERT INTO carriers(id,name,code,enabled) VALUES(900190,'Original tracking carrier','TRACKING-RESTORE-CARRIER',true),(900191,'Original disabled tracking carrier','TRACKING-RESTORE-DISABLED',false);
INSERT INTO shipping_carrier_services(id,carrier_id,service_code,service_name,currency,enabled) VALUES(900190,900190,'TRACKING-RESTORE-SERVICE','Original tracking service','USD',true);
INSERT INTO shipping_tracking_providers(id,provider_code,provider_name,environment,base_url,api_key,webhook_secret,webhook_enabled,auto_register,polling_enabled,polling_interval_minutes,request_timeout_seconds,sort_order,enabled,deleted_at,created_at,updated_at)
VALUES(900190,'17TRACK','Original tracking provider','production','https://synthetic.invalid/tracking','synthetic-original-tracking-api-key-must-not-be-task-data','synthetic-original-webhook-secret-must-not-be-task-data',true,true,true,37,19,7,true,null,'2025-01-02T03:04:05Z','2025-01-02T03:04:05Z'),
(900191,'FUTURE-TRACKING','Original disabled sandbox provider','sandbox','','','',false,false,false,0,0,8,false,null,'2025-01-02T03:04:05Z','2025-01-02T03:04:05Z'),
(900192,'MOCK','Original mock provider','production','','','',false,false,false,60,15,9,true,null,'2025-01-02T03:04:05Z','2025-01-02T03:04:05Z'),
(900193,'MOCK','Original soft deleted provider','sandbox','','','',false,false,false,60,15,10,false,'2025-06-01','2025-01-02T03:04:05Z','2025-01-02T03:04:05Z'),
(900197,'PRESERVED-PENDING-WORK','Preserved pending-work provider','sandbox','','','',false,false,false,60,15,20,false,null,'2025-01-02T03:04:05Z','2025-01-02T03:04:05Z');
INSERT INTO shipping_tracking_carrier_mappings(id,provider_id,scope,carrier_id,carrier_service_id,provider_carrier_code,enabled,priority,deleted_at)
VALUES(900190,900190,'carrier',900190,null,'ORIGINAL-CARRIER',true,17,null),
(900191,900190,'carrier_service',null,900190,'ORIGINAL-SERVICE',true,23,null),
(900192,900191,'carrier',900191,null,'ORIGINAL-DISABLED',false,29,null),
(900193,900190,'carrier',900190,null,'ORIGINAL-DELETED',false,31,'2025-06-01');
UPDATE tax_rate_source_configs SET enabled=false,refresh_interval_hours=731,last_checked_at='2025-01-02T03:04:05Z',last_successful_sync_at='2025-01-01T03:04:05Z',last_sync_error='synthetic-original-tax-source-error-must-not-be-task-data',created_at='2025-01-01T03:04:05Z',updated_at='2025-01-02T03:04:05Z' WHERE id=1;
INSERT INTO shipping_tracking_shipments(id,order_id,tracking_provider_id,tracking_number,provider_carrier_code,registration_status,sync_status,event_count,last_synced_at,next_sync_at,sync_lease_owner,sync_lease_generation,sync_lease_expires_at,last_error)
VALUES(900197,900001,900197,'PRESERVED-TRACKING-NUMBER','PRESERVED-CARRIER','pending','syncing',1,'2025-01-02T03:04:05Z','2099-01-01','original-polling-lease-owner',7,'2099-01-01','original tracking error');
INSERT INTO tracking_events(id,order_id,tracking_number,provider_carrier_code,status,description,event_time) VALUES(900197,900001,'PRESERVED-TRACKING-NUMBER','PRESERVED-CARRIER','in_transit','Original tracking fact','2025-01-02T03:04:05Z');
INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status,attempts,available_at)
VALUES(900197,'preserved-tracking-registration-command','tracking.shipment_registration_requested','order','900001','{"order_id":900001,"tracking_provider_id":900197,"tracking_number":"PRESERVED-TRACKING-NUMBER","provider_carrier_code":"PRESERVED-CARRIER","requested_at":"2025-01-02T03:04:05Z"}','pending',3,'2099-01-01');
`)
	require.NoError(t, err)
	mutation := map[string]string{
		"collection_partial_enabled_tax_control_restore":     `UPDATE tax_rate_source_configs SET enabled=true WHERE id=1`,
		"tracking_invalid_provider_name":                     `UPDATE shipping_tracking_providers SET provider_name=' ' WHERE id=900190`,
		"tracking_invalid_environment":                       `UPDATE shipping_tracking_providers SET environment='test' WHERE id=900190`,
		"tracking_invalid_interval":                          `UPDATE shipping_tracking_providers SET polling_interval_minutes=-1 WHERE id=900190`,
		"tracking_unsupported_active_provider":               `UPDATE shipping_tracking_providers SET provider_code='UNSUPPORTED' WHERE id=900190`,
		"tracking_missing_api_key":                           `UPDATE shipping_tracking_providers SET api_key='' WHERE id=900190`,
		"tracking_invalid_endpoint":                          `UPDATE shipping_tracking_providers SET base_url='file:///invalid' WHERE id=900190`,
		"tracking_missing_webhook_secret":                    `UPDATE shipping_tracking_providers SET webhook_secret='' WHERE id=900190`,
		"tracking_invalid_mapping_scope":                     `UPDATE shipping_tracking_carrier_mappings SET scope='other' WHERE id=900190`,
		"tracking_dual_mapping_targets":                      `UPDATE shipping_tracking_carrier_mappings SET carrier_service_id=900190 WHERE id=900190`,
		"tracking_negative_mapping_priority":                 `UPDATE shipping_tracking_carrier_mappings SET priority=-1 WHERE id=900190`,
		"tracking_empty_provider_carrier_code":               `UPDATE shipping_tracking_carrier_mappings SET provider_carrier_code=' ' WHERE id=900190`,
		"tracking_partial_mapping_invalid_existing_provider": `UPDATE shipping_tracking_providers SET provider_name=' ' WHERE id=900190`,
		"tracking_partial_mapping_disabled_existing_service": `UPDATE shipping_carrier_services SET enabled=false WHERE id=900190`,
		"tracking_disabled_local_carrier":                    `UPDATE carriers SET enabled=false WHERE id=900190`,
		"tracking_deleted_provider":                          `UPDATE shipping_tracking_providers SET deleted_at='2025-06-01' WHERE id=900190`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
}

func deleteConfigurationTrackingProviderMappingTaxControlFixtureRowsAndPreserveTargetOnlyConfiguration(t *testing.T, database *sql.DB, scenario string) int {
	t.Helper()
	statement, count := `DELETE FROM shipping_tracking_carrier_mappings WHERE id BETWEEN 900190 AND 900193;DELETE FROM shipping_tracking_providers WHERE id BETWEEN 900190 AND 900193;DELETE FROM tax_rate_source_configs WHERE id=1;`, 9
	switch scenario {
	case "tracking_partial_carrier_mapping_restore", "tracking_partial_mapping_invalid_existing_provider", "tracking_mapping_natural_identity_conflict", "tracking_changed_existing_credentials":
		statement, count = `DELETE FROM shipping_tracking_carrier_mappings WHERE id=900190;`, 1
	case "tracking_partial_service_mapping_restore", "tracking_partial_mapping_disabled_existing_service":
		statement, count = `DELETE FROM shipping_tracking_carrier_mappings WHERE id=900191;`, 1
	case "tracking_partial_disabled_provider_and_mapping_restore":
		statement, count = `DELETE FROM shipping_tracking_providers WHERE id=900191;`, 2
	case "tracking_partial_deleted_provider_restore":
		statement, count = `DELETE FROM shipping_tracking_providers WHERE id=900193;`, 1
	case "collection_partial_enabled_tax_control_restore":
		statement, count = `DELETE FROM tax_rate_source_configs WHERE id=1;`, 1
	case "collection_changed_existing_tax_control":
		statement, count = `DELETE FROM shipping_tracking_carrier_mappings WHERE id BETWEEN 900190 AND 900193;DELETE FROM shipping_tracking_providers WHERE id BETWEEN 900190 AND 900193;`, 8
	}
	_, err := database.ExecContext(context.Background(), statement+`
INSERT INTO shipping_tracking_providers(id,provider_code,provider_name,environment,enabled) VALUES(900199,'TARGET-ONLY-PROVIDER','Preserved current provider','sandbox',false);
`)
	require.NoError(t, err)
	mutation := map[string]string{
		"tracking_provider_case_identity_conflict":   `INSERT INTO shipping_tracking_providers(id,provider_code,provider_name,environment,enabled) VALUES(900198,'17track','Current code alias','sandbox',false)`,
		"tracking_mapping_natural_identity_conflict": `INSERT INTO shipping_tracking_carrier_mappings(id,provider_id,scope,carrier_id,provider_carrier_code) VALUES(900198,900190,'carrier',900190,'CURRENT-CARRIER')`,
		"tracking_changed_existing_credentials":      `UPDATE shipping_tracking_providers SET api_key='newer-current-key-must-not-be-overwritten' WHERE id=900190`,
		"collection_changed_existing_tax_control":    `UPDATE tax_rate_source_configs SET enabled=true,refresh_interval_hours=732 WHERE id=1`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
	return count
}

func configurationTrackingProviderMappingOrTaxCollectionIntegrationScenarioRequiresRejection(scenario string) bool {
	return (strings.HasPrefix(scenario, "tracking_") || strings.HasPrefix(scenario, "collection_")) && !strings.HasSuffix(scenario, "_restore")
}

func verifyRestoredTrackingCredentialsCarrierResolutionAndTaxControlThroughReadOnlyQueriesWithoutReplayingPendingWork(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	tx, err := database.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	require.NoError(t, err)
	defer tx.Rollback()
	gormDatabase, err := gorm.Open(postgres.New(postgres.Config{Conn: tx}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	shippingService := service.NewShippingService(repository.NewShippingRepository(gormDatabase))
	provider, err := shippingService.GetTrackingProviderConfig(900190)
	require.NoError(t, err)
	require.Equal(t, configurationTrackingOriginalSyntheticAPIKey, provider.APIKey)
	require.Equal(t, configurationTrackingOriginalSyntheticWebhookSecret, provider.WebhookSecret)
	require.True(t, provider.Enabled && provider.WebhookEnabled && provider.AutoRegister && provider.PollingEnabled)
	require.Equal(t, 37, provider.PollingIntervalMinutes)
	require.Equal(t, 19, provider.RequestTimeoutSeconds)
	require.Equal(t, "production", provider.Environment)
	require.Equal(t, "2025-01-02T03:04:05Z", provider.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"))
	carrierID, serviceID := uint(900190), uint(900190)
	for _, input := range []service.TrackingCarrierResolutionInput{{ProviderID: 900190, CarrierID: &carrierID}, {ProviderID: 900190, CarrierServiceID: &serviceID}} {
		resolution, err := shippingService.ResolveTrackingCarrier(input)
		require.NoError(t, err)
		expected := "ORIGINAL-CARRIER"
		if input.CarrierServiceID != nil {
			expected = "ORIGINAL-SERVICE"
		}
		require.Equal(t, expected, resolution.ProviderCarrierCode)
	}
	disabled, err := shippingService.GetTrackingProviderConfig(900191)
	require.NoError(t, err)
	require.False(t, disabled.Enabled)
	require.Equal(t, "sandbox", disabled.Environment)
	_, err = shippingService.GetTrackingProviderConfig(900193)
	require.Error(t, err)
	var control taxrate.TaxRateSourceConfig
	// The existing GetSourceConfiguration getter attempts INSERT ... DO NOTHING;
	// use an actual read-only model query instead of invoking its initialization.
	require.NoError(t, gormDatabase.First(&control, 1).Error)
	require.Equal(t, scenario == "collection_partial_enabled_tax_control_restore", control.Enabled)
	require.Equal(t, 731, control.RefreshIntervalHours)
	require.Equal(t, configurationTaxCollectionOriginalSyntheticError, control.LastSyncError)
	require.Equal(t, "2025-01-02T03:04:05Z", control.LastCheckedAt.UTC().Format("2006-01-02T15:04:05Z"))
	require.Equal(t, "2025-01-01T03:04:05Z", control.LastSuccessfulSyncAt.UTC().Format("2006-01-02T15:04:05Z"))
	var retainedShipment, retainedEvent, retainedCommand, targetOnly int
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT count(*) FROM shipping_tracking_shipments WHERE id=900197 AND registration_status='pending' AND sync_status='syncing' AND event_count=1 AND last_synced_at='2025-01-02T03:04:05Z' AND next_sync_at='2099-01-01' AND sync_lease_owner='original-polling-lease-owner' AND sync_lease_generation=7 AND sync_lease_expires_at='2099-01-01' AND last_error='original tracking error'`).Scan(&retainedShipment))
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT count(*) FROM tracking_events WHERE id=900197 AND status='in_transit' AND description='Original tracking fact'`).Scan(&retainedEvent))
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT count(*) FROM outbox_events WHERE id=900197 AND status='pending' AND attempts=3 AND processed_at IS NULL AND available_at='2099-01-01'`).Scan(&retainedCommand))
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT count(*) FROM shipping_tracking_providers WHERE id=900199 AND provider_code='TARGET-ONLY-PROVIDER' AND NOT enabled`).Scan(&targetOnly))
	require.Equal(t, 1, retainedShipment)
	require.Equal(t, 1, retainedEvent)
	require.Equal(t, 1, retainedCommand)
	require.Equal(t, 1, targetOnly)
}
