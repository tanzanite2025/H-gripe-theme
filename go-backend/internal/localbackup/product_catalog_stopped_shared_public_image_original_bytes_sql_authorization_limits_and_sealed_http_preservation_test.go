package localbackup

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"image/color"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestStoppedProductPublicImageCaptureBoundsCancellationAndPrivateDigestOnlySerialization(t *testing.T) {
	body := createDecodableProductCatalogPNGImageFixture(t, 1, color.RGBA{R: 180, A: 255})
	require.NoError(t, validateStoppedProductCapturedPublicImageResponse(200, map[string]string{"Content-Type": "image/png"}, body))
	for _, mimeType := range []string{"text/html", "image/svg+xml", "image/avif", "image/webp"} {
		require.Error(t, validateStoppedProductCapturedPublicImageResponse(200, map[string]string{"Content-Type": mimeType}, body))
	}
	require.Error(t, validateStoppedProductCapturedPublicImageResponse(200, map[string]string{"Content-Type": "image/png"}, []byte("<html>not an image</html>")))
	require.Error(t, validateStoppedProductCapturedPublicImageResponse(206, map[string]string{"Content-Type": "image/png"}, body))
	response := &stoppedProductMediaHashingHTTPResponse{header: make(http.Header), digest: sha256.New(), ctx: context.Background(), publicImageBody: &bytes.Buffer{}}
	n, err := response.Write(make([]byte, stoppedProductPublicImageMaximumBytes))
	require.NoError(t, err)
	require.Equal(t, stoppedProductPublicImageMaximumBytes, n)
	n, err = response.Write([]byte{1})
	require.Error(t, err)
	require.Zero(t, n)
	require.Equal(t, stoppedProductPublicImageMaximumBytes, response.publicImageBody.Len())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	response = &stoppedProductMediaHashingHTTPResponse{header: make(http.Header), digest: sha256.New(), ctx: ctx, publicImageBody: &bytes.Buffer{}}
	_, err = response.Write(body)
	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, response.publicImageBody.Bytes())
	encoded, err := json.Marshal(StoppedProductReadOnlyMediaHTTPResult{capturedPublicImageBytes: []byte("must-never-serialize-private-body")})
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "must-never-serialize-private-body")
	require.NotContains(t, string(encoded), "response_bytes_base64")
}

