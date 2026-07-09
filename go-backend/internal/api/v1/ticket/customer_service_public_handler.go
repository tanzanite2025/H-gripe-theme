package ticket

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) EnsurePublicCustomerServiceConversation(c *gin.Context) {
	var req struct {
		AgentID string `json:"agent_id"`
	}
	if c.Request != nil && c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
			return
		}
	}

	t, err := h.ticketService.GetOrCreatePublicCustomerServiceConversation(
		h.publicCustomerOwner(c),
		parseCustomerServiceAgentID(req.AgentID),
	)
	if err != nil {
		writePublicCustomerServiceError(c, err)
		return
	}

	conversationID := publicConversationID(t)
	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"hasConversation": true,
		"conversation_id": conversationID,
		"lastAgentId":     zeroToNil(t.AssignedTo),
		"data": gin.H{
			"conversation_id": conversationID,
			"lastAgentId":     zeroToNil(t.AssignedTo),
		},
	})
}

func (h *Handler) HasPublicCustomerServiceConversation(c *gin.Context) {
	hasConversation, conversationID, lastAgentID, err := h.ticketService.HasPublicCustomerServiceConversation(h.publicCustomerOwner(c))
	if err != nil {
		writePublicCustomerServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hasConversation": hasConversation,
		"conversation_id": conversationID,
		"lastAgentId":     zeroToNil(lastAgentID),
	})
}

func (h *Handler) SendPublicCustomerServiceMessage(c *gin.Context) {
	var req struct {
		ConversationID string      `json:"conversation_id"`
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
	if message == "" || senderName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] missing required parameters"})
		return
	}

	t, msg, err := h.ticketService.AddPublicCustomerServiceMessage(
		conversationID,
		h.publicCustomerOwner(c),
		message,
		parseCustomerServiceAgentID(req.AgentID),
	)
	if err != nil {
		writePublicCustomerServiceError(c, err)
		return
	}
	if msg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "[CRITICAL] AddPublicCustomerServiceMessage returned nil message"})
		return
	}

	conversationID = publicConversationID(t)
	response := publicCustomerServiceMessageResponse(*msg, conversationID, senderName, req.MessageType, req.Metadata)
	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message_id":      msg.ID,
		"conversation_id": conversationID,
		"data":            response,
	})
}

func (h *Handler) GetWelcomeMessage(c *gin.Context) {
	conversationID := strings.TrimSpace(c.Query("conversation_id"))
	agentID := parseCustomerServiceAgentID(c.Query("agent_id"))
	owner := h.publicCustomerOwner(c)

	if conversationID == "" {
		t, err := h.ticketService.GetOrCreatePublicCustomerServiceConversation(owner, agentID)
		if err != nil {
			writePublicCustomerServiceError(c, err)
			return
		}
		conversationID = publicConversationID(t)
	}

	reply, alreadySent, err := h.ticketService.GetWelcomeMessage(conversationID, owner, agentID)
	if err != nil {
		writePublicCustomerServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"conversation_id": conversationID,
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
		AgentID        string `json:"agent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
		return
	}
	req.ConversationID = strings.TrimSpace(req.ConversationID)
	if req.ConversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "[CRITICAL] conversation_id is required"})
		return
	}

	reply, ruleID, err := h.ticketService.MatchKeywordMessage(
		req.ConversationID,
		req.Message,
		h.publicCustomerOwner(c),
		parseCustomerServiceAgentID(req.AgentID),
	)
	if err != nil {
		writePublicCustomerServiceError(c, err)
		return
	}

	if reply == "" {
		c.JSON(http.StatusOK, gin.H{
			"success":         true,
			"conversation_id": req.ConversationID,
			"data": gin.H{
				"reply": "",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"conversation_id": req.ConversationID,
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

	messages, err := h.ticketService.GetPublicCustomerServiceMessages(conversationID, h.publicCustomerOwner(c), limit, offset)
	if err != nil {
		writePublicCustomerServiceError(c, err)
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
