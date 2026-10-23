package localbackup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
	"commerce-platform/internal/service"
)

func TestProductCatalogStoppedHandoffPreparationAndCancellationRequireRestorePermissionAndStrictRequests(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager := &Manager{Configuration: cfg}
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	for _, route := range []string{"product-catalog-stopped-handoff-preparation", "product-catalog-stopped-handoff-cancellation", "product-catalog-stopped-isolated-switch", "product-catalog-stopped-post-switch-inspection", "product-catalog-stopped-switch-rollback", "product-catalog-stopped-detail-cache-invalidation", "product-catalog-stopped-projection-inspection", "product-catalog-stopped-projection-evidence", "product-catalog-stopped-read-only-product-http-query", "product-catalog-stopped-read-only-media-http-query", "product-catalog-stopped-read-only-storefront-http-query", "product-catalog-stopped-read-only-public-image-http-query"} {
		for _, scenario := range []struct {
			permission, body string
			status           int
		}{
			{PermissionView, `{}`, http.StatusForbidden},
			{PermissionExport, `{}`, http.StatusForbidden},
			{PermissionUpload, `{}`, http.StatusForbidden},
			{PermissionRestore, `{"redis_password":"forged"}`, http.StatusConflict},
			{PermissionRestore, `{"phase":"ready"}`, http.StatusConflict},
			{PermissionRestore, `{} {}`, http.StatusConflict},
		} {
			t.Run(route+"/"+scenario.permission+"/"+scenario.body, func(t *testing.T) {
				token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic", []string{scenario.permission})
				require.NoError(t, err)
				request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/tasks/"+strings.Repeat("a", 32)+"/"+route, strings.NewReader(scenario.body))
				request.Header.Set("Authorization", "Bearer "+token)
				response := httptest.NewRecorder()
				manager.HTTPHandler().ServeHTTP(response, request)
				require.Equal(t, scenario.status, response.Code, response.Body.String())
			})
		}
	}
}

func TestProductCatalogStoppedHandoffCancellationFinishesDurableReleaseIntentAfterFenceRemovalAndRestart(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	task, err := manager.createTask("upload", "synthetic-cancellation-operator")
	require.NoError(t, err)
	task.Manifest = &Manifest{Scope: productCatalogBackupScope, PackageID: newTaskID()}
	task.ExpiresAt = time.Now().Add(-time.Hour)
	task.State = "running"
	task.ProductCatalogPreviewBinding = &productCatalogRecoveryPreviewBinding{Version: 1}
	configurationDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(cfg)
	require.NoError(t, err)
	task.ProductViewStoppedDrain = &ProductViewStoppedDrainAndCandidateRebuildProgress{Version: 1, TaskID: task.ID, PackageID: task.Manifest.PackageID, ConfigurationSHA256: configurationDigest, Operator: task.Actor, TargetDatabaseOID: 42, Phase: "candidate_rebuilt", RebuiltPreviewSHA256: "synthetic-preview", After: &service.ProductViewCountRecoveryBoundaryInspection{RedisBuffersEmpty: true}}
	require.NoError(t, manager.saveProductViewStoppedDrainProgress(task))
	packageDigest, err := productCatalogStoppedHandoffImmutablePackageSHA256(task)
	require.NoError(t, err)
	previewDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(task.ProductCatalogPreviewBinding)
	require.NoError(t, err)
	task.ProductCatalogStoppedHandoffPreparation = &ProductCatalogStoppedHandoffPreparationJournal{Version: 1, WindowID: newTaskID(), TaskID: task.ID, PackageSHA256: packageDigest, ConfigurationSHA256: configurationDigest, Operator: task.Actor, Phase: "canceling", StartedAt: time.Now().Add(-time.Hour), PreparedAt: time.Now().Add(-time.Minute), CancellationVerifiedAt: time.Now(), FenceSHA256: "synthetic-removed-fence", PreviewBindingSHA256: previewDigest, DrainProgressSHA256: task.ProductViewStoppedDrain.SHA256}
	require.NoError(t, manager.saveProductCatalogStoppedHandoffPreparationJournal(task))
	require.NoError(t, manager.Close())
	manager, err = NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	_, err = manager.StartVerifiedCancellationOfProductCatalogStoppedHandoffPreparation(task.ID, task.Actor, ProductCatalogStoppedHandoffCancellationRequest{Confirmation: "CANCEL STOPPED PRODUCT HANDOFF " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID})
	require.NoError(t, err)
	cancelled := waitForIndependentRecoveryTask(t, manager, task.ID)
	require.Equal(t, "product_module_isolated_verified", cancelled.State, cancelled.Error)
	require.Equal(t, "cancelled", cancelled.ProductCatalogStoppedHandoffPreparation.Phase)
	require.True(t, time.Now().After(cancelled.ExpiresAt), "finishing cancellation never renews expired work")
	require.Nil(t, manager.productCatalogStoppedHandoffLease)
	fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
	require.NoError(t, err)
	require.False(t, fenced, "finishing the release record never creates a new fence")
}

