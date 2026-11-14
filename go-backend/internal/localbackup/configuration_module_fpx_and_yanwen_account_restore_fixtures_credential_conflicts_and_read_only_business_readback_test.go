package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"commerce-platform/internal/pkg/secretbox"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const configurationLogisticsAccountSyntheticMasterKey = "synthetic-original-logistics-master-key-must-not-be-task-data"
const configurationLogisticsAccountSyntheticCredential = "synthetic-original-logistics-credential-must-not-be-task-data"

func TestConfigurationFpxAndYanwenAccountsRequireOriginalBusinessKeysWithoutNormalizingCiphertextOrEnvironment(t *testing.T) {
	encrypted, err := secretbox.EncryptString(configurationLogisticsAccountSyntheticCredential, configurationLogisticsAccountSyntheticMasterKey)
	require.NoError(t, err)
	row, err := json.Marshal(map[string]any{"id": 1, "environment": "production", "endpoint": "https://open.4px.com/router/api/service", "enabled": true, "app_key_encrypted": encrypted, "app_secret_encrypted": encrypted, "access_token_encrypted": "", "last_sync_scanned": 13, "last_sync_added": 3, "last_sync_updated": 5, "last_sync_preserved_enabled": 7, "last_synced_at": "2025-01-02T03:04:05Z"})
	require.NoError(t, err)
	for _, sample := range []struct {
		name, column string
		value        any
		key          string
		valid        bool
	}{
		{"original_credentials", "", nil, configurationLogisticsAccountSyntheticMasterKey, true},
		{"missing_original_key", "", nil, "", false},
		{"wrong_original_key", "", nil, "different-synthetic-key", false},
		{"wrong_environment", "environment", "fat", configurationLogisticsAccountSyntheticMasterKey, false},
		{"wrong_environment_gateway", "endpoint", "https://open-test.4px.com/router/api/service", configurationLogisticsAccountSyntheticMasterKey, false},
		{"gateway_userinfo", "endpoint", "https://secret@open.4px.com/router/api/service", configurationLogisticsAccountSyntheticMasterKey, false},
		{"gateway_query", "endpoint", "https://open.4px.com/router/api/service?secret=hidden", configurationLogisticsAccountSyntheticMasterKey, false},
		{"gateway_fragment", "endpoint", "https://open.4px.com/router/api/service#fragment", configurationLogisticsAccountSyntheticMasterKey, false},
		{"ciphertext_corrupted", "app_secret_encrypted", "v1:corrupted", configurationLogisticsAccountSyntheticMasterKey, false},
		{"ciphertext_plaintext", "app_secret_encrypted", configurationLogisticsAccountSyntheticCredential, configurationLogisticsAccountSyntheticMasterKey, false},
		{"enabled_missing_secret", "app_secret_encrypted", "", configurationLogisticsAccountSyntheticMasterKey, false},
		{"null_credential", "app_key_encrypted", nil, configurationLogisticsAccountSyntheticMasterKey, false},
		{"negative_sync_count", "last_sync_scanned", -1, configurationLogisticsAccountSyntheticMasterKey, false},
		{"fractional_sync_count", "last_sync_scanned", 1.5, configurationLogisticsAccountSyntheticMasterKey, false},
	} {
		t.Run(sample.name, func(t *testing.T) {
			var fields map[string]any
			require.NoError(t, json.Unmarshal(row, &fields))
			if sample.column != "" {
				fields[sample.column] = sample.value
			}
			original, err := json.Marshal(fields)
			require.NoError(t, err)
			cfg := Configuration{RecoverySecrets: map[string]string{"FPX_API_MASTER_KEY": sample.key}}
			before := string(original)
			err = cfg.validateConfigurationModuleMissingFpxAndYanwenAccountCredentialsWithOriginalBusinessKeys(map[string][]json.RawMessage{"shipping_fpx_api_configs": {original}})
			require.Equal(t, sample.valid, err == nil, "%v", err)
			if err != nil {
				require.NotContains(t, err.Error(), configurationLogisticsAccountSyntheticCredential)
				require.NotContains(t, err.Error(), encrypted)
				require.NotContains(t, err.Error(), configurationLogisticsAccountSyntheticMasterKey)
			}
			require.Equal(t, before, string(original))
		})
	}
}

func configurationFpxAndYanwenAccountRestoreIntegrationScenarios() []string {
	return []string{"account_partial_fpx_production_restore", "account_partial_yanwen_production_restore", "account_partial_disabled_empty_restore", "account_fpx_wrong_environment", "account_yanwen_test_alias", "account_fpx_wrong_gateway", "account_yanwen_gateway_query", "account_fpx_missing_enabled_secret", "account_yanwen_missing_enabled_token", "account_corrupt_ciphertext", "account_wrong_encryption_key", "account_missing_original_business_key", "account_negative_sync_count", "account_changed_existing_credentials", "account_current_environment_identity_conflict", "legacy_v11_policy_requires_rehearsal"}
}

