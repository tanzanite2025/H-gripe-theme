package middleware

import (
	"commerce-platform/internal/pkg/config"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查是否在允许列表中
		allowed := false
		isWildcard := false
		for _, allowedOrigin := range cfg.AllowedOrigins {
			if allowedOrigin == "*" {
				isWildcard = true
				break
			}
			if allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if isWildcard && !cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", joinStrings(cfg.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", joinStrings(corsAllowedHeaders(cfg.AllowedHeaders), ", "))
		c.Header("Access-Control-Expose-Headers", joinStrings(corsExposedHeaders(cfg.ExposeHeaders), ", "))

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func corsAllowedHeaders(headers []string) []string {
	headers = appendHeaderIfMissing(headers, idempotencyKeyHeader)
	return appendHeaderIfMissing(headers, "X-Anonymous-ID")
}

func corsExposedHeaders(headers []string) []string {
	headers = appendHeaderIfMissing(headers, idempotencyReplayHeader)
	return appendHeaderIfMissing(headers, "Retry-After")
}

func appendHeaderIfMissing(headers []string, required string) []string {
	for _, header := range headers {
		if strings.EqualFold(strings.TrimSpace(header), required) {
			return headers
		}
	}
	return append(append([]string(nil), headers...), required)
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