func TestProductCatalogStoppedHandoffPreparationRetainsWorkCopiesAndRefusesAllOrdinaryReleasePathsAfterRestart(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	task, err := manager.createTask("upload", "synthetic-preparation-operator")
	require.NoError(t, err)
	task.Manifest = &Manifest{Scope: productCatalogBackupScope, PackageID: newTaskID()}
	task.State, task.ExpiresAt = "running", time.Now().Add(-time.Hour)
	packageDigest, err := productCatalogStoppedHandoffImmutablePackageSHA256(task)
	require.NoError(t, err)
	configurationDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(cfg)
	require.NoError(t, err)
	task.ProductCatalogStoppedHandoffPreparation = &ProductCatalogStoppedHandoffPreparationJournal{Version: 1, WindowID: newTaskID(), TaskID: task.ID, PackageSHA256: packageDigest, ConfigurationSHA256: configurationDigest, Operator: task.Actor, Phase: "preparing", StartedAt: time.Now().UTC(), FenceSHA256: "synthetic-fence-sha256"}
	require.NoError(t, manager.saveProductCatalogStoppedHandoffPreparationJournal(task))
	require.NoError(t, writeDurableJSON(filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename), map[string]string{"task_id": task.ID}))
	require.NoError(t, manager.Close())
	manager, err = NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	retained, err := manager.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, "interrupted", retained.State)
	require.Nil(t, manager.productCatalogStoppedHandoffLease, "restart cannot inherit a continuous lease")
	require.Equal(t, task.ProductCatalogStoppedHandoffPreparation.SHA256, retained.ProductCatalogStoppedHandoffPreparation.SHA256)
	require.NoError(t, manager.ExpireWorkCopies(context.Background()))
	require.False(t, retained.WorkCopyExpired)
	require.ErrorContains(t, manager.ReleaseUnmodifiedMaintenance(task.ID), "dedicated verification")
	require.ErrorContains(t, manager.releaseMaintenance(retained), "ordinary release is forbidden")
	_, err = manager.VerifyUploadedPackage(task.ID)
	require.ErrorContains(t, err, "stopped product handoff")
	_, err = manager.BeginUpload("synthetic", 1024)
	require.ErrorContains(t, err, "work copies must be retained")
	_, err = manager.StartStoppedProductViewDrainAndFreshCandidateRebuild(task.ID, "synthetic", ProductViewStoppedDrainAndCandidateRebuildRequest{})
	require.ErrorContains(t, err, "resume or cancel")
	require.Error(t, manager.rehearseProductCatalogPairedSwitchAndVerifiedRollback(context.Background(), task.ID, nil))
	require.NoError(t, os.Remove(filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename)))
	_, err = manager.acquireMaintenance(context.Background(), retained)
	require.ErrorContains(t, err, "fence is missing")
	require.ErrorContains(t, manager.ExpireWorkCopies(context.Background()), "work copies must be retained")
	other, err := manager.createTask("backup", "synthetic-other")
	require.NoError(t, err)
	_, err = manager.acquireMaintenance(context.Background(), other)
	require.ErrorContains(t, err, "unresolved stopped product handoff")
	retained.ProductCatalogStoppedHandoffPreparation.Phase = "cancelled"
	require.ErrorContains(t, manager.validateProductCatalogStoppedHandoffPreparationJournal(retained), "journal changed")
	require.True(t, taskHasUnresolvedProductCatalogStoppedHandoffPreparation(retained), "a damaged terminal marker cannot disable maintenance protection")
	require.Error(t, manager.releaseMaintenance(retained))
}
