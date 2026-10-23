package admin

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/localbackup"
)

func issueLocalComputerBackupMaintenanceSession(c *gin.Context) {
	keyFile := os.Getenv("LOCAL_BACKUP_AUTHENTICATION_KEY_FILE")
	key, err := os.ReadFile(keyFile) // #nosec G304,G703 -- The credential path is administrator-controlled environment configuration, never an HTTP parameter.
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "独立备份维护服务尚未配置；请先按运行手册配置维护进程、私有工作区及独立凭据"})
		return
	}
	role := auth.Role(c.GetString("user_role"))
	permissions := []string{}
	for _, permission := range []auth.Permission{auth.PermOpsBackupView, auth.PermOpsBackupExport, auth.PermOpsBackupUpload, auth.PermOpsBackupRestore} {
		if role.HasPermission(permission) {
			permissions = append(permissions, string(permission))
		}
	}
	actor := fmt.Sprintf("admin:%d", c.GetUint("user_id"))
	token, err := localbackup.IssueMaintenanceSession(strings.TrimRight(string(key), "\r\n"), os.Getenv("LOCAL_BACKUP_SITE_ID"), os.Getenv("LOCAL_BACKUP_ENVIRONMENT"), actor, permissions)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "独立备份维护服务的站点身份或凭据尚未配置"})
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"token": token, "console_path": localbackup.ConsolePrefix + "/", "expires_in_seconds": 1800})
}
