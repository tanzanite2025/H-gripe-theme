package localbackup

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// Serve non-action status, task, backup-creation and resumable-upload routes after authentication and audit.
func (manager *Manager) serveMaintenanceStatusTasksBackupAndUploadRoute(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, path string, segments []string) bool {
	if path == "/api/status" && request.Method == http.MethodGet {
		fence, readErr := os.ReadFile(filepath.Join(manager.Configuration.MaintenanceDirectory, "application-maintenance.json"))
		if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
			writeMaintenanceError(response, readErr)
			return true
		}
		writeMaintenanceJSON(response, http.StatusOK, map[string]any{"site_id": manager.Configuration.SiteID, "environment": manager.Configuration.Environment, "application_version": manager.Configuration.ApplicationVersion, "database_name": manager.Configuration.DatabaseName, "storage_roots": manager.Configuration.StorageRoots, "package_version": PackageVersion, "chunk_bytes": ChunkBytes, "max_package_bytes": manager.Configuration.MaxPackageBytes, "work_copy_lifetime_hours": manager.Configuration.WorkCopyLifetimeHours, "application_fenced": len(fence) > 0, "permissions": session.Permissions,
			"product_view_stopped_drain_configured":  manager.Configuration.ProductViewRecoveryRedis != nil && len(manager.Configuration.ProductViewRecoveryWriterInventory) > 0,
			"business_module_backup_options":         availableBusinessModuleBackupOptions(),
			"product_view_recovery_writer_inventory": manager.Configuration.ProductViewRecoveryWriterInventory})
		return true
	}
	if path == "/api/tasks" && request.Method == http.MethodGet {
		tasks, err := manager.ListTasks()
		if err != nil {
			writeMaintenanceError(response, err)
			return true
		}
		for index, task := range tasks {
			tasks[index] = summarizeTaskForMaintenanceHTTP(task)
		}
		writeMaintenanceJSON(response, http.StatusOK, map[string]any{"tasks": tasks})
		return true
	}
	if path == "/api/backups" && request.Method == http.MethodPost {
		var body struct {
			Confirmation string `json:"confirmation"`
		}
		if !readMaintenanceJSON(response, request, &body) {
			return true
		}
		task, err := manager.CreateBackup(session.Subject, body.Confirmation)
		writeMaintenanceTask(response, task, err)
		return true
	}
	if path == "/api/product-catalog-backups" && request.Method == http.MethodPost {
		var body struct {
			Confirmation string `json:"confirmation"`
		}
		if !readMaintenanceJSON(response, request, &body) {
			return true
		}
		task, err := manager.CreateProductCatalogBackup(session.Subject, body.Confirmation)
		writeMaintenanceTask(response, task, err)
		return true
	}
	if (path == "/api/module-backups" || path == "/api/module-backup-preview") && request.Method == http.MethodPost {
		var body struct {
			Modules      []string `json:"modules"`
			Confirmation string   `json:"confirmation"`
		}
		if !readMaintenanceJSON(response, request, &body) {
			return true
		}
		if path == "/api/module-backup-preview" {
			preview, err := manager.PreviewSelectedBusinessModuleBackup(request.Context(), body.Modules)
			if err != nil {
				writeMaintenanceError(response, err)
				return true
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"preview": preview})
			return true
		}
		task, err := manager.CreateSelectedBusinessModuleBackup(session.Subject, body.Confirmation, body.Modules)
		writeMaintenanceTask(response, task, err)
		return true
	}
	if path == "/api/uploads" && request.Method == http.MethodPost {
		var body struct {
			Size int64 `json:"size"`
		}
		if !readMaintenanceJSON(response, request, &body) {
			return true
		}
		task, err := manager.BeginUpload(session.Subject, body.Size)
		writeMaintenanceTask(response, task, err)
		return true
	}
	if len(segments) == 3 && segments[0] == "api" && segments[1] == "tasks" && request.Method == http.MethodGet {
		task, err := manager.GetTask(segments[2])
		writeMaintenanceTask(response, task, err)
		return true
	}
	if len(segments) == 4 && segments[0] == "api" && segments[1] == "uploads" {
		id, action := segments[2], segments[3]
		if action == "chunks" && request.Method == http.MethodPut {
			offset, err := strconv.ParseInt(request.Header.Get("Upload-Offset"), 10, 64)
			if err != nil {
				writeMaintenanceError(response, errors.New("invalid chunk offset"))
				return true
			}
			data, err := readUploadChunk(request.Body)
			if err != nil {
				writeMaintenanceError(response, err)
				return true
			}
			task, err := manager.AppendUploadChunk(id, offset, data, request.Header.Get("X-Chunk-SHA256"))
			writeMaintenanceTask(response, task, err)
			return true
		}
		if action == "commit" && request.Method == http.MethodPost {
			task, err := manager.CommitUpload(id)
			writeMaintenanceTask(response, task, err)
			return true
		}
	}
	return false
}
