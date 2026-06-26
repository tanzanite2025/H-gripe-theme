package middleware

import (
	"net/http"
	"tanzanite/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

// RequirePermission 权限检查中间件
func RequirePermission(permission auth.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户角色（由 AuthMiddleware 设置）
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		role := auth.Role(userRole.(string))

		// 检查角色是否有效
		if !role.IsValid() {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role"})
			c.Abort()
			return
		}

		// 检查权限
		if !role.HasPermission(permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Permission denied",
				"message": "You don't have permission to access this resource",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission 要求任意一个权限
func RequireAnyPermission(permissions ...auth.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		role := auth.Role(userRole.(string))

		if !role.IsValid() {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role"})
			c.Abort()
			return
		}

		// 检查是否拥有任意一个权限
		hasPermission := false
		for _, perm := range permissions {
			if role.HasPermission(perm) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Permission denied",
				"message": "You don't have permission to access this resource",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllPermissions 要求所有权限
func RequireAllPermissions(permissions ...auth.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		role := auth.Role(userRole.(string))

		if !role.IsValid() {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role"})
			c.Abort()
			return
		}

		// 检查是否拥有所有权限
		for _, perm := range permissions {
			if !role.HasPermission(perm) {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Permission denied",
					"message": "You don't have permission to access this resource",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequireRoleAuth 角色检查中间件（新版本，使用 auth.Role）
func RequireRoleAuth(roles ...auth.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		role := auth.Role(userRole.(string))

		// 检查是否是指定角色之一
		hasRole := false
		for _, r := range roles {
			if role == r {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Access denied",
				"message": "You don't have the required role to access this resource",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireBackofficeAccess requires an active staff role for admin surfaces.
func RequireBackofficeAccess() gin.HandlerFunc {
	return RequireRoleAuth(auth.RoleAdmin, auth.RoleManager, auth.RoleEditor, auth.RoleSupport)
}

// AdminOnly 仅管理员中间件
func AdminOnly() gin.HandlerFunc {
	return RequireRoleAuth(auth.RoleAdmin)
}

// ManagerOrAbove 经理及以上中间件
func ManagerOrAbove() gin.HandlerFunc {
	return RequireRoleAuth(auth.RoleAdmin, auth.RoleManager)
}
