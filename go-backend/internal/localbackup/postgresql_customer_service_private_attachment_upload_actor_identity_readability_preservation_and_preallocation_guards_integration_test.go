package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

const customerServiceRecoveryValidVisitorUploadActorSyntheticRows = `
UPDATE tickets SET user_id=NULL,customer_user_id=NULL,visitor_session_hash=repeat('a',64) WHERE id=1;
UPDATE media_upload_sessions SET upload_actor_type='visitor',upload_actor_account_id=0,upload_actor_visitor_session_hash=repeat('a',64);
UPDATE media_assets SET uploader_actor_type='visitor',uploader_id=0,uploader_visitor_session_hash=repeat('a',64) WHERE id=1;
`

const customerServiceRecoveryPublicReusableAttachmentSyntheticRows = `
DELETE FROM ticket_message_media_attachments; DELETE FROM media_upload_sessions;
INSERT INTO ticket_message_media_attachments(id,ticket_message_id,media_asset_id) VALUES(1,1,2);
`

func TestPostgreSQLCustomerServicePrivateAttachmentUploadActorsMatchActualReadIdentityWithoutMutatingData(t *testing.T) {
	for _, scenario := range []struct {
		name, bothSQL, sourceSQL, targetSQL, code string
		choices                                   map[string]string
		incomplete, originalOnly                  bool
	}{
		{name: "valid_customer_upload"},
		{name: "valid_customer_legacy_ticket_owner_fallback", bothSQL: `UPDATE tickets SET customer_user_id=NULL WHERE id=1;`},
		{name: "valid_administrator_upload", bothSQL: `UPDATE media_upload_sessions SET upload_actor_type='administrator',upload_actor_account_id=3; UPDATE media_assets SET uploader_actor_type='administrator',uploader_id=3 WHERE id=1;`},
		{name: "valid_visitor_upload", bothSQL: customerServiceRecoveryValidVisitorUploadActorSyntheticRows},
		{name: "same_key_different_registered_account", bothSQL: `UPDATE media_assets SET uploader_id=2 WHERE id=1;`, code: "upload_receipt_asset_actor_identity_conflict"},
		{name: "same_account_different_actor_namespace", bothSQL: `UPDATE media_assets SET uploader_actor_type='administrator' WHERE id=1;`, code: "upload_receipt_asset_actor_identity_conflict"},
		{name: "legacy_media_identity_is_not_a_customer_identity", bothSQL: `UPDATE media_assets SET uploader_actor_type='legacy_account' WHERE id=1;`, code: "upload_receipt_asset_actor_identity_conflict"},
		{name: "matching_legacy_receipt_is_still_unsupported", bothSQL: `UPDATE media_assets SET uploader_actor_type='legacy_account' WHERE id=1; UPDATE media_upload_sessions SET upload_actor_type='legacy_account';`, code: "upload_actor_identity_invalid"},
		{name: "matching_customer_identity_cannot_carry_visitor_hash", bothSQL: `UPDATE media_upload_sessions SET upload_actor_visitor_session_hash=repeat('a',64); UPDATE media_assets SET uploader_visitor_session_hash=repeat('a',64) WHERE id=1;`, code: "upload_actor_identity_invalid"},
		{name: "administrator_account_zero_is_invalid", bothSQL: `UPDATE media_upload_sessions SET upload_actor_type='administrator',upload_actor_account_id=0; UPDATE media_assets SET uploader_actor_type='administrator',uploader_id=0 WHERE id=1;`, code: "upload_actor_identity_invalid"},
		{name: "administrator_cannot_carry_visitor_hash", bothSQL: `UPDATE media_upload_sessions SET upload_actor_type='administrator',upload_actor_visitor_session_hash=repeat('a',64); UPDATE media_assets SET uploader_actor_type='administrator',uploader_visitor_session_hash=repeat('a',64) WHERE id=1;`, code: "upload_actor_identity_invalid"},
		{name: "customer_account_zero_is_invalid", bothSQL: `UPDATE media_upload_sessions SET upload_actor_account_id=0; UPDATE media_assets SET uploader_id=0 WHERE id=1;`, code: "upload_actor_identity_invalid"},
		{name: "null_receipt_account_cannot_be_treated_as_zero", bothSQL: customerServiceRecoveryValidVisitorUploadActorSyntheticRows + `UPDATE media_upload_sessions SET upload_actor_account_id=NULL;`, code: "upload_actor_identity_invalid"},
		{name: "null_registered_visitor_account_cannot_be_treated_as_zero", bothSQL: customerServiceRecoveryValidVisitorUploadActorSyntheticRows + `UPDATE media_assets SET uploader_id=NULL WHERE id=1;`, code: "upload_receipt_asset_actor_identity_conflict"},
		{name: "visitor_cannot_carry_account", bothSQL: customerServiceRecoveryValidVisitorUploadActorSyntheticRows + `UPDATE media_upload_sessions SET upload_actor_account_id=1; UPDATE media_assets SET uploader_id=1 WHERE id=1;`, code: "upload_actor_identity_invalid"},
		{name: "matching_short_visitor_hash_is_invalid", bothSQL: customerServiceRecoveryValidVisitorUploadActorSyntheticRows + `UPDATE tickets SET visitor_session_hash='short-private-visitor-hash' WHERE id=1; UPDATE media_upload_sessions SET upload_actor_visitor_session_hash='short-private-visitor-hash'; UPDATE media_assets SET uploader_visitor_session_hash='short-private-visitor-hash' WHERE id=1;`, code: "upload_actor_identity_invalid"},
		{name: "matching_uppercase_visitor_hash_is_invalid", bothSQL: customerServiceRecoveryValidVisitorUploadActorSyntheticRows + `UPDATE tickets SET visitor_session_hash=repeat('A',64) WHERE id=1; UPDATE media_upload_sessions SET upload_actor_visitor_session_hash=repeat('A',64); UPDATE media_assets SET uploader_visitor_session_hash=repeat('A',64) WHERE id=1;`, code: "upload_actor_identity_invalid"},
		{name: "matching_nonhex_visitor_hash_is_invalid", bothSQL: customerServiceRecoveryValidVisitorUploadActorSyntheticRows + `UPDATE tickets SET visitor_session_hash=repeat('g',64) WHERE id=1; UPDATE media_upload_sessions SET upload_actor_visitor_session_hash=repeat('g',64); UPDATE media_assets SET uploader_visitor_session_hash=repeat('g',64) WHERE id=1;`, code: "upload_actor_identity_invalid"},
		{name: "registered_visitor_hash_differs", bothSQL: customerServiceRecoveryValidVisitorUploadActorSyntheticRows + `UPDATE media_assets SET uploader_visitor_session_hash=repeat('b',64) WHERE id=1;`, code: "upload_receipt_asset_actor_identity_conflict"},
		{name: "registered_identity_alone_is_invalid", bothSQL: `UPDATE media_assets SET uploader_visitor_session_hash=repeat('a',64) WHERE id=1;`, code: "upload_receipt_asset_actor_identity_conflict"},
		{name: "private_binding_without_receipt", bothSQL: `UPDATE ticket_message_media_attachments SET upload_session_id=NULL; DELETE FROM media_upload_sessions;`, code: "private_attachment_upload_receipt_missing"},
		{name: "bound_receipt_without_registered_asset", bothSQL: `UPDATE media_upload_sessions SET media_asset_id=NULL;`, code: "upload_receipt_registered_asset_missing"},
		{name: "public_reusable_attachment_without_receipt", bothSQL: customerServiceRecoveryPublicReusableAttachmentSyntheticRows},
		{name: "public_legacy_reuse_does_not_require_modern_actor_columns", bothSQL: customerServiceRecoveryPublicReusableAttachmentSyntheticRows + `ALTER TABLE media_assets DROP COLUMN uploader_actor_type; ALTER TABLE media_assets DROP COLUMN uploader_visitor_session_hash;`},
		{name: "source_only_registered_actor_conflict", sourceSQL: `UPDATE media_assets SET uploader_id=2 WHERE id=1;`, code: "upload_receipt_asset_actor_identity_conflict", originalOnly: true},
		{name: "target_only_registered_actor_conflict", targetSQL: `UPDATE media_assets SET uploader_id=2 WHERE id=1;`, code: "upload_receipt_asset_actor_identity_conflict"},
		{name: "preserving_current_message_cannot_waive_bad_registered_actor", bothSQL: `UPDATE media_assets SET uploader_id=2 WHERE id=1;`, targetSQL: `UPDATE ticket_messages SET content='private newer current message' WHERE id=1;`, choices: map[string]string{"ticket_messages:1": "preserve_current_row"}, code: "upload_receipt_asset_actor_identity_conflict"},
		{name: "source_missing_registered_actor_column", sourceSQL: `ALTER TABLE media_assets DROP COLUMN uploader_actor_type;`, code: "upload_actor_identity_schema_missing", incomplete: true, originalOnly: true},
		{name: "target_missing_registered_visitor_column", targetSQL: `ALTER TABLE media_assets DROP COLUMN uploader_visitor_session_hash;`, code: "upload_actor_identity_schema_missing", incomplete: true},
		{name: "both_missing_registered_account_column", bothSQL: `ALTER TABLE media_assets DROP COLUMN uploader_id;`, code: "upload_actor_identity_schema_missing", incomplete: true},
		{name: "both_missing_receipt_actor_column", bothSQL: `ALTER TABLE media_upload_sessions DROP COLUMN upload_actor_type;`, code: "upload_actor_identity_schema_missing", incomplete: true},
		{name: "both_missing_receipt_account_column", bothSQL: `ALTER TABLE media_upload_sessions DROP COLUMN upload_actor_account_id;`, code: "upload_actor_identity_schema_missing", incomplete: true},
		{name: "both_missing_receipt_visitor_column", bothSQL: `ALTER TABLE media_upload_sessions DROP COLUMN upload_actor_visitor_session_hash;`, code: "upload_actor_identity_schema_missing", incomplete: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
			ctx := context.Background()
			for index, database := range []*sql.DB{fixture.Source, fixture.Target} {
				statement := scenario.bothSQL
				if index == 0 {
					statement += scenario.sourceSQL
				} else {
					statement += scenario.targetSQL
				}
				if statement != "" {
					_, err := database.ExecContext(context.Background(), statement)
					require.NoError(t, err)
				}
			}
			var before []string
			for _, database := range []*sql.DB{fixture.Source, fixture.Target} {
				_, fingerprint, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				before = append(before, fingerprint)
			}
			mediaBefore, err := fixture.Configuration.mediaFingerprint(ctx)
			require.NoError(t, err)
			source, target, plan, review, request := fixture.inspectAndBindSavedCandidateRehearsalSelections(t, scenario.choices)
			selection, err := fixture.Configuration.inspectCustomerServiceEffectiveSelectedRows(ctx, source, target, plan, review, request, fixture.Manifest, fixture.VerifiedDirectory)
			require.NoError(t, err)
			require.False(t, plan.ExecutionAvailable)
			require.Equal(t, scenario.code == "", selection.SelectionDependenciesSatisfied, "%v", selection.DependencyChecks.IssueCounts)
			if scenario.code == "" {
				require.Empty(t, plan.BlockingIssueCounts)
				require.Zero(t, selection.DependencyChecks.BlockingIssues)
			} else {
				require.Positive(t, plan.BlockingIssueCounts[scenario.code], "%v", plan.BlockingIssueCounts)
				if scenario.originalOnly {
					require.Zero(t, selection.DependencyChecks.IssueCounts[scenario.code])
				} else {
					require.Positive(t, selection.DependencyChecks.IssueCounts[scenario.code], "%v", selection.DependencyChecks.IssueCounts)
				}
			}
			require.Equal(t, !scenario.incomplete, plan.InspectionComplete)
			require.Equal(t, !scenario.incomplete || scenario.originalOnly, selection.DependencyChecks.Complete)
			for index, database := range []*sql.DB{fixture.Source, fixture.Target} {
				_, fingerprint, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				require.Equal(t, before[index], fingerprint)
			}
			mediaAfter, err := fixture.Configuration.mediaFingerprint(ctx)
			require.NoError(t, err)
			require.Equal(t, mediaBefore, mediaAfter)
			encoded, err := json.Marshal(selection)
			require.NoError(t, err)
			for _, privateValue := range []string{strings.Repeat("a", 64), strings.Repeat("A", 64), strings.Repeat("g", 64), "short-private-visitor-hash", "synthetic-upload-one", "private newer current message", "originals/synthetic.txt"} {
				require.NotContains(t, string(encoded), privateValue)
			}
			if directory := os.Getenv("LOCAL_BACKUP_INTEGRATION_VALIDATION_DIRECTORY"); directory != "" {
				require.NoError(t, os.MkdirAll(directory, 0700))
				require.NoError(t, writeDurableJSON(filepath.Join(directory, scenario.name+"-upload-actor-selection.json"), selection))
			}
		})
	}
}

