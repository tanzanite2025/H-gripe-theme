package audit

import (
	"net/http"
	"strconv"
	"tanzanite/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	auditRepo *repository.AuditRepository
}

func NewHandler(auditRepo *repository.AuditRepository) *Handler {
	return &Handler{
		auditRepo: auditRepo,
	}
}

// GetAuditLog 获取审计日志详情
// @Summary 获取审计日志详情
// @Tags Audit
// @Produce json
// @Param id path int true "日志ID"
// @Success 200 {object} audit.AuditLog
// @Router /api/v1/admin/audit/logs/{id} [get]
func (h *Handler) GetAuditLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid log id"})
		return
	}

	log, err := h.auditRepo.FindAuditLogByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, log)
}

// ListUserAuditLogs 获取用户的审计日志
// @Summary 获取用户的审计日志
// @Tags Audit
// @Produce json
// @Param user_id path int true "用户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/audit/users/{user_id}/logs [get]
func (h *Handler) ListUserAuditLogs(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
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

	logs, total, err := h.auditRepo.FindAuditLogsByUserID(uint(userID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

// ListEntityAuditLogs 获取实体的审计日志
// @Summary 获取实体的审计日志
// @Tags Audit
// @Produce json
// @Param entity_type query string true "实体类型"
// @Param entity_id query int true "实体ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/audit/entities/logs [get]
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := h.auditRepo.FindAuditLogsByEntity(entityType, uint(entityID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

// ListAllAuditLogs 获取所有审计日志（管理员）
// @Summary 获取所有审计日志
// @Tags Audit
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param action query string false "操作类型"
// @Param entity_type query string false "实体类型"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/audit/logs [get]
func (h *Handler) ListAllAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	action := c.Query("action")
	entityType := c.Query("entity_type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := h.auditRepo.FindAllAuditLogs(page, pageSize, action, entityType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

// ListAuditLogsByDateRange 根据日期范围获取审计日志
// @Summary 根据日期范围获取审计日志
// @Tags Audit
// @Produce json
// @Param start_date query string true "开始日期 (YYYY-MM-DD)"
// @Param end_date query string true "结束日期 (YYYY-MM-DD)"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/audit/logs/date-range [get]
func (h *Handler) ListAuditLogsByDateRange(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format (use YYYY-MM-DD)"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format (use YYYY-MM-DD)"})
		return
	}

	// 设置结束日期为当天的23:59:59
	endDate = endDate.Add(24*time.Hour - time.Second)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := h.auditRepo.FindAuditLogsByDateRange(startDate, endDate, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

// ListAuditLogsByIP 根据IP地址获取审计日志
// @Summary 根据IP地址获取审计日志
// @Tags Audit
// @Produce json
// @Param ip_address path string true "IP地址"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/audit/ip/{ip_address}/logs [get]
func (h *Handler) ListAuditLogsByIP(c *gin.Context) {
	ipAddress := c.Param("ip_address")
	if ipAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip_address is required"})
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

	logs, total, err := h.auditRepo.FindAuditLogsByIP(ipAddress, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

// SearchAuditLogs 搜索审计日志
// @Summary 搜索审计日志
// @Tags Audit
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/audit/logs/search [get]
func (h *Handler) SearchAuditLogs(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keyword is required"})
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

	logs, total, err := h.auditRepo.SearchAuditLogs(keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

// GetAuditStats 获取审计统计
// @Summary 获取审计统计
// @Tags Audit
// @Produce json
// @Param start_date query string false "开始日期 (YYYY-MM-DD)"
// @Param end_date query string false "结束日期 (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/audit/stats [get]
func (h *Handler) GetAuditStats(c *gin.Context) {
	var startDate, endDate time.Time
	var err error

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format (use YYYY-MM-DD)"})
			return
		}
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format (use YYYY-MM-DD)"})
			return
		}
		// 设置结束日期为当天的23:59:59
		endDate = endDate.Add(24*time.Hour - time.Second)
	}

	stats, err := h.auditRepo.GetAuditStats(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetRecentActivities 获取最近活动
// @Summary 获取最近活动
// @Tags Audit
// @Produce json
// @Param limit query int false "数量限制" default(50)
// @Success 200 {array} audit.AuditLog
// @Router /api/v1/admin/audit/activities/recent [get]
func (h *Handler) GetRecentActivities(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 200 {
		limit = 50
	}

	logs, err := h.auditRepo.GetRecentActivities(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

// DeleteOldAuditLogs 删除旧的审计日志（管理员）
// @Summary 删除旧的审计日志
// @Tags Audit
// @Accept json
// @Produce json
// @Param request body map[string]string true "删除日期"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/audit/logs/cleanup [post]
func (h *Handler) DeleteOldAuditLogs(c *gin.Context) {
	var req struct {
		BeforeDate string `json:"before_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	beforeDate, err := time.Parse("2006-01-02", req.BeforeDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format (use YYYY-MM-DD)"})
		return
	}

	if err := h.auditRepo.DeleteOldAuditLogs(beforeDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "old audit logs deleted"})
}
