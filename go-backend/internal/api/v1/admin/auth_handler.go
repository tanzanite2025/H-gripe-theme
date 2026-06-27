package admin

import (
	"net/http"
	"tanzanite/internal/domain/auth"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// AdminLogin 管理员登录
// POST /api/admin/auth/login
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 登录
	token, user, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// 检查用户角色（只允许管理员角色登录）
	role := auth.NormalizeRole(user.Role)
	if role != auth.RoleAdmin && role != auth.RoleManager && role != auth.RoleEditor && role != auth.RoleSupport {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "You don't have permission to access the admin panel",
		})
		return
	}

	// 设置 HttpOnly Cookie
	c.SetCookie("auth_token", token, 3600*24*7, "/", "", true, true)

	// 返回用户信息和权限
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":          user.ID,
			"email":       user.Email,
			"username":    user.Username,
			"first_name":  user.FirstName,
			"last_name":   user.LastName,
			"role":        user.Role,
			"permissions": role.GetPermissions(),
		},
	})
}

// GetProfile 获取当前用户信息
// GET /api/admin/auth/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	role := auth.Role(user.Role)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":          user.ID,
			"email":       user.Email,
			"username":    user.Username,
			"first_name":  user.FirstName,
			"last_name":   user.LastName,
			"role":        user.Role,
			"locale":      user.Locale,
			"status":      user.Status,
			"created_at":  user.CreatedAt,
			"permissions": role.GetPermissions(),
		},
	})
}

// RefreshToken 刷新令牌
// POST /api/admin/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 生成新令牌
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.SetCookie("auth_token", token, 3600*24*7, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
	})
}

// Logout 登出
// POST /api/admin/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// 在实际应用中，可以在这里将令牌加入黑名单
	c.SetCookie("auth_token", "", -1, "/", "", true, true)

	// 目前只返回成功消息
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// GetPermissions 获取当前用户权限
// GET /api/admin/auth/permissions
func (h *AuthHandler) GetPermissions(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	role := auth.Role(userRole.(string))
	permissions := role.GetPermissions()

	c.JSON(http.StatusOK, gin.H{
		"role":        role.String(),
		"permissions": permissions,
	})
}
