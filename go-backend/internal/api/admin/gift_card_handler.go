package admin

import (
	"strconv"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *MarketingHandler) ListGiftCards(c *gin.Context) {
	params := pagination.ParsePagination(c)
	status := c.Query("status")

	giftCards, total, err := h.marketingService.ListGiftCardsAdmin(params.Page, params.PageSize, status)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Paged(c, gin.H{"gift_cards": giftCards}, params.Page, params.PageSize, total)
}

func (h *MarketingHandler) GetGiftCard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid gift card ID")
		return
	}

	detail, err := h.marketingService.GetGiftCard(uint(id))
	if err != nil {
		respondMarketingError(c, err, "gift card")
		return
	}

	response.Success(c, gin.H{
		"gift_card":    detail.GiftCard,
		"transactions": detail.Transactions,
	})
}

func (h *MarketingHandler) UpdateGiftCardStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid gift card ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active used expired cancelled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	giftCard, err := h.marketingService.UpdateGiftCardStatus(uint(id), req.Status)
	if err != nil {
		respondMarketingError(c, err, "gift card")
		return
	}

	response.Success(c, gin.H{"gift_card": giftCard})
}
