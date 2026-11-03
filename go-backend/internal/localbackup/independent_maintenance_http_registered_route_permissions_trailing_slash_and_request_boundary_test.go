package localbackup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaintenanceDownloadTicketTrailingSlashCannotGrantExportToViewOnlySession(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	task, err := manager.createTask("backup", "synthetic-operator")
	require.NoError(t, err)
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	viewer, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-viewer", []string{PermissionView})
	require.NoError(t, err)
	for _, prefix := range []string{"", ConsolePrefix} {
		for _, suffix := range []string{"", "/", "//"} {
			t.Run(prefix+"/download-ticket"+suffix, func(t *testing.T) {
				request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, prefix+"/api/tasks/"+task.ID+"/download-ticket"+suffix, strings.NewReader(`{}`))
				request.Header.Set("Authorization", "Bearer "+viewer)
				response := httptest.NewRecorder()
				manager.HTTPHandler().ServeHTTP(response, request)
				require.Equal(t, http.StatusForbidden, response.Code)
				require.Empty(t, response.Result().Cookies(), "view permission must never mint an export credential")
			})
		}
	}
}

func TestMaintenanceRegisteredTaskActionsAuthorizeBeforeParsingForEveryScopePrefixAndTrailingSlash(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	permissions := []string{PermissionView, PermissionExport, PermissionUpload, PermissionRestore}
	tokens := map[string]string{}
	for _, permission := range permissions {
		tokens[permission], err = IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-permission-check", []string{permission})
		require.NoError(t, err)
	}
	// Exercise the HTTP boundary with independent expected permissions. A malformed
	// body and invalid identity prevent accepted requests from touching a database.
	routes := []struct{ action, method, permission string }{
		{"download-ticket", http.MethodPost, PermissionExport},
		{"verify", http.MethodPost, PermissionUpload},
		{"selected-module-isolated-restore-preview", http.MethodPost, PermissionRestore},
		{"module-dependency-report", http.MethodPost, PermissionRestore},
		{"selected-customer-service-restore-choice", http.MethodPost, PermissionRestore},
		{"selected-customer-service-candidate-rehearsal", http.MethodPost, PermissionRestore},
		{"customer-service-effective-selection-inspection", http.MethodPost, PermissionRestore},
		{"customer-service-conflict-review", http.MethodPost, PermissionRestore},
		{"customer-service-conflict-review", http.MethodGet, PermissionRestore},
		{"configuration-site-content-formal-restore", http.MethodPost, PermissionRestore},
		{"configuration-site-content-interrupted-restore-recovery", http.MethodPost, PermissionRestore},
		{"configuration-site-content-reviewed-service-release", http.MethodPost, PermissionRestore},
		{"media-module-formal-restore", http.MethodPost, PermissionRestore},
		{"media-module-interrupted-restore-recovery", http.MethodPost, PermissionRestore},
		{"customer-profile-module-formal-restore", http.MethodPost, PermissionRestore},
		{"customer-profile-module-interrupted-restore-recovery", http.MethodPost, PermissionRestore},
		{"customer-service-module-formal-restore", http.MethodPost, PermissionRestore},
		{"customer-service-module-interrupted-restore-recovery", http.MethodPost, PermissionRestore},
		{"order-module-formal-history-restore", http.MethodPost, PermissionRestore},
		{"order-module-interrupted-restore-recovery", http.MethodPost, PermissionRestore},
		{"product-catalog-stopped-read-only-public-image-http-query", http.MethodPost, PermissionRestore},
		{"product-catalog-stopped-read-only-storefront-http-query", http.MethodPost, PermissionRestore},
		{"product-catalog-stopped-read-only-media-http-query", http.MethodPost, PermissionRestore},
		{"product-catalog-stopped-read-only-product-http-query", http.MethodPost, PermissionRestore},
		{"product-catalog-stopped-projection-inspection", http.MethodPost, PermissionRestore},
		{"product-catalog-stopped-projection-evidence", http.MethodPost, PermissionRestore},
		{"product-catalog-formal-restore-acceptance", http.MethodPost, PermissionRestore},
		{"product-catalog-stopped-isolated-switch", http.MethodPost, PermissionRestore},
		{"product-catalog-stopped-switch-rollback", http.MethodPost, PermissionRestore},
		{"product-catalog-stopped-detail-cache-invalidation", http.MethodPost, PermissionRestore},
		{"product-catalog-stopped-post-switch-inspection", http.MethodPost, PermissionRestore},
		{"product-catalog-stopped-handoff-preparation", http.MethodPost, PermissionRestore},
		{"product-catalog-stopped-handoff-cancellation", http.MethodPost, PermissionRestore},
		{"product-view-stopped-drain-and-rebuild", http.MethodPost, PermissionRestore},
		{"product-catalog-handoff-inspection", http.MethodPost, PermissionRestore},
		{"reconciliation-report", http.MethodGet, PermissionRestore},
		{"reconciliation-connections", http.MethodPost, PermissionRestore},
		{"apply", http.MethodPost, PermissionRestore},
		{"rollback", http.MethodPost, PermissionRestore},
		{"complete", http.MethodPost, PermissionRestore},
		{"release", http.MethodPost, PermissionRestore},
	}
	handler := manager.HTTPHandler()
	for _, route := range routes {
		t.Run(route.method+"/"+route.action, func(t *testing.T) {
			for _, permission := range permissions {
				for _, prefix := range []string{"", ConsolePrefix} {
					for _, suffix := range []string{"", "/", "//"} {
						request := httptest.NewRequestWithContext(context.Background(), route.method, prefix+"/api/tasks/invalid-task/"+route.action+suffix, strings.NewReader(`{`))
						request.Header.Set("Authorization", "Bearer "+tokens[permission])
						response := httptest.NewRecorder()
						handler.ServeHTTP(response, request)
						want := http.StatusForbidden
						if permission == route.permission {
							want = http.StatusConflict
						}
						require.Equal(t, want, response.Code, "%s %s permission=%s", route.method, request.URL.Path, permission)
						require.Empty(t, response.Result().Cookies())
					}
				}
			}
		})
	}
}

