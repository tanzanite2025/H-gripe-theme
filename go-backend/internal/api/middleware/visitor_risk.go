package middleware

import (
	"strings"
	"time"

	"commerce-platform/internal/pkg/visitorcookie"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func VisitorRiskTelemetry(visitorRiskService *service.VisitorRiskService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if visitorRiskService == nil || !visitorRiskService.Enabled() {
			c.Next()
			return
		}

		startedAt := time.Now().UTC()
		path := ""
		if c.Request != nil && c.Request.URL != nil {
			path = c.Request.URL.Path
		}

		c.Next()

		visitorRiskService.RecordRequest(service.VisitorRiskRecordInput{
			IPAddress:        c.ClientIP(),
			UserAgent:        c.GetHeader("User-Agent"),
			Path:             path,
			CountryCode:      visitorRiskCountryCode(c),
			AnonymousID:      visitorRiskAnonymousID(c),
			SessionID:        visitorRiskSessionID(c),
			HasCookieHeader:  strings.TrimSpace(c.GetHeader("Cookie")) != "",
			StatusCode:       c.Writer.Status(),
			OccurredAt:       startedAt,
			MeaningfulAction: visitorRiskMeaningfulAction(c.Request.Method, path, c.Writer.Status()),
		})
	}
}

func visitorRiskCountryCode(c *gin.Context) string {
	for _, header := range []string{"CF-IPCountry", "CloudFront-Viewer-Country", "X-Vercel-IP-Country"} {
		if value := strings.TrimSpace(c.GetHeader(header)); value != "" && !strings.EqualFold(value, "XX") {
			return value
		}
	}
	return ""
}

func visitorRiskAnonymousID(c *gin.Context) string {
	for _, header := range []string{"X-Commerce-Platform-Anonymous-ID", "X-Anonymous-ID"} {
		if value := strings.TrimSpace(c.GetHeader(header)); value != "" {
			return value
		}
	}
	return ""
}

func visitorRiskSessionID(c *gin.Context) string {
	for _, cookieName := range []string{"session_id", visitorcookie.CustomerServiceVisitorCookie} {
		if value, err := c.Cookie(cookieName); err == nil && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func visitorRiskMeaningfulAction(method string, path string, statusCode int) bool {
	if statusCode >= 400 {
		return false
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	path = strings.ToLower(strings.TrimSpace(path))
	if method == "GET" || method == "HEAD" || path == "" {
		return false
	}

	meaningfulPrefixes := []string{
		"/api/v1/cart/add",
		"/api/v1/cart/sync",
		"/api/v1/cart/items/",
		"/api/v1/cart/clear",
		"/api/v1/customer-service/messages",
		"/api/v1/orders",
		"/api/v1/checkout",
		"/api/v1/wishlist",
		"/api/v1/spoke/calc",
		"/api/v1/subscriptions",
		"/api/v1/registrations",
	}
	for _, prefix := range meaningfulPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
