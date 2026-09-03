package admin

import (
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *MarketingHandler) GetPromotionRiskAnalysis(c *gin.Context) {
	analysis, err := h.marketingService.AnalyzePromotionStackingRisk()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"analysis": analysis})
}
