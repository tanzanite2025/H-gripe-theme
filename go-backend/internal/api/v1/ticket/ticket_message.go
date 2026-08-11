package ticket

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"commerce-platform/internal/domain/ticket"

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
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, ticketMessageJSONMaxBytes)
	if err := c.ShouldBindJSON(&req); err != nil {
		respondTicketJSONBindError(c, err)
		return
	}

	attachmentValues, err := h.sanitizeTicketMessageAttachments(req.Attachments, ticketMessageAttachmentMaxCount)
	if err != nil {
		respondTicketAttachmentError(c, err)
		return
	}
	attachments := marshalTicketMessageAttachments(attachmentValues...)

	msg := &ticket.TicketMessage{
		TicketID:    uint(ticketID),
		Content:     req.Content,
		MessageType: "text",
		Attachments: attachments,
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
	attachments := parseTicketMessageAttachments(item.Attachments)
	attachmentURL := firstTicketMessageAttachment(attachments)

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
		"source":          ticketMessageSource(item.Metadata),
		"attachment_url":  attachmentURL,
		"attachments":     attachments,
		"created_at":      item.CreatedAt,
		"is_read":         item.IsRead,
		"is_agent":        item.IsStaff,
	}
}

func publicCustomerServiceMessageResponse(item ticket.TicketMessage, conversationID, senderName, messageType string, metadata interface{}) gin.H {
	attachments := parseTicketMessageAttachments(item.Attachments)
	attachmentURL := firstTicketMessageAttachment(attachments)

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
		messageType = item.MessageType
	}
	messageType = normalizeTicketMessageType(messageType)
	if metadata == nil {
		metadata = parseTicketMessageMetadata(item.Metadata)
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
		"source":          ticketMessageSource(item.Metadata),
		"attachment_url":  attachmentURL,
		"attachments":     attachments,
		"created_at":      item.CreatedAt,
		"is_agent":        item.IsStaff,
	}
}

func normalizeTicketMessageType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "product", "order", "image", "link", "faq", "config_confirm":
		return value
	default:
		return "text"
	}
}

func marshalTicketMessageMetadata(value interface{}) (string, error) {
	if value == nil {
		return "", nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	if string(payload) == "null" {
		return "", nil
	}
	return string(payload), nil
}

func parseTicketMessageAttachments(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{}
	}
	var attachments []string
	if err := json.Unmarshal([]byte(value), &attachments); err != nil || attachments == nil {
		return []string{}
	}
	return attachments
}

func firstTicketMessageAttachment(attachments []string) string {
	if len(attachments) == 0 {
		return ""
	}
	return attachments[0]
}

func ticketMessageSource(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return ""
	}
	source, _ := payload["_source"].(string)
	return strings.TrimSpace(source)
}

func marshalTicketMessageAttachments(values ...string) string {
	attachments := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			attachments = append(attachments, value)
		}
	}
	if len(attachments) == 0 {
		return ""
	}
	payload, err := json.Marshal(attachments)
	if err != nil {
		return ""
	}
	return string(payload)
}

func parseTicketMessageMetadata(value string) interface{} {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	var payload interface{}
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return nil
	}
	return payload
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
