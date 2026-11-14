package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"

	domainmoney "commerce-platform/internal/domain/money"
	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const configurationPaymentMethodOriginalSyntheticSettings = "synthetic-original-payment-method-settings-must-not-be-task-data"
const configurationPaymentProtectionOriginalSyntheticReason = "synthetic-original-payment-protection-reason-must-not-be-task-data"

func TestConfigurationPaymentMethodAndProtectionRecoveryPreservesExactFeesAndOriginalExpiredOrRevokedControls(t *testing.T) {
	method := json.RawMessage("{\"id\":1,\"name\":\"Original method\",\"code\":\"original\",\"fee_type\":\"percentage\",\"fee_rate_decimal\":2.123456789012345,\"fee_value_minor\":901,\"min_amount_minor\":9007199254740993,\"max_amount_minor\":0,\"enabled\":false}")
	control := json.RawMessage("{\"id\":2,\"action\":\"pause_payment\",\"scope_type\":\"global\",\"scope_value\":\"\",\"reason\":\"synthetic-original-payment-protection-reason-must-not-be-task-data\",\"created_by\":3,\"updated_by\":4,\"enabled\":true,\"created_at\":\"2025-01-01T00:00:00Z\",\"updated_at\":\"2025-01-01T00:01:00Z\",\"expires_at\":\"2025-01-01T01:00:00Z\"}")
	for _, sample := range []struct {
		name, table, column, replacement string
		valid                            bool
	}{
		{"original_exact_percentage_and_disabled", "payment_methods", "", "", true},
		{"original_fixed_fee", "payment_methods", "fee_type", "\"fixed\"", true},
		{"valid_unicode_name", "payment_methods", "name", "\"" + strings.Repeat("支付", 100) + "\"", true},
		{"unsupported_fee_type", "payment_methods", "fee_type", "\"dynamic\"", false},
		{"missing_rate", "payment_methods", "fee_rate_decimal", "null", false},
		{"excess_decimal_precision", "payment_methods", "fee_rate_decimal", "2.1234567890123456", false},
		{"rate_over_100", "payment_methods", "fee_rate_decimal", "100.000000000000001", false},
		{"negative_fee", "payment_methods", "fee_value_minor", "-1", false},
		{"fractional_minor_units", "payment_methods", "fee_value_minor", "1.5", false},
		{"minor_units_overflow", "payment_methods", "min_amount_minor", "9223372036854775808", false},
		{"inverted_amount_range", "payment_methods", "max_amount_minor", "10", false},
		{"missing_original_method_switch", "payment_methods", "enabled", "null", false},
		{"expired_original_control_is_not_renewed", "payment_protection_controls", "", "", true},
		{"revoked_control_stays_revoked", "payment_protection_controls", "enabled", "false", true},
		{"system_actor_zero_is_preserved", "payment_protection_controls", "created_by", "0", true},
		{"provider_scope_requires_value", "payment_protection_controls", "scope_type", "\"provider\"", false},
		{"unsupported_action", "payment_protection_controls", "action", "\"allow_all\"", false},
		{"unsupported_scope", "payment_protection_controls", "scope_type", "\"user\"", false},
		{"global_scope_with_value", "payment_protection_controls", "scope_value", "\"stripe\"", false},
		{"missing_original_control_switch", "payment_protection_controls", "enabled", "null", false},
		{"missing_original_actor", "payment_protection_controls", "created_by", "null", false},
		{"negative_original_actor", "payment_protection_controls", "updated_by", "-1", false},
		{"invalid_expiry", "payment_protection_controls", "expires_at", "\"invalid\"", false},
		{"expiry_before_creation", "payment_protection_controls", "expires_at", "\"2024-12-31T00:00:00Z\"", false},
		{"missing_creation_time", "payment_protection_controls", "created_at", "null", false},
		{"empty_reason", "payment_protection_controls", "reason", "\" \"", false},
	} {
		t.Run(sample.name, func(t *testing.T) {
			original := method
			if sample.table == "payment_protection_controls" {
				original = control
			}
			var fields map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(original, &fields))
			if sample.column != "" {
				fields[sample.column] = json.RawMessage(sample.replacement)
			}
			row, err := json.Marshal(fields)
			require.NoError(t, err)
			before := string(row)
			err = validateConfigurationModuleMissingPaymentMethodAndTimeBoundedProtectionDomainRows(map[string][]json.RawMessage{sample.table: {row}})
			require.Equal(t, sample.valid, err == nil, "%v", err)
			require.Equal(t, before, string(row))
			if err != nil {
				require.NotContains(t, err.Error(), configurationPaymentProtectionOriginalSyntheticReason)
			}
		})
	}
	for _, sample := range []struct {
		scope, value string
		valid        bool
	}{
		{"provider", "stripe", true}, {"provider", "paypal", true}, {"provider", "other", false}, {"provider", "Stripe", false},
		{"country", "NZ", true}, {"country", "nz", false}, {"country", "N", false},
		{"payment_method", "card", true}, {"payment_method", "", false}, {"payment_method", "Card", false},
	} {
		t.Run("scope_"+sample.scope+"_"+sample.value, func(t *testing.T) {
			var fields map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(control, &fields))
			fields["scope_type"], _ = json.Marshal(sample.scope)
			fields["scope_value"], _ = json.Marshal(sample.value)
			row, err := json.Marshal(fields)
			require.NoError(t, err)
			err = validateConfigurationModuleMissingPaymentMethodAndTimeBoundedProtectionDomainRows(map[string][]json.RawMessage{"payment_protection_controls": {row}})
			require.Equal(t, sample.valid, err == nil, "%v", err)
		})
	}
	for _, table := range strings.Fields("settings social_oauth_connections transactions refunds payment_refund_executions payment_operation_idempotencies outbox_events orders") {
		require.False(t, configurationModuleHasMissingSiteContentWriteContract(selectedBusinessModuleCompleteSnapshot{Modules: []string{"configuration"}}, table), table)
	}
}

