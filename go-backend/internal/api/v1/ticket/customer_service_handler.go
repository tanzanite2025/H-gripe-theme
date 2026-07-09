package ticket

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListPublicCustomerServiceAgents(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	agents, err := h.ticketService.ListCustomerServiceAgentProfiles(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
		return
	}
	if agents == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "[CRITICAL] ListCustomerServiceAgentProfiles returned nil"})
		return
	}

	items := make([]gin.H, 0, len(agents))
	for _, agent := range agents {
		if agent.UserID == nil {
			continue
		}
		items = append(items, gin.H{
			"id":       *agent.UserID,
			"user_id":  *agent.UserID,
			"agent_id": agent.AgentID,
			"name":     agent.DisplayName(),
			"email":    agent.PublicEmail(),
			"avatar":   agent.Avatar,
			"whatsapp": agent.WhatsApp,
			"status":   emptyToDefault(agent.OnlineStatus, "offline"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
		"emailSettings": gin.H{
			"preSalesEmail":   "",
			"afterSalesEmail": "",
		},
	})
}

func emptyToDefault(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}
