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

const customerServiceRowInspectionSyntheticSchema = `
CREATE TABLE schema_migrations(version bigint,dirty boolean); INSERT INTO schema_migrations VALUES(434,false);
CREATE TABLE users(id bigint PRIMARY KEY,name text);
CREATE TABLE tickets(id bigint PRIMARY KEY,category text,conversation_id text,user_id bigint REFERENCES users(id),customer_user_id bigint,visitor_session_hash text,assigned_to bigint,deleted_at timestamptz);
CREATE TABLE ticket_messages(id bigint PRIMARY KEY,ticket_id bigint REFERENCES tickets(id),user_id bigint REFERENCES users(id),content text,attachments text,attachment_reference_format text,deleted_at timestamptz);
ALTER TABLE ticket_messages ADD COLUMN message_type text DEFAULT 'text';
ALTER TABLE ticket_messages ADD COLUMN metadata text DEFAULT '';
CREATE TABLE ticket_auto_replies(id bigint PRIMARY KEY,message_type text DEFAULT 'text',metadata text DEFAULT '',attachments text DEFAULT '[]',is_active boolean DEFAULT true);
CREATE TABLE customer_service_inbox_states(id bigint PRIMARY KEY,ticket_id bigint REFERENCES tickets(id),recipient_user_id bigint REFERENCES users(id),last_read_message_id bigint);
CREATE TABLE visitor_profiles(id bigint PRIMARY KEY,customer_service_visitor_hash text,user_id bigint,deleted_at timestamptz);
CREATE TABLE media_assets(id bigint PRIMARY KEY,storage_key text,url text,visibility text,size bigint,content_sha256 text,library_partition text,customer_service_media_purpose text,classification_method text,upload_origin text,uploader_id bigint,customer_service_derivative_write_token text,customer_service_derivative_write_expires_at timestamptz);
ALTER TABLE media_assets ADD COLUMN uploader_actor_type text NOT NULL DEFAULT 'legacy_account';
ALTER TABLE media_assets ADD COLUMN uploader_visitor_session_hash text NOT NULL DEFAULT '';
CREATE TABLE media_asset_derivatives(id bigint PRIMARY KEY,media_asset_id bigint REFERENCES media_assets(id),storage_key text,url text,size bigint);
CREATE TABLE media_upload_sessions(id text PRIMARY KEY,business_context_type text,business_context_id bigint REFERENCES tickets(id),upload_actor_type text,upload_actor_account_id bigint,upload_actor_visitor_session_hash text,media_asset_id bigint REFERENCES media_assets(id),object_key text,status text);
CREATE TABLE ticket_message_media_attachments(id bigint PRIMARY KEY,ticket_message_id bigint REFERENCES ticket_messages(id),media_asset_id bigint REFERENCES media_assets(id),upload_session_id text REFERENCES media_upload_sessions(id));
CREATE TABLE outbox_events(id bigint PRIMARY KEY,event_key text UNIQUE NOT NULL,event_type text NOT NULL,aggregate_type text NOT NULL,aggregate_id text NOT NULL,payload jsonb NOT NULL DEFAULT '{}',status text NOT NULL,attempts int NOT NULL DEFAULT 0,locked_at timestamptz,locked_by text NOT NULL DEFAULT '',processed_at timestamptz,uncertain_at timestamptz,reconcile_after timestamptz,deleted_at timestamptz);
CREATE TABLE customer_service_agent_profiles(id bigint PRIMARY KEY,avatar text);
CREATE TABLE customer_service_retention_policies(id bigint PRIMARY KEY,keep_days bigint);
CREATE TABLE orders(id bigint PRIMARY KEY,user_id bigint REFERENCES users(id));
CREATE TABLE product_media(id bigint PRIMARY KEY,media_asset_id bigint REFERENCES media_assets(id));
CREATE TABLE synthetic_composite_parents(first_key bigint,second_key text,PRIMARY KEY(first_key,second_key));
ALTER TABLE tickets ADD COLUMN synthetic_first_key bigint; ALTER TABLE tickets ADD COLUMN synthetic_second_key text;
ALTER TABLE tickets ADD FOREIGN KEY(synthetic_first_key,synthetic_second_key) REFERENCES synthetic_composite_parents(first_key,second_key);
ALTER TABLE tickets ADD COLUMN ticket_number text;
CREATE UNIQUE INDEX synthetic_ticket_number_unique ON tickets(ticket_number);
CREATE UNIQUE INDEX synthetic_conversation_id_unique ON tickets(conversation_id) WHERE conversation_id IS NOT NULL;
CREATE UNIQUE INDEX synthetic_inbox_recipient_ticket_unique ON customer_service_inbox_states(recipient_user_id,ticket_id);
ALTER TABLE ticket_message_media_attachments ADD COLUMN attachment_index integer DEFAULT 0;
CREATE UNIQUE INDEX synthetic_attachment_position_unique ON ticket_message_media_attachments(ticket_message_id,attachment_index);
CREATE UNIQUE INDEX synthetic_attachment_asset_unique ON ticket_message_media_attachments(ticket_message_id,media_asset_id);
CREATE UNIQUE INDEX synthetic_attachment_upload_unique ON ticket_message_media_attachments(upload_session_id);
CREATE UNIQUE INDEX synthetic_upload_object_unique ON media_upload_sessions(object_key);
CREATE UNIQUE INDEX synthetic_upload_asset_unique ON media_upload_sessions(media_asset_id);
ALTER TABLE tickets ADD COLUMN status_version integer NOT NULL DEFAULT 1;
ALTER TABLE customer_service_inbox_states ADD COLUMN assignment_version integer NOT NULL DEFAULT 1;
ALTER TABLE customer_service_inbox_states ADD COLUMN unread_count integer NOT NULL DEFAULT 0;
`

