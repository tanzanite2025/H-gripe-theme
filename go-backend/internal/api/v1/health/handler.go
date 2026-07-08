package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Handler struct {
	db        *gorm.DB
	redis     *redis.Client
	version   string
	buildTime string
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	BuildTime string            `json:"buildTime,omitempty"`
	Time      string            `json:"time"`
	Services  map[string]string `json:"services"`
}

func NewHandler(db *gorm.DB, redis *redis.Client, version, buildTime string) *Handler {
	return &Handler{
		db:        db,
		redis:     redis,
		version:   version,
		buildTime: buildTime,
	}
}

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, redis *redis.Client, version, buildTime string) {
	handler := NewHandler(db, redis, version, buildTime)

	r.GET("/health", handler.Health)
	r.GET("/readiness", handler.Readiness)
	r.GET("/ready", handler.Readiness)
	r.GET("/liveness", handler.Liveness)
}

func (h *Handler) Health(c *gin.Context) {
	services := map[string]string{
		"database": h.databaseStatus(),
		"redis":    h.redisStatus(c.Request.Context()),
	}

	status := overallStatus(services)
	statusCode := http.StatusOK
	if status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, HealthResponse{
		Status:    status,
		Version:   h.version,
		BuildTime: h.buildTime,
		Time:      time.Now().Format(time.RFC3339),
		Services:  services,
	})
}

func (h *Handler) Readiness(c *gin.Context) {
	if h.databaseStatus() != "healthy" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "database not available",
		})
		return
	}

	if h.redisStatus(c.Request.Context()) != "healthy" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "redis not available",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (h *Handler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (h *Handler) databaseStatus() string {
	if h.db == nil {
		return "not configured"
	}

	sqlDB, err := h.db.DB()
	if err != nil {
		return "error: " + err.Error()
	}
	if err := sqlDB.Ping(); err != nil {
		return "unhealthy: " + err.Error()
	}
	return "healthy"
}

func (h *Handler) redisStatus(ctx context.Context) string {
	if h.redis == nil {
		return "not configured"
	}
	if err := h.redis.Ping(ctx).Err(); err != nil {
		return "unhealthy: " + err.Error()
	}
	return "healthy"
}

func overallStatus(services map[string]string) string {
	unhealthyCount := 0
	for _, status := range services {
		if status != "healthy" {
			unhealthyCount++
		}
	}

	switch unhealthyCount {
	case 0:
		return "healthy"
	case len(services):
		return "unhealthy"
	default:
		return "degraded"
	}
}
