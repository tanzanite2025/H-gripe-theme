package localbackup

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofrs/flock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

func verifyStoppedProductCatalogIsolatedSwitchPostSwitchInspectionAndInterruptedPairedRollback(t *testing.T, managerReference **Manager, packageBytes []byte, redisOptions *redis.Options) {
	t.Helper()
	for _, scenario := range []string{"authenticated_switch_inspection_and_verified_rollback", "restart_after_journal_persisted", "restart_after_media_0_renamed_before_acknowledgement", "restart_after_candidate_database_renamed_before_acknowledgement", "reverse_move_then_fence_release_acknowledgement_loss", "cache_intent_then_executor_restart"} {
		t.Run("stopped_product_isolated_switch_"+scenario, func(t *testing.T) {
			ctx := context.Background()
			manager := *managerReference
			cfg := manager.Configuration
			candidate := uploadProductCatalogPackageForIntegrationTest(t, manager, packageBytes)
			require.Equal(t, "product_module_isolated_verified", candidate.State, candidate.Error)
			key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
			require.NoError(t, err)
			token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-stopped-switch-operator", []string{PermissionRestore})
			require.NoError(t, err)
			invokeProductSwitchHTTP := func(route string, body any, expectedStatus int) *httptest.ResponseRecorder {
				encoded, err := json.Marshal(body)
				require.NoError(t, err)
				request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/tasks/"+candidate.ID+"/"+route, strings.NewReader(string(encoded)))
				request.Header.Set("Authorization", "Bearer "+token)
				response := httptest.NewRecorder()
				manager.HTTPHandler().ServeHTTP(response, request)
				require.Equal(t, expectedStatus, response.Code, response.Body.String())
				return response
			}
			invokeProductSwitchHTTP("product-catalog-stopped-handoff-preparation", ProductCatalogStoppedHandoffPreparationRequest{Confirmation: "PREPARE STOPPED PRODUCT HANDOFF " + cfg.SiteID + "/" + cfg.Environment + "/" + candidate.ID, AllListedWritersStoppedAndUseThisSingleHostGate: true}, http.StatusOK)
			prepared := waitForIndependentRecoveryTask(t, manager, candidate.ID)
			require.Equal(t, "product_module_stopped_handoff_prepared", prepared.State, prepared.Error)
			originalOID := prepared.ProductCatalogPreviewBinding.Target.OID
			originalFence, err := os.ReadFile(filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename))
			require.NoError(t, err)
			probe := flock.New(filepath.Join(cfg.MaintenanceDirectory, "application-runtime.lock"))
			defer probe.Close()
			assertProductWindowStillStopped := func() {
				locked, err := probe.TryRLock()
				require.NoError(t, err)
				require.False(t, locked)
				actual, err := os.ReadFile(filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename))
				require.NoError(t, err)
				require.Equal(t, originalFence, actual)
			}
			restartProductMaintenanceExecutor := func() {
				require.NoError(t, manager.Close())
				manager, err = NewManager(cfg)
				require.NoError(t, err)
				*managerReference = manager
			}
			if strings.HasPrefix(scenario, "restart_after_") {
				fault := strings.TrimPrefix(scenario, "restart_after_")
				require.NoError(t, manager.startTaskOperation(prepared, "synthetic_switch_fault_injection", func(ctx context.Context, task *Task) error {
					return manager.switchProductCatalogWithinContinuouslyHeldStoppedWindow(ctx, task, "synthetic-fault-operator", func(point string) error {
						if point == fault {
							return errors.New("synthetic stopped product switch acknowledgement loss: " + point)
						}
						return nil
					})
				}))
				failed := waitForIndependentRecoveryTask(t, manager, candidate.ID)
				require.Equal(t, "failed", failed.State)
				require.Contains(t, failed.Error, fault)
				assertProductWindowStillStopped()
				restartProductMaintenanceExecutor()
				interrupted, err := manager.GetTask(candidate.ID)
				require.NoError(t, err)
				require.Equal(t, "interrupted", interrupted.State)
				require.Equal(t, "product_stopped_switch_requires_explicit_verified_rollback", interrupted.Phase)
				require.Nil(t, manager.productCatalogStoppedHandoffLease)
				invokeProductSwitchHTTP("product-catalog-stopped-post-switch-inspection", struct{}{}, http.StatusConflict)
				invokeProductSwitchHTTP("product-catalog-stopped-isolated-switch", ProductCatalogStoppedSwitchConfirmationRequest{Confirmation: "SWITCH STOPPED PRODUCT CATALOG " + cfg.SiteID + "/" + cfg.Environment + "/" + candidate.ID}, http.StatusConflict)
			} else {
				invokeProductSwitchHTTP("product-catalog-stopped-isolated-switch", ProductCatalogStoppedSwitchConfirmationRequest{Confirmation: "SWITCH STOPPED PRODUCT CATALOG " + cfg.SiteID + "/" + cfg.Environment + "/" + candidate.ID}, http.StatusOK)
				switched := waitForIndependentRecoveryTask(t, manager, candidate.ID)
				require.Equal(t, "product_module_switched_still_isolated", switched.State, switched.Error)
				require.Nil(t, switched.Journal.ProductCatalog, "operator-requested switching cannot borrow a rehearsal journal")
				require.False(t, switched.ProductCatalogRecovery.OnlineOverwriteAvailable)
				assertProductWindowStillStopped()
				if scenario == "authenticated_switch_inspection_and_verified_rollback" {
					before, err := manager.GetTask(candidate.ID)
					require.NoError(t, err)
					firstResponse := invokeProductSwitchHTTP("product-catalog-stopped-post-switch-inspection", struct{}{}, http.StatusOK)
					var first, second struct {
						Report ProductCatalogStoppedPostSwitchInspection
					}
					require.NoError(t, json.Unmarshal(firstResponse.Body.Bytes(), &first))
					require.NoError(t, json.Unmarshal(invokeProductSwitchHTTP("product-catalog-stopped-post-switch-inspection", struct{}{}, http.StatusOK).Body.Bytes(), &second))
					require.Equal(t, first.Report.ContractSHA256, second.Report.ContractSHA256)
					require.Equal(t, switched.Journal.NewDatabaseOID, first.Report.DatabaseOID)
					require.NotEqual(t, originalOID, first.Report.DatabaseOID)
					require.Equal(t, switched.ProductCatalogStoppedHandoffPreparation.WindowID, first.Report.WindowID)
					require.Len(t, first.Report.ProjectionRequirements, 5)
					require.True(t, first.Report.ApplicationRemainsStopped)
					require.True(t, first.Report.CandidateRemainsIsolated)
					require.False(t, first.Report.ExternalActionsExecuted)
					require.False(t, first.Report.OnlineOverwriteAvailable)
					after, err := manager.GetTask(candidate.ID)
					require.NoError(t, err)
					require.Equal(t, before, after, "fresh inspection is read-only and never saves release approval")
					invokeProductSwitchHTTP("product-catalog-stopped-projection-inspection", struct{}{}, http.StatusConflict)
					verifyStoppedProductDetailCacheExactInvalidationPartialAcknowledgementRetryAndPreservedRedisData(t, manager, candidate.ID, redisOptions, invokeProductSwitchHTTP)
					verifyStoppedProductProjectionActualSQLBaselineOperatorRecordsAndStaleCacheRefusal(t, manager, candidate.ID, redisOptions, invokeProductSwitchHTTP)
					verifyStoppedProductReadOnlyActualPublicHTTPQueriesPreserveSQLRedisAndSealedDatabases(t, manager, candidate.ID, redisOptions, invokeProductSwitchHTTP)
					verifyStoppedProductMediaActualGuestAuthorizationHashesRangesAndUnchangedRecoveryBoundary(t, manager, candidate.ID, invokeProductSwitchHTTP)
					verifyStoppedProductMediaActualMemberAccountAndAgentPermissionsWithoutChangingRecoveryBoundary(t, manager, candidate.ID, invokeProductSwitchHTTP)
					verifyStoppedStorefrontActualNetworkHTTPUsesRecoveredSQLWithoutWritingOrReleasing(t, manager, candidate.ID)
					verifyStoppedProductPublicImageActualNetworkBytesAndFailureResealingPreserveRecoveryBoundary(t, manager, candidate.ID)
					verifyStoppedProductCurrentSiteIdentitySettingsAndImageBytesWithinSameRecoveryWindow(t, manager, candidate.ID)
					require.Error(t, manager.ReleaseUnmodifiedMaintenance(candidate.ID))
					_, err = manager.RollbackRecovery(candidate.ID, "ROLLBACK "+cfg.SiteID+"/"+candidate.ID)
					require.ErrorContains(t, err, "dedicated")
					_, err = manager.StartVerifiedCancellationOfProductCatalogStoppedHandoffPreparation(candidate.ID, "synthetic", ProductCatalogStoppedHandoffCancellationRequest{Confirmation: "CANCEL STOPPED PRODUCT HANDOFF " + cfg.SiteID + "/" + cfg.Environment + "/" + candidate.ID})
					require.Error(t, err)
					require.Error(t, manager.CompleteReconciledRecovery(candidate.ID, "RESUME "+cfg.SiteID+"/"+cfg.Environment+"/"+candidate.ID, ReconciliationEvidence{}))
					// A changed fence cannot be adopted, even with an unchanged task ID.
					fencePath := filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename)
					require.NoError(t, os.WriteFile(fencePath, append(append([]byte{}, originalFence...), '\n'), 0600))
					invokeProductSwitchHTTP("product-catalog-stopped-post-switch-inspection", struct{}{}, http.StatusConflict)
					require.NoError(t, os.WriteFile(fencePath, originalFence, 0600))
					// New deltas block both inspection and rollback release. The fixture
					// removes only the uncommitted synthetic data it deliberately added.
					client := redis.NewClient(redisOptions)
					require.NoError(t, client.HSet(ctx, "product:view_counts", "1", "1").Err())
					require.NoError(t, client.Close())
					invokeProductSwitchHTTP("product-catalog-stopped-post-switch-inspection", struct{}{}, http.StatusConflict)
					invokeProductSwitchHTTP("product-catalog-stopped-switch-rollback", ProductCatalogStoppedSwitchConfirmationRequest{Confirmation: "ROLLBACK STOPPED PRODUCT CATALOG " + cfg.SiteID + "/" + cfg.Environment + "/" + candidate.ID}, http.StatusOK)
					failed := waitForIndependentRecoveryTask(t, manager, candidate.ID)
					require.Equal(t, "failed", failed.State)
					require.Contains(t, failed.Error, "buffers changed")
					assertProductWindowStillStopped()
					client = redis.NewClient(redisOptions)
					require.NoError(t, client.Del(ctx, "product:view_counts").Err())
					require.NoError(t, client.Close())
				}
			}
			if scenario == "cache_intent_then_executor_restart" {
				switched, err := manager.GetTask(candidate.ID)
				require.NoError(t, err)
				interruptStoppedProductCacheIntentThenVerifyExecutorRestartCannotContinue(t, manager, switched)
				restartProductMaintenanceExecutor()
				invokeProductSwitchHTTP("product-catalog-stopped-detail-cache-invalidation", ProductCatalogStoppedSwitchConfirmationRequest{Confirmation: "INVALIDATE STOPPED PRODUCT CACHE " + cfg.SiteID + "/" + cfg.Environment + "/" + candidate.ID}, http.StatusConflict)
				invokeProductSwitchHTTP("product-catalog-stopped-post-switch-inspection", struct{}{}, http.StatusConflict)
			}
			if scenario == "reverse_move_then_fence_release_acknowledgement_loss" {
				switched, err := manager.GetTask(candidate.ID)
				require.NoError(t, err)
				require.NoError(t, manager.startTaskOperation(switched, "synthetic_reverse_acknowledgement_loss", func(ctx context.Context, task *Task) error {
					return manager.rollbackStoppedProductCatalogSwitchWithVerifiedOriginalRelease(ctx, task, func(point string) error {
						if strings.Contains(point, "reversed_before_acknowledgement") {
							return errors.New("synthetic reverse acknowledgement loss")
						}
						return nil
					})
				}))
				failed := waitForIndependentRecoveryTask(t, manager, candidate.ID)
				require.Equal(t, "failed", failed.State)
				require.Contains(t, failed.Error, "synthetic reverse")
				assertProductWindowStillStopped()
				restartProductMaintenanceExecutor()
				interrupted, err := manager.GetTask(candidate.ID)
				require.NoError(t, err)
				require.NoError(t, manager.startTaskOperation(interrupted, "synthetic_fence_release_acknowledgement_loss", func(ctx context.Context, task *Task) error {
					return manager.rollbackStoppedProductCatalogSwitchWithVerifiedOriginalRelease(ctx, task, func(point string) error {
						if point == "fence_removed_before_release_acknowledgement" {
							return errors.New("synthetic release acknowledgement loss")
						}
						return nil
					})
				}))
				failed = waitForIndependentRecoveryTask(t, manager, candidate.ID)
				require.Equal(t, "failed", failed.State)
				require.Contains(t, failed.Error, "synthetic release")
				restartProductMaintenanceExecutor()
			} else {
				invokeProductSwitchHTTP("product-catalog-stopped-switch-rollback", ProductCatalogStoppedSwitchConfirmationRequest{Confirmation: "ROLLBACK STOPPED PRODUCT CATALOG " + cfg.SiteID + "/" + cfg.Environment + "/" + candidate.ID}, http.StatusOK)
			}
			rolledBack := waitForIndependentRecoveryTask(t, manager, candidate.ID)
			require.Equal(t, "rolled_back", rolledBack.State, rolledBack.Error)
			require.Equal(t, "original_released", rolledBack.Journal.ProductCatalogStoppedSwitch.Phase)
			require.True(t, taskHasDurablyReleasedOriginalAfterProductCatalogStoppedSwitchRollback(rolledBack))
			require.False(t, taskHasUnresolvedProductCatalogStoppedHandoffPreparation(rolledBack))
			require.NoError(t, manager.validateProductCatalogStoppedSwitchJournalBinding(rolledBack))
			fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
			require.NoError(t, err)
			require.False(t, fenced)
			locked, err := probe.TryRLock()
			require.NoError(t, err)
			require.True(t, locked)
			require.NoError(t, probe.Unlock())
			control, err := cfg.connectDatabase(ctx, "postgres")
			require.NoError(t, err)
			var currentOID int64
			require.NoError(t, readDatabaseOID(ctx, control, cfg.DatabaseName, &currentOID))
			require.Equal(t, originalOID, currentOID)
			require.NoError(t, control.Close())
			// Cleanup reads only bound candidate/previous volumes; the original
			// service has resumed and is not fingerprinted or modified again.
			require.NoError(t, manager.cleanupVerifiedRolledBackStoppedProductSwitchMediaVolumes(ctx, rolledBack))
		})
	}
}