func configurationPaymentMethodAndTimeBoundedProtectionRestoreIntegrationScenarios() []string {
	return []string{
		"payment_partial_methods_restore", "payment_partial_expired_control_restore", "payment_partial_disabled_method_restore", "payment_partial_revoked_control_restore", "payment_system_actor_restore",
		"payment_invalid_method_fee_type", "payment_invalid_control_action", "payment_changed_original_actor", "payment_missing_original_actor",
		"payment_changed_existing_method", "payment_current_method_code_conflict", "payment_current_control_revocation_conflict", "legacy_v13_policy_requires_rehearsal",
	}
}

func createConfigurationPaymentMethodAndProtectionFixtureWithPreservedTransactions(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `
INSERT INTO users(id,email,username,password,role) VALUES(900260,'protection-creator@synthetic.invalid','protection-creator','synthetic-original-password','admin'),(900261,'protection-updater@synthetic.invalid','protection-updater','synthetic-original-password','admin');
INSERT INTO payment_methods(id,name,code,fee_type,fee_rate_decimal,fee_value_minor,min_amount_minor,max_amount_minor,enabled,sort_order,settings,created_at,updated_at,deleted_at)
VALUES(900260,'Original fixed method','restore-fixed','fixed',0,901,9007199254740993,0,true,17,'{"original_note":"synthetic-original-payment-method-settings-must-not-be-task-data"}','2025-01-01','2025-01-02',null),
(900261,'Original percentage method','restore-percent','percentage',2.123456789012345,0,0,0,true,19,'{}','2025-01-01','2025-01-02',null),
(900262,'Original disabled method','restore-disabled','fixed',0,23,0,0,false,21,null,'2025-01-01','2025-01-02',null),
(900263,'Original deleted method','restore-deleted','fixed',0,0,0,0,false,-1,'{}','2025-01-01','2025-01-02','2025-01-03');
INSERT INTO payment_protection_controls(id,action,scope_type,scope_value,reason,expires_at,enabled,created_by,updated_by,created_at,updated_at)
VALUES(900260,'force_3ds','provider','stripe','synthetic-original-payment-protection-reason-must-not-be-task-data',now()+interval '2 hours',true,900260,900261,now()-interval '10 minutes',now()-interval '5 minutes'),
(900261,'pause_payment','global','','Original expired global pause','2025-01-01 01:00:00Z',true,900260,900261,'2025-01-01','2025-01-01 00:05:00Z'),
(900262,'pause_payment','global','','Original revoked global pause',now()+interval '2 hours',false,900260,900261,now()-interval '10 minutes',now()-interval '5 minutes'),
(900263,'pause_payment','country','GB','Original country pause',now()+interval '2 hours',true,900260,900261,now()-interval '10 minutes',now()-interval '5 minutes');
INSERT INTO transactions(id,order_id,transaction_id,payment_method,amount_minor,currency,status,attempt_key,provider_request_key,completed_at)
VALUES(900268,900001,'preserved-original-completed-payment','card',123,'USD','completed','original-completed-attempt','original-completed-request','2025-01-02 03:04:05Z'),
(900269,900001,'preserved-original-pending-payment','card',123,'USD','requires_action','original-pending-attempt','original-pending-request',null);
`)
	require.NoError(t, err)
	mutation := map[string]string{
		"payment_system_actor_restore":    `UPDATE payment_protection_controls SET created_by=0,updated_by=0 WHERE id BETWEEN 900260 AND 900263`,
		"payment_invalid_method_fee_type": `UPDATE payment_methods SET fee_type='dynamic' WHERE id=900260`,
		"payment_invalid_control_action":  `UPDATE payment_protection_controls SET action='allow_all' WHERE id=900260`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
}

