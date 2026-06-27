package admin

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/auth"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/repository"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
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

	users, total, err := h.userRepo.FindAllWithFilters(page, pageSize, role, status, search)
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

	user, err := h.userRepo.FindByID(uint(id))
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
	var req struct {
		Email     string `json:"email" binding:"required,email"`
		Username  string `json:"username" binding:"required,min=3,max=50"`
		Password  string `json:"password" binding:"required,min=6"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Role      string `json:"role" binding:"required,oneof=user admin manager editor support agent viewer"`
		Locale    string `json:"locale"`
		Status    string `json:"status" binding:"required,oneof=active inactive suspended"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查邮箱是否已存在
	existingUser, _ := h.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// 检查用户名是否已存在
	existingUser, _ = h.userRepo.FindByUsername(req.Username)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// 创建用户
	newUser := &user.User{
		Email:     req.Email,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      string(auth.NormalizeRole(req.Role)),
		Locale:    req.Locale,
		Status:    req.Status,
	}

	// 加密密码
	if err := newUser.HashPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := h.userRepo.Create(newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
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

	var req struct {
		Email     string `json:"email" binding:"omitempty,email"`
		Username  string `json:"username" binding:"omitempty,min=3,max=50"`
		Password  string `json:"password" binding:"omitempty,min=6"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Role      string `json:"role" binding:"omitempty,oneof=user admin manager editor support agent viewer"`
		Locale    string `json:"locale"`
		Status    string `json:"status" binding:"omitempty,oneof=active inactive suspended"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取现有用户
	existingUser, err := h.userRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 更新字段
	if req.Email != "" && req.Email != existingUser.Email {
		// 检查新邮箱是否已被使用
		emailUser, _ := h.userRepo.FindByEmail(req.Email)
		if emailUser != nil && emailUser.ID != existingUser.ID {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		existingUser.Email = req.Email
	}

	if req.Username != "" && req.Username != existingUser.Username {
		// 检查新用户名是否已被使用
		usernameUser, _ := h.userRepo.FindByUsername(req.Username)
		if usernameUser != nil && usernameUser.ID != existingUser.ID {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}
		existingUser.Username = req.Username
	}

	if req.Password != "" {
		if err := existingUser.HashPassword(req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
	}

	if req.FirstName != "" {
		existingUser.FirstName = req.FirstName
	}
	if req.LastName != "" {
		existingUser.LastName = req.LastName
	}
	if req.Role != "" {
		existingUser.Role = string(auth.NormalizeRole(req.Role))
	}
	if req.Locale != "" {
		existingUser.Locale = req.Locale
	}
	if req.Status != "" {
		existingUser.Status = req.Status
	}

	if err := h.userRepo.Update(existingUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    existingUser.ToResponse(),
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

	// 不允许删除自己
	currentUserID, _ := c.Get("user_id")
	if currentUserID.(uint) == uint(id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete yourself"})
		return
	}

	if err := h.userRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
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

	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive suspended"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 不允许修改自己的状态
	currentUserID, _ := c.Get("user_id")
	if currentUserID.(uint) == uint(id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify your own status"})
		return
	}

	if err := h.userRepo.UpdateStatus(uint(id), req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User status updated successfully",
	})
}

// GetUserStats 获取用户统计
// GET /api/admin/users/stats
func (h *UserHandler) GetUserStats(c *gin.Context) {
	stats, err := h.userRepo.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// BatchDeleteUsers 批量删除用户
// POST /api/admin/users/batch-delete
func (h *UserHandler) BatchDeleteUsers(c *gin.Context) {
	var req struct {
		UserIDs []uint `json:"user_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 不允许删除自己
	currentUserID, _ := c.Get("user_id")
	for _, id := range req.UserIDs {
		if id == currentUserID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete yourself"})
			return
		}
	}

	deleted := 0
	for _, id := range req.UserIDs {
		if err := h.userRepo.Delete(id); err == nil {
			deleted++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch delete completed",
		"deleted": deleted,
		"total":   len(req.UserIDs),
	})
}
