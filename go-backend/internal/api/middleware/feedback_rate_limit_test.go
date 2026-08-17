package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/pkg/config"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestFeedbackReadRateLimiterBlocksAfterIPBucketIsDrained(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter, cleanup := newTestFeedbackReadRateLimiter(t, config.FeedbackRateLimitConfig{
		Enabled:                    true,
		ReadIPRequestsPerMinute:    60,
		ReadIPBurst:                2,
		WriteIPRequestsPerMinute:   60,
		WriteIPBurst:               2,
		WriteUserRequestsPerMinute: 60,
		WriteUserBurst:             2,
		FailOpen:                   false,
	})
	defer cleanup()

	router := gin.New()
	router.GET("/api/v1/feedback", limiter.Middleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for attempt := 1; attempt <= 2; attempt++ {
		recorder := feedbackRateLimitRequest(router, http.MethodGet, "203.0.113.10", "/api/v1/feedback?thread=support-payment", 0)
		if recorder.Code != http.StatusOK {
			t.Fatalf("attempt %d status = %d, want %d: %s", attempt, recorder.Code, http.StatusOK, recorder.Body.String())
		}
	}

	recorder := feedbackRateLimitRequest(router, http.MethodGet, "203.0.113.10", "/api/v1/feedback?thread=support-payment", 0)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("third IP request status = %d, want %d: %s", recorder.Code, http.StatusTooManyRequests, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"dimension":"ip"`) {
		t.Fatalf("response does not expose IP dimension: %s", recorder.Body.String())
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}

	otherIPRecorder := feedbackRateLimitRequest(router, http.MethodGet, "203.0.113.11", "/api/v1/feedback?thread=support-payment", 0)
	if otherIPRecorder.Code != http.StatusOK {
		t.Fatalf("different IP status = %d, want %d: %s", otherIPRecorder.Code, http.StatusOK, otherIPRecorder.Body.String())
	}
}

func TestFeedbackWriteRateLimiterBlocksSameUserAcrossIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter, cleanup := newTestFeedbackWriteRateLimiter(t, config.FeedbackRateLimitConfig{
		Enabled:                    true,
		ReadIPRequestsPerMinute:    60,
		ReadIPBurst:                100,
		WriteIPRequestsPerMinute:   60,
		WriteIPBurst:               100,
		WriteUserRequestsPerMinute: 60,
		WriteUserBurst:             1,
		FailOpen:                   false,
	})
	defer cleanup()

	router := gin.New()
	router.POST("/api/v1/feedback", func(c *gin.Context) {
		userID, _ := strconv.ParseUint(c.GetHeader("X-Test-User"), 10, 64)
		c.Set("user_id", uint(userID))
	}, limiter.Middleware(), func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	first := feedbackRateLimitRequest(router, http.MethodPost, "203.0.113.10", "/api/v1/feedback", 7)
	if first.Code != http.StatusCreated {
		t.Fatalf("first user request status = %d, want %d: %s", first.Code, http.StatusCreated, first.Body.String())
	}

	second := feedbackRateLimitRequest(router, http.MethodPost, "203.0.113.11", "/api/v1/feedback", 7)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second user request status = %d, want %d: %s", second.Code, http.StatusTooManyRequests, second.Body.String())
	}
	if !strings.Contains(second.Body.String(), `"dimension":"user"`) {
		t.Fatalf("response does not expose user dimension: %s", second.Body.String())
	}

	otherUser := feedbackRateLimitRequest(router, http.MethodPost, "203.0.113.11", "/api/v1/feedback", 8)
	if otherUser.Code != http.StatusCreated {
		t.Fatalf("different user status = %d, want %d: %s", otherUser.Code, http.StatusCreated, otherUser.Body.String())
	}
}

func TestFeedbackRateLimiterRecordsBlockedCounters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, redisClient := newFeedbackRateLimitRedis(t)
	limiter := &FeedbackRateLimiter{
		store: redisFeedbackRateLimitStore{client: redisClient},
		cfg: config.FeedbackRateLimitConfig{
			Enabled:                    true,
			ReadIPRequestsPerMinute:    60,
			ReadIPBurst:                1,
			WriteIPRequestsPerMinute:   60,
			WriteIPBurst:               1,
			WriteUserRequestsPerMinute: 60,
			WriteUserBurst:             1,
			FailOpen:                   false,
		},
		kind: feedbackRateLimitRead,
		now:  time.Now,
	}

	first, err := limiter.Allow(context.Background(), feedbackRateLimitIdentity{IPAddress: "203.0.113.10"})
	if err != nil || !first.Allowed {
		t.Fatalf("first Allow() = %#v, err = %v; want allowed", first, err)
	}
	second, err := limiter.Allow(context.Background(), feedbackRateLimitIdentity{IPAddress: "203.0.113.10"})
	if err != nil {
		t.Fatalf("second Allow() error = %v", err)
	}
	if second.Allowed {
		t.Fatalf("second Allow() = %#v, want blocked", second)
	}

	summary, err := FeedbackRateLimitBlockedCounts(context.Background(), redisClient, 1)
	if err != nil {
		t.Fatalf("FeedbackRateLimitBlockedCounts() error = %v", err)
	}
	if summary.Total != 1 || summary.ReadIP != 1 {
		t.Fatalf("blocked summary = %+v, want total=1 read_ip=1", summary)
	}
}

func TestFeedbackRateLimiterFallsBackLocallyWhenRedisUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	currentTime := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	localStore := newFeedbackLocalRateLimitStore(func() time.Time {
		return currentTime
	})
	limiter := &FeedbackRateLimiter{
		cfg: config.FeedbackRateLimitConfig{
			Enabled:                    true,
			ReadIPRequestsPerMinute:    60,
			ReadIPBurst:                1,
			WriteIPRequestsPerMinute:   60,
			WriteIPBurst:               1,
			WriteUserRequestsPerMinute: 60,
			WriteUserBurst:             1,
			FailOpen:                   false,
		},
		kind:          feedbackRateLimitRead,
		now:           func() time.Time { return currentTime },
		localFallback: localStore,
	}

	router := gin.New()
	router.GET("/api/v1/feedback", limiter.Middleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	first := feedbackRateLimitRequest(router, http.MethodGet, "203.0.113.10", "/api/v1/feedback?thread=support-payment", 0)
	if first.Code != http.StatusOK {
		t.Fatalf("first fallback request status = %d, want %d: %s", first.Code, http.StatusOK, first.Body.String())
	}
	if first.Header().Get("X-Feedback-RateLimit-Mode") != feedbackRateLimitModeLocalFallback {
		t.Fatalf("first fallback mode header = %q, want %q", first.Header().Get("X-Feedback-RateLimit-Mode"), feedbackRateLimitModeLocalFallback)
	}

	second := feedbackRateLimitRequest(router, http.MethodGet, "203.0.113.10", "/api/v1/feedback?thread=support-payment", 0)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second fallback request status = %d, want %d: %s", second.Code, http.StatusTooManyRequests, second.Body.String())
	}
	if !strings.Contains(second.Body.String(), `"mode":"local_fallback"`) {
		t.Fatalf("fallback block response does not expose mode: %s", second.Body.String())
	}
	if second.Header().Get("X-Feedback-RateLimit-Mode") != feedbackRateLimitModeLocalFallback {
		t.Fatalf("second fallback mode header = %q, want %q", second.Header().Get("X-Feedback-RateLimit-Mode"), feedbackRateLimitModeLocalFallback)
	}

	summary := localStore.BlockedCounts(1)
	if summary.Total != 1 || summary.ReadIP != 1 || summary.RedisUnavailable != 2 {
		t.Fatalf("local fallback summary = %+v, want total=1 read_ip=1 redis_unavailable=2", summary)
	}
}

func TestFeedbackRateLimiterLocalFallbackCapsBucketCardinality(t *testing.T) {
	currentTime := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	localStore := newFeedbackLocalRateLimitStore(func() time.Time {
		return currentTime
	})
	localStore.mu.Lock()
	for index := 0; index < feedbackLocalRateLimitMaxBuckets; index++ {
		localStore.buckets[strconv.Itoa(index)] = feedbackLocalRateLimitBucket{
			Tokens:    1,
			UpdatedAt: currentTime,
		}
	}
	localStore.mu.Unlock()

	decision, err := localStore.Allow(currentTime, feedbackRateLimitRead, config.FeedbackRateLimitConfig{
		Enabled:                 true,
		ReadIPRequestsPerMinute: 60,
		ReadIPBurst:             1,
	}, feedbackRateLimitIdentity{IPAddress: "203.0.113.10"})
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if decision.Allowed || decision.Mode != feedbackRateLimitModeLocalFallback || decision.Dimension != "ip" {
		t.Fatalf("capacity decision = %+v, want local fallback IP block", decision)
	}
}

func TestFeedbackRateLimiterKeysDoNotLeakRawIdentifiers(t *testing.T) {
	redisServer, redisClient := newFeedbackRateLimitRedis(t)
	limiter := &FeedbackRateLimiter{
		store: redisFeedbackRateLimitStore{client: redisClient},
		cfg: config.FeedbackRateLimitConfig{
			Enabled:                    true,
			ReadIPRequestsPerMinute:    60,
			ReadIPBurst:                10,
			WriteIPRequestsPerMinute:   60,
			WriteIPBurst:               10,
			WriteUserRequestsPerMinute: 60,
			WriteUserBurst:             10,
		},
		kind: feedbackRateLimitWrite,
		now: func() time.Time {
			return time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
		},
	}

	_, err := limiter.Allow(context.Background(), feedbackRateLimitIdentity{
		IPAddress: "203.0.113.10",
		UserID:    "customer-secret-identifier",
	})
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}

	keys := redisServer.Keys()
	joined := strings.Join(keys, "|")
	for _, raw := range []string{"203.0.113.10", "customer-secret-identifier"} {
		if strings.Contains(joined, raw) {
			t.Fatalf("feedback rate limit key leaked raw value %q in %q", raw, joined)
		}
	}
}

func newTestFeedbackReadRateLimiter(t *testing.T, cfg config.FeedbackRateLimitConfig) (*FeedbackRateLimiter, func()) {
	t.Helper()

	_, redisClient := newFeedbackRateLimitRedis(t)
	limiter := &FeedbackRateLimiter{
		store: redisFeedbackRateLimitStore{client: redisClient},
		cfg:   cfg,
		kind:  feedbackRateLimitRead,
		now: func() time.Time {
			return time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
		},
	}
	return limiter, func() {
		_ = redisClient.Close()
	}
}

func newTestFeedbackWriteRateLimiter(t *testing.T, cfg config.FeedbackRateLimitConfig) (*FeedbackRateLimiter, func()) {
	t.Helper()

	_, redisClient := newFeedbackRateLimitRedis(t)
	limiter := &FeedbackRateLimiter{
		store: redisFeedbackRateLimitStore{client: redisClient},
		cfg:   cfg,
		kind:  feedbackRateLimitWrite,
		now: func() time.Time {
			return time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
		},
	}
	return limiter, func() {
		_ = redisClient.Close()
	}
}

func newFeedbackRateLimitRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})
	return redisServer, redisClient
}

func feedbackRateLimitRequest(router *gin.Engine, method string, ip string, path string, userID uint) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	request.RemoteAddr = ip + ":12345"
	request.Header.Set("X-Forwarded-For", ip)
	if userID > 0 {
		request.Header.Set("X-Test-User", strconv.FormatUint(uint64(userID), 10))
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}
