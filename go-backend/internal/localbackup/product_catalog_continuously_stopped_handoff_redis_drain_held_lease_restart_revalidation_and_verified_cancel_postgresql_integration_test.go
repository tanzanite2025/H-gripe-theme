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

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/gofrs/flock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// Reuses the preceding fixture's explicitly enrolled disposable Redis instance;
// no production connection or ownership enrollment is inferred here.
func verifyContinuouslyStoppedProductCatalogHandoffPreparationRestartAndCancellation(t *testing.T, managerReference **Manager, packageBytes []byte, redisOptions *redis.Options) {
	t.Helper()
	t.Run("continuous_handoff_preparation_held_lease_restart_and_verified_cancellation", func(t *testing.T) {
		ctx := context.Background()
		manager := *managerReference
		cfg := manager.Configuration
		candidate := uploadProductCatalogPackageForIntegrationTest(t, manager, packageBytes)
		require.Equal(t, "product_module_isolated_verified", candidate.State, candidate.Error)
		live, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
		require.NoError(t, err)
		defer live.Close()
		var productID uint
		var originalViews int64
		require.NoError(t, live.QueryRowContext(ctx, `SELECT id,view_count FROM products WHERE slug='product-backup-original' AND locale='en'`).Scan(&productID, &originalViews))
		client := redis.NewClient(redisOptions)
		require.NoError(t, client.HSet(ctx, "product:view_counts", strconv.FormatUint(uint64(productID), 10), "11").Err())
		require.NoError(t, client.Close())
		key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
		require.NoError(t, err)
		token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-continuous-window-operator", []string{PermissionRestore})
		require.NoError(t, err)
		invokeStoppedProductHandoffRouteAndWait := func(route string, body any) *Task {
			encoded, err := json.Marshal(body)
			require.NoError(t, err)
			request := httptest.NewRequest(http.MethodPost, "/api/tasks/"+candidate.ID+"/"+route, strings.NewReader(string(encoded)))
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			manager.HTTPHandler().ServeHTTP(response, request)
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			return waitForIndependentRecoveryTask(t, manager, candidate.ID)
		}
		prepare := func() *Task {
			return invokeStoppedProductHandoffRouteAndWait("product-catalog-stopped-handoff-preparation", ProductCatalogStoppedHandoffPreparationRequest{Confirmation: "PREPARE STOPPED PRODUCT HANDOFF " + cfg.SiteID + "/" + cfg.Environment + "/" + candidate.ID, AllListedWritersStoppedAndUseThisSingleHostGate: true})
		}
		cancel := func() *Task {
			return invokeStoppedProductHandoffRouteAndWait("product-catalog-stopped-handoff-cancellation", ProductCatalogStoppedHandoffCancellationRequest{Confirmation: "CANCEL STOPPED PRODUCT HANDOFF " + cfg.SiteID + "/" + cfg.Environment + "/" + candidate.ID})
		}
		prepared := prepare()
		require.Equal(t, "product_module_stopped_handoff_prepared", prepared.State, prepared.Error)
		require.Equal(t, "ready", prepared.ProductCatalogStoppedHandoffPreparation.Phase)
		require.Nil(t, prepared.Journal, "preparation never creates or consumes a switch journal")
		require.False(t, prepared.ProductCatalogRecovery.OnlineOverwriteAvailable)
		require.Equal(t, candidate.ProductCatalogPreviewBinding.Target.OID, prepared.ProductCatalogPreviewBinding.Target.OID)
		require.NotEqual(t, candidate.ProductCatalogPreviewBinding.Candidate.OID, prepared.ProductCatalogPreviewBinding.Candidate.OID)
		probe := flock.New(filepath.Join(cfg.MaintenanceDirectory, "application-runtime.lock"))
		defer probe.Close()
		locked, err := probe.TryRLock()
		require.NoError(t, err)
		require.False(t, locked, "exclusive runtime lease must remain held after the HTTP task finishes")
		require.Error(t, manager.ReleaseUnmodifiedMaintenance(candidate.ID))
		_, err = manager.VerifyUploadedPackage(candidate.ID)
		require.Error(t, err)
		_, err = manager.ApplyVerifiedRecovery(candidate.ID, "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+candidate.Manifest.PackageID)
		require.Error(t, err)
		require.NoError(t, manager.ExpireWorkCopies(ctx))
		// A changed fence cannot be silently adopted, even with the same task ID.
		fencePath := filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename)
		fence, err := os.ReadFile(fencePath)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(fencePath, append(append([]byte{}, fence...), '\n'), 0600))
		refused := cancel()
		require.Equal(t, "failed", refused.State)
		require.Contains(t, refused.Error, "fence changed")
		require.NoError(t, os.WriteFile(fencePath, fence, 0600)) // restore only this fixture's known original
		// Unfenced external writes invalidate readiness. Cancellation must not
		// release; the explicit preparation retries from the latest target.
		_, err = live.ExecContext(ctx, `UPDATE products SET view_count=view_count+2 WHERE id=$1`, productID)
		require.NoError(t, err)
		refused = cancel()
		require.Equal(t, "failed", refused.State)
		require.Contains(t, refused.Error, "target changed")
		require.NoError(t, manager.Close())
		manager, err = NewManager(cfg)
		require.NoError(t, err)
		*managerReference = manager
		restarted, err := manager.GetTask(candidate.ID)
		require.NoError(t, err)
		require.Equal(t, "interrupted", restarted.State)
		require.Nil(t, manager.productCatalogStoppedHandoffLease)
		require.Error(t, manager.ReleaseUnmodifiedMaintenance(candidate.ID))
		reprepared := prepare()
		require.Equal(t, "product_module_stopped_handoff_prepared", reprepared.State, reprepared.Error)
		require.Equal(t, prepared.ProductCatalogStoppedHandoffPreparation.WindowID, reprepared.ProductCatalogStoppedHandoffPreparation.WindowID)
		require.NotEqual(t, prepared.ProductCatalogPreviewBinding.Candidate.OID, reprepared.ProductCatalogPreviewBinding.Candidate.OID)
		var views int64
		require.NoError(t, live.QueryRowContext(ctx, `SELECT view_count FROM products WHERE id=$1`, productID).Scan(&views))
		require.Equal(t, originalViews+13, views, "retry must not replay the already committed delta")
		cancelled := cancel()
		require.Equal(t, "product_module_isolated_verified", cancelled.State, cancelled.Error)
		require.Equal(t, "cancelled", cancelled.ProductCatalogStoppedHandoffPreparation.Phase)
		require.Equal(t, "product_handoff_cancelled_original_service_released", cancelled.Phase)
		fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
		require.NoError(t, err)
		require.False(t, fenced)
		locked, err = probe.TryRLock()
		require.NoError(t, err)
		require.True(t, locked, "verified cancellation releases the held runtime lease")
		require.NoError(t, probe.Unlock())
		var targetOID int64
		control, err := cfg.connectDatabase(ctx, "postgres")
		require.NoError(t, err)
		require.NoError(t, readDatabaseOID(ctx, control, cfg.DatabaseName, &targetOID))
		require.NoError(t, control.Close())
		require.Equal(t, prepared.ProductCatalogPreviewBinding.Target.OID, targetOID, "preparation and cancellation never switch the target")
		_, err = manager.ApplyVerifiedRecovery(candidate.ID, "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+candidate.Manifest.PackageID)
		require.Error(t, err, "cancellation does not enable product overwrite")
	})
}
