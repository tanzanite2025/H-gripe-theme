package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TraceMiddleware generates a unique UUID and sets it in the request context
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := uuid.New().String()
		c.Set("X-Trace-ID", traceID)
		
		// Set in standard request context so services can extract it
		ctx := context.WithValue(c.Request.Context(), "X-Trace-ID", traceID)
		c.Request = c.Request.WithContext(ctx)
		
		c.Next()
	}
}