func configureOriginalSyntheticLogisticsAccountBusinessKeysForRecovery(t *testing.T, cfg *Configuration, scenario string) {
	t.Helper()
	cfg.RecoverySecrets["FPX_API_MASTER_KEY"] = configurationLogisticsAccountSyntheticMasterKey
	cfg.RecoverySecrets["YANWEN_API_MASTER_KEY"] = configurationLogisticsAccountSyntheticMasterKey
	if scenario == "account_missing_original_business_key" {
		delete(cfg.RecoverySecrets, "FPX_API_MASTER_KEY")
	}
}

func createConfigurationFpxAndYanwenAccountRestoreFixture(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	encrypted, err := secretbox.EncryptString(configurationLogisticsAccountSyntheticCredential, configurationLogisticsAccountSyntheticMasterKey)
	require.NoError(t, err)
	_, err = database.ExecContext(context.Background(), `
INSERT INTO shipping_fpx_api_configs(id,environment,endpoint,app_key_encrypted,app_secret_encrypted,access_token_encrypted,enabled,last_sync_status,last_synced_at,last_error,last_sync_scanned,last_sync_added,last_sync_updated,last_sync_preserved_enabled,created_at,updated_at)
VALUES(900220,'production','https://open.4px.com/router/api/service',$1,$1,$1,true,'error','2025-01-02T03:04:05Z','Original historical sync failure',13,3,5,7,'2025-01-01','2025-01-02'),
(900221,'test','https://open-test.4px.com/router/api/service','','','',false,'',null,'',0,0,0,0,'2025-01-01','2025-01-02');`, encrypted)
	require.NoError(t, err)
	_, err = database.ExecContext(context.Background(), `
INSERT INTO shipping_yanwen_api_configs(id,environment,endpoint,user_id_encrypted,api_token_encrypted,enabled,created_at,updated_at)
VALUES(900220,'production','https://open.yw56.com.cn/api/order',$1,$1,true,'2025-01-01','2025-01-02'),
(900221,'fat','https://open-fat.yw56.com.cn/api/order','','',false,'2025-01-01','2025-01-02');`, encrypted)
	require.NoError(t, err)
	mutation := map[string]string{
		"account_fpx_wrong_environment":        `UPDATE shipping_fpx_api_configs SET environment='fat' WHERE id=900220`,
		"account_yanwen_test_alias":            `UPDATE shipping_yanwen_api_configs SET environment='test' WHERE id=900220`,
		"account_fpx_wrong_gateway":            `UPDATE shipping_fpx_api_configs SET endpoint='https://open-test.4px.com/router/api/service' WHERE id=900220`,
		"account_yanwen_gateway_query":         `UPDATE shipping_yanwen_api_configs SET endpoint='https://open.yw56.com.cn/api/order?secret=hidden' WHERE id=900220`,
		"account_fpx_missing_enabled_secret":   `UPDATE shipping_fpx_api_configs SET app_secret_encrypted='' WHERE id=900220`,
		"account_yanwen_missing_enabled_token": `UPDATE shipping_yanwen_api_configs SET api_token_encrypted='' WHERE id=900220`,
		"account_corrupt_ciphertext":           `UPDATE shipping_fpx_api_configs SET access_token_encrypted='v1:corrupted' WHERE id=900220`,
		"account_negative_sync_count":          `UPDATE shipping_fpx_api_configs SET last_sync_scanned=-1 WHERE id=900220`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
	if scenario == "account_wrong_encryption_key" {
		wrongCiphertext, err := secretbox.EncryptString(configurationLogisticsAccountSyntheticCredential, "different-original-business-key")
		require.NoError(t, err)
		_, err = database.ExecContext(context.Background(), `UPDATE shipping_yanwen_api_configs SET api_token_encrypted=$1 WHERE id=900220`, wrongCiphertext)
		require.NoError(t, err)
	}
}

func deleteConfigurationFpxAndYanwenAccountFixtureRowsAndPreserveExistingCredentials(t *testing.T, database *sql.DB, scenario string) int {
	t.Helper()
	statement, count := `DELETE FROM shipping_fpx_api_configs WHERE id BETWEEN 900220 AND 900221;DELETE FROM shipping_yanwen_api_configs WHERE id BETWEEN 900220 AND 900221;`, 4
	switch scenario {
	case "account_partial_fpx_production_restore":
		statement, count = `DELETE FROM shipping_fpx_api_configs WHERE id=900220`, 1
	case "account_partial_yanwen_production_restore", "account_changed_existing_credentials":
		statement, count = `DELETE FROM shipping_yanwen_api_configs WHERE id=900220`, 1
	case "account_partial_disabled_empty_restore":
		statement, count = `DELETE FROM shipping_fpx_api_configs WHERE id=900221;DELETE FROM shipping_yanwen_api_configs WHERE id=900221`, 2
	}
	_, err := database.ExecContext(context.Background(), statement)
	require.NoError(t, err)
	if scenario == "account_changed_existing_credentials" {
		_, err = database.ExecContext(context.Background(), `UPDATE shipping_fpx_api_configs SET app_secret_encrypted='newer-current-credentials' WHERE id=900220`)
		require.NoError(t, err)
	}
	if scenario == "account_current_environment_identity_conflict" {
		_, err = database.ExecContext(context.Background(), `INSERT INTO shipping_fpx_api_configs(id,environment,endpoint) VALUES(900229,'production','https://open.4px.com/router/api/service')`)
		require.NoError(t, err)
	}
	return count
}

func configurationFpxAndYanwenAccountIntegrationScenarioRequiresRejection(scenario string) bool {
	return strings.HasPrefix(scenario, "account_") && !strings.HasSuffix(scenario, "_restore")
}

func verifyRestoredFpxAndYanwenAccountsThroughReadOnlyBusinessViewsAndLocalDecryptionWithoutGatewayCalls(t *testing.T, database *sql.DB) {
	t.Helper()
	tx, err := database.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	require.NoError(t, err)
	defer tx.Rollback()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: tx}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	fpxRepository := repository.NewFpxAPIConfigRepository(db)
	yanwenRepository := repository.NewYanwenAPIConfigRepository(db)
	fpxService := service.NewFpxAPIService(fpxRepository, nil)
	yanwenService := service.NewYanwenGatewayConfigurationService(yanwenRepository, nil)
	fpx, err := fpxService.View("production")
	require.NoError(t, err)
	require.True(t, fpx.Enabled && fpx.AppKeyConfigured && fpx.AppSecretConfigured && fpx.AccessTokenConfigured)
	require.Equal(t, "error", fpx.LastSyncStatus)
	require.Equal(t, "Original historical sync failure", fpx.LastError)
	require.Equal(t, "2025-01-02T03:04:05Z", fpx.LastSyncedAt.UTC().Format("2006-01-02T15:04:05Z"))
	require.Equal(t, []int{13, 3, 5, 7}, []int{fpx.LastSyncScanned, fpx.LastSyncAdded, fpx.LastSyncUpdated, fpx.LastSyncPreservedEnabled})
	yanwen, err := yanwenService.GetYanwenAPIConfigurationView("production")
	require.NoError(t, err)
	require.True(t, yanwen.Enabled && yanwen.UserIDConfigured && yanwen.APITokenConfigured)
	fpxTest, err := fpxService.View("test")
	require.NoError(t, err)
	require.False(t, fpxTest.Enabled || fpxTest.AppKeyConfigured || fpxTest.AppSecretConfigured || fpxTest.AccessTokenConfigured)
	yanwenFAT, err := yanwenService.GetYanwenAPIConfigurationView("fat")
	require.NoError(t, err)
	require.False(t, yanwenFAT.Enabled || yanwenFAT.UserIDConfigured || yanwenFAT.APITokenConfigured)
	for _, table := range configurationModuleMissingFpxAndYanwenAccountWritableTableNames() {
		rows, err := readProductCatalogRows(context.Background(), tx, table, "true")
		require.NoError(t, err)
		for _, row := range rows {
			for _, column := range configurationModuleFpxOrYanwenAccountEncryptedCredentialColumns(table) {
				encrypted, _ := readBusinessModuleSnapshotRowColumnText(row, column)
				if encrypted == "" {
					continue
				}
				plaintext, err := secretbox.DecryptString(encrypted, configurationLogisticsAccountSyntheticMasterKey)
				require.NoError(t, err)
				require.True(t, plaintext == configurationLogisticsAccountSyntheticCredential, "original credentials must remain usable")
			}
		}
	}
	var auditJSON string
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT coalesce(jsonb_agg(to_jsonb(a)),'[]'::jsonb)::text FROM audit_logs a WHERE resource='configuration_module_site_content_restore'`).Scan(&auditJSON))
	require.NotContains(t, auditJSON, configurationLogisticsAccountSyntheticMasterKey)
	require.NotContains(t, auditJSON, configurationLogisticsAccountSyntheticCredential)
	require.NotContains(t, auditJSON, "v1:")
}
