package localbackup

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

func TestProductCatalogPackagesCannotContainWholeDatabaseSecretsOrPrivateMedia(t *testing.T) {
	for _, filename := range []string{"database.dump", "recovery-secrets.json", "media/uploads_private/chat.png", "media/site_logo_uploads/logo.png", "product-static/../secrets", "product-static/company/../../secrets", "media/uploads/../secrets"} {
		require.False(t, isAllowedBackupFileForScope(productCatalogBackupScope, filename), filename)
	}
	require.True(t, isAllowedBackupFileForScope(productCatalogBackupScope, "media/uploads/products/photo.webp"))
	require.True(t, isAllowedBackupFileForScope(productCatalogBackupScope, "product-static/company/products/photo.webp"))
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	for _, permission := range []string{PermissionView, PermissionUpload, PermissionRestore} {
		token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic", []string{permission})
		require.NoError(t, err)
		request := httptest.NewRequestWithContext(context.Background(), "POST", "/api/product-catalog-backups", strings.NewReader(`{}`))
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		manager.HTTPHandler().ServeHTTP(response, request)
		require.Equal(t, 403, response.Code, permission)
	}
	task, err := manager.createTask("upload", "synthetic")
	require.NoError(t, err)
	task.State = "verified"
	task.Manifest = &Manifest{Scope: productCatalogBackupScope, PackageID: newTaskID()}
	require.NoError(t, manager.saveTask(task))
	_, err = manager.ApplyVerifiedRecovery(task.ID, "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+task.Manifest.PackageID)
	require.Error(t, err)
	_, err = cfg.collectProductCatalogMediaReferences(productCatalogModuleSnapshot{Tables: map[string]productCatalogTableSnapshot{"products": {Rows: []json.RawMessage{json.RawMessage(`{"description":"<img src='data:image/png;base64,AAAA'>","short_desc":""}`)}}}})
	require.NoError(t, err)
	_, _, err = cfg.resolveProductCatalogImageSource("https://unregistered.invalid/photo.webp")
	require.Error(t, err)
}

func TestProductCatalogPackagesRejectNoncanonicalPrimaryKeysThatCouldBypassSharedMediaIdentityLookup(t *testing.T) {
	for _, rawID := range []string{`"01"`, `0`, `-1`, `9223372036854775808`, `1e2`} {
		t.Run(rawID, func(t *testing.T) {
			snapshot := productCatalogModuleSnapshot{ContractVersion: 1, Tables: map[string]productCatalogTableSnapshot{}}
			manifest := Manifest{Tables: map[string]int64{}}
			for name := range productCatalogAllowedTableNames() {
				snapshot.Tables[name] = productCatalogTableSnapshot{Rows: []json.RawMessage{}}
				manifest.Tables[name] = 0
			}
			snapshot.Tables["media_assets"] = productCatalogTableSnapshot{Rows: []json.RawMessage{json.RawMessage(`{"id":` + rawID + `}`)}}
			manifest.Tables["media_assets"] = 1
			contents, err := json.Marshal(snapshot)
			require.NoError(t, err)
			manifest.DatabaseFingerprint = productCatalogBytesSHA256(contents)
			directory := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(directory, productCatalogRowsFilename), contents, 0600))
			_, err = readAndValidateProductCatalogModuleSnapshot(directory, manifest)
			require.ErrorContains(t, err, "invalid or duplicate product row identity")
		})
	}
}

