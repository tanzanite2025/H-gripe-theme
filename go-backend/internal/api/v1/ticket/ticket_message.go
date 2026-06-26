package ticket

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"tanzanite/internal/domain/ticket"

	"github.com/gin-gonic/gin"
)

// AddMessage 添加消息
// @Summary 添加消息
// @Tags Tickets
// @Accept json
// @Produce json
// @Param id path int true "工单ID"
// @Param message body map[string]interface{} true "消息内容"
// @Success 201 {object} ticket.TicketMessage
// @Router /api/v1/tickets/{id}/messages [post]
func (h *Handler) AddMessage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	isStaff := false
	if role, exists := c.Get("role"); exists && role == "admin" {
		isStaff = true
	}

	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}

	var req struct {
		Content     string   `json:"content" binding:"required"`
		Attachments []string `json:"attachments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	attachments, _ := json.Marshal(req.Attachments)

	msg := &ticket.TicketMessage{
		TicketID:    uint(ticketID),
		Content:     req.Content,
		Attachments: string(attachments),
	}

	if err := h.ticketService.AddMessage(msg, userID.(uint), isStaff); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, msg)
}

// GetMessages 获取工单消息列表
// @Summary 获取工单消息列表
// @Tags Tickets
// @Produce json
// @Param id path int true "工单ID"
// @Success 200 {array} ticket.TicketMessage
// @Router /api/v1/tickets/{id}/messages [get]
func (h *Handler) GetMessages(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	isStaff := false
	if role, exists := c.Get("role"); exists && role == "admin" {
		isStaff = true
	}

	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}

	messages, err := h.ticketService.GetMessages(uint(ticketID), userID.(uint), isStaff)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": messages})
}

// ============ 辅助函数 ============

func ticketConversationResponse(item ticket.Ticket) gin.H {
	customerName := "Customer"
	customerAvatar := ""
	if item.User != nil {
		customerName = displayName(item.User.FirstName, item.User.LastName, item.User.Username, item.User.Email)
	}
	customerID := item.UserID
	if item.Category == "customer_service" && strings.HasPrefix(item.Tags, "conversation_id:") {
		customerName = "Customer"
		customerID = 0
	}

	lastMessage := ""
	lastMessageTime := item.UpdatedAt
	unreadCount := 0
	if len(item.Messages) > 0 {
		last := item.Messages[len(item.Messages)-1]
		lastMessage = last.Content
		lastMessageTime = last.CreatedAt
		for _, message := range item.Messages {
			if !message.IsStaff && !message.IsRead {
				unreadCount++
			}
		}
	}
	conversationID := ""
	if item.ConversationID != nil && strings.TrimSpace(*item.ConversationID) != "" {
		conversationID = strings.TrimSpace(*item.ConversationID)
	} else if strings.HasPrefix(item.Tags, "conversation_id:") {
		conversationID = strings.TrimPrefix(item.Tags, "conversation_id:")
	}

	return gin.H{
		"id":                item.ID,
		"ticket_id":         item.ID,
		"conversation_id":   conversationID,
		"customer_id":       customerID,
		"customer_name":     customerName,
		"customer_avatar":   customerAvatar,
		"agent_id":          zeroToNil(item.AssignedTo),
		"status":            customerServiceStatus(item.Status),
		"unread_count":      unreadCount,
		"last_message":      lastMessage,
		"last_message_time": lastMessageTime,
		"created_at":        item.CreatedAt,
		"updated_at":        item.UpdatedAt,
	}
}

func ticketMessageResponse(item ticket.TicketMessage) gin.H {
	attachmentURL := ""
	var attachments []string
	if err := json.Unmarshal([]byte(item.Attachments), &attachments); err == nil && len(attachments) > 0 {
		attachmentURL = attachments[0]
	}

	senderName := "Customer"
	if item.IsStaff {
		senderName = "Agent"
	}
	if item.User != nil {
		senderName = displayName(item.User.FirstName, item.User.LastName, item.User.Username, item.User.Email)
	}

	return gin.H{
		"id":              item.ID,
		"conversation_id": item.TicketID,
		"sender_id":       item.UserID,
		"sender_name":     senderName,
		"message":         item.Content,
		"attachment_url":  attachmentURL,
		"created_at":      item.CreatedAt,
		"is_read":         item.IsRead,
		"is_agent":        item.IsStaff,
	}
}

func publicCustomerServiceMessageResponse(item ticket.TicketMessage, conversationID, senderName, messageType string, metadata interface{}) gin.H {
	if strings.TrimSpace(senderName) == "" {
		if item.IsStaff {
			senderName = "Agent"
		} else {
			senderName = "Customer"
		}
		if item.User != nil {
			senderName = displayName(item.User.FirstName, item.User.LastName, item.User.Username, item.User.Email)
		}
	}
	if strings.TrimSpace(messageType) == "" {
		messageType = "text"
	}

	return gin.H{
		"id":              item.ID,
		"conversation_id": conversationID,
		"sender_id":       item.UserID,
		"sender_name":     senderName,
		"sender_email":    "",
		"message":         item.Content,
		"message_type":    messageType,
		"metadata":        metadata,
		"created_at":      item.CreatedAt,
		"is_agent":        item.IsStaff,
	}
}

func displayName(firstName, lastName, username, email string) string {
	fullName := strings.TrimSpace(strings.TrimSpace(firstName) + " " + strings.TrimSpace(lastName))
	if fullName != "" {
		return fullName
	}
	if strings.TrimSpace(username) != "" {
		return strings.TrimSpace(username)
	}
	if strings.TrimSpace(email) != "" {
		return strings.TrimSpace(email)
	}
	return "Customer"
}

func customerServiceStatus(status string) string {
	switch status {
	case "resolved", "closed":
		return "closed"
	case "in_progress", "open":
		return "active"
	default:
		return "pending"
	}
}

func zeroToNil(value uint) interface{} {
	if value == 0 {
		return nil
	}
	return value
}
