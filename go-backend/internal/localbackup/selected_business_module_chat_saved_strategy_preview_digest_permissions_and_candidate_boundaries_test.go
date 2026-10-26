package localbackup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createSelectedBusinessModuleChatPreviewForSavedStrategyBoundaryTest(t *testing.T, manager *Manager) *Task {
	t.Helper()
	task, err := manager.createTask("upload", "synthetic")
	require.NoError(t, err)
	task.State = "module_restore_conflicts_found"
	task.Manifest = &Manifest{Scope: selectedBusinessModulesBackupScope, PackageID: newTaskID(), SelectedModules: []string{"customer_service"}}
	strategy := &SelectedCustomerServiceModuleOverwriteStrategyInspection{Policy: "use_package_conversation_rows", Complete: true, DomainChecksSatisfied: true}
	task.SelectedModuleRestorePreview = &SelectedBusinessModuleIsolatedRestorePreview{Modules: []string{"customer_service"}, TargetUnchanged: true, DisposableStageRemoved: true,
		Binding:                         &SelectedCustomerServiceModuleSnapshotBinding{Version: 1, TaskID: task.ID, PackageID: task.Manifest.PackageID, WorkCopyExpiresAt: task.ExpiresAt},
		CustomerServiceOverwritePreview: &SelectedCustomerServiceModuleBoundedOverwritePreview{PrerequisitesSatisfied: true, PackageRows: strategy}}
	task.SelectedModuleRestorePreview.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(task.SelectedModuleRestorePreview)
	require.NoError(t, err)
	return task
}

func TestSelectedBusinessModuleChatSavedStrategyAndCandidateRequireRestorePermission(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	for _, permission := range []string{PermissionView, PermissionExport, PermissionUpload} {
		for _, action := range []string{"selected-customer-service-restore-choice", "selected-customer-service-candidate-rehearsal"} {
			token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic", []string{permission})
			require.NoError(t, err)
			request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, ConsolePrefix+"/api/tasks/"+newTaskID()+"/"+action, strings.NewReader(`{}`))
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			manager.HTTPHandler().ServeHTTP(response, request)
			require.Equal(t, http.StatusForbidden, response.Code, action)
		}
	}
}

func TestSelectedBusinessModuleChatSavedStrategyRejectsChangedExpiredOrUnrelatedPreview(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	for _, scenario := range []struct {
		name   string
		mutate func(*Task)
	}{
		{"missing_preview", func(task *Task) { task.SelectedModuleRestorePreview = nil }},
		{"changed_counts", func(task *Task) { task.SelectedModuleRestorePreview.RestoredRows++ }},
		{"unremoved_stage", func(task *Task) { task.SelectedModuleRestorePreview.DisposableStageRemoved = false }},
		{"online_authority", func(task *Task) { task.SelectedModuleRestorePreview.OnlineOverwriteAvailable = true }},
		{"expired", func(task *Task) { task.ExpiresAt = time.Now().Add(-time.Hour) }},
		{"combined_package", func(task *Task) { task.Manifest.SelectedModules = append(task.Manifest.SelectedModules, "orders") }},
		{"whole_application", func(task *Task) { task.Manifest.Scope = "whole_application" }},
		{"running", func(task *Task) { task.State = "running" }},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			task := createSelectedBusinessModuleChatPreviewForSavedStrategyBoundaryTest(t, manager)
			digest := task.SelectedModuleRestorePreview.SHA256
			require.NoError(t, validateSelectedCustomerServiceModuleTaskAndPreview(task, digest))
			scenario.mutate(task)
			require.NoError(t, manager.saveTask(task))
			_, err := manager.SaveSelectedCustomerServiceModuleRestoreChoice(context.Background(), task.ID, "synthetic", digest, "use_package_conversation_rows")
			require.Error(t, err)
			latest, err := manager.GetTask(task.ID)
			require.NoError(t, err)
			require.Nil(t, latest.SelectedCustomerServiceRestoreChoice)
		})
	}
}

func TestSelectedBusinessModuleChatSavedStrategyCannotBypassDomainOrSharedWriteBlockers(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	for _, scenario := range []struct {
		name   string
		mutate func(*SelectedBusinessModuleIsolatedRestorePreview)
	}{
		{"domain_issue", func(report *SelectedBusinessModuleIsolatedRestorePreview) {
			report.CustomerServiceOverwritePreview.PackageRows.BlockingIssues = 1
		}},
		{"incomplete_domain", func(report *SelectedBusinessModuleIsolatedRestorePreview) {
			report.CustomerServiceOverwritePreview.PackageRows.Complete = false
		}},
		{"missing_shared_metadata", func(report *SelectedBusinessModuleIsolatedRestorePreview) {
			report.Tables = []SelectedBusinessModuleTableRestorePreview{{Table: "media_assets", MissingRows: 1}}
		}},
		{"missing_execution_fact", func(report *SelectedBusinessModuleIsolatedRestorePreview) {
			report.Tables = []SelectedBusinessModuleTableRestorePreview{{Table: "outbox_events", ProtectedNewRows: 1}}
		}},
		{"shared_content_conflict", func(report *SelectedBusinessModuleIsolatedRestorePreview) {
			report.Tables = []SelectedBusinessModuleTableRestorePreview{{Table: "users", ConflictingRows: 1}}
		}},
		{"file_conflict", func(report *SelectedBusinessModuleIsolatedRestorePreview) { report.FileConflicts = 1 }},
		{"dependency_conflict", func(report *SelectedBusinessModuleIsolatedRestorePreview) { report.DependencyConflicts = 1 }},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			task := createSelectedBusinessModuleChatPreviewForSavedStrategyBoundaryTest(t, manager)
			scenario.mutate(task.SelectedModuleRestorePreview)
			choice := &SelectedCustomerServiceModuleSavedRestoreChoice{Policy: "use_package_conversation_rows"}
			require.Error(t, validateSelectedCustomerServiceModuleChosenStrategyForFullRowRehearsal(choice, task.SelectedModuleRestorePreview))
		})
	}
	task := createSelectedBusinessModuleChatPreviewForSavedStrategyBoundaryTest(t, manager)
	choice := &SelectedCustomerServiceModuleSavedRestoreChoice{Version: 1, State: "recorded_not_executable", TaskID: task.ID, Operator: "synthetic", Policy: "use_package_conversation_rows", PreviewSHA256: task.SelectedModuleRestorePreview.SHA256, Binding: *task.SelectedModuleRestorePreview.Binding}
	choice.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(choice)
	require.NoError(t, err)
	task.SelectedCustomerServiceRestoreChoice = choice
	require.NoError(t, manager.saveTask(task))
	_, err = manager.StartSelectedCustomerServiceModuleCandidateRehearsal(task.ID, choice.SHA256, "wrong")
	require.ErrorContains(t, err, "维护窗口")
	choice.Policy = "preserve_current_conflicting_conversation_rows"
	require.NoError(t, manager.saveTask(task))
	_, err = manager.StartSelectedCustomerServiceModuleCandidateRehearsal(task.ID, choice.SHA256, "")
	require.ErrorContains(t, err, "摘要不一致")
}
