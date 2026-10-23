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

func createSyntheticProductCatalogCandidateHandoffTaskAndRuntimeBoundary(t *testing.T) (*Manager, *Task, maintenancegate.ApplicationBoundary) {
	t.Helper()
	cfg := createIndependentRecoveryTestConfiguration(t)
	id := newTaskID()
	task := &Task{ID: id, Kind: "upload", State: "product_module_isolated_verified", ExpiresAt: time.Now().Add(time.Hour), CiphertextSHA256: strings.Repeat("a", 64), StageDatabase: "local_backup_stage_" + id[:16], StageFingerprint: "candidate-fingerprint", PreviewTargetFingerprint: "combined-target-fingerprint",
		Manifest:                     &Manifest{Scope: productCatalogBackupScope, PackageID: newTaskID(), MigrationVersion: 435},
		ProductCatalogRecovery:       &ProductCatalogIsolatedRecoveryResult{PreservedTables: 201, PreservedTargetOnlyRows: true, InventoryPreservation: &ProductCatalogInventoryPreservationResult{ExistingInventoryAndIdentityFactsVerified: true}},
		ProductCatalogPreviewBinding: &productCatalogRecoveryPreviewBinding{Version: 1, Target: productCatalogRecoveryDatabaseBinding{OID: 1, Owner: "original-owner", Fingerprint: "target-fingerprint"}, Candidate: productCatalogRecoveryDatabaseBinding{OID: 2, Owner: "local_backup_isolated_owner", Fingerprint: "candidate-fingerprint"}, TargetMedia: &moduleRecoveryCompleteMediaInventory{}, CandidateMedia: &moduleRecoveryCompleteMediaInventory{}, RequiredImageCleanupProtection: &productCatalogRequiredImageCleanupProtection{Version: 1, CleanupRowsSHA256: strings.Repeat("b", 64)}},
	}
	projections := syntheticExternalProjectionBoundaryForRecoveryReleaseTests()
	projections.ProductCatalog = &maintenancegate.ProductCatalogExternalProjectionBoundary{Version: 1, SearchMode: "sql_only", DetailCacheMode: "redis", DetailCacheContractSHA256: strings.Repeat("c", 64), ViewCountMode: "sql_direct"}
	return &Manager{Configuration: cfg}, task, maintenancegate.ApplicationBoundary{Protocol: 2, ExternalProjections: &projections}
}

func TestProductCatalogHandoffInspectionBindsIdentitiesPreservesCustodyAndSeparatesWholeStorefrontEvidence(t *testing.T) {
	manager, task, boundary := createSyntheticProductCatalogCandidateHandoffTaskAndRuntimeBoundary(t)
	before, err := json.Marshal(task)
	require.NoError(t, err)
	report, err := manager.buildProductCatalogCandidateHandoffInspection(task, boundary)
	require.NoError(t, err)
	repeat, err := manager.buildProductCatalogCandidateHandoffInspection(task, boundary)
	require.NoError(t, err)
	require.Equal(t, report.ContractSHA256, repeat.ContractSHA256)
	require.False(t, report.CustodyTransferAvailable)
	require.False(t, report.OnlineOverwriteAvailable)
	require.True(t, report.StoppedRevalidationRequired)
	require.Equal(t, "maintenance_executor", report.CandidateCustodian)
	require.Len(t, report.ProjectionRequirements, 5)
	require.Len(t, report.Blockers, 3)
	after, err := json.Marshal(task)
	require.NoError(t, err)
	require.Equal(t, before, after)
	encoded, err := json.Marshal(report)
	require.NoError(t, err)
	for _, private := range []string{manager.Configuration.DatabaseURL, manager.Configuration.Workspace, task.StageDatabase} {
		if private != "" {
			require.NotContains(t, string(encoded), private)
		}
	}
	wholeEvidence := &ExternalProjectionReconciliationEvidence{ContractSHA256: report.ContractSHA256}
	require.Error(t, validateTaskBoundExternalProjectionReconciliationEvidence(&ExternalProjectionReconciliationContract{Version: 1, Scope: report.Scope, ContractSHA256: report.ContractSHA256}, wholeEvidence))
}

