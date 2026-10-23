package localbackup

import (
	"embed"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//go:embed maintenance_console/*
var maintenanceConsoleFiles embed.FS

const ConsolePrefix = "/local-backup-maintenance"

func (manager *Manager) HTTPHandler() http.Handler {
	staticFiles, _ := fs.Sub(maintenanceConsoleFiles, "maintenance_console")
	static := http.FileServer(http.FS(staticFiles))
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Cache-Control", "no-store")
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.Header().Set("Referrer-Policy", "no-referrer")
		response.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; frame-ancestors 'self'; object-src 'none'; base-uri 'none'; form-action 'self'")
		path := strings.TrimPrefix(request.URL.Path, ConsolePrefix)
		if path == "/favicon.ico" {
			response.WriteHeader(http.StatusNoContent)
			return
		}
		if path == "" {
			path = "/"
		}
		if path == "/" || path == "/index.html" || path == "/maintenance-console.js" || path == "/maintenance-console.css" {
			if request.Method != http.MethodGet && request.Method != http.MethodHead {
				response.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			clone := request.Clone(request.Context())
			clone.URL.Path = path
			static.ServeHTTP(response, clone)
			return
		}
		if !strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/downloads/") {
			http.NotFound(response, request)
			return
		}
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			// Mutations accept only explicit bearer authentication, never cookies.
			if request.Header.Get("Sec-Fetch-Site") == "cross-site" {
				response.WriteHeader(http.StatusForbidden)
				return
			}
		}
		authorization := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
		if strings.HasPrefix(path, "/downloads/") && authorization == "" {
			if cookie, err := request.Cookie("local_backup_download"); err == nil {
				authorization = cookie.Value
			}
		}
		session, err := manager.Configuration.AuthenticateMaintenanceSession(authorization)
		if err != nil {
			writeMaintenanceJSON(response, http.StatusUnauthorized, map[string]string{"error": "维护凭据无效或已过期，请重新登录独立恢复入口"})
			return
		}
		if session.DownloadTask != "" && !strings.HasPrefix(path, "/downloads/") {
			response.WriteHeader(http.StatusForbidden)
			return
		}
		permission := PermissionView
		if strings.HasSuffix(path, "/product-catalog-stopped-read-only-public-image-http-query") {
			permission = PermissionRestore
		}
		if strings.HasPrefix(path, "/downloads/") || strings.HasSuffix(path, "/download-ticket") || path == "/api/backups" || path == "/api/product-catalog-backups" {
			permission = PermissionExport
		}
		if strings.HasPrefix(path, "/api/uploads") || strings.HasSuffix(path, "/verify") {
			permission = PermissionUpload
		}
		if strings.HasSuffix(path, "/apply") || strings.HasSuffix(path, "/rollback") || strings.HasSuffix(path, "/complete") || strings.HasSuffix(path, "/release") {
			permission = PermissionRestore
		}
		if strings.HasSuffix(path, "/reconciliation-report") || strings.HasSuffix(path, "/reconciliation-connections") || strings.HasSuffix(path, "/module-dependency-report") || strings.HasSuffix(path, "/customer-service-conflict-review") || strings.HasSuffix(path, "/customer-service-effective-selection-inspection") || strings.HasSuffix(path, "/product-catalog-handoff-inspection") || strings.HasSuffix(path, "/product-view-stopped-drain-and-rebuild") {
			permission = PermissionRestore
		}
		if strings.HasSuffix(path, "/product-catalog-stopped-handoff-preparation") || strings.HasSuffix(path, "/product-catalog-stopped-handoff-cancellation") || strings.HasSuffix(path, "/product-catalog-stopped-isolated-switch") || strings.HasSuffix(path, "/product-catalog-stopped-post-switch-inspection") || strings.HasSuffix(path, "/product-catalog-stopped-switch-rollback") || strings.HasSuffix(path, "/product-catalog-stopped-detail-cache-invalidation") {
			permission = PermissionRestore
		}
		if strings.HasSuffix(path, "/product-catalog-stopped-projection-inspection") || strings.HasSuffix(path, "/product-catalog-stopped-projection-evidence") || strings.HasSuffix(path, "/product-catalog-stopped-read-only-product-http-query") || strings.HasSuffix(path, "/product-catalog-stopped-read-only-media-http-query") || strings.HasSuffix(path, "/product-catalog-stopped-read-only-storefront-http-query") {
			permission = PermissionRestore
		}
		if !session.HasPermission(permission) {
			response.WriteHeader(http.StatusForbidden)
			return
		}
		if request.Method != http.MethodGet && request.Method != http.MethodHead || strings.HasPrefix(path, "/downloads/") {
			if err := manager.recordMaintenanceOperatorAction(session.Subject, request.Method, path); err != nil {
				writeMaintenanceError(response, errors.New("independent audit could not be persisted; operation refused"))
				return
			}
		}
		if strings.HasPrefix(path, "/downloads/") {
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
			file, err := os.Open(filename) // #nosec G304,G703 -- taskPackagePath validates the task ID and returns only its fixed private package path.
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
			response.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": backupFilename}))
			http.ServeContent(response, request, info.Name(), info.ModTime(), file)
			return
		}
		if path == "/api/status" && request.Method == http.MethodGet {
			fence, readErr := os.ReadFile(filepath.Join(manager.Configuration.MaintenanceDirectory, "application-maintenance.json"))
			if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
				writeMaintenanceError(response, readErr)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"site_id": manager.Configuration.SiteID, "environment": manager.Configuration.Environment, "application_version": manager.Configuration.ApplicationVersion, "database_name": manager.Configuration.DatabaseName, "storage_roots": manager.Configuration.StorageRoots, "package_version": PackageVersion, "chunk_bytes": ChunkBytes, "max_package_bytes": manager.Configuration.MaxPackageBytes, "work_copy_lifetime_hours": manager.Configuration.WorkCopyLifetimeHours, "application_fenced": len(fence) > 0, "permissions": session.Permissions,
				"product_view_stopped_drain_configured":  manager.Configuration.ProductViewRecoveryRedis != nil && len(manager.Configuration.ProductViewRecoveryWriterInventory) > 0,
				"product_view_recovery_writer_inventory": manager.Configuration.ProductViewRecoveryWriterInventory})
			return
		}
		if path == "/api/tasks" && request.Method == http.MethodGet {
			tasks, err := manager.ListTasks()
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			for index, task := range tasks {
				tasks[index] = summarizeTaskForMaintenanceHTTP(task)
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"tasks": tasks})
			return
		}
		if path == "/api/backups" && request.Method == http.MethodPost {
			var body struct {
				Confirmation string `json:"confirmation"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.CreateBackup(session.Subject, body.Confirmation)
			writeMaintenanceTask(response, task, err)
			return
		}
		if path == "/api/product-catalog-backups" && request.Method == http.MethodPost {
			var body struct {
				Confirmation string `json:"confirmation"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.CreateProductCatalogBackup(session.Subject, body.Confirmation)
			writeMaintenanceTask(response, task, err)
			return
		}
		if path == "/api/uploads" && request.Method == http.MethodPost {
			var body struct {
				Size int64 `json:"size"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.BeginUpload(session.Subject, body.Size)
			writeMaintenanceTask(response, task, err)
			return
		}
		segments := strings.Split(strings.Trim(path, "/"), "/")
		if len(segments) == 4 && segments[0] == "api" && segments[1] == "tasks" && request.Method == http.MethodPost {
			switch segments[3] {
			case "product-catalog-stopped-read-only-public-image-http-query":
				var body StoppedProductReadOnlyPublicImageHTTPRequest
				if !readMaintenanceJSON(response, request, &body) {
					return
				}
				result, err := manager.ExecuteStoppedProductReadOnlyPublicImageHTTPQuery(request.Context(), segments[2], body)
				if err != nil {
					writeMaintenanceError(response, err)
					return
				}
				writeMaintenanceJSON(response, http.StatusOK, map[string]any{"result": result})
				return
			case "product-catalog-stopped-read-only-storefront-http-query":
				var body StoppedProductReadOnlyStorefrontHTTPRequest
				if !readMaintenanceJSON(response, request, &body) {
					return
				}
				result, err := manager.ExecuteStoppedProductReadOnlyStorefrontHTTPQuery(request.Context(), segments[2], body)
				if err != nil {
					writeMaintenanceError(response, err)
					return
				}
				writeMaintenanceJSON(response, http.StatusOK, map[string]any{"result": result})
				return
			case "product-catalog-stopped-read-only-media-http-query":
				var body StoppedProductReadOnlyMediaHTTPRequest
				if !readMaintenanceJSON(response, request, &body) {
					return
				}
				result, err := manager.ExecuteStoppedProductReadOnlyMediaHTTPQuery(request.Context(), segments[2], body)
				if err != nil {
					writeMaintenanceError(response, err)
					return
				}
				writeMaintenanceJSON(response, http.StatusOK, map[string]any{"result": result})
				return
			case "product-catalog-stopped-read-only-product-http-query":
				var body StoppedProductReadOnlyPublicHTTPRequest
				if !readMaintenanceJSON(response, request, &body) {
					return
				}
				result, err := manager.ExecuteStoppedProductReadOnlyPublicHTTPQuery(request.Context(), segments[2], body)
				if err != nil {
					writeMaintenanceError(response, err)
					return
				}
				writeMaintenanceJSON(response, http.StatusOK, map[string]any{"result": result})
				return
			case "product-catalog-stopped-projection-inspection":
				var body struct{}
				if !readMaintenanceJSON(response, request, &body) {
					return
				}
				report, err := manager.InspectStoppedProductStorefrontProjectionBaseline(request.Context(), segments[2])
				if err != nil {
					writeMaintenanceError(response, err)
					return
				}
				writeMaintenanceJSON(response, http.StatusOK, map[string]any{"report": report})
				return
			case "product-catalog-stopped-projection-evidence":
				var body ProductCatalogStoppedProjectionEvidenceRequest
				if !readMaintenanceJSON(response, request, &body) {
					return
				}
				record, err := manager.SaveStoppedProductProjectionIndependentOperatorEvidence(request.Context(), segments[2], session.Subject, body)
				if err != nil {
					writeMaintenanceError(response, err)
					return
				}
				writeMaintenanceJSON(response, http.StatusOK, map[string]any{"record": record})
				return
			case "product-catalog-stopped-isolated-switch", "product-catalog-stopped-switch-rollback":
				var body ProductCatalogStoppedSwitchConfirmationRequest
				if !readMaintenanceJSON(response, request, &body) {
					return
				}
				var task *Task
				var err error
				if segments[3] == "product-catalog-stopped-isolated-switch" {
					task, err = manager.StartProductCatalogIsolatedSwitchWithinContinuouslyStoppedWindow(segments[2], session.Subject, body)
				} else {
					task, err = manager.StartVerifiedPairedRollbackOfStoppedProductCatalogSwitch(segments[2], session.Subject, body)
				}
				writeMaintenanceTask(response, task, err)
				return
			case "product-catalog-stopped-detail-cache-invalidation":
				var body ProductCatalogStoppedSwitchConfirmationRequest
				if !readMaintenanceJSON(response, request, &body) {
					return
				}
				task, err := manager.StartStoppedProductDetailCacheInvalidation(segments[2], session.Subject, body)
				writeMaintenanceTask(response, task, err)
				return
			case "product-catalog-stopped-post-switch-inspection":
				var body struct{}
				if !readMaintenanceJSON(response, request, &body) {
					return
				}
				report, err := manager.InspectStoppedProductCatalogPostSwitch(request.Context(), segments[2])
				if err != nil {
					writeMaintenanceError(response, err)
					return
				}
				writeMaintenanceJSON(response, http.StatusOK, map[string]any{"report": report})
				return
			case "product-catalog-stopped-handoff-preparation":
				var body ProductCatalogStoppedHandoffPreparationRequest
				if !readMaintenanceJSON(response, request, &body) {
					return
				}
				task, err := manager.StartProductCatalogContinuouslyStoppedHandoffPreparation(segments[2], session.Subject, body)
				writeMaintenanceTask(response, task, err)
				return
			case "product-catalog-stopped-handoff-cancellation":
				var body ProductCatalogStoppedHandoffCancellationRequest
				if !readMaintenanceJSON(response, request, &body) {
					return
				}
				task, err := manager.StartVerifiedCancellationOfProductCatalogStoppedHandoffPreparation(segments[2], session.Subject, body)
				writeMaintenanceTask(response, task, err)
				return
			}
		}
		if len(segments) == 4 && segments[0] == "api" && segments[1] == "tasks" && segments[3] == "product-view-stopped-drain-and-rebuild" && request.Method == http.MethodPost {
			var body ProductViewStoppedDrainAndCandidateRebuildRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			task, err := manager.StartStoppedProductViewDrainAndFreshCandidateRebuild(segments[2], session.Subject, body)
			writeMaintenanceTask(response, task, err)
			return
		}
		if len(segments) == 4 && segments[0] == "api" && segments[1] == "tasks" && segments[3] == "customer-service-effective-selection-inspection" && request.Method == http.MethodPost {
			var body CustomerServiceEffectiveSelectionInspectionRequest
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			report, err := manager.InspectCustomerServiceRecoveryEffectiveSelection(request.Context(), segments[2], body)
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"report": report})
			return
		}
		if len(segments) == 4 && segments[0] == "api" && segments[1] == "tasks" && segments[3] == "customer-service-conflict-review" && (request.Method == http.MethodGet || request.Method == http.MethodPost) {
			var review *CustomerServiceRecoveryConflictReview
			var err error
			if request.Method == http.MethodGet {
				review, err = manager.ReadCustomerServiceRecoveryConflictReview(segments[2])
			} else {
				var body CustomerServiceRecoveryConflictReviewRequest
				if !readMaintenanceJSON(response, request, &body) {
					return
				}
				review, err = manager.SaveCustomerServiceRecoveryConflictReview(request.Context(), segments[2], session.Subject, body)
			}
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"review": review})
			return
		}
		if len(segments) == 4 && segments[0] == "api" && segments[1] == "tasks" && segments[3] == "reconciliation-report" && request.Method == http.MethodGet {
			report, err := manager.InspectReconciliation(segments[2])
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"report": report})
			return
		}
		if len(segments) == 3 && segments[0] == "api" && segments[1] == "tasks" && request.Method == http.MethodGet {
			task, err := manager.GetTask(segments[2])
			writeMaintenanceTask(response, task, err)
			return
		}
		if len(segments) == 4 && segments[0] == "api" && segments[1] == "tasks" && segments[3] == "product-catalog-handoff-inspection" && request.Method == http.MethodPost {
			var body struct{}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			report, err := manager.InspectProductCatalogCandidateHandoff(request.Context(), segments[2])
			if err != nil {
				writeMaintenanceError(response, err)
				return
			}
			writeMaintenanceJSON(response, http.StatusOK, map[string]any{"report": report})
			return
		}
		if len(segments) == 4 && segments[0] == "api" && segments[1] == "uploads" {
			id, action := segments[2], segments[3]
			if action == "chunks" && request.Method == http.MethodPut {
				offset, err := strconv.ParseInt(request.Header.Get("Upload-Offset"), 10, 64)
				if err != nil {
					writeMaintenanceError(response, errors.New("invalid chunk offset"))
					return
				}
				data, err := readUploadChunk(request.Body)
				if err != nil {
					writeMaintenanceError(response, err)
					return
				}
				task, err := manager.AppendUploadChunk(id, offset, data, request.Header.Get("X-Chunk-SHA256"))
				writeMaintenanceTask(response, task, err)
				return
			}
			if action == "commit" && request.Method == http.MethodPost {
				task, err := manager.CommitUpload(id)
				writeMaintenanceTask(response, task, err)
				return
			}
		}
		if len(segments) == 4 && segments[0] == "api" && segments[1] == "tasks" && request.Method == http.MethodPost {
			id, action := segments[2], segments[3]
			if action == "module-dependency-report" {
				var body struct {
					Module string `json:"module"`
				}
				if !readMaintenanceJSON(response, request, &body) {
					return
				}
				report, err := manager.InspectVerifiedPackageModuleDependencies(request.Context(), id, body.Module)
				if err != nil {
					writeMaintenanceError(response, err)
					return
				}
				writeMaintenanceJSON(response, http.StatusOK, map[string]any{"report": report})
				return
			}
			if action == "verify" {
				task, err := manager.VerifyUploadedPackage(id)
				writeMaintenanceTask(response, task, err)
				return
			}
			if action == "download-ticket" {
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
				http.SetCookie(response, &http.Cookie{Name: "local_backup_download", Value: value, Path: cookiePath, HttpOnly: true, Secure: secure, SameSite: http.SameSiteStrictMode, MaxAge: 900}) // #nosec G124 -- Secure follows the TLS/proxy scheme; HTTP is supported for the private SSH loopback console, with HttpOnly and SameSiteStrict.
				writeMaintenanceJSON(response, http.StatusOK, map[string]string{"download_path": cookiePath})
				return
			}
			var body struct {
				Confirmation string                 `json:"confirmation"`
				Evidence     ReconciliationEvidence `json:"evidence"`
			}
			if !readMaintenanceJSON(response, request, &body) {
				return
			}
			switch action {
			case "reconciliation-connections":
				err := manager.OpenReconciliationConnections(id, body.Confirmation)
				writeMaintenanceTask(response, nil, err)
				return
			case "apply":
				task, err := manager.ApplyVerifiedRecovery(id, body.Confirmation)
				writeMaintenanceTask(response, task, err)
				return
			case "rollback":
				task, err := manager.RollbackRecovery(id, body.Confirmation)
				writeMaintenanceTask(response, task, err)
				return
			case "complete":
				body.Evidence.Operator = session.Subject
				err := manager.CompleteReconciledRecovery(id, body.Confirmation, body.Evidence)
				writeMaintenanceTask(response, nil, err)
				return
			case "release":
				err := manager.ReleaseUnmodifiedMaintenance(id)
				writeMaintenanceTask(response, nil, err)
				return
			}
		}
		http.NotFound(response, request)
	})
}

func readMaintenanceJSON(response http.ResponseWriter, request *http.Request, value any) bool {
	request.Body = http.MaxBytesReader(response, request.Body, 32<<10)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		writeMaintenanceError(response, errors.New("invalid request body"))
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeMaintenanceError(response, errors.New("unexpected trailing request body"))
		return false
	}
	return true
}
func writeMaintenanceJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}
func writeMaintenanceError(response http.ResponseWriter, err error) {
	writeMaintenanceJSON(response, http.StatusConflict, map[string]string{"error": err.Error()})
}
func writeMaintenanceTask(response http.ResponseWriter, task *Task, err error) {
	if err != nil {
		writeMaintenanceError(response, err)
		return
	}
	writeMaintenanceJSON(response, http.StatusOK, map[string]any{"task": summarizeTaskForMaintenanceHTTP(task)})
}

func summarizeTaskForMaintenanceHTTP(task *Task) *Task {
	if task == nil {
		return nil
	}
	summary := *task
	if task.Manifest != nil {
		manifest := *task.Manifest
		manifest.Files = nil
		manifest.MediaFileCount = 0
		manifest.MediaFileBytes = 0
		for _, entry := range task.Manifest.Files {
			if strings.HasPrefix(entry.Path, "media/") || strings.HasPrefix(entry.Path, "product-static/") {
				manifest.MediaFileCount++
				manifest.MediaFileBytes += entry.Size
			}
		}
		summary.Manifest = &manifest
	}
	return &summary
}
