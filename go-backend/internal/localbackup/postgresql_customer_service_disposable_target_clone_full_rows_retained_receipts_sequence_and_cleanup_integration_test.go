package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type customerServiceDisposableCandidateTestFixture struct {
	Configuration                         Configuration
	Control, Source, Target               *sql.DB
	SourceName, TaskID, VerifiedDirectory string
	Manifest                              *Manifest
}

func createCustomerServiceDisposableCandidateRehearsalTestFixture(t *testing.T) *customerServiceDisposableCandidateTestFixture {
	t.Helper()
	if os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" || os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL") == "" {
		t.Skip("requires a newly created disposable PostgreSQL server with pg_dump and pg_restore")
	}
	ctx := context.Background()
	cfg := createIndependentRecoveryTestConfiguration(t)
	cfg.DatabaseURL = os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	if capacityDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_POSTGRESQL_CAPACITY_DIRECTORY"); capacityDirectory != "" {
		cfg.PostgreSQLCapacityDirectory = capacityDirectory
	}
	cfg.DatabaseName = "local_backup_test_" + newTaskID()[:12]
	cfg.ApplicationDatabaseRole = "synthetic_candidate_app_" + newTaskID()[:12]
	control, err := cfg.connectDatabase(ctx, "postgres")
	require.NoError(t, err)
	taskID := newTaskID()
	fixture := &customerServiceDisposableCandidateTestFixture{Configuration: cfg, Control: control, TaskID: taskID, SourceName: "local_backup_stage_" + taskID[:16], VerifiedDirectory: filepath.Join(t.TempDir(), "verified")}
	_, err = control.ExecContext(ctx, `CREATE ROLE `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole)+` NOLOGIN`)
	require.NoError(t, err)
	t.Cleanup(func() {
		for _, name := range []string{fixture.SourceName, cfg.DatabaseName} {
			_, _ = control.ExecContext(context.Background(), `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(name)+` WITH (FORCE)`)
		}
		_, _ = control.ExecContext(context.Background(), `DROP ROLE `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole))
		control.Close()
	})
	for index, name := range []string{fixture.SourceName, cfg.DatabaseName} {
		_, err := control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(name))
		require.NoError(t, err)
		database, err := cfg.connectDatabase(ctx, name)
		require.NoError(t, err)
		t.Cleanup(func() { database.Close() })
		_, err = database.ExecContext(ctx, customerServiceRowInspectionSyntheticSchema+`
ALTER TABLE ticket_messages ADD COLUMN is_internal boolean DEFAULT false;
CREATE SEQUENCE tickets_id_seq OWNED BY tickets.id;
ALTER TABLE tickets ALTER COLUMN id SET DEFAULT nextval('tickets_id_seq');
CREATE SEQUENCE ticket_messages_id_seq OWNED BY ticket_messages.id;
ALTER TABLE ticket_messages ALTER COLUMN id SET DEFAULT nextval('ticket_messages_id_seq');
CREATE SEQUENCE orders_id_seq OWNED BY orders.id;
ALTER TABLE orders ALTER COLUMN id SET DEFAULT nextval('orders_id_seq');
CREATE SEQUENCE synthetic_unowned_sequence;
`+customerServiceRowInspectionSyntheticRows+`
SELECT setval('tickets_id_seq',20,false),setval('ticket_messages_id_seq',10,false),setval('orders_id_seq',2000,true),setval('synthetic_unowned_sequence',700,false);
UPDATE ticket_messages SET is_internal=true WHERE id=1;
`)
		require.NoError(t, err)
		if index == 0 {
			fixture.Source = database
		} else {
			fixture.Target = database
		}
	}
	fixture.Manifest = &Manifest{FormatVersion: PackageVersion, PackageID: newTaskID(), Scope: "whole_application", SiteID: cfg.SiteID, Environment: cfg.Environment, ApplicationVersion: cfg.ApplicationVersion, MigrationVersion: cfg.ExpectedMigrationVersion}
	for _, path := range []string{"media/uploads_private/customer-service/attachments/originals/synthetic.txt", "media/uploads_private/customer-service/attachments/derivatives/synthetic.txt", "media/uploads/legacy-attachment.txt"} {
		parts := strings.SplitN(strings.TrimPrefix(path, "media/"), "/", 2)
		for _, filename := range []string{filepath.Join(fixture.VerifiedDirectory, filepath.FromSlash(path)), filepath.Join(cfg.StorageRoots[parts[0]], filepath.FromSlash(parts[1]))} {
			require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
			require.NoError(t, os.WriteFile(filename, []byte("original"), 0600))
		}
		identity, err := fileIdentity(filepath.Join(fixture.VerifiedDirectory, filepath.FromSlash(path)), path)
		require.NoError(t, err)
		fixture.Manifest.Files = append(fixture.Manifest.Files, identity)
	}
	return fixture
}

func (fixture *customerServiceDisposableCandidateTestFixture) inspectAndBindSavedCandidateRehearsalSelections(t *testing.T, choices map[string]string) (*moduleRecoveryRowDependencyGraph, *moduleRecoveryRowDependencyGraph, *CustomerServiceConversationRecoveryPlan, *CustomerServiceRecoveryConflictReview, CustomerServiceEffectiveSelectionInspectionRequest) {
	t.Helper()
	ctx := context.Background()
	cfg := fixture.Configuration
	var transactions []*sql.Tx
	var observations []*moduleRecoveryDatabaseObservation
	for _, database := range []*sql.DB{fixture.Source, fixture.Target} {
		transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
		require.NoError(t, err)
		t.Cleanup(func() { transaction.Rollback() })
		observation, err := inspectModuleRecoveryDatabaseObservation(ctx, transaction)
		require.NoError(t, err)
		transactions = append(transactions, transaction)
		observations = append(observations, observation)
	}
	var source, target *moduleRecoveryRowDependencyGraph
	inspection, err := cfg.inspectCustomerServiceConversationRows(ctx, transactions[0], transactions[1], observations[0], observations[1], fixture.Manifest, fixture.VerifiedDirectory, defaultModuleRecoveryRowInspectionLimits(), func(packageGraph, currentGraph *moduleRecoveryRowDependencyGraph) {
		source, target = packageGraph, currentGraph
	})
	require.NoError(t, err)
	plan := inspection.ConversationRecoveryPlan
	fixture.Manifest.DatabaseFingerprint = observations[0].Fingerprint
	task := &Task{ID: fixture.TaskID, StageDatabase: fixture.SourceName, Manifest: fixture.Manifest, ExpiresAt: time.Now().UTC().Add(time.Hour)}
	report := &ModuleRecoveryDependencyReport{SourceDatabaseFingerprint: observations[0].Fingerprint, TargetDatabaseFingerprint: observations[1].Fingerprint, Blockers: []ModuleRecoveryDependencyBlocker{{Code: "module_recovery_not_enabled"}, {Code: "row_and_file_ownership_review_required"}}}
	mediaFingerprint, err := cfg.mediaFingerprint(ctx)
	require.NoError(t, err)
	require.NoError(t, cfg.bindCustomerServiceRecoveryPlanToObservedSnapshots(plan, task, report, mediaFingerprint, observations[0], observations[1]))
	reviewRequest := CustomerServiceRecoveryConflictReviewRequest{PlanSHA256: plan.SHA256, Confirmation: "REVIEW " + cfg.SiteID + "/" + cfg.Environment + "/" + task.Manifest.PackageID}
	for _, row := range plan.RowActions {
		if !customerServiceRecoveryRowRequiresContentDecision(row) {
			continue
		}
		choice := choices[row.Table+":"+row.RowIdentity]
		if choice == "" {
			choice = "use_package_row"
		}
		reviewRequest.Decisions = append(reviewRequest.Decisions, CustomerServiceRecoveryContentDecision{Table: row.Table, IdentitySHA256: row.IdentitySHA256, Choice: choice, Reason: "synthetic disposable clone verification"})
	}
	var review *CustomerServiceRecoveryConflictReview
	request := CustomerServiceEffectiveSelectionInspectionRequest{PlanSHA256: plan.SHA256}
	for start := 0; start < len(reviewRequest.Decisions); start += customerServiceConflictReviewMaximumDecisionsPerSave {
		chunk := reviewRequest
		chunk.Decisions = reviewRequest.Decisions[start:min(start+customerServiceConflictReviewMaximumDecisionsPerSave, len(reviewRequest.Decisions))]
		if review != nil {
			chunk.ExpectedReviewSHA256 = review.SHA256
		}
		review, err = buildCustomerServiceRecoveryConflictReview(plan, review, chunk, "synthetic-operator", time.Now().UTC())
		require.NoError(t, err)
	}
	if review != nil {
		request.ReviewSHA256 = review.SHA256
	}
	return source, target, plan, review, request
}

func TestPostgreSQLCustomerServiceDisposableTargetCloneCopiesFullRowsPreservesReceiptsAndAdvancesOnlySelectedSequences(t *testing.T) {
	for _, scenario := range []struct {
		name, bothSQL, sourceSQL, targetSQL string
		choices                             map[string]string
		selectedRows, adjustments           int
	}{
		{name: "unchanged_target_clone_retains_everything"},
		{name: "replace_complete_message_and_insert_historical_id", sourceSQL: `INSERT INTO ticket_messages(id,ticket_id,user_id,content,attachments,attachment_reference_format,is_internal) VALUES(100,1,1,'private historical inserted message','[]','legacy_urls',true);`, targetSQL: `UPDATE ticket_messages SET content='private newer current message',is_internal=false WHERE id=1; INSERT INTO ticket_messages(id,ticket_id,content,attachments,attachment_reference_format) VALUES(7,1,'private current-only message','[]','legacy_urls'); UPDATE outbox_events SET status='ignored',attempts=7 WHERE id=1; INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,status) VALUES(77,'preserved-registration-reward','referral.registered','user','1','processed');`, selectedRows: 2, adjustments: 1},
		{name: "preserve_current_content_keeps_its_full_fingerprint", targetSQL: `UPDATE ticket_messages SET content='private chosen current message' WHERE id=1;`, choices: map[string]string{"ticket_messages:1": "preserve_current_row"}},
		{name: "source_only_conversation_parent_before_children", sourceSQL: `INSERT INTO tickets(id,category,conversation_id,ticket_number) VALUES(100,'customer_service','synthetic-source-only','synthetic-source-only-number'); INSERT INTO ticket_messages(id,ticket_id,content,attachments,attachment_reference_format) VALUES(101,100,'private source-only conversation body','[]','legacy_urls');`, selectedRows: 2, adjustments: 2},
		{name: "advanced_current_sequence_is_never_rewound", sourceSQL: `INSERT INTO ticket_messages(id,ticket_id,content,attachments,attachment_reference_format) VALUES(100,1,'private restored lower id','[]','legacy_urls');`, targetSQL: `SELECT setval('ticket_messages_id_seq',500,true);`, selectedRows: 1},
		{name: "descending_sequence_advances_in_its_own_direction", bothSQL: `ALTER SEQUENCE ticket_messages_id_seq INCREMENT BY -1 MINVALUE -1000 MAXVALUE 1000; SELECT setval('ticket_messages_id_seq',-10,false);`, sourceSQL: `INSERT INTO ticket_messages(id,ticket_id,content,attachments,attachment_reference_format) VALUES(-100,1,'private restored negative id','[]','legacy_urls');`, selectedRows: 1, adjustments: 1},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
			ctx := context.Background()
			if scenario.bothSQL != "" {
				for _, database := range []*sql.DB{fixture.Source, fixture.Target} {
					_, err := database.ExecContext(ctx, scenario.bothSQL)
					require.NoError(t, err)
				}
			}
			for index, statement := range []string{scenario.sourceSQL, scenario.targetSQL} {
				if statement != "" {
					database := fixture.Source
					if index == 1 {
						database = fixture.Target
					}
					_, err := database.ExecContext(ctx, statement)
					require.NoError(t, err)
				}
			}
			source, target, plan, review, request := fixture.inspectAndBindSavedCandidateRehearsalSelections(t, scenario.choices)
			result, err := fixture.Configuration.rehearseCustomerServiceSelectedRowsInDisposableTargetClone(ctx, source, target, plan, review, request, fixture.Manifest, fixture.VerifiedDirectory)
			require.NoError(t, err)
			require.Equal(t, "rehearsed_not_switchable", result.State)
			require.False(t, result.ExecutionAvailable)
			require.True(t, result.CandidateDeleted)
			require.Equal(t, scenario.selectedRows, result.SelectedFullRows)
			require.Len(t, result.SequenceAdjustments, scenario.adjustments)
			require.Equal(t, len(target.Observation.Tables), result.VerifiedRetainedTables)
			require.Equal(t, 4, result.VerifiedSequenceStates)
			require.Len(t, result.DependencyBlockers, 2)
			if scenario.selectedRows == 0 {
				require.Equal(t, result.BaselineFingerprint, result.CandidateFingerprint)
			} else {
				require.NotEqual(t, result.BaselineFingerprint, result.CandidateFingerprint)
			}
			var count int
			require.NoError(t, fixture.Control.QueryRowContext(ctx, `SELECT count(*) FROM pg_database WHERE datname LIKE 'local_backup_candidate_%'`).Scan(&count))
			require.Zero(t, count)
			entries, err := os.ReadDir(fixture.Configuration.Workspace)
			require.NoError(t, err)
			require.Empty(t, entries)
			encoded, err := json.Marshal(result)
			require.NoError(t, err)
			require.NotContains(t, string(encoded), "private ")
			require.NotContains(t, string(encoded), fixture.SourceName)
			t.Logf("selected complete rows=%d, retained tables=%d, checked sequence states=%d, adjustments=%d, candidate deleted", result.SelectedFullRows, result.VerifiedRetainedTables, result.VerifiedSequenceStates, len(result.SequenceAdjustments))
		})
	}
}

func TestPostgreSQLCustomerServiceDisposableCandidateRejectsUnsafeSelectionsBindingsCapacityAndWriteContracts(t *testing.T) {
	for _, scenario := range []struct {
		name, bothSQL, sourceSQL, targetSQL, expectedError string
		mutate                                             func(*customerServiceDisposableCandidateTestFixture, *CustomerServiceConversationRecoveryPlan, *CustomerServiceRecoveryConflictReview, *CustomerServiceEffectiveSelectionInspectionRequest)
	}{
		{name: "oversized_full_body_is_not_a_small_projection", sourceSQL: `UPDATE ticket_messages SET content=repeat('private body ',6000) WHERE id=1;`, expectedError: "row size limit"},
		{name: "complete_row_total_byte_budget_before_allocation", sourceSQL: `INSERT INTO ticket_messages(id,ticket_id,content,attachments,attachment_reference_format) SELECT 1000+generated_id,1,repeat('a',60000),'[]','legacy_urls' FROM generate_series(1,300) generated_id;`, expectedError: "resource budget"},
		{name: "unvalidated_constraint_rejected_before_allocation", bothSQL: `ALTER TABLE ticket_messages ADD CHECK(content<>'private rejected complete body') NOT VALID;`, expectedError: "unvalidated constraints"},
		{name: "disabled_foreign_key_triggers_rejected_before_allocation", bothSQL: `ALTER TABLE ticket_messages DISABLE TRIGGER ALL;`, expectedError: "disabled internal constraint triggers"},
		{name: "selected_sequence_exhaustion_discards_candidate", bothSQL: `ALTER SEQUENCE ticket_messages_id_seq MAXVALUE 100;`, sourceSQL: `INSERT INTO ticket_messages(id,ticket_id,content,attachments,attachment_reference_format) VALUES(100,1,'private sequence boundary','[]','legacy_urls');`, expectedError: "exhausted"},
		{name: "unsafe_current_sequence_is_not_silently_repaired", targetSQL: `SELECT setval('ticket_messages_id_seq',3,true); UPDATE ticket_messages SET content='private changed current content' WHERE id=1;`, expectedError: "baseline has unsafe"},
		{name: "user_trigger_is_never_executed", bothSQL: `CREATE FUNCTION synthetic_candidate_rejected_trigger() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'private trigger must not run'; END $$; CREATE TRIGGER synthetic_candidate_rejected_trigger BEFORE INSERT ON ticket_messages FOR EACH ROW EXECUTE FUNCTION synthetic_candidate_rejected_trigger();`, expectedError: "repair blockers"},
		{name: "generated_columns_need_an_explicit_contract", bothSQL: `ALTER TABLE ticket_messages ADD COLUMN is_read boolean GENERATED ALWAYS AS (length(content)>0) STORED;`, expectedError: "unsupported type"},
		{name: "row_security_needs_an_explicit_contract", bothSQL: `ALTER TABLE ticket_messages ENABLE ROW LEVEL SECURITY;`, expectedError: "unsupported type"},
		{name: "user_constraint_function_needs_an_explicit_contract", bothSQL: `CREATE FUNCTION synthetic_candidate_constraint_function(text) RETURNS boolean LANGUAGE plpgsql IMMUTABLE AS $$ BEGIN RETURN true; END $$; ALTER TABLE ticket_messages ADD CHECK (synthetic_candidate_constraint_function(content));`, expectedError: "write expression"},
		{name: "stale_target_binding", expectedError: "snapshot binding", mutate: func(f *customerServiceDisposableCandidateTestFixture, p *CustomerServiceConversationRecoveryPlan, r *CustomerServiceRecoveryConflictReview, q *CustomerServiceEffectiveSelectionInspectionRequest) {
			p.Binding.TargetDatabaseFingerprint = strings.Repeat("a", 64)
			p.SHA256 = ""
			p.SHA256, _ = customerServiceRecoveryPlanCanonicalJSONSHA256(p)
			q.PlanSHA256 = p.SHA256
		}},
		{name: "workspace_capacity_before_database_allocation", expectedError: "workspace capacity", mutate: func(f *customerServiceDisposableCandidateTestFixture, p *CustomerServiceConversationRecoveryPlan, r *CustomerServiceRecoveryConflictReview, q *CustomerServiceEffectiveSelectionInspectionRequest) {
			f.Configuration.MaxWorkspaceBytes = 1
		}},
		{name: "superuser_application_role_is_rejected", expectedError: "snapshot binding", mutate: func(f *customerServiceDisposableCandidateTestFixture, p *CustomerServiceConversationRecoveryPlan, r *CustomerServiceRecoveryConflictReview, q *CustomerServiceEffectiveSelectionInspectionRequest) {
			f.Configuration.ApplicationDatabaseRole = "postgres"
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
			ctx := context.Background()
			if scenario.bothSQL != "" {
				for _, database := range []*sql.DB{fixture.Source, fixture.Target} {
					_, err := database.ExecContext(ctx, scenario.bothSQL)
					require.NoError(t, err)
				}
			}
			for index, statement := range []string{scenario.sourceSQL, scenario.targetSQL} {
				if statement != "" {
					database := fixture.Source
					if index == 1 {
						database = fixture.Target
					}
					_, err := database.ExecContext(ctx, statement)
					require.NoError(t, err)
				}
			}
			source, target, plan, review, request := fixture.inspectAndBindSavedCandidateRehearsalSelections(t, nil)
			if scenario.mutate != nil {
				scenario.mutate(fixture, plan, review, &request)
			}
			result, err := fixture.Configuration.rehearseCustomerServiceSelectedRowsInDisposableTargetClone(ctx, source, target, plan, review, request, fixture.Manifest, fixture.VerifiedDirectory)
			require.ErrorContains(t, err, scenario.expectedError)
			require.Nil(t, result)
			_, sourceFingerprint, _, err := captureDatabaseInventory(ctx, fixture.Source)
			require.NoError(t, err)
			require.Equal(t, source.Observation.Fingerprint, sourceFingerprint)
			_, targetFingerprint, _, err := captureDatabaseInventory(ctx, fixture.Target)
			require.NoError(t, err)
			require.Equal(t, target.Observation.Fingerprint, targetFingerprint)
			var count int
			require.NoError(t, fixture.Control.QueryRowContext(ctx, `SELECT count(*) FROM pg_database WHERE datname LIKE 'local_backup_candidate_%'`).Scan(&count))
			require.Zero(t, count)
		})
	}
}

func TestPostgreSQLCustomerServiceDisposableCandidateCleanupRequiresTheRecordedOIDNameAndUnprivilegedOwner(t *testing.T) {
	fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
	ctx := context.Background()
	require.NoError(t, ensureCustomerServiceDisposableCandidateOwnerAndApplicationIsolation(ctx, fixture.Control, fixture.Configuration.ApplicationDatabaseRole))
	require.ErrorContains(t, ensureCustomerServiceDisposableCandidateOwnerAndApplicationIsolation(ctx, fixture.Control, "postgres"), "unprivileged application")
	name := "local_backup_candidate_" + newTaskID()[:16]
	_, err := fixture.Control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(name)+` WITH OWNER `+pq.QuoteIdentifier(customerServiceDisposableCandidateOwner))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = fixture.Control.ExecContext(context.Background(), `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(name)+` WITH (FORCE)`)
	})
	identity := customerServiceDisposableCandidateDatabaseIdentity{Version: 1, DatabaseName: name}
	require.NoError(t, fixture.Control.QueryRowContext(ctx, `SELECT oid::bigint FROM pg_database WHERE datname=$1`, name).Scan(&identity.DatabaseOID))
	t.Run("missing_persisted_oid_never_guesses", func(t *testing.T) {
		invalid := identity
		invalid.DatabaseOID = 0
		require.ErrorContains(t, cleanupCustomerServiceDisposableCandidateDatabaseByOID(ctx, fixture.Control, invalid), "persisted")
	})
	t.Run("recorded_oid_cannot_drop_another_database", func(t *testing.T) {
		invalid := identity
		require.NoError(t, fixture.Control.QueryRowContext(ctx, `SELECT oid::bigint FROM pg_database WHERE datname=$1`, fixture.Configuration.DatabaseName).Scan(&invalid.DatabaseOID))
		require.ErrorContains(t, cleanupCustomerServiceDisposableCandidateDatabaseByOID(ctx, fixture.Control, invalid), "changed")
	})
	t.Run("changed_owner_requires_inspection", func(t *testing.T) {
		_, err := fixture.Control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(name)+` OWNER TO postgres`)
		require.NoError(t, err)
		require.ErrorContains(t, cleanupCustomerServiceDisposableCandidateDatabaseByOID(ctx, fixture.Control, identity), "owner changed")
		_, err = fixture.Control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(name)+` OWNER TO `+pq.QuoteIdentifier(customerServiceDisposableCandidateOwner))
		require.NoError(t, err)
	})
	require.NoError(t, cleanupCustomerServiceDisposableCandidateDatabaseByOID(ctx, fixture.Control, identity))
	require.NoError(t, cleanupCustomerServiceDisposableCandidateDatabaseByOID(ctx, fixture.Control, identity))
	t.Run("reused_name_does_not_authorize_dropping_a_new_oid", func(t *testing.T) {
		_, err := fixture.Control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(name)+` WITH OWNER `+pq.QuoteIdentifier(customerServiceDisposableCandidateOwner))
		require.NoError(t, err)
		require.NoError(t, cleanupCustomerServiceDisposableCandidateDatabaseByOID(ctx, fixture.Control, identity))
		var replacementOID int64
		require.NoError(t, fixture.Control.QueryRowContext(ctx, `SELECT oid::bigint FROM pg_database WHERE datname=$1`, name).Scan(&replacementOID))
		require.NotEqual(t, identity.DatabaseOID, replacementOID)
	})
}