func uploadProductCatalogPackageForIntegrationTest(t *testing.T, manager *Manager, data []byte) *Task {
	t.Helper()
	key, err := readProtectedCredential(manager.Configuration.AuthenticationKeyFile)
	require.NoError(t, err)
	token, err := IssueMaintenanceSession(key, manager.Configuration.SiteID, manager.Configuration.Environment, "synthetic-product-operator", []string{PermissionUpload})
	require.NoError(t, err)
	invoke := func(method, path string, body []byte, headers map[string]string) *Task {
		request := httptest.NewRequestWithContext(context.Background(), method, path, bytes.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+token)
		for name, value := range headers {
			request.Header.Set(name, value)
		}
		response := httptest.NewRecorder()
		manager.HTTPHandler().ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		var result struct {
			Task *Task `json:"task"`
		}
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
		require.NotNil(t, result.Task)
		return result.Task
	}
	task := invoke("POST", "/api/uploads", []byte(fmt.Sprintf(`{"size":%d}`, len(data))), nil)
	for offset := 0; offset < len(data); {
		end := offset + int(ChunkBytes)
		if end > len(data) {
			end = len(data)
		}
		chunk := data[offset:end]
		digest := sha256.Sum256(chunk)
		invoke("PUT", "/api/uploads/"+task.ID+"/chunks", chunk, map[string]string{"Upload-Offset": fmt.Sprint(offset), "X-Chunk-SHA256": hex.EncodeToString(digest[:])})
		offset = end
	}
	invoke("POST", "/api/uploads/"+task.ID+"/commit", []byte(`{}`), nil)
	invoke("POST", "/api/tasks/"+task.ID+"/verify", []byte(`{}`), nil)
	return waitForIndependentRecoveryTask(t, manager, task.ID)
}