const customerServiceRowInspectionSyntheticRows = `
TRUNCATE users,tickets,ticket_messages,ticket_auto_replies,customer_service_inbox_states,visitor_profiles,media_assets,media_asset_derivatives,media_upload_sessions,ticket_message_media_attachments,outbox_events,customer_service_agent_profiles,customer_service_retention_policies,orders,product_media,synthetic_composite_parents;
INSERT INTO users VALUES(1,'private synthetic customer name'),(2,'unrelated customer'),(3,'staff');
INSERT INTO synthetic_composite_parents VALUES(10,'private composite dependency');
INSERT INTO tickets(id,category,conversation_id,user_id,customer_user_id,visitor_session_hash,assigned_to,synthetic_first_key,synthetic_second_key) VALUES(1,'customer_service','synthetic-conversation-one',1,1,'',3,10,'private composite dependency'),(2,'order','synthetic-order-ticket',2,2,'',NULL,NULL,NULL),(3,'customer_service','synthetic-visitor-conversation',NULL,NULL,repeat('a',64),NULL,NULL,NULL);
UPDATE tickets SET ticket_number='synthetic-ticket-'||id;
INSERT INTO ticket_messages VALUES(1,1,1,'private retained synthetic message','[]','media_asset_bindings','2026-10-01'),(2,2,2,'unrelated order question','[]','legacy_urls',NULL),(3,3,NULL,'private visitor message','[]','legacy_urls',NULL),(4,1,1,'private legacy message','["/uploads/legacy-attachment.txt"]','legacy_urls',NULL);
INSERT INTO customer_service_inbox_states VALUES(1,1,3,1),(2,3,3,0);
INSERT INTO visitor_profiles VALUES(1,repeat('a',64),NULL,NULL);
INSERT INTO media_assets VALUES(1,'customer-service/attachments/originals/synthetic.txt','/uploads-private/customer-service/attachments/originals/synthetic.txt','private',8,'','customer_service','conversation_attachment','recorded_upload_context','customer_service_conversation_upload',1,'',NULL),(2,'legacy-attachment.txt','/uploads/legacy-attachment.txt','public',8,'','customer_service','conversation_attachment','manual_reviewed','legacy_unknown',1,'',NULL);
UPDATE media_assets SET uploader_actor_type='customer' WHERE id=1;
INSERT INTO media_asset_derivatives VALUES(1,1,'customer-service/attachments/derivatives/synthetic.txt','/uploads-private/customer-service/attachments/derivatives/synthetic.txt',8);
INSERT INTO media_upload_sessions VALUES('synthetic-upload-one','customer_service',1,'customer',1,'',1,'customer-service/attachments/originals/synthetic.txt','bound');
INSERT INTO ticket_message_media_attachments VALUES(1,1,1,'synthetic-upload-one');
INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,status) VALUES(1,'synthetic-order-event-1','order.paid','order','1','processed'),(2,'synthetic-order-event-2','order.paid','order','1','pending'); INSERT INTO orders VALUES(1,2); INSERT INTO customer_service_retention_policies VALUES(1,365);
`

