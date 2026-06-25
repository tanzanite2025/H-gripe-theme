package ticket

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"tanzanite/internal/domain/ticket"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListCustomerServiceConversations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	tickets, total, err := h.ticketService.GetCustomerServiceConversations(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "customer_service_error", "message": "[CRITICAL] " + err.Error()})
		return
	}
	if tickets == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "customer_service_error", "message": "[CRITICAL] GetCustomerServiceConversations returned nil"})
		return
	}

	items := make([]gin.H, 0, len(tickets))
	for _, item := range tickets {
		items = append(items, ticketConversationResponse(item))
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"items": items,
			"pagination": gin.H{
				"page":       page,
				"page_size":  pageSize,
				"total":      total,
				"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		},
		"conversations": items,
	})
}

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
			"id":              *agent.UserID,
			"agent_id":        agent.AgentID,
			"legacy_agent_id": agent.AgentID,
			"wp_user_id":      *agent.UserID,
			"name":            agent.DisplayName(),
			"email":           agent.PublicEmail(),
			"avatar":          agent.Avatar,
			"whatsapp":        agent.WhatsApp,
			"status":          emptyToDefault(agent.OnlineStatus, "offline"),
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

func (h *Handler) HasPublicCustomerServiceConversation(c *gin.Context) {
	conversationID := strings.TrimSpace(c.Query("visitor_id"))
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] visitor_id is required"})
		return
	}

	hasConversation, lastAgentID, err := h.ticketService.HasPublicCustomerServiceConversation(conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "[CRITICAL] " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hasConversation": hasConversation,
		"lastAgentId":     zeroToNil(lastAgentID),
	})
}

func (h *Handler) SendPublicCustomerServiceMessage(c *gin.Context) {
	var req struct {
		ConversationID string      `json:"conversation_id" binding:"required"`
		Message        string      `json:"message" binding:"required"`
		SenderType     string      `json:"sender_type"`
		SenderName     string      `json:"sender_name" binding:"required"`
		SenderEmail    string      `json:"sender_email"`
		AgentID        string      `json:"agent_id"`
		MessageType    string      `json:"message_type"`
		Metadata       interface{} `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
		return
	}

	conversationID := strings.TrimSpace(req.ConversationID)
	message := strings.TrimSpace(req.Message)
	senderName := strings.TrimSpace(req.SenderName)
	if conversationID == "" || message == "" || senderName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] missing required parameters"})
		return
	}

	var userID uint
	if value, exists := c.Get("user_id"); exists {
		if parsed, ok := value.(uint); ok {
			userID = parsed
		}
	}

	var agentID uint
	if strings.TrimSpace(req.AgentID) != "" {
		if parsed, err := strconv.ParseUint(strings.TrimSpace(req.AgentID), 10, 32); err == nil {
			agentID = uint(parsed)
		}
	}

	_, msg, err := h.ticketService.AddPublicCustomerServiceMessage(conversationID, message, userID, agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
		return
	}
	if msg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "[CRITICAL] AddPublicCustomerServiceMessage returned nil message"})
		return
	}

	response := publicCustomerServiceMessageResponse(*msg, conversationID, senderName, req.MessageType, req.Metadata)
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message_id": msg.ID,
		"data":       response,
	})
}

func (h *Handler) GetWelcomeMessage(c *gin.Context) {
	conversationID := strings.TrimSpace(c.Query("conversation_id"))
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] conversation_id is required"})
		return
	}

	reply, alreadySent, err := h.ticketService.GetWelcomeMessage(conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message":      reply,
			"already_sent": alreadySent,
		},
	})
}

func (h *Handler) MatchKeywordMessage(c *gin.Context) {
	var req struct {
		Message        string `json:"message" binding:"required"`
		ConversationID string `json:"conversation_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
		return
	}

	reply, ruleID, err := h.ticketService.MatchKeywordMessage(req.ConversationID, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
		return
	}

	if reply == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"reply": "",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"reply":   reply,
			"rule_id": ruleID,
		},
	})
}

