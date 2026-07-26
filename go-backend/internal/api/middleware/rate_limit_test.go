package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimitByUserUsesUserIDContextKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/limited", func(c *gin.Context) {
		c.Set("user_id", c.GetHeader("X-Test-User"))
	}, RateLimitByUser(1), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	firstUserRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/limited", nil)
	firstUserRequest.Header.Set("X-Test-User", "1")
	firstUserRecorder := httptest.NewRecorder()
	router.ServeHTTP(firstUserRecorder, firstUserRequest)
	if firstUserRecorder.Code != http.StatusOK {
		t.Fatalf("first user first request status = %d, want %d", firstUserRecorder.Code, http.StatusOK)
	}

	secondUserRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/limited", nil)
	secondUserRequest.Header.Set("X-Test-User", "1")
	secondUserRecorder := httptest.NewRecorder()
	router.ServeHTTP(secondUserRecorder, secondUserRequest)
	if secondUserRecorder.Code != http.StatusOK {
		t.Fatalf("same user burst request status = %d, want %d", secondUserRecorder.Code, http.StatusOK)
	}

	thirdUserRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/limited", nil)
	thirdUserRequest.Header.Set("X-Test-User", "1")
	thirdUserRecorder := httptest.NewRecorder()
	router.ServeHTTP(thirdUserRecorder, thirdUserRequest)
	if thirdUserRecorder.Code != http.StatusTooManyRequests {
		t.Fatalf("same user over burst request status = %d, want %d", thirdUserRecorder.Code, http.StatusTooManyRequests)
	}

	otherUserRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/limited", nil)
	otherUserRequest.Header.Set("X-Test-User", "2")
	otherUserRecorder := httptest.NewRecorder()
	router.ServeHTTP(otherUserRecorder, otherUserRequest)
	if otherUserRecorder.Code != http.StatusOK {
		t.Fatalf("different user request status = %d, want %d", otherUserRecorder.Code, http.StatusOK)
	}
}
