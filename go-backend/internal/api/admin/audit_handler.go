package admin

import (
	"errors"
	"net/http"
	"strconv"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	auditService *service.AuditService
}

func NewAuditHandler(auditService *service.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	page, pageSize := auditPagination(c)
	userID, err := parseOptionalAuditUserID(c.Query("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	logs, total, err := h.auditService.ListAuditLogs(service.AuditListInput{
		Page:      page,
		PageSize:  pageSize,
		Action:    c.Query("action"),
		Resource:  c.Query("resource"),
		UserID:    userID,
		IPAddress: c.Query("ip_address"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidAuditDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch audit logs"})
		return
	}

	writeAuditList(c, logs, total, page, pageSize)
}

func (h *AuditHandler) GetAuditLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid audit log ID"})
		return
	}

	log, err := h.auditService.GetAuditLog(uint(id))
	if err != nil {
		if service.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "audit log not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch audit log"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"log": log})
}

func (h *AuditHandler) GetAuditStats(c *gin.Context) {
	stats, err := h.auditService.GetAuditStats(c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		if errors.Is(err, service.ErrInvalidAuditDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch audit stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *AuditHandler) GetRecentActivities(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	logs, err := h.auditService.GetRecentActivities(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recent activities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"activities": logs})
}

func (h *AuditHandler) SearchAuditLogs(c *gin.Context) {
	page, pageSize := auditPagination(c)

	logs, total, err := h.auditService.SearchAuditLogs(c.Query("keyword"), page, pageSize)
	if err != nil {
		if errors.Is(err, service.ErrAuditKeywordRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "keyword is required"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search audit logs"})
		return
	}

	writeAuditList(c, logs, total, page, pageSize)
}

func (h *AuditHandler) GetUserAuditLogs(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	page, pageSize := auditPagination(c)

	logs, total, err := h.auditService.GetUserAuditLogs(uint(userID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user audit logs"})
		return
	}

	writeAuditList(c, logs, total, page, pageSize)
}

func (h *AuditHandler) DeleteOldLogs(c *gin.Context) {
	var req struct {
		BeforeDate string `json:"before_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.auditService.DeleteOldLogs(req.BeforeDate); err != nil {
		if errors.Is(err, service.ErrInvalidAuditDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete old audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted successfully"})
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

func parseOptionalAuditUserID(value string) (uint, error) {
	if value == "" {
		return 0, nil
	}
	userID, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(userID), nil
}

func writeAuditList(c *gin.Context, logs interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
