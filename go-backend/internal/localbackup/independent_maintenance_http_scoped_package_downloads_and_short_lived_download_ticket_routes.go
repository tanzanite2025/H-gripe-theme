package localbackup

import (
	"mime"
	"net/http"
	"os"
	"strings"
)

// Register each supported task action with its required permission and original request/response behavior.
func (manager *Manager) registeredMaintenancePackageDownloadsHTTPTaskActions() map[string]maintenanceHTTPTaskActionDefinition {
	return map[string]maintenanceHTTPTaskActionDefinition{
		"download-ticket": {Permission: PermissionExport, Methods: []string{http.MethodPost}, Handle: func(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, segments []string) {
			id := segments[2]

			if _, err := manager.GetTask(id); err != nil {
				writeMaintenanceError(response, err)
				return
			}
			value, err := manager.Configuration.issueDownloadSession(session.Subject, id)
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			secure := request.TLS != nil || request.Header.Get("X-Forwarded-Proto") == "https"
			cookiePath := ConsolePrefix + "/downloads/" + id
			if !strings.HasPrefix(request.URL.Path, ConsolePrefix) {
				cookiePath = "/downloads/" + id
			}
			http.SetCookie(response, &http.Cookie{Name: "local_backup_download", Value: value, Path: cookiePath, HttpOnly: true, Secure: secure, SameSite: http.SameSiteStrictMode, MaxAge: 900}) // #nosec G124 -- The scoped HttpOnly SameSiteStrict ticket cookie is Secure on HTTPS; loopback-only HTTP remains supported for the SSH maintenance tunnel.
			writeMaintenanceJSON(response, http.StatusOK, map[string]string{"download_path": cookiePath})
		}},
	}
}

func (manager *Manager) serveScopedMaintenancePackageDownload(response http.ResponseWriter, request *http.Request, session *MaintenanceSession, path string) {
	id := strings.TrimPrefix(path, "/downloads/")
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		response.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if session.DownloadTask != "" && session.DownloadTask != id {
		response.WriteHeader(http.StatusForbidden)
		return
	}
	if err := manager.recordDownloadRequest(id); err != nil {
		writeMaintenanceError(response, err)
		return
	}
	filename, err := manager.taskPackagePath(id)
	if err != nil {
		writeMaintenanceError(response, err)
		return
	}
	file, err := os.Open(filename) // #nosec G304 G703 -- The authenticated task ID is validated by taskPackagePath and resolves to the fixed package file in its private task directory.
	if err != nil {
		writeMaintenanceError(response, err)
		return
	}
	defer func(closeResource func() error) { _ = closeResource() }(file.Close)
	info, err := file.Stat()
	if err != nil {
		writeMaintenanceError(response, err)
		return
	}
	response.Header().Set("Content-Type", "application/octet-stream")
	backupFilename := "storefront-whole-backup-" + id + ".age"
	if task, err := manager.GetTask(id); err == nil && task.Manifest != nil && task.Manifest.Scope == productCatalogBackupScope {
		backupFilename = "product-catalog-with-images-backup-" + id + ".age"
	}
	if task, err := manager.GetTask(id); err == nil && task.Manifest != nil && task.Manifest.Scope == selectedBusinessModulesBackupScope {
		backupFilename = "storefront-modules-" + strings.Join(task.Manifest.SelectedModules, "-") + "-" + id + ".age"
	}
	response.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": backupFilename}))
	http.ServeContent(response, request, info.Name(), info.ModTime(), file)
}
