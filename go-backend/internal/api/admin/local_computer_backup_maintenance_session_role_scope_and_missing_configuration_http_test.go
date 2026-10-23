package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/localbackup"
)

func TestLocalBackupSessionPreservesRoleScopesWithoutOpeningBusinessDatabase(t *testing.T) {
	keyFile := filepath.Join(t.TempDir(), "independent-maintenance-session-authentication-key")
	require.NoError(t, os.WriteFile(keyFile, []byte(strings.Repeat("synthetic-session-key-", 3)), 0600))
	t.Setenv("LOCAL_BACKUP_AUTHENTICATION_KEY_FILE", keyFile)
	t.Setenv("LOCAL_BACKUP_SITE_ID", "synthetic-store")
	t.Setenv("LOCAL_BACKUP_ENVIRONMENT", "test")
	for _, role := range []auth.Role{auth.RoleAdmin, auth.RoleManager, auth.RoleSupport} {
		t.Run(string(role), func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) { c.Set("user_role", string(role)); c.Set("user_id", uint(42)) })
			router.POST("/session", middleware.RequirePermission(auth.PermOpsBackupView), issueLocalComputerBackupMaintenanceSession)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/session", strings.NewReader(`{}`)))
			if role == auth.RoleSupport {
				require.Equal(t, 403, response.Code)
				return
			}
			require.Equal(t, 200, response.Code)
			var body struct {
				Token string `json:"token"`
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
			cfg := localbackup.Configuration{SiteID: "synthetic-store", Environment: "test", AuthenticationKeyFile: keyFile}
			session, err := cfg.AuthenticateMaintenanceSession(body.Token)
			require.NoError(t, err)
			require.Equal(t, "admin:42", session.Subject)
			require.True(t, session.HasPermission(localbackup.PermissionView))
			require.Equal(t, role == auth.RoleAdmin, session.HasPermission(localbackup.PermissionRestore))
			require.Equal(t, role == auth.RoleAdmin, session.HasPermission(localbackup.PermissionExport))
			cfg.SiteID = "another-store"
			_, err = cfg.AuthenticateMaintenanceSession(body.Token)
			require.Error(t, err)
		})
	}
	t.Setenv("LOCAL_BACKUP_AUTHENTICATION_KEY_FILE", filepath.Join(t.TempDir(), "not-configured"))
	router := gin.New()
	router.POST("/session", issueLocalComputerBackupMaintenanceSession)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/session", nil))
	require.Equal(t, 503, response.Code)
}
