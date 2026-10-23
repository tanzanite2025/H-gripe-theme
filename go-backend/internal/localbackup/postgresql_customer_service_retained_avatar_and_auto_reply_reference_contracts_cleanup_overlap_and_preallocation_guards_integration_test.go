package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLCustomerServiceRetainedAvatarAndAutoReplyUnknownColumnsCannotBeWaivedByCurrentContentOrEmptyValues(t *testing.T) {
	for _, scenario := range []struct {
		name, table, definition string
		source, target          bool
	}{
		{"both_empty_profile_json", "customer_service_agent_profiles", "custom_reference jsonb DEFAULT '{}'", true, true},
		{"both_nullable_disabled_reply_html", "ticket_auto_replies", "custom_reference text", true, true},
		{"both_retained_profile_foreign_key", "customer_service_agent_profiles", "custom_reference bigint REFERENCES media_assets(id)", true, true},
		{"source_only_reply_json", "ticket_auto_replies", "custom_reference jsonb", true, false},
		{"target_only_profile_json", "customer_service_agent_profiles", "custom_reference jsonb", false, true},
		{"both_legacy_named_avatar_field", "customer_service_agent_profiles", "avatar_url text DEFAULT ''", true, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
			ctx := context.Background()
			for index, database := range []*sql.DB{fixture.Source, fixture.Target} {
				if index == 0 && !scenario.source || index == 1 && !scenario.target {
					continue
				}
				_, err := database.Exec(`ALTER TABLE ` + pq.QuoteIdentifier(scenario.table) + ` ADD COLUMN ` + scenario.definition)
				require.NoError(t, err)
			}
			_, err := fixture.Target.Exec(`INSERT INTO ticket_auto_replies(id,message_type,metadata,attachments,is_active) VALUES(1,'text','','[]',false); UPDATE ticket_messages SET content='private chosen current message' WHERE id=1;`)
			require.NoError(t, err)
			before := []string{}
			for _, database := range []*sql.DB{fixture.Source, fixture.Target} {
				_, fingerprint, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				before = append(before, fingerprint)
			}
			source, target, plan, review, request := fixture.inspectAndBindSavedCandidateRehearsalSelections(t, map[string]string{"ticket_messages:1": "preserve_current_row"})
			require.False(t, plan.InspectionComplete)
			require.Positive(t, plan.BlockingIssueCounts["retained_media_owner_column_reference_contract_unregistered"])
			selection, err := fixture.Configuration.inspectCustomerServiceEffectiveSelectedRows(ctx, source, target, plan, review, request, fixture.Manifest, fixture.VerifiedDirectory)
			require.NoError(t, err)
			require.False(t, selection.SelectionDependenciesSatisfied)
			require.Equal(t, 1, selection.EffectiveRowActionCounts["selected_preserve_current"])
			if scenario.target {
				require.Positive(t, selection.DependencyChecks.BlockingIssueCounts["retained_media_owner_column_reference_contract_unregistered"])
			}
			if scenario.source {
				_, _, err := readCustomerServiceCandidateSelectedFullRows(ctx, source, selection)
				require.ErrorContains(t, err, "retained avatar profile or auto reply")
			}
			if scenario.target {
				require.ErrorContains(t, assertCustomerServiceCandidateTablesHaveSupportedWriteContracts(ctx, target.Transaction, target.Observation), "retained avatar profile or auto reply")
			}
			_, err = fixture.Configuration.rehearseCustomerServiceSelectedRowsInDisposableTargetClone(ctx, source, target, plan, review, request, fixture.Manifest, fixture.VerifiedDirectory)
			require.Error(t, err)
			for index, database := range []*sql.DB{fixture.Source, fixture.Target} {
				_, fingerprint, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				require.Equal(t, before[index], fingerprint)
			}
			encoded, err := json.Marshal(selection)
			require.NoError(t, err)
			require.NotContains(t, string(encoded), "private chosen current message")
			if directory := os.Getenv("LOCAL_BACKUP_INTEGRATION_VALIDATION_DIRECTORY"); directory != "" {
				require.NoError(t, os.MkdirAll(directory, 0700))
				require.NoError(t, writeDurableJSON(filepath.Join(directory, scenario.name+"-retained-media-owner-blocked-selection.json"), selection))
			}
		})
	}
}

