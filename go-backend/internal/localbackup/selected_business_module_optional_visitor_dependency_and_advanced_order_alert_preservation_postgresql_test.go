package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestSelectedBusinessModuleExternalDependencyReasonsKeepMissingAccountsAndChangedVisitorProfilesBlocked(t *testing.T) {
	hash := strings.Repeat("a", 64)
	reference := selectedBusinessModuleExternalReference{Table: "visitor_profiles", Values: map[string]string{"customer_service_visitor_hash": hash}, RowSHA256: "missing"}
	snapshot := selectedBusinessModuleCompleteSnapshot{Modules: []string{"customer_service"}, Tables: map[string]productCatalogTableSnapshot{"tickets": {Rows: []json.RawMessage{json.RawMessage(`{"visitor_session_hash":"` + hash + `"}`)}}}, ExternalDependencies: []selectedBusinessModuleExternalReference{reference}}
	for _, scenario := range []struct {
		name      string
		change    func(*selectedBusinessModuleCompleteSnapshot, *moduleRecoveryDatabaseObservation, *[]selectedBusinessModuleExternalReference)
		conflicts int
		reviews   int
		reason    string
	}{
		{name: "optional_browsing_context_absent_on_both_sides", reviews: 1, reason: "visitor_browsing_context_absent_at_backup_and_target"},
		{name: "actual_account_missing_on_both_sides", conflicts: 1, reason: "dependency_absent_at_backup_and_target", change: func(snapshot *selectedBusinessModuleCompleteSnapshot, _ *moduleRecoveryDatabaseObservation, actual *[]selectedBusinessModuleExternalReference) {
			snapshot.ExternalDependencies[0].Table = "users"
			(*actual)[0].Table = "users"
		}},
		{name: "profile_lost_after_backup", conflicts: 1, reason: "target_dependency_missing", change: func(snapshot *selectedBusinessModuleCompleteSnapshot, _ *moduleRecoveryDatabaseObservation, _ *[]selectedBusinessModuleExternalReference) {
			snapshot.ExternalDependencies[0].RowSHA256 = strings.Repeat("b", 64)
		}},
		{name: "profile_created_after_backup", conflicts: 1, reason: "package_dependency_missing", change: func(_ *selectedBusinessModuleCompleteSnapshot, _ *moduleRecoveryDatabaseObservation, actual *[]selectedBusinessModuleExternalReference) {
			(*actual)[0].RowSHA256 = strings.Repeat("b", 64)
		}},
		{name: "profile_content_changed", conflicts: 1, reason: "dependency_content_changed", change: func(snapshot *selectedBusinessModuleCompleteSnapshot, _ *moduleRecoveryDatabaseObservation, actual *[]selectedBusinessModuleExternalReference) {
			snapshot.ExternalDependencies[0].RowSHA256 = strings.Repeat("b", 64)
			(*actual)[0].RowSHA256 = strings.Repeat("c", 64)
		}},
		{name: "profile_foreign_key_is_mandatory", conflicts: 1, reason: "dependency_absent_at_backup_and_target", change: func(_ *selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation, _ *[]selectedBusinessModuleExternalReference) {
			observation.References = []ModuleRecoveryDependencyReference{{SourceTable: "tickets", TargetTable: "visitor_profiles"}}
		}},
		{name: "combination_scope_does_not_inherit_chat_exception", conflicts: 1, reason: "dependency_absent_at_backup_and_target", change: func(snapshot *selectedBusinessModuleCompleteSnapshot, _ *moduleRecoveryDatabaseObservation, _ *[]selectedBusinessModuleExternalReference) {
			snapshot.Modules = []string{"customer_service", "orders"}
		}},
		{name: "missing_expected_reference_counts_once", conflicts: 1, reason: "expected_reference_not_reproduced", change: func(_ *selectedBusinessModuleCompleteSnapshot, _ *moduleRecoveryDatabaseObservation, actual *[]selectedBusinessModuleExternalReference) {
			*actual = nil
		}},
		{name: "unexpected_reference_counts_once", conflicts: 1, reason: "unexpected_target_reference", change: func(snapshot *selectedBusinessModuleCompleteSnapshot, _ *moduleRecoveryDatabaseObservation, _ *[]selectedBusinessModuleExternalReference) {
			snapshot.ExternalDependencies = nil
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			encoded, err := json.Marshal(snapshot)
			require.NoError(t, err)
			var candidate selectedBusinessModuleCompleteSnapshot
			require.NoError(t, json.Unmarshal(encoded, &candidate))
			actual := []selectedBusinessModuleExternalReference{reference}
			observation := &moduleRecoveryDatabaseObservation{}
			if scenario.change != nil {
				scenario.change(&candidate, observation, &actual)
			}
			report := &SelectedBusinessModuleIsolatedRestorePreview{}
			classifySelectedBusinessModuleExternalDependencyIssues(candidate, observation, actual, report)
			require.Equal(t, scenario.conflicts, report.DependencyConflicts)
			require.Equal(t, scenario.reviews, report.DependencyReviews)
			require.Len(t, report.DependencyIssues, 1)
			require.Equal(t, scenario.reason, report.DependencyIssues[0].Reason)
			public, err := json.Marshal(report)
			require.NoError(t, err)
			require.NotContains(t, string(public), hash)
		})
	}
	for _, invalid := range []string{"", "a", strings.Repeat("A", 64), strings.Repeat("g", 64), strings.Repeat("0", 63) + "/"} {
		candidate := reference
		candidate.Values = map[string]string{"customer_service_visitor_hash": invalid}
		require.False(t, selectedChatExternalReferenceIsOptionalAbsentVisitorBrowsingContext(snapshot, &moduleRecoveryDatabaseObservation{}, candidate))
	}
}

