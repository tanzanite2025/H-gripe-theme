package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func verifyProductCatalogMediaPlacementOwnersAssetBindingsAndLegacyURLAtomicRollback(t *testing.T, manager *Manager, stage *sql.DB, snapshot productCatalogModuleSnapshot, packageData []byte) {
	t.Helper()
	ctx, cfg := context.Background(), manager.Configuration
	var mediaID, productID, assetID, otherProductID, otherVariantID, templateDefinitionID int64
	require.NoError(t, stage.QueryRowContext(context.Background(), `SELECT id,product_id,media_asset_id FROM product_media WHERE url='/uploads/products/product-original.png'`).Scan(&mediaID, &productID, &assetID))
	require.NoError(t, stage.QueryRowContext(context.Background(), `SELECT id FROM products WHERE slug='target-only-product'`).Scan(&otherProductID))
	require.NoError(t, stage.QueryRowContext(context.Background(), `SELECT id FROM product_variants WHERE sku='target-only-translated-sku'`).Scan(&otherVariantID))
	templateDefinitionID = 910011
	_, err := stage.ExecContext(context.Background(), `INSERT INTO product_specification_templates(id,name,slug) VALUES(910010,'Synthetic media placement template','synthetic-media-placement-template'); INSERT INTO product_spec_definitions(id,product_specification_template_id,name,slug) VALUES(910011,910010,'Synthetic media placement definition','synthetic-media-placement-definition')`)
	require.NoError(t, err)
	defer func() {
		_, err := stage.ExecContext(context.Background(), `DELETE FROM product_spec_definitions WHERE id=910011; DELETE FROM product_specification_templates WHERE id=910010`)
		require.NoError(t, err)
	}()
	snapshot = copyProductCatalogSnapshotWithMatchingRowJSONFieldsChanged(t, snapshot, "", "", "")
	for _, table := range []struct{ name, predicate string }{
		{"product_specification_templates", "id=910010"}, {"product_spec_definitions", "id=910011"},
		{"products", fmt.Sprintf("id=%d", otherProductID)}, {"product_variants", fmt.Sprintf("id=%d", otherVariantID)},
	} {
		tx, err := stage.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
		require.NoError(t, err)
		rows, err := readProductCatalogRows(ctx, tx, table.name, table.predicate)
		require.NoError(t, err)
		require.NoError(t, tx.Rollback())
		selected := snapshot.Tables[table.name]
		selected.Rows = append(selected.Rows, rows...)
		snapshot.Tables[table.name] = selected
	}
	var originalMedia string
	require.NoError(t, stage.QueryRowContext(context.Background(), `SELECT to_jsonb(m)::text FROM product_media m WHERE id=$1`, mediaID).Scan(&originalMedia))
	for _, scenario := range []struct {
		name, mutationJSON, setupSQL, expectedError string
		newPlacement                                bool
	}{
		{name: "sku_belongs_to_another_product", mutationJSON: fmt.Sprintf(`{"variant_id":%d}`, otherVariantID), expectedError: "belonging to another product"},
		{name: "option_belongs_to_another_product", setupSQL: fmt.Sprintf(`INSERT INTO product_variant_option_values(id,product_id,spec_definition_id,value_key,label) VALUES(910002,%d,%d,'synthetic-placement-option','Synthetic placement option')`, otherProductID, templateDefinitionID), mutationJSON: `{"variant_option_value_id":910002}`, expectedError: "belonging to another product"},
		{name: "existing_media_cannot_move_product", mutationJSON: fmt.Sprintf(`{"product_id":%d,"variant_id":%d}`, otherProductID, otherVariantID), expectedError: "cannot move to another product"},
		{name: "asset_id_and_original_url_disagree", setupSQL: `INSERT INTO media_assets(id,filename,url,storage_key) VALUES(910001,'other.png','/uploads/products/other.png','products/other.png')`, mutationJSON: `{"media_asset_id":910001}`, expectedError: "asset ID and URL identify different files"},
		{name: "static_url_cannot_claim_registered_asset", mutationJSON: `{"url":"/company/products/bundled-product-photo.png"}`, expectedError: "asset ID and URL identify different files"},
		{name: "missing_main_url", mutationJSON: `{"url":""}`, expectedError: "requires a persisted URL"},
		{name: "main_media_type_disagrees", mutationJSON: `{"media_type":"video"}`, expectedError: "type does not match"},
		{name: "poster_must_be_an_image", setupSQL: `INSERT INTO media_assets(id,filename,url,storage_key,media_type) VALUES(910001,'video.mp4','/uploads/products/video.mp4','products/video.mp4','video')`, mutationJSON: `{"poster_url":"/uploads/products/video.mp4"}`, expectedError: "type does not match"},
		{name: "thumbnail_must_be_an_image", setupSQL: `INSERT INTO media_assets(id,filename,url,storage_key,media_type) VALUES(910001,'video.mp4','/uploads/products/video.mp4','products/video.mp4','video')`, mutationJSON: `{"thumbnail_url":"/uploads/products/video.mp4"}`, expectedError: "type does not match"},
		{name: "generated_variant_wrong_parent", setupSQL: `INSERT INTO media_assets(id,filename,url,storage_key) VALUES(910001,'other.png','/uploads/products/other.png','products/other.png')`, mutationJSON: `{"image_variants":{"small":{"url":"/uploads/products/other.png"}}}`, expectedError: "does not belong to its parent"},
		{name: "generated_variant_unknown_field", mutationJSON: `{"image_variants":{"small":{"url":"/uploads/products/product-small.png","future_reference":"/uploads/other.png"}}}`, expectedError: "registered object format"},
		{name: "generated_variant_wrong_shape", mutationJSON: `{"image_variants":[]}`, expectedError: "registered object format"},
		{name: "generated_variant_null_shape", mutationJSON: `{"image_variants":null}`, expectedError: "row restore failed in product_media"},
		{name: "generated_variant_missing_url", mutationJSON: `{"image_variants":{"small":{"width":1}}}`, expectedError: "does not belong to its parent"},
		{name: "generated_variant_empty_preset", mutationJSON: `{"image_variants":{"":{"url":"/uploads/products/product-small.png"}}}`, expectedError: "registered parent image"},
		{name: "generated_variant_negative_dimensions", mutationJSON: `{"image_variants":{"small":{"url":"/uploads/products/product-small.png","width":-1}}}`, expectedError: "registered parent image"},
		{name: "new_unregistered_main_is_rejected", newPlacement: true, mutationJSON: `{"media_asset_id":null,"url":"/uploads/products/new-unregistered.png"}`, expectedError: "exact existing placement and field"},
		{name: "changed_unregistered_main_is_rejected", mutationJSON: `{"media_asset_id":null,"url":"/uploads/products/new-unregistered.png"}`, expectedError: "exact existing placement and field"},
		{name: "copying_legacy_main_into_thumbnail_is_rejected", setupSQL: fmt.Sprintf(`UPDATE product_media SET url='/uploads/products/legacy-unregistered.png',media_asset_id=NULL WHERE id=%d`, mediaID), mutationJSON: `{"thumbnail_url":"/uploads/products/legacy-unregistered.png"}`, expectedError: "exact existing placement and field"},
		{name: "valid_same_parent_generated_variant", mutationJSON: `{"image_variants":{"small":{"url":"/uploads/products/product-small.png","width":1,"height":1,"mime_type":"image/webp"}}}`},
		{name: "valid_derivative_main_with_parent_asset", mutationJSON: `{"url":"/uploads/products/product-small.png"}`},
		{name: "valid_url_only_registered_main", mutationJSON: `{"media_asset_id":null}`},
		{name: "valid_url_only_registered_derivative_main", mutationJSON: `{"media_asset_id":null,"url":"/uploads/products/product-small.png"}`},
		{name: "url_only_derivative_cannot_hide_wrong_media_type", mutationJSON: `{"media_asset_id":null,"url":"/uploads/products/product-small.png","media_type":"video"}`, expectedError: "type does not match"},
		{name: "valid_new_registered_placement", newPlacement: true, mutationJSON: `{"thumbnail_url":"/uploads/products/product-small.png"}`},
		{name: "valid_independent_registered_thumbnail", setupSQL: `INSERT INTO media_assets(id,filename,url,storage_key) VALUES(910001,'other.png','/uploads/products/other.png','products/other.png')`, mutationJSON: `{"thumbnail_url":"/uploads/products/other.png"}`},
		{name: "unchanged_unregistered_legacy_main_is_preserved", setupSQL: fmt.Sprintf(`UPDATE product_media SET url='/uploads/products/legacy-unregistered.png',media_asset_id=NULL WHERE id=%d`, mediaID), mutationJSON: `{"url":"/uploads/products/legacy-unregistered.png","media_asset_id":null}`},
	} {
		t.Run("product_media_placement_"+scenario.name, func(t *testing.T) {
			t.Cleanup(func() {
				_, err := stage.ExecContext(context.Background(), `DELETE FROM product_media WHERE id=910003; DELETE FROM product_variant_option_values WHERE id=910002; DELETE FROM media_assets WHERE id=910001`)
				require.NoError(t, err)
				_, err = stage.ExecContext(context.Background(), `UPDATE product_media SET url=r.url,media_asset_id=r.media_asset_id,variant_id=r.variant_id,variant_option_value_id=r.variant_option_value_id,product_id=r.product_id,media_type=r.media_type,thumbnail_url=r.thumbnail_url,poster_url=r.poster_url,image_variants=r.image_variants FROM jsonb_populate_record(NULL::product_media,$1::jsonb) r WHERE product_media.id=r.id`, originalMedia)
				require.NoError(t, err)
			})
			if scenario.setupSQL != "" {
				_, err := stage.ExecContext(context.Background(), scenario.setupSQL)
				require.NoError(t, err)
			}
			changed := copyProductCatalogSnapshotWithMatchingRowJSONFieldsChanged(t, snapshot, "", "", "")
			mediaTable := changed.Tables["product_media"]
			matched := false
			for i, row := range mediaTable.Rows {
				id, err := productCatalogRowString(row, "id")
				require.NoError(t, err)
				if id != fmt.Sprint(mediaID) {
					continue
				}
				updated := copyProductCatalogRowWithJSONFieldsChanged(t, row, scenario.mutationJSON)
				if scenario.newPlacement {
					updated = copyProductCatalogRowWithJSONFieldsChanged(t, updated, `{"id":910003}`)
					mediaTable.Rows = append(mediaTable.Rows, updated)
				} else {
					mediaTable.Rows[i] = updated
				}
				matched = true
				break
			}
			require.True(t, matched)
			changed.Tables["product_media"] = mediaTable
			// Include synthetic referenced rows so a separate FK absence cannot mask
			// the ownership/identity rule being exercised.
			for _, table := range []struct{ name, predicate string }{{"media_assets", "id=910001"}, {"product_variant_option_values", "id=910002"}} {
				tx, err := stage.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
				require.NoError(t, err)
				rows, err := readProductCatalogRows(ctx, tx, table.name, table.predicate)
				require.NoError(t, err)
				require.NoError(t, tx.Rollback())
				selected := changed.Tables[table.name]
				selected.Rows = append(selected.Rows, rows...)
				changed.Tables[table.name] = selected
			}
			_, before, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			_, _, result, err := cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, changed)
			if scenario.expectedError == "" {
				require.NoError(t, err)
				require.NotNil(t, result)
			} else {
				require.ErrorContains(t, err, scenario.expectedError)
				require.Nil(t, result)
				_, after, _, err := captureDatabaseInventory(ctx, stage)
				require.NoError(t, err)
				require.Equal(t, before, after, "invalid placements must preserve every row and nontransactional sequence")
			}
		})
	}

	for _, tableName := range []string{"product_variant_option_values", "product_spec_option_items"} {
		for _, scenario := range []struct{ name, mutationJSON, expectedError string }{
			{"asset_url_disagree", `{"swatch_media_asset_id":910001}`, "asset ID and URL identify different files"},
			{"asset_without_url", `{"swatch_url":null}`, "requires a persisted URL"},
			{"unregistered_changed_url", `{"swatch_media_asset_id":null,"swatch_url":"/uploads/products/new-unregistered.png"}`, "exact existing placement and field"},
			{"valid_derivative_url", `{"swatch_url":"/uploads/products/product-small.png"}`, ""},
		} {
			t.Run("product_media_placement_swatch_"+tableName+"_"+scenario.name, func(t *testing.T) {
				var insert string
				if tableName == "product_variant_option_values" {
					insert = `INSERT INTO product_variant_option_values(id,product_id,spec_definition_id,value_key,label,swatch_media_asset_id,swatch_url) VALUES(910005,$1,$2,'synthetic-swatch','Synthetic swatch',$3,'/uploads/products/product-original.png')`
				} else {
					insert = `INSERT INTO product_spec_option_items(id,spec_definition_id,value_key,default_label,swatch_media_asset_id,swatch_url) VALUES(910005,$2,'synthetic-swatch','Synthetic swatch',$3,'/uploads/products/product-original.png')`
					// Keep all three typed parameters in both fixture statements.
					insert += ` RETURNING $1::bigint`
				}
				_, err := stage.ExecContext(context.Background(), insert, productID, templateDefinitionID, assetID)
				require.NoError(t, err)
				_, err = stage.ExecContext(context.Background(), `INSERT INTO media_assets(id,filename,url,storage_key) VALUES(910001,'other.png','/uploads/products/other.png','products/other.png')`)
				require.NoError(t, err)
				t.Cleanup(func() {
					_, err := stage.ExecContext(context.Background(), `DELETE FROM `+tableName+` WHERE id=910005; DELETE FROM media_assets WHERE id=910001`)
					require.NoError(t, err)
				})
				changed := copyProductCatalogSnapshotWithMatchingRowJSONFieldsChanged(t, snapshot, "", "", "")
				tx, err := stage.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
				require.NoError(t, err)
				rows, err := readProductCatalogRows(ctx, tx, tableName, "id=910005")
				require.NoError(t, err)
				require.Len(t, rows, 1)
				otherAssets, err := readProductCatalogRows(ctx, tx, "media_assets", "id=910001")
				require.NoError(t, err)
				require.NoError(t, tx.Rollback())
				selected := changed.Tables[tableName]
				selected.Rows = append(selected.Rows, copyProductCatalogRowWithJSONFieldsChanged(t, rows[0], scenario.mutationJSON))
				changed.Tables[tableName] = selected
				selected = changed.Tables["media_assets"]
				selected.Rows = append(selected.Rows, otherAssets...)
				changed.Tables["media_assets"] = selected
				_, before, _, err := captureDatabaseInventory(ctx, stage)
				require.NoError(t, err)
				_, _, _, err = cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, changed)
				if scenario.expectedError == "" {
					require.NoError(t, err)
				} else {
					require.ErrorContains(t, err, scenario.expectedError)
					_, after, _, err := captureDatabaseInventory(ctx, stage)
					require.NoError(t, err)
					require.Equal(t, before, after)
				}
			})
		}
	}
	for _, field := range []string{"variant_id", "media_asset_id"} {
		t.Run("product_media_placement_retained_invalid_"+field, func(t *testing.T) {
			_, err := stage.ExecContext(context.Background(), `INSERT INTO product_media(id,product_id,media_asset_id,url) VALUES(910004,$1,$2,'/uploads/products/product-original.png')`, otherProductID, assetID)
			require.NoError(t, err)
			t.Cleanup(func() {
				_, err := stage.ExecContext(context.Background(), `DELETE FROM product_media WHERE id=910004`)
				require.NoError(t, err)
			})
			expectedError := "belonging to another product"
			if field == "variant_id" {
				_, err = stage.ExecContext(context.Background(), `UPDATE product_media SET variant_id=(SELECT id FROM product_variants WHERE sku='product-backup-sku') WHERE id=910004`)
			} else {
				_, err = stage.ExecContext(context.Background(), `UPDATE product_media SET url='/uploads/products/product-small.png',media_asset_id=NULL,media_type='video' WHERE id=910004`)
				expectedError = "type does not match"
			}
			require.NoError(t, err)
			_, before, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			_, _, _, err = cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, snapshot)
			require.ErrorContains(t, err, expectedError)
			_, after, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			require.Equal(t, before, after)
			tx, err := stage.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
			require.NoError(t, err)
			_, err = cfg.captureProductCatalogModuleSnapshot(ctx, tx)
			require.ErrorContains(t, err, expectedError)
			require.NoError(t, tx.Rollback())
		})
	}
	for _, version := range []int{0, productCatalogMediaPlacementGuardVersion + 1} {
		t.Run(fmt.Sprintf("product_media_placement_stopped_guard_version_%d", version), func(t *testing.T) {
			task := uploadProductCatalogPackageForIntegrationTest(t, manager, packageData)
			require.Equal(t, "product_module_isolated_verified", task.State, task.Error)
			require.Equal(t, productCatalogMediaPlacementGuardVersion, task.ProductCatalogPreviewBinding.RequiredImageCleanupProtection.MediaPlacementGuardVersion)
			task.ProductCatalogPreviewBinding.RequiredImageCleanupProtection.MediaPlacementGuardVersion = version
			require.NoError(t, manager.saveTask(task))
			require.ErrorContains(t, manager.rehearseProductCatalogPairedSwitchAndVerifiedRollback(ctx, task.ID, nil), "cleanup protection is missing or changed")
			task, err := manager.GetTask(task.ID)
			require.NoError(t, err)
			require.Nil(t, task.Journal)
		})
	}
}

func TestProductCatalogPreviousStandardTriggerPreviewProtectionPreservesRollbackDigestWithoutMediaPlacementGuard(t *testing.T) {
	legacy := `{"version":1,"required_images":3,"inspected_registrations":2,"retained_cleanup_receipts":0,"required_paths_sha256":"paths","registration_rows_sha256":"registrations","cleanup_rows_sha256":"receipts","standard_write_guard_version":1}`
	var decoded productCatalogRequiredImageCleanupProtection
	require.NoError(t, json.Unmarshal([]byte(legacy), &decoded))
	actual, err := json.Marshal(decoded)
	require.NoError(t, err)
	require.Equal(t, legacy, string(actual), "keep the previous interrupted journal's exact digest input")
}
