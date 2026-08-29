package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	appLogger "commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/pkg/metrics"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GlobalIPBlocker applies durable IP/CIDR rules before any public or admin
// handler runs. The rule-management path is deliberately exempt so an
// authenticated administrator can recover from an accidental self-block.
func GlobalIPBlocker(blockService *service.GlobalIPBlockService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c == nil || c.Request == nil {
			c.Next()
			return
		}
		if isGlobalIPBlockBypassPath(c.Request.URL.Path) {
			c.Next()
			return
		}
		if blockService == nil {
			metrics.GlobalIPBlockChecks.WithLabelValues("service_unavailable").Inc()
			appLogger.Error("global IP block service is not configured")
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "IP block enforcement unavailable",
				"code":  "ip_block_unavailable",
			})
			return
		}

		blocked, rule, err := blockService.IsBlocked(c.Request.Context(), c.ClientIP(), time.Now().UTC())
		if err != nil {
			metrics.GlobalIPBlockChecks.WithLabelValues("service_unavailable").Inc()
			appLogger.Error(
				"global IP block enforcement unavailable",
				zap.Error(err),
				zap.Bool("cache_unavailable", errors.Is(err, service.ErrIPBlockCacheUnavailable)),
			)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "IP block enforcement unavailable",
				"code":  "ip_block_unavailable",
			})
			return
		}
		if !blocked {
			metrics.GlobalIPBlockChecks.WithLabelValues("allowed").Inc()
			c.Next()
			return
		}

		metrics.GlobalIPBlockChecks.WithLabelValues("blocked").Inc()
		c.Header("Cache-Control", "no-store, max-age=0")
		c.Header("X-Robots-Tag", "noindex, nofollow, noarchive")
		c.Header("X-Access-Block", "ip")
		response := gin.H{
			"error": "access blocked",
			"code":  "ip_blocked",
		}
		if rule != nil {
			response["rule_id"] = rule.ID
		}
		c.AbortWithStatusJSON(http.StatusForbidden, response)
	}
}

func isGlobalIPBlockBypassPath(path string) bool {
	path = strings.TrimSpace(path)
	switch path {
	case "/health", "/readiness", "/ready", "/liveness", "/_internal/security/ip-block-check":
		return true
	}
	return path == "/api/admin/security/ip-blocks" ||
		strings.HasPrefix(path, "/api/admin/security/ip-blocks/")
}
