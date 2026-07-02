package auth

import (
	"strconv"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type BrowsingHistoryHandler struct {
	userService *service.UserService
}

func NewBrowsingHistoryHandler(userService *service.UserService) *BrowsingHistoryHandler {
	return &BrowsingHistoryHandler{userService: userService}
}

func (h *BrowsingHistoryHandler) AddBrowsingHistory(c *gin.Context) {
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

	if err := h.userService.AddBrowsingHistory(userID.(uint), req.ProductID); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Browsing history saved", nil)
}

func (h *BrowsingHistoryHandler) GetBrowsingHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	limit := pagination.ParseLimit(c)
	history, err := h.userService.GetBrowsingHistory(userID.(uint), limit)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"history": history,
		"count":   len(history),
	})
}

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

	if err := h.userService.DeleteBrowsingHistory(userID.(uint), uint(productID)); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Browsing history deleted", nil)
}

func (h *BrowsingHistoryHandler) ClearBrowsingHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	if err := h.userService.ClearBrowsingHistory(userID.(uint)); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Browsing history cleared", nil)
}
