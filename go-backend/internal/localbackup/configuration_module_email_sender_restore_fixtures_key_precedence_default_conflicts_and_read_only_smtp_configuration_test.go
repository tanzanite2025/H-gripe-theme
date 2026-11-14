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

const configurationEmailSenderSyntheticOriginalMasterKey = "synthetic-original-email-sender-master-key-must-not-be-task-data"
const configurationEmailSenderSyntheticOriginalPassword = "synthetic-original-email-sender-password-must-not-be-task-data"
const configurationEmailSenderSyntheticOriginalTestError = "synthetic-original-email-sender-test-error-must-not-be-task-data"

func TestConfigurationEmailSenderRecoveryPreservesSMTPFieldsAndUsesOnlyOriginalBoundKeyPrecedence(t *testing.T) {
	t.Setenv("EMAIL_PROVIDER_MASTER_KEY", configurationEmailSenderSyntheticOriginalMasterKey)
	t.Setenv("SECURITY_MASTER_KEY", configurationEmailSenderSyntheticOriginalMasterKey)
	t.Setenv("PAYMENT_CONFIG_MASTER_KEY", configurationEmailSenderSyntheticOriginalMasterKey)
	encrypted, err := secretbox.EncryptString(configurationEmailSenderSyntheticOriginalPassword, configurationEmailSenderSyntheticOriginalMasterKey)
	require.NoError(t, err)
	original := map[string]any{"id": 1, "code": "ORIGINAL", "name": "Original SMTP", "driver": "smtp", "host": "smtp.synthetic.invalid", "port": 465, "username": "original-user", "password_encrypted": encrypted, "from_name": "Original store", "from_email": "store@synthetic.invalid", "reply_to": "support@synthetic.invalid", "encryption_type": "tls", "is_active": true, "is_default": true}
	for _, sample := range []struct {
		name, column string
		value        any
		keys         map[string]string
		valid        bool
	}{
		{"preferred_original_key", "", nil, map[string]string{"EMAIL_PROVIDER_MASTER_KEY": configurationEmailSenderSyntheticOriginalMasterKey, "SECURITY_MASTER_KEY": "different-security-key"}, true},
		{"security_fallback", "", nil, map[string]string{"SECURITY_MASTER_KEY": configurationEmailSenderSyntheticOriginalMasterKey, "PAYMENT_CONFIG_MASTER_KEY": "different-payment-key"}, true},
		{"payment_fallback", "", nil, map[string]string{"PAYMENT_CONFIG_MASTER_KEY": configurationEmailSenderSyntheticOriginalMasterKey}, true},
		{"no_process_environment_fallback", "", nil, nil, false},
		{"wrong_preferred_key_must_not_try_valid_lower_key", "", nil, map[string]string{"EMAIL_PROVIDER_MASTER_KEY": "wrong-preferred-key", "SECURITY_MASTER_KEY": configurationEmailSenderSyntheticOriginalMasterKey}, false},
		{"wrong_security_key_must_not_try_valid_payment_key", "", nil, map[string]string{"SECURITY_MASTER_KEY": "wrong-preferred-key", "PAYMENT_CONFIG_MASTER_KEY": configurationEmailSenderSyntheticOriginalMasterKey}, false},
		{"damaged_ciphertext", "password_encrypted", "v1:corrupted", map[string]string{"EMAIL_PROVIDER_MASTER_KEY": configurationEmailSenderSyntheticOriginalMasterKey}, false},
		{"plaintext_password_rejected", "password_encrypted", configurationEmailSenderSyntheticOriginalPassword, map[string]string{"EMAIL_PROVIDER_MASTER_KEY": configurationEmailSenderSyntheticOriginalMasterKey}, false},
		{"anonymous_smtp_without_key", "password_encrypted", "", nil, true},
		{"null_password_is_not_anonymous_smtp", "password_encrypted", nil, nil, false},
		{"inactive_original_default", "is_active", false, nil, false},
		{"missing_original_switch", "is_default", nil, nil, false},
		{"invalid_sender_address", "from_email", "invalid-address", nil, false},
		{"invalid_reply_address", "reply_to", "invalid-address", nil, false},
		{"empty_host", "host", " ", nil, false},
		{"unsupported_transport", "driver", "unsupported", nil, false},
		{"unsupported_encryption", "encryption_type", "unsupported", nil, false},
		{"invalid_port", "port", 65536, nil, false},
	} {
		t.Run(sample.name, func(t *testing.T) {
			fields := map[string]any{}
			for column, value := range original {
				fields[column] = value
			}
			if sample.column != "" {
				fields[sample.column] = sample.value
			}
			if sample.name == "anonymous_smtp_without_key" {
				fields["username"] = ""
			}
			row, err := json.Marshal(fields)
			require.NoError(t, err)
			before := string(row)
			cfg := Configuration{RecoverySecrets: sample.keys}
			err = cfg.validateConfigurationModuleMissingEmailSenderCredentialsWithOriginalBoundKeyPrecedence(map[string][]json.RawMessage{"email_provider_configs": {row}})
			require.Equal(t, sample.valid, err == nil, "%v", err)
			if err != nil {
				require.NotContains(t, err.Error(), encrypted)
				require.NotContains(t, err.Error(), configurationEmailSenderSyntheticOriginalPassword)
				require.NotContains(t, err.Error(), configurationEmailSenderSyntheticOriginalMasterKey)
			}
			require.Equal(t, before, string(row))
		})
	}
}

