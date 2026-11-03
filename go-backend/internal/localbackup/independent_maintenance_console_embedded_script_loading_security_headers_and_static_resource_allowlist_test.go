package localbackup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	adminstyletokens "commerce-platform/web/admin/src/styles"

	"github.com/stretchr/testify/require"
)

func TestMaintenanceConsoleDocumentLoadsEveryEmbeddedResponsibilityScriptExactlyOnceBeforeEventEntry(t *testing.T) {
	manager := &Manager{}
	handler := manager.HTTPHandler()
	files, err := fs.Glob(maintenanceConsoleFiles, "maintenance_console/*.js")
	require.NoError(t, err)
	require.Greater(t, len(files), 1)
	scriptPattern := regexp.MustCompile(`<script src="(maintenance-console(?:-[a-z-]+)?\.js)" defer></script>`)
	for _, prefix := range []string{"", ConsolePrefix} {
		document := httptest.NewRecorder()
		handler.ServeHTTP(document, httptest.NewRequestWithContext(context.Background(), http.MethodGet, prefix+"/", nil))
		require.Equal(t, http.StatusOK, document.Code)
		scripts := scriptPattern.FindAllStringSubmatch(document.Body.String(), -1)
		require.Len(t, scripts, len(files), "every embedded script must be part of the actual document")
		require.Equal(t, "maintenance-console.js", scripts[len(scripts)-1][1])
		seen := map[string]bool{}
		for _, script := range scripts {
			name := script[1]
			require.False(t, seen[name], "script loaded twice: %s", name)
			seen[name] = true
			expected, err := maintenanceConsoleFiles.ReadFile("maintenance_console/" + name)
			require.NoError(t, err)
			for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodDelete} {
				response := httptest.NewRecorder()
				handler.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), method, prefix+"/"+name, nil))
				require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
				require.Equal(t, "nosniff", response.Header().Get("X-Content-Type-Options"))
				require.Contains(t, response.Header().Get("Content-Security-Policy"), "script-src 'self'")
				if method == http.MethodGet || method == http.MethodHead {
					require.Equal(t, http.StatusOK, response.Code, "%s %s", method, name)
					require.Contains(t, response.Header().Get("Content-Type"), "javascript")
					if method == http.MethodGet {
						require.Equal(t, string(expected), response.Body.String(), "must serve script bytes, not an HTML fallback")
					} else {
						require.Empty(t, response.Body.String())
					}
				} else {
					require.Equal(t, http.StatusMethodNotAllowed, response.Code)
				}
			}
		}
		for _, file := range files {
			require.True(t, seen[strings.TrimPrefix(file, "maintenance_console/")])
		}
	}
}

func TestMaintenanceConsoleStaticAllowlistDoesNotExposeUnregisteredPathsOrBypassAPIAuthentication(t *testing.T) {
	configuration := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(configuration)
	require.NoError(t, err)
	defer manager.Close()
	handler := manager.HTTPHandler()
	for _, prefix := range []string{"", ConsolePrefix} {
		for _, path := range []string{
			"/maintenance-console-unregistered.js", "/maintenance-console.js.map", "/maintenance-console.js/",
			"/maintenance_console/maintenance-console.js", "/../maintenance-console.js", "/index.html/private.json",
			"/MapleUI-CJK.f8ce6d72e8cb.woff2.bak", "/embedded_existing_maple_ui_cjk_font_for_independent_maintenance_page.go",
			"/embedded_shared_admin_theme_tokens_for_independent_maintenance_page.go",
		} {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, prefix+path, nil))
			require.Equal(t, http.StatusNotFound, response.Code, "%s", prefix+path)
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, prefix+"/api/status", nil))
		require.Equal(t, http.StatusUnauthorized, response.Code)
	}
}

func TestIndependentMaintenanceConsoleServesTheSameThemeTokenFileAsTheAdminWithoutApplicationAccess(t *testing.T) {
	manager := &Manager{}
	handler := manager.HTTPHandler()
	expected, err := adminstyletokens.SharedAdminThemeTokenFiles.ReadFile(adminstyletokens.SharedAdminThemeTokensFilename)
	require.NoError(t, err)
	for _, prefix := range []string{"", ConsolePrefix} {
		for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost} {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), method, prefix+"/"+adminstyletokens.SharedAdminThemeTokensFilename, nil))
			if method == http.MethodPost {
				require.Equal(t, http.StatusMethodNotAllowed, response.Code)
				continue
			}
			require.Equal(t, http.StatusOK, response.Code)
			require.Contains(t, response.Header().Get("Content-Type"), "text/css")
			if method == http.MethodHead {
				require.Empty(t, response.Body.Bytes())
			} else {
				require.Equal(t, expected, response.Body.Bytes())
			}
		}
	}
}

func TestIndependentMaintenanceConsoleServesOnlyApprovedExistingAdminFontWithoutApplicationAccess(t *testing.T) {
	// An empty manager proves typography does not require a business database,
	// an admin session, or a running admin web server during recovery.
	manager := &Manager{}
	handler := manager.HTTPHandler()
	for _, prefix := range []string{"", ConsolePrefix} {
		for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost} {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), method, prefix+"/MapleUI-CJK.f8ce6d72e8cb.woff2", nil))
			if method == http.MethodPost {
				require.Equal(t, http.StatusMethodNotAllowed, response.Code)
				continue
			}
			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, "font/woff2", response.Header().Get("Content-Type"))
			require.Equal(t, "nosniff", response.Header().Get("X-Content-Type-Options"))
			if method == http.MethodHead {
				require.Empty(t, response.Body.Bytes())
				continue
			}
			digest := sha256.Sum256(response.Body.Bytes())
			require.Equal(t, "f8ce6d72e8cbda22971300b1367fe49ba98f5f798e0a1a297f4dea90c7fb8be9", hex.EncodeToString(digest[:]))
		}
	}
}