func TestPostgreSQLCustomerServiceDisposableCandidateSnapshotDumpCannotExceedItsWorkspaceByteLimit(t *testing.T) {
	fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
	ctx := context.Background()
	_, target, _, _, _ := fixture.inspectAndBindSavedCandidateRehearsalSelections(t, nil)
	var snapshot string
	require.NoError(t, target.Transaction.QueryRowContext(ctx, `SELECT pg_export_snapshot()`).Scan(&snapshot))
	filename := filepath.Join(t.TempDir(), "bounded-disposable-customer-service-candidate-snapshot.dump")
	require.ErrorContains(t, fixture.Configuration.dumpCustomerServiceTargetSnapshotWithAnExplicitByteLimit(ctx, fixture.Configuration.DatabaseName, snapshot, filename, 1), "exceeded its byte limit")
	info, err := os.Stat(filename)
	require.NoError(t, err)
	require.LessOrEqual(t, info.Size(), int64(1))
	_, fingerprint, _, err := captureDatabaseInventory(ctx, fixture.Target)
	require.NoError(t, err)
	require.Equal(t, target.Observation.Fingerprint, fingerprint)
}

func verifyCustomerServiceDisposableCandidateRehearsalWithCompleteMigrationFixture(t *testing.T, cfg Configuration, currentDatabase string, files []FileIdentity, verifiedDirectory string) {
	t.Helper()
	t.Run("complete_migration_disposable_customer_service_candidate_retains_all_current_tables", func(t *testing.T) {
		ctx := context.Background()
		cfg.DatabaseName = currentDatabase
		cfg.ApplicationDatabaseRole = "synthetic_candidate_app_" + newTaskID()[:12]
		control, err := cfg.connectDatabase(ctx, "postgres")
		require.NoError(t, err)
		defer control.Close()
		_, err = control.ExecContext(ctx, `CREATE ROLE `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole)+` NOLOGIN`)
		require.NoError(t, err)
		defer func() { _, _ = control.ExecContext(ctx, `DROP ROLE `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole)) }()
		id := newTaskID()
		sourceName := "local_backup_stage_" + id[:16]
		defer func() {
			_, _ = control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(sourceName)+` WITH (FORCE)`)
		}()
		directory := t.TempDir()
		manifest := &Manifest{FormatVersion: PackageVersion, PackageID: newTaskID(), Scope: "whole_application", SiteID: cfg.SiteID, Environment: cfg.Environment, ApplicationVersion: cfg.ApplicationVersion, Files: files}
		require.NoError(t, cfg.captureSnapshotAndDump(ctx, directory, manifest))
		require.NoError(t, cfg.restoreIsolatedDatabase(ctx, directory, sourceName, *manifest))
		sourceDatabase, err := cfg.connectDatabase(ctx, sourceName)
		require.NoError(t, err)
		defer sourceDatabase.Close()
		targetDatabase, err := cfg.connectDatabase(ctx, currentDatabase)
		require.NoError(t, err)
		defer targetDatabase.Close()
		_, err = sourceDatabase.ExecContext(ctx, `UPDATE ticket_messages SET content='private full-migration historical content'; INSERT INTO ticket_messages(id,ticket_id,user_id,content,attachments,attachment_reference_format) SELECT 10000,ticket_id,user_id,'private full-migration historical inserted message','[]','legacy_urls' FROM ticket_messages ORDER BY id LIMIT 1;`)
		require.NoError(t, err)
		fixture := &customerServiceDisposableCandidateTestFixture{Configuration: cfg, Control: control, Source: sourceDatabase, Target: targetDatabase, SourceName: sourceName, TaskID: id, VerifiedDirectory: verifiedDirectory, Manifest: manifest}
		source, target, plan, review, request := fixture.inspectAndBindSavedCandidateRehearsalSelections(t, nil)
		result, err := cfg.rehearseCustomerServiceSelectedRowsInDisposableTargetClone(ctx, source, target, plan, review, request, manifest, verifiedDirectory)
		require.NoError(t, err)
		require.True(t, result.CandidateDeleted)
		require.False(t, result.ExecutionAvailable)
		require.Equal(t, 2, result.SelectedFullRows)
		require.Equal(t, len(registeredStorefrontModuleRecoveryTableNames()), result.VerifiedRetainedTables)
		require.Len(t, result.SequenceAdjustments, 1)
		require.Equal(t, "ticket_messages_id_seq", result.SequenceAdjustments[0].Sequence)
		require.EqualValues(t, 10000, result.SequenceAdjustments[0].CandidateValue)
		t.Logf("complete migration candidate: %d full rows, %d retained tables, %d sequence states; payment/refund/mail receipts and registration data retained; temporary candidate deleted", result.SelectedFullRows, result.VerifiedRetainedTables, result.VerifiedSequenceStates)
		if validationDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_VALIDATION_DIRECTORY"); validationDirectory != "" {
			require.NoError(t, os.MkdirAll(validationDirectory, 0700))
			require.NoError(t, writeDurableJSON(filepath.Join(validationDirectory, "customer-service-complete-migration-disposable-candidate-rehearsal.json"), result))
		}
		verifyCustomerServicePersistentCandidateWithCompleteMigrationFixture(t, fixture)
	})
}

