package admin

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/repository"
)

func (h *DashboardHandler) GetCommandCenterOutboxFailures(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, h.businessFollowUpReadService.GetCommandCenterOutboxFailures(c.Request.Context(), auth.Role(c.GetString("user_role")), c.GetUint("user_id")))
}

func (h *DashboardHandler) GetBusinessFollowUpList(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, h.businessFollowUpReadService.GetDashboardBusinessFollowUpList(c.Request.Context(), auth.Role(c.GetString("user_role")), c.GetUint("user_id")))
}

func (h *DashboardHandler) GetBusinessFollowUpDetail(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || eventID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid follow-up ID"})
		return
	}
	item, err := h.businessFollowUpReadService.GetDashboardBusinessFollowUpDetail(c.Request.Context(), auth.Role(c.GetString("user_role")), c.GetUint("user_id"), uint(eventID))
	if errors.Is(err, repository.ErrOutboxEventNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Follow-up not found or unavailable to this account"})
		return
	}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Business follow-up data is unavailable"})
		return
	}
	c.JSON(http.StatusOK, item)
}
