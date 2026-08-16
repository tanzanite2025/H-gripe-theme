package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/ops"

	"github.com/gin-gonic/gin"
)

func TestWorkflowCancelPermissionByMode(t *testing.T) {
	tests := []struct {
		name       string
		mode       string
		permission auth.Permission
		ok         bool
	}{
		{
			name:       "dry run",
			mode:       ops.DeploymentWorkflowModeDryRun,
			permission: auth.PermOpsDeployDryRun,
			ok:         true,
		},
		{
			name:       "production",
			mode:       ops.DeploymentWorkflowModeProduction,
			permission: auth.PermOpsDeployExecute,
			ok:         true,
		},
		{
			name: "unknown mode",
			mode: "unknown",
			ok:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			permission, ok := workflowCancelPermission(test.mode)
			if permission != test.permission || ok != test.ok {
				t.Fatalf("workflowCancelPermission(%q) = (%q, %t), want (%q, %t)", test.mode, permission, ok, test.permission, test.ok)
			}
		})
	}
}

func TestWorkflowRollbackPermissionIsDedicated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		role auth.Role
		code int
	}{
		{name: "manager can rollback", role: auth.RoleManager, code: http.StatusOK},
		{name: "editor cannot rollback", role: auth.RoleEditor, code: http.StatusForbidden},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/rollback",
				func(c *gin.Context) {
					c.Set("user_role", string(test.role))
					c.Next()
				},
				middleware.RequirePermission(auth.PermOpsDeployRollback),
				func(c *gin.Context) {
					c.Status(http.StatusOK)
				},
			)
			request := httptest.NewRequest(http.MethodPost, "/rollback", nil)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			if recorder.Code != test.code {
				t.Fatalf("rollback permission for %s = %d, want %d", test.role, recorder.Code, test.code)
			}
		})
	}
}