func TestPostgreSQLProductCatalogModuleExportsActualImagesAndRestoresOnlyProductsInAnIsolatedClone(t *testing.T) {
	connectionURL := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	migrationsDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || migrationsDirectory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires an explicitly disposable PostgreSQL server with pg_dump/restore and all migrations")
	}
	ctx := context.Background()
	originalImage := createDecodableProductCatalogPNGImageFixture(t, 2, color.RGBA{R: 200, G: 40, B: 80, A: 255})
	thumbnailImage := createDecodableProductCatalogPNGImageFixture(t, 1, color.RGBA{R: 40, G: 200, B: 80, A: 255})
	cfg := createIndependentRecoveryTestConfiguration(t)
	cfg.DatabaseURL = connectionURL
	cfg.DatabaseName = "local_backup_test_" + newTaskID()[:12]
	cfg.MaxFiles = 10000
	cfg.MaxExpandedBytes = 256 << 20
	cfg.MaxWorkspaceBytes = 1 << 30
	cfg.ProductStaticAssetsDirectory = filepath.Join(t.TempDir(), "static")
	require.NoError(t, os.Mkdir(cfg.ProductStaticAssetsDirectory, 0700))
	require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
	control, err := cfg.connectDatabase(ctx, "postgres")
	require.NoError(t, err)
	defer control.Close()
	cfg.ApplicationDatabaseRole = "synthetic_product_app_" + newTaskID()[:12]
	_, err = control.ExecContext(ctx, `CREATE ROLE `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole)+` NOLOGIN`)
	require.NoError(t, err)
	defer func() {
		_, _ = control.ExecContext(context.Background(), `DROP ROLE `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole))
	}()
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
	require.NoError(t, db.QueryRowContext(ctx, `SELECT version FROM schema_migrations WHERE NOT dirty`).Scan(&cfg.ExpectedMigrationVersion))
	_, err = db.ExecContext(ctx, `DO $$ DECLARE customer bigint; product bigint; variant bigint; asset bigint; ticket bigint; order_id bigint; BEGIN
INSERT INTO users(email,username,password) VALUES('product-backup@example.invalid','product-backup-customer','synthetic-password') RETURNING id INTO customer;
INSERT INTO products(name,slug,description,price_minor,currency) VALUES('Original product','product-backup-original','Original complete description',12345,'USD') RETURNING id INTO product;
INSERT INTO customs_classification_profiles(name,slug,hs_code) VALUES('Original shared customs profile','product-backup-customs-profile','123456');
UPDATE products SET customs_classification_profile_id=(SELECT id FROM customs_classification_profiles WHERE slug='product-backup-customs-profile') WHERE id=product;
INSERT INTO product_variants(product_id,sku,price_minor,currency,stock) VALUES(product,'product-backup-sku',12345,'USD',19) RETURNING id INTO variant;
INSERT INTO product_variants(product_id,sku,option_values,price_minor,currency,stock) VALUES(product,'product-backup-new-sku','{"size":"new"}',12345,'USD',3);
INSERT INTO products(name,slug,locale,parent_id,price_minor,currency,view_count) VALUES('Original translated product','product-backup-original','zh',product,12345,'USD',5);
INSERT INTO product_variants(product_id,sku,master_variant_id,price_minor,currency,stock) VALUES((SELECT id FROM products WHERE slug='product-backup-original' AND locale='zh'),'product-backup-translated-sku',variant,12345,'USD',0);
UPDATE products SET view_count=5 WHERE id=product;
INSERT INTO product_view_count_flush_receipts(batch_id,product_id,delta,product_increment_applied) VALUES('11111111-1111-4111-8111-111111111111',product,4,true);
INSERT INTO media_assets(filename,url,storage_key,size,visibility) VALUES('product-original.png','/uploads/products/product-original.png','products/product-original.png',20,'public') RETURNING id INTO asset;
INSERT INTO media_asset_derivatives(media_asset_id,preset,url,storage_key,mime_type,size) VALUES(asset,'small','/uploads/products/product-small.png','products/product-small.png','image/webp',17);
INSERT INTO product_media(product_id,variant_id,media_asset_id,media_type,url) VALUES(product,variant,asset,'image','/uploads/products/product-original.png');
INSERT INTO product_media(product_id,media_type,url) VALUES(product,'image','/company/products/bundled-product-photo.png');
INSERT INTO tickets(ticket_number,user_id,customer_user_id,conversation_id,subject,category) VALUES('product-backup-chat',customer,customer,'product-backup-chat','Retained customer conversation','customer_service') RETURNING id INTO ticket;
INSERT INTO ticket_messages(ticket_id,user_id,content,attachments,attachment_reference_format) VALUES(ticket,customer,'Retained customer message','[]','media_asset_bindings');
INSERT INTO orders(order_number,user_id,currency,payment_currency,payment_amount_minor,subtotal_amount_minor,total_amount_minor,shipping_fee_minor,tax_amount_minor,discount_amount_minor,points_value_minor) VALUES('product-backup-order',customer,'USD','USD',12345,12345,12345,0,0,0,0) RETURNING id INTO order_id;
INSERT INTO order_items(order_id,product_id,variant_id,product_name,quantity,currency,price_minor,subtotal_minor,total_minor,tax_amount_minor,discount_minor) VALUES(order_id,product,variant,'Historical order description',1,'USD',12345,12345,12345,0,0);
INSERT INTO outbox_events(event_key,event_type,aggregate_type,aggregate_id,payload,status,processed_at) VALUES('product-backup-completed','order.completed','order',order_id::text,'{}','processed',now());
END $$`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `UPDATE media_assets SET size=$1 WHERE storage_key='products/product-original.png'`, len(originalImage))
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `UPDATE media_asset_derivatives SET size=$1 WHERE storage_key='products/product-small.png'`, len(thumbnailImage))
	require.NoError(t, err)
	for key, body := range map[string][]byte{"products/product-original.png": originalImage, "products/product-small.png": thumbnailImage, "unselected/retained.txt": []byte("retained-public-file")} {
		filename := filepath.Join(cfg.StorageRoots["uploads"], filepath.FromSlash(key))
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
		require.NoError(t, os.WriteFile(filename, body, 0600))
	}
	require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads_private"], "retained-chat.txt"), []byte("retained private attachment"), 0600))
	require.NoError(t, os.MkdirAll(filepath.Join(cfg.StorageRoots["uploads_private"], "retained-empty-directory", "nested-empty"), 0700))
	seedStoppedProductMediaPrivateGuestConversationBeforeRecoverySnapshots(t, db, cfg)
	seedStoppedProductMediaMemberAndStaffAccountsBeforeRecoverySnapshots(t, db)
	seedStoppedStorefrontDatabaseMarketSettingsAndWeightedProductShippingBeforeSnapshots(t, db)
	seedStoppedStorefrontPublishedPageShellSettingsRedirectRulesAndProductFAQ(t, db)
	seedStoppedStorefrontApprovedPendingRejectedAndDeletedReviewPaginationBeforeSnapshots(t, db)
	seedStoppedProductCurrentSiteLogoAndFaviconBeforeRecoverySnapshots(t, db, cfg)
	verifyStoppedStorefrontComposedDependenciesInReadOnlySQLBeforeBackupRoundTrip(t, cfg, db)
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	snapshot, err := cfg.captureProductCatalogModuleSnapshot(ctx, tx)
	require.NoError(t, err)
	require.NotEmpty(t, snapshot.ExternalDependencies)
	require.NotContains(t, snapshot.Tables, "customs_classification_profiles")
	require.NoError(t, tx.Rollback())
	references, err := cfg.collectProductCatalogMediaReferences(snapshot)
	require.NoError(t, err)
	for reference := range references {
		if strings.HasPrefix(reference, "/company/") {
			filename := filepath.Join(cfg.ProductStaticAssetsDirectory, filepath.FromSlash(strings.TrimPrefix(reference, "/")))
			require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
			require.NoError(t, os.WriteFile(filename, originalImage, 0600))
		}
	}
	require.NoError(t, os.WriteFile(filepath.Join(cfg.ProductStaticAssetsDirectory, "company", companyStaticReleaseVersionFilename), []byte(cfg.ApplicationVersion), 0600))
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer func() {
		tasks, _ := manager.ListTasks()
		for _, task := range tasks {
			if task.StageDatabase != "" {
				require.NoError(t, cfg.dropOwnedStageDatabase(ctx, task.StageDatabase))
			}
		}
		require.NoError(t, manager.Close())
	}()
	backup, err := manager.CreateProductCatalogBackup("synthetic-operator", "BACKUP PRODUCTS "+cfg.SiteID+"/"+cfg.Environment)
	require.NoError(t, err)
	backup = waitForIndependentRecoveryTask(t, manager, backup.ID)
	require.Equal(t, "ready_to_download", backup.State, backup.Error)
	require.Equal(t, productCatalogBackupScope, backup.Manifest.Scope)
	packagePath, _ := manager.taskPackagePath(backup.ID)
	data, err := os.ReadFile(packagePath)
	require.NoError(t, err)
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	exportToken, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-exporter", []string{PermissionExport})
	require.NoError(t, err)
	downloadRequest := httptest.NewRequestWithContext(context.Background(), "GET", "/downloads/"+backup.ID, nil)
	downloadRequest.Header.Set("Authorization", "Bearer "+exportToken)
	downloadResponse := httptest.NewRecorder()
	manager.HTTPHandler().ServeHTTP(downloadResponse, downloadRequest)
	require.Equal(t, 200, downloadResponse.Code)
	require.Contains(t, downloadResponse.Header().Get("Content-Disposition"), "product-catalog-with-images-backup-")
	require.Equal(t, data, downloadResponse.Body.Bytes())
	passphrase, err := readProtectedCredential(cfg.PassphraseFile)
	require.NoError(t, err)
	_, err = ExtractAndVerifyEncryptedPackage(packagePath, filepath.Join(t.TempDir(), "whole-extractor-refuses-product"), passphrase, cfg.MaxExpandedBytes, cfg.MaxFiles)
	require.Error(t, err)
	extracted := filepath.Join(t.TempDir(), "verified-product-package")
	manifest, err := extractAndVerifyEncryptedPackageWithAllowedScopes(packagePath, extracted, passphrase, cfg.MaxExpandedBytes, cfg.MaxFiles, true)
	require.NoError(t, err)
	require.NotContains(t, manifest.Tables, "orders")
	require.NotContains(t, manifest.Tables, "users")
	require.NotContains(t, manifest.Tables, "product_view_count_flush_receipts")
	for _, filename := range []string{"database.dump", "recovery-secrets.json", "media/uploads_private/retained-chat.txt", "media/uploads/unselected/retained.txt"} {
		_, err := os.Stat(filepath.Join(extracted, filepath.FromSlash(filename)))
		require.True(t, os.IsNotExist(err), filename)
	}
	image, err := os.ReadFile(filepath.Join(extracted, "media/uploads/products/product-original.png"))
	require.NoError(t, err)
	require.Equal(t, originalImage, image)
	decoded, err := png.Decode(bytes.NewReader(image))
	require.NoError(t, err)
	require.Equal(t, 2, decoded.Bounds().Dx())
	_, err = db.ExecContext(ctx, `UPDATE products SET name='Current edited product',description='Current edited description',view_count=37 WHERE slug='product-backup-original'; UPDATE product_variants SET stock=7 WHERE sku='product-backup-sku'; DELETE FROM product_variants WHERE sku='product-backup-new-sku'; INSERT INTO products(name,slug,locale,parent_id,price_minor,currency) VALUES('Target-only product','target-only-product','fr',(SELECT id FROM products WHERE slug='product-backup-original' AND locale='en'),99,'USD'); INSERT INTO product_variants(product_id,sku,master_variant_id,price_minor,currency,stock) VALUES((SELECT id FROM products WHERE slug='target-only-product'),'target-only-translated-sku',(SELECT id FROM product_variants WHERE sku='product-backup-sku'),99,'USD',0); UPDATE ticket_messages SET content='Latest retained customer message' WHERE ticket_id=(SELECT id FROM tickets WHERE ticket_number='product-backup-chat')`)
	require.NoError(t, err)
	beforeTables, beforeFingerprint, _, err := captureDatabaseInventory(ctx, db)
	require.NoError(t, err)
	beforeMedia, err := cfg.mediaFingerprint(ctx)
	require.NoError(t, err)
	verified := uploadProductCatalogPackageForIntegrationTest(t, manager, data)
	require.Equal(t, "product_module_isolated_verified", verified.State, verified.Error)
	require.NotNil(t, verified.ProductCatalogRecovery)
	require.False(t, verified.ProductCatalogRecovery.OnlineOverwriteAvailable)
	require.NotNil(t, verified.ProductCatalogRecovery.InventoryPreservation)
	require.Equal(t, productCatalogCurrentInventoryPreservationPolicy, verified.ProductCatalogRecovery.InventoryPreservation.Policy)
	require.Equal(t, 2, verified.ProductCatalogRecovery.InventoryPreservation.ExistingVariantStocksPreserved)
	require.Equal(t, 1, verified.ProductCatalogRecovery.InventoryPreservation.DifferentSnapshotStockValuesNotApplied)
	require.Equal(t, 1, verified.ProductCatalogRecovery.InventoryPreservation.NewVariantStocksRestoredFromSnapshot)
	require.Equal(t, 2, verified.ProductCatalogRecovery.InventoryPreservation.ExistingProductViewCountsPreserved)
	require.True(t, verified.ProductCatalogRecovery.InventoryPreservation.ExistingInventoryAndIdentityFactsVerified)
	require.Greater(t, verified.ProductCatalogRecovery.PreservedTables, 100)
	stage, err := cfg.connectDatabase(ctx, verified.StageDatabase)
	require.NoError(t, err)
	defer stage.Close()
	var name, description, message string
	var stock int
	require.NoError(t, stage.QueryRowContext(ctx, `SELECT name,description FROM products WHERE slug='product-backup-original' AND locale='en'`).Scan(&name, &description))
	require.Equal(t, "Original product", name)
	require.Equal(t, "Original complete description", description)
	require.NoError(t, stage.QueryRowContext(ctx, `SELECT stock FROM product_variants WHERE sku='product-backup-sku'`).Scan(&stock))
	require.Equal(t, 7, stock, "a product-only restore must preserve current inventory after orders/refunds")
	require.NoError(t, stage.QueryRowContext(ctx, `SELECT stock FROM product_variants WHERE sku='product-backup-new-sku'`).Scan(&stock))
	require.Equal(t, 3, stock, "a new variant has no current inventory to replace; its snapshot quantity is reported separately")
	var currentViewCount int
	require.NoError(t, stage.QueryRowContext(ctx, `SELECT view_count FROM products WHERE slug='product-backup-original' AND locale='en'`).Scan(&currentViewCount))
	require.Equal(t, 37, currentViewCount)
	var retainedViewReceiptCount int
	require.NoError(t, stage.QueryRowContext(ctx, `SELECT count(*) FROM product_view_count_flush_receipts WHERE batch_id='11111111-1111-4111-8111-111111111111' AND delta=4 AND product_increment_applied`).Scan(&retainedViewReceiptCount))
	require.Equal(t, 1, retainedViewReceiptCount, "product-only recovery preserves the current SQL replay ledger")
	var translatedStock, retainedTranslatedStock int
	require.NoError(t, stage.QueryRowContext(ctx, `SELECT stock FROM product_variants WHERE sku='product-backup-translated-sku'`).Scan(&translatedStock))
	require.NoError(t, stage.QueryRowContext(ctx, `SELECT stock FROM product_variants WHERE sku='target-only-translated-sku'`).Scan(&retainedTranslatedStock))
	require.Zero(t, translatedStock)
	require.Zero(t, retainedTranslatedStock)
	require.NoError(t, stage.QueryRowContext(ctx, `SELECT content FROM ticket_messages WHERE ticket_id=(SELECT id FROM tickets WHERE ticket_number='product-backup-chat')`).Scan(&message))
	require.Equal(t, "Latest retained customer message", message)
	var targetOnly int
	require.NoError(t, stage.QueryRowContext(ctx, `SELECT count(*) FROM products WHERE slug='target-only-product'`).Scan(&targetOnly))
	require.Equal(t, 1, targetOnly)
	afterTables, afterFingerprint, _, err := captureDatabaseInventory(ctx, db)
	require.NoError(t, err)
	require.Equal(t, beforeTables, afterTables)
	require.Equal(t, beforeFingerprint, afterFingerprint)
	afterMedia, err := cfg.mediaFingerprint(ctx)
	require.NoError(t, err)
	require.Equal(t, beforeMedia, afterMedia)
	_, err = manager.ApplyVerifiedRecovery(verified.ID, "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+manifest.PackageID)
	require.Error(t, err)
	t.Logf("product module restored %d rows and %d actual images; preserved %d other tables, target-only products, and all live database/media fingerprints", verified.ProductCatalogRecovery.RestoredRows, verified.ProductCatalogRecovery.VerifiedImageFiles, verified.ProductCatalogRecovery.PreservedTables)
	verifyProductCatalogHandoffHTTPInspectsActualCandidateWithoutChangingCustodyDatabaseMediaOrTask(t, manager, verified)
	verifyProductCatalogStoppedRevalidationPairedSwitchAndEveryInterruptedRenameRollback(t, &manager, data)
	verifyProductCatalogCurrentInventoryStableIdentitiesAndInvalidSharedOwnerAtomicRollback(t, cfg, stage, snapshot)
	verifyProductCatalogRequiredImagesSharedRegistrationsAndRetainedCleanupReceiptAtomicRollback(t, manager, stage, snapshot, data)
	verifyProductCatalogStandardTriggerContractsBeforeAnyWriteAndSystemCategoryAtomicRollback(t, manager, stage, snapshot, data)
	verifyProductCatalogMediaPlacementOwnersAssetBindingsAndLegacyURLAtomicRollback(t, manager, stage, snapshot, data)
	verifyProductCatalogRichTextMediaTypesLegacyOccurrencesAndAtomicRollback(t, manager, stage, snapshot, data)
	verifyProductCatalogTemplateAndCategoryRichTextMediaTypesLegacyOccurrencesAndAtomicRollback(t, manager, stage, data)
	verifyProductViewStoppedRedisOwnershipDrainRestartRetryAndFreshCandidatePreservation(t, &manager, data)
	serveOptInDisposableProductCatalogMaintenanceBrowserFixture(t, manager)
	// A missing source image produces no downloadable package.
	imagePath := filepath.Join(cfg.StorageRoots["uploads"], "products/product-original.png")
	require.NoError(t, os.Remove(imagePath))
	missingBackup, err := manager.CreateProductCatalogBackup("synthetic-operator", "BACKUP PRODUCTS "+cfg.SiteID+"/"+cfg.Environment)
	require.NoError(t, err)
	missingBackup = waitForIndependentRecoveryTask(t, manager, missingBackup.ID)
	require.Equal(t, "failed", missingBackup.State)
	missingPackage, _ := manager.taskPackagePath(missingBackup.ID)
	_, err = os.Stat(missingPackage)
	require.True(t, os.IsNotExist(err))
	require.NoError(t, os.WriteFile(imagePath, originalImage, 0600))
	// Same-path conflicts and corrupt ciphertext must fail without online changes.
	require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads"], "products/product-original.png"), []byte("conflicting live image"), 0600))
	conflicting := uploadProductCatalogPackageForIntegrationTest(t, manager, data)
	require.Equal(t, "failed", conflicting.State)
	require.Contains(t, conflicting.Error, "conflicts")
	corrupt := append([]byte(nil), data...)
	corrupt[len(corrupt)-1] ^= 1
	failed := uploadProductCatalogPackageForIntegrationTest(t, manager, corrupt)
	require.Equal(t, "failed", failed.State)
	require.NoError(t, os.WriteFile(imagePath, originalImage, 0600))
	_, err = db.ExecContext(ctx, `UPDATE customs_classification_profiles SET name='Changed current shared customs profile' WHERE slug='product-backup-customs-profile'`)
	require.NoError(t, err)
	dependencyConflict := uploadProductCatalogPackageForIntegrationTest(t, manager, data)
	require.Equal(t, "failed", dependencyConflict.State)
	require.Contains(t, dependencyConflict.Error, "shared product dependencies differ")
	_, err = db.ExecContext(ctx, `CREATE TABLE product_images(id bigint PRIMARY KEY, url text)`)
	require.NoError(t, err)
	unregistered, err := manager.CreateProductCatalogBackup("synthetic-operator", "BACKUP PRODUCTS "+cfg.SiteID+"/"+cfg.Environment)
	require.NoError(t, err)
	unregistered = waitForIndependentRecoveryTask(t, manager, unregistered.ID)
	require.Equal(t, "failed", unregistered.State)
	require.Contains(t, unregistered.Error, "unregistered storefront table")
	require.Contains(t, unregistered.Error, "product_images")
	_, err = db.ExecContext(ctx, `DROP TABLE product_images`)
	require.NoError(t, err)
}

