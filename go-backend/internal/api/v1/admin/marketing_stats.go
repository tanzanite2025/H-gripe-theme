package admin

import (
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *MarketingHandler) GetMarketingStats(c *gin.Context) {
	stats, err := h.marketingService.GetMarketingStats()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, stats)
}