func TestPostgreSQLCustomerServiceRetainedOwnerUnknownColumnsRejectPersistentCandidateBeforeDatabaseOrMediaAllocation(t *testing.T) {
	for _, table := range []string{"customer_service_agent_profiles", "ticket_auto_replies"} {
		t.Run(table, func(t *testing.T) {
			fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
			cfg := fixture.Configuration
			ctx := context.Background()
			require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
			for _, database := range []*sql.DB{fixture.Source, fixture.Target} {
				_, err := database.Exec(`ALTER TABLE ` + pq.QuoteIdentifier(table) + ` ADD COLUMN custom_reference jsonb`)
				require.NoError(t, err)
			}
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
			require.NoError(t, fixture.Control.QueryRow(`SELECT count(*) FROM pg_database WHERE datname LIKE 'local_backup_candidate_%'`).Scan(&candidates))
			require.Zero(t, candidates)
			after, err := cfg.mediaFingerprint(ctx)
			require.NoError(t, err)
			require.Equal(t, before, after)
			require.NoError(t, manager.discardCustomerServicePersistentCandidate(ctx, task.ID))
			require.NoError(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
		})
	}
}

func verifyCustomerServiceRetainedAvatarCleanupAgainstActualMigrationColumns(t *testing.T, cfg Configuration, database *sql.DB, manifest *Manifest, verifiedDirectory string) {
	t.Helper()
	for _, scenario := range []struct {
		name, eventType, status, payload string
		blocked                          bool
	}{
		{"inactive_deleted_profile_avatar_processed", "customer_service.avatar_cleanup", "processed", `{"url":"/uploads/customer-service/avatars/private-retained-profile.png"}`, true},
		{"inactive_deleted_profile_avatar_pending", "customer_service.avatar_cleanup", "pending", `{"url":"/uploads/customer-service/avatars/private-retained-profile.png"}`, true},
		{"avatar_shared_by_generic_cleanup", "storage.object_cleanup_requested", "processed", `{"resource_type":"private_order_business_attachment_failed_upload","resource_id":"55","object_keys":["customer-service/avatars/private-retained-profile.png"]}`, true},
		{"avatar_shared_by_retention_cleanup", "customer_service.retention_cleanup_requested", "pending", `{"ticket_id":999,"attachment_references":["/uploads/customer-service/avatars/private-retained-profile.png"]}`, true},
		{"ignored_avatar_receipt_stays_suppressed", "customer_service.avatar_cleanup", "ignored", `{"url":"/uploads/customer-service/avatars/private-retained-profile.png"}`, false},
		{"different_complete_avatar_key_has_no_overlap", "customer_service.avatar_cleanup", "processed", `{"url":"/uploads/customer-service/avatars/private-retained-profile.png.extra"}`, false},
	} {
		t.Run("complete_migration_retained_avatar_"+scenario.name, func(t *testing.T) {
			ctx := context.Background()
			aggregateType, aggregateID := "customer_service_agent_profile", ""
			require.NoError(t, database.QueryRow(`SELECT id::text FROM customer_service_agent_profiles WHERE agent_id='synthetic-retained-avatar-owner'`).Scan(&aggregateID))
			if scenario.eventType == "storage.object_cleanup_requested" {
				aggregateType, aggregateID = "order", "55"
			}
			if scenario.eventType == "customer_service.retention_cleanup_requested" {
				aggregateType, aggregateID = "customer_service_conversation", "999"
			}
			_, err := database.Exec(`UPDATE outbox_events SET event_type=$1,aggregate_type=$2,aggregate_id=$3,payload=$4::jsonb,status=$5 WHERE event_key='synthetic-full-schema-avatar-cleanup'`, scenario.eventType, aggregateType, aggregateID, scenario.payload, scenario.status)
			require.NoError(t, err)
			t.Cleanup(func() {
				_, err := database.Exec(`UPDATE outbox_events SET event_type='customer_service.avatar_cleanup',aggregate_type='customer_service_agent_profile',aggregate_id=(SELECT id::text FROM customer_service_agent_profiles WHERE agent_id='synthetic-retained-avatar-owner'),payload='{"url":"/uploads/customer-service/avatars/private-retained-profile.png"}',status='ignored' WHERE event_key='synthetic-full-schema-avatar-cleanup'`)
				require.NoError(t, err)
			})
			transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
			require.NoError(t, err)
			defer transaction.Rollback()
			observation, err := inspectModuleRecoveryDatabaseObservation(ctx, transaction)
			require.NoError(t, err)
			report, err := cfg.inspectCustomerServiceConversationRows(ctx, transaction, transaction, observation, observation, manifest, verifiedDirectory, defaultModuleRecoveryRowInspectionLimits())
			require.NoError(t, err)
			require.True(t, report.Complete)
			require.Zero(t, report.IssueCounts["avatar_cleanup_owner_schema_missing"])
			require.Equal(t, scenario.blocked, report.IssueCounts["durable_cleanup_overlaps_retained_recovery_object"] > 0, "%v", report.IssueCounts)
			if !scenario.blocked {
				require.Zero(t, report.BlockingIssues, "%v", report.IssueCounts)
			}
			_, after, _, err := captureDatabaseTransactionInventory(ctx, transaction)
			require.NoError(t, err)
			require.Equal(t, observation.Fingerprint, after)
			encoded, err := json.Marshal(report)
			require.NoError(t, err)
			for _, privateValue := range []string{"private-retained-profile.png", "private retained avatar profile", "synthetic-retained-avatar-owner"} {
				require.NotContains(t, string(encoded), privateValue)
			}
			if directory := os.Getenv("LOCAL_BACKUP_INTEGRATION_VALIDATION_DIRECTORY"); directory != "" {
				require.NoError(t, os.MkdirAll(directory, 0700))
				require.NoError(t, writeDurableJSON(filepath.Join(directory, "complete-migration-"+scenario.name+"-avatar-cleanup-report.json"), report))
			}
		})
	}
}
