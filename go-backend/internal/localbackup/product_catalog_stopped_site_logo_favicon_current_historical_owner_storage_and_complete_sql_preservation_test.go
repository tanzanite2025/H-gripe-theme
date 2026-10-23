package localbackup

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"image/color"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func seedStoppedProductCurrentSiteLogoAndFaviconBeforeRecoverySnapshots(t *testing.T, database *sql.DB, configuration Configuration) map[string][]byte {
	t.Helper()
	bodies := map[string][]byte{
		"site-logo/stopped-current.png":    createDecodableProductCatalogPNGImageFixture(t, 512, color.RGBA{R: 190, B: 70, A: 255}),
		"site-favicon/stopped-current.png": createDecodableProductCatalogPNGImageFixture(t, 512, color.RGBA{G: 190, B: 70, A: 255}),
	}
	for key, body := range bodies {
		table, root := "site_favicon_assets", configuration.StorageRoots["uploads"]
		if stoppedProductPublicSiteIdentityImageOwnerForObjectKey(key) == "site_logo" {
			table, root = "site_logo_assets", configuration.StorageRoots["site_logo_uploads"]
			decoy := filepath.Join(configuration.StorageRoots["uploads"], filepath.FromSlash(key))
			require.NoError(t, os.MkdirAll(filepath.Dir(decoy), 0700))
			require.NoError(t, os.WriteFile(decoy, []byte("same logo key in the wrong shared directory"), 0600))
		}
		filename := filepath.Join(root, filepath.FromSlash(key))
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
		require.NoError(t, os.WriteFile(filename, body, 0600))
		_, err := database.ExecContext(context.Background(), `INSERT INTO `+table+`(id,filename,url,storage_key,mime_type,size,content_sha256,width,height) VALUES(1,$1,$2,$1,'image/png',$3,$4,512,512)`, key, "/uploads/"+key, len(body), productCatalogBytesSHA256(body))
		require.NoError(t, err)
	}
	return bodies
}

