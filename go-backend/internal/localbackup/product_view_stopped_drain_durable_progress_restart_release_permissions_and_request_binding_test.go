package localbackup

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

func TestProductViewStoppedDrainUnfinishedProgressSurvivesRestartAndCannotReleaseExpireOrReverify(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = manager.Close() })
	task, err := manager.createTask("upload", "synthetic-drain-operator")
	require.NoError(t, err)
	task.State, task.ExpiresAt = "running", time.Now().Add(-time.Hour)
	task.Manifest = &Manifest{Scope: productCatalogBackupScope, PackageID: newTaskID()}
	configurationDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(cfg)
	require.NoError(t, err)
	task.ProductViewStoppedDrain = &ProductViewStoppedDrainAndCandidateRebuildProgress{Version: 1, TaskID: task.ID, PackageID: task.Manifest.PackageID, ConfigurationSHA256: configurationDigest, Operator: task.Actor, TargetDatabaseOID: 42, Phase: "draining"}
	require.NoError(t, manager.saveProductViewStoppedDrainProgress(task))
	require.NoError(t, writeDurableJSON(filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename), map[string]string{"task_id": task.ID}))
	require.ErrorContains(t, manager.ReleaseUnmodifiedMaintenance(task.ID), "active")
	_, err = manager.VerifyUploadedPackage(task.ID)
	require.Error(t, err)
	require.NoError(t, manager.ExpireWorkCopies(context.Background()))
	require.NoError(t, manager.Close())
	manager, err = NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	retained, err := manager.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, "interrupted", retained.State)
	require.Equal(t, task.ProductViewStoppedDrain.SHA256, retained.ProductViewStoppedDrain.SHA256)
	require.False(t, retained.WorkCopyExpired)
	require.Nil(t, retained.ProductCatalogRecovery)
	require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
	require.NoError(t, os.Remove(filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename)))
	_, err = manager.acquireMaintenance(context.Background(), retained)
	require.ErrorContains(t, err, "fence is missing")
	require.ErrorContains(t, manager.ExpireWorkCopies(context.Background()), "unresolved product view drain")
	other, err := manager.createTask("backup", "synthetic-other")
	require.NoError(t, err)
	_, err = manager.acquireMaintenance(context.Background(), other)
	require.ErrorContains(t, err, "unresolved product view drain")
	retained.ProductViewStoppedDrain.Phase = "candidate_rebuilt"
	require.ErrorContains(t, manager.releaseMaintenance(retained), "binding changed")
}

func TestProductViewStoppedDrainHTTPRequiresRestorePermissionAndRejectsClientConnectionOrFakeAttestation(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager := &Manager{Configuration: cfg}
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	for _, scenario := range []struct {
		name, permission, body string
		status                 int
	}{
		{"view", PermissionView, `{}`, http.StatusForbidden},
		{"export", PermissionExport, `{}`, http.StatusForbidden},
		{"upload", PermissionUpload, `{}`, http.StatusForbidden},
		{"injected_connection", PermissionRestore, `{"database_url":"postgres://forged.invalid"}`, http.StatusConflict},
		{"injected_redis_credentials", PermissionRestore, `{"redis_password":"fake"}`, http.StatusConflict},
		{"trailing_json", PermissionRestore, `{} {}`, http.StatusConflict},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic", []string{scenario.permission})
			require.NoError(t, err)
			request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/tasks/"+strings.Repeat("a", 32)+"/product-view-stopped-drain-and-rebuild", strings.NewReader(scenario.body))
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			manager.HTTPHandler().ServeHTTP(response, request)
			require.Equal(t, scenario.status, response.Code, response.Body.String())
		})
	}
	task, err := manager.createTask("upload", "synthetic")
	require.NoError(t, err)
	_, err = manager.StartStoppedProductViewDrainAndFreshCandidateRebuild(task.ID, "synthetic", ProductViewStoppedDrainAndCandidateRebuildRequest{Confirmation: "DRAIN PRODUCT VIEWS " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID})
	require.ErrorContains(t, err, "attestation")
	encoded, err := json.Marshal(cfg)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "product_view_recovery_redis")
}
