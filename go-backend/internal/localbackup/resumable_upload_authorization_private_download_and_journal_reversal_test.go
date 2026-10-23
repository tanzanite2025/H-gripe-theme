package localbackup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/stretchr/testify/require"
)

func createIndependentRecoveryTestConfiguration(t *testing.T) Configuration {
	t.Helper()
	root := t.TempDir()
	cfg := Configuration{SiteID: "synthetic-store", Environment: "test", ApplicationVersion: "synthetic-build-v1", ExpectedMigrationVersion: 434, ApplicationDatabaseRole: "postgres", DatabaseName: "isolated_synthetic", Workspace: filepath.Join(root, "workspace"), MaintenanceDirectory: filepath.Join(root, "gate"), StorageRoots: map[string]string{}, RecoverySecrets: map[string]string{"JWT_SECRET": "synthetic-jwt-key", "PAYMENT_CONFIG_MASTER_KEY": "synthetic-payment-key", "OPS_CONNECTOR_MASTER_KEY": "synthetic-connector-key"}, PassphraseFile: filepath.Join(root, "passphrase"), AuthenticationKeyFile: filepath.Join(root, "authentication"), MaxPackageBytes: 8 << 20, MaxExpandedBytes: 64 << 20, MaxWorkspaceBytes: 256 << 20, MaxFiles: 100, WorkCopyLifetimeHours: 24}
	cfg.PostgreSQLCapacityDirectory = filepath.Join(root, "postgresql-capacity")
	require.NoError(t, os.Mkdir(cfg.PostgreSQLCapacityDirectory, 0700))
	for _, name := range mediaRoots {
		cfg.StorageRoots[name] = filepath.Join(root, name)
	}
	for _, directory := range append([]string{cfg.Workspace, cfg.MaintenanceDirectory}, cfg.StorageRoots["uploads"], cfg.StorageRoots["uploads_private"], cfg.StorageRoots["site_logo_uploads"]) {
		require.NoError(t, os.Mkdir(directory, 0700))
	}
	for _, filename := range []string{cfg.PassphraseFile, cfg.AuthenticationKeyFile} {
		require.NoError(t, os.WriteFile(filename, []byte(filepath.Base(filename)+strings.Repeat("synthetic-test-credential-", 3)), 0600))
	}
	require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
	return cfg
}

func TestUploadChunkRetriesPersistAcrossRestartAndRejectWrongDataOrOffsets(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	task, err := manager.BeginUpload("operator", 12)
	require.NoError(t, err)
	chunk := []byte("first-chunk!")
	digest := sha256.Sum256(chunk)
	checksum := hex.EncodeToString(digest[:])
	_, err = manager.AppendUploadChunk(task.ID, 0, chunk, "bad")
	require.Error(t, err)
	updated, err := manager.AppendUploadChunk(task.ID, 0, chunk, checksum)
	require.NoError(t, err)
	require.EqualValues(t, 12, updated.UploadedBytes)
	_, err = manager.AppendUploadChunk(task.ID, 0, chunk, checksum)
	require.NoError(t, err)
	_, err = manager.AppendUploadChunk(task.ID, 13, chunk, checksum)
	require.Error(t, err)
	require.NoError(t, manager.Close())
	manager, err = NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	committed, err := manager.CommitUpload(task.ID)
	require.NoError(t, err)
	require.Equal(t, "uploaded", committed.State)
	filename, _ := manager.taskPackagePath(task.ID)
	contents, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, chunk, contents)
}

func TestMaintenanceHTTPScopesDoNotDependOnTheBusinessDatabaseAndDownloadsSupportRanges(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	viewer, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "viewer", []string{PermissionView})
	require.NoError(t, err)
	for _, scenario := range []struct {
		method, path, token string
		status              int
	}{{"GET", "/api/status", "", 401}, {"GET", "/api/tasks", viewer, 200}, {"POST", "/api/backups", viewer, 403}, {"POST", "/api/uploads", viewer, 403}, {"POST", "/api/tasks/invalid/apply", viewer, 403}} {
		request := httptest.NewRequest(scenario.method, scenario.path, strings.NewReader(`{}`))
		if scenario.token != "" {
			request.Header.Set("Authorization", "Bearer "+scenario.token)
		}
		response := httptest.NewRecorder()
		manager.HTTPHandler().ServeHTTP(response, request)
		require.Equal(t, scenario.status, response.Code, scenario.path)
	}
	task, err := manager.createTask("backup", "operator")
	require.NoError(t, err)
	task.State = "ready_to_download"
	task.Size = 12
	require.NoError(t, manager.saveTask(task))
	filename, _ := manager.taskPackagePath(task.ID)
	require.NoError(t, os.WriteFile(filename, []byte("0123456789ab"), 0600))
	downloadToken, err := cfg.issueDownloadSession("operator", task.ID)
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodGet, "/downloads/"+task.ID, nil)
	request.AddCookie(&http.Cookie{Name: "local_backup_download", Value: downloadToken})
	request.Header.Set("Range", "bytes=3-6")
	response := httptest.NewRecorder()
	manager.HTTPHandler().ServeHTTP(response, request)
	require.Equal(t, 206, response.Code)
	require.Equal(t, "3456", response.Body.String())
	request = httptest.NewRequest(http.MethodPost, "/api/backups", strings.NewReader(`{}`))
	request.AddCookie(&http.Cookie{Name: "local_backup_download", Value: downloadToken})
	response = httptest.NewRecorder()
	manager.HTTPHandler().ServeHTTP(response, request)
	require.Equal(t, 401, response.Code)
	claims, err := cfg.AuthenticateMaintenanceSession(downloadToken)
	require.NoError(t, err)
	require.Equal(t, task.ID, claims.DownloadTask)
}

func TestFileMoveJournalDistinguishesNotStartedMovedAndAlreadyRevertedOperations(t *testing.T) {
	root := t.TempDir()
	source, destination := filepath.Join(root, "source"), filepath.Join(root, "destination")
	require.NoError(t, os.WriteFile(source, []byte("retained-original"), 0600))
	move := RecoveryFileMove{Source: source, Destination: destination, Started: true}
	require.NoError(t, performJournaledFileMove(&move))
	require.True(t, move.Completed)
	// Replaying a journal after rename but before its completion write is safe.
	move.Completed = false
	require.NoError(t, performJournaledFileMove(&move))
	require.True(t, move.Completed)
	reverse := RecoveryFileMove{Source: destination, Destination: source}
	require.NoError(t, performJournaledFileMove(&reverse))
	contents, err := os.ReadFile(source)
	require.NoError(t, err)
	require.Equal(t, "retained-original", string(contents))
	require.NoError(t, os.WriteFile(destination, []byte("occupied"), 0600))
	require.Error(t, performJournaledFileMove(&move))
}

func TestExpirationCannotRemoveAnActiveRecoveryFenceOrPairedRollbackJournal(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	task, err := manager.createTask("upload", "operator")
	require.NoError(t, err)
	task.State = "failed"
	task.ExpiresAt = time.Now().Add(-time.Hour)
	task.Journal = &RecoveryJournal{}
	require.NoError(t, manager.saveTask(task))
	filename, _ := manager.taskPackagePath(task.ID)
	require.NoError(t, os.WriteFile(filename, []byte("retained"), 0600))
	require.NoError(t, manager.ExpireWorkCopies(context.Background()))
	_, err = os.Stat(filename)
	require.NoError(t, err)
	require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
}
