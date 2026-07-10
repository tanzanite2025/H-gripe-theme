package admin

import (
	"strconv"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"
	"time"

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

func (h *MarketingHandler) CreateGiftCard(c *gin.Context) {
	var req struct {
		Code           string     `json:"code" binding:"required"`
		InitialValue   float64    `json:"initial_value" binding:"required,gt=0"`
		Currency       string     `json:"currency"`
		RecipientEmail string     `json:"recipient_email"`
		RecipientName  string     `json:"recipient_name"`
		SenderName     string     `json:"sender_name"`
		Message        string     `json:"message"`
		CoverImage     string     `json:"cover_image"`
		ExpiresAt      *time.Time `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	giftCard, err := h.marketingService.CreateGiftCardAdmin(service.GiftCardCreateInput{
		Code:           req.Code,
		InitialValue:   req.InitialValue,
		Currency:       req.Currency,
		RecipientEmail: req.RecipientEmail,
		RecipientName:  req.RecipientName,
		SenderName:     req.SenderName,
		Message:        req.Message,
		CoverImage:     req.CoverImage,
		ExpiresAt:      req.ExpiresAt,
	})
	if err != nil {
		respondMarketingError(c, err, "gift card")
		return
	}

	response.Created(c, gin.H{"gift_card": giftCard})
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
