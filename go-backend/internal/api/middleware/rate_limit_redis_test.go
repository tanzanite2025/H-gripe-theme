package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestRateLimitByUserPerMinuteRedisSharesWindowAcrossRouters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	buildRouter := func() *gin.Engine {
		router := gin.New()
		router.POST(
			"/limited",
			func(c *gin.Context) {
				c.Set("user_id", c.GetHeader("X-Test-User"))
				c.Next()
			},
			RateLimitByUserPerMinuteRedis(client),
			func(c *gin.Context) {
				c.Status(http.StatusOK)
			},
		)
		return router
	}

	firstRouter := buildRouter()
	secondRouter := buildRouter()

	first := rateLimitRedisRequest(firstRouter, "1")
	require.Equal(t, http.StatusOK, first.Code)

	second := rateLimitRedisRequest(secondRouter, "1")
	require.Equal(t, http.StatusTooManyRequests, second.Code)
	require.Equal(t, "60", second.Header().Get("Retry-After"))
	require.Contains(t, second.Body.String(), "rate_limit_exceeded")

	otherUser := rateLimitRedisRequest(secondRouter, "2")
	require.Equal(t, http.StatusOK, otherUser.Code)

	server.FastForward(time.Minute)
	afterWindow := rateLimitRedisRequest(secondRouter, "1")
	require.Equal(t, http.StatusOK, afterWindow.Code)
}

func TestRateLimitByUserPerMinuteRedisFailsClosedWithoutRedis(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST(
		"/limited",
		func(c *gin.Context) {
			c.Set("user_id", "1")
			c.Next()
		},
		RateLimitByUserPerMinuteRedis(nil),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/limited", nil)
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusServiceUnavailable, response.Code)
	require.Contains(t, response.Body.String(), "rate_limit_service_unavailable")
}

func rateLimitRedisRequest(router *gin.Engine, userID string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/limited", nil)
	request.Header.Set("X-Test-User", userID)
	router.ServeHTTP(response, request)
	return response
}