func TestPostgreSQLStoppedSiteIdentityImagesUseCurrentHistoricalOwnersSeparateRootsAndPreserveAllSQL(t *testing.T) {
	connectionURL := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL")
	migrationsDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || migrationsDirectory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires an explicitly disposable PostgreSQL server and complete application migrations")
	}
	ctx := context.Background()
	configuration := createIndependentRecoveryTestConfiguration(t)
	configuration.DatabaseURL, configuration.DatabaseName = connectionURL, "local_backup_test_"+newTaskID()[:12]
	control, err := configuration.connectDatabase(ctx, "postgres")
	require.NoError(t, err)
	defer control.Close()
	_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(configuration.DatabaseName))
	require.NoError(t, err)
	defer func() {
		_, _ = control.ExecContext(context.Background(), `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(configuration.DatabaseName)+` WITH (FORCE)`)
	}()
	database, err := configuration.connectDatabase(ctx, configuration.DatabaseName)
	require.NoError(t, err)
	driver, err := migrationpostgres.WithInstance(database, &migrationpostgres.Config{})
	require.NoError(t, err)
	migrator, err := migrate.NewWithDatabaseInstance("file://"+filepath.ToSlash(migrationsDirectory), "postgres", driver)
	require.NoError(t, err)
	if err := migrator.Up(); err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
	sourceCloseError, databaseCloseError := migrator.Close()
	require.NoError(t, sourceCloseError)
	require.NoError(t, databaseCloseError)
	database, err = configuration.connectDatabase(ctx, configuration.DatabaseName)
	require.NoError(t, err)
	defer database.Close()
	configuration.StoppedProductLocalMediaVerification = &StoppedProductLocalMediaVerificationConfiguration{SharedBaseURL: "https://synthetic.example.test", LogoBaseURL: "https://logos.example.test"}
	bodies := seedStoppedProductCurrentSiteLogoAndFaviconBeforeRecoverySnapshots(t, database, configuration)
	_, err = database.ExecContext(context.Background(), `INSERT INTO users(email,username,password) VALUES('stopped-site-identity-author@example.invalid','stopped-site-identity-author','synthetic-password')`)
	require.NoError(t, err)
	for _, namespace := range []string{"site-logo", "site-favicon"} {
		root := configuration.StorageRoots["uploads"]
		if namespace == "site-logo" {
			root = configuration.StorageRoots["site_logo_uploads"]
		}
		for _, suffix := range []string{"published-old", "draft-old", "orphaned", "published-own-origin", "published-other-origin"} {
			key := namespace + "/" + suffix + ".png"
			body := createDecodableProductCatalogPNGImageFixture(t, 2, color.RGBA{R: 80, G: 40, B: 210, A: 255})
			bodies[key] = body
			filename := filepath.Join(root, filepath.FromSlash(key))
			require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
			require.NoError(t, os.WriteFile(filename, body, 0600))
			// A shared registration is deliberately present even for orphaned
			// dedicated objects; only their owning service may grant access.
			_, err = database.ExecContext(context.Background(), `INSERT INTO media_assets(filename,url,storage_key,size,mime_type,media_type,visibility,status) VALUES($1,$2,$1,$3,'image/png','image','public','active')`, key, "/uploads/"+key, len(body))
			require.NoError(t, err)
			if suffix != "orphaned" {
				status := "draft"
				if suffix != "draft-old" {
					status = "published"
				}
				reference := "/uploads/" + key
				if suffix == "published-own-origin" {
					origin := configuration.StoppedProductLocalMediaVerification.SharedBaseURL
					if namespace == "site-logo" {
						origin = configuration.StoppedProductLocalMediaVerification.LogoBaseURL
					}
					reference = origin + reference
				}
				if suffix == "published-other-origin" {
					origin := configuration.StoppedProductLocalMediaVerification.LogoBaseURL
					if namespace == "site-logo" {
						origin = configuration.StoppedProductLocalMediaVerification.SharedBaseURL
					}
					reference = origin + reference
				}
				_, err = database.ExecContext(context.Background(), `INSERT INTO posts(title,slug,content,status,author_id) SELECT $1,$1,$2,$3,id FROM users WHERE username='stopped-site-identity-author'`, namespace+"-"+suffix, `<img src="`+reference+`">`, status)
				require.NoError(t, err)
			}
		}
	}
	manager := &Manager{Configuration: configuration}
	verifyStoppedSiteIdentityReadWithoutSQLMutation := func(t *testing.T, key string, expectedStatus int) {
		t.Helper()
		tx, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
		require.NoError(t, err)
		defer tx.Rollback()
		result, err := manager.invokeStoppedProductMediaHandlersWithReadOnlyTransactionAndOptionalPublicImageCapture(ctx, tx, StoppedProductReadOnlyMediaHTTPRequest{Operation: "public_upload", ObjectKey: key}, true)
		require.NoError(t, err)
		require.Equal(t, expectedStatus, result.HTTPStatus)
		if expectedStatus == 200 {
			require.Equal(t, bodies[key], result.capturedPublicImageBytes)
			require.Equal(t, productCatalogBytesSHA256(bodies[key]), result.ResponseSHA256)
		} else {
			require.Empty(t, result.capturedPublicImageBytes)
		}
		require.NoError(t, verifyStoppedProductSQLTransactionRemainsReadOnly(ctx, tx))
		require.NoError(t, tx.Rollback())
	}
	beforeCounts, beforeFingerprint, beforeMigration, err := captureDatabaseInventory(ctx, database)
	require.NoError(t, err)
	for _, scenario := range []struct {
		Name, Key string
		Status    int
	}{
		{"current_logo_uses_logo_root_not_shared_decoy", "site-logo/stopped-current.png", 200},
		{"current_favicon_uses_shared_root", "site-favicon/stopped-current.png", 200},
		{"published_old_logo", "site-logo/published-old.png", 200}, {"published_old_favicon", "site-favicon/published-old.png", 200},
		{"draft_old_logo", "site-logo/draft-old.png", 404}, {"draft_old_favicon", "site-favicon/draft-old.png", 404},
		{"shared_registration_cannot_authorize_logo", "site-logo/orphaned.png", 404}, {"shared_registration_cannot_authorize_favicon", "site-favicon/orphaned.png", 404},
		{"unknown_logo", "site-logo/unknown.png", 404}, {"unknown_favicon", "site-favicon/unknown.png", 404},
		{"published_own_logo_origin", "site-logo/published-own-origin.png", 200}, {"published_own_favicon_origin", "site-favicon/published-own-origin.png", 200},
		{"published_shared_origin_cannot_authorize_logo", "site-logo/published-other-origin.png", 404}, {"published_logo_origin_cannot_authorize_favicon", "site-favicon/published-other-origin.png", 404},
	} {
		t.Run(scenario.Name, func(t *testing.T) { verifyStoppedSiteIdentityReadWithoutSQLMutation(t, scenario.Key, scenario.Status) })
	}
	afterCounts, afterFingerprint, afterMigration, err := captureDatabaseInventory(ctx, database)
	require.NoError(t, err)
	require.Equal(t, beforeCounts, afterCounts)
	require.Equal(t, beforeFingerprint, afterFingerprint)
	require.Equal(t, beforeMigration, afterMigration)
	_, err = database.ExecContext(context.Background(), `UPDATE posts SET status='draft' WHERE slug IN ('site-logo-published-old','site-favicon-published-old')`)
	require.NoError(t, err)
	beforeCounts, beforeFingerprint, beforeMigration, err = captureDatabaseInventory(ctx, database)
	require.NoError(t, err)
	for _, namespace := range []string{"site-logo", "site-favicon"} {
		t.Run("revoked_published_consumer_"+namespace, func(t *testing.T) {
			verifyStoppedSiteIdentityReadWithoutSQLMutation(t, namespace+"/published-old.png", 404)
		})
	}
	afterCounts, afterFingerprint, afterMigration, err = captureDatabaseInventory(ctx, database)
	require.NoError(t, err)
	require.Equal(t, beforeCounts, afterCounts)
	require.Equal(t, beforeFingerprint, afterFingerprint)
	require.Equal(t, beforeMigration, afterMigration)
	_, err = database.ExecContext(context.Background(), `UPDATE site_favicon_assets SET storage_key='site-favicon/missing.png',url='/uploads/site-favicon/missing.png' WHERE id=1`)
	require.NoError(t, err)
	beforeCounts, beforeFingerprint, beforeMigration, err = captureDatabaseInventory(ctx, database)
	require.NoError(t, err)
	t.Run("current_favicon_missing_file", func(t *testing.T) {
		verifyStoppedSiteIdentityReadWithoutSQLMutation(t, "site-favicon/missing.png", 404)
	})
	t.Run("legacy_digest_only_contract_still_rejects_dedicated_owner", func(t *testing.T) {
		tx, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
		require.NoError(t, err)
		defer tx.Rollback()
		result, err := manager.invokeStoppedProductMediaHandlersWithReadOnlyTransaction(ctx, tx, StoppedProductReadOnlyMediaHTTPRequest{Operation: "public_upload", ObjectKey: "site-logo/stopped-current.png"})
		require.Error(t, err)
		require.Nil(t, result)
	})
	afterCounts, afterFingerprint, afterMigration, err = captureDatabaseInventory(ctx, database)
	require.NoError(t, err)
	require.Equal(t, beforeCounts, afterCounts)
	require.Equal(t, beforeFingerprint, afterFingerprint)
	require.Equal(t, beforeMigration, afterMigration)
}

func verifyStoppedProductCurrentSiteIdentitySettingsAndImageBytesWithinSameRecoveryWindow(t *testing.T, manager *Manager, taskID string) {
	t.Helper()
	ctx := context.Background()
	before, err := manager.GetTask(taskID)
	require.NoError(t, err)
	baseline, err := manager.InspectStoppedProductStorefrontProjectionBaseline(ctx, taskID)
	require.NoError(t, err)
	credential, err := readProtectedCredential(manager.Configuration.AuthenticationKeyFile)
	require.NoError(t, err)
	token, err := IssueMaintenanceSession(credential, manager.Configuration.SiteID, manager.Configuration.Environment, "synthetic-site-identity-image-network-verifier", []string{PermissionRestore})
	require.NoError(t, err)
	server := httptest.NewServer(manager.HTTPHandler())
	defer server.Close()
	client := &http.Client{Timeout: 2 * time.Minute}
	invokeStoppedSiteIdentityMaintenanceHTTPRead := func(t *testing.T, operation string, body any, payload any) {
		t.Helper()
		encoded, err := json.Marshal(body)
		require.NoError(t, err)
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/api/tasks/"+taskID+"/"+operation, bytes.NewReader(encoded))
		require.NoError(t, err)
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("Content-Type", "application/json")
		response, err := client.Do(request)
		require.NoError(t, err)
		defer response.Body.Close()
		require.Equal(t, http.StatusOK, response.StatusCode)
		require.NoError(t, json.NewDecoder(response.Body).Decode(payload))
	}
	t.Run("site_identity_settings_current_owner_urls", func(t *testing.T) {
		var payload struct {
			Result StoppedProductReadOnlyStorefrontHTTPResult
		}
		invokeStoppedSiteIdentityMaintenanceHTTPRead(t, "product-catalog-stopped-read-only-storefront-http-query", StoppedProductReadOnlyStorefrontHTTPRequest{ContractSHA256: baseline.ContractSHA256, Operation: "site_settings", Locale: "en", Country: "AQ", Currency: "USD"}, &payload)
		require.Contains(t, string(payload.Result.ResponseBytesBase64), "/uploads/site-logo/stopped-current.png")
		require.Contains(t, string(payload.Result.ResponseBytesBase64), "/uploads/site-favicon/stopped-current.png")
	})
	for _, key := range []string{"site-logo/stopped-current.png", "site-favicon/stopped-current.png"} {
		t.Run("site_identity_original_bytes_"+stoppedProductPublicSiteIdentityImageOwnerForObjectKey(key), func(t *testing.T) {
			var payload struct {
				Result StoppedProductReadOnlyPublicImageHTTPResult
			}
			invokeStoppedSiteIdentityMaintenanceHTTPRead(t, "product-catalog-stopped-read-only-public-image-http-query", StoppedProductReadOnlyPublicImageHTTPRequest{ContractSHA256: baseline.ContractSHA256, ObjectKey: key}, &payload)
			result := payload.Result
			require.Equal(t, 200, result.HTTPStatus)
			require.Equal(t, stoppedProductPublicSiteIdentityImageOwnerForObjectKey(key), result.ObjectOwner)
			require.Equal(t, stoppedProductSiteIdentityImageVerification, result.Verification)
			require.Equal(t, baseline.ContractSHA256, result.ContractSHA256)
			root := manager.Configuration.StorageRoots["uploads"]
			if result.ObjectOwner == "site_logo" {
				root = manager.Configuration.StorageRoots["site_logo_uploads"]
			}
			expected, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(key)))
			require.NoError(t, err)
			body, err := base64.StdEncoding.DecodeString(result.ResponseBytesBase64)
			require.NoError(t, err)
			require.Equal(t, expected, body)
			require.False(t, result.OnlineOverwriteAvailable)
		})
	}
	after, err := manager.GetTask(taskID)
	require.NoError(t, err)
	require.Equal(t, before, after)
	current, err := manager.InspectStoppedProductStorefrontProjectionBaseline(ctx, taskID)
	require.NoError(t, err)
	require.Equal(t, baseline.ContractSHA256, current.ContractSHA256)
}
