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

func TestSelectedBusinessModuleIsolatedRestoreRequiresRestorePermissionConfirmationAndUnexpiredValidatedUpload(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	for _, permission := range []string{PermissionView, PermissionExport, PermissionUpload} {
		token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic", []string{permission})
		require.NoError(t, err)
		request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, ConsolePrefix+"/api/tasks/"+newTaskID()+"/selected-module-isolated-restore-preview", strings.NewReader(`{"confirmation":""}`))
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		manager.HTTPHandler().ServeHTTP(response, request)
		require.Equal(t, http.StatusForbidden, response.Code)
	}
	task, err := manager.createTask("upload", "synthetic")
	require.NoError(t, err)
	task.Manifest = &Manifest{Scope: selectedBusinessModulesBackupScope}
	task.State = "module_package_integrity_verified"
	require.NoError(t, manager.saveTask(task))
	_, err = manager.StartSelectedBusinessModuleIsolatedRestorePreview(task.ID, "wrong confirmation")
	require.ErrorContains(t, err, "维护窗口")
	phrase := "VERIFY MODULE RESTORE " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID
	task.ExpiresAt = time.Now().Add(-time.Second)
	require.NoError(t, manager.saveTask(task))
	_, err = manager.StartSelectedBusinessModuleIsolatedRestorePreview(task.ID, phrase)
	require.ErrorContains(t, err, "未到期")
	task.ExpiresAt = time.Now().Add(time.Hour)
	task.State = "verified"
	require.NoError(t, manager.saveTask(task))
	_, err = manager.StartSelectedBusinessModuleIsolatedRestorePreview(task.ID, phrase)
	require.ErrorContains(t, err, "重新校验")
}