func configurationEmailSenderRestoreIntegrationScenarios() []string {
	return []string{"email_sender_partial_default_restore", "email_sender_partial_disabled_restore", "email_sender_security_key_fallback_restore", "email_sender_payment_key_fallback_restore", "email_sender_no_default_restore", "email_sender_corrupt_ciphertext", "email_sender_wrong_preferred_key", "email_sender_inactive_default", "email_sender_invalid_from_address", "email_sender_changed_existing_password", "email_sender_current_default_conflict", "email_sender_current_code_identity_conflict", "legacy_v12_policy_requires_rehearsal"}
}

func configureOriginalSyntheticEmailSenderBoundBusinessKeysForRecovery(t *testing.T, cfg *Configuration, scenario string) {
	t.Helper()
	cfg.RecoverySecrets["EMAIL_PROVIDER_MASTER_KEY"] = configurationEmailSenderSyntheticOriginalMasterKey
	cfg.RecoverySecrets["SECURITY_MASTER_KEY"] = "synthetic-original-email-security-fallback-key"
	if scenario == "email_sender_security_key_fallback_restore" || scenario == "email_sender_payment_key_fallback_restore" {
		delete(cfg.RecoverySecrets, "EMAIL_PROVIDER_MASTER_KEY")
	}
	if scenario == "email_sender_payment_key_fallback_restore" {
		delete(cfg.RecoverySecrets, "SECURITY_MASTER_KEY")
	}
}

