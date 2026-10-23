package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const orderTransactionRowInspectionSyntheticSchema = `
CREATE TABLE schema_migrations(version bigint,dirty boolean); INSERT INTO schema_migrations VALUES(434,false);
CREATE TABLE users(id bigserial PRIMARY KEY,email text);
CREATE TABLE products(id bigserial PRIMARY KEY,name text);
CREATE TABLE product_variants(id bigserial PRIMARY KEY,product_id bigint,stock int);
CREATE TABLE synthetic_order_composite_parents(owner_code text,revision int,PRIMARY KEY(owner_code,revision));
CREATE TABLE orders(id bigserial PRIMARY KEY,order_number text UNIQUE,user_id bigint,customer_note text,deleted_at timestamptz);
CREATE TABLE order_items(id bigserial PRIMARY KEY,order_id bigint,product_id bigint,variant_id bigint,quantity int,currency text,price_minor bigint,owner_code text,revision int,FOREIGN KEY(owner_code,revision) REFERENCES synthetic_order_composite_parents(owner_code,revision));
CREATE TABLE transactions(id bigserial PRIMARY KEY,order_id bigint,amount_minor bigint,currency text,status text,gateway_response text);
CREATE TABLE refunds(id bigserial PRIMARY KEY,order_id bigint,transaction_id bigint,amount_minor bigint,status text);
CREATE TABLE refund_line_items(id bigserial PRIMARY KEY,order_id bigint,refund_id bigint,order_item_id bigint,product_id bigint,variant_id bigint,quantity int);
CREATE TABLE payment_refund_executions(id bigserial PRIMARY KEY,order_id bigint,refund_id bigint,transaction_id bigint,status text,provider text DEFAULT 'synthetic-provider',idempotency_key text DEFAULT 'private-refund-execution-key',attempt_count int DEFAULT 1);
CREATE TABLE after_sales_cases(id bigserial PRIMARY KEY,order_id bigint);
CREATE TABLE after_sales_case_items(id bigserial PRIMARY KEY,order_id bigint,case_id bigint,order_item_id bigint,product_id bigint,variant_id bigint,quantity int);
CREATE TABLE after_sales_refund_reviews(id bigserial PRIMARY KEY,case_id bigint,linked_refund_id bigint);
CREATE TABLE order_evidence_snapshots(id bigserial PRIMARY KEY,order_id bigint);
CREATE TABLE order_evidence_packages(id bigserial PRIMARY KEY,order_id bigint,snapshot_id bigint);
CREATE TABLE order_evidence_items(id bigserial PRIMARY KEY,order_id bigint,package_id bigint,order_item_id bigint,snapshot_id bigint);
CREATE TABLE order_evidence_attachments(id bigserial PRIMARY KEY,evidence_item_id bigint,storage_key text,size_bytes bigint,sha256 text);
CREATE TABLE after_sales_case_attachments(id bigserial PRIMARY KEY,case_id bigint,storage_url text,size_bytes bigint);
CREATE TABLE coupons(id bigserial PRIMARY KEY);
CREATE TABLE coupon_usage(id bigserial PRIMARY KEY,order_id bigint REFERENCES orders(id),user_id bigint,coupon_id bigint);
CREATE TABLE referrals(id bigserial PRIMARY KEY,completed_order_id bigint);
CREATE TABLE referral_records(id bigserial PRIMARY KEY,order_id bigint);
CREATE TABLE referral_reward_logs(id bigserial PRIMARY KEY,referral_record_id bigint REFERENCES referral_records(id));
CREATE TABLE referral_transitions(id bigserial PRIMARY KEY,referral_record_id bigint REFERENCES referral_records(id));
CREATE TABLE loyalty_transactions(id bigserial PRIMARY KEY,source text,source_id bigint,user_id bigint,points int,description text);
CREATE TABLE payment_operation_idempotencies(id bigserial PRIMARY KEY,status text,response_body text,user_id bigint DEFAULT 3,scope text DEFAULT 'private-capture-scope',idempotency_key text DEFAULT 'private-capture-key',request_hash text DEFAULT repeat('a',64),status_code int DEFAULT 200,content_type text DEFAULT 'application/json',claim_token text DEFAULT '',lease_expires_at timestamptz,reconciliation_started_at timestamptz,created_at timestamptz DEFAULT '2026-10-01',updated_at timestamptz DEFAULT '2026-10-01',UNIQUE(user_id,scope,idempotency_key));
CREATE TABLE refund_idempotencies(id bigserial PRIMARY KEY,admin_user_id bigint,scope text,idempotency_key text,request_hash text,refund_id bigint,created_at timestamptz DEFAULT '2026-10-01',updated_at timestamptz DEFAULT '2026-10-01',UNIQUE(scope,idempotency_key));
CREATE TABLE email_delivery_records(id bigserial PRIMARY KEY,outbox_event_id bigint,event_key text UNIQUE,event_type text,template_code text,template_locale text DEFAULT 'en',template_version int DEFAULT 1,recipient_email text,subject text,reference_type text,reference_number text,status text,attempt_count int DEFAULT 1,provider_code text DEFAULT 'synthetic-mail-provider',last_error text DEFAULT '',first_attempt_at timestamptz DEFAULT '2026-10-01',last_attempt_at timestamptz DEFAULT '2026-10-01',sent_at timestamptz DEFAULT '2026-10-01',created_at timestamptz DEFAULT '2026-10-01',updated_at timestamptz DEFAULT '2026-10-01');
CREATE TABLE payment_risk_snapshots(id bigserial PRIMARY KEY,provider text);
CREATE TABLE outbox_events(id bigserial PRIMARY KEY,status text,payload jsonb,event_key text UNIQUE,event_type text,aggregate_type text,aggregate_id text,attempts int DEFAULT 0,max_attempts int DEFAULT 10,available_at timestamptz DEFAULT '2026-10-01',locked_at timestamptz,locked_by text DEFAULT '',processed_at timestamptz,last_attempt_at timestamptz,uncertain_at timestamptz,reconcile_after timestamptz,deleted_at timestamptz);
CREATE TABLE synthetic_preserved_order_references(id bigserial PRIMARY KEY,order_id bigint REFERENCES orders(id));
`

