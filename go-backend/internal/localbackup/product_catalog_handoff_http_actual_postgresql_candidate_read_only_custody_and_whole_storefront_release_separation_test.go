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

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/stretchr/testify/require"
)

func verifyProductCatalogHandoffHTTPInspectsActualCandidateWithoutChangingCustodyDatabaseMediaOrTask(t *testing.T, manager *Manager, task *Task) {
	t.Helper()
	ctx, cfg := context.Background(), manager.Configuration
	boundaryPath := filepath.Join(cfg.MaintenanceDirectory, "application-boundary.json")
	originalBoundary, err := os.ReadFile(boundaryPath)
	require.NoError(t, err)
	defer func() { require.NoError(t, os.WriteFile(boundaryPath, originalBoundary, 0600)) }()
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-product-handoff-inspector", []string{PermissionRestore})
	require.NoError(t, err)
	for _, mode := range []string{"unregistered", "sql_direct", "redis_delta_buffer", "idempotent_redis_delta_buffer"} {
		t.Run("product_handoff_read_only_http_"+mode, func(t *testing.T) {
			if mode != "unregistered" {
				projection := syntheticExternalProjectionBoundaryForRecoveryReleaseTests()
				projection.ProductCatalog = &maintenancegate.ProductCatalogExternalProjectionBoundary{Version: 1, SearchMode: "sql_only", DetailCacheMode: "redis", DetailCacheContractSHA256: strings.Repeat("a", 64), ViewCountMode: mode}
				if mode == "idempotent_redis_delta_buffer" {
					projection.ProductCatalog.Version = 2
					projection.ProductCatalog.ViewCountMode = "redis_delta_buffer"
					projection.ProductCatalog.ViewCountKeysSHA256 = strings.Repeat("b", 64)
					projection.ProductCatalog.ViewCountFlushProtocol = maintenancegate.ProductViewCountImmutableBatchSQLReceiptProtocol
					projection.ProductCatalog.ViewCountFlushContractSHA256 = strings.Repeat("c", 64)
				}
				if mode == "redis_delta_buffer" {
					projection.ProductCatalog.ViewCountKeysSHA256 = strings.Repeat("b", 64)
				}
				require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets, projection))
			}
			beforeTarget, err := manager.captureProductCatalogRecoveryPreviewBaseline(ctx)
			require.NoError(t, err)
			beforeCandidate, err := cfg.captureProductCatalogRecoveryDatabaseBinding(ctx, task.StageDatabase)
			require.NoError(t, err)
			beforeMedia, err := manager.productCatalogCandidateMediaConfiguration(task).inventoryModuleRecoveryCompleteMedia(ctx)
			require.NoError(t, err)
			beforeTask, err := manager.GetTask(task.ID)
			require.NoError(t, err)
			beforeTaskJSON, err := json.Marshal(beforeTask)
			require.NoError(t, err)
			request := httptest.NewRequest(http.MethodPost, "/api/tasks/"+task.ID+"/product-catalog-handoff-inspection", strings.NewReader("{}"))
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			manager.HTTPHandler().ServeHTTP(response, request)
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			var body struct {
				Report ProductCatalogCandidateHandoffInspection `json:"report"`
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
			report := body.Report
			require.Equal(t, task.ID, report.TaskID)
			require.Equal(t, beforeTarget.Target.OID, report.TargetDatabaseOID)
			require.Equal(t, beforeCandidate.OID, report.CandidateDatabaseOID)
			require.Equal(t, "local_backup_isolated_owner", report.CandidateDatabaseOwner)
			require.False(t, report.OnlineOverwriteAvailable)
			require.False(t, report.CustodyTransferAvailable)
			require.True(t, report.StoppedRevalidationRequired)
			if mode == "unregistered" {
				require.Contains(t, report.Blockers, "product_projection_configuration_unknown_reregister_current_business_build")
			} else if mode == "redis_delta_buffer" {
				require.Contains(t, report.Blockers, "buffered_product_view_delta_drain_and_inflight_idempotency_contract_not_implemented")
			} else if mode == "idempotent_redis_delta_buffer" {
				require.Contains(t, report.Blockers, "buffered_product_view_delta_stopped_handoff_and_target_verification_not_implemented")
				require.Equal(t, 2, report.ProjectionConfiguration.ProductCatalog.Version)
			} else {
				require.Len(t, report.Blockers, 3)
			}
			_, err = manager.ApplyVerifiedRecovery(task.ID, "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+task.Manifest.PackageID)
			require.Error(t, err)
			_, err = manager.InspectReconciliation(task.ID)
			require.Error(t, err)
			require.Error(t, manager.CompleteReconciledRecovery(task.ID, "RESUME "+cfg.SiteID+"/"+cfg.Environment+"/"+task.ID, ReconciliationEvidence{}))
			afterTarget, err := manager.captureProductCatalogRecoveryPreviewBaseline(ctx)
			require.NoError(t, err)
			require.Equal(t, beforeTarget, afterTarget)
			afterCandidate, err := cfg.captureProductCatalogRecoveryDatabaseBinding(ctx, task.StageDatabase)
			require.NoError(t, err)
			require.Equal(t, beforeCandidate, afterCandidate)
			afterMedia, err := manager.productCatalogCandidateMediaConfiguration(task).inventoryModuleRecoveryCompleteMedia(ctx)
			require.NoError(t, err)
			require.Equal(t, beforeMedia, afterMedia)
			afterTask, err := manager.GetTask(task.ID)
			require.NoError(t, err)
			afterTaskJSON, err := json.Marshal(afterTask)
			require.NoError(t, err)
			require.Equal(t, beforeTaskJSON, afterTaskJSON)
			fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
			require.NoError(t, err)
			require.False(t, fenced)
		})
	}
}