func createConfigurationEmailSenderRestoreFixtureAndPreservedDeliveryHistory(t *testing.T, database *sql.DB, cfg Configuration, scenario string) {
	t.Helper()
	key := cfg.resolveConfigurationEmailSenderOriginalBoundBusinessKey()
	if scenario == "email_sender_wrong_preferred_key" {
		key = cfg.RecoverySecrets["SECURITY_MASTER_KEY"]
	}
	encrypted, err := secretbox.EncryptString(configurationEmailSenderSyntheticOriginalPassword, key)
	require.NoError(t, err)
	_, err = database.ExecContext(context.Background(), `
INSERT INTO email_provider_configs(id,code,name,driver,host,port,username,password_encrypted,from_name,from_email,reply_to,encryption_type,is_active,is_default,last_tested_at,last_test_status,last_test_error,created_at,updated_at)
VALUES(900240,'RESTORE-ORIGINAL-SMTP','Original default sender','smtp','smtp.synthetic.invalid',465,'original-smtp-user',$1,'Original store','store@synthetic.invalid','support@synthetic.invalid','tls',true,true,'2025-01-02T03:04:05Z','failed','synthetic-original-email-sender-test-error-must-not-be-task-data','2025-01-01','2025-01-02'),
(900241,'RESTORE-DISABLED-SMTP','Original disabled relay','smtp','relay.synthetic.invalid',25,'','','Original relay','relay@synthetic.invalid','','none',false,false,null,'','','2025-01-01','2025-01-02'),
(900242,'RESTORE-ANONYMOUS-SMTP','Original anonymous sender','smtp','relay.synthetic.invalid',587,'','','Original anonymous','anonymous@synthetic.invalid','','starttls',true,false,null,'','','2025-01-01','2025-01-02');`, encrypted)
	require.NoError(t, err)
	_, err = database.ExecContext(context.Background(), `
INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status,attempts,available_at,processed_at)
VALUES(900247,'preserved-email-delivery-pending','order.paid','order','900001','{"order_id":900001}','pending',3,'2099-01-01',null),
(900248,'preserved-email-delivery-sent','order.paid','order','900001','{"order_id":900001}','processed',1,'2025-01-01','2025-01-02T03:04:05Z');
INSERT INTO email_delivery_records(id,outbox_event_id,event_key,event_type,template_code,recipient_email,status,attempt_count,provider_code,last_error,first_attempt_at,last_attempt_at,sent_at)
VALUES(900247,900247,'preserved-email-delivery-pending','order.paid','order_confirmation','customer@synthetic.invalid','unknown',3,'RESTORE-ORIGINAL-SMTP','Original ambiguous delivery result','2025-01-01','2025-01-02',null),
(900248,900248,'preserved-email-delivery-sent','order.paid','order_confirmation','customer@synthetic.invalid','sent',1,'RESTORE-ORIGINAL-SMTP','','2025-01-01','2025-01-02','2025-01-02T03:04:05Z');`)
	require.NoError(t, err)
	mutation := map[string]string{
		"email_sender_no_default_restore":   `UPDATE email_provider_configs SET is_default=false WHERE id=900240`,
		"email_sender_corrupt_ciphertext":   `UPDATE email_provider_configs SET password_encrypted='v1:corrupted' WHERE id=900240`,
		"email_sender_inactive_default":     `UPDATE email_provider_configs SET is_active=false WHERE id=900240`,
		"email_sender_invalid_from_address": `UPDATE email_provider_configs SET from_email='invalid-address' WHERE id=900240`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
}

func deleteConfigurationEmailSenderFixtureRowsAndPreserveCurrentSenders(t *testing.T, database *sql.DB, scenario string) int {
	t.Helper()
	statement, count := `DELETE FROM email_provider_configs WHERE id BETWEEN 900240 AND 900242;`, 3
	switch scenario {
	case "email_sender_partial_default_restore":
		statement, count = `DELETE FROM email_provider_configs WHERE id=900240;`, 1
	case "email_sender_partial_disabled_restore", "email_sender_changed_existing_password":
		statement, count = `DELETE FROM email_provider_configs WHERE id=900241;`, 1
	}
	_, err := database.ExecContext(context.Background(), statement+`
INSERT INTO email_provider_configs(id,code,name,host,port,from_name,from_email,encryption_type,is_active,is_default) VALUES(900249,'TARGET-ONLY-SMTP','Preserved current relay','current.synthetic.invalid',25,'Current relay','current@synthetic.invalid','none',false,false);`)
	require.NoError(t, err)
	mutation := map[string]string{
		"email_sender_changed_existing_password":      `UPDATE email_provider_configs SET password_encrypted='newer-current-email-password-must-not-be-overwritten' WHERE id=900240`,
		"email_sender_current_default_conflict":       `UPDATE email_provider_configs SET is_active=true,is_default=true WHERE id=900249`,
		"email_sender_current_code_identity_conflict": `UPDATE email_provider_configs SET code='RESTORE-ORIGINAL-SMTP' WHERE id=900249`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
	return count
}

func configurationEmailSenderIntegrationScenarioRequiresRejection(scenario string) bool {
	return strings.HasPrefix(scenario, "email_sender_") && !strings.HasSuffix(scenario, "_restore")
}

func verifyRestoredEmailSenderThroughReadOnlyDefaultSMTPResolutionAndUnchangedDeliveryHistory(t *testing.T, database *sql.DB, cfg Configuration, scenario string) {
	t.Helper()
	for _, name := range []string{"EMAIL_PROVIDER_MASTER_KEY", "SECURITY_MASTER_KEY", "PAYMENT_CONFIG_MASTER_KEY"} {
		t.Setenv(name, cfg.RecoverySecrets[name])
	}
	tx, err := database.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	require.NoError(t, err)
	defer tx.Rollback()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: tx}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	providers := service.NewEmailProviderService(repository.NewEmailProviderRepository(db))
	views, err := providers.List()
	require.NoError(t, err)
	require.Len(t, views, 4)
	primary, err := providers.Get(900240)
	require.NoError(t, err)
	require.True(t, primary.IsActive && primary.HasPassword)
	require.Equal(t, scenario != "email_sender_no_default_restore", primary.IsDefault)
	require.Equal(t, "failed", primary.LastTestStatus)
	require.Equal(t, configurationEmailSenderSyntheticOriginalTestError, primary.LastTestError)
	require.Equal(t, "2025-01-02T03:04:05Z", primary.LastTestedAt.UTC().Format("2006-01-02T15:04:05Z"))
	smtpConfig, configured, err := providers.DefaultSMTPConfig()
	require.NoError(t, err)
	if scenario == "email_sender_no_default_restore" {
		require.False(t, configured)
		require.Nil(t, smtpConfig)
		require.Empty(t, providers.DefaultProviderCode())
	} else {
		require.True(t, configured)
		require.Equal(t, "RESTORE-ORIGINAL-SMTP", providers.DefaultProviderCode())
		require.Equal(t, "smtp.synthetic.invalid", smtpConfig.Host)
		require.Equal(t, 465, smtpConfig.Port)
		require.Equal(t, "original-smtp-user", smtpConfig.Username)
		require.True(t, smtpConfig.Password == configurationEmailSenderSyntheticOriginalPassword, "restored original password must remain usable")
		require.Equal(t, "tls", smtpConfig.EncryptionType)
		require.Equal(t, "store@synthetic.invalid", smtpConfig.From)
		require.Equal(t, "support@synthetic.invalid", smtpConfig.ReplyTo)
	}
	disabled, err := providers.Get(900241)
	require.NoError(t, err)
	require.False(t, disabled.IsActive || disabled.IsDefault || disabled.HasPassword)
	require.Equal(t, "none", disabled.EncryptionType)
	anonymous, err := providers.Get(900242)
	require.NoError(t, err)
	require.True(t, anonymous.IsActive)
	require.False(t, anonymous.IsDefault || anonymous.HasPassword)
	require.Equal(t, "starttls", anonymous.EncryptionType)
	var retained int
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT count(*) FROM outbox_events WHERE (id=900247 AND status='pending' AND attempts=3 AND available_at='2099-01-01' AND processed_at IS NULL) OR (id=900248 AND status='processed' AND attempts=1 AND processed_at='2025-01-02T03:04:05Z')`).Scan(&retained))
	require.Equal(t, 2, retained)
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT count(*) FROM email_delivery_records WHERE (id=900247 AND status='unknown' AND attempt_count=3 AND sent_at IS NULL AND last_error='Original ambiguous delivery result') OR (id=900248 AND status='sent' AND attempt_count=1 AND sent_at='2025-01-02T03:04:05Z')`).Scan(&retained))
	require.Equal(t, 2, retained)
	var auditJSON string
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT coalesce(jsonb_agg(to_jsonb(a)),'[]'::jsonb)::text FROM audit_logs a WHERE resource='configuration_module_site_content_restore'`).Scan(&auditJSON))
	for _, confidential := range []string{configurationEmailSenderSyntheticOriginalPassword, configurationEmailSenderSyntheticOriginalMasterKey, configurationEmailSenderSyntheticOriginalTestError, "v1:"} {
		require.NotContains(t, auditJSON, confidential)
	}
}
