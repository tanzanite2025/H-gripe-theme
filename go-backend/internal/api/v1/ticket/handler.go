package ticket

import (
	"encoding/json"
	"net/http"
	"strconv"
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