func TestPostgreSQLSelectedBusinessModuleAdvancedOrderAlertAndOptionalVisitorContextRecoveryBoundaries(t *testing.T) {
	connectionURL, migrationsDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL"), os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || migrationsDirectory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires an explicitly disposable PostgreSQL server and complete application migrations")
	}
	ctx := context.Background()
	cfg := createIndependentRecoveryTestConfiguration(t)
	cfg.DatabaseURL, cfg.DatabaseName = connectionURL, "local_backup_test_"+newTaskID()[:12]
	control, err := cfg.connectDatabase(ctx, "postgres")
	require.NoError(t, err)
	defer control.Close()
	_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName))
	require.NoError(t, err)
	defer func() {
		_, _ = control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(cfg.DatabaseName)+` WITH (FORCE)`)
	}()
	db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	driver, err := migrationpostgres.WithInstance(db, &migrationpostgres.Config{})
	require.NoError(t, err)
	migrator, err := migrate.NewWithDatabaseInstance("file://"+filepath.ToSlash(migrationsDirectory), "postgres", driver)
	require.NoError(t, err)
	err = migrator.Up()
	if err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
	_, err = migrator.Close()
	require.NoError(t, err)
	db, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	defer db.Close()
	baseline, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	observation, err := inspectModuleRecoveryDatabaseObservation(ctx, baseline)
	require.NoError(t, err)
	require.NoError(t, baseline.Rollback())
	readRows := func(tx *sql.Tx, table string) productCatalogTableSnapshot {
		t.Helper()
		contents, err := readSelectedBusinessModuleCompleteTableSnapshot(ctx, tx, table, "true")
		require.NoError(t, err)
		return contents
	}
	for _, scenario := range []struct {
		name, change                string
		allowed, missing, protected bool
	}{
		{name: "same_level_newer_valid_snapshot_preserves_runtime_state", allowed: true},
		{name: "missing_archival_snapshot_restored_without_rewinding_alert_or_replaying_event", allowed: true, missing: true},
		{name: "changed_alert_level_blocks", change: `UPDATE payment_risk_alert_states SET current_level='critical'`},
		{name: "snapshot_from_another_provider_blocks", change: `UPDATE payment_risk_snapshots SET provider='paypal' WHERE id=(SELECT current_snapshot_id FROM payment_risk_alert_states)`},
		{name: "snapshot_level_disagrees_with_state_blocks", change: `UPDATE payment_risk_snapshots SET level='critical' WHERE id=(SELECT current_snapshot_id FROM payment_risk_alert_states)`},
		{name: "snapshot_computed_time_disagrees_with_state_blocks", change: `UPDATE payment_risk_snapshots SET computed_at='2026-01-04Z' WHERE id=(SELECT current_snapshot_id FROM payment_risk_alert_states)`},
		{name: "time_moved_backwards_blocks", change: `UPDATE payment_risk_alert_states SET updated_at='2025-12-31Z'`},
		{name: "missing_snapshot_reference_blocks", change: `UPDATE payment_risk_alert_states SET current_snapshot_id=99999999`},
		{name: "same_snapshot_with_changed_time_blocks", change: `UPDATE payment_risk_alert_states SET current_snapshot_id=(SELECT min(id) FROM payment_risk_snapshots)`},
		{name: "original_snapshot_content_changed_blocks", change: `UPDATE payment_risk_snapshots SET window_days=60 WHERE id=(SELECT min(id) FROM payment_risk_snapshots)`},
		{name: "missing_alert_execution_state_requires_dedicated_recovery", change: `DELETE FROM payment_risk_alert_states`, protected: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			tx, err := db.BeginTx(ctx, nil)
			require.NoError(t, err)
			defer tx.Rollback()
			var originalID, archiveID int64
			require.NoError(t, tx.QueryRowContext(ctx, `INSERT INTO payment_risk_snapshots(provider,window_days,window_start,window_end,computed_at,level) VALUES('stripe',30,'2025-12-01Z','2026-01-01Z','2026-01-01Z','normal') RETURNING id`).Scan(&originalID))
			require.NoError(t, tx.QueryRowContext(ctx, `INSERT INTO payment_risk_snapshots(provider,window_days,window_start,window_end,computed_at,level) VALUES('stripe',30,'2025-11-01Z','2025-12-01Z','2025-12-01Z','normal') RETURNING id`).Scan(&archiveID))
			_, err = tx.ExecContext(ctx, `INSERT INTO payment_risk_alert_states(provider,current_level,current_snapshot_id,updated_at) VALUES('stripe','normal',$1,'2026-01-01Z')`, originalID)
			require.NoError(t, err)
			snapshot := selectedBusinessModuleCompleteSnapshot{Modules: []string{"orders"}, Tables: map[string]productCatalogTableSnapshot{"payment_risk_alert_states": readRows(tx, "payment_risk_alert_states"), "payment_risk_snapshots": readRows(tx, "payment_risk_snapshots")}, ExternalDependencies: []selectedBusinessModuleExternalReference{}}
			_, err = tx.ExecContext(ctx, `WITH next AS (INSERT INTO payment_risk_snapshots(provider,window_days,window_start,window_end,computed_at,level) VALUES('stripe',30,'2025-12-02Z','2026-01-02Z','2026-01-02Z','normal') RETURNING id) UPDATE payment_risk_alert_states SET current_snapshot_id=(SELECT id FROM next),updated_at='2026-01-02Z'`)
			require.NoError(t, err)
			_, err = tx.ExecContext(ctx, `INSERT INTO outbox_events(event_key,event_type,aggregate_type,aggregate_id,payload,status,processed_at) VALUES('preserved-runtime-alert','payment.risk.level_changed','payment_risk','stripe','{}','processed',now())`)
			require.NoError(t, err)
			if scenario.change != "" {
				_, err = tx.ExecContext(ctx, scenario.change)
				require.NoError(t, err)
			}
			if scenario.missing {
				_, err = tx.ExecContext(ctx, `DELETE FROM payment_risk_snapshots WHERE id=$1`, archiveID)
				require.NoError(t, err)
			}
			stateBefore := readRows(tx, "payment_risk_alert_states")
			eventsBefore := readRows(tx, "outbox_events")
			report := &SelectedBusinessModuleIsolatedRestorePreview{}
			missing, blocked, err := inspectSelectedBusinessModuleRowsForRestore(ctx, tx, snapshot, observation, report)
			require.NoError(t, err)
			require.Equal(t, !scenario.allowed, blocked)
			for _, table := range report.Tables {
				if table.Table == "payment_risk_alert_states" {
					if scenario.allowed {
						require.Equal(t, 1, table.PreservedCurrentRuntimeRows)
						require.Zero(t, table.IdenticalRows)
						require.Zero(t, table.ConflictingRows)
					}
					if scenario.protected {
						require.Equal(t, 1, table.ProtectedNewRows)
					}
				}
			}
			if scenario.allowed {
				var scansBefore, scansAfter int64
				require.NoError(t, tx.QueryRowContext(ctx, `SELECT COALESCE(sum(seq_scan),0) FROM pg_stat_xact_user_tables`).Scan(&scansBefore))
				count, err := restoreSelectedBusinessModuleMissingRowsAndVerifyPreservation(ctx, tx, snapshot, observation, missing)
				require.NoError(t, err)
				require.NoError(t, tx.QueryRowContext(ctx, `SELECT COALESCE(sum(seq_scan),0) FROM pg_stat_xact_user_tables`).Scan(&scansAfter))
				if scenario.missing {
					require.Equal(t, 1, count)
				} else {
					require.Zero(t, count)
					require.Equal(t, scansBefore, scansAfter, "no-op recovery must not scan unrelated rows")
				}
				require.Equal(t, stateBefore, readRows(tx, "payment_risk_alert_states"))
				require.Equal(t, eventsBefore, readRows(tx, "outbox_events"))
			}
		})
	}
	t.Run("chat_missing_optional_browsing_profile_preserves_signed_session_and_avoids_unrelated_content_scan", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)
		defer tx.Rollback()
		_, err = tx.ExecContext(ctx, `WITH agent AS (INSERT INTO users(email,username,password,role) VALUES('visitor-context-agent@example.invalid','visitor-context-agent','synthetic-password','admin') RETURNING id) INSERT INTO tickets(ticket_number,user_id,conversation_id,visitor_session_hash,subject,category) SELECT 'visitor-context-review',id,'visitor-context-review',repeat('a',64),'Synthetic retained visitor conversation','customer_service' FROM agent`)
		require.NoError(t, err)
		snapshot, err := cfg.captureSelectedBusinessModuleCompleteSnapshot(ctx, tx, []string{"customer_service"})
		require.NoError(t, err)
		var scansBefore, scansAfter int64
		require.NoError(t, tx.QueryRowContext(ctx, `SELECT seq_scan FROM pg_stat_xact_user_tables WHERE relname='site_quality_runs'`).Scan(&scansBefore))
		snapshot.ExternalDependencies, err = captureSelectedBusinessModuleExternalReferenceFingerprints(ctx, tx, snapshot)
		require.NoError(t, err)
		require.NoError(t, tx.QueryRowContext(ctx, `SELECT seq_scan FROM pg_stat_xact_user_tables WHERE relname='site_quality_runs'`).Scan(&scansAfter))
		require.Equal(t, scansBefore, scansAfter, "reference capture must only inspect catalog and actual referenced rows")
		report := &SelectedBusinessModuleIsolatedRestorePreview{}
		missing, blocked, err := inspectSelectedBusinessModuleRowsForRestore(ctx, tx, snapshot, observation, report)
		require.NoError(t, err)
		require.False(t, blocked)
		require.Empty(t, missing)
		require.Zero(t, report.DependencyConflicts)
		require.Equal(t, 1, report.DependencyReviews)
		strategies, err := cfg.inspectSelectedCustomerServiceModuleBoundedOverwriteStrategies(ctx, tx, snapshot, observation, missing, &Manifest{}, t.TempDir(), report)
		require.NoError(t, err)
		require.True(t, strategies.PrerequisitesSatisfied)
		require.True(t, strategies.PackageRows.DomainChecksSatisfied, "%v", strategies.PackageRows.IssueCounts)
		require.True(t, strategies.PreserveCurrentRows.DomainChecksSatisfied, "%v", strategies.PreserveCurrentRows.IssueCounts)
		graphReport := newSelectedCustomerServiceModuleDomainChecks()
		graph := newModuleRecoveryRowDependencyGraph(tx, observation, "target", graphReport, defaultModuleRecoveryRowInspectionLimits())
		require.NoError(t, graph.selectCustomerServiceConversationScopeRows(ctx))
		require.NoError(t, graph.inspectConversationBusinessReferences(ctx))
		require.Equal(t, 1, graphReport.IssueCounts["visitor_context_profile_missing"])
		require.Equal(t, snapshot.Tables["tickets"], readRows(tx, "tickets"))
		var profiles int
		require.NoError(t, tx.QueryRowContext(ctx, `SELECT count(*) FROM visitor_profiles`).Scan(&profiles))
		require.Zero(t, profiles, "recovery must never manufacture a browsing profile")
	})
}
