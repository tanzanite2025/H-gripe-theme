package localbackup

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/pkg/maintenancegate"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func verifyProductViewStoppedRedisOwnershipDrainRestartRetryAndFreshCandidatePreservation(t *testing.T, managerReference **Manager, packageBytes []byte) {
	t.Helper()
	if os.Getenv("LOCAL_BACKUP_INTEGRATION_REDIS_URL") == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_REDIS_TEST_SERVER") != "true" {
		return
	}
	t.Run("product_view_stopped_owned_redis_drain_restart_retry_and_fresh_candidate", func(t *testing.T) {
		ctx := context.Background()
		manager := *managerReference
		originalConfiguration := manager.Configuration
		boundaryPath := filepath.Join(originalConfiguration.MaintenanceDirectory, "application-boundary.json")
		originalBoundary, err := os.ReadFile(boundaryPath)
		require.NoError(t, err)
		options, err := redis.ParseURL(os.Getenv("LOCAL_BACKUP_INTEGRATION_REDIS_URL"))
		require.NoError(t, err)
		client := redis.NewClient(options)
		keys := []string{maintenancegate.ProductViewRedisOwnershipKey, "product:view_counts", "product:view_counts:processing", "product:view_counts:processing_batch_id"}
		existing, err := client.Exists(ctx, keys...).Result()
		require.NoError(t, err)
		require.Zero(t, existing, "disposable Redis fixture must start without product ownership or buffers")
		t.Cleanup(func() {
			cleanupClient := redis.NewClient(options)
			defer cleanupClient.Close()
			require.NoError(t, cleanupClient.Del(context.Background(), keys...).Err())
			_ = client.Close()
			(*managerReference).Configuration = originalConfiguration
			require.NoError(t, os.WriteFile(boundaryPath, originalBoundary, 0600))
		})
		cfg := manager.Configuration
		cfg.StoppedProductLocalMediaVerification = &StoppedProductLocalMediaVerificationConfiguration{SharedBaseURL: "https://synthetic.example.test", LogoBaseURL: "https://synthetic.example.test"}
		cfg.ProductViewRecoveryRedis = &ProductViewRecoveryRedisConnectionConfiguration{Mode: "standalone", Addresses: []string{options.Addr}, Database: options.DB, Username: options.Username, Password: options.Password}
		cfg.ProductViewRecoveryWriterInventory = []string{"synthetic API and its workers and schedulers using this one gate", "synthetic migration and importer commands using this one gate"}
		manager.Configuration = cfg
		projection := syntheticExternalProjectionBoundaryForRecoveryReleaseTests()
		productProjection := service.DescribeProductCatalogRuntimeProjectionBoundaryForRecovery(true)
		projection.ProductCatalog = &productProjection
		projection.MediaDeliveryConfigurationSHA256 = maintenancegate.FingerprintExternalProjectionConfiguration([]string{"https://synthetic.example.test", "", "https://synthetic.example.test", ""})
		projection.RedisDestinationSHA256 = maintenancegate.FingerprintRedisRecoveryDestination("standalone", []string{options.Addr}, options.DB, "")
		require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets, projection))
		registerSyntheticProductViewRedisOwnershipBeforeStartingAnyWriters := func() error {
			return maintenancegate.RegisterProductViewRedisRuntimeOwnershipBeforeStartingWriters(ctx, cfg.MaintenanceDirectory, cfg.SiteID, cfg.Environment, cfg.DatabaseName, projection.RedisDestinationSHA256, productProjection.ViewCountKeysSHA256, client)
		}
		require.NoError(t, client.HSet(ctx, keys[1], "1", "2").Err())
		require.ErrorContains(t, registerSyntheticProductViewRedisOwnershipBeforeStartingAnyWriters(), "enrollment refused")
		require.NoError(t, client.Del(ctx, keys[1]).Err()) // only synthetic data created above
		require.NoError(t, registerSyntheticProductViewRedisOwnershipBeforeStartingAnyWriters())
		otherClient := redis.NewClient(options)
		require.NoError(t, otherClient.Ping(ctx).Err())
		require.ErrorContains(t, requireNoOtherProductViewRedisLogicalDatabaseClients(ctx, client, options.DB), "other Redis clients")
		require.NoError(t, otherClient.Close())
		subscriptionClient := redis.NewClient(options)
		subscription := subscriptionClient.Subscribe(ctx, "synthetic-product-view-stopped-inspection")
		_, err = subscription.Receive(ctx)
		require.NoError(t, err)
		require.ErrorContains(t, requireNoOtherProductViewRedisLogicalDatabaseClients(ctx, client, options.DB), "other Redis clients")
		require.NoError(t, subscription.Close())
		require.NoError(t, subscriptionClient.Close())
		ownership, err := maintenancegate.ReadRegisteredProductViewRedisRuntimeOwnership(cfg.MaintenanceDirectory)
		require.NoError(t, err)
		require.NoError(t, maintenancegate.VerifyRegisteredProductViewRedisRuntimeOwnership(ctx, client, ownership))
		wrongInstance := *ownership
		wrongInstance.ServerRunIDSHA256 = strings.Repeat("0", 64)
		require.ErrorContains(t, maintenancegate.VerifyRegisteredProductViewRedisRuntimeOwnership(ctx, client, &wrongInstance), "instance restarted or changed")
		wrongDatabaseOptions := *options
		wrongDatabaseOptions.DB = options.DB + 1
		wrongDatabaseClient := redis.NewClient(&wrongDatabaseOptions)
		require.ErrorContains(t, maintenancegate.VerifyRegisteredProductViewRedisRuntimeOwnership(ctx, wrongDatabaseClient, ownership), "marker mismatch")
		require.NoError(t, wrongDatabaseClient.Close())
		marker, err := client.Get(ctx, keys[0]).Result()
		require.NoError(t, err)
		require.NoError(t, client.Del(ctx, keys[0]).Err())
		require.ErrorContains(t, registerSyntheticProductViewRedisOwnershipBeforeStartingAnyWriters(), "marker changed or disappeared")
		require.NoError(t, client.Set(ctx, keys[0], marker, 0).Err())
		require.NoError(t, registerSyntheticProductViewRedisOwnershipBeforeStartingAnyWriters())
		live, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
		require.NoError(t, err)
		defer live.Close()
		var productID uint
		var beforeViews int64
		require.NoError(t, live.QueryRowContext(ctx, `SELECT id,view_count FROM products WHERE slug='product-backup-original' AND locale='en'`).Scan(&productID, &beforeViews))
		database, err := gorm.Open(postgres.New(postgres.Config{Conn: live}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), DisableAutomaticPing: true})
		require.NoError(t, err)
		batchID := uuid.NewString()
		applied, err := repository.NewProductRepository(database).ApplyProductViewCountBatchWithAtomicReplayReceipts(ctx, batchID, map[uint]int64{productID: 7})
		require.NoError(t, err)
		require.Equal(t, 1, applied)
		require.NoError(t, client.HSet(ctx, keys[2], strconv.FormatUint(uint64(productID), 10), "7").Err())
		require.NoError(t, client.Set(ctx, keys[3], batchID, 0).Err())
		require.NoError(t, client.HSet(ctx, keys[1], strconv.FormatUint(uint64(productID), 10), "5").Err())
		candidate := uploadProductCatalogPackageForIntegrationTest(t, manager, packageBytes)
		require.Equal(t, "product_module_isolated_verified", candidate.State, candidate.Error)
		oldOID := candidate.ProductCatalogPreviewBinding.Candidate.OID
		// Legitimate target rows advance after the original preview. The stopped
		// operation must rebuild from these latest rows instead of rejecting them.
		_, err = live.ExecContext(ctx, `UPDATE products SET view_count=view_count+3 WHERE id=$1`, productID)
		require.NoError(t, err)
		require.NoError(t, client.Del(ctx, keys[3]).Err())
		require.NoError(t, client.Close())
		key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
		require.NoError(t, err)
		token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-stopped-view-operator", []string{PermissionRestore})
		require.NoError(t, err)
		startStoppedProductViewDrainThroughAuthenticatedMaintenanceHTTPAndWaitForResult := func() *Task {
			body, _ := json.Marshal(ProductViewStoppedDrainAndCandidateRebuildRequest{Confirmation: "DRAIN PRODUCT VIEWS " + cfg.SiteID + "/" + cfg.Environment + "/" + candidate.ID, AllListedWritersStoppedAndUseThisSingleHostGate: true})
			request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/tasks/"+candidate.ID+"/product-view-stopped-drain-and-rebuild", strings.NewReader(string(body)))
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			manager.HTTPHandler().ServeHTTP(response, request)
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			return waitForIndependentRecoveryTask(t, manager, candidate.ID)
		}
		failed := startStoppedProductViewDrainThroughAuthenticatedMaintenanceHTTPAndWaitForResult()
		require.Equal(t, "failed", failed.State, failed.Error)
		require.NotNil(t, failed.ProductViewStoppedDrain)
		require.Equal(t, "draining", failed.ProductViewStoppedDrain.Phase)
		require.Nil(t, failed.ProductCatalogRecovery)
		require.Error(t, manager.ReleaseUnmodifiedMaintenance(candidate.ID))
		_, err = manager.VerifyUploadedPackage(candidate.ID)
		require.Error(t, err)
		require.NoError(t, manager.Close())
		manager, err = NewManager(cfg)
		require.NoError(t, err)
		*managerReference = manager
		retained, err := manager.GetTask(candidate.ID)
		require.NoError(t, err)
		require.Equal(t, "interrupted", retained.State)
		require.Equal(t, failed.ProductViewStoppedDrain.SHA256, retained.ProductViewStoppedDrain.SHA256)
		// An unfinished drain must remain recoverable after its work-copy TTL.
		retained.ExpiresAt = time.Now().Add(-time.Hour)
		require.NoError(t, manager.saveTask(retained))
		client = redis.NewClient(options)
		// Restore the original fixture identity from its SQL receipt and known
		// source. Production code never invents an identity for legacy buffers.
		require.NoError(t, client.Set(ctx, keys[3], batchID, 0).Err())
		require.NoError(t, client.Close())
		completed := startStoppedProductViewDrainThroughAuthenticatedMaintenanceHTTPAndWaitForResult()
		require.Equal(t, "product_module_isolated_verified", completed.State, completed.Error)
		require.Equal(t, "candidate_rebuilt", completed.ProductViewStoppedDrain.Phase)
		require.True(t, completed.ProductViewStoppedDrain.After.RedisBuffersEmpty)
		require.Equal(t, 1, completed.ProductViewStoppedDrain.Before.CommittedProcessingReceipts)
		require.Equal(t, 2, completed.ProductViewStoppedDrain.FlushCalls)
		require.NotEqual(t, oldOID, completed.ProductCatalogPreviewBinding.Candidate.OID)
		require.False(t, completed.ProductCatalogRecovery.OnlineOverwriteAvailable)
		require.Error(t, validateFreshProductCatalogCandidateTaskForHandoffInspection(completed), "finishing an expired drain does not renew the candidate or authorize handoff")
		var currentViews, candidateViews int64
		require.NoError(t, live.QueryRowContext(ctx, `SELECT view_count FROM products WHERE id=$1`, productID).Scan(&currentViews))
		require.Equal(t, beforeViews+15, currentViews, "the committed processing batch must not be counted twice")
		stage, err := cfg.connectDatabase(ctx, completed.StageDatabase)
		require.NoError(t, err)
		defer stage.Close()
		require.NoError(t, stage.QueryRowContext(ctx, `SELECT view_count FROM products WHERE id=$1`, productID).Scan(&candidateViews))
		require.Equal(t, currentViews, candidateViews)
		var receiptCount int
		require.NoError(t, stage.QueryRowContext(ctx, `SELECT count(*) FROM product_view_count_flush_receipts WHERE product_id=$1 AND delta IN (5,7)`, productID).Scan(&receiptCount))
		require.Equal(t, 2, receiptCount)
		fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
		require.NoError(t, err)
		require.False(t, fenced)
		_, err = manager.ApplyVerifiedRecovery(candidate.ID, "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+completed.Manifest.PackageID)
		require.Error(t, err, "stopped drain must not enable production product overwrite")
		verifyContinuouslyStoppedProductCatalogHandoffPreparationRestartAndCancellation(t, managerReference, packageBytes, options)
		verifyStoppedProductCatalogIsolatedSwitchPostSwitchInspectionAndInterruptedPairedRollback(t, managerReference, packageBytes, options)
	})
}
