package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders 设置安全相关的 HTTP 响应头
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "frame-ancestors 'none'; object-src 'none'; base-uri 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(self)")
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Next()
	}
}
