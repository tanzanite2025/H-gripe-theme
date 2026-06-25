package ticket

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"tanzanite/internal/domain/ticket"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	ticketService *service.TicketService
}

func NewHandler(ticketService *service.TicketService) *Handler {
	return &Handler{
		ticketService: ticketService,
	}
}

// CreateTicketRequest 鍒涘缓宸ュ崟璇锋眰
type CreateTicketRequest struct {
	Subject  string `json:"subject" binding:"required"`
	Category string `json:"category" binding:"required"`
	Priority string `json:"priority"`
	Content  string `json:"content" binding:"required"`
}

// CreateTicket 鍒涘缓宸ュ崟
// @Summary 鍒涘缓宸ュ崟
// @Tags Tickets
// @Accept json
// @Produce json
// @Param ticket body CreateTicketRequest true "宸ュ崟淇℃伅"
// @Success 201 {object} ticket.Ticket
// @Router /api/v1/tickets [post]
func (h *Handler) CreateTicket(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t := &ticket.Ticket{
		UserID:   userID.(uint),
		Subject:  req.Subject,
		Category: req.Category,
		Priority: req.Priority,
	}

	if err := h.ticketService.CreateTicket(t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 鍒涘缓鍒濆娑堟伅
	msg := &ticket.TicketMessage{
		TicketID: t.ID,
		UserID:   userID.(uint),
		IsStaff:  false,
		Content:  req.Content,
	}
	if err := h.ticketService.AddMessage(msg, userID.(uint), false); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, t)
}

// GetTicket 鑾峰彇宸ュ崟璇︽儏
// @Summary 鑾峰彇宸ュ崟璇︽儏
// @Tags Tickets
// @Produce json
// @Param id path int true "宸ュ崟ID"
// @Success 200 {object} ticket.Ticket
// @Router /api/v1/tickets/{id} [get]
func (h *Handler) GetTicket(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 妫€鏌ユ槸鍚︽槸绠＄悊鍛?
	isStaff := false
	if role, exists := c.Get("role"); exists && role == "admin" {
		isStaff = true
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}

	t, err := h.ticketService.GetTicket(uint(id), userID.(uint), isStaff)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, t)
}

// ListTickets 鑾峰彇宸ュ崟鍒楄〃
// @Summary 鑾峰彇宸ュ崟鍒楄〃
// @Tags Tickets
// @Produce json
// @Param page query int false "椤电爜" default(1)
// @Param page_size query int false "姣忛〉鏁伴噺" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tickets [get]
func (h *Handler) ListTickets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tickets, total, err := h.ticketService.GetUserTickets(userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":   true,
		"data": tickets,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// ListAllTickets 鑾峰彇鎵€鏈夊伐鍗曪紙绠＄悊鍛橈級
// @Summary 鑾峰彇鎵€鏈夊伐鍗?
// @Tags Tickets
// @Produce json
// @Param page query int false "椤电爜" default(1)
// @Param page_size query int false "姣忛〉鏁伴噺" default(20)
// @Param status query string false "鐘舵€?
// @Param priority query string false "浼樺厛绾?
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/tickets [get]
func (h *Handler) ListAllTickets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	priority := c.Query("priority")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tickets, total, err := h.ticketService.GetAllTickets(page, pageSize, status, priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":   true,
		"data": tickets,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// UpdateTicketStatus 鏇存柊宸ュ崟鐘舵€?
// @Summary 鏇存柊宸ュ崟鐘舵€?
// @Tags Tickets
// @Accept json
// @Produce json
// @Param id path int true "宸ュ崟ID"
// @Param request body map[string]string true "鐘舵€?
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tickets/{id}/status [put]
func (h *Handler) UpdateTicketStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ticketService.UpdateTicketStatus(uint(id), req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ticket status updated"})
}

// AssignTicket 鍒嗛厤宸ュ崟锛堢鐞嗗憳锛?
// @Summary 鍒嗛厤宸ュ崟
// @Tags Tickets
// @Accept json
// @Produce json
// @Param id path int true "宸ュ崟ID"
// @Param request body map[string]uint true "鍒嗛厤淇℃伅"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/tickets/{id}/assign [post]
func (h *Handler) AssignTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}

	var req struct {
		AssignedTo uint `json:"assigned_to" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ticketService.AssignTicket(uint(id), req.AssignedTo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ticket assigned"})
}

// CloseTicket 鍏抽棴宸ュ崟
// @Summary 鍏抽棴宸ュ崟
// @Tags Tickets
// @Produce json
// @Param id path int true "宸ュ崟ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tickets/{id}/close [post]
func (h *Handler) CloseTicket(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	isStaff := false
	if role, exists := c.Get("role"); exists && role == "admin" {
		isStaff = true
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}

	if err := h.ticketService.CloseTicket(uint(id), userID.(uint), isStaff); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ticket closed"})
}

// AddMessage 娣诲姞娑堟伅
// @Summary 娣诲姞娑堟伅
// @Tags Tickets
// @Accept json
// @Produce json
// @Param id path int true "宸ュ崟ID"
// @Param message body map[string]interface{} true "娑堟伅鍐呭"
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

// GetMessages 鑾峰彇宸ュ崟娑堟伅鍒楄〃
// @Summary 鑾峰彇宸ュ崟娑堟伅鍒楄〃
// @Tags Tickets
// @Produce json
// @Param id path int true "宸ュ崟ID"
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

// GetTicketStats 鑾峰彇宸ュ崟缁熻
// @Summary 鑾峰彇宸ュ崟缁熻
// @Tags Tickets
// @Produce json
// @Success 200 {object} map[string]int64
// @Router /api/v1/tickets/stats [get]
func (h *Handler) GetTicketStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	stats, err := h.ticketService.GetTicketStats(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetDashboard 鑾峰彇瀹㈡湇浠〃鏉匡紙绠＄悊鍛橈級
// @Summary 鑾峰彇瀹㈡湇浠〃鏉?
// @Tags Tickets
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/tickets/dashboard [get]
func (h *Handler) GetDashboard(c *gin.Context) {
	dashboard, err := h.ticketService.GetDashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// GetRecentTickets 鑾峰彇鏈€杩戠殑宸ュ崟锛堢鐞嗗憳锛?
// @Summary 鑾峰彇鏈€杩戠殑宸ュ崟
// @Tags Tickets
// @Produce json
// @Param limit query int false "鏁伴噺闄愬埗" default(10)
// @Success 200 {array} ticket.Ticket
// @Router /api/v1/admin/tickets/recent [get]
func (h *Handler) GetRecentTickets(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	tickets, err := h.ticketService.GetRecentTickets(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tickets})
}

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

	return gin.H{
		"id":                item.ID,
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
