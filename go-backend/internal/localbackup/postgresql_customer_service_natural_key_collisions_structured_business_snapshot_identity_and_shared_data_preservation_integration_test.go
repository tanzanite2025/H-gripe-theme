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

func TestPostgreSQLCustomerServiceNaturalKeysAndBusinessSnapshotsProtectCurrentSharedDataWithoutWrites(t *testing.T) {
	if os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" || os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL") == "" {
		t.Skip("requires a new disposable PostgreSQL 15 server")
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
		_, err = database.ExecContext(ctx, customerServiceRowInspectionSyntheticSchema+`
ALTER TABLE orders ADD COLUMN order_number text UNIQUE; ALTER TABLE orders ADD COLUMN status text DEFAULT 'paid'; ALTER TABLE orders ADD COLUMN total_minor bigint DEFAULT 1200;
CREATE TABLE products(id bigint PRIMARY KEY,sku text,slug text,name text,price_minor bigint,stock integer);
CREATE TABLE product_variants(id bigint PRIMARY KEY,product_id bigint,sku text,stock integer);
CREATE TABLE order_items(id bigint PRIMARY KEY,order_id bigint,product_id bigint,variant_id bigint,quantity integer,price_minor bigint);
CREATE TABLE faqs(id bigint PRIMARY KEY,page_id text,locale text,question text);
CREATE TABLE selection_assistant_flows(id bigint PRIMARY KEY,slug text UNIQUE,name text);
CREATE TABLE selection_assistant_flow_versions(id bigint PRIMARY KEY,flow_id bigint,version_number integer,config jsonb,status text,UNIQUE(flow_id,version_number));`)
		require.NoError(t, err)
	}
	verifiedDirectory := filepath.Join(t.TempDir(), "verified-customer-service-snapshot-package")
	manifest := &Manifest{}
	for _, path := range []string{"media/uploads_private/customer-service/attachments/originals/synthetic.txt", "media/uploads_private/customer-service/attachments/derivatives/synthetic.txt", "media/uploads/legacy-attachment.txt"} {
		parts := strings.SplitN(strings.TrimPrefix(path, "media/"), "/", 2)
		for _, filename := range []string{filepath.Join(verifiedDirectory, filepath.FromSlash(path)), filepath.Join(cfg.StorageRoots[parts[0]], filepath.FromSlash(parts[1]))} {
			require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
			require.NoError(t, os.WriteFile(filename, []byte("original"), 0600))
		}
		identity, err := fileIdentity(filepath.Join(verifiedDirectory, filepath.FromSlash(path)), path)
		require.NoError(t, err)
		manifest.Files = append(manifest.Files, identity)
	}
	resetDatabases := func() {
		for _, database := range databases {
			_, err := database.ExecContext(ctx, `
DROP INDEX IF EXISTS case_expression_unique,case_partial_unique,case_collation_unique,case_nulls_not_distinct,case_non_unique;
CREATE UNIQUE INDEX IF NOT EXISTS synthetic_ticket_number_unique ON tickets(ticket_number);
`+customerServiceRowInspectionSyntheticRows+`
TRUNCATE products,product_variants,order_items,faqs,selection_assistant_flows,selection_assistant_flow_versions;
UPDATE orders SET user_id=1,order_number='private-order-one'; INSERT INTO orders(id,user_id,order_number) VALUES(2,2,'private-order-two');
INSERT INTO products VALUES(100,'private-product-sku','private-product-slug','private product name',1200,5),(200,'private-second-sku','private-second-slug','second product',500,2),(9007199254740997,'private-large-sku','private-large-slug','large identity',500,2);
INSERT INTO product_variants VALUES(101,100,'private-variant-sku',5),(201,200,'private-second-variant',3);
INSERT INTO order_items VALUES(10,1,100,101,1,1200),(20,2,200,201,1,500);
INSERT INTO faqs VALUES(20,'private-faq-page','en','private historical question');
INSERT INTO selection_assistant_flows VALUES(30,'private-wheelset-flow','private flow name');
INSERT INTO selection_assistant_flow_versions VALUES(31,30,4,'{"nodes":[{"key":"rim"}]}','published');`)
			require.NoError(t, err)
		}
	}
	inspectionLimits := defaultModuleRecoveryRowInspectionLimits()
	inspectReadOnlySnapshots := func() *ModuleRecoveryRowInspection {
		var transactions []*sql.Tx
		var observations []*moduleRecoveryDatabaseObservation
		for _, database := range databases {
			transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
			require.NoError(t, err)
			transactions = append(transactions, transaction)
			defer transaction.Rollback()
			observation, err := inspectModuleRecoveryDatabaseObservation(ctx, transaction)
			require.NoError(t, err)
			observations = append(observations, observation)
		}
		report, err := cfg.inspectCustomerServiceConversationRows(ctx, transactions[0], transactions[1], observations[0], observations[1], manifest, verifiedDirectory, inspectionLimits)
		require.NoError(t, err)
		for index, transaction := range transactions {
			_, fingerprint, _, err := captureDatabaseTransactionInventory(ctx, transaction)
			require.NoError(t, err)
			require.Equal(t, observations[index].Fingerprint, fingerprint, "inspection must preserve all rows and sequences")
			indexes, err := inspectModuleRecoveryDatabaseIndexDefinitions(ctx, transaction)
			require.NoError(t, err)
			require.Equal(t, observations[index].Indexes, indexes)
		}
		encoded, err := json.Marshal(report)
		require.NoError(t, err)
		for _, secret := range []string{"private-order-one", "private-product-sku", "private-product-slug", "private-variant-sku", "private-faq-page", "private-wheelset-flow", "private product name", "private historical question"} {
			require.NotContains(t, string(encoded), secret)
		}
		require.False(t, report.ConversationRecoveryPlan.ExecutionAvailable)
		return report
	}
	for _, scenario := range []struct {
		name, sourceSQL, targetSQL, code string
		incomplete                       bool
	}{
		{"unchanged_registered_natural_keys", "", "", "", false},
		{"source_only_ticket_number_owned_by_preserved_order_ticket", `INSERT INTO tickets(id,category,conversation_id,user_id,customer_user_id,ticket_number) VALUES(4,'customer_service','new-source-conversation',1,1,'shared-natural-number')`, `UPDATE tickets SET ticket_number='shared-natural-number' WHERE id=2`, "conversation_natural_key_owned_by_different_row", false},
		{"source_only_conversation_identity_owned_by_other_category", `INSERT INTO tickets(id,category,conversation_id,user_id,customer_user_id,ticket_number) VALUES(4,'customer_service','reused-conversation',1,1,'new-source-number')`, `UPDATE tickets SET conversation_id='reused-conversation' WHERE id=2`, "conversation_natural_key_owned_by_different_row", false},
		{"inbox_natural_key_moved_to_other_primary_key", `INSERT INTO customer_service_inbox_states VALUES(4,1,2,0)`, `INSERT INTO customer_service_inbox_states VALUES(99,1,2,0)`, "conversation_natural_key_owned_by_different_row", false},
		{"attachment_natural_keys_moved_to_other_primary_key", "", `UPDATE ticket_message_media_attachments SET id=99 WHERE id=1`, "conversation_natural_key_owned_by_different_row", false},
		{"upload_receipt_natural_keys_moved_to_other_primary_key", "", `UPDATE ticket_message_media_attachments SET upload_session_id=NULL; UPDATE media_upload_sessions SET id='different-target-upload'`, "conversation_natural_key_owned_by_different_row", false},
		{"nullable_upload_key_for_public_reuse_does_not_collide", `UPDATE ticket_message_media_attachments SET upload_session_id=NULL,media_asset_id=2`, `UPDATE ticket_message_media_attachments SET upload_session_id=NULL,media_asset_id=2`, "", false},
		{"nullable_conversation_key_does_not_collide", `UPDATE tickets SET conversation_id=NULL WHERE id=3`, `UPDATE tickets SET conversation_id=NULL WHERE id=3`, "", false},
		{"required_natural_key_missing", `UPDATE tickets SET ticket_number=NULL WHERE id=3`, "", "conversation_required_natural_key_missing", false},
		{"required_unique_index_missing", "", `DROP INDEX synthetic_ticket_number_unique`, "conversation_natural_key_index_missing", true},
		{"expression_unique_index_requires_contract", "", `CREATE UNIQUE INDEX case_expression_unique ON tickets(lower(ticket_number))`, "conversation_unique_index_contract_unsupported", true},
		{"partial_unique_index_requires_contract", "", `CREATE UNIQUE INDEX case_partial_unique ON tickets(ticket_number) WHERE deleted_at IS NULL`, "conversation_unique_index_contract_unsupported", true},
		{"alternate_collation_requires_contract", "", `CREATE UNIQUE INDEX case_collation_unique ON tickets(ticket_number COLLATE "C")`, "conversation_unique_index_contract_unsupported", true},
		{"nulls_not_distinct_requires_contract", "", `CREATE UNIQUE INDEX case_nulls_not_distinct ON tickets(conversation_id) NULLS NOT DISTINCT`, "conversation_unique_index_contract_unsupported", true},
		{"additional_non_unique_index_is_preserved", "", `CREATE INDEX case_non_unique ON tickets(category)`, "", false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			resetDatabases()
			for index, statement := range []string{scenario.sourceSQL, scenario.targetSQL} {
				if statement != "" {
					_, err := databases[index].ExecContext(ctx, statement)
					require.NoError(t, err)
				}
			}
			report := inspectReadOnlySnapshots()
			if scenario.code == "" {
				require.Zero(t, report.BlockingIssues, "%+v", report.IssueCounts)
			} else {
				require.Positive(t, report.IssueCounts[scenario.code])
				if scenario.code == "conversation_natural_key_owned_by_different_row" {
					require.Positive(t, report.ConversationRecoveryPlan.RowActionCounts["blocked_reference_contract_conflict"])
				}
			}
			require.Equal(t, !scenario.incomplete, report.Complete, "%+v", report.IssueCounts)
		})
	}
	const orderSnapshot = `{"order_number":"private-order-one","items":[{"id":10,"product_id":100,"variant_id":101,"title":"private captured title","quantity":1,"price_minor":1200}]}`
	for _, scenario := range []struct {
		name, messageType, metadata, sourceSQL, targetSQL, code string
		incomplete                                              bool
	}{
		{"product_and_variant_reference", "product", `{"product_id":100,"variant_id":101}`, "", "", "", false},
		{"large_product_id_keeps_all_digits", "product", `{"product_id":9007199254740997}`, "", "", "", false},
		{"current_product_price_stock_and_name_are_preserved", "product", `{"product_id":100,"variant_id":101}`, "", `UPDATE products SET name='new current title',price_minor=2300,stock=1; UPDATE product_variants SET stock=1`, "", false},
		{"product_identity_changed", "product", `{"product_id":100}`, "", `UPDATE products SET sku='different-target-sku' WHERE id=100`, "business_snapshot_shared_identity_changed", false},
		{"variant_parent_reassigned", "product", `{"product_id":100,"variant_id":101}`, "", `UPDATE product_variants SET product_id=200 WHERE id=101`, "business_snapshot_relationship_identity_conflict", false},
		{"historical_missing_product_retains_static_message", "product", `{"product_id":999999}`, "", "", "historical_business_snapshot_reference_unavailable", false},
		{"historical_product_id_exists_only_in_target", "product", `{"product_id":999999}`, "", `INSERT INTO products VALUES(999999,'new-identity','new-slug','new product',10,1)`, "business_snapshot_source_identity_unverified", false},
		{"required_target_product_missing", "product", `{"product_id":100}`, "", `DELETE FROM products WHERE id=100`, "business_snapshot_target_dependency_missing", false},
		{"fractional_product_id_rejected", "product", `{"product_id":100.5}`, "", "", "business_snapshot_reference_shape_unsupported", true},
		{"noncanonical_string_product_id_rejected", "product", `{"product_id":"0100"}`, "", "", "business_snapshot_reference_shape_unsupported", true},
		{"configuration_variant_reference", "config_confirm", `{"product":{"id":100,"variant_id":101},"selections":{"variant_id":101,"options":[{"key":"color","value":"blue"}]}}`, "", "", "", false},
		{"configuration_variant_alias_conflict", "config_confirm", `{"product":{"id":100,"variant_id":101},"selections":{"variant_id":201}}`, "", "", "business_snapshot_reference_shape_unsupported", true},
		{"order_and_line_references", "order", orderSnapshot, "", "", "", false},
		{"current_order_status_and_amount_are_preserved", "order", orderSnapshot, "", `UPDATE orders SET status='completed',total_minor=1800 WHERE id=1; UPDATE order_items SET price_minor=1800 WHERE id=10`, "", false},
		{"order_customer_reassigned", "order", orderSnapshot, "", `UPDATE orders SET user_id=2 WHERE id=1`, "business_snapshot_shared_identity_changed", false},
		{"visitor_order_customer_cannot_be_inferred_from_placeholder_user", "order", orderSnapshot, `UPDATE tickets SET customer_user_id=NULL,visitor_session_hash='private-visitor-session' WHERE id=1`, "", "business_snapshot_order_customer_unverified", false},
		{"order_line_reassigned", "order", orderSnapshot, "", `UPDATE order_items SET order_id=2 WHERE id=10`, "business_snapshot_relationship_identity_conflict", false},
		{"order_line_product_reassigned", "order", orderSnapshot, "", `UPDATE order_items SET product_id=200 WHERE id=10`, "business_snapshot_relationship_identity_conflict", false},
		{"order_line_without_order_number", "order", `{"items":[{"id":10,"product_id":100}]}`, "", "", "business_snapshot_order_identity_missing", false},
		{"order_line_thumbnail_is_included", "order", `{"order_number":"private-order-one","items":[{"id":10,"product_id":100,"variant_id":101,"thumbnail":"/uploads/legacy-attachment.txt"}]}`, "", "", "", false},
		{"order_line_unknown_thumbnail_blocks", "order", `{"order_number":"private-order-one","items":[{"id":10,"thumbnail":"/uploads/missing-snapshot-file.png"}]}`, "", "", "legacy_attachment_registration_missing", true},
		{"faq_reference", "faq", `{"faq_id":20,"page_id":"private-faq-page","locale":"en"}`, "", "", "", false},
		{"faq_question_update_keeps_history", "faq", `{"faq_id":20,"page_id":"private-faq-page","locale":"en"}`, "", `UPDATE faqs SET question='new current question' WHERE id=20`, "", false},
		{"faq_page_reassigned", "faq", `{"faq_id":20,"page_id":"private-faq-page","locale":"en"}`, "", `UPDATE faqs SET page_id='another-page' WHERE id=20`, "business_snapshot_relationship_identity_conflict", false},
		{"wheelset_flow_uses_version_number", "wheelset_selection_request", `{"kind":"wheelset_selection_request","version":1,"flow_slug":"private-wheelset-flow","flow_version":4}`, "", "", "", false},
		{"wheelset_archived_version_keeps_snapshot", "wheelset_selection_request", `{"version":1,"flow_slug":"private-wheelset-flow","flow_version":4}`, "", `UPDATE selection_assistant_flow_versions SET status='archived' WHERE id=31`, "", false},
		{"wheelset_same_version_changed_configuration", "wheelset_selection_request", `{"version":1,"flow_slug":"private-wheelset-flow","flow_version":4}`, "", `UPDATE selection_assistant_flow_versions SET config='{"nodes":[]}' WHERE id=31`, "business_snapshot_shared_identity_changed", false},
		{"wheelset_unknown_payload_version", "wheelset_selection_request", `{"version":2,"flow_slug":"private-wheelset-flow","flow_version":4}`, "", "", "business_snapshot_reference_shape_unsupported", true},
		{"wheelset_zero_payload_version_is_not_an_absent_version", "wheelset_selection_request", `{"version":0,"flow_slug":"private-wheelset-flow","flow_version":4}`, "", "", "business_snapshot_reference_shape_unsupported", true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			resetDatabases()
			for _, database := range databases {
				_, err := database.ExecContext(ctx, `UPDATE ticket_messages SET message_type=$1,metadata=$2 WHERE id=1`, scenario.messageType, scenario.metadata)
				require.NoError(t, err)
			}
			for index, statement := range []string{scenario.sourceSQL, scenario.targetSQL} {
				if statement != "" {
					_, err := databases[index].ExecContext(ctx, statement)
					require.NoError(t, err)
				}
			}
			report := inspectReadOnlySnapshots()
			if scenario.code == "" {
				require.Zero(t, report.BlockingIssues, "%+v", report.IssueCounts)
			} else {
				require.Positive(t, report.IssueCounts[scenario.code], "%+v", report.IssueCounts)
			}
			require.Equal(t, !scenario.incomplete, report.Complete, "%+v", report.IssueCounts)
			for _, action := range report.ConversationRecoveryPlan.RowActions {
				require.NotContains(t, []string{"orders", "order_items", "products", "product_variants", "faqs", "selection_assistant_flows", "selection_assistant_flow_versions"}, action.Table, "shared business references must not enter replacement graph")
			}
		})
	}
	t.Run("preserved_order_line_thumbnail_blocks_cleanup", func(t *testing.T) {
		resetDatabases()
		_, err := databases[1].ExecContext(ctx, `UPDATE ticket_messages SET message_type='order',metadata='{"order_number":"private-order-two","items":[{"id":20,"product_id":200,"variant_id":201,"thumbnail":"/uploads/retained-order-line.png"}]}' WHERE id=2; INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(3,'private-order-line-cleanup','storage.object_cleanup_requested','media_asset','999','{"resource_type":"media_asset","resource_id":"999","object_keys":["retained-order-line.png"]}','pending')`)
		require.NoError(t, err)
		report := inspectReadOnlySnapshots()
		require.Positive(t, report.IssueCounts["durable_cleanup_overlaps_retained_recovery_object"])
	})
	t.Run("disabled_auto_reply_business_reference_is_still_checked", func(t *testing.T) {
		resetDatabases()
		for _, database := range databases {
			_, err := database.ExecContext(ctx, `INSERT INTO ticket_auto_replies(id,message_type,metadata,is_active) VALUES(1,'product','{"product_id":100,"variant_id":101}',false)`)
			require.NoError(t, err)
		}
		_, err := databases[1].ExecContext(ctx, `UPDATE product_variants SET product_id=200 WHERE id=101`)
		require.NoError(t, err)
		report := inspectReadOnlySnapshots()
		require.Positive(t, report.IssueCounts["business_snapshot_relationship_identity_conflict"])
		for _, action := range report.ConversationRecoveryPlan.RowActions {
			require.NotEqual(t, "ticket_auto_replies", action.Table)
		}
	})
	t.Run("missing_business_identity_column_is_incomplete", func(t *testing.T) {
		resetDatabases()
		_, err := databases[0].ExecContext(ctx, `UPDATE ticket_messages SET message_type='product',metadata='{"product_id":100}' WHERE id=1`)
		require.NoError(t, err)
		_, err = databases[1].ExecContext(ctx, `ALTER TABLE products RENAME COLUMN sku TO unavailable_snapshot_identity`)
		require.NoError(t, err)
		defer func() {
			_, err := databases[1].ExecContext(ctx, `ALTER TABLE products RENAME COLUMN unavailable_snapshot_identity TO sku`)
			require.NoError(t, err)
		}()
		report := inspectReadOnlySnapshots()
		require.False(t, report.Complete)
		require.False(t, report.ConversationReferenceChecks.Complete)
		require.Positive(t, report.IssueCounts["business_snapshot_reference_schema_missing"])
	})
	t.Run("ambiguous_order_number_cannot_select_an_arbitrary_order", func(t *testing.T) {
		resetDatabases()
		_, err := databases[0].ExecContext(ctx, `UPDATE ticket_messages SET message_type='order',metadata=$1 WHERE id=1`, orderSnapshot)
		require.NoError(t, err)
		_, err = databases[1].ExecContext(ctx, `ALTER TABLE orders DROP CONSTRAINT orders_order_number_key; UPDATE orders SET order_number='private-order-one' WHERE id=2`)
		require.NoError(t, err)
		defer func() {
			_, err := databases[1].ExecContext(ctx, `UPDATE orders SET order_number='private-order-two' WHERE id=2; ALTER TABLE orders ADD CONSTRAINT orders_order_number_key UNIQUE(order_number)`)
			require.NoError(t, err)
		}()
		report := inspectReadOnlySnapshots()
		require.Positive(t, report.IssueCounts["business_snapshot_natural_identity_ambiguous"])
		require.Positive(t, report.ConversationRecoveryPlan.RowActionCounts["blocked_reference_contract_conflict"])
	})
	t.Run("business_snapshot_line_reference_limit_is_explicit", func(t *testing.T) {
		resetDatabases()
		inspectionLimits.Rows = 100
		defer func() { inspectionLimits = defaultModuleRecoveryRowInspectionLimits() }()
		metadata := `{"order_number":"private-order-one","items":[` + strings.Repeat(`{},`, 100) + `{}]}`
		_, err := databases[0].ExecContext(ctx, `UPDATE ticket_messages SET message_type='order',metadata=$1 WHERE id=1`, metadata)
		require.NoError(t, err)
		report := inspectReadOnlySnapshots()
		require.False(t, report.Complete)
		require.Positive(t, report.IssueCounts["business_snapshot_reference_resource_limit"])
	})
	if outputDirectory := os.Getenv("LOCAL_BACKUP_MODULE_REPORT_VALIDATION_OUTPUT_DIRECTORY"); outputDirectory != "" {
		resetDatabases()
		for _, database := range databases {
			_, err := database.ExecContext(ctx, `UPDATE ticket_messages SET message_type='order',metadata=$1 WHERE id=1`, orderSnapshot)
			require.NoError(t, err)
		}
		_, err := databases[1].ExecContext(ctx, `UPDATE orders SET user_id=2 WHERE id=1`)
		require.NoError(t, err)
		require.NoError(t, writeDurableJSON(filepath.Join(outputDirectory, "synthetic-customer-service-natural-key-and-business-snapshot-identity-conflicts.json"), inspectReadOnlySnapshots()))
	}
}