func TestPostgreSQLCustomerServiceDisposableCandidateRejectsCurrentChangesAfterObservationAndDiscardsItsClone(t *testing.T) {
	for _, scenario := range []struct{ name, statement, expectedError string }{
		{"retained_event_state_changes", `UPDATE outbox_events SET status='ignored',attempts=9 WHERE id=1`, "changed during candidate rehearsal"},
		{"sequence_changes_outside_mvcc", `SELECT nextval('orders_id_seq')`, "differs from the observed"},
		{"schema_changes_after_snapshot", `CREATE INDEX synthetic_candidate_late_index ON tickets(category)`, "changed during candidate rehearsal"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
			ctx := context.Background()
			source, target, plan, review, request := fixture.inspectAndBindSavedCandidateRehearsalSelections(t, nil)
			_, err := fixture.Target.ExecContext(ctx, scenario.statement)
			require.NoError(t, err)
			result, err := fixture.Configuration.rehearseCustomerServiceSelectedRowsInDisposableTargetClone(ctx, source, target, plan, review, request, fixture.Manifest, fixture.VerifiedDirectory)
			require.ErrorContains(t, err, scenario.expectedError)
			require.Nil(t, result)
			var count int
			require.NoError(t, fixture.Control.QueryRowContext(ctx, `SELECT count(*) FROM pg_database WHERE datname LIKE 'local_backup_candidate_%'`).Scan(&count))
			require.Zero(t, count)
		})
	}
}

