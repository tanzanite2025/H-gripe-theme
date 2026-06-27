package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Handler 健康检查处理器
type Handler struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewHandler 创建处理器
func NewHandler(db *gorm.DB, redis *redis.Client) *Handler {
	return &Handler{
		db:    db,
		redis: redis,
	}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status   string            `json:"status"`
	Version  string            `json:"version"`
	Time     string            `json:"time"`
	Services map[string]string `json:"services"`
}

// Health 健康检查端点
func (h *Handler) Health(c *gin.Context) {
	ctx := context.Background()
	services := make(map[string]string)

	// 检查数据库
	dbStatus := "healthy"
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil {
			dbStatus = "error: " + err.Error()
		} else {
			if err := sqlDB.Ping(); err != nil {
				dbStatus = "unhealthy: " + err.Error()
			}
		}
	} else {
		dbStatus = "not configured"
	}
	services["database"] = dbStatus

	// 检查Redis
	redisStatus := "healthy"
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			redisStatus = "unhealthy: " + err.Error()
		}
	} else {
		redisStatus = "not configured"
	}
	services["redis"] = redisStatus

	// 确定总体状态
	overallStatus := "healthy"
	if dbStatus != "healthy" || redisStatus != "healthy" {
		overallStatus = "degraded"
	}
	if dbStatus != "healthy" && redisStatus != "healthy" {
		overallStatus = "unhealthy"
	}

	statusCode := http.StatusOK
	switch overallStatus {
	case "unhealthy":
		statusCode = http.StatusServiceUnavailable
	case "degraded":
		statusCode = http.StatusOK // 降级但仍可服务
	}

	response := HealthResponse{
		Status:   overallStatus,
		Version:  "1.0.0",
		Time:     time.Now().Format(time.RFC3339),
		Services: services,
	}

	c.JSON(statusCode, response)
}

// Readiness 就绪检查端点（Kubernetes探针）
func (h *Handler) Readiness(c *gin.Context) {
	ctx := context.Background()

	// 检查数据库连接
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"reason": "database not available",
			})
			return
		}
	}

	// 检查Redis连接
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"reason": "redis not available",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// Liveness 存活检查端点（Kubernetes探针）
func (h *Handler) Liveness(c *gin.Context) {
	// 简单的存活检查，只要服务在运行就返回200
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// DetailedHealth 详细健康检查（仅供管理员）
func (h *Handler) DetailedHealth(c *gin.Context) {
	ctx := context.Background()
	details := make(map[string]interface{})

	// 数据库详情
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err == nil {
			stats := sqlDB.Stats()
			details["database"] = map[string]interface{}{
				"status":           "healthy",
				"max_open_conns":   stats.MaxOpenConnections,
				"open_conns":       stats.OpenConnections,
				"in_use":           stats.InUse,
				"idle":             stats.Idle,
				"wait_count":       stats.WaitCount,
				"wait_duration_ms": stats.WaitDuration.Milliseconds(),
			}
		} else {
			details["database"] = map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			}
		}
	}

	// Redis详情
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err == nil {
			poolStats := h.redis.PoolStats()
			details["redis"] = map[string]interface{}{
				"status":      "healthy",
				"hits":        poolStats.Hits,
				"misses":      poolStats.Misses,
				"timeouts":    poolStats.Timeouts,
				"total_conns": poolStats.TotalConns,
				"idle_conns":  poolStats.IdleConns,
				"stale_conns": poolStats.StaleConns,
			}
		} else {
			details["redis"] = map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"time":    time.Now().Format(time.RFC3339),
		"version": "1.0.0",
		"details": details,
	})
}

// RegisterRoutes 注册健康检查路由
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, redis *redis.Client) {
	handler := NewHandler(db, redis)

	// 公开健康检查端点
	r.GET("/health", handler.Health)
	r.GET("/readiness", handler.Readiness)
	r.GET("/liveness", handler.Liveness)

	// 详细健康检查（需要认证）
	// r.GET("/health/detailed", middleware.Auth(), handler.DetailedHealth)
}
