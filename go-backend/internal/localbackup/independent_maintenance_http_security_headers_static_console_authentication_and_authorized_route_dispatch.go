package localbackup

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"strings"

	mapleuifont "commerce-platform/web/admin/src/assets/fonts/maple-ui"
	adminstyletokens "commerce-platform/web/admin/src/styles"
)

// Authenticate, authorize and audit before dispatch. A task action obtains its permission from the same registration as its handler.
//
//go:embed maintenance_console/*
var maintenanceConsoleFiles embed.FS

const ConsolePrefix = "/local-backup-maintenance"

// Keep this explicit: embedding a file must not automatically publish it.
func isPublicMaintenanceConsoleStaticResourcePath(path string) bool {
	switch path {
	case "/", "/index.html", "/maintenance-console.css", "/maintenance-console.js",
		"/" + mapleuifont.ApprovedMapleUICJKFontFilename,
		"/" + adminstyletokens.SharedAdminThemeTokensFilename,
		"/maintenance-console-admin-shared-action-buttons-and-selection-controls.css",
		"/maintenance-console-embedded-admin-presentation-and-theme-synchronization.js",
		"/maintenance-console-shared-dom-formatting-and-task-state-labels.js",
		"/maintenance-console-session-authenticated-requests-and-task-refresh.js",
		"/maintenance-console-business-module-backup-selection-and-range-preview.js",
		"/maintenance-console-task-list-actions-and-package-scope-preview.js",
		"/maintenance-console-selected-module-isolated-restore-preview-and-dependency-reports.js",
		"/maintenance-console-order-history-restore-eligibility-confirmation-and-row-inspection.js",
		"/maintenance-console-customer-profile-restore-eligibility-and-confirmation.js",
		"/maintenance-console-configuration-site-content-restore-and-reviewed-service-release.js",
		"/maintenance-console-media-restore-eligibility-confirmation-and-interrupted-recovery.js",
		"/maintenance-console-customer-service-saved-strategy-rehearsal-and-formal-restore.js",
		"/maintenance-console-customer-service-conversation-recovery-plan-and-conflict-review.js",
		"/maintenance-console-product-candidate-stopped-handoff-switch-and-acceptance.js",
		"/maintenance-console-stopped-product-projection-evidence-and-read-only-http-queries.js",
		"/maintenance-console-recovery-durable-event-and-side-effect-receipt-reports.js",
		"/maintenance-console-local-package-resumable-upload-and-ticket-download.js",
		"/maintenance-console-external-projection-evidence-and-whole-storefront-reconciliation.js":
		return true
	default:
		return false
	}
}

func (manager *Manager) HTTPHandler() http.Handler {
	staticFiles, _ := fs.Sub(maintenanceConsoleFiles, "maintenance_console")
	static := http.FileServer(http.FS(staticFiles))
	approvedFont := http.FileServer(http.FS(mapleuifont.ApprovedMapleUICJKFontFiles))
	sharedAdminTheme := http.FileServer(http.FS(adminstyletokens.SharedAdminThemeTokenFiles))
	taskActions := manager.registeredMaintenanceHTTPTaskActions()
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
		if isPublicMaintenanceConsoleStaticResourcePath(path) {
			if request.Method != http.MethodGet && request.Method != http.MethodHead {
				response.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			clone := request.Clone(request.Context())
			clone.URL.Path = path
			if path == "/"+mapleuifont.ApprovedMapleUICJKFontFilename {
				approvedFont.ServeHTTP(response, clone)
				return
			}
			if path == "/"+adminstyletokens.SharedAdminThemeTokensFilename {
				sharedAdminTheme.ServeHTTP(response, clone)
				return
			}
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
		segments := strings.Split(strings.Trim(path, "/"), "/")
		var actionDefinition maintenanceHTTPTaskActionDefinition
		registeredAction := false
		if len(segments) == 4 && segments[0] == "api" && segments[1] == "tasks" {
			actionDefinition, registeredAction = taskActions[segments[3]]
		}
		permission := requiredMaintenancePermissionForUnregisteredRequestPath(path)
		if registeredAction {
			permission = actionDefinition.Permission
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
			manager.serveScopedMaintenancePackageDownload(response, request, session, path)
			return
		}
		if manager.serveMaintenanceStatusTasksBackupAndUploadRoute(response, request, session, path, segments) {
			return
		}
		if registeredAction {
			for _, method := range actionDefinition.Methods {
				if request.Method == method {
					actionDefinition.Handle(response, request, session, segments)
					return
				}
			}
		}
		if serveUnregisteredMaintenanceTaskPostBodyCompatibility(response, request, segments) {
			return
		}
		http.NotFound(response, request)

	})
}