func deleteConfigurationPaymentMethodAndProtectionFixtureRowsWithoutChangingCurrentPaymentFacts(t *testing.T, database *sql.DB, scenario string) int {
	t.Helper()
	statement, count := `DELETE FROM payment_methods WHERE id BETWEEN 900260 AND 900263; DELETE FROM payment_protection_controls WHERE id BETWEEN 900260 AND 900263;`, 8
	switch scenario {
	case "payment_partial_methods_restore":
		statement, count = `DELETE FROM payment_methods WHERE id BETWEEN 900260 AND 900263;`, 4
	case "payment_partial_expired_control_restore", "payment_current_control_revocation_conflict":
		statement, count = `DELETE FROM payment_protection_controls WHERE id=900261;`, 1
	case "payment_partial_disabled_method_restore", "payment_changed_existing_method":
		statement, count = `DELETE FROM payment_methods WHERE id=900262;`, 1
	case "payment_partial_revoked_control_restore":
		statement, count = `DELETE FROM payment_protection_controls WHERE id=900262;`, 1
	}
	_, err := database.ExecContext(context.Background(), statement+`
INSERT INTO payment_methods(id,name,code,fee_type,fee_rate_decimal,fee_value_minor,min_amount_minor,max_amount_minor,enabled,settings) VALUES(900269,'Preserved target method','target-only-method','fixed',0,73,0,0,false,'{"preserved":true}');
INSERT INTO payment_protection_controls(id,action,scope_type,scope_value,reason,expires_at,enabled,created_by,updated_by,created_at,updated_at) VALUES(900269,'pause_payment','country','US','Preserved target expired pause','2025-01-01 02:00:00Z',true,0,0,'2025-01-01','2025-01-01 00:05:00Z');
`)
	require.NoError(t, err)
	mutation := map[string]string{
		"payment_changed_original_actor":              `UPDATE users SET username='changed-original-protection-creator' WHERE id=900260`,
		"payment_missing_original_actor":              `DELETE FROM users WHERE id=900260`,
		"payment_changed_existing_method":             `UPDATE payment_methods SET fee_value_minor=999 WHERE id=900260`,
		"payment_current_method_code_conflict":        `UPDATE payment_methods SET code='restore-fixed' WHERE id=900269`,
		"payment_current_control_revocation_conflict": `UPDATE payment_protection_controls SET enabled=false WHERE id=900260`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
	return count
}

func configurationPaymentMethodOrProtectionIntegrationScenarioRequiresRejection(scenario string) bool {
	return strings.HasPrefix(scenario, "payment_") && !strings.HasSuffix(scenario, "_restore")
}

func verifyConfigurationPaymentProtectionPackageWithoutOriginalActorFingerprintsIsRejected(t *testing.T, database *sql.DB, cfg Configuration) {
	t.Helper()
	tx, err := database.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	require.NoError(t, err)
	defer tx.Rollback()
	// Match the real package/preview transaction before collecting hashes. The
	// fixture database deliberately defaults to Asia/Shanghai; an observation
	// otherwise switches to UTC after capture and changes timestamp JSON text.
	_, err = tx.ExecContext(context.Background(), `SET LOCAL timezone='UTC'`)
	require.NoError(t, err)
	snapshot, err := cfg.captureSelectedBusinessModuleCompleteSnapshot(context.Background(), tx, []string{"configuration"})
	require.NoError(t, err)
	filtered := []selectedBusinessModuleExternalReference{}
	removed := 0
	for _, reference := range snapshot.ExternalDependencies {
		if reference.Table == "users" && (reference.Values["id"] == "900260" || reference.Values["id"] == "900261") {
			removed++
			continue
		}
		filtered = append(filtered, reference)
	}
	require.Equal(t, 2, removed)
	snapshot.ExternalDependencies = filtered
	observation, err := inspectModuleRecoveryDatabaseObservation(context.Background(), tx)
	require.NoError(t, err)
	report := &SelectedBusinessModuleIsolatedRestorePreview{}
	_, blocked, err := inspectSelectedBusinessModuleRowsForRestore(context.Background(), tx, snapshot, observation, report)
	require.NoError(t, err)
	require.True(t, blocked)
	require.Equal(t, 2, report.DependencyConflicts)
}

func verifyRestoredPaymentMethodsAndProtectionThroughReadOnlyBusinessServicesWithoutPaymentOrAuditReplay(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	tx, err := database.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	require.NoError(t, err)
	defer tx.Rollback()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: tx}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	payments := service.NewPaymentService(nil, repository.NewPaymentRepository(db))
	fixed, err := payments.GetPaymentMethod(900260)
	require.NoError(t, err)
	require.Equal(t, int64(9007199254740993), fixed.MinAmountMinor)
	require.Equal(t, int64(901), fixed.FeeValueMinor)
	require.Contains(t, fixed.Settings, configurationPaymentMethodOriginalSyntheticSettings)
	fee, err := fixed.CalculateFeeMoney(domainmoney.MustNew(9007199254740993, "USD"))
	require.NoError(t, err)
	require.Equal(t, int64(901), fee.AmountMinor())
	percentage, err := payments.GetPaymentMethod(900261)
	require.NoError(t, err)
	require.Equal(t, "2.123456789012345", percentage.FeeRateDecimal)
	fee, err = percentage.CalculateFeeMoney(domainmoney.MustNew(9007199254740993, "USD"))
	require.NoError(t, err)
	require.Equal(t, int64(191263984074667), fee.AmountMinor())
	disabled, err := payments.GetPaymentMethod(900262)
	require.NoError(t, err)
	require.False(t, disabled.Enabled)
	_, err = payments.GetPaymentMethod(900263)
	require.True(t, repository.IsRecordNotFound(err))
	enabled, err := payments.ListPaymentMethods(true)
	require.NoError(t, err)
	for _, method := range enabled {
		require.NotContains(t, []uint{900262, 900263, 900269}, method.ID)
	}
	protection := service.NewPaymentProtectionService(repository.NewPaymentProtectionRepository(db), config.PaymentProtectionConfig{Enabled: true})
	controls, err := protection.ListControls(time.Now().UTC(), true)
	require.NoError(t, err)
	indexed := map[uint]paymentdomain.PaymentProtectionControl{}
	for _, control := range controls {
		indexed[control.ID] = control
	}
	require.Len(t, indexed, 5)
	require.Equal(t, "active", indexed[900260].Status)
	require.Equal(t, "expired", indexed[900261].Status)
	require.Equal(t, "revoked", indexed[900262].Status)
	require.Equal(t, "2025-01-01T01:00:00Z", indexed[900261].ExpiresAt.UTC().Format(time.RFC3339))
	require.Equal(t, configurationPaymentProtectionOriginalSyntheticReason, indexed[900260].Reason)
	actor := uint(900260)
	if scenario == "payment_system_actor_restore" {
		actor = 0
	}
	require.Equal(t, actor, indexed[900260].CreatedBy)
	decision, err := protection.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{Provider: "stripe", Country: "NZ", PaymentMethod: "card"})
	require.NoError(t, err)
	require.True(t, decision.Force3DS)
	require.False(t, decision.PausePayment)
	decision, err = protection.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{Provider: "paypal", Country: "NZ", PaymentMethod: "card"})
	require.NoError(t, err)
	require.False(t, decision.Force3DS || decision.PausePayment)
	decision, err = protection.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{Provider: "stripe", Country: "GB", PaymentMethod: "card"})
	require.NoError(t, err)
	require.True(t, decision.Force3DS && decision.PausePayment)
	var retained int
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT count(*) FROM transactions WHERE (id=900268 AND status='completed' AND completed_at='2025-01-02 03:04:05Z' AND provider_request_key='original-completed-request') OR (id=900269 AND status='requires_action' AND completed_at IS NULL AND provider_request_key='original-pending-request')`).Scan(&retained))
	require.Equal(t, 2, retained)
	var auditJSON string
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT coalesce(jsonb_agg(to_jsonb(a)),'[]'::jsonb)::text FROM audit_logs a WHERE resource='configuration_module_site_content_restore'`).Scan(&auditJSON))
	for _, confidential := range []string{configurationPaymentMethodOriginalSyntheticSettings, configurationPaymentProtectionOriginalSyntheticReason} {
		require.NotContains(t, auditJSON, confidential)
	}
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT count(*) FROM audit_logs WHERE resource='payment_protection_control' AND resource_id BETWEEN 900260 AND 900269`).Scan(&retained))
	require.Zero(t, retained, "recovery must not replay create/revoke business audits")
}