func (h *Handler) GetPublicCustomerServiceMessages(c *gin.Context) {
	conversationID := strings.TrimSpace(c.Param("conversation_id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] missing conversation id"})
		return
	}

	messages, err := h.ticketService.GetPublicCustomerServiceMessages(conversationID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
		return
	}
	if messages == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "[CRITICAL] GetPublicCustomerServiceMessages returned nil"})
		return
	}

	items := make([]gin.H, 0, len(messages))
	for _, item := range messages {
		items = append(items, publicCustomerServiceMessageResponse(item, conversationID, "", "", nil))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
		"total":   len(items),
	})
}

func (h *Handler) GetCustomerServiceMessages(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "[CRITICAL] invalid conversation id"})
		return
	}

	messages, err := h.ticketService.GetMessages(uint(ticketID), 0, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "[CRITICAL] " + err.Error()})
		return
	}
	if messages == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "[CRITICAL] GetMessages returned nil"})
		return
	}

	items := make([]gin.H, 0, len(messages))
	for _, item := range messages {
		items = append(items, ticketMessageResponse(item))
	}

	if err := h.ticketService.MarkMessagesAsRead(uint(ticketID), true); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "[CRITICAL] " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     gin.H{"items": items},
		"messages": items,
	})
}

func (h *Handler) SendCustomerServiceMessage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "[CRITICAL] unauthorized"})
		return
	}

	var req struct {
		ConversationID uint   `json:"conversation_id" binding:"required"`
		Message        string `json:"message" binding:"required"`
		AttachmentURL  string `json:"attachment_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "[CRITICAL] " + err.Error()})
		return
	}

	attachments := []string{}
	if strings.TrimSpace(req.AttachmentURL) != "" {
		attachments = append(attachments, strings.TrimSpace(req.AttachmentURL))
	}
	attachmentsJSON, _ := json.Marshal(attachments)

	msg := &ticket.TicketMessage{
		TicketID:    req.ConversationID,
		UserID:      userID.(uint),
		IsStaff:     true,
		Content:     strings.TrimSpace(req.Message),
		Attachments: string(attachmentsJSON),
		IsRead:      false,
		IsInternal:  false,
	}
	if err := h.ticketService.AddMessage(msg, userID.(uint), true); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "[CRITICAL] " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"ok": true, "message": ticketMessageResponse(*msg)})
}

func (h *Handler) MarkCustomerServiceMessagesRead(c *gin.Context) {
	var req struct {
		ConversationID uint `json:"conversation_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "[CRITICAL] " + err.Error()})
		return
	}

	if err := h.ticketService.MarkMessagesAsRead(req.ConversationID, true); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "[CRITICAL] " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "success": true})
}

func (h *Handler) TransferCustomerServiceConversation(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] invalid conversation id"})
		return
	}

	var req struct {
		ToAgentID string `json:"to_agent_id" binding:"required"`
		Note      string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
		return
	}

	assignedTo, err := strconv.ParseUint(req.ToAgentID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] invalid agent id"})
		return
	}

	if err := h.ticketService.AssignTicket(uint(ticketID), uint(assignedTo)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"success": true,
		"data": gin.H{
			"conversation_id": ticketID,
			"to_agent_id":     assignedTo,
			"to_agent":        req.ToAgentID,
		},
	})
}

func (h *Handler) GetCustomerServiceAgentStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ok": true,
		"data": gin.H{
			"status": "online",
		},
	})
}

func (h *Handler) UpdateCustomerServiceAgentStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": "[CRITICAL] " + err.Error()})
		return
	}

	allowed := map[string]bool{"online": true, "busy": true, "away": true, "offline": true}
	if !allowed[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": "[CRITICAL] invalid status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
		"data": gin.H{
			"status": req.Status,
		},
	})
}