func TestProductCatalogHandoffInspectionDigestChangesWithPackagePreviewRuntimeAndPreservationFacts(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		change func(*Manager, *Task, *maintenancegate.ApplicationBoundary)
	}{
		{"package", func(_ *Manager, t *Task, _ *maintenancegate.ApplicationBoundary) { t.Manifest.PackageID = newTaskID() }},
		{"ciphertext", func(_ *Manager, t *Task, _ *maintenancegate.ApplicationBoundary) {
			t.CiphertextSHA256 = strings.Repeat("b", 64)
		}},
		{"scope_rows", func(_ *Manager, t *Task, _ *maintenancegate.ApplicationBoundary) {
			t.Manifest.Tables = map[string]int64{"products": 2}
		}},
		{"expiry", func(_ *Manager, t *Task, _ *maintenancegate.ApplicationBoundary) {
			t.ExpiresAt = t.ExpiresAt.Add(time.Minute)
		}},
		{"target_oid", func(_ *Manager, t *Task, _ *maintenancegate.ApplicationBoundary) {
			t.ProductCatalogPreviewBinding.Target.OID = 3
		}},
		{"target_snapshot", func(_ *Manager, t *Task, _ *maintenancegate.ApplicationBoundary) {
			t.ProductCatalogPreviewBinding.Target.Fingerprint = "changed"
		}},
		{"candidate_schema", func(_ *Manager, t *Task, _ *maintenancegate.ApplicationBoundary) {
			t.ProductCatalogPreviewBinding.Candidate.SchemaSHA256 = "changed"
		}},
		{"cleanup_receipts", func(_ *Manager, t *Task, _ *maintenancegate.ApplicationBoundary) {
			t.ProductCatalogPreviewBinding.RequiredImageCleanupProtection.CleanupRowsSHA256 = "changed"
		}},
		{"inventory_summary", func(_ *Manager, t *Task, _ *maintenancegate.ApplicationBoundary) {
			t.ProductCatalogRecovery.InventoryPreservation.ExistingVariantStocksPreserved++
		}},
		{"business_build", func(m *Manager, _ *Task, _ *maintenancegate.ApplicationBoundary) {
			m.Configuration.ApplicationVersion += "-new"
		}},
		{"redis_destination", func(_ *Manager, _ *Task, b *maintenancegate.ApplicationBoundary) {
			b.ExternalProjections.RedisDestinationSHA256 = strings.Repeat("d", 64)
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			manager, task, boundary := createSyntheticProductCatalogCandidateHandoffTaskAndRuntimeBoundary(t)
			before, err := manager.buildProductCatalogCandidateHandoffInspection(task, boundary)
			require.NoError(t, err)
			scenario.change(manager, task, &boundary)
			after, err := manager.buildProductCatalogCandidateHandoffInspection(task, boundary)
			require.NoError(t, err)
			require.NotEqual(t, before.ContractSHA256, after.ContractSHA256)
		})
	}
}

func TestProductCatalogHandoffInspectionKeepsUnknownAndBufferedViewCountsBlocked(t *testing.T) {
	for _, mode := range []string{"unregistered", "sql_direct", "redis_delta_buffer", "idempotent_redis_delta_buffer"} {
		t.Run(mode, func(t *testing.T) {
			manager, task, boundary := createSyntheticProductCatalogCandidateHandoffTaskAndRuntimeBoundary(t)
			switch mode {
			case "unregistered":
				boundary.ExternalProjections.ProductCatalog = nil
			case "redis_delta_buffer":
				boundary.ExternalProjections.ProductCatalog.ViewCountMode = mode
				boundary.ExternalProjections.ProductCatalog.ViewCountKeysSHA256 = strings.Repeat("e", 64)
			case "idempotent_redis_delta_buffer":
				boundary.Protocol = 3
				boundary.ExternalProjections.ProductCatalog.Version = 2
				boundary.ExternalProjections.ProductCatalog.ViewCountMode = "redis_delta_buffer"
				boundary.ExternalProjections.ProductCatalog.ViewCountKeysSHA256 = strings.Repeat("e", 64)
				boundary.ExternalProjections.ProductCatalog.ViewCountFlushProtocol = maintenancegate.ProductViewCountImmutableBatchSQLReceiptProtocol
				boundary.ExternalProjections.ProductCatalog.ViewCountFlushContractSHA256 = strings.Repeat("f", 64)
			}
			report, err := manager.buildProductCatalogCandidateHandoffInspection(task, boundary)
			require.NoError(t, err)
			require.False(t, report.OnlineOverwriteAvailable)
			if mode == "unregistered" {
				require.Contains(t, report.Blockers, "product_projection_configuration_unknown_reregister_current_business_build")
			}
			if mode == "redis_delta_buffer" {
				require.Contains(t, report.Blockers, "buffered_product_view_delta_drain_and_inflight_idempotency_contract_not_implemented")
			}
			if mode == "idempotent_redis_delta_buffer" {
				require.Contains(t, report.Blockers, "buffered_product_view_delta_stopped_handoff_and_target_verification_not_implemented")
				require.NotContains(t, report.Blockers, "buffered_product_view_delta_drain_and_inflight_idempotency_contract_not_implemented")
			}
			require.Contains(t, report.ProjectionRequirements[1].Description, "product:*")
			require.Contains(t, report.ProjectionRequirements[3].Description, "不清客服")
		})
	}
}

