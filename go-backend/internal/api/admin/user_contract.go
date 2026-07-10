package admin

import (
	"errors"
	"net/http"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type userCreateRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role" binding:"required,oneof=user admin manager editor support viewer"`
	Locale    string `json:"locale"`
	Status    string `json:"status" binding:"required,oneof=active inactive suspended"`
}

type userUpdateRequest struct {
	Email     string `json:"email" binding:"omitempty,email"`
	Username  string `json:"username" binding:"omitempty,min=3,max=50"`
	Password  string `json:"password" binding:"omitempty,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role" binding:"omitempty,oneof=user admin manager editor support viewer"`
	Locale    string `json:"locale"`
	Status    string `json:"status" binding:"omitempty,oneof=active inactive suspended"`
}

type userStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive suspended"`
}

type userBatchDeleteRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required,min=1"`
}

func currentAdminActor(c *gin.Context) (uint, string, bool) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return 0, "", false
	}

	userID, ok := userIDValue.(uint)
	if !ok || userID == 0 {
		return 0, "", false
	}

	roleValue, exists := c.Get("user_role")
	if !exists {
		roleValue, exists = c.Get("role")
	}
	if !exists {
		return 0, "", false
	}

	role, ok := roleValue.(string)
	return userID, role, ok && role != ""
}

func respondUserServiceError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	case errors.Is(err, service.ErrEmailExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
	case errors.Is(err, service.ErrUsernameExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
	case errors.Is(err, service.ErrSelfDelete):
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete yourself"})
	case errors.Is(err, service.ErrSelfStatusChange):
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify your own status"})
	case errors.Is(err, service.ErrSelfRoleChange):
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify your own role"})
	case errors.Is(err, service.ErrRoleForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient privileges for requested role change"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
	}
}