func TestPostgreSQLCustomerServicePrivateUploadActorConflictRejectsPersistentCandidateBeforeAllocation(t *testing.T) {
	fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
	ctx := context.Background()
	_, err := fixture.Target.ExecContext(context.Background(), `UPDATE media_assets SET uploader_actor_type='administrator' WHERE id=1;`)
	require.NoError(t, err)
	cfg := fixture.Configuration
	require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
	manager, task, request := registerVerifiedUploadAndSavedSelectionsForCustomerServicePersistentCandidateTest(t, fixture)
	before, err := cfg.mediaFingerprint(ctx)
	require.NoError(t, err)
	_, err = manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(ctx, task.ID, request)
	require.Error(t, err)
	record, err := manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
	require.NoError(t, err)
	require.Zero(t, record.Identity.DatabaseOID)
	require.Nil(t, record.MediaUnion)
	var candidates int
	require.NoError(t, fixture.Control.QueryRowContext(context.Background(), `SELECT count(*) FROM pg_database WHERE datname LIKE 'local_backup_candidate_%'`).Scan(&candidates))
	require.Zero(t, candidates)
	after, err := cfg.mediaFingerprint(ctx)
	require.NoError(t, err)
	require.Equal(t, before, after)
	require.NoError(t, manager.discardCustomerServicePersistentCandidate(ctx, task.ID))
	require.NoError(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
}

func verifyCustomerServicePrivateUploadActorReadabilityAgainstActualMigrationSchema(t *testing.T, cfg Configuration, database *sql.DB, manifest *Manifest, verifiedDirectory string) {
	t.Helper()
	for _, scenario := range []struct{ name, statement, code string }{
		{name: "valid_customer_upload"},
		{name: "different_account_same_path", statement: `UPDATE media_assets SET uploader_id=uploader_id+100000 WHERE visibility='private';`, code: "upload_receipt_asset_actor_identity_conflict"},
		{name: "same_account_different_actor_type", statement: `UPDATE media_assets SET uploader_actor_type='administrator' WHERE visibility='private';`, code: "upload_receipt_asset_actor_identity_conflict"},
		{name: "legacy_identity_not_inferred", statement: `UPDATE media_assets SET uploader_actor_type='legacy_account' WHERE visibility='private';`, code: "upload_receipt_asset_actor_identity_conflict"},
		{name: "private_binding_missing_receipt", statement: `UPDATE ticket_message_media_attachments SET upload_session_id=NULL;`, code: "private_attachment_upload_receipt_missing"},
	} {
		t.Run("complete_migration_upload_actor_"+scenario.name, func(t *testing.T) {
			ctx := context.Background()
			transaction, err := database.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
			require.NoError(t, err)
			defer transaction.Rollback()
			if scenario.statement != "" {
				_, err = transaction.ExecContext(context.Background(), scenario.statement)
				require.NoError(t, err)
			}
			observation, err := inspectModuleRecoveryDatabaseObservation(ctx, transaction)
			require.NoError(t, err)
			report, err := cfg.inspectCustomerServiceConversationRows(ctx, transaction, transaction, observation, observation, manifest, verifiedDirectory, defaultModuleRecoveryRowInspectionLimits())
			require.NoError(t, err)
			require.True(t, report.Complete)
			if scenario.code == "" {
				require.Zero(t, report.BlockingIssues, "%v", report.IssueCounts)
			} else {
				require.Positive(t, report.IssueCounts[scenario.code], "%v", report.IssueCounts)
			}
			_, after, _, err := captureDatabaseTransactionInventory(ctx, transaction)
			require.NoError(t, err)
			require.Equal(t, observation.Fingerprint, after)
			if directory := os.Getenv("LOCAL_BACKUP_INTEGRATION_VALIDATION_DIRECTORY"); directory != "" {
				require.NoError(t, os.MkdirAll(directory, 0700))
				require.NoError(t, writeDurableJSON(filepath.Join(directory, "complete-migration-"+scenario.name+"-upload-actor-report.json"), report))
			}
		})
	}
}
