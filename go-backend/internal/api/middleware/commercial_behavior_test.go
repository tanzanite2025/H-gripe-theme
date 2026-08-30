package middleware

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestCommercialBehaviorTrackerDetectsCatalogBurst(t *testing.T) {
	tracker := newCommercialBehaviorTracker()
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)

	for index := 0; index < commercialInventoryTargetThreshold; index++ {
		_, unique, _, _ := tracker.observe(
			context.Background(),
			"203.0.113.10",
			"product:"+itoa(index+1),
			nil,
			now,
		)
		if unique != index+1 {
			t.Fatalf("unique target count = %d, want %d", unique, index+1)
		}
	}

	requests, unique, _, _ := tracker.observe(context.Background(), "203.0.113.10", "product:next", nil, now)
	if requests != commercialInventoryTargetThreshold+1 || unique != commercialInventoryTargetThreshold+1 {
		t.Fatalf("tracker counts = (%d, %d), want (%d, %d)", requests, unique, commercialInventoryTargetThreshold+1, commercialInventoryTargetThreshold+1)
	}
}

func TestCommercialBehaviorTrackerFallsBackWhenRedisUnavailable(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Dialer: func(context.Context, string, string) (net.Conn, error) {
			return nil, errors.New("redis unavailable")
		},
	})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	tracker := newCommercialBehaviorTrackerWithRedis(redisClient)
	requests, unique, _, _ := tracker.observe(
		context.Background(),
		"203.0.113.18",
		"product:18",
		nil,
		time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC),
	)
	if requests != 1 || unique != 1 {
		t.Fatalf("fallback tracker counts = (%d, %d), want (1, 1)", requests, unique)
	}
}

func TestCommercialInventoryProbeGuardFallsBackWhenRedisIsSlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	clientConn, serverConn := net.Pipe()
	go func() {
		_, _ = io.Copy(io.Discard, serverConn)
	}()

	redisClient := redis.NewClient(&redis.Options{
		Dialer: func(_ context.Context, _, _ string) (net.Conn, error) {
			return clientConn, nil
		},
		MaxRetries:            0,
		ContextTimeoutEnabled: true,
	})
	t.Cleanup(func() {
		_ = clientConn.Close()
		_ = serverConn.Close()
		_ = redisClient.Close()
	})

	router := gin.New()
	router.Use(CommercialInventoryProbeGuard(redisClient))
	router.GET("/api/v1/products/:id", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/products/1", nil)
	request.Header.Set("X-Forwarded-For", "203.0.113.21")
	recorder := httptest.NewRecorder()
	start := time.Now()
	router.ServeHTTP(recorder, request)
	elapsed := time.Since(start)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("request took %s, want < 500ms", elapsed)
	}
}

func TestCommercialInventoryProbeGuardBlocksCatalogEnumeration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CommercialInventoryProbeGuard())
	router.GET("/api/v1/products/:id", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for id := 1; id <= commercialInventoryTargetThreshold+1; id++ {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/products/"+itoa(id), nil)
		request.Header.Set("X-Forwarded-For", "203.0.113.12")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		want := http.StatusOK
		if id == commercialInventoryTargetThreshold+1 {
			want = http.StatusTooManyRequests
		}
		if recorder.Code != want {
			t.Fatalf("product %d status = %d, want %d", id, recorder.Code, want)
		}
	}
}

func TestCommercialInventoryProbeGuardSharesLimitWithPublicChatSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CommercialInventoryProbeGuard())
	router.GET("/api/v1/products/:id", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/api/v1/customer-service/products", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for id := 1; id <= commercialInventoryTargetThreshold; id++ {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/products/"+itoa(id), nil)
		request.Header.Set("X-Forwarded-For", "203.0.113.15")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("product %d status = %d, want %d", id, recorder.Code, http.StatusOK)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/customer-service/products", nil)
	request.Header.Set("X-Forwarded-For", "203.0.113.15")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("customer-service search status = %d, want %d", recorder.Code, http.StatusTooManyRequests)
	}
}

func TestCommercialOrderEnumerationGuardBlocksHighVolumeOpaqueOrderGuesses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(7))
		c.Next()
	})
	router.Use(CommercialOrderEnumerationGuard())
	router.GET("/api/v1/orders/:order_number", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for index := 1; index <= commercialOrderRequestThreshold+1; index++ {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/orders/TZ-2026-OPAQUE"+itoa(index), nil)
		request.Header.Set("X-Forwarded-For", "203.0.113.11")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		want := http.StatusOK
		if index == commercialOrderRequestThreshold+1 {
			want = http.StatusTooManyRequests
		}
		if recorder.Code != want {
			t.Fatalf("order request %d status = %d, want %d", index, recorder.Code, want)
		}
	}
}

