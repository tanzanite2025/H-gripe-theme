package middleware

import (
	"strconv"
	"time"

	"commerce-platform/internal/pkg/metrics"

	"github.com/gin-gonic/gin"
)

func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		method := c.Request.Method
		metrics.HTTPRequests.WithLabelValues(method, route, strconv.Itoa(c.Writer.Status())).Inc()
		metrics.HTTPDuration.WithLabelValues(method, route).Observe(time.Since(start).Seconds())
	}
}
