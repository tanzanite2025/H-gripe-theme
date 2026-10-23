package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const moduleRecoverySideEffectReceiptSyntheticRows = `
INSERT INTO email_delivery_records(id,outbox_event_id,event_key,event_type,template_code,recipient_email,subject,reference_type,reference_number,status)
VALUES(1,10,'notification:private-order-completion-key:order_completed','order.completed','order_completed','private-receipt-recipient@example.invalid','private rendered mail subject','order','private-synthetic-order-one','sent');
INSERT INTO refund_idempotencies(id,admin_user_id,scope,idempotency_key,request_hash,refund_id)
VALUES(1,3,'private-refund-scope','private-admin-refund-key',repeat('b',64),1),(2,3,'private-refund-scope','private-unbound-refund-key',repeat('c',64),NULL);
`

func TestPostgreSQLModuleRecoveryPreservesGlobalPaymentRefundAndNotificationReceiptsWithoutReplay(t *testing.T) {
	if os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" || os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL") == "" {
		t.Skip("requires an explicitly disposable PostgreSQL 15 server")
	}
	ctx := context.Background()
	cfg := createIndependentRecoveryTestConfiguration(t)
	cfg.DatabaseURL = os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	control, err := cfg.connectDatabase(ctx, "postgres")
	require.NoError(t, err)
	t.Cleanup(func() { control.Close() })
	var databases []*sql.DB
	for range 2 {
		name := "local_backup_test_" + newTaskID()[:12]
		_, err := control.ExecContext(ctx, "CREATE DATABASE "+pq.QuoteIdentifier(name))
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = control.ExecContext(ctx, "DROP DATABASE IF EXISTS "+pq.QuoteIdentifier(name)+" WITH (FORCE)")
		})
		database, err := cfg.connectDatabase(ctx, name)
		require.NoError(t, err)
		t.Cleanup(func() { database.Close() })
		databases = append(databases, database)
		_, err = database.ExecContext(ctx, orderTransactionRowInspectionSyntheticSchema)
		require.NoError(t, err)
	}
	resetSyntheticModuleRecoveryReceiptDatabases := func() {
		for _, database := range databases {
			rows, err := database.QueryContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename<>'schema_migrations' ORDER BY tablename`)
			require.NoError(t, err)
			var tables []string
			for rows.Next() {
				var table string
				require.NoError(t, rows.Scan(&table))
				tables = append(tables, pq.QuoteIdentifier(table))
			}
			require.NoError(t, rows.Err())
			require.NoError(t, rows.Close())
			_, err = database.ExecContext(ctx, "TRUNCATE "+strings.Join(tables, ",")+" RESTART IDENTITY CASCADE;"+orderTransactionRowInspectionSyntheticRows+orderDurableEventCompletedSyntheticRow+moduleRecoverySideEffectReceiptSyntheticRows)
			require.NoError(t, err)
		}
	}
	inspectSyntheticModuleRecoveryReceiptDatabases := func(limits moduleRecoveryRowInspectionLimits) *ModuleRecoveryRowInspection {
		report := &ModuleRecoveryRowInspection{Complete: true, IssueSampleLimit: limits.IssueSamples, IssueCounts: map[string]int{}, Issues: []ModuleRecoveryRowIssue{}}
		var graphs []*moduleRecoveryRowDependencyGraph
		var fingerprints []string
		for index, database := range databases {
			transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
			require.NoError(t, err)
			defer transaction.Rollback()
			observation, err := inspectModuleRecoveryDatabaseObservation(ctx, transaction)
			require.NoError(t, err)
			fingerprints = append(fingerprints, observation.Fingerprint)
			side := "package"
			if index == 1 {
				side = "target"
			}
			graphs = append(graphs, newModuleRecoveryRowDependencyGraph(transaction, observation, side, report, limits))
		}
		require.NoError(t, inspectModuleRecoveryPaymentRefundAndNotificationReceiptPreservation(ctx, graphs, report))
		require.NotNil(t, report.SideEffectReceipts)
		require.Equal(t, "preserve_all_target_receipts_no_automatic_import_or_replay", report.SideEffectReceipts.StatePolicy)
		encoded, err := json.Marshal(report)
		require.NoError(t, err)
		for _, secret := range []string{"private-capture-key", "private-capture-scope", "private-admin-refund-key", "private-unbound-refund-key", "private-refund-scope", "private-receipt-recipient@example.invalid", "private rendered mail subject", "private idempotency response", "private-order-completion-key", "private-order-completion-event", "private-synthetic-order-one", "private claim token", "private new response", "private provider error"} {
			require.NotContains(t, string(encoded), secret)
		}
		for index, database := range databases {
			transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
			require.NoError(t, err)
			_, fingerprint, _, err := captureDatabaseTransactionInventory(ctx, transaction)
			require.NoError(t, err)
			require.NoError(t, transaction.Rollback())
			require.Equal(t, fingerprints[index], fingerprint, "receipt, event, business rows and sequences must not change")
			var registrationGift int
			require.NoError(t, database.QueryRowContext(ctx, "SELECT points FROM loyalty_transactions WHERE source='referral_referee'").Scan(&registrationGift))
			require.Equal(t, 50, registrationGift)
		}
		return report
	}
	for _, scenario := range []struct {
		name, statement, code  string
		incomplete, reviewOnly bool
	}{
		{"matching_receipts_keep_other_customer_capture_and_unbound_refund", "SELECT 1", "", false, false},
		{"sent_receipt_with_pending_event_blocks_resending", "UPDATE outbox_events SET status='pending' WHERE id=10", "notification_sent_receipt_has_replayable_event", false, false},
		{"sent_receipt_with_failed_event_still_blocks_resending", "UPDATE outbox_events SET status='failed' WHERE id=10", "notification_sent_receipt_has_replayable_event", false, false},
		{"ignored_event_keeps_sent_receipt", "UPDATE outbox_events SET status='ignored' WHERE id=10", "", false, false},
		{"consumed_event_keeps_sent_receipt", "UPDATE outbox_events SET status='pending',deleted_at='2026-10-02' WHERE id=10", "", false, false},
		{"mail_audit_missing_does_not_prove_unsent", "DELETE FROM email_delivery_records", "notification_audit_absence_not_delivery_evidence", false, true},
		{"historic_event_missing_keeps_receipt_without_recreation", "DELETE FROM outbox_events WHERE id=10", "notification_receipt_event_unavailable", false, true},
		{"sending_mail_needs_external_reconciliation", "UPDATE email_delivery_records SET status='sending',sent_at=NULL", "notification_receipt_outcome_requires_reconciliation", false, false},
		{"unknown_mail_needs_external_reconciliation", "UPDATE email_delivery_records SET status='unknown',sent_at=NULL", "notification_receipt_outcome_requires_reconciliation", false, false},
		{"unknown_mail_status_marks_incomplete", "UPDATE email_delivery_records SET status='future-status'", "notification_receipt_status_unregistered", true, false},
		{"sent_mail_requires_sent_timestamp", "UPDATE email_delivery_records SET sent_at=NULL", "notification_receipt_sent_time_missing", false, false},
		{"mail_outbox_key_is_not_notification_key", "UPDATE email_delivery_records SET event_key='private-order-completion-event'", "notification_receipt_event_identity_conflict", false, false},
		{"mail_wrong_template_is_identity_conflict", "UPDATE email_delivery_records SET template_code='order_delivered'", "notification_receipt_event_identity_conflict", false, false},
		{"mail_wrong_business_reference_is_identity_conflict", "UPDATE email_delivery_records SET reference_number='private-other-order'", "notification_receipt_business_reference_conflict", false, false},
		{"mail_wrong_event_type_is_identity_conflict", "UPDATE email_delivery_records SET event_type='order.delivered'", "notification_receipt_event_identity_conflict", false, false},
		{"mail_event_id_reused_by_unrelated_fact", "UPDATE email_delivery_records SET outbox_event_id=1", "notification_receipt_event_identity_conflict", false, false},
		{"mail_event_number_zero_is_invalid", "UPDATE email_delivery_records SET outbox_event_id=0", "notification_receipt_event_identity_invalid", false, false},
		{"provider_error_change_preserves_current_row_privately", "UPDATE email_delivery_records SET last_error='private provider error'", "side_effect_receipt_current_state_differs", false, false},
		{"global_capture_for_unrelated_customer_is_not_ignored", "UPDATE payment_operation_idempotencies SET response_body='private new response'", "side_effect_receipt_current_state_differs", false, false},
		{"pending_capture_with_expired_lease_not_replayed", "UPDATE payment_operation_idempotencies SET status='pending',claim_token='private claim token',lease_expires_at='2020-01-01'", "payment_receipt_external_outcome_requires_reconciliation", false, false},
		{"reconciling_capture_not_replayed", "UPDATE payment_operation_idempotencies SET status='reconciling'", "payment_receipt_external_outcome_requires_reconciliation", false, false},
		{"completed_capture_with_claim_blocks", "UPDATE payment_operation_idempotencies SET claim_token='private claim token'", "completed_payment_receipt_execution_lease_present", false, false},
		{"unknown_capture_status_marks_incomplete", "UPDATE payment_operation_idempotencies SET status='future-status'", "payment_receipt_status_unregistered", true, false},
		{"same_capture_key_different_request_hash_blocks", "UPDATE payment_operation_idempotencies SET request_hash=repeat('d',64)", "side_effect_receipt_natural_key_identity_conflict", false, false},
		{"capture_request_digest_must_be_hex", "UPDATE payment_operation_idempotencies SET request_hash=repeat('z',64)", "payment_receipt_request_identity_invalid", false, false},
		{"capture_same_user_new_scope_is_independent", "INSERT INTO payment_operation_idempotencies(id,status,response_body,scope) VALUES(2,'completed','private new response','private-other-scope')", "", false, false},
		{"capture_same_key_other_user_is_independent", "INSERT INTO payment_operation_idempotencies(id,status,response_body,user_id) VALUES(2,'completed','private new response',1)", "", false, false},
		{"target_missing_receipt_is_not_automatically_imported", "DELETE FROM payment_operation_idempotencies", "source_side_effect_receipt_missing_from_target", false, false},
		{"same_receipt_key_moved_primary_id", "UPDATE payment_operation_idempotencies SET id=9", "side_effect_receipt_identifier_moved", false, false},
		{"same_primary_id_reassigned_to_other_key", "UPDATE payment_operation_idempotencies SET idempotency_key='private-other-key'", "side_effect_receipt_primary_key_reassigned", false, false},
		{"refund_natural_key_ignores_admin_but_admin_identity_must_match", "UPDATE refund_idempotencies SET admin_user_id=1 WHERE id=1", "side_effect_receipt_natural_key_identity_conflict", false, false},
		{"unbound_refund_key_never_loses_request_hash_fence", "UPDATE refund_idempotencies SET request_hash=repeat('d',64) WHERE id=2", "side_effect_receipt_natural_key_identity_conflict", false, false},
		{"refund_link_cannot_change_under_same_key", "UPDATE refund_idempotencies SET refund_id=NULL WHERE id=1", "side_effect_receipt_natural_key_identity_conflict", false, false},
		{"missing_refund_blocks_receipt_rebinding", "UPDATE refund_idempotencies SET refund_id=999 WHERE id=1", "refund_receipt_business_reference_missing", false, false},
		{"zero_refund_is_invalid_not_global", "UPDATE refund_idempotencies SET refund_id=0 WHERE id=1", "refund_receipt_business_identity_invalid", false, false},
		{"missing_natural_key_blocks", "UPDATE refund_idempotencies SET idempotency_key='' WHERE id=1", "side_effect_receipt_natural_key_invalid", false, false},
		{"huge_response_is_hashed_without_transfer", "UPDATE payment_operation_idempotencies SET response_body=repeat('private new response',100000)", "side_effect_receipt_current_state_differs", false, false},
		{"huge_natural_key_is_not_transferred", "UPDATE refund_idempotencies SET idempotency_key=repeat('private-key',10000) WHERE id=1", "row_projection_too_large", true, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			resetSyntheticModuleRecoveryReceiptDatabases()
			_, err := databases[1].ExecContext(ctx, scenario.statement)
			require.NoError(t, err)
			report := inspectSyntheticModuleRecoveryReceiptDatabases(defaultModuleRecoveryRowInspectionLimits())
			require.Equal(t, !scenario.incomplete, report.SideEffectReceipts.Complete)
			if scenario.code == "" {
				require.Zero(t, report.BlockingIssues, "%+v", report.IssueCounts)
			} else {
				require.Positive(t, report.IssueCounts[scenario.code], "%+v", report.IssueCounts)
				if scenario.reviewOnly {
					for _, issue := range report.Issues {
						if issue.Code == scenario.code {
							require.Equal(t, "review", issue.Severity)
						}
					}
				} else {
					require.Positive(t, report.BlockingIssues)
				}
			}
			if scenario.name == "sent_receipt_with_pending_event_blocks_resending" {
				require.Positive(t, report.SideEffectReceipts.Tables[0].BlockingIssues, "matching audit rows must still show replayable-event risks")
			}
		})
	}
	for _, scenario := range []struct{ name, statement string }{
		{"historical_notification_key_falls_back_to_outbox_key", `UPDATE outbox_events SET payload=payload-'idempotency_key' WHERE id=10;UPDATE email_delivery_records SET event_key='notification:private-order-completion-event:order_completed'`},
		{"historical_notification_key_falls_back_to_event_number", `UPDATE outbox_events SET payload=payload-'idempotency_key',event_key='' WHERE id=10;UPDATE email_delivery_records SET event_key='notification:event:10:order_completed'`},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			resetSyntheticModuleRecoveryReceiptDatabases()
			for _, database := range databases {
				_, err := database.ExecContext(ctx, scenario.statement)
				require.NoError(t, err)
			}
			report := inspectSyntheticModuleRecoveryReceiptDatabases(defaultModuleRecoveryRowInspectionLimits())
			require.True(t, report.SideEffectReceipts.Complete)
			require.Zero(t, report.BlockingIssues, "%+v", report.IssueCounts)
		})
	}
	for _, scenario := range []struct{ name, statement, code string }{
		{"after_sales_template_and_case_reference_match_producer", "SELECT 1", ""},
		{"after_sales_audit_must_not_reference_another_case", "UPDATE email_delivery_records SET reference_number='2'", "notification_receipt_business_reference_conflict"},
		{"non_notifying_after_sales_state_cannot_have_registered_mail", `UPDATE outbox_events SET payload=payload||'{"new_status":"inspecting"}'::jsonb WHERE id=10`, "notification_receipt_event_identity_conflict"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			resetSyntheticModuleRecoveryReceiptDatabases()
			for _, database := range databases {
				_, err := database.ExecContext(ctx, `UPDATE outbox_events SET event_type='after_sales.status_changed',aggregate_type='after_sales_case',payload=payload||'{"case_id":1,"new_status":"approved"}'::jsonb WHERE id=10;UPDATE email_delivery_records SET event_type='after_sales.status_changed',template_code='after_sales_approved',event_key='notification:private-order-completion-key:after_sales_approved',reference_type='after_sales',reference_number='1'`)
				require.NoError(t, err)
			}
			_, err := databases[1].ExecContext(ctx, scenario.statement)
			require.NoError(t, err)
			report := inspectSyntheticModuleRecoveryReceiptDatabases(defaultModuleRecoveryRowInspectionLimits())
			if scenario.code == "" {
				require.Zero(t, report.BlockingIssues, "%+v", report.IssueCounts)
			} else {
				require.Positive(t, report.IssueCounts[scenario.code])
			}
		})
	}
	t.Run("schema_missing_receipt_identity_is_incomplete", func(t *testing.T) {
		resetSyntheticModuleRecoveryReceiptDatabases()
		_, err := databases[1].ExecContext(ctx, "ALTER TABLE payment_operation_idempotencies DROP COLUMN lease_expires_at")
		require.NoError(t, err)
		defer func() {
			_, err := databases[1].ExecContext(ctx, "ALTER TABLE payment_operation_idempotencies ADD COLUMN lease_expires_at timestamptz")
			require.NoError(t, err)
		}()
		report := inspectSyntheticModuleRecoveryReceiptDatabases(defaultModuleRecoveryRowInspectionLimits())
		require.False(t, report.SideEffectReceipts.Complete)
		require.Positive(t, report.IssueCounts["side_effect_receipt_schema_missing"])
	})
	t.Run("shared_limits_and_bounded_samples_remain_enforced", func(t *testing.T) {
		resetSyntheticModuleRecoveryReceiptDatabases()
		limits := defaultModuleRecoveryRowInspectionLimits()
		limits.IssueSamples = 1
		report := inspectSyntheticModuleRecoveryReceiptDatabases(limits)
		require.True(t, report.SideEffectReceipts.ComparisonsTruncated)
		require.Equal(t, 4, report.SideEffectReceipts.ComparedIdentities)
		limits.Rows = 1
		report = inspectSyntheticModuleRecoveryReceiptDatabases(limits)
		require.False(t, report.SideEffectReceipts.Complete)
		limits = defaultModuleRecoveryRowInspectionLimits()
		limits.ProjectionBytes = 100
		report = inspectSyntheticModuleRecoveryReceiptDatabases(limits)
		require.False(t, report.SideEffectReceipts.Complete)
	})
	if directory := os.Getenv("LOCAL_BACKUP_MODULE_REPORT_VALIDATION_OUTPUT_DIRECTORY"); directory != "" {
		resetSyntheticModuleRecoveryReceiptDatabases()
		_, err := databases[1].ExecContext(ctx, "UPDATE outbox_events SET status='pending' WHERE id=10;UPDATE payment_operation_idempotencies SET status='reconciling',claim_token='private claim token'")
		require.NoError(t, err)
		report := inspectSyntheticModuleRecoveryReceiptDatabases(defaultModuleRecoveryRowInspectionLimits())
		require.NoError(t, writeDurableJSON(filepath.Join(directory, "synthetic-module-recovery-payment-refund-notification-receipt-preservation-conflicts.json"), report))
	}
}
