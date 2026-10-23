package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func verifyProductCatalogRichTextMediaTypesLegacyOccurrencesAndAtomicRollback(t *testing.T, manager *Manager, stage *sql.DB, snapshot productCatalogModuleSnapshot, packageData []byte) {
	t.Helper()
	ctx, cfg := context.Background(), manager.Configuration
	var productID, translatedProductID int64
	require.NoError(t, stage.QueryRow(`SELECT id FROM products WHERE slug='product-backup-original' AND locale='en'`).Scan(&productID))
	require.NoError(t, stage.QueryRow(`SELECT id FROM products WHERE slug='product-backup-original' AND locale='zh'`).Scan(&translatedProductID))
	_, err := stage.Exec(`INSERT INTO media_assets(id,filename,url,storage_key,media_type) VALUES
(920001,'synthetic-video.mp4','/uploads/products/synthetic-video.mp4','products/synthetic-video.mp4','video'),
(920002,'synthetic-manual.pdf','/uploads/products/synthetic-manual','products/synthetic-manual','document'),
(920003,'synthetic-owned.png','/uploads/visual-showcase/synthetic-owned.png','visual-showcase/synthetic-owned.png','image');
INSERT INTO media_asset_derivatives(id,media_asset_id,preset,url,storage_key) VALUES(920004,920003,'synthetic-owned-derivative','/uploads/products/synthetic-owned-derivative.png','products/synthetic-owned-derivative.png')`)
	require.NoError(t, err)
	defer func() {
		_, err := stage.Exec(`DELETE FROM media_asset_derivatives WHERE id=920004; DELETE FROM media_assets WHERE id IN (920001,920002,920003)`)
		require.NoError(t, err)
	}()

	const image = `<img src="/uploads/products/product-original.png">`
	const video = `<video src="/uploads/products/synthetic-video.mp4" poster="/uploads/products/product-small.png"></video>`
	const legacy = `<img src="/uploads/products/synthetic-legacy.png">`
	for _, column := range []string{"description", "short_description"} {
		for _, scenario := range []struct {
			name, previous, incoming, otherField, translated, expectedError string
			newProduct, includeOwnedRegistration                            bool
		}{
			{name: "registered_image", incoming: image},
			{name: "registered_derivative_image", incoming: `<img src='/uploads/products/product-small.png'>`},
			{name: "registered_video_and_poster", incoming: video},
			{name: "registered_extensionless_file_link", incoming: `<a href="/uploads/products/synthetic-manual#page=2">Manual</a>`},
			{name: "video_cannot_be_image", incoming: `<img src="/uploads/products/synthetic-video.mp4">`, expectedError: "type does not match"},
			{name: "image_cannot_be_video", incoming: `<video src="/uploads/products/product-original.png"></video>`, expectedError: "type does not match"},
			{name: "video_cannot_be_poster", incoming: `<video src="/uploads/products/synthetic-video.mp4" poster="/uploads/products/synthetic-video.mp4"></video>`, expectedError: "type does not match"},
			{name: "unchanged_wrong_type_is_blocked", previous: `<img src="/uploads/products/synthetic-video.mp4">`, incoming: `<img src="/uploads/products/synthetic-video.mp4">`, expectedError: "type does not match"},
			{name: "new_unregistered_reference", incoming: legacy, expectedError: "existing product, field, attribute, URL and occurrence count"},
			{name: "same_legacy_occurrence_survives_text_edit", previous: legacy, incoming: `<p>Updated explanation</p>` + legacy},
			{name: "legacy_occurrence_can_be_removed", previous: legacy, incoming: `<p>Updated explanation</p>`},
			{name: "legacy_duplicate_is_rejected", previous: legacy, incoming: legacy + legacy, expectedError: "occurrence count"},
			{name: "two_legacy_occurrences_can_remain", previous: legacy + legacy, incoming: `<p>Updated</p>` + legacy + legacy},
			{name: "legacy_cannot_move_from_other_field", otherField: legacy, incoming: legacy, expectedError: "occurrence count"},
			{name: "legacy_cannot_move_from_translated_product", translated: legacy, incoming: legacy, expectedError: "occurrence count"},
			{name: "legacy_cannot_be_copied_to_new_product", previous: legacy, incoming: legacy, newProduct: true, expectedError: "occurrence count"},
			{name: "legacy_image_cannot_become_file_link", previous: legacy, incoming: `<a href="/uploads/products/synthetic-legacy.png">Copy</a>`, expectedError: "occurrence count"},
			{name: "legacy_file_link_cannot_become_image", previous: `<a href="/uploads/products/synthetic-legacy.png">Old</a>`, incoming: legacy, expectedError: "occurrence count"},
			{name: "legacy_text_does_not_grant_attachment", previous: `<p>/uploads/products/synthetic-legacy.png</p>`, incoming: legacy, expectedError: "occurrence count"},
			{name: "legacy_alt_does_not_grant_attachment", previous: `<img src="/uploads/products/product-original.png" alt="/uploads/products/synthetic-legacy.png">`, incoming: legacy, expectedError: "occurrence count"},
			{name: "legacy_url_alias_cannot_replace_original", previous: legacy, incoming: `<img src="/uploads/products/synthetic-legacy.png?alias=1">`, expectedError: "occurrence count"},
			{name: "html_entity_decoding_preserves_exact_url_and_html", previous: `<img src="/uploads/products/synthetic-legacy.png?a=1&amp;b=2">`, incoming: `<p>Edited</p><img src='/uploads/products/synthetic-legacy.png?a=1&#38;b=2'>`},
			{name: "ordinary_links_text_and_removed_elements_are_not_files", incoming: `<p>/uploads/products/missing.png</p><a href="https://ordinary.example.invalid/photo.jpg">External</a><a href="/products/catalog">Catalog</a><script><img src="/uploads/products/missing.png"></script><img src="/uploads/products/product-original.png" alt="/uploads/products/missing.png">`},
			{name: "inline_image_is_not_portable", incoming: `<img src="data:image/png;base64,AAAA">`, expectedError: "durable local or trusted HTTP URL"},
			{name: "private_media_is_rejected", incoming: `<a href="/uploads-private/synthetic-secret">Private</a>`, expectedError: "registered local storage"},
			{name: "foreign_image_origin_is_rejected", incoming: `<img src="https://foreign.invalid/uploads/products/product-original.png">`, expectedError: "stored locally before backup"},
			{name: "registered_dedicated_owner_cannot_be_newly_attached", incoming: `<img src="/uploads/visual-showcase/synthetic-owned.png">`, expectedError: "dedicated lifecycle"},
			{name: "registered_dedicated_parent_derivative_cannot_be_newly_attached", incoming: `<img src="/uploads/products/synthetic-owned-derivative.png">`, expectedError: "dedicated lifecycle"},
			{name: "uploaded_dedicated_parent_derivative_cannot_be_newly_attached", incoming: `<img src="/uploads/products/synthetic-owned-derivative.png">`, includeOwnedRegistration: true, expectedError: "dedicated lifecycle"},
			{name: "registered_dedicated_occurrence_can_remain", previous: `<img src="/uploads/visual-showcase/synthetic-owned.png">`, incoming: `<p>Updated</p><img src="/uploads/visual-showcase/synthetic-owned.png">`},
		} {
			t.Run("product_rich_text_"+column+"_"+scenario.name, func(t *testing.T) {
				t.Cleanup(func() {
					_, err := stage.Exec(`DELETE FROM products WHERE id=920010`)
					require.NoError(t, err)
					_, err = stage.Exec(`UPDATE products SET description=NULL,short_description=NULL WHERE id=$1`, translatedProductID)
					require.NoError(t, err)
					_, err = stage.Exec(`UPDATE products SET description='Original complete description',short_description=NULL WHERE id=$1`, productID)
					require.NoError(t, err)
				})
				otherColumn := "short_description"
				if column == "short_description" {
					otherColumn = "description"
				}
				_, err := stage.Exec(`UPDATE products SET `+column+`=$1,`+otherColumn+`=$2 WHERE id=$3`, scenario.previous, scenario.otherField, productID)
				require.NoError(t, err)
				_, err = stage.Exec(`UPDATE products SET `+column+`=$1 WHERE id=$2`, scenario.translated, translatedProductID)
				require.NoError(t, err)
				mutation, err := json.Marshal(map[string]any{column: scenario.incoming})
				require.NoError(t, err)
				changed := copyProductCatalogSnapshotWithMatchingRowJSONFieldsChanged(t, snapshot, "", "", "")
				products := changed.Tables["products"]
				for index, row := range products.Rows {
					id, _ := productCatalogRowString(row, "id")
					if id == fmt.Sprint(productID) {
						products.Rows[index] = copyProductCatalogRowWithJSONFieldsChanged(t, row, string(mutation))
					}
				}
				changed.Tables["products"] = products
				if scenario.newProduct {
					products := changed.Tables["products"]
					for _, row := range products.Rows {
						id, _ := productCatalogRowString(row, "id")
						if id == fmt.Sprint(productID) {
							products.Rows = append(products.Rows, copyProductCatalogRowWithJSONFieldsChanged(t, row, `{"id":920010,"slug":"synthetic-rich-text-new-product"}`))
						}
					}
					changed.Tables["products"] = products
				}
				if scenario.includeOwnedRegistration {
					for _, tableName := range []string{"media_assets", "media_asset_derivatives"} {
						tx, err := stage.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
						require.NoError(t, err)
						rows, err := readProductCatalogRows(ctx, tx, tableName, "id IN (920003,920004)")
						require.NoError(t, err)
						require.NoError(t, tx.Rollback())
						selected := changed.Tables[tableName]
						selected.Rows = append(selected.Rows, rows...)
						changed.Tables[tableName] = selected
					}
				}
				_, before, _, err := captureDatabaseInventory(ctx, stage)
				require.NoError(t, err)
				_, _, result, err := cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, changed)
				if scenario.expectedError != "" {
					require.ErrorContains(t, err, scenario.expectedError)
					require.Nil(t, result)
					_, after, _, err := captureDatabaseInventory(ctx, stage)
					require.NoError(t, err)
					require.Equal(t, before, after, "rejected HTML must preserve all business rows and sequences")
					return
				}
				require.NoError(t, err)
				require.NotNil(t, result)
				var persisted string
				require.NoError(t, stage.QueryRow(`SELECT `+column+` FROM products WHERE id=$1`, productID).Scan(&persisted))
				require.Equal(t, scenario.incoming, persisted, "restore must preserve the exact HTML instead of rewriting URLs")
			})
		}
	}
	for _, deleted := range []bool{false, true} {
		t.Run(fmt.Sprintf("product_rich_text_retained_invalid_type_deleted_%t", deleted), func(t *testing.T) {
			_, err := stage.Exec(`UPDATE products SET description='<img src="/uploads/products/synthetic-video.mp4">',deleted_at=CASE WHEN $1 THEN now() ELSE NULL END WHERE slug='target-only-product'`, deleted)
			require.NoError(t, err)
			t.Cleanup(func() {
				_, err := stage.Exec(`UPDATE products SET description=NULL,deleted_at=NULL WHERE slug='target-only-product'`)
				require.NoError(t, err)
			})
			_, before, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			_, _, _, err = cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, snapshot)
			require.ErrorContains(t, err, "type does not match")
			_, after, _, err := captureDatabaseInventory(ctx, stage)
			require.NoError(t, err)
			require.Equal(t, before, after)
			tx, err := stage.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
			require.NoError(t, err)
			_, err = cfg.captureProductCatalogModuleSnapshot(ctx, tx)
			require.ErrorContains(t, err, "type does not match")
			require.NoError(t, tx.Rollback())
		})
	}
	for _, version := range []int{0, productCatalogRichTextMediaGuardVersion + 1} {
		t.Run(fmt.Sprintf("product_rich_text_stopped_guard_version_%d", version), func(t *testing.T) {
			task := uploadProductCatalogPackageForIntegrationTest(t, manager, packageData)
			require.Equal(t, "product_module_isolated_verified", task.State, task.Error)
			require.Equal(t, productCatalogRichTextMediaGuardVersion, task.ProductCatalogPreviewBinding.RequiredImageCleanupProtection.RichTextMediaGuardVersion)
			task.ProductCatalogPreviewBinding.RequiredImageCleanupProtection.RichTextMediaGuardVersion = version
			require.NoError(t, manager.saveTask(task))
			require.ErrorContains(t, manager.rehearseProductCatalogPairedSwitchAndVerifiedRollback(ctx, task.ID, nil), "cleanup protection is missing or changed")
			current, err := manager.GetTask(task.ID)
			require.NoError(t, err)
			require.Nil(t, current.Journal)
		})
	}
	t.Run("product_rich_text_actual_encrypted_video_and_extensionless_file_round_trip", func(t *testing.T) {
		verifyProductCatalogRichTextVideoAndExtensionlessFileEncryptedRoundTrip(t, manager)
	})
}

