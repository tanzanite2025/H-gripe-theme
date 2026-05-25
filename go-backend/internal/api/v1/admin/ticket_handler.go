package admin

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/ticket"
	"tanzanite/internal/repository"

	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	ticketRepo *repository.TicketRepository
}

func NewTicketHandler(ticketRepo *repository.TicketRepository) *TicketHandler {
	return &TicketHandler{
		ticketRepo: ticketRepo,
	}
}

// ListTickets 获取工单列表
// GET /api/admin/tickets
func (h *TicketHandler) ListTickets(c *gin.Context) {
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

	tickets, total, err := h.ticketRepo.FindAllTickets(page, pageSize, status, priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tickets"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"tickets": tickets,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetTicket 获取工单详情
// GET /api/admin/tickets/:id
func (h *TicketHandler) GetTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	ticketItem, err := h.ticketRepo.FindTicketByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ticket": ticketItem,
	})
}

// UpdateTicketStatus 更新工单状态
// PATCH /api/admin/tickets/:id/status
func (h *TicketHandler) UpdateTicketStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=open in_progress resolved closed"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ticketRepo.UpdateTicketStatus(uint(id), req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ticket status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ticket status updated successfully",
	})
}

// AssignTicket 分配工单
// PATCH /api/admin/tickets/:id/assign
func (h *TicketHandler) AssignTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	var req struct {
		AssignedTo uint `json:"assigned_to" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ticketRepo.AssignTicket(uint(id), req.AssignedTo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ticket assigned successfully",
	})
}

// UpdateTicket 更新工单
// PUT /api/admin/tickets/:id
func (h *TicketHandler) UpdateTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	var req struct {
		Subject    string `json:"subject"`
		Priority   string `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
		Status     string `json:"status" binding:"omitempty,oneof=open in_progress resolved closed"`
		AssignedTo *uint  `json:"assigned_to"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingTicket, err := h.ticketRepo.FindTicketByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	if req.Subject != "" {
		existingTicket.Subject = req.Subject
	}
	if req.Priority != "" {
		existingTicket.Priority = req.Priority
	}
	if req.Status != "" {
		existingTicket.Status = req.Status
	}
	if req.AssignedTo != nil {
		existingTicket.AssignedTo = *req.AssignedTo
	}

	if err := h.ticketRepo.UpdateTicket(existingTicket); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ticket updated successfully",
		"ticket":  existingTicket,
	})
}

// DeleteTicket 删除工单
// DELETE /api/admin/tickets/:id
func (h *TicketHandler) DeleteTicket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	if err := h.ticketRepo.DeleteTicket(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ticket deleted successfully",
	})
}

// GetTicketStats 获取工单统计
// GET /api/admin/tickets/stats
func (h *TicketHandler) GetTicketStats(c *gin.Context) {
	stats, err := h.ticketRepo.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ticket stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CreateMessage 创建工单消息
// POST /api/admin/tickets/:id/messages
func (h *TicketHandler) CreateMessage(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	var req struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	newMessage := &ticket.TicketMessage{
		TicketID: uint(ticketID),
		UserID:   userID.(uint),
		Content:  req.Message,
		IsStaff:  true, // 管理员发送的消息
	}

	if err := h.ticketRepo.CreateTicketMessage(newMessage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Message created successfully",
		"data":    newMessage,
	})
}

// GetMessages 获取工单消息列表
// GET /api/admin/tickets/:id/messages
func (h *TicketHandler) GetMessages(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	messages, err := h.ticketRepo.FindMessagesByTicketID(uint(ticketID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
	})
}

// MarkMessagesAsRead 标记消息为已读
// POST /api/admin/tickets/:id/messages/mark-read
func (h *TicketHandler) MarkMessagesAsRead(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	if err := h.ticketRepo.MarkMessagesAsRead(uint(ticketID), true); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark messages as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Messages marked as read",
	})
}
