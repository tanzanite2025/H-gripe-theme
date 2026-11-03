package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestOrderModuleNormalConstraintRestoreRequiresCurrentPreviewPolicy(t *testing.T) {
	for _, policy := range []string{orderModuleBoundedOrderRestorePolicy, "restore_terminal_and_settled_active_orders_with_original_component_and_master_inventory_dependencies_v5", "restore_terminal_and_settled_active_orders_with_completed_facts_and_original_payment_receipts_v4"} {
		report := &SelectedBusinessModuleIsolatedRestorePreview{Binding: &BusinessModuleRestoreSnapshotBinding{}, TargetUnchanged: true, DisposableStageRemoved: true, IsolatedRestoreVerified: true, Modules: []string{"orders"}, OrderHistoryRestoreEligibility: &OrderModuleHistoryRestoreEligibility{Eligible: true, Policy: policy}}
		var err error
		report.SHA256, err = businessModuleRecoveryCanonicalJSONSHA256(report)
		require.NoError(t, err)
		task := &Task{Kind: "upload", State: "module_isolated_restore_verified", ExpiresAt: time.Now().Add(time.Hour), Manifest: &Manifest{Scope: selectedBusinessModulesBackupScope, SelectedModules: []string{"orders"}}, SelectedModuleRestorePreview: report}
		err = validateOrderModuleFormalHistoryRestorePreview(task, report.SHA256)
		if policy == orderModuleBoundedOrderRestorePolicy {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}

func TestOrderModuleNormalConstraintInsertPlanRequiresParentsBeforeChildrenAndRejectsCycles(t *testing.T) {
	missing := map[string][]json.RawMessage{}
	for _, table := range []string{"orders", "order_items", "transactions", "refunds", "outbox_events", "email_delivery_records"} {
		missing[table] = []json.RawMessage{json.RawMessage(`{}`)}
	}
	references := []ModuleRecoveryDependencyReference{
		{Kind: "foreign_key", SourceTable: "order_items", TargetTable: "orders"},
		{Kind: "foreign_key", SourceTable: "refunds", TargetTable: "transactions"},
	}
	ordered, err := planOrderModuleMissingTableInsertsWithNormalForeignKeyEnforcement(missing, references)
	require.NoError(t, err)
	positions := map[string]int{}
	for index, table := range ordered {
		positions[table] = index
	}
	for _, dependency := range [][2]string{{"order_items", "orders"}, {"transactions", "orders"}, {"refunds", "transactions"}, {"outbox_events", "refunds"}, {"email_delivery_records", "outbox_events"}} {
		require.Greater(t, positions[dependency[0]], positions[dependency[1]])
	}
	_, err = planOrderModuleMissingTableInsertsWithNormalForeignKeyEnforcement(missing, append(references, ModuleRecoveryDependencyReference{Kind: "foreign_key", SourceTable: "orders", TargetTable: "order_items"}))
	require.ErrorContains(t, err, "cyclic table dependency")
}

func readOrderModuleNormalConstraintTestCatalogIdentity(t *testing.T, database *sql.DB) string {
	t.Helper()
	var identity string
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT jsonb_build_object('constraints',(SELECT jsonb_agg(jsonb_build_array(c.oid,c.conrelid,c.confrelid,c.conname,pg_get_constraintdef(c.oid),c.convalidated) ORDER BY c.oid) FROM pg_constraint c JOIN pg_class r ON r.oid=c.conrelid WHERE r.relnamespace='public'::regnamespace),'triggers',(SELECT jsonb_agg(jsonb_build_array(t.oid,t.tgrelid,t.tgconstraint,t.tgenabled) ORDER BY t.oid) FROM pg_trigger t JOIN pg_class r ON r.oid=t.tgrelid WHERE r.relnamespace='public'::regnamespace))::text`).Scan(&identity))
	return identity
}

func TestPostgreSQLOrderModuleNormalConstraintWriterRejectsInvalidRowsAndPreservesConstraintIdentity(t *testing.T) {
	connectionURL := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	if connectionURL == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires a disposable PostgreSQL instance")
	}
	for _, scenario := range []string{"parent_first_self_reference_and_transactional_rollback", "missing_foreign_parent", "negative_quantity", "duplicate_order_identity", "disabled_foreign_key_trigger", "unvalidated_foreign_key", "replica_connection_rejected", "deferred_foreign_key_violation"} {
		t.Run(scenario, func(t *testing.T) {
			ctx := context.Background()
			cfg := createIndependentRecoveryTestConfiguration(t)
			cfg.DatabaseURL, cfg.DatabaseName = connectionURL, "local_backup_normal_order_"+newTaskID()[:12]
			control, err := cfg.connectDatabase(ctx, "postgres")
			require.NoError(t, err)
			_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName))
			require.NoError(t, err)
			t.Cleanup(func() {
				_, err := control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(cfg.DatabaseName)+` WITH (FORCE)`)
				require.NoError(t, err)
				require.NoError(t, control.Close())
			})
			database, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, database.Close()) })
			_, err = database.ExecContext(ctx, `
CREATE TABLE schema_migrations(version bigint PRIMARY KEY,dirty boolean NOT NULL);
INSERT INTO schema_migrations VALUES(436,false);
CREATE TABLE orders(id bigserial PRIMARY KEY,order_number text UNIQUE NOT NULL,parent_order_id bigint REFERENCES orders(id));
CREATE TABLE order_items(id bigserial PRIMARY KEY,order_id bigint NOT NULL REFERENCES orders(id),quantity integer NOT NULL CHECK(quantity>0));
CREATE TABLE users(id bigserial PRIMARY KEY,username text NOT NULL);
INSERT INTO users(id,username) VALUES(7,'current-customer');
INSERT INTO orders(id,order_number) VALUES(7,'preserved-current-order');
`)
			require.NoError(t, err)
			if scenario == "unvalidated_foreign_key" {
				_, err = database.ExecContext(ctx, `ALTER TABLE order_items DROP CONSTRAINT order_items_order_id_fkey; ALTER TABLE order_items ADD CONSTRAINT order_items_order_id_fkey FOREIGN KEY(order_id) REFERENCES orders(id) NOT VALID`)
				require.NoError(t, err)
			}
			if scenario == "disabled_foreign_key_trigger" {
				_, err = database.ExecContext(ctx, `ALTER TABLE order_items DISABLE TRIGGER ALL`)
				require.NoError(t, err)
			}
			if scenario == "deferred_foreign_key_violation" {
				_, err = database.ExecContext(ctx, `ALTER TABLE order_items ALTER CONSTRAINT order_items_order_id_fkey DEFERRABLE INITIALLY DEFERRED`)
				require.NoError(t, err)
			}
			orderRows := []json.RawMessage{json.RawMessage(`{"id":42,"order_number":"restored-child","parent_order_id":43}`), json.RawMessage(`{"id":43,"order_number":"restored-parent","parent_order_id":null}`)}
			itemRows := []json.RawMessage{json.RawMessage(`{"id":42,"order_id":42,"quantity":1}`)}
			if scenario == "missing_foreign_parent" || scenario == "deferred_foreign_key_violation" {
				itemRows[0] = json.RawMessage(`{"id":42,"order_id":999,"quantity":1}`)
			}
			if scenario == "negative_quantity" {
				itemRows[0] = json.RawMessage(`{"id":42,"order_id":42,"quantity":-1}`)
			}
			if scenario == "duplicate_order_identity" {
				orderRows[0] = json.RawMessage(`{"id":42,"order_number":"preserved-current-order","parent_order_id":43}`)
			}
			snapshot := selectedBusinessModuleCompleteSnapshot{Modules: []string{"orders"}, Tables: map[string]businessModuleCompleteTableSnapshot{"orders": {Rows: orderRows}, "order_items": {Rows: itemRows}}}
			missing := map[string][]json.RawMessage{"order_items": itemRows, "orders": orderRows}
			beforeCatalog := readOrderModuleNormalConstraintTestCatalogIdentity(t, database)
			_, beforeFingerprint, _, err := captureDatabaseInventory(ctx, database)
			require.NoError(t, err)
			tx, err := database.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
			require.NoError(t, err)
			defer tx.Rollback()
			if scenario == "replica_connection_rejected" {
				_, err = tx.ExecContext(ctx, `SET LOCAL session_replication_role='replica'`)
				require.NoError(t, err)
			}
			observation, err := inspectModuleRecoveryDatabaseObservation(ctx, tx)
			require.NoError(t, err)
			beforeSequences, err := readBusinessModuleRestoreSequenceStates(ctx, tx)
			require.NoError(t, err)
			restored, writeErr := restoreOrderModuleMissingRowsWithNormalConstraintsTransactionalSequencesAndVerifiedPreservation(ctx, tx, snapshot, observation, missing)
			if scenario == "parent_first_self_reference_and_transactional_rollback" {
				require.NoError(t, writeErr)
				require.Equal(t, 3, restored)
				var role string
				require.NoError(t, tx.QueryRowContext(ctx, `SHOW session_replication_role`).Scan(&role))
				require.Equal(t, "origin", role)
				advanced, err := readBusinessModuleRestoreSequenceStates(ctx, tx)
				require.NoError(t, err)
				require.False(t, reflect.DeepEqual(beforeSequences, advanced))
			} else {
				require.Error(t, writeErr)
				require.Zero(t, restored)
				if strings.Contains(scenario, "foreign_key_trigger") || scenario == "unvalidated_foreign_key" {
					require.Contains(t, writeErr.Error(), "validated foreign keys")
				}
				if scenario == "replica_connection_rejected" {
					require.Contains(t, writeErr.Error(), "origin replication role")
				}
			}
			require.NoError(t, tx.Rollback())
			_, afterFingerprint, _, err := captureDatabaseInventory(ctx, database)
			require.NoError(t, err)
			require.Equal(t, beforeFingerprint, afterFingerprint, "rows and owned sequences must roll back together")
			require.Equal(t, beforeCatalog, readOrderModuleNormalConstraintTestCatalogIdentity(t, database), "constraint and trigger OIDs, definitions and enabled states must remain unchanged")
		})
	}
}