func verifyProductCatalogRichTextVideoAndExtensionlessFileEncryptedRoundTrip(t *testing.T, manager *Manager) {
	t.Helper()
	ctx, cfg := context.Background(), manager.Configuration
	db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	const html = `<p>/uploads/products/not-an-attachment.png</p><video src="/uploads/products/synthetic-video.mp4" poster="/uploads/products/product-small.png"></video><a href="/uploads/products/synthetic-manual#page=2">Manual</a><a href="https://external.invalid/product.jpg">External link</a>`
	const summary = `<p>Canonical summary &amp; exact</p><img src="/uploads/products/product-original.png">`
	const seo = "Canonical SEO description & exact"
	_, err = db.Exec(`INSERT INTO media_assets(id,filename,url,storage_key,media_type) VALUES(920001,'synthetic-video.mp4','/uploads/products/synthetic-video.mp4','products/synthetic-video.mp4','video'),(920002,'synthetic-manual.pdf','/uploads/products/synthetic-manual','products/synthetic-manual','document'); INSERT INTO products(id,name,slug,price_minor,currency) VALUES(920010,'Synthetic rich text product','synthetic-rich-text-product',100,'USD')`)
	require.NoError(t, err)
	_, err = db.Exec(`UPDATE products SET description=$1,short_description=$2,meta_description=$3 WHERE id=920010`, html, summary, seo)
	require.NoError(t, err)
	files := map[string][]byte{"products/synthetic-video.mp4": []byte("synthetic-video-file-content"), "products/synthetic-manual": []byte("synthetic-extensionless-document-content")}
	t.Cleanup(func() {
		_, err := db.Exec(`DELETE FROM products WHERE id=920010; DELETE FROM media_assets WHERE id IN (920001,920002)`)
		require.NoError(t, err)
		for key := range files {
			require.NoError(t, os.Remove(filepath.Join(cfg.StorageRoots["uploads"], filepath.FromSlash(key))))
		}
	})
	for key, body := range files {
		require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads"], filepath.FromSlash(key)), body, 0600))
		_, err := db.Exec(`UPDATE media_assets SET size=$1 WHERE storage_key=$2`, len(body), key)
		require.NoError(t, err)
	}
	_, before, _, err := captureDatabaseInventory(ctx, db)
	require.NoError(t, err)
	beforeMedia, err := cfg.mediaFingerprint(ctx)
	require.NoError(t, err)
	backup, err := manager.CreateProductCatalogBackup("synthetic-rich-text-operator", "BACKUP PRODUCTS "+cfg.SiteID+"/"+cfg.Environment)
	require.NoError(t, err)
	backup = waitForIndependentRecoveryTask(t, manager, backup.ID)
	require.Equal(t, "ready_to_download", backup.State, backup.Error)
	packagePath, err := manager.taskPackagePath(backup.ID)
	require.NoError(t, err)
	data, err := os.ReadFile(packagePath)
	require.NoError(t, err)
	verified := uploadProductCatalogPackageForIntegrationTest(t, manager, data)
	require.Equal(t, "product_module_isolated_verified", verified.State, verified.Error)
	directory, err := manager.taskDirectory(verified.ID)
	require.NoError(t, err)
	for key, expected := range files {
		actual, err := os.ReadFile(filepath.Join(directory, "product-isolated-target/media/uploads", filepath.FromSlash(key)))
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	}
	clone, err := cfg.connectDatabase(ctx, verified.StageDatabase)
	require.NoError(t, err)
	defer clone.Close()
	var restoredHTML, restoredSummary, restoredSEO string
	require.NoError(t, clone.QueryRow(`SELECT description,short_description,meta_description FROM products WHERE id=920010`).Scan(&restoredHTML, &restoredSummary, &restoredSEO))
	require.Equal(t, html, restoredHTML)
	require.Equal(t, summary, restoredSummary)
	require.Equal(t, seo, restoredSEO)
	_, after, _, err := captureDatabaseInventory(ctx, db)
	require.NoError(t, err)
	require.Equal(t, before, after)
	afterMedia, err := cfg.mediaFingerprint(ctx)
	require.NoError(t, err)
	require.Equal(t, beforeMedia, afterMedia)
	t.Logf("rich text encrypted export/upload restored %d rows and %d referenced files; video/document bytes, exact HTML, canonical summary and SEO verified", verified.ProductCatalogRecovery.RestoredRows, verified.ProductCatalogRecovery.VerifiedImageFiles)
}

