package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func verifyProductCatalogTemplateAndCategoryRichTextMediaTypesLegacyOccurrencesAndAtomicRollback(t *testing.T, manager *Manager, stage *sql.DB, packageData []byte) {
	t.Helper()
	ctx, cfg := context.Background(), manager.Configuration
	_, err := stage.ExecContext(context.Background(), `INSERT INTO media_assets(id,filename,url,storage_key,media_type) VALUES
(930001,'template-video.mp4','/uploads/products/template-video.mp4','products/template-video.mp4','video'),
(930002,'template-manual.pdf','/uploads/products/template-manual','products/template-manual','document'),
(930003,'template-owned.png','/uploads/visual-showcase/template-owned.png','visual-showcase/template-owned.png','image');
INSERT INTO media_asset_derivatives(id,media_asset_id,preset,url,storage_key) VALUES
(930004,930003,'template-owned-derivative','/uploads/products/template-owned-derivative.png','products/template-owned-derivative.png');
INSERT INTO product_information_templates(id,kind,name,slug,locale,is_enabled) VALUES
(930010,'packaging','Synthetic template','template-contract-en','en',false),
(930011,'packaging','Synthetic translated template','template-contract-fr','fr',true);
INSERT INTO product_categories(id,name,slug,is_enabled) VALUES
(930020,'Synthetic category','template-contract-category',false),
(930021,'Synthetic other category','template-contract-other-category',true);
INSERT INTO product_category_translations(id,product_category_id,locale,name) VALUES
(930030,930020,'fr','Synthetic French category'),(930031,930020,'de','Synthetic German category')`)
	require.NoError(t, err)
	defer func() {
		_, err := stage.ExecContext(context.Background(), `DELETE FROM product_information_templates WHERE id IN (930010,930011,930012);
DELETE FROM product_categories WHERE id IN (930020,930021,930022);
DELETE FROM media_asset_derivatives WHERE id=930004; DELETE FROM media_assets WHERE id IN (930001,930002,930003)`)
		require.NoError(t, err)
	}()
	tx, err := stage.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	require.NoError(t, err)
	snapshot, err := cfg.captureProductCatalogModuleSnapshot(ctx, tx)
	require.NoError(t, err)
	registeredRows, err := readProductCatalogRows(ctx, tx, "media_assets", "id IN (930001,930002)")
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
	registered := snapshot.Tables["media_assets"]
	registered.Rows = append(registered.Rows, registeredRows...)
	snapshot.Tables["media_assets"] = registered

	const image = `<img src="/uploads/products/product-original.png">`
	const legacy = `<img src="/uploads/products/template-legacy.png">`
	for _, source := range []struct {
		table, column      string
		id, otherID, newID int64
	}{
		{"product_information_templates", "content", 930010, 930011, 930012},
		{"product_categories", "seo_intro", 930020, 930021, 930022},
		{"product_category_translations", "seo_intro", 930030, 930031, 930032},
	} {
		for _, scenario := range []struct {
			name, previous, incoming, otherPrevious, expectedError string
			newRecord, includeOwnedRegistration, deletedTemplate   bool
		}{
			{name: "registered_image", incoming: image},
			{name: "registered_derivative", incoming: `<img src='/uploads/products/product-small.png'>`},
			{name: "registered_video_poster_and_extensionless_link", incoming: `<video src="/uploads/products/template-video.mp4" poster="/uploads/products/product-small.png"></video><a href="/uploads/products/template-manual#page=2">Manual</a>`},
			{name: "video_cannot_be_image", incoming: `<img src="/uploads/products/template-video.mp4">`, expectedError: "type does not match"},
			{name: "image_cannot_be_video", incoming: `<video src="/uploads/products/product-original.png"></video>`, expectedError: "type does not match"},
			{name: "video_cannot_be_poster", incoming: `<video src="/uploads/products/template-video.mp4" poster="/uploads/products/template-video.mp4"></video>`, expectedError: "type does not match"},
			{name: "unchanged_wrong_type_still_blocks", previous: `<img src="/uploads/products/template-video.mp4">`, incoming: `<img src="/uploads/products/template-video.mp4">`, expectedError: "type does not match"},
			{name: "new_unregistered_reference", incoming: legacy, expectedError: "occurrence count"},
			{name: "same_legacy_occurrence_allows_text_edit", previous: legacy, incoming: `<p>Edited</p>` + legacy},
			{name: "legacy_occurrences_can_decrease", previous: legacy + legacy, incoming: legacy},
			{name: "legacy_occurrence_can_be_removed", previous: legacy, incoming: `<p>Removed attachment</p>`},
			{name: "legacy_duplicate_is_rejected", previous: legacy, incoming: legacy + legacy, expectedError: "occurrence count"},
			{name: "legacy_cannot_move_from_other_record_or_translation", otherPrevious: legacy, incoming: legacy, expectedError: "occurrence count"},
			{name: "legacy_cannot_be_copied_to_new_record", previous: legacy, incoming: legacy, newRecord: true, expectedError: "occurrence count"},
			{name: "legacy_image_cannot_become_link", previous: legacy, incoming: `<a href="/uploads/products/template-legacy.png">File</a>`, expectedError: "occurrence count"},
			{name: "plain_text_does_not_grant_attachment", previous: `<p>/uploads/products/template-legacy.png</p>`, incoming: legacy, expectedError: "occurrence count"},
			{name: "alt_does_not_grant_attachment", previous: `<img src="/uploads/products/product-original.png" alt="/uploads/products/template-legacy.png">`, incoming: legacy, expectedError: "occurrence count"},
			{name: "exact_html_and_decoded_entity_url_survive", previous: `<img src="/uploads/products/template-legacy.png?a=1&amp;b=2">`, incoming: `<p>Edited</p><img src='/uploads/products/template-legacy.png?a=1&#38;b=2'>`},
			{name: "ordinary_text_links_alt_and_removed_script_are_not_files", incoming: `<p>/uploads/products/missing.png</p><a href="https://external.invalid/photo.jpg">External</a><script><img src="/uploads/products/missing.png"></script><img src="/uploads/products/product-original.png" alt="/uploads/products/missing-alt.png">`},
			{name: "dedicated_registration_cannot_authorize_new_placement", incoming: `<img src="/uploads/visual-showcase/template-owned.png">`, expectedError: "dedicated lifecycle"},
			{name: "dedicated_parent_derivative_cannot_authorize_new_placement", incoming: `<img src="/uploads/products/template-owned-derivative.png">`, expectedError: "dedicated lifecycle"},
			{name: "uploaded_dedicated_parent_cannot_authorize_new_placement", incoming: `<img src="/uploads/products/template-owned-derivative.png">`, includeOwnedRegistration: true, expectedError: "dedicated lifecycle"},
			{name: "existing_dedicated_occurrence_can_remain", previous: `<img src="/uploads/visual-showcase/template-owned.png">`, incoming: `<p>Edited</p><img src="/uploads/visual-showcase/template-owned.png">`},
			{name: "deleted_template_invalid_type_still_blocks", incoming: `<img src="/uploads/products/template-video.mp4">`, deletedTemplate: true, expectedError: "type does not match"},
		} {
			if scenario.deletedTemplate && source.table != "product_information_templates" {
				continue
			}
			t.Run("product_template_category_rich_text_"+source.table+"_"+scenario.name, func(t *testing.T) {
				t.Cleanup(func() {
					_, err := stage.ExecContext(context.Background(), `DELETE FROM `+pq.QuoteIdentifier(source.table)+` WHERE id=$1`, source.newID)
					require.NoError(t, err)
					_, err = stage.ExecContext(context.Background(), `UPDATE `+pq.QuoteIdentifier(source.table)+` SET `+pq.QuoteIdentifier(source.column)+`='' WHERE id IN ($1,$2)`, source.id, source.otherID)
					require.NoError(t, err)
					if source.table == "product_information_templates" {
						_, err = stage.ExecContext(context.Background(), `UPDATE product_information_templates SET deleted_at=NULL WHERE id=$1`, source.id)
						require.NoError(t, err)
					}
				})
				_, err := stage.ExecContext(context.Background(), `UPDATE `+pq.QuoteIdentifier(source.table)+` SET `+pq.QuoteIdentifier(source.column)+`=$1 WHERE id=$2`, scenario.previous, source.id)
				require.NoError(t, err)
				_, err = stage.ExecContext(context.Background(), `UPDATE `+pq.QuoteIdentifier(source.table)+` SET `+pq.QuoteIdentifier(source.column)+`=$1 WHERE id=$2`, scenario.otherPrevious, source.otherID)
				require.NoError(t, err)
				mutation := map[string]any{source.column: scenario.incoming}
				if scenario.deletedTemplate {
					mutation["deleted_at"] = "2026-10-20T00:00:00Z"
				}
				changed := copyProductCatalogSnapshotWithMatchingRowJSONFieldsChanged(t, snapshot, "", "", "")
				selected := changed.Tables[source.table]
				for index, row := range selected.Rows {
					id, _ := productCatalogRowString(row, "id")
					if id != fmt.Sprint(source.id) {
						continue
					}
					encoded, err := json.Marshal(mutation)
					require.NoError(t, err)
					selected.Rows[index] = copyProductCatalogRowWithJSONFieldsChanged(t, row, string(encoded))
					if scenario.newRecord {
						mutation["id"] = source.newID
						mutation["slug"] = "template-contract-new-record"
						if source.table == "product_category_translations" {
							delete(mutation, "slug")
							mutation["locale"] = "es"
						}
						encoded, err := json.Marshal(mutation)
						require.NoError(t, err)
						selected.Rows = append(selected.Rows, copyProductCatalogRowWithJSONFieldsChanged(t, row, string(encoded)))
					}
				}
				changed.Tables[source.table] = selected
				if scenario.includeOwnedRegistration {
					for _, table := range []string{"media_assets", "media_asset_derivatives"} {
						tx, err := stage.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
						require.NoError(t, err)
						rows, err := readProductCatalogRows(ctx, tx, table, "id IN (930003,930004)")
						require.NoError(t, err)
						require.NoError(t, tx.Rollback())
						selected := changed.Tables[table]
						selected.Rows = append(selected.Rows, rows...)
						changed.Tables[table] = selected
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
					require.Equal(t, before, after, "rejection must preserve all business rows and sequences")
					return
				}
				require.NoError(t, err)
				require.NotNil(t, result)
				var restored string
				require.NoError(t, stage.QueryRowContext(context.Background(), `SELECT `+pq.QuoteIdentifier(source.column)+` FROM `+pq.QuoteIdentifier(source.table)+` WHERE id=$1`, source.id).Scan(&restored))
				require.Equal(t, scenario.incoming, restored)
			})
		}
		for _, deleted := range []bool{false, true} {
			if deleted && source.table != "product_information_templates" {
				continue
			}
			t.Run(fmt.Sprintf("product_template_category_rich_text_%s_target_only_bad_type_deleted_%t", source.table, deleted), func(t *testing.T) {
				// Remove this row from the upload: its retained target content still participates.
				changed := copyProductCatalogSnapshotWithMatchingRowJSONFieldsChanged(t, snapshot, "", "", "")
				selected := changed.Tables[source.table]
				selected.Rows = nil
				for _, row := range snapshot.Tables[source.table].Rows {
					id, _ := productCatalogRowString(row, "id")
					if id != fmt.Sprint(source.id) {
						selected.Rows = append(selected.Rows, row)
					}
				}
				changed.Tables[source.table] = selected
				if source.table == "product_categories" {
					// A target-only category also retains its translations. Uploading
					// those children without their parent fails dependency validation
					// before reaching the retained rich-text type check.
					translations := changed.Tables["product_category_translations"]
					translations.Rows = nil
					for _, row := range snapshot.Tables["product_category_translations"].Rows {
						parentID, _ := productCatalogRowString(row, "product_category_id")
						if parentID != fmt.Sprint(source.id) {
							translations.Rows = append(translations.Rows, row)
						}
					}
					changed.Tables["product_category_translations"] = translations
				}
				_, err := stage.ExecContext(context.Background(), `UPDATE `+pq.QuoteIdentifier(source.table)+` SET `+pq.QuoteIdentifier(source.column)+`='<img src="/uploads/products/template-video.mp4">' WHERE id=$1`, source.id)
				require.NoError(t, err)
				if deleted {
					_, err = stage.ExecContext(context.Background(), `UPDATE product_information_templates SET deleted_at=now() WHERE id=$1`, source.id)
					require.NoError(t, err)
				}
				t.Cleanup(func() {
					_, err := stage.ExecContext(context.Background(), `UPDATE `+pq.QuoteIdentifier(source.table)+` SET `+pq.QuoteIdentifier(source.column)+`='' WHERE id=$1`, source.id)
					require.NoError(t, err)
					if deleted {
						_, err = stage.ExecContext(context.Background(), `UPDATE product_information_templates SET deleted_at=NULL WHERE id=$1`, source.id)
						require.NoError(t, err)
					}
				})
				_, before, _, err := captureDatabaseInventory(ctx, stage)
				require.NoError(t, err)
				_, _, _, err = cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, changed)
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
	}
	for _, version := range []int{0, productCatalogTemplateAndCategoryRichTextGuardVersion + 1} {
		t.Run(fmt.Sprintf("product_template_category_rich_text_stopped_guard_version_%d", version), func(t *testing.T) {
			task := uploadProductCatalogPackageForIntegrationTest(t, manager, packageData)
			require.Equal(t, "product_module_isolated_verified", task.State, task.Error)
			require.Equal(t, productCatalogTemplateAndCategoryRichTextGuardVersion, task.ProductCatalogPreviewBinding.RequiredImageCleanupProtection.TemplateAndCategoryRichTextGuardVersion)
			task.ProductCatalogPreviewBinding.RequiredImageCleanupProtection.TemplateAndCategoryRichTextGuardVersion = version
			require.NoError(t, manager.saveTask(task))
			require.ErrorContains(t, manager.rehearseProductCatalogPairedSwitchAndVerifiedRollback(ctx, task.ID, nil), "cleanup protection is missing or changed")
			current, err := manager.GetTask(task.ID)
			require.NoError(t, err)
			require.Nil(t, current.Journal)
		})
	}
	t.Run("product_template_category_rich_text_encrypted_video_document_and_exact_rows_round_trip", func(t *testing.T) {
		verifyProductCatalogTemplateAndCategoryRichTextEncryptedFileAndExactPersistedRowRoundTrip(t, manager)
	})
}

func verifyProductCatalogTemplateAndCategoryRichTextEncryptedFileAndExactPersistedRowRoundTrip(t *testing.T, manager *Manager) {
	t.Helper()
	ctx, cfg := context.Background(), manager.Configuration
	db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	_, err = db.ExecContext(context.Background(), `INSERT INTO media_assets(id,filename,url,storage_key,media_type) VALUES
(930001,'template-video.mp4','/uploads/products/template-video.mp4','products/template-video.mp4','video'),
(930002,'template-manual.pdf','/uploads/products/template-manual','products/template-manual','document');
INSERT INTO product_information_templates(id,kind,name,slug,locale,is_enabled,deleted_at) VALUES
(930010,'packaging','Synthetic /uploads/products/missing-name.png','template-round-trip-disabled','en',false,NULL),
(930011,'after_sales','Synthetic deleted template','template-round-trip-deleted','en',true,'2026-10-20T00:00:00Z');
INSERT INTO product_categories(id,name,slug,is_enabled) VALUES(930020,'Synthetic category','template-round-trip-category',false);
INSERT INTO product_category_translations(id,product_category_id,locale,name) VALUES(930030,930020,'fr','Synthetic French category')`)
	require.NoError(t, err)
	const html = `<p>Exact &amp; body /uploads/products/missing-body.png</p><video src='/uploads/products/template-video.mp4' poster="/uploads/products/product-small.png"></video><a href="/uploads/products/template-manual#page=2">Manual</a><img src="/uploads/products/product-original.png" alt="/uploads/products/missing-alt.png"><a href="https://external.invalid/photo.jpg">External</a>`
	const text = `<img src="/uploads/products/missing-plain-text.png">`
	_, err = db.ExecContext(context.Background(), `UPDATE product_information_templates SET content=$1 WHERE id IN (930010,930011)`, html)
	require.NoError(t, err)
	for _, table := range []string{"product_categories", "product_category_translations"} {
		_, err = db.ExecContext(context.Background(), `UPDATE `+pq.QuoteIdentifier(table)+` SET seo_intro=$1,description=$2::text,meta_title=$2::text,meta_description=$2::text WHERE id IN (930020,930030)`, html, text)
		require.NoError(t, err)
	}
	files := map[string][]byte{
		"products/template-video.mp4": []byte("synthetic-template-video-file-content"),
		"products/template-manual":    []byte("synthetic-template-extensionless-document-content"),
	}
	t.Cleanup(func() {
		_, err := db.ExecContext(context.Background(), `DELETE FROM product_information_templates WHERE id IN (930010,930011); DELETE FROM product_categories WHERE id=930020; DELETE FROM media_assets WHERE id IN (930001,930002)`)
		require.NoError(t, err)
		for key := range files {
			require.NoError(t, os.Remove(filepath.Join(cfg.StorageRoots["uploads"], filepath.FromSlash(key))))
		}
	})
	for key, body := range files {
		require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads"], filepath.FromSlash(key)), body, 0600))
		_, err := db.ExecContext(context.Background(), `UPDATE media_assets SET size=$1 WHERE storage_key=$2`, len(body), key)
		require.NoError(t, err)
	}
	_, before, _, err := captureDatabaseInventory(ctx, db)
	require.NoError(t, err)
	beforeMedia, err := cfg.mediaFingerprint(ctx)
	require.NoError(t, err)
	backup, err := manager.CreateProductCatalogBackup("synthetic-template-category-operator", "BACKUP PRODUCTS "+cfg.SiteID+"/"+cfg.Environment)
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
	for _, source := range []struct {
		table     string
		predicate string
	}{
		{"product_information_templates", "id IN (930010,930011)"},
		{"product_categories", "id=930020"},
		{"product_category_translations", "id=930030"},
	} {
		query := `SELECT jsonb_agg(to_jsonb(p) ORDER BY id)::text FROM ` + pq.QuoteIdentifier(source.table) + ` p WHERE ` + source.predicate
		var expected, actual string
		require.NoError(t, db.QueryRowContext(context.Background(), query).Scan(&expected))
		require.NoError(t, clone.QueryRowContext(context.Background(), query).Scan(&actual))
		require.JSONEq(t, expected, actual, "all persisted row fields, HTML and disabled/deleted status must survive")
	}
	_, after, _, err := captureDatabaseInventory(ctx, db)
	require.NoError(t, err)
	require.Equal(t, before, after)
	afterMedia, err := cfg.mediaFingerprint(ctx)
	require.NoError(t, err)
	require.Equal(t, beforeMedia, afterMedia)
	t.Logf("template/category encrypted export and HTTP upload restored %d rows and %d referenced files; four exact rows, disabled/deleted status, video/document bytes and unchanged source fingerprints verified", verified.ProductCatalogRecovery.RestoredRows, verified.ProductCatalogRecovery.VerifiedImageFiles)
}