func TestCommercialCartProbeGuardRejectsOversizedQuantity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handlerReached := false
	router.Use(CommercialCartProbeGuard())
	router.POST("/api/v1/cart/add", func(c *gin.Context) {
		handlerReached = true
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/cart/add",
		strings.NewReader(`{"product_id":1,"quantity":9999}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Forwarded-For", "203.0.113.13")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if handlerReached {
		t.Fatal("cart handler ran for an oversized quantity request")
	}
	if strings.Contains(recorder.Body.String(), "stock") {
		t.Fatalf("quantity rejection exposes inventory: %s", recorder.Body.String())
	}
}

func TestCommercialCartProbeGuardBlocksAnonymousBurst(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CommercialCartProbeGuard())
	router.POST("/api/v1/cart/add", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for attempt := 1; attempt <= commercialAnonymousCartWriteThreshold+1; attempt++ {
		request := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/cart/add",
			strings.NewReader(`{"product_id":1,"quantity":1}`),
		)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Forwarded-For", "203.0.113.14")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		want := http.StatusOK
		if attempt == commercialAnonymousCartWriteThreshold+1 {
			want = http.StatusTooManyRequests
		}
		if recorder.Code != want {
			t.Fatalf("attempt %d status = %d, want %d", attempt, recorder.Code, want)
		}
	}
}

func TestCommercialCartProbeGuardBlocksRepeatedSKUWithSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CommercialCartProbeGuard())
	router.POST("/api/v1/cart/add", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for attempt := 1; attempt <= commercialCartTargetRequestThreshold+1; attempt++ {
		request := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/cart/add",
			strings.NewReader(`{"product_id":7,"variant_id":9,"quantity":1}`),
		)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Forwarded-For", "203.0.113.16")
		request.AddCookie(&http.Cookie{Name: "session_id", Value: "repeat-sku-session"})
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		want := http.StatusOK
		if attempt == commercialCartTargetRequestThreshold+1 {
			want = http.StatusTooManyRequests
		}
		if recorder.Code != want {
			t.Fatalf("attempt %d status = %d, want %d", attempt, recorder.Code, want)
		}
	}
}

func TestCommercialCartProbeGuardKeepsAccountSKUThrottleWhenIPChanges(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(71))
		c.Next()
	})
	router.Use(CommercialCartProbeGuard())
	router.POST("/api/v1/cart/add", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for attempt := 1; attempt <= commercialCartTargetRequestThreshold+1; attempt++ {
		request := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/cart/add",
			strings.NewReader(`{"product_id":8,"variant_id":10,"quantity":1}`),
		)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Forwarded-For", "203.0.113."+itoa(50+attempt))
		request.AddCookie(&http.Cookie{Name: "session_id", Value: "account-session-" + itoa(attempt)})
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		want := http.StatusOK
		if attempt == commercialCartTargetRequestThreshold+1 {
			want = http.StatusTooManyRequests
		}
		if recorder.Code != want {
			t.Fatalf("attempt %d status = %d, want %d", attempt, recorder.Code, want)
		}
	}
}

func TestCommercialCartIdentitiesUseHashedMultiSignalKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/cart/add", nil)
	request.Header.Set("X-Forwarded-For", "203.0.113.17")
	request.AddCookie(&http.Cookie{Name: "session_id", Value: "cart-session-17"})
	request.AddCookie(&http.Cookie{Name: "tz_customer_service_visitor", Value: "visitor-cookie-17"})
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = request
	context.Set("user_id", uint(17))

	identities := commercialCartIdentities(context)
	if len(identities) != 4 {
		t.Fatalf("identity count = %d, want 4", len(identities))
	}
	for _, identity := range identities {
		if strings.Contains(identity, "203.0.113.17") ||
			strings.Contains(identity, "cart-session-17") ||
			strings.Contains(identity, "visitor-cookie-17") {
			t.Fatalf("identity contains raw identifier: %q", identity)
		}
	}
}

func TestCommercialProtectionAuditContextRedactsRequestIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/cart/add", nil)
	request.Header.Set("X-Forwarded-For", "203.0.113.19")
	request.Header.Set("User-Agent", "commercial-test-agent")
	request.AddCookie(&http.Cookie{Name: "session_id", Value: "audit-session-19"})
	request.AddCookie(&http.Cookie{Name: "tz_customer_service_visitor", Value: "audit-visitor-19"})
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = request
	context.Set("user_id", uint(19))

	auditContext := buildCommercialProtectionAuditContext(
		context,
		commercialProtectionAuditInput{
			SeedID:      "inventory-crawlers",
			Outcome:     "rate_limited",
			Action:      "rate_limit_429",
			Reason:      "behavior_threshold_exceeded",
			Window:      time.Minute,
			ReleaseMode: "automatic_window_expiry",
		},
		time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC),
	)
	if auditContext.RuleVersion != commercialProtectionRuleVersion {
		t.Fatalf("rule version = %q, want %q", auditContext.RuleVersion, commercialProtectionRuleVersion)
	}
	if auditContext.ExpiresAt == nil {
		t.Fatal("audit context omitted expiry time")
	}
	if len(auditContext.IdentityKeys) != 4 {
		t.Fatalf("identity key count = %d, want 4", len(auditContext.IdentityKeys))
	}
	serialized := strings.Join(auditContext.IdentityKeys, "|") + "|" + auditContext.UserAgentHash
	for _, raw := range []string{"203.0.113.19", "audit-session-19", "audit-visitor-19", "commercial-test-agent"} {
		if strings.Contains(serialized, raw) {
			t.Fatalf("audit context contains raw identifier %q: %s", raw, serialized)
		}
	}
}

func TestCommercialInventoryProbeGuardIgnoresInternalSSRIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CommercialInventoryProbeGuard())
	router.GET("/api/v1/products/:id", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for id := 1; id <= commercialInventoryTargetThreshold+5; id++ {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/products/"+itoa(id), nil)
		request.RemoteAddr = "172.20.0.5:1234"
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("internal SSR product %d status = %d, want %d", id, recorder.Code, http.StatusOK)
		}
	}
}

func itoa(value int) string {
	if value < 0 {
		return ""
	}
	digits := make([]byte, 0, 12)
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	if len(digits) == 0 {
		return "0"
	}
	return string(digits)
}
