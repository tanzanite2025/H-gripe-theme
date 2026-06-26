package auth

import (
	"net/http"
	"strconv"
	
	"tanzanite/internal/repository"
	
	"github.com/gin-gonic/gin"
)

type BrowsingHistoryHandler struct {
	userRepo *repository.UserRepository
}

func NewBrowsingHistoryHandler(userRepo *repository.UserRepository) *BrowsingHistoryHandler {
	return &BrowsingHistoryHandler{userRepo: userRepo}
}

// AddBrowsingHistory 添加浏览历史
// POST /api/v1/user/browsing-history
func (h *BrowsingHistoryHandler) AddBrowsingHistory(c *gin.Context) {
	// 获取用户ID（需要认证）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		ProductID uint `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 添加到浏览历史
	if err := h.userRepo.AddBrowsingHistory(userID.(uint), req.ProductID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save browsing history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Browsing history saved"
	})
}

// GetBrowsingHistory 获取浏览历史
// GET /api/v1/user/browsing-history?limit=20
func (h *BrowsingHistoryHandler) GetBrowsingHistory(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 获取limit参数
	limit := 20 // 默认20条
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// 查询浏览历史
	history, err := h.userRepo.GetBrowsingHistory(userID.(uint), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get browsing history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"count":   len(history)
	})
}

// DeleteBrowsingHistory 删除特定浏览记录
// DELETE /api/v1/user/browsing-history/:product_id
func (h *BrowsingHistoryHandler) DeleteBrowsingHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	productID, err := strconv.ParseUint(c.Param("product_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	if err := h.userRepo.DeleteBrowsingHistory(userID.(uint), uint(productID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete browsing history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Browsing history deleted"
	})
}

// ClearBrowsingHistory 清空浏览历史
// DELETE /api/v1/user/browsing-history
func (h *BrowsingHistoryHandler) ClearBrowsingHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.userRepo.ClearBrowsingHistory(userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear browsing history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Browsing history cleared"
	})
}