func TestPostgreSQLStoppedSharedPublicImageBytesUseRealAuthorizationAndPreserveAllTablesAndSequences(t *testing.T) {
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
	imageBytes := createDecodableProductCatalogPNGImageFixture(t, 2, color.RGBA{G: 170, A: 255})
	configuration.StoppedProductLocalMediaVerification = &StoppedProductLocalMediaVerificationConfiguration{SharedBaseURL: "https://synthetic.example.test", LogoBaseURL: "https://synthetic.example.test"}
	for _, key := range []string{"products/original.png", "products/private.png", "products/deleted.png", "products/unregistered.png", "products/invalid.png", "products/oversized.png", "site-logo/owned.png"} {
		body := imageBytes
		if key == "products/invalid.png" {
			body = []byte("<html>not an image</html>")
		}
		if key == "products/oversized.png" {
			body = append(append([]byte{}, imageBytes...), make([]byte, stoppedProductPublicImageMaximumBytes)...)
		}
		filename := filepath.Join(configuration.StorageRoots["uploads"], filepath.FromSlash(key))
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
		require.NoError(t, os.WriteFile(filename, body, 0600))
		if key == "products/unregistered.png" {
			continue
		}
		visibility := "public"
		if key == "products/private.png" {
			visibility = "private"
		}
		_, err = database.ExecContext(context.Background(), `INSERT INTO media_assets(filename,url,storage_key,size,mime_type,media_type,visibility,status) VALUES($1,$2,$1,$3,'image/png','image',$4,'active')`, key, "/uploads/"+key, len(body), visibility)
		require.NoError(t, err)
	}
	_, err = database.ExecContext(context.Background(), `UPDATE media_assets SET deleted_at=now() WHERE storage_key='products/deleted.png';
INSERT INTO media_asset_derivatives(media_asset_id,preset,url,storage_key,mime_type,size)
SELECT id,'small','/uploads/products/small.png','products/small.png','image/png',size FROM media_assets WHERE storage_key='products/original.png'`)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(configuration.StorageRoots["uploads"], "products/small.png"), imageBytes, 0600))
	manager := &Manager{Configuration: configuration}
	beforeCounts, beforeFingerprint, beforeMigration, err := captureDatabaseInventory(ctx, database)
	require.NoError(t, err)
	for _, scenario := range []struct {
		Name    string
		Key     string
		Status  int
		Refused bool
	}{
		{"registered_original", "products/original.png", 200, false}, {"registered_derivative", "products/small.png", 200, false},
		{"private_registration", "products/private.png", 404, false}, {"deleted_registration", "products/deleted.png", 404, false},
		{"existing_unregistered_file", "products/unregistered.png", 404, false}, {"unknown_file", "products/missing.png", 404, false},
		{"dedicated_owner_collision", "site-logo/owned.png", 404, false}, {"mismatched_mime_signature", "products/invalid.png", 0, true},
		{"oversized_image", "products/oversized.png", 0, true},
	} {
		t.Run(scenario.Name, func(t *testing.T) {
			tx, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
			require.NoError(t, err)
			defer tx.Rollback()
			result, err := manager.invokeStoppedProductMediaHandlersWithReadOnlyTransactionAndOptionalPublicImageCapture(ctx, tx, StoppedProductReadOnlyMediaHTTPRequest{Operation: "public_upload", ObjectKey: scenario.Key}, true)
			if scenario.Refused {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, scenario.Status, result.HTTPStatus)
				require.Equal(t, productCatalogBytesSHA256(result.capturedPublicImageBytes), result.ResponseSHA256)
				if scenario.Status == 200 {
					require.Equal(t, imageBytes, result.capturedPublicImageBytes)
				} else {
					require.NotContains(t, string(result.capturedPublicImageBytes), string(imageBytes))
				}
			}
			require.NoError(t, verifyStoppedProductSQLTransactionRemainsReadOnly(ctx, tx))
			require.NoError(t, tx.Rollback())
		})
	}
	afterCounts, afterFingerprint, afterMigration, err := captureDatabaseInventory(ctx, database)
	require.NoError(t, err)
	require.Equal(t, beforeCounts, afterCounts)
	require.Equal(t, beforeFingerprint, afterFingerprint)
	require.Equal(t, beforeMigration, afterMigration)
}