func TestPostgreSQLCustomerServiceDisposableCandidateFullRowWriterEnforcesConstraintsDigestsAndRetainedFingerprints(t *testing.T) {
	for _, scenario := range []struct{ name, bothSQL, mutation, collateralSQL, expectedError string }{
		{"foreign_key_violation", "", `jsonb_set(to_jsonb(record),'{ticket_id}','999'::jsonb)`, "", "violates a database constraint"},
		{"check_violation", `ALTER TABLE ticket_messages ADD CHECK(content<>'private forbidden content');`, `jsonb_set(to_jsonb(record),'{content}','"private forbidden content"'::jsonb)`, "", "violates a database constraint"},
		{"unique_violation", `CREATE UNIQUE INDEX synthetic_candidate_content_unique ON ticket_messages(content);`, `jsonb_set(to_jsonb(record),'{content}','"private visitor message"'::jsonb)`, "", "violates a database constraint"},
		{"incorrect_selected_digest", "", `to_jsonb(record)`, "", "does not match the selected source digest"},
		{"collateral_retained_event_change", "", `to_jsonb(record)`, `UPDATE outbox_events SET attempts=11 WHERE id=1`, "changed retained target rows"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
			ctx := context.Background()
			if scenario.bothSQL != "" {
				for _, database := range []*sql.DB{fixture.Source, fixture.Target} {
					_, err := database.ExecContext(ctx, scenario.bothSQL)
					require.NoError(t, err)
				}
			}
			source, target, plan, review, request := fixture.inspectAndBindSavedCandidateRehearsalSelections(t, nil)
			selection, err := fixture.Configuration.inspectCustomerServiceEffectiveSelectedRows(ctx, source, target, plan, review, request, fixture.Manifest, fixture.VerifiedDirectory)
			require.NoError(t, err)
			if scenario.name == "unique_violation" {
				require.False(t, selection.SelectionDependenciesSatisfied, "unregistered content uniqueness stays blocked by the outer inspector")
			} else {
				require.True(t, selection.SelectionDependenciesSatisfied)
			}
			transaction, err := fixture.Target.BeginTx(ctx, nil)
			require.NoError(t, err)
			defer transaction.Rollback()
			// Exercise the write kernel inside a rollback-only disposable fixture.
			var content, digest string
			require.NoError(t, transaction.QueryRowContext(ctx, `SELECT (`+scenario.mutation+`)::text,encode(sha256(convert_to(to_jsonb(record)::text,'UTF8')),'hex') FROM ticket_messages record WHERE id=1`).Scan(&content, &digest))
			if scenario.name == "incorrect_selected_digest" {
				digest = strings.Repeat("a", 64)
			}
			if scenario.collateralSQL != "" {
				_, err := transaction.ExecContext(ctx, scenario.collateralSQL)
				require.NoError(t, err)
			}
			row := customerServiceCandidateSelectedFullRow{Table: "ticket_messages", Identity: `{"id":1}`, Digest: digest, Content: json.RawMessage(content)}
			err = writeCustomerServiceSelectedFullRowsAndValidateRetainedTables(ctx, transaction, target.Observation, selection, []customerServiceCandidateSelectedFullRow{row}, nil)
			require.ErrorContains(t, err, scenario.expectedError)
			require.NotContains(t, err.Error(), "private forbidden content")
			require.NoError(t, transaction.Rollback())
			_, fingerprint, _, err := captureDatabaseInventory(ctx, fixture.Target)
			require.NoError(t, err)
			require.Equal(t, target.Observation.Fingerprint, fingerprint)
		})
	}
}
