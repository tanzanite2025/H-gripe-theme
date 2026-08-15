package ticket

import (
	"commerce-platform/internal/domain/user"
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
	groups, err := h.ticketService.ListCustomerServiceAgentGroups(limit, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
		return
	}

	items := make([]gin.H, 0, len(agents))
	for _, agent := range agents {
		if agent.UserID == nil {
			continue
		}
		onlineStatus := emptyToDefault(agent.OnlineStatus, "offline")
		items = append(items, gin.H{
			"id":            *agent.UserID,
			"user_id":       *agent.UserID,
			"agent_id":      agent.AgentID,
			"name":          agent.DisplayName(),
			"email":         agent.PublicEmail(),
			"avatar":        agent.Avatar,
			"whatsapp":      agent.WhatsApp,
			"online_status": onlineStatus,
			"status":        onlineStatus,
			"group_ids":     publicCustomerServiceAgentGroupIDs(agent.Groups),
			"groups":        publicCustomerServiceAgentGroupsResponse(agent.Groups),
			"primary_group": publicCustomerServicePrimaryAgentGroup(agent.Groups),
		})
	}
	if len(items) == 0 {
		fallbackAgents, err := h.ticketService.ListCustomerServiceAgents(limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
			return
		}
		for _, agent := range fallbackAgents {
			items = append(items, gin.H{
				"id":            agent.ID,
				"user_id":       agent.ID,
				"agent_id":      "user-" + strconv.FormatUint(uint64(agent.ID), 10),
				"name":          "Customer Service",
				"email":         "",
				"avatar":        "",
				"whatsapp":      "",
				"online_status": "offline",
				"status":        "offline",
				"group_ids":     []uint{},
				"groups":        []gin.H{},
				"primary_group": nil,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
		"groups":  publicCustomerServiceGroupsResponse(groups),
		"emailSettings": gin.H{
			"preSalesEmail":   "",
			"afterSalesEmail": "",
		},
	})
}

func publicCustomerServiceGroupsResponse(groups []user.AgentGroup) []gin.H {
	items := make([]gin.H, 0, len(groups))
	for _, group := range groups {
		items = append(items, publicCustomerServiceGroupResponse(group))
	}
	return items
}

func publicCustomerServiceAgentGroupsResponse(groups []user.AgentGroup) []gin.H {
	items := make([]gin.H, 0, len(groups))
	for _, group := range groups {
		if group.Status != "" && group.Status != "active" {
			continue
		}
		items = append(items, publicCustomerServiceGroupResponse(group))
	}
	return items
}

func publicCustomerServiceGroupResponse(group user.AgentGroup) gin.H {
	return gin.H{
		"id":          group.ID,
		"code":        group.Code,
		"name":        group.Name,
		"description": group.Description,
		"status":      group.Status,
		"sort_order":  group.SortOrder,
	}
}

func publicCustomerServiceAgentGroupIDs(groups []user.AgentGroup) []uint {
	ids := make([]uint, 0, len(groups))
	for _, group := range groups {
		if group.ID > 0 && (group.Status == "" || group.Status == "active") {
			ids = append(ids, group.ID)
		}
	}
	return ids
}

func publicCustomerServicePrimaryAgentGroup(groups []user.AgentGroup) interface{} {
	for _, group := range groups {
		if group.ID > 0 && (group.Status == "" || group.Status == "active") {
			return publicCustomerServiceGroupResponse(group)
		}
	}
	return nil
}

func emptyToDefault(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}