func verifyStoppedProductPublicImageActualNetworkBytesAndFailureResealingPreserveRecoveryBoundary(t *testing.T, manager *Manager, taskID string) {
	t.Helper()
	ctx := context.Background()
	before, err := manager.GetTask(taskID)
	require.NoError(t, err)
	baseline, err := manager.InspectStoppedProductStorefrontProjectionBaseline(ctx, taskID)
	require.NoError(t, err)
	key, err := readProtectedCredential(manager.Configuration.AuthenticationKeyFile)
	require.NoError(t, err)
	token, err := IssueMaintenanceSession(key, manager.Configuration.SiteID, manager.Configuration.Environment, "synthetic-public-image-network-verifier", []string{PermissionRestore})
	require.NoError(t, err)
	server := httptest.NewServer(manager.HTTPHandler())
	defer server.Close()
	client := &http.Client{Timeout: 2 * time.Minute}
	for _, scenario := range []struct {
		Name, ObjectKey string
		Status          int
	}{
		{"original_png_bytes", "products/product-original.png", 200}, {"derivative_png_bytes", "products/product-small.png", 200},
		{"private_attachment_direct_path", stoppedProductMediaSyntheticOriginalKey, 404}, {"unknown_public_file", "unselected/retained.txt", 404},
		{"dedicated_owner_refusal_and_reseal", "site-logo/unwired.png", 404},
		{"unsupported_owner_error_and_reseal", "customer-service/avatars/unwired.png", 0},
	} {
		t.Run("public_image_network_"+scenario.Name, func(t *testing.T) {
			query := StoppedProductReadOnlyPublicImageHTTPRequest{ContractSHA256: baseline.ContractSHA256, ObjectKey: scenario.ObjectKey}
			encoded, err := json.Marshal(query)
			require.NoError(t, err)
			request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL+"/api/tasks/"+taskID+"/product-catalog-stopped-read-only-public-image-http-query", bytes.NewReader(encoded))
			require.NoError(t, err)
			request.Header.Set("Authorization", "Bearer "+token)
			request.Header.Set("Content-Type", "application/json")
			response, err := client.Do(request)
			require.NoError(t, err)
			defer response.Body.Close()
			if scenario.Status == 0 {
				require.Equal(t, http.StatusConflict, response.StatusCode)
			} else {
				require.Equal(t, http.StatusOK, response.StatusCode)
				var payload struct {
					Result StoppedProductReadOnlyPublicImageHTTPResult `json:"result"`
				}
				require.NoError(t, json.NewDecoder(response.Body).Decode(&payload))
				result := payload.Result
				require.Equal(t, scenario.Status, result.HTTPStatus)
				require.Equal(t, query, result.Request)
				require.Equal(t, baseline.ContractSHA256, result.ContractSHA256)
				require.Equal(t, baseline.DatabaseOID, result.DatabaseOID)
				require.True(t, result.SQLReadOnly)
				require.True(t, result.ApplicationRemainsStopped)
				require.False(t, result.OnlineOverwriteAvailable)
				body, err := base64.StdEncoding.DecodeString(result.ResponseBytesBase64)
				require.NoError(t, err)
				require.Equal(t, productCatalogBytesSHA256(body), result.ResponseSHA256)
				require.Equal(t, int64(len(body)), result.ResponseBytes)
				if scenario.Status == 200 {
					expected, err := os.ReadFile(filepath.Join(manager.Configuration.StorageRoots["uploads"], filepath.FromSlash(query.ObjectKey)))
					require.NoError(t, err)
					require.Equal(t, expected, body)
					require.NotNil(t, result.MatchesPackageFile)
					require.True(t, *result.MatchesPackageFile)
				} else {
					require.NotContains(t, string(body), "private-original-fixture-v1")
				}
			}
			// A new full inspection also verifies that SQL was resealed after
			// successful reads, authorization denials and errors inside the tx.
			afterBaseline, err := manager.InspectStoppedProductStorefrontProjectionBaseline(ctx, taskID)
			require.NoError(t, err)
			require.Equal(t, baseline.ContractSHA256, afterBaseline.ContractSHA256)
		})
	}
	cancelledContext, cancel := context.WithCancel(ctx)
	cancel()
	result, err := manager.ExecuteStoppedProductReadOnlyPublicImageHTTPQuery(cancelledContext, taskID, StoppedProductReadOnlyPublicImageHTTPRequest{ContractSHA256: baseline.ContractSHA256, ObjectKey: "products/product-original.png"})
	require.Error(t, err)
	require.Nil(t, result)
	result, err = manager.ExecuteStoppedProductReadOnlyPublicImageHTTPQuery(ctx, taskID, StoppedProductReadOnlyPublicImageHTTPRequest{ContractSHA256: strings.Repeat("f", 64), ObjectKey: "products/product-original.png"})
	require.Error(t, err)
	require.Nil(t, result)
	after, err := manager.GetTask(taskID)
	require.NoError(t, err)
	require.Equal(t, before, after)
	afterBaseline, err := manager.InspectStoppedProductStorefrontProjectionBaseline(ctx, taskID)
	require.NoError(t, err)
	require.Equal(t, baseline.ContractSHA256, afterBaseline.ContractSHA256)
}