func TestPostgreSQLCustomerServiceRowInspectionPreservesSharedDataAndDetectsSpecificConversationAndFileConflicts(t *testing.T) {
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
		_, err = database.ExecContext(ctx, customerServiceRowInspectionSyntheticSchema)
		require.NoError(t, err)
	}
	verifiedDirectory := filepath.Join(t.TempDir(), "verified")
	filePaths := []string{"media/uploads_private/customer-service/attachments/originals/synthetic.txt", "media/uploads_private/customer-service/attachments/derivatives/synthetic.txt", "media/uploads/legacy-attachment.txt"}
	resetSyntheticConversationDatabasesAndAttachmentFiles := func() {
		for _, database := range databases {
			_, err := database.ExecContext(ctx, customerServiceRowInspectionSyntheticRows)
			require.NoError(t, err)
		}
		for _, path := range filePaths {
			parts := strings.SplitN(strings.TrimPrefix(path, "media/"), "/", 2)
			for _, filename := range []string{filepath.Join(verifiedDirectory, filepath.FromSlash(path)), filepath.Join(cfg.StorageRoots[parts[0]], filepath.FromSlash(parts[1]))} {
				require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
				require.NoError(t, os.WriteFile(filename, []byte("original"), 0600))
			}
		}
	}
	manifest := &Manifest{Files: []FileIdentity{}}
	resetSyntheticConversationDatabasesAndAttachmentFiles()
	for _, path := range filePaths {
		identity, err := fileIdentity(filepath.Join(verifiedDirectory, filepath.FromSlash(path)), path)
		require.NoError(t, err)
		manifest.Files = append(manifest.Files, identity)
	}
	inspectSyntheticCustomerServiceConversationDependencies := func(limits moduleRecoveryRowInspectionLimits) *ModuleRecoveryRowInspection {
		var transactions []*sql.Tx
		var observations []*moduleRecoveryDatabaseObservation
		for _, database := range databases {
			transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
			require.NoError(t, err)
			defer transaction.Rollback()
			transactions = append(transactions, transaction)
			observation, err := inspectModuleRecoveryDatabaseObservation(ctx, transaction)
			require.NoError(t, err)
			observations = append(observations, observation)
		}
		report, err := cfg.inspectCustomerServiceConversationRows(ctx, transactions[0], transactions[1], observations[0], observations[1], manifest, verifiedDirectory, limits)
		require.NoError(t, err)
		encoded, err := json.Marshal(report)
		require.NoError(t, err)
		for _, privateValue := range []string{"private retained synthetic message", "private synthetic customer name", strings.Repeat("a", 64), "private composite dependency", "synthetic-conversation-one"} {
			require.NotContains(t, string(encoded), privateValue)
		}
		return report
	}
	for _, scenario := range []struct {
		name            string
		database        int
		statement, code string
	}{
		{"unrelated_shared_customer_changes", 1, `UPDATE users SET name='unrelated newer name' WHERE id=2`, ""},
		{"preserved_configuration_and_outbox_changes", 1, `UPDATE customer_service_retention_policies SET keep_days=999; UPDATE outbox_events SET status='pending' WHERE id=1`, ""},
		{"referenced_shared_customer_changes", 1, `UPDATE users SET name='newer referenced name' WHERE id=1`, "shared_dependency_row_differs"},
		{"message_identifier_reassigned_to_order_ticket", 1, `UPDATE ticket_messages SET ticket_id=2 WHERE id=1`, "row_identity_reassigned"},
		{"retained_message_content_changes", 1, `UPDATE ticket_messages SET content='newer retained message' WHERE id=1`, "conversation_row_content_differs"},
		{"missing_non_foreign_key_customer", 0, `UPDATE tickets SET customer_user_id=99 WHERE id=1`, "referenced_row_missing"},
		{"missing_visitor_browser_context_is_review_only", 0, `DELETE FROM visitor_profiles`, "visitor_context_profile_missing"},
		{"read_cursor_crosses_conversations", 0, `UPDATE customer_service_inbox_states SET last_read_message_id=2 WHERE id=1`, "inbox_read_cursor_outside_conversation"},
		{"zero_read_cursor_is_allowed", 0, `UPDATE customer_service_inbox_states SET last_read_message_id=0 WHERE id=1`, ""},
		{"upload_receipt_crosses_conversations", 0, `UPDATE media_upload_sessions SET business_context_id=3 WHERE id='synthetic-upload-one'`, "attachment_binding_conversation_conflict"},
		{"upload_actor_crosses_customers", 0, `UPDATE media_upload_sessions SET upload_actor_account_id=2 WHERE id='synthetic-upload-one'`, "upload_actor_conversation_identity_conflict"},
		{"unfinished_upload", 1, `UPDATE media_upload_sessions SET status='ready' WHERE id='synthetic-upload-one'`, "unfinished_conversation_upload"},
		{"malformed_legacy_attachment_json", 0, `UPDATE ticket_messages SET attachments='broken-json' WHERE id=4`, "legacy_attachment_json_invalid"},
		{"external_legacy_attachment", 0, `UPDATE ticket_messages SET attachments='["https://unrelated.invalid/file"]' WHERE id=4`, "legacy_attachment_reference_unresolved"},
		{"historical_media_ownership_remains_unknown", 0, `UPDATE media_assets SET library_partition='unclassified',classification_method='historical_unknown',customer_service_media_purpose='' WHERE id=2`, "historical_attachment_media_ownership_unknown"},
		{"reusable_attachment_keeps_shared_ownership", 0, `UPDATE media_assets SET customer_service_media_purpose='reusable_reply' WHERE id=2`, "attachment_media_ownership_review_required"},
		{"derivative_write_lease", 1, `UPDATE media_assets SET customer_service_derivative_write_token='synthetic-lease' WHERE id=1`, "attachment_derivative_write_lease_present"},
		{"preserved_product_references_attachment", 1, `INSERT INTO product_media VALUES(1,1)`, "preserved_row_references_conversation_object"},
		{"new_target_conversation_row", 1, `INSERT INTO ticket_messages VALUES(5,1,1,'newer message','[]','legacy_urls',NULL)`, "target_only_conversation_row"},
		{"metadata_image_registered_original", 0, `UPDATE ticket_messages SET message_type='image',metadata='{"url":"/uploads-private/customer-service/attachments/originals/synthetic.txt","caption":"private metadata caption"}' WHERE id=3`, ""},
		{"metadata_video_registered_derivative", 0, `UPDATE ticket_messages SET message_type='video',metadata='{"thumbnail":"/uploads-private/customer-service/attachments/derivatives/synthetic.txt"}' WHERE id=3`, ""},
		{"metadata_product_navigation_is_not_media", 0, `UPDATE ticket_messages SET message_type='product',metadata='{"url":"https://navigation.invalid","caption":"/uploads/imaginary.txt"}' WHERE id=3`, "conversation_row_content_differs"},
		{"metadata_faq_unregistered_file", 0, `UPDATE ticket_messages SET message_type='faq',metadata='{"faq":{"answer_image_url":"/uploads/missing.png"}}' WHERE id=3`, "legacy_attachment_registration_missing"},
		{"metadata_wrong_media_field_type", 0, `UPDATE ticket_messages SET message_type='image',metadata='{"url":123}' WHERE id=3`, "structured_message_metadata_media_fields_invalid"},
		{"metadata_unknown_type_requires_contract", 0, `UPDATE ticket_messages SET message_type='custom_card',metadata='{"image":"/uploads/legacy-attachment.txt"}' WHERE id=3`, "structured_message_type_reference_contract_missing"},
		{"unknown_type_empty_metadata_source", 0, `UPDATE ticket_messages SET message_type='custom_html',metadata='' WHERE id=3`, "structured_message_type_reference_contract_missing"},
		{"unknown_type_null_metadata_source", 0, `UPDATE ticket_messages SET message_type='custom_html',metadata=NULL WHERE id=3`, "structured_message_type_reference_contract_missing"},
		{"unknown_type_retained_other_ticket", 1, `UPDATE ticket_messages SET message_type='custom_html',metadata='' WHERE id=2`, "structured_message_type_reference_contract_missing"},
		{"unknown_type_disabled_retained_auto_reply", 1, `INSERT INTO ticket_auto_replies(id,message_type,metadata,is_active) VALUES(1,'custom_html','',false)`, "structured_message_type_reference_contract_missing"},
		{"metadata_external_media_blocks", 0, `UPDATE ticket_messages SET message_type='image',metadata='{"url":"https://external.invalid/a.png"}' WHERE id=3`, "structured_message_media_reference_unresolved"},
		{"metadata_retained_auto_reply_invalid", 1, `INSERT INTO ticket_auto_replies(id,message_type,metadata,is_active) VALUES(1,'image','{"url":[]}',false)`, "structured_message_metadata_media_fields_invalid"},
		{"metadata_retained_other_ticket_invalid", 1, `UPDATE ticket_messages SET message_type='image',metadata='broken-json' WHERE id=2`, "structured_message_metadata_media_fields_invalid"},
		{"target_shared_row_missing", 1, `UPDATE tickets SET user_id=2,customer_user_id=2 WHERE id=1; UPDATE ticket_messages SET user_id=2 WHERE user_id=1; UPDATE media_assets SET uploader_id=2; UPDATE media_upload_sessions SET upload_actor_account_id=2; DELETE FROM users WHERE id=1`, "target_dependency_row_missing"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			resetSyntheticConversationDatabasesAndAttachmentFiles()
			_, err := databases[scenario.database].ExecContext(ctx, scenario.statement)
			require.NoError(t, err)
			report := inspectSyntheticCustomerServiceConversationDependencies(defaultModuleRecoveryRowInspectionLimits())
			if scenario.name == "metadata_image_registered_original" || scenario.name == "metadata_video_registered_derivative" {
				require.Zero(t, report.IssueCounts["legacy_attachment_registration_missing"])
				require.Zero(t, report.IssueCounts["structured_message_metadata_media_fields_invalid"])
				require.Equal(t, 3, report.Files.PackageVerifiedFiles)
			}
			if scenario.code != "" {
				require.Positive(t, report.IssueCounts[scenario.code], "%+v", report.IssueCounts)
			} else if scenario.name != "zero_read_cursor_is_allowed" && !strings.HasPrefix(scenario.name, "metadata_") {
				require.Zero(t, report.BlockingIssues, "%+v", report.IssueCounts)
			}
			if scenario.name == "missing_visitor_browser_context_is_review_only" {
				for _, issue := range report.Issues {
					if issue.Code == scenario.code {
						require.Equal(t, "review", issue.Severity)
					}
				}
			}
		})
	}
	t.Run("recovery_plan_actions_preserve_target_baseline_and_never_authorize_writes", func(t *testing.T) {
		resetSyntheticConversationDatabasesAndAttachmentFiles()
		baseline := inspectSyntheticCustomerServiceConversationDependencies(defaultModuleRecoveryRowInspectionLimits()).ConversationRecoveryPlan
		require.NotNil(t, baseline)
		require.False(t, baseline.ExecutionAvailable)
		require.Len(t, baseline.TargetPreservation, 17)
		_, err := databases[0].ExecContext(ctx, `INSERT INTO ticket_messages VALUES(6,1,1,'source only history','[]','legacy_urls',NULL)`)
		require.NoError(t, err)
		_, err = databases[1].ExecContext(ctx, `UPDATE ticket_messages SET content='newer message' WHERE id=1; INSERT INTO ticket_messages VALUES(5,1,1,'target only history','[]','legacy_urls',NULL); UPDATE customer_service_retention_policies SET keep_days=700; UPDATE outbox_events SET status='failed' WHERE id=1`)
		require.NoError(t, err)
		report := inspectSyntheticCustomerServiceConversationDependencies(defaultModuleRecoveryRowInspectionLimits())
		plan := report.ConversationRecoveryPlan
		require.Equal(t, 1, plan.RowActionCounts["propose_insert_after_review"])
		require.Equal(t, 1, plan.RowActionCounts["propose_replace_after_review"])
		require.Equal(t, 1, plan.RowActionCounts["preserve_target_only"])
		require.Positive(t, plan.BlockingIssueCounts["conversation_row_content_differs"])
		require.Positive(t, plan.BlockingIssueCounts["target_only_conversation_row"])
		preserved := map[string]CustomerServiceRecoveryPlanTablePreservation{}
		for _, table := range plan.TargetPreservation {
			preserved[table.Table] = table
		}
		require.EqualValues(t, 4, preserved["ticket_messages"].RetainedRecords)
		require.Equal(t, 1, preserved["ticket_messages"].ProposedReplacementRecords)
		for _, table := range baseline.TargetPreservation {
			if table.Table == "customer_service_retention_policies" || table.Table == "outbox_events" {
				require.NotEqual(t, table.ContentFingerprint, preserved[table.Table].ContentFingerprint)
			}
			if table.Table == "orders" || table.Table == "users" {
				require.Equal(t, table.ContentFingerprint, preserved[table.Table].ContentFingerprint)
			}
		}
		for _, action := range plan.RowActions {
			if action.Action == "propose_replace_after_review" || action.Action == "propose_insert_after_review" {
				require.Equal(t, "ticket_messages", action.Table)
			}
		}
	})
	for _, retainedTable := range []string{"ticket_auto_replies", "ticket_messages"} {
		t.Run("cleanup_protects_metadata_in_preserved_"+retainedTable, func(t *testing.T) {
			resetSyntheticConversationDatabasesAndAttachmentFiles()
			statement := `INSERT INTO ticket_auto_replies(id,message_type,metadata,is_active) VALUES(1,'image','{"url":"/uploads/retained-only.png"}',false)`
			if retainedTable == "ticket_messages" {
				statement = `UPDATE ticket_messages SET message_type='image',metadata='{"url":"/uploads/retained-only.png"}' WHERE id=2`
			}
			_, err := databases[1].ExecContext(ctx, statement+`; INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(3,'synthetic-preserved-metadata-cleanup','storage.object_cleanup_requested','media_asset','999','{"object_keys":["retained-only.png"]}','pending')`)
			require.NoError(t, err)
			report := inspectSyntheticCustomerServiceConversationDependencies(defaultModuleRecoveryRowInspectionLimits())
			require.Positive(t, report.DurableEvents.TargetRelatedEvents)
			require.Positive(t, report.IssueCounts["durable_cleanup_overlaps_retained_recovery_object"])
			for _, comparison := range report.Tables {
				require.NotEqual(t, "ticket_auto_replies", comparison.Table)
			}
		})
	}
	t.Run("unchanged_scope_keeps_other_tickets_and_includes_retained_messages_and_derivatives", func(t *testing.T) {
		resetSyntheticConversationDatabasesAndAttachmentFiles()
		report := inspectSyntheticCustomerServiceConversationDependencies(defaultModuleRecoveryRowInspectionLimits())
		require.True(t, report.Complete)
		require.Zero(t, report.BlockingIssues, "%+v", report.IssueCounts)
		require.Equal(t, 3, report.Files.PackageVerifiedFiles)
		require.Equal(t, 3, report.Files.TargetMatchingFiles)
		for _, comparison := range report.Tables {
			switch comparison.Table {
			case "tickets":
				require.Equal(t, 2, comparison.SourceScoped)
			case "ticket_messages":
				require.Equal(t, 3, comparison.SourceScoped)
			case "synthetic_composite_parents":
				require.Equal(t, 1, comparison.Matching)
			case "media_asset_derivatives":
				require.Equal(t, 1, comparison.SourceSelected)
			case "orders", "customer_service_retention_policies", "outbox_events":
				t.Errorf("preserved table entered the row scope: %s", comparison.Table)
			}
		}
	})
	for _, scenario := range []struct {
		name, code string
		change     func()
	}{
		{"target_same_path_different_bytes", "target_attachment_file_content_conflict", func() {
			require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads"], "legacy-attachment.txt"), []byte("newbytes"), 0600))
		}},
		{"target_unreferenced_path_still_conflicts_with_package", "target_attachment_file_content_conflict", func() {
			_, err := databases[1].ExecContext(ctx, `UPDATE ticket_messages SET attachments='[]' WHERE id=4; DELETE FROM media_assets WHERE id=2`)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads"], "legacy-attachment.txt"), []byte("newbytes"), 0600))
		}},
		{"target_same_path_larger_file", "target_attachment_file_content_conflict", func() {
			require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads"], "legacy-attachment.txt"), []byte("newer larger bytes"), 0600))
		}},
		{"target_attachment_missing", "target_attachment_file_missing", func() {
			require.NoError(t, os.Remove(filepath.Join(cfg.StorageRoots["uploads"], "legacy-attachment.txt")))
		}},
		{"package_extracted_file_changed", "verified_package_attachment_changed_after_validation", func() {
			require.NoError(t, os.WriteFile(filepath.Join(verifiedDirectory, "media/uploads/legacy-attachment.txt"), []byte("tampered"), 0600))
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			resetSyntheticConversationDatabasesAndAttachmentFiles()
			scenario.change()
			report := inspectSyntheticCustomerServiceConversationDependencies(defaultModuleRecoveryRowInspectionLimits())
			require.Positive(t, report.IssueCounts[scenario.code])
			if scenario.name == "package_extracted_file_changed" {
				require.False(t, report.Complete)
			}
			if scenario.name == "target_same_path_larger_file" || scenario.name == "target_same_path_different_bytes" {
				require.Positive(t, report.ConversationRecoveryPlan.FileActionCounts["blocked_same_path_content_conflict"])
			}
			if scenario.name == "target_attachment_missing" {
				require.Positive(t, report.ConversationRecoveryPlan.FileActionCounts["propose_add_source_file_after_review"])
			}
		})
	}
	t.Run("row_and_byte_bounds_remain_explicit", func(t *testing.T) {
		for _, limits := range []moduleRecoveryRowInspectionLimits{{Rows: 2, ProjectionBytes: 16 << 20, IssueSamples: 2}, {Rows: 20000, ProjectionBytes: 100, IssueSamples: 2}} {
			resetSyntheticConversationDatabasesAndAttachmentFiles()
			report := inspectSyntheticCustomerServiceConversationDependencies(limits)
			require.False(t, report.Complete)
			require.Positive(t, report.IssueCounts["row_inspection_resource_limit"])
			require.LessOrEqual(t, len(report.Issues), 2)
			require.True(t, report.IssuesTruncated)
		}
	})
	t.Run("projection_over_64_kib_is_not_transferred", func(t *testing.T) {
		resetSyntheticConversationDatabasesAndAttachmentFiles()
		_, err := databases[0].ExecContext(ctx, `UPDATE ticket_messages SET attachments=repeat('x',70000) WHERE id=4`)
		require.NoError(t, err)
		report := inspectSyntheticCustomerServiceConversationDependencies(defaultModuleRecoveryRowInspectionLimits())
		require.False(t, report.Complete)
		require.Positive(t, report.IssueCounts["row_projection_too_large"])
	})
	t.Run("metadata_over_64_kib_does_not_leave_database", func(t *testing.T) {
		resetSyntheticConversationDatabasesAndAttachmentFiles()
		_, err := databases[0].ExecContext(ctx, `UPDATE ticket_messages SET message_type='image',metadata=jsonb_build_object('url',repeat('x',70000))::text WHERE id=3`)
		require.NoError(t, err)
		report := inspectSyntheticCustomerServiceConversationDependencies(defaultModuleRecoveryRowInspectionLimits())
		require.False(t, report.Complete)
		require.Positive(t, report.IssueCounts["row_projection_too_large"])
	})
	if validationDirectory := os.Getenv("LOCAL_BACKUP_MODULE_REPORT_VALIDATION_OUTPUT_DIRECTORY"); validationDirectory != "" {
		resetSyntheticConversationDatabasesAndAttachmentFiles()
		_, err := databases[1].ExecContext(ctx, `UPDATE users SET name='newer referenced customer' WHERE id=1; UPDATE customer_service_inbox_states SET last_read_message_id=2 WHERE id=1`)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads"], "legacy-attachment.txt"), []byte("newbytes"), 0600))
		report := inspectSyntheticCustomerServiceConversationDependencies(defaultModuleRecoveryRowInspectionLimits())
		require.NoError(t, writeDurableJSON(filepath.Join(validationDirectory, "synthetic-customer-service-conversation-row-and-attachment-conflicts-from-postgresql-integration.json"), report))
	}
}
