package admin

import (
	"net/http"
	"strconv"
	"tanzanite/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditHandler 审计日志处理器
type AuditHandler struct {
	auditRepo *repository.AuditRepository
}

// NewAuditHandler 创建审计日志处理器
func NewAuditHandler(auditRepo *repository.AuditRepository) *AuditHandler {
	return &AuditHandler{
		auditRepo: auditRepo,
	}
}

// ListAuditLogs 获取审计日志列表
func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	action := c.Query("action")
	resource := c.Query("resource")
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	ipAddress := c.Query("ip_address")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var logs interface{}
	var total int64
	var err error

	// 根据不同的查询条件调用不同的方法
	if userID > 0 {
		logs, total, err = h.auditRepo.FindAuditLogsByUserID(uint(userID), page, pageSize)
	} else if ipAddress != "" {
		logs, total, err = h.auditRepo.FindAuditLogsByIP(ipAddress, page, pageSize)
	} else if startDate != "" && endDate != "" {
		start, _ := time.Parse("2006-01-02", startDate)
		end, _ := time.Parse("2006-01-02", endDate)
		logs, total, err = h.auditRepo.FindAuditLogsByDateRange(start, end, page, pageSize)
	} else {
		logs, total, err = h.auditRepo.FindAllAuditLogs(page, pageSize, action, resource)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取审计日志失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetAuditLog 获取审计日志详情
func (h *AuditHandler) GetAuditLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的日志ID"})
		return
	}

	log, err := h.auditRepo.FindAuditLogByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "审计日志不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"log": log})
}

// GetAuditStats 获取审计统计
func (h *AuditHandler) GetAuditStats(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var start, end time.Time
	if startDate != "" {
		start, _ = time.Parse("2006-01-02", startDate)
	}
	if endDate != "" {
		end, _ = time.Parse("2006-01-02", endDate)
	}

	stats, err := h.auditRepo.GetAuditStats(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetRecentActivities 获取最近活动
func (h *AuditHandler) GetRecentActivities(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	logs, err := h.auditRepo.GetRecentActivities(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取最近活动失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"activities": logs})
}

// SearchAuditLogs 搜索审计日志
func (h *AuditHandler) SearchAuditLogs(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供搜索关键词"})
		return
	}

	logs, total, err := h.auditRepo.SearchAuditLogs(keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetUserAuditLogs 获取用户的审计日志
func (h *AuditHandler) GetUserAuditLogs(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	logs, total, err := h.auditRepo.FindAuditLogsByUserID(uint(userID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户审计日志失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteOldLogs 删除旧日志
func (h *AuditHandler) DeleteOldLogs(c *gin.Context) {
	var req struct {
		BeforeDate string `json:"before_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	beforeDate, err := time.Parse("2006-01-02", req.BeforeDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误"})
		return
	}

	if err := h.auditRepo.DeleteOldAuditLogs(beforeDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除旧日志失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
