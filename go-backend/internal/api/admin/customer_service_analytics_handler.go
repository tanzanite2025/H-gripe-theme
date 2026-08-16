package admin

import (
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *TicketHandler) GetCustomerServiceAnalytics(c *gin.Context) {
	if h.customerServiceAnalytics == nil {
		apierror.RespondInternalError(c, errors.New("customer service analytics service is not configured"))
		return
	}

	timezoneOffsetMinutes, err := strconv.Atoi(c.DefaultQuery("tz_offset_minutes", "0"))
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid timezone offset")
		return
	}

	agentUserID, canViewAll := adminCustomerServiceScope(c)
	analytics, err := h.customerServiceAnalytics.ForAgent(service.CustomerServiceAnalyticsInput{
		Date:                  strings.TrimSpace(c.Query("date")),
		TimezoneOffsetMinutes: timezoneOffsetMinutes,
		AgentUserID:           agentUserID,
		CanViewAll:            canViewAll,
	})
	if err != nil {
		respondAdminCustomerServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"analytics": analytics})
}
