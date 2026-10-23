package localbackup

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func verifyProductCatalogRequiredImagesSharedRegistrationsAndRetainedCleanupReceiptAtomicRollback(t *testing.T, manager *Manager, stage *sql.DB, snapshot productCatalogModuleSnapshot, packageData []byte) {
	t.Helper()
	ctx, cfg := context.Background(), manager.Configuration
	var assetID int64
	require.NoError(t, stage.QueryRow(`SELECT id FROM media_assets WHERE storage_key='products/product-original.png'`).Scan(&assetID))
	basePayload := fmt.Sprintf(`{"resource_type":"media_asset","resource_id":"%d","object_keys":["products/product-original.png"]}`, assetID)
	for _, scenario := range []struct {
		name, payload, eventType, aggregateType, aggregateID, status, extraSQL, expectedError string
		expectedCleanupInspectionError, isolatedRetirementInspectionFixtureSQL                string
	}{
		{name: "pending_original", payload: basePayload, expectedError: "can delete a required product image"},
		{name: "processed_original_can_reconcile", payload: basePayload, status: "processed", expectedError: "can delete a required product image"},
		{name: "dead_letter_original_can_reconcile", payload: basePayload, status: "dead_letter", expectedError: "can delete a required product image"},
		{name: "failed_original_can_retry", payload: basePayload, status: "failed", expectedError: "can delete a required product image"},
		{name: "ignored_original_is_preserved", payload: basePayload, status: "ignored"},
		{name: "deleted_original_is_preserved", payload: basePayload, extraSQL: `UPDATE outbox_events SET deleted_at=now() WHERE id=900000`},
		{name: "ignored_but_leased_is_blocked", payload: basePayload, status: "ignored", extraSQL: `UPDATE outbox_events SET locked_by='synthetic-cleanup-worker' WHERE id=900000`, expectedError: "execution lease"},
		{name: "unknown_result_is_blocked", payload: basePayload, status: "unknown", expectedError: "unresolved execution state"},
		{name: "unknown_payload_field_is_blocked", payload: fmt.Sprintf(`{"resource_type":"media_asset","resource_id":"%d","object_keys":["disjoint/photo.png"],"future_prefix":"products/"}`, assetID), expectedError: "unknown fields"},
		{name: "unknown_resource_is_blocked", payload: `{"resource_type":"future_media_namespace","resource_id":"x","object_keys":["disjoint/photo.png"]}`, expectedError: "resource contract"},
		{name: "unknown_event_is_blocked", payload: `{}`, eventType: "storage.future_cleanup_requested", expectedError: "event type"},
		{name: "noncanonical_resource_is_blocked", payload: fmt.Sprintf(`{"resource_type":"media_asset","resource_id":"0%d","object_keys":["disjoint/photo.png"]}`, assetID), expectedError: "identity is invalid"},
		{name: "empty_original_keys_are_blocked", payload: fmt.Sprintf(`{"resource_type":"media_asset","resource_id":"%d"}`, assetID), expectedError: "exact object keys"},
		{name: "original_resource_hits_registered_derivative_even_if_key_list_is_old", payload: fmt.Sprintf(`{"resource_type":"media_asset","resource_id":"%d","object_keys":["disjoint/photo.png"]}`, assetID), expectedError: "can delete a required product image"},
		{name: "noncanonical_key_is_blocked", payload: `{"resource_type":"unregistered_media_upload","resource_id":"synthetic","object_keys":["products\\product-original.png"]}`, expectedError: "noncanonical"},
		{name: "same_filename_in_different_directory_is_safe", payload: `{"resource_type":"unregistered_media_upload","resource_id":"synthetic","object_keys":["disjoint/product-original.png"]}`},
		{name: "required_thumbnail_key_is_blocked", payload: `{"resource_type":"unregistered_media_upload","resource_id":"synthetic","object_keys":["products/product-small.png"]}`, expectedError: "can delete a required product image"},
		{name: "failed_derivative_wrong_parent_namespace_is_blocked", payload: fmt.Sprintf(`{"resource_type":"unregistered_media_derivative","resource_id":"%d","object_keys":["media-derivatives/900009/original.png"]}`, assetID), expectedError: "parent namespace"},
		{name: "retired_derivative_wrong_aggregate_parent_is_blocked", payload: `{"resource_type":"retired_media_derivative","resource_id":"900002","object_keys":["disjoint/retired.png"]}`, extraSQL: `INSERT INTO media_assets(id,filename,url,storage_key) VALUES(900001,'unrelated.png','/uploads/disjoint/unrelated.png','disjoint/unrelated.png'); INSERT INTO media_asset_derivatives(id,media_asset_id,preset,url,storage_key,deleted_at) VALUES(900002,900001,'retired','/uploads/disjoint/retired.png','disjoint/retired.png',now())`, expectedError: "registration or parent"},
		{name: "retired_derivative_missing_registration_does_not_become_asset_id", payload: `{"resource_type":"retired_media_derivative","resource_id":"900002","object_keys":["disjoint/retired.png"]}`},
		{name: "retired_derivative_valid_disjoint_registration_is_preserved", payload: `{"resource_type":"retired_media_derivative","resource_id":"900002","object_keys":["media-derivatives/900001/retired.png"]}`, aggregateID: "900001", extraSQL: `INSERT INTO media_assets(id,filename,url,storage_key) VALUES(900001,'unrelated.png','/uploads/disjoint/unrelated.png','disjoint/unrelated.png'); INSERT INTO media_asset_derivatives(id,media_asset_id,preset,url,storage_key,deleted_at) VALUES(900002,900001,'retired','/uploads/media-derivatives/900001/retired.png','media-derivatives/900001/retired.png',now())`},
		{name: "retained_product_customer_upload_cannot_become_public_catalog_media", payload: `{"resource_type":"unregistered_media_upload","resource_id":"synthetic","object_keys":["disjoint/photo.png"]}`, extraSQL: `INSERT INTO media_assets(id,filename,url,storage_key,uploader_actor_type,uploader_id,library_partition) VALUES(900001,'customer.png','/uploads/products/customer.png','products/customer.png','customer',(SELECT min(id) FROM users),'customer_submissions'); UPDATE products SET description='<img src="/uploads/products/customer.png">' WHERE slug='target-only-product'`, expectedError: "not eligible for public business reuse"},
		{name: "required_derivative_namespace_cannot_claim_another_parent", payload: `{"resource_type":"unregistered_media_upload","resource_id":"synthetic","object_keys":["disjoint/photo.png"]}`, extraSQL: fmt.Sprintf(`INSERT INTO media_asset_derivatives(id,media_asset_id,preset,url,storage_key) VALUES(900002,%d,'wrong-parent','/uploads/media-derivatives/900009/wrong-parent.png','media-derivatives/900009/wrong-parent.png')`, assetID), expectedError: "namespace disagrees with its parent"},
		{name: "retained_product_media_active_write_lease_is_blocked", payload: `{"resource_type":"unregistered_media_upload","resource_id":"synthetic","object_keys":["disjoint/photo.png"]}`, extraSQL: `INSERT INTO media_assets(id,filename,url,storage_key,customer_service_derivative_write_token,customer_service_derivative_write_expires_at) VALUES(900001,'leased.png','/uploads/products/leased.png','products/leased.png','synthetic-active-write-lease',now()+interval '1 hour'); UPDATE products SET description='<img src="/uploads/products/leased.png">' WHERE slug='target-only-product'`, expectedError: "active write lease"},
		{name: "retained_product_only_reference_is_also_protected", payload: `{"resource_type":"unregistered_media_upload","resource_id":"synthetic","object_keys":["products/target-only.png"]}`, extraSQL: `UPDATE products SET description='<img src="/uploads/products/target-only.png">' WHERE slug='target-only-product'`, expectedError: "can delete a required product image"},
		{name: "legacy_trusted_url_cleanup_is_blocked", payload: `{"ticket_id":99,"attachment_references":["/uploads/products/product-original.png?historical=1"]}`, eventType: "customer_service.retention_cleanup_requested", aggregateType: "customer_service_conversation", aggregateID: "99", expectedError: "can delete a required product image"},
		{name: "foreign_url_cannot_match_by_filename", payload: `{"ticket_id":99,"attachment_references":["https://foreign.invalid/uploads/products/product-original.png"]}`, eventType: "customer_service.retention_cleanup_requested", aggregateType: "customer_service_conversation", aggregateID: "99", expectedError: "cannot be resolved locally"},
		{name: "active_duplicate_original_registration_is_blocked", payload: `{"resource_type":"unregistered_media_upload","resource_id":"synthetic","object_keys":["disjoint/photo.png"]}`, extraSQL: `INSERT INTO media_assets(id,filename,url,storage_key) VALUES(900001,'duplicate.png','/uploads/products/product-original.png?other-registration=1','products/product-original.png')`, expectedError: "multiple shared object registrations"},
		{name: "soft_deleted_original_still_blocks_reuse", payload: `{"resource_type":"unregistered_media_upload","resource_id":"synthetic","object_keys":["disjoint/photo.png"]}`, extraSQL: `INSERT INTO media_assets(id,filename,url,storage_key,deleted_at) VALUES(900001,'retired.png','/uploads/products/product-original.png?retired=1','products/product-original.png',now())`, expectedError: "multiple shared object registrations", expectedCleanupInspectionError: "retired registration", isolatedRetirementInspectionFixtureSQL: fmt.Sprintf(`DELETE FROM media_assets WHERE id=900001; UPDATE media_assets SET deleted_at=now() WHERE id=%d`, assetID)},
		{name: "retired_derivative_same_key_blocks_original_reuse", payload: `{"resource_type":"unregistered_media_upload","resource_id":"synthetic","object_keys":["disjoint/photo.png"]}`, extraSQL: fmt.Sprintf(`INSERT INTO media_asset_derivatives(id,media_asset_id,preset,url,storage_key,deleted_at) VALUES(900002,%d,'retired-overlap','/uploads/products/product-original.png?retired-derivative=1','products/product-original.png',now())`, assetID), expectedError: "multiple shared object registrations", expectedCleanupInspectionError: "retired registration", isolatedRetirementInspectionFixtureSQL: fmt.Sprintf(`DELETE FROM media_asset_derivatives WHERE id=900002; UPDATE media_asset_derivatives SET deleted_at=now() WHERE media_asset_id=%d AND preset='small'`, assetID)},
	} {
		t.Run("product_cleanup_transaction_"+scenario.name, func(t *testing.T) {
			if scenario.eventType == "" {
				scenario.eventType = "storage.object_cleanup_requested"
			}
			if scenario.aggregateType == "" {
				scenario.aggregateType = "media_asset"
			}
			if scenario.aggregateID == "" {
				scenario.aggregateID = fmt.Sprint(assetID)
			}
			if scenario.status == "" {
				scenario.status = "pending"
			}
			_, err := stage.ExecContext(ctx, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(900000,'synthetic-product-cleanup-guard',$1,$2,$3,$4::jsonb,$5)`, scenario.eventType, scenario.aggregateType, scenario.aggregateID, scenario.payload, scenario.status)
			require.NoError(t, err)
			t.Cleanup(func() {
				_, err := stage.ExecContext(ctx, `DELETE FROM outbox_events WHERE id=900000; DELETE FROM media_asset_derivatives WHERE id=900002; DELETE FROM media_assets WHERE id=900001; UPDATE products SET name='Original product' WHERE slug='product-backup-original' AND locale='en'; UPDATE products SET description=NULL WHERE slug='target-only-product'`)
				require.NoError(t, err)
			})
			if scenario.extraSQL != "" {
				_, err = stage.ExecContext(ctx, scenario.extraSQL)
				require.NoError(t, err)
			}
			_, err = stage.ExecContext(ctx, `UPDATE products SET name='Marker before rejected merge' WHERE slug='product-backup-original' AND locale='en'`)
			require.NoError(t, err)
			_, before, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			var originalReceipt string
			require.NoError(t, stage.QueryRow(`SELECT to_jsonb(r)::text FROM outbox_events r WHERE id=900000`).Scan(&originalReceipt))
			if scenario.expectedCleanupInspectionError != "" {
				// Isolate retirement on the existing required original/thumbnail
				// inside a rolled-back fixture transaction. Then the untouched
				// overlap fixture must fail the full merge's duplicate check too.
				inspection, err := stage.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
				require.NoError(t, err)
				defer inspection.Rollback()
				_, err = inspection.ExecContext(ctx, scenario.isolatedRetirementInspectionFixtureSQL)
				require.NoError(t, err)
				protection, inspectionErr := cfg.inspectProductCatalogRequiredImagesAndRetainedCleanupReceipts(ctx, inspection)
				require.NoError(t, inspection.Rollback())
				require.ErrorContains(t, inspectionErr, scenario.expectedCleanupInspectionError)
				require.Nil(t, protection)
			}
			_, _, result, err := cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, snapshot)
			if scenario.expectedError != "" {
				require.ErrorContains(t, err, scenario.expectedError)
				require.Nil(t, result)
				_, after, _, err := captureDatabaseInventory(ctx, stage)
				require.NoError(t, err)
				require.Equal(t, before, after, "rejected cleanup/shared ownership must roll back prior product writes and preserve all sequences")
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
			var retainedReceipt string
			require.NoError(t, stage.QueryRow(`SELECT to_jsonb(r)::text FROM outbox_events r WHERE id=900000`).Scan(&retainedReceipt))
			require.Equal(t, originalReceipt, retainedReceipt, "never reset, consume, suppress or rewrite the original receipt")
		})
	}
	for _, protectionChange := range []string{"missing", "changed"} {
		t.Run("product_cleanup_stopped_binding_"+protectionChange, func(t *testing.T) {
			task := uploadProductCatalogPackageForIntegrationTest(t, manager, packageData)
			require.Equal(t, "product_module_isolated_verified", task.State, task.Error)
			require.NotNil(t, task.ProductCatalogPreviewBinding.RequiredImageCleanupProtection)
			if protectionChange == "missing" {
				task.ProductCatalogPreviewBinding.RequiredImageCleanupProtection = nil
			} else {
				task.ProductCatalogPreviewBinding.RequiredImageCleanupProtection.CleanupRowsSHA256 = "changed"
			}
			require.NoError(t, manager.saveTask(task))
			err := manager.rehearseProductCatalogPairedSwitchAndVerifiedRollback(ctx, task.ID, nil)
			require.ErrorContains(t, err, "cleanup protection is missing or changed")
			task, err = manager.GetTask(task.ID)
			require.NoError(t, err)
			require.Nil(t, task.Journal, "reject before any rename or connection sealing")
		})
	}
	for _, databaseSide := range []string{"target", "candidate"} {
		t.Run("product_cleanup_stopped_rejects_new_"+databaseSide+"_receipt", func(t *testing.T) {
			task := uploadProductCatalogPackageForIntegrationTest(t, manager, packageData)
			require.Equal(t, "product_module_isolated_verified", task.State, task.Error)
			databaseName := task.StageDatabase
			if databaseSide == "target" {
				databaseName = cfg.DatabaseName
			}
			db, err := cfg.connectDatabase(ctx, databaseName)
			require.NoError(t, err)
			defer db.Close()
			_, err = db.Exec(`INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(900000,'synthetic-new-cleanup-after-product-preview','storage.object_cleanup_requested','media_asset',$1,$2::jsonb,'processed')`, fmt.Sprint(assetID), basePayload)
			require.NoError(t, err)
			defer func() { _, err := db.Exec(`DELETE FROM outbox_events WHERE id=900000`); require.NoError(t, err) }()
			err = manager.rehearseProductCatalogPairedSwitchAndVerifiedRollback(ctx, task.ID, nil)
			require.ErrorContains(t, err, "changed after preview")
			task, err = manager.GetTask(task.ID)
			require.NoError(t, err)
			require.Nil(t, task.Journal)
		})
	}
	t.Run("product_cleanup_encrypted_upload_rejects_retained_pending_cleanup", func(t *testing.T) {
		live, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
		require.NoError(t, err)
		defer live.Close()
		_, err = live.ExecContext(ctx, `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(900000,'synthetic-product-cleanup-before-http-upload','storage.object_cleanup_requested','media_asset',$1,$2::jsonb,'processed')`, fmt.Sprint(assetID), basePayload)
		require.NoError(t, err)
		defer func() {
			_, err := live.ExecContext(ctx, `DELETE FROM outbox_events WHERE id=900000`)
			require.NoError(t, err)
		}()
		_, before, _, err := captureDatabaseInventory(ctx, live)
		require.NoError(t, err)
		mediaBefore, err := cfg.mediaFingerprint(ctx)
		require.NoError(t, err)
		task := uploadProductCatalogPackageForIntegrationTest(t, manager, packageData)
		require.Equal(t, "failed", task.State)
		require.Contains(t, task.Error, "can delete a required product image")
		_, after, _, err := captureDatabaseInventory(ctx, live)
		require.NoError(t, err)
		require.Equal(t, before, after)
		mediaAfter, err := cfg.mediaFingerprint(ctx)
		require.NoError(t, err)
		require.Equal(t, mediaBefore, mediaAfter)
		var stageExists bool
		require.NoError(t, live.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)`, task.StageDatabase).Scan(&stageExists))
		require.False(t, stageExists, "a failed upload must remove only its disposable candidate")
	})
	t.Run("product_cleanup_ignored_receipt_survives_encrypted_upload_paired_switch_and_rollback", func(t *testing.T) {
		live, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
		require.NoError(t, err)
		_, err = live.Exec(`INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(900000,'synthetic-ignored-cleanup-through-product-pair','storage.object_cleanup_requested','media_asset',$1,$2::jsonb,'ignored')`, fmt.Sprint(assetID), basePayload)
		require.NoError(t, err)
		var originalReceipt string
		require.NoError(t, live.QueryRow(`SELECT to_jsonb(r)::text FROM outbox_events r WHERE id=900000`).Scan(&originalReceipt))
		require.NoError(t, live.Close())
		defer func() {
			live, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			defer live.Close()
			_, err = live.Exec(`DELETE FROM outbox_events WHERE id=900000`)
			require.NoError(t, err)
		}()
		task := uploadProductCatalogPackageForIntegrationTest(t, manager, packageData)
		require.Equal(t, "product_module_isolated_verified", task.State, task.Error)
		require.Equal(t, 1, task.ProductCatalogPreviewBinding.RequiredImageCleanupProtection.RetainedCleanupReceipts)
		candidate, err := cfg.connectDatabase(ctx, task.StageDatabase)
		require.NoError(t, err)
		var candidateReceipt string
		require.NoError(t, candidate.QueryRow(`SELECT to_jsonb(r)::text FROM outbox_events r WHERE id=900000`).Scan(&candidateReceipt))
		require.NoError(t, candidate.Close())
		require.Equal(t, originalReceipt, candidateReceipt)
		require.NoError(t, manager.rehearseProductCatalogPairedSwitchAndVerifiedRollback(ctx, task.ID, nil))
		task, err = manager.GetTask(task.ID)
		require.NoError(t, err)
		require.Equal(t, "rolled_back", task.State, task.Error)
		requireProductCatalogOriginalDatabaseAndCompleteMediaRestored(t, manager, task)
		restored, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
		require.NoError(t, err)
		var retainedReceipt string
		require.NoError(t, restored.QueryRow(`SELECT to_jsonb(r)::text FROM outbox_events r WHERE id=900000`).Scan(&retainedReceipt))
		require.NoError(t, restored.Close())
		require.Equal(t, originalReceipt, retainedReceipt)
	})
}