func TestProductCatalogHandoffInspectionRejectsUnverifiedExpiredWrongOwnerAndJournalTasksBeforeDatabaseAccess(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		change func(*Task)
	}{
		{"unverified", func(t *Task) { t.State = "verified" }},
		{"whole_storefront", func(t *Task) { t.Manifest.Scope = "whole_application" }},
		{"expired", func(t *Task) { t.ExpiresAt = time.Now().Add(-time.Second) }},
		{"removed", func(t *Task) { t.WorkCopyExpired = true }},
		{"journal", func(t *Task) { t.Journal = &RecoveryJournal{} }},
		{"missing_inventory", func(t *Task) { t.ProductCatalogRecovery.InventoryPreservation = nil }},
		{"wrong_owner", func(t *Task) { t.ProductCatalogPreviewBinding.Candidate.Owner = "business_user" }},
		{"same_database", func(t *Task) {
			t.ProductCatalogPreviewBinding.Candidate.OID = t.ProductCatalogPreviewBinding.Target.OID
		}},
		{"wrong_stage", func(t *Task) { t.StageDatabase = "unrelated" }},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			manager, task, _ := createSyntheticProductCatalogCandidateHandoffTaskAndRuntimeBoundary(t)
			require.NoError(t, os.Mkdir(filepath.Join(manager.Configuration.Workspace, task.ID), 0700))
			scenario.change(task)
			require.NoError(t, manager.saveTask(task))
			_, err := manager.InspectProductCatalogCandidateHandoff(context.Background(), task.ID)
			require.Error(t, err)
			fenced, err := maintenancegate.IsApplicationFenced(manager.Configuration.MaintenanceDirectory)
			require.NoError(t, err)
			require.False(t, fenced)
		})
	}
}

func TestProductCatalogHandoffHTTPRequiresRestorePermissionAndRejectsClientSuppliedEvidence(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	task, err := manager.createTask("upload", "synthetic-operator")
	require.NoError(t, err)
	endpoint := ConsolePrefix + "/api/tasks/" + task.ID + "/product-catalog-handoff-inspection"
	for _, permission := range []string{PermissionView, PermissionExport, PermissionUpload, PermissionRestore} {
		t.Run(permission, func(t *testing.T) {
			token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-operator", []string{permission})
			require.NoError(t, err)
			request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, strings.NewReader(`{}`))
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			manager.HTTPHandler().ServeHTTP(response, request)
			if permission == PermissionRestore {
				require.Equal(t, http.StatusConflict, response.Code)
				require.Contains(t, response.Body.String(), "fresh isolated product candidate")
			} else {
				require.Equal(t, http.StatusForbidden, response.Code)
			}
		})
	}
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, strings.NewReader(`{"contract_sha256":"forged","online_overwrite_available":true}`))
	request.Header.Set("Authorization", "Bearer "+key)
	response := httptest.NewRecorder()
	manager.HTTPHandler().ServeHTTP(response, request)
	require.Equal(t, http.StatusConflict, response.Code)
	require.Contains(t, response.Body.String(), "invalid request body")
}