func TestProductCatalogPreviousMediaPlacementPreviewProtectionPreservesRollbackDigestWithoutRichTextGuard(t *testing.T) {
	legacy := `{"version":1,"required_images":3,"inspected_registrations":2,"retained_cleanup_receipts":0,"required_paths_sha256":"paths","registration_rows_sha256":"registrations","cleanup_rows_sha256":"receipts","standard_write_guard_version":1,"media_placement_guard_version":1}`
	var decoded productCatalogRequiredImageCleanupProtection
	require.NoError(t, json.Unmarshal([]byte(legacy), &decoded))
	actual, err := json.Marshal(decoded)
	require.NoError(t, err)
	require.Equal(t, legacy, string(actual), "preserve the previous interrupted journal's exact digest input")
}

func TestProductCatalogRichTextMediaUsesEachPersistedShortDescriptionColumnAndOnlyActualMediaAttributes(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	cfg.PublicOrigins = []string{"https://catalog.example.invalid"}
	cfg.ProductStaticAssetsDirectory = t.TempDir()
	row := json.RawMessage(`{"id":1,"description":"<p>/uploads/missing-text.png</p><img src='/uploads/photo.png?a=1&amp;b=2' alt='/uploads/missing-alt.png'><script><img src='/uploads/missing-script.png'></script><a href='https://external.invalid/photo.jpg'>External</a><a href='/products/page'>Navigation</a>","short_description":"<a href='/uploads/manual#page=2'>Manual</a>","short_desc":"<video src='https://catalog.example.invalid/uploads/video.mp4' poster='/uploads/photo.png'></video>"}`)
	snapshot := productCatalogModuleSnapshot{Tables: map[string]productCatalogTableSnapshot{
		"products": {Rows: []json.RawMessage{row}},
		"media_assets": {Rows: []json.RawMessage{
			json.RawMessage(`{"id":1,"media_type":"image","storage_key":"photo.png"}`),
			json.RawMessage(`{"id":2,"media_type":"video","storage_key":"video.mp4"}`),
			json.RawMessage(`{"id":3,"media_type":"document","storage_key":"manual"}`),
		}},
	}}
	before, err := json.Marshal(snapshot)
	require.NoError(t, err)
	references, err := cfg.collectProductCatalogMediaReferences(snapshot)
	require.NoError(t, err)
	require.Equal(t, map[string]bool{"/uploads/photo.png?a=1&b=2": true, "/uploads/manual#page=2": true, "https://catalog.example.invalid/uploads/video.mp4": true, "/uploads/photo.png": true}, references)
	require.NoError(t, cfg.validateProductCatalogCompleteRichTextRegisteredMediaTypes(snapshot))
	after, err := json.Marshal(snapshot)
	require.NoError(t, err)
	require.Equal(t, before, after)
	for _, column := range []string{"short_description", "short_desc"} {
		mutation, err := json.Marshal(map[string]string{column: `<img src='/uploads/video.mp4'>`})
		require.NoError(t, err)
		changed := copyProductCatalogSnapshotWithMatchingRowJSONFieldsChanged(t, snapshot, "", "", "")
		products := changed.Tables["products"]
		products.Rows[0] = copyProductCatalogRowWithJSONFieldsChanged(t, row, string(mutation))
		changed.Tables["products"] = products
		require.ErrorContains(t, cfg.validateProductCatalogCompleteRichTextRegisteredMediaTypes(changed), "product rich text "+column+" img.src media type does not match")
	}
}