const orderTransactionRowInspectionSyntheticRows = `
INSERT INTO users VALUES(1,'private-order-customer@example.invalid'),(2,'other-private@example.invalid'),(3,'unrelated-private@example.invalid');
INSERT INTO products VALUES(1,'private historical product'),(2,'other product'),(3,'unrelated product');
INSERT INTO product_variants VALUES(1,1,20),(2,2,30),(3,3,40);
INSERT INTO synthetic_order_composite_parents VALUES('private-composite-owner',7);
INSERT INTO orders VALUES(1,'private-synthetic-order-one',1,'private order customer note','2026-10-01'),(2,'private-synthetic-order-two',2,'private order customer note',NULL);
INSERT INTO order_items VALUES(1,1,1,1,3,'USD',100,'private-composite-owner',7),(2,2,2,2,2,'USD',100,'private-composite-owner',7);
INSERT INTO transactions VALUES(1,1,300,'USD','completed','private gateway response'),(2,2,200,'USD','completed','private gateway response');
INSERT INTO refunds VALUES(1,1,1,100,'completed');
INSERT INTO refund_line_items VALUES(1,1,1,1,1,1,1);
INSERT INTO payment_refund_executions(id,order_id,refund_id,transaction_id,status) VALUES(1,1,1,1,'succeeded');
INSERT INTO after_sales_cases VALUES(1,1);
INSERT INTO after_sales_case_items VALUES(1,1,1,1,1,1,1);
INSERT INTO after_sales_refund_reviews VALUES(1,1,1);
INSERT INTO order_evidence_snapshots VALUES(1,1);
INSERT INTO order_evidence_packages VALUES(1,1,1);
INSERT INTO order_evidence_items VALUES(1,1,1,1,1);
INSERT INTO order_evidence_attachments VALUES(1,1,'order-evidence/1/1/synthetic-proof.bin',8,encode(sha256('original'::bytea),'hex'));
INSERT INTO coupons VALUES(1);
INSERT INTO coupon_usage VALUES(1,1,1,1);
INSERT INTO loyalty_transactions VALUES(1,'order',1,1,10,'private loyalty description'),(2,'checkin',999,3,5,'private unrelated description');
INSERT INTO payment_operation_idempotencies(id,status,response_body) VALUES(1,'completed','private idempotency response');
INSERT INTO payment_risk_snapshots VALUES(1,'synthetic-provider');
INSERT INTO outbox_events(id,status,payload,event_key,event_type,aggregate_type,aggregate_id) VALUES(1,'processed','{"private":"private event body"}','private-unrelated-event','merchant.product_upsert','product','1');
`

