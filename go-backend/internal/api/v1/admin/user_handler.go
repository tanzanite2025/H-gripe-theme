package admin

import (
	"net/http"
	"strconv"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// ListUsers 获取用户列表
// GET /api/admin/users
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	role := c.Query("role")
	status := c.Query("status")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := h.userService.ListUsers(page, pageSize, role, status, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// 转换为响应格式（不包含密码）
	userResponses := make([]interface{}, len(users))
	for i, u := range users {
		userResponses[i] = u.ToResponse()
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"users": userResponses,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetUser 获取用户详情
// GET /api/admin/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUser(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user.ToResponse(),
	})
}

// CreateUser 创建用户
// POST /api/admin/users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req userCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, actorRole, ok := currentAdminActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid admin session"})
		return
	}

	newUser, err := h.userService.CreateUser(service.UserCreateInput{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
		Locale:    req.Locale,
		Status:    req.Status,
	}, actorRole)
	if err != nil {
		respondUserServiceError(c, err, "Failed to create user")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    newUser.ToResponse(),
	})
}

// UpdateUser 更新用户
// PUT /api/admin/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req userUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID, actorRole, ok := currentAdminActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid admin session"})
		return
	}

	updatedUser, err := h.userService.UpdateUser(uint(id), actorID, actorRole, service.UserUpdateInput{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
		Locale:    req.Locale,
		Status:    req.Status,
	})
	if err != nil {
		respondUserServiceError(c, err, "Failed to update user")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    updatedUser.ToResponse(),
	})
}

// DeleteUser 删除用户
// DELETE /api/admin/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	actorID, _, ok := currentAdminActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid admin session"})
		return
	}

	if err := h.userService.DeleteUser(uint(id), actorID); err != nil {
		respondUserServiceError(c, err, "Failed to delete user")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// UpdateUserStatus 更新用户状态
// PATCH /api/admin/users/:id/status
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req userStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID, actorRole, ok := currentAdminActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid admin session"})
		return
	}

	if err := h.userService.UpdateUserStatus(uint(id), actorID, actorRole, req.Status); err != nil {
		respondUserServiceError(c, err, "Failed to update user status")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User status updated successfully",
	})
}

// GetUserStats 获取用户统计
// GET /api/admin/users/stats
func (h *UserHandler) GetUserStats(c *gin.Context) {
	stats, err := h.userService.GetUserStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// BatchDeleteUsers 批量删除用户
// POST /api/admin/users/batch-delete
func (h *UserHandler) BatchDeleteUsers(c *gin.Context) {
	var req userBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID, _, ok := currentAdminActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid admin session"})
		return
	}

	deleted, err := h.userService.BatchDeleteUsers(req.UserIDs, actorID)
	if err != nil {
		respondUserServiceError(c, err, "Batch delete failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch delete completed",
		"deleted": deleted,
		"total":   len(req.UserIDs),
	})
}
