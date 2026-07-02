package audit

import (
	"errors"
	"net/http"
	"strconv"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	auditService *service.AuditService
}

func NewHandler(auditService *service.AuditService) *Handler {
	return &Handler{auditService: auditService}
}

func (h *Handler) GetAuditLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid log id"})
		return
	}

	log, err := h.auditService.GetAuditLog(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "audit log not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, log)
}

func (h *Handler) ListUserAuditLogs(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	page, pageSize := auditPagination(c)
	logs, total, err := h.auditService.GetUserAuditLogs(uint(userID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAuditPage(c, logs, total, page, pageSize)
}

func (h *Handler) ListEntityAuditLogs(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")
	if entityType == "" || entityIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_type and entity_id are required"})
		return
	}

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity id"})
		return
	}

	page, pageSize := auditPagination(c)
	logs, total, err := h.auditService.GetEntityAuditLogs(entityType, uint(entityID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAuditPage(c, logs, total, page, pageSize)
}

func (h *Handler) ListAllAuditLogs(c *gin.Context) {
	page, pageSize := auditPagination(c)
	logs, total, err := h.auditService.GetAllAuditLogs(page, pageSize, c.Query("action"), c.Query("entity_type"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAuditPage(c, logs, total, page, pageSize)
}

func (h *Handler) ListAuditLogsByDateRange(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	page, pageSize := auditPagination(c)
	logs, total, err := h.auditService.GetAuditLogsByDateRange(startDate, endDate, page, pageSize)
	if err != nil {
		if errors.Is(err, service.ErrInvalidAuditDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format (use YYYY-MM-DD)"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAuditPage(c, logs, total, page, pageSize)
}

func (h *Handler) ListAuditLogsByIP(c *gin.Context) {
	ipAddress := c.Param("ip_address")
	if ipAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip_address is required"})
		return
	}

	page, pageSize := auditPagination(c)
	logs, total, err := h.auditService.GetAuditLogsByIP(ipAddress, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAuditPage(c, logs, total, page, pageSize)
}

func (h *Handler) SearchAuditLogs(c *gin.Context) {
	page, pageSize := auditPagination(c)
	logs, total, err := h.auditService.SearchAuditLogs(c.Query("keyword"), page, pageSize)
	if err != nil {
		if errors.Is(err, service.ErrAuditKeywordRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "keyword is required"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAuditPage(c, logs, total, page, pageSize)
}

func (h *Handler) GetAuditStats(c *gin.Context) {
	stats, err := h.auditService.GetAuditStats(c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		if errors.Is(err, service.ErrInvalidAuditDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format (use YYYY-MM-DD)"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) GetRecentActivities(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 200 {
		limit = 50
	}

	logs, err := h.auditService.GetRecentActivities(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func (h *Handler) DeleteOldAuditLogs(c *gin.Context) {
	var req struct {
		BeforeDate string `json:"before_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.auditService.DeleteOldLogs(req.BeforeDate); err != nil {
		if errors.Is(err, service.ErrInvalidAuditDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format (use YYYY-MM-DD)"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "old audit logs deleted"})
}

func auditPagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func writeAuditPage(c *gin.Context, logs interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}