func TestPostgreSQLOrderTransactionRowsProtectFinancialInventoryAndEvidenceFactsWithoutMutations(t *testing.T) {
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
		_, err := control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(name))
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(name)+` WITH (FORCE)`)
		})
		database, err := cfg.connectDatabase(ctx, name)
		require.NoError(t, err)
		t.Cleanup(func() { database.Close() })
		databases = append(databases, database)
		_, err = database.ExecContext(ctx, orderTransactionRowInspectionSyntheticSchema)
		require.NoError(t, err)
		for table, column := range registeredOrderTransactionOwnershipColumns() {
			_, err := database.ExecContext(ctx, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(id bigserial PRIMARY KEY,%s bigint)", pq.QuoteIdentifier(table), pq.QuoteIdentifier(column)))
			require.NoError(t, err)
		}
	}
	verifiedDirectory := filepath.Join(t.TempDir(), "synthetic-order-verified-package")
	key := "order-evidence/1/1/synthetic-proof.bin"
	var manifest Manifest
	resetSyntheticOrderRecoveryDatabasesAndFiles := func() {
		for _, database := range databases {
			var tables []string
			rows, err := database.QueryContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename<>'schema_migrations' ORDER BY tablename`)
			require.NoError(t, err)
			for rows.Next() {
				var table string
				require.NoError(t, rows.Scan(&table))
				tables = append(tables, pq.QuoteIdentifier(table))
			}
			require.NoError(t, rows.Err())
			require.NoError(t, rows.Close())
			_, err = database.ExecContext(ctx, "TRUNCATE "+strings.Join(tables, ",")+" RESTART IDENTITY CASCADE;"+orderTransactionRowInspectionSyntheticRows)
			require.NoError(t, err)
		}
		for _, filename := range []string{filepath.Join(verifiedDirectory, "media", "uploads_private", filepath.FromSlash(key)), filepath.Join(cfg.StorageRoots["uploads_private"], filepath.FromSlash(key))} {
			require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
			require.NoError(t, os.WriteFile(filename, []byte("original"), 0600))
		}
		file, err := fileIdentity(filepath.Join(verifiedDirectory, "media", "uploads_private", filepath.FromSlash(key)), "media/uploads_private/"+key)
		require.NoError(t, err)
		manifest = Manifest{Files: []FileIdentity{file}}
	}
	inspectSyntheticOrderRecoverySnapshots := func(limits moduleRecoveryRowInspectionLimits) *ModuleRecoveryRowInspection {
		var transactions []*sql.Tx
		var observations []*moduleRecoveryDatabaseObservation
		for _, database := range databases {
			transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
			require.NoError(t, err)
			defer transaction.Rollback()
			observation, err := inspectModuleRecoveryDatabaseObservation(ctx, transaction)
			require.NoError(t, err)
			transactions = append(transactions, transaction)
			observations = append(observations, observation)
		}
		report, err := cfg.inspectOrderTransactionRows(ctx, transactions[0], transactions[1], observations[0], observations[1], &manifest, verifiedDirectory, limits)
		require.NoError(t, err)
		require.Equal(t, "preserve_financial_inventory_and_idempotency_facts_no_replay", report.StatePolicy)
		require.NotNil(t, report.DurableEvents)
		require.Zero(t, report.DurableEvents.SourceRelatedEvents, "unrelated global events remain outside order diagnostics")
		encoded, err := json.Marshal(report)
		require.NoError(t, err)
		for _, value := range []string{"private-order-customer@example.invalid", "private-synthetic-order-one", "private order customer note", "private gateway response", "private-composite-owner", "private event body", "synthetic-proof.bin", "private idempotency response"} {
			require.NotContains(t, string(encoded), value)
		}
		for index, database := range databases {
			transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
			require.NoError(t, err)
			_, fingerprint, _, err := captureDatabaseTransactionInventory(ctx, transaction)
			require.NoError(t, err)
			require.NoError(t, transaction.Rollback())
			require.Equal(t, observations[index].Fingerprint, fingerprint, "all table and sequence facts must remain unchanged")
		}
		return report
	}
	for _, scenario := range []struct{ name, statement, code string }{
		{"matching_retained_orders_composite_dependencies_and_evidence", "SELECT 1", ""},
		{"unreferenced_customer_changes", "UPDATE users SET email='changed-unrelated@example.invalid' WHERE id=3", ""},
		{"unreferenced_inventory_changes", "UPDATE product_variants SET stock=1 WHERE id=3", ""},
		{"global_payment_idempotency_requires_reconciliation_and_stays_preserved", "UPDATE payment_operation_idempotencies SET status='pending';UPDATE outbox_events SET status='pending';UPDATE payment_risk_snapshots SET provider='changed'", "payment_receipt_external_outcome_requires_reconciliation"},
		{"referenced_customer_changes", "UPDATE users SET email='changed-linked@example.invalid' WHERE id=1", "shared_dependency_row_differs"},
		{"referenced_inventory_changes", "UPDATE product_variants SET stock=19 WHERE id=1", "shared_dependency_row_differs"},
		{"historical_soft_deleted_order_changes", "UPDATE orders SET customer_note='new historical note' WHERE id=1", "order_transaction_row_content_differs"},
		{"order_number_reassigned_at_same_primary_key", "UPDATE orders SET order_number='another-order' WHERE id=1", "row_identity_reassigned"},
		{"order_number_now_belongs_to_another_primary_key", "UPDATE orders SET order_number='temporary-order-number' WHERE id=1;UPDATE orders SET order_number='private-synthetic-order-one' WHERE id=2", "order_number_primary_key_identity_conflict"},
		{"order_line_moves_to_another_order", "UPDATE order_items SET order_id=2 WHERE id=1", "row_identity_reassigned"},
		{"zero_order_owner_is_not_an_unbound_global_record", "UPDATE transactions SET order_id=0 WHERE id=1", "order_owner_identity_invalid"},
		{"refund_transaction_crosses_orders", "UPDATE refunds SET transaction_id=2 WHERE id=1", "transaction_reference_crosses_orders"},
		{"refund_line_crosses_orders", "UPDATE refund_line_items SET order_item_id=2 WHERE id=1", "transaction_reference_crosses_orders"},
		{"refund_execution_crosses_orders", "UPDATE payment_refund_executions SET transaction_id=2 WHERE id=1", "transaction_reference_crosses_orders"},
		{"after_sales_line_crosses_orders", "UPDATE after_sales_case_items SET order_item_id=2 WHERE id=1", "transaction_reference_crosses_orders"},
		{"after_sales_review_refund_crosses_orders", "INSERT INTO refunds VALUES(2,2,2,100,'completed');UPDATE after_sales_refund_reviews SET linked_refund_id=2", "after_sales_refund_review_crosses_orders"},
		{"evidence_package_crosses_orders", "UPDATE order_evidence_packages SET order_id=2 WHERE id=1", "transaction_reference_crosses_orders"},
		{"variant_belongs_to_other_product", "UPDATE order_items SET variant_id=2 WHERE id=1", "order_variant_product_identity_conflict"},
		{"refund_product_snapshot_is_inconsistent", "UPDATE refund_line_items SET product_id=2 WHERE id=1", "refund_or_after_sales_product_identity_conflict"},
		{"refund_quantity_exceeds_original_order", "UPDATE refund_line_items SET quantity=4 WHERE id=1", "refund_or_after_sales_quantity_exceeds_order_line"},
		{"negative_minor_amount_is_not_recomputed", "UPDATE transactions SET amount_minor=-1 WHERE id=1", "order_monetary_minor_snapshot_invalid"},
		{"non_fk_transaction_reference_missing", "UPDATE refunds SET transaction_id=999 WHERE id=1", "referenced_row_missing"},
		{"non_fk_product_reference_missing", "DELETE FROM products WHERE id=1", "referenced_row_missing"},
		{"non_fk_customer_reference_missing", "DELETE FROM users WHERE id=1", "referenced_row_missing"},
		{"preserved_business_row_references_order", "INSERT INTO synthetic_preserved_order_references VALUES(1,1)", "preserved_row_references_order_object"},
		{"invalid_order_line_quantity", "UPDATE order_items SET quantity=0 WHERE id=1", "order_line_quantity_invalid"},
		{"oversized_order_identity_projection_is_not_transferred", "UPDATE orders SET order_number=repeat('private-order-number',5000) WHERE id=1", "row_projection_too_large"},
		{"oversized_gateway_response_does_not_leave_database", "UPDATE transactions SET gateway_response=repeat('private gateway response',100000)", "order_transaction_row_content_differs"},
		{"new_target_payment_fact", "INSERT INTO transactions VALUES(3,1,300,'USD','completed','private newer gateway fact')", "target_only_order_transaction_row"},
		{"new_target_shared_loyalty_fact", "INSERT INTO loyalty_transactions VALUES(3,'order',1,1,2,'private new reward')", "target_only_shared_order_fact"},
		{"new_target_order_with_existing_customer", "INSERT INTO orders VALUES(3,'private-new-order',1,'private new note',NULL)", "target_only_order_transaction_row"},
		{"evidence_file_owner_path_conflict", "UPDATE order_evidence_attachments SET storage_key='order-evidence/2/1/wrong.bin'", "order_evidence_attachment_owner_identity_conflict"},
		{"zero_registered_size_is_checked", "UPDATE order_evidence_attachments SET size_bytes=0", "target_attachment_registration_content_conflict"},
		{"evidence_digest_does_not_match_actual_file", "UPDATE order_evidence_attachments SET sha256=repeat('a',64)", "target_attachment_registration_content_conflict"},
		{"external_after_sales_reference_is_incomplete", "INSERT INTO after_sales_case_attachments VALUES(1,1,'https://foreign.invalid/private.jpg',8)", "order_attachment_reference_unresolved"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			resetSyntheticOrderRecoveryDatabasesAndFiles()
			_, err := databases[1].ExecContext(ctx, scenario.statement)
			require.NoError(t, err)
			report := inspectSyntheticOrderRecoverySnapshots(defaultModuleRecoveryRowInspectionLimits())
			if scenario.code == "" {
				require.True(t, report.Complete, "%+v", report.IssueCounts)
				require.Zero(t, report.BlockingIssues, "%+v", report.IssueCounts)
				require.Equal(t, 1, report.Files.PackageVerifiedFiles)
			} else {
				require.Positive(t, report.IssueCounts[scenario.code], "%+v", report.IssueCounts)
			}
			if scenario.name == "oversized_gateway_response_does_not_leave_database" {
				require.True(t, report.Complete)
				require.Zero(t, report.IssueCounts["row_projection_too_large"])
			}
		})
	}
	for _, scenario := range []struct {
		name, code string
		change     func()
	}{
		{"same_path_changed_bytes", "target_attachment_file_content_conflict", func() {
			require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads_private"], filepath.FromSlash(key)), []byte("changed!"), 0600))
		}},
		{"target_missing_evidence_file", "target_attachment_file_missing", func() {
			require.NoError(t, os.Remove(filepath.Join(cfg.StorageRoots["uploads_private"], filepath.FromSlash(key))))
		}},
		{"package_file_changed_after_verification", "verified_package_attachment_changed_after_validation", func() {
			require.NoError(t, os.WriteFile(filepath.Join(verifiedDirectory, "media", "uploads_private", filepath.FromSlash(key)), []byte("changed!"), 0600))
		}},
		{"unreferenced_target_path_still_conflicts_with_package", "target_attachment_file_content_conflict", func() {
			_, err := databases[1].ExecContext(ctx, "DELETE FROM order_evidence_attachments")
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads_private"], filepath.FromSlash(key)), []byte("changed!"), 0600))
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			resetSyntheticOrderRecoveryDatabasesAndFiles()
			scenario.change()
			report := inspectSyntheticOrderRecoverySnapshots(defaultModuleRecoveryRowInspectionLimits())
			require.Positive(t, report.IssueCounts[scenario.code])
		})
	}
	t.Run("resource_limits_and_comparison_samples_remain_explicit", func(t *testing.T) {
		resetSyntheticOrderRecoveryDatabasesAndFiles()
		_, err := databases[1].ExecContext(ctx, "UPDATE refunds SET transaction_id=2;UPDATE product_variants SET stock=19 WHERE id=1")
		require.NoError(t, err)
		limits := defaultModuleRecoveryRowInspectionLimits()
		limits.IssueSamples = 1
		report := inspectSyntheticOrderRecoverySnapshots(limits)
		require.True(t, report.IssuesTruncated)
		require.Len(t, report.Issues, 1)
		limits.Rows = 10
		report = inspectSyntheticOrderRecoverySnapshots(limits)
		require.False(t, report.Complete)
		require.Positive(t, report.IssueCounts["row_inspection_resource_limit"])
		limits = defaultModuleRecoveryRowInspectionLimits()
		limits.ProjectionBytes = 500
		report = inspectSyntheticOrderRecoverySnapshots(limits)
		require.False(t, report.Complete)
	})
	if directory := os.Getenv("LOCAL_BACKUP_MODULE_REPORT_VALIDATION_OUTPUT_DIRECTORY"); directory != "" {
		resetSyntheticOrderRecoveryDatabasesAndFiles()
		_, err := databases[1].ExecContext(ctx, "UPDATE refunds SET transaction_id=2;UPDATE product_variants SET stock=19 WHERE id=1")
		require.NoError(t, err)
		report := inspectSyntheticOrderRecoverySnapshots(defaultModuleRecoveryRowInspectionLimits())
		require.NoError(t, writeDurableJSON(filepath.Join(directory, "synthetic-order-transaction-row-refund-inventory-and-evidence-conflicts-from-postgresql-integration.json"), report))
	}
}
