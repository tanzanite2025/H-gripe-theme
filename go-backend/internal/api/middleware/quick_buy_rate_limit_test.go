package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/pkg/config"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestQuickBuyRateLimiterBlocksAfterIPBucketIsDrained(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter, cleanup := newTestQuickBuyRateLimiter(t, config.QuickBuyRateLimitConfig{
		Enabled:                  true,
		IPRequestsPerMinute:      60,
		IPBurst:                  2,
		SessionRequestsPerMinute: 60,
		SessionBurst:             100,
		FailOpen:                 false,
	})
	defer cleanup()

	router := gin.New()
	router.GET("/api/v1/quick-buy/sessions/:token", limiter.Middleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for attempt := 1; attempt <= 2; attempt++ {
		recorder := quickBuyRateLimitRequest(router, "203.0.113.10", "/api/v1/quick-buy/sessions/session-a")
		if recorder.Code != http.StatusOK {
			t.Fatalf("attempt %d status = %d, want %d: %s", attempt, recorder.Code, http.StatusOK, recorder.Body.String())
		}
	}

	recorder := quickBuyRateLimitRequest(router, "203.0.113.10", "/api/v1/quick-buy/sessions/session-b")
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("third IP request status = %d, want %d: %s", recorder.Code, http.StatusTooManyRequests, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"dimension":"ip"`) {
		t.Fatalf("response does not expose IP dimension: %s", recorder.Body.String())
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}

	otherIPRecorder := quickBuyRateLimitRequest(router, "203.0.113.11", "/api/v1/quick-buy/sessions/session-b")
	if otherIPRecorder.Code != http.StatusOK {
		t.Fatalf("different IP status = %d, want %d: %s", otherIPRecorder.Code, http.StatusOK, otherIPRecorder.Body.String())
	}
}

func TestQuickBuyRateLimiterBlocksSameSessionAcrossIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter, cleanup := newTestQuickBuyRateLimiter(t, config.QuickBuyRateLimitConfig{
		Enabled:                  true,
		IPRequestsPerMinute:      60,
		IPBurst:                  100,
		SessionRequestsPerMinute: 60,
		SessionBurst:             1,
		FailOpen:                 false,
	})
	defer cleanup()

	router := gin.New()
	router.GET("/api/v1/quick-buy/sessions/:token", limiter.Middleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	first := quickBuyRateLimitRequest(router, "203.0.113.10", "/api/v1/quick-buy/sessions/shared-session")
	if first.Code != http.StatusOK {
		t.Fatalf("first session request status = %d, want %d: %s", first.Code, http.StatusOK, first.Body.String())
	}

	second := quickBuyRateLimitRequest(router, "203.0.113.11", "/api/v1/quick-buy/sessions/shared-session")
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second session request status = %d, want %d: %s", second.Code, http.StatusTooManyRequests, second.Body.String())
	}
	if !strings.Contains(second.Body.String(), `"dimension":"session"`) {
		t.Fatalf("response does not expose session dimension: %s", second.Body.String())
	}

	otherSession := quickBuyRateLimitRequest(router, "203.0.113.11", "/api/v1/quick-buy/sessions/other-session")
	if otherSession.Code != http.StatusOK {
		t.Fatalf("different session status = %d, want %d: %s", otherSession.Code, http.StatusOK, otherSession.Body.String())
	}
}

func TestQuickBuyRateLimiterCreationRequestFallsBackToAnonymousHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter, cleanup := newTestQuickBuyRateLimiter(t, config.QuickBuyRateLimitConfig{
		Enabled:                  true,
		IPRequestsPerMinute:      60,
		IPBurst:                  100,
		SessionRequestsPerMinute: 60,
		SessionBurst:             1,
		FailOpen:                 false,
	})
	defer cleanup()

	router := gin.New()
	router.POST("/api/v1/quick-buy/sessions", limiter.Middleware(), func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	first := quickBuyRateLimitRequestWithHeader(router, http.MethodPost, "203.0.113.10", "/api/v1/quick-buy/sessions", "X-Anonymous-ID", "anonymous-a")
	if first.Code != http.StatusCreated {
		t.Fatalf("first anonymous request status = %d, want %d: %s", first.Code, http.StatusCreated, first.Body.String())
	}

	second := quickBuyRateLimitRequestWithHeader(router, http.MethodPost, "203.0.113.11", "/api/v1/quick-buy/sessions", "X-Anonymous-ID", "anonymous-a")
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second anonymous request status = %d, want %d: %s", second.Code, http.StatusTooManyRequests, second.Body.String())
	}
	if !strings.Contains(second.Body.String(), `"dimension":"session"`) {
		t.Fatalf("response does not expose session dimension: %s", second.Body.String())
	}
}

func TestQuickBuyRateLimiterKeysDoNotLeakRawIdentifiers(t *testing.T) {
	redisServer, redisClient := newQuickBuyRateLimitRedis(t)
	limiter := &QuickBuyRateLimiter{
		store: redisQuickBuyRateLimitStore{client: redisClient},
		cfg: config.QuickBuyRateLimitConfig{
			Enabled:                  true,
			IPRequestsPerMinute:      60,
			IPBurst:                  10,
			SessionRequestsPerMinute: 60,
			SessionBurst:             10,
		},
		now: func() time.Time {
			return time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
		},
	}

	_, err := limiter.Allow(context.Background(), quickBuyRateLimitIdentity{
		IPAddress: "203.0.113.10",
		SessionID: "quick-session-token",
	})
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}

	keys := redisServer.Keys()
	joined := strings.Join(keys, "|")
	for _, raw := range []string{"203.0.113.10", "quick-session-token"} {
		if strings.Contains(joined, raw) {
			t.Fatalf("quick buy rate limit key leaked raw value %q in %q", raw, joined)
		}
	}
}

func TestQuickBuyRateLimiterFailsOpenWhenConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewQuickBuyRateLimiter(nil, config.QuickBuyRateLimitConfig{
		Enabled:  true,
		FailOpen: true,
	})
	router := gin.New()
	router.GET("/api/v1/quick-buy/sessions/:token", limiter.Middleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	recorder := quickBuyRateLimitRequest(router, "203.0.113.10", "/api/v1/quick-buy/sessions/session-a")
	if recorder.Code != http.StatusOK {
		t.Fatalf("fail-open request status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
}

func TestQuickBuyRateLimiterFailsClosedWhenConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewQuickBuyRateLimiter(nil, config.QuickBuyRateLimitConfig{
		Enabled:  true,
		FailOpen: false,
	})
	router := gin.New()
	router.GET("/api/v1/quick-buy/sessions/:token", limiter.Middleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	recorder := quickBuyRateLimitRequest(router, "203.0.113.10", "/api/v1/quick-buy/sessions/session-a")
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("fail-closed request status = %d, want %d: %s", recorder.Code, http.StatusServiceUnavailable, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "quick_buy_rate_limit_unavailable") {
		t.Fatalf("fail-closed response did not expose unavailable error: %s", recorder.Body.String())
	}
}

func TestQuickBuyRateLimiterRefillsDrainedBucket(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	_, redisClient := newQuickBuyRateLimitRedis(t)
	limiter := &QuickBuyRateLimiter{
		store: redisQuickBuyRateLimitStore{client: redisClient},
		cfg: config.QuickBuyRateLimitConfig{
			Enabled:                  true,
			IPRequestsPerMinute:      60,
			IPBurst:                  1,
			SessionRequestsPerMinute: 60,
			SessionBurst:             100,
			FailOpen:                 false,
		},
		now: func() time.Time {
			return now
		},
	}

	first, err := limiter.Allow(context.Background(), quickBuyRateLimitIdentity{
		IPAddress: "203.0.113.10",
		SessionID: "session-a",
	})
	if err != nil || !first.Allowed {
		t.Fatalf("first Allow() = %#v, err = %v; want allowed", first, err)
	}

	second, err := limiter.Allow(context.Background(), quickBuyRateLimitIdentity{
		IPAddress: "203.0.113.10",
		SessionID: "session-b",
	})
	if err != nil {
		t.Fatalf("second Allow() error = %v", err)
	}
	if second.Allowed || second.Dimension != "ip" {
		t.Fatalf("second Allow() = %#v, want IP limit", second)
	}

	now = now.Add(time.Second)
	third, err := limiter.Allow(context.Background(), quickBuyRateLimitIdentity{
		IPAddress: "203.0.113.10",
		SessionID: "session-c",
	})
	if err != nil || !third.Allowed {
		t.Fatalf("third Allow() after refill = %#v, err = %v; want allowed", third, err)
	}
}

func newTestQuickBuyRateLimiter(t *testing.T, cfg config.QuickBuyRateLimitConfig) (*QuickBuyRateLimiter, func()) {
	t.Helper()

	_, redisClient := newQuickBuyRateLimitRedis(t)
	limiter := &QuickBuyRateLimiter{
		store: redisQuickBuyRateLimitStore{client: redisClient},
		cfg:   cfg,
		now: func() time.Time {
			return time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
		},
	}
	return limiter, func() {
		_ = redisClient.Close()
	}
}

func newQuickBuyRateLimitRedis(t *testing.T) (*miniredis.Miniredis, redis.UniversalClient) {
	t.Helper()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})
	return redisServer, redisClient
}

func quickBuyRateLimitRequest(router *gin.Engine, ip string, path string) *httptest.ResponseRecorder {
	return quickBuyRateLimitRequestWithHeader(router, http.MethodGet, ip, path, "", "")
}

func quickBuyRateLimitRequestWithHeader(router *gin.Engine, method string, ip string, path string, header string, value string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	request.RemoteAddr = ip + ":12345"
	request.Header.Set("X-Forwarded-For", ip)
	if header != "" {
		request.Header.Set(header, value)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}