func TestMaintenanceRegisteredTaskRoutesRetainCrossSiteDownloadScopeAndUnknownBodyBoundaries(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	restorer, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-restorer", []string{PermissionRestore})
	require.NoError(t, err)
	task, err := manager.createTask("backup", "synthetic-operator")
	require.NoError(t, err)
	download, err := cfg.issueDownloadSession("synthetic-operator", task.ID)
	require.NoError(t, err)
	handler := manager.HTTPHandler()
	for _, scenario := range []struct {
		name, method, path, token, body, fetchSite string
		want                                       int
	}{
		{"cross_site_mutation", http.MethodPost, "/api/tasks/invalid/apply", restorer, `{}`, "cross-site", http.StatusForbidden},
		{"download_ticket_cannot_call_restore", http.MethodPost, "/api/tasks/invalid/apply/", download, `{}`, "", http.StatusForbidden},
		{"unknown_action_malformed_body", http.MethodPost, "/api/tasks/invalid/not-registered", key, `{`, "", http.StatusConflict},
		{"unknown_action_strict_fields", http.MethodPost, "/api/tasks/invalid/not-registered", key, `{"unexpected":true}`, "", http.StatusConflict},
		{"unknown_action_valid_body", http.MethodPost, "/api/tasks/invalid/not-registered", key, `{}`, "", http.StatusNotFound},
		{"wrong_method_on_restore", http.MethodGet, "/api/tasks/invalid/apply", restorer, "", "", http.StatusNotFound},
		{"wrong_method_with_body_on_report", http.MethodPost, "/api/tasks/invalid/reconciliation-report", restorer, `{`, "", http.StatusConflict},
		{"trailing_json_rejected", http.MethodPost, "/api/tasks/invalid/apply", restorer, `{} {}`, "", http.StatusConflict},
		{"oversize_body_rejected", http.MethodPost, "/api/tasks/invalid/apply", restorer, `{"confirmation":"` + strings.Repeat("x", 33<<10) + `"}`, "", http.StatusConflict},
		{"static_script_public", http.MethodGet, "/maintenance-console.js", "", "", "", http.StatusOK},
		{"static_script_no_post", http.MethodPost, "/maintenance-console.js", "", "", "", http.StatusMethodNotAllowed},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(context.Background(), scenario.method, scenario.path, strings.NewReader(scenario.body))
			if scenario.token != "" {
				request.Header.Set("Authorization", "Bearer "+scenario.token)
			}
			request.Header.Set("Sec-Fetch-Site", scenario.fetchSite)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			require.Equal(t, scenario.want, response.Code)
			require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
			require.Equal(t, "nosniff", response.Header().Get("X-Content-Type-Options"))
		})
	}
}
