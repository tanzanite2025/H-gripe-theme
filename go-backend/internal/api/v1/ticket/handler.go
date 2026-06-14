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

// CreateTicketRequest 创建工单请求
type CreateTicketRequest struct {
	Subject  string `json:"subject" binding:"required"`
	Category string `json:"category" binding:"required"`
	Priority string `json:"priority"`
	Content  string `json:"content" binding:"required"`
}

// CreateTicket 创建工单
// @Summary 创建工单
// @Tags Tickets
// @Accept json
// @Produce json
// @Param ticket body CreateTicketRequest true "工单信息"
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

	// 创建初始消息
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

// GetTicket 获取工单详情
// @Summary 获取工单详情
// @Tags Tickets
// @Produce json
// @Param id path int true "工单ID"
// @Success 200 {object} ticket.Ticket
// @Router /api/v1/tickets/{id} [get]
func (h *Handler) GetTicket(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 检查是否是管理员
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

// ListTickets 获取工单列表
// @Summary 获取工单列表
// @Tags Tickets
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
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

// ListAllTickets 获取所有工单（管理员）
// @Summary 获取所有工单
// @Tags Tickets
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "状态"
// @Param priority query string false "优先级"
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

// UpdateTicketStatus 更新工单状态
// @Summary 更新工单状态
// @Tags Tickets
// @Accept json
// @Produce json
// @Param id path int true "工单ID"
// @Param request body map[string]string true "状态"
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

// AssignTicket 分配工单（管理员）
// @Summary 分配工单
// @Tags Tickets
// @Accept json
// @Produce json
// @Param id path int true "工单ID"
// @Param request body map[string]uint true "分配信息"
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

// CloseTicket 关闭工单
// @Summary 关闭工单
// @Tags Tickets
// @Produce json
// @Param id path int true "工单ID"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "customer_service_error", "message": err.Error()})
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
	agents, err := h.ticketService.ListCustomerServiceAgents(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	items := make([]gin.H, 0, len(agents))
	for _, agent := range agents {
		items = append(items, gin.H{
			"id":         agent.ID,
			"wp_user_id": 0,
			"name":       displayName(agent.FirstName, agent.LastName, agent.Username, agent.Email),
			"email":      agent.Email,
			"avatar":     "",
			"whatsapp":   "",
			"status":     "online",
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

func (h *Handler) HasPublicCustomerServiceConversation(c *gin.Context) {
	conversationID := strings.TrimSpace(c.Query("visitor_id"))
	if conversationID == "" {
		c.JSON(http.StatusOK, gin.H{"hasConversation": false})
		return
	}

	hasConversation, lastAgentID, err := h.ticketService.HasPublicCustomerServiceConversation(conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	conversationID := strings.TrimSpace(req.ConversationID)
	message := strings.TrimSpace(req.Message)
	senderName := strings.TrimSpace(req.SenderName)
	if conversationID == "" || message == "" || senderName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "missing required parameters"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	response := publicCustomerServiceMessageResponse(*msg, conversationID, senderName, req.MessageType, req.Metadata)
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message_id": msg.ID,
		"data":       response,
	})
}

func (h *Handler) GetPublicCustomerServiceMessages(c *gin.Context) {
	conversationID := strings.TrimSpace(c.Param("conversation_id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "missing conversation id"})
		return
	}

	messages, err := h.ticketService.GetPublicCustomerServiceMessages(conversationID, limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": []gin.H{}, "total": 0})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	messages, err := h.ticketService.GetMessages(uint(ticketID), 0, true)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	items := make([]gin.H, 0, len(messages))
	for _, item := range messages {
		items = append(items, ticketMessageResponse(item))
	}

	if err := h.ticketService.MarkMessagesAsRead(uint(ticketID), true); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ConversationID uint   `json:"conversation_id" binding:"required"`
		Message        string `json:"message" binding:"required"`
		AttachmentURL  string `json:"attachment_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"ok": true, "message": ticketMessageResponse(*msg)})
}

func (h *Handler) MarkCustomerServiceMessagesRead(c *gin.Context) {
	var req struct {
		ConversationID uint `json:"conversation_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ticketService.MarkMessagesAsRead(req.ConversationID, true); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "success": true})
}

func (h *Handler) TransferCustomerServiceConversation(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid conversation id"})
		return
	}

	var req struct {
		ToAgentID string `json:"to_agent_id" binding:"required"`
		Note      string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	assignedTo, err := strconv.ParseUint(req.ToAgentID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid agent id"})
		return
	}

	if err := h.ticketService.AssignTicket(uint(ticketID), uint(assignedTo)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": err.Error()})
		return
	}

	allowed := map[string]bool{"online": true, "busy": true, "away": true, "offline": true}
	if !allowed[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": "invalid status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
		"data": gin.H{
			"status": req.Status,
		},
	})
}

// GetTicketStats 获取工单统计
// @Summary 获取工单统计
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

// GetDashboard 获取客服仪表板（管理员）
// @Summary 获取客服仪表板
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

// GetRecentTickets 获取最近的工单（管理员）
// @Summary 获取最近的工单
// @Tags Tickets
// @Produce json
// @Param limit query int false "数量限制" default(10)
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
