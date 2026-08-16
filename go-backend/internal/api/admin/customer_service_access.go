package admin

import (
	"commerce-platform/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

// Customer-service access reads the canonical role claim set by AuthMiddleware.
// Do not couple individual handlers to compatibility context keys.
func adminCustomerServiceRole(c *gin.Context) auth.Role {
	if c == nil {
		return auth.RoleUser
	}
	value, _ := c.Get("role")
	rawRole, _ := value.(string)
	return auth.NormalizeRole(rawRole)
}

func adminCustomerServiceScope(c *gin.Context) (uint, bool) {
	if c == nil {
		return 0, false
	}
	value, _ := c.Get("user_id")
	userID, _ := value.(uint)
	role := adminCustomerServiceRole(c)
	return userID, role == auth.RoleAdmin || role == auth.RoleManager
}

func adminCustomerServiceCanEdit(c *gin.Context) bool {
	return adminCustomerServiceRole(c).HasPermission(auth.PermTicketEdit)
}
