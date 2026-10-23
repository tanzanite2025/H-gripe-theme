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
)

func TestModuleDependencyReportProtectsIncomingReferencesAndTraversesCompositeDependencies(t *testing.T) {
	contract := moduleRecoveryDependencyContract{Name: "orders", Label: "订单", OwnedTables: []string{"orders"}, SharedTables: []string{"users"}}
	observation := &moduleRecoveryDatabaseObservation{
		Migration: 434,
		Tables:    map[string]int64{"orders": 1, "users": 1, "products": 1, "product_variants": 1, "showcases": 1, "schema_migrations": 1, "unregistered_empty_table": 0},
		Columns: map[string]map[string]bool{
			"orders": {"id": true, "product_id": true, "variant_id": true},
			"users":  {"id": true}, "products": {"id": true},
			"product_variants": {"id": true, "product_id": true}, "showcases": {"order_id": true},
		},
		Shapes: map[string]string{},
		References: []ModuleRecoveryDependencyReference{
			{SourceTable: "orders", SourceColumns: []string{"product_id", "variant_id"}, TargetTable: "product_variants", TargetColumns: []string{"product_id", "id"}, Kind: "foreign_key"},
			{SourceTable: "product_variants", SourceColumns: []string{"product_id"}, TargetTable: "products", TargetColumns: []string{"id"}, Kind: "foreign_key"},
			{SourceTable: "showcases", SourceColumns: []string{"order_id"}, TargetTable: "orders", TargetColumns: []string{"id"}, Kind: "foreign_key"},
		},
	}
	task := &Task{Manifest: &Manifest{PackageID: "synthetic-package"}}
	report := buildModuleRecoveryDependencyReport(contract, task, observation, observation)
	sortModuleRecoveryDependencyReport(report)
	require.False(t, report.RestoreAvailable)
	roles := map[string]string{}
	for _, table := range report.Tables {
		roles[table.Table] = table.Role
	}
	require.Equal(t, "module_owned", roles["orders"])
	require.Equal(t, "shared_dependency", roles["products"])
	require.Equal(t, "shared_dependency", roles["product_variants"])
	require.Equal(t, "preserved", roles["showcases"])
	require.Contains(t, collectModuleDependencyReportBlockerCodes(report), "preserved_table_references_module:showcases")
	require.Contains(t, collectModuleDependencyReportBlockerCodes(report), "unregistered_table:unregistered_empty_table")
	require.Contains(t, report.References, observation.References[0])
	require.Len(t, report.References, 3, "source and target constraints must be deduplicated")

	missingTarget := buildModuleRecoveryDependencyReport(contract, task, observation, nil)
	require.True(t, missingTarget.TargetDatabaseAbsent)
	require.Contains(t, collectModuleDependencyReportBlockerCodes(missingTarget), "target_database_absent:")
	require.False(t, missingTarget.RestoreAvailable)

	delete(observation.Tables, "users")
	missingSource := buildModuleRecoveryDependencyReport(contract, task, observation, nil)
	require.Contains(t, collectModuleDependencyReportBlockerCodes(missingSource), "source_dependency_table_missing:users")
}

func collectModuleDependencyReportBlockerCodes(report *ModuleRecoveryDependencyReport) []string {
	codes := make([]string, 0, len(report.Blockers))
	for _, blocker := range report.Blockers {
		codes = append(codes, blocker.Code+":"+blocker.Table)
	}
	return codes
}

func TestModuleDependencyInspectionRequiresRestorePermissionAndVerifiedUnexpiredUpload(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	task, err := manager.createTask("upload", "synthetic-operator")
	require.NoError(t, err)
	path := ConsolePrefix + "/api/tasks/" + task.ID + "/module-dependency-report"
	for _, permission := range []string{PermissionView, PermissionExport, PermissionUpload} {
		token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "operator", []string{permission})
		require.NoError(t, err)
		request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, path, strings.NewReader(`{"module":"customer_service"}`))
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		manager.HTTPHandler().ServeHTTP(response, request)
		require.Equal(t, http.StatusForbidden, response.Code, permission)
	}
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, path, strings.NewReader(`{"module":"customer_service"}`))
	request.Header.Set("Authorization", "Bearer "+key)
	response := httptest.NewRecorder()
	manager.HTTPHandler().ServeHTTP(response, request)
	require.Equal(t, http.StatusConflict, response.Code)
	require.Contains(t, response.Body.String(), "verified whole-storefront upload")
	request = httptest.NewRequestWithContext(context.Background(), http.MethodPost, ConsolePrefix+"/api/tasks/"+task.ID+"/apply", strings.NewReader(`{"module":"customer_service","confirmation":"synthetic"}`))
	request.Header.Set("Authorization", "Bearer "+key)
	response = httptest.NewRecorder()
	manager.HTTPHandler().ServeHTTP(response, request)
	require.Equal(t, http.StatusConflict, response.Code)
	require.Contains(t, response.Body.String(), "invalid request body", "module parameters cannot be passed to whole-storefront apply")
	_, err = manager.InspectVerifiedPackageModuleDependencies(context.Background(), task.ID, "../orders")
	require.ErrorContains(t, err, "only orders and customer_service")
	_, err = os.Stat(filepath.Join(cfg.Workspace, task.ID, "module-dependency-report-customer_service.json"))
	require.ErrorIs(t, err, os.ErrNotExist)
	fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
	require.NoError(t, err)
	require.False(t, fenced)
	task.State = "verified"
	task.StageDatabase = "local_backup_stage_" + task.ID[:16]
	task.StageFingerprint = "synthetic-fingerprint"
	task.Manifest = &Manifest{}
	task.ExpiresAt = time.Now().Add(-time.Second)
	require.NoError(t, manager.saveTask(task))
	_, err = manager.InspectVerifiedPackageModuleDependencies(context.Background(), task.ID, "customer_service")
	require.ErrorContains(t, err, "verified whole-storefront upload")
}
