package auth

import (
	"strconv"

	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"
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
		apierror.RespondUnauthorized(c)
		return
	}

	var req struct {
		ProductID uint `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	// 添加到浏览历史
	if err := h.userRepo.AddBrowsingHistory(userID.(uint), req.ProductID); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Browsing history saved", nil)
}

// GetBrowsingHistory 获取浏览历史
// GET /api/v1/user/browsing-history?limit=20
func (h *BrowsingHistoryHandler) GetBrowsingHistory(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	// 使用统一的limit解析
	limit := pagination.ParseLimit(c)

	// 查询浏览历史
	history, err := h.userRepo.GetBrowsingHistory(userID.(uint), limit)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"history": history,
		"count":   len(history),
	})
}

// DeleteBrowsingHistory 删除特定浏览记录
// DELETE /api/v1/user/browsing-history/:product_id
func (h *BrowsingHistoryHandler) DeleteBrowsingHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	productID, err := strconv.ParseUint(c.Param("product_id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid product ID")
		return
	}

	if err := h.userRepo.DeleteBrowsingHistory(userID.(uint), uint(productID)); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Browsing history deleted", nil)
}

// ClearBrowsingHistory 清空浏览历史
// DELETE /api/v1/user/browsing-history
func (h *BrowsingHistoryHandler) ClearBrowsingHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	if err := h.userRepo.ClearBrowsingHistory(userID.(uint)); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Browsing history cleared", nil)
}