// Only the explicitly disposable integration fixture can expose this listener.
// Docker must publish its port on loopback; the production handler is unchanged.
func serveOptInDisposableProductCatalogMaintenanceBrowserFixture(t *testing.T, manager *Manager) {
	t.Helper()
	if os.Getenv("LOCAL_BACKUP_PRODUCT_CATALOG_BROWSER_FIXTURE") != "true" {
		return
	}
	done := make(chan struct{})
	var finish sync.Once
	mux := http.NewServeMux()
	mux.HandleFunc("/__synthetic-browser-fixture/restored-product-image", func(response http.ResponseWriter, request *http.Request) {
		tasks, err := manager.ListTasks()
		if err != nil {
			response.WriteHeader(500)
			return
		}
		for _, task := range tasks {
			if task.State == "product_module_isolated_verified" {
				directory, _ := manager.taskDirectory(task.ID)
				response.Header().Set("Content-Type", "image/png")
				http.ServeFile(response, request, filepath.Join(directory, "product-isolated-target/media/uploads/products/product-original.png"))
				return
			}
		}
		response.WriteHeader(404)
	})
	mux.HandleFunc("/__synthetic-browser-fixture/finish", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			response.WriteHeader(405)
			return
		}
		finish.Do(func() { close(done) })
		response.WriteHeader(204)
	})
	mux.Handle("/", manager.HTTPHandler())
	server := &http.Server{Addr: ":19093", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("synthetic browser fixture: %v", err)
			finish.Do(func() { close(done) })
		}
	}()
	t.Log("Disposable product catalog browser fixture ready on container port 19093")
	select {
	case <-done:
	case <-time.After(10 * time.Minute):
		t.Error("synthetic browser fixture was not finished within ten minutes")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, server.Shutdown(ctx))
}

func createDecodableProductCatalogPNGImageFixture(t *testing.T, size int, pixel color.RGBA) []byte {
	t.Helper()
	fixture := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			fixture.SetRGBA(x, y, pixel)
		}
	}
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, fixture))
	return encoded.Bytes()
}
