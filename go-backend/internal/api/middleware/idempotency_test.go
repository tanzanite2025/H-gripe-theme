package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdempotencyReplaysCompletedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := newTestIdempotencyRedisClient(t)
	router := gin.New()

	var calls int32
	router.POST("/api/v1/orders", Idempotency(client), func(c *gin.Context) {
		atomic.AddInt32(&calls, 1)
		c.JSON(http.StatusCreated, gin.H{
			"order_number": "TZ-123456",
			"status":       "created",
		})
	})

	reqBody := []byte(`{"payment_method":"card","shipping_method":"standard"}`)
	first := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/orders", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(idempotencyKeyHeader, "checkout-uuid-1")
	router.ServeHTTP(first, req)

	require.Equal(t, http.StatusCreated, first.Code)
	require.Equal(t, int32(1), atomic.LoadInt32(&calls))
	require.Empty(t, first.Header().Get(idempotencyReplayHeader))
	require.Contains(t, first.Body.String(), `"order_number":"TZ-123456"`)

	second := httptest.NewRecorder()
	req2 := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/orders", bytes.NewReader(reqBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set(idempotencyKeyHeader, "checkout-uuid-1")
	router.ServeHTTP(second, req2)

	require.Equal(t, http.StatusCreated, second.Code)
	require.Equal(t, int32(1), atomic.LoadInt32(&calls))
	assert.Equal(t, "true", second.Header().Get(idempotencyReplayHeader))
	assert.Equal(t, first.Body.String(), second.Body.String())
}

func TestIdempotencyRejectsSameKeyWithDifferentPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := newTestIdempotencyRedisClient(t)
	router := gin.New()

	var calls int32
	router.POST("/api/v1/orders", Idempotency(client), func(c *gin.Context) {
		atomic.AddInt32(&calls, 1)
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	req1 := httptest.NewRecorder()
	first := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/orders", bytes.NewReader([]byte(`{"payment_method":"card"}`)))
	first.Header.Set("Content-Type", "application/json")
	first.Header.Set(idempotencyKeyHeader, "checkout-uuid-2")
	router.ServeHTTP(req1, first)
	require.Equal(t, http.StatusCreated, req1.Code)

	req2 := httptest.NewRecorder()
	second := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/orders", bytes.NewReader([]byte(`{"payment_method":"paypal"}`)))
	second.Header.Set("Content-Type", "application/json")
	second.Header.Set(idempotencyKeyHeader, "checkout-uuid-2")
	router.ServeHTTP(req2, second)

	require.Equal(t, http.StatusConflict, req2.Code)
	require.Equal(t, int32(1), atomic.LoadInt32(&calls))
	require.Contains(t, req2.Body.String(), "idempotency_key_conflict")
}

func TestIdempotencyWaitsForInFlightRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := newTestIdempotencyRedisClient(t)
	router := gin.New()

	started := make(chan struct{})
	release := make(chan struct{})
	var calls int32

	router.POST("/api/v1/orders", Idempotency(client), func(c *gin.Context) {
		if atomic.AddInt32(&calls, 1) == 1 {
			close(started)
			<-release
		}
		c.JSON(http.StatusCreated, gin.H{
			"order_number": "TZ-999999",
			"status":       "created",
		})
	})

	reqBody := []byte(`{"payment_method":"card","shipping_method":"standard"}`)

	firstRecorder := httptest.NewRecorder()
	firstRequest := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/orders", bytes.NewReader(reqBody))
	firstRequest.Header.Set("Content-Type", "application/json")
	firstRequest.Header.Set(idempotencyKeyHeader, "checkout-uuid-3")

	secondRecorder := httptest.NewRecorder()
	secondRequest := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/orders", bytes.NewReader(reqBody))
	secondRequest.Header.Set("Content-Type", "application/json")
	secondRequest.Header.Set(idempotencyKeyHeader, "checkout-uuid-3")

	doneFirst := make(chan struct{})
	doneSecond := make(chan struct{})

	go func() {
		defer close(doneFirst)
		router.ServeHTTP(firstRecorder, firstRequest)
	}()

	<-started

	go func() {
		defer close(doneSecond)
		router.ServeHTTP(secondRecorder, secondRequest)
	}()

	select {
	case <-doneSecond:
		t.Fatal("expected duplicate request to wait for the in-flight result")
	case <-time.After(200 * time.Millisecond):
	}

	close(release)

	select {
	case <-doneFirst:
	case <-time.After(3 * time.Second):
		t.Fatal("first request did not finish")
	}
	select {
	case <-doneSecond:
	case <-time.After(3 * time.Second):
		t.Fatal("duplicate request did not replay the result")
	}

	require.Equal(t, http.StatusCreated, firstRecorder.Code)
	require.Equal(t, http.StatusCreated, secondRecorder.Code)
	require.Equal(t, int32(1), atomic.LoadInt32(&calls))
	assert.Equal(t, "true", secondRecorder.Header().Get(idempotencyReplayHeader))
	assert.Equal(t, firstRecorder.Body.String(), secondRecorder.Body.String())
}

func TestIdempotencyNormalizesJSONBodyForHash(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := newTestIdempotencyRedisClient(t)
	router := gin.New()

	var calls int32
	router.POST("/api/v1/orders", Idempotency(client), func(c *gin.Context) {
		atomic.AddInt32(&calls, 1)
		c.JSON(http.StatusCreated, gin.H{
			"order_number": "TZ-123456",
		})
	})

	first := httptest.NewRecorder()
	req1 := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/orders", bytes.NewReader([]byte(`{"payment_method":"card","shipping_method":"standard"}`)))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set(idempotencyKeyHeader, "checkout-json-normalized")
	router.ServeHTTP(first, req1)

	require.Equal(t, http.StatusCreated, first.Code)
	require.Equal(t, int32(1), atomic.LoadInt32(&calls))

	second := httptest.NewRecorder()
	req2 := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/orders", bytes.NewReader([]byte("{\n  \"shipping_method\": \"standard\",\n  \"payment_method\": \"card\"\n}")))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set(idempotencyKeyHeader, "checkout-json-normalized")
	router.ServeHTTP(second, req2)

	require.Equal(t, http.StatusCreated, second.Code)
	require.Equal(t, int32(1), atomic.LoadInt32(&calls))
	assert.Equal(t, "true", second.Header().Get(idempotencyReplayHeader))
	assert.Equal(t, first.Body.String(), second.Body.String())
}

func TestIdempotencyDoesNotCollapseDistinctLargeJSONNumbers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := newTestIdempotencyRedisClient(t)
	router := gin.New()

	var calls int32
	router.POST("/api/v1/orders", Idempotency(client), func(c *gin.Context) {
		atomic.AddInt32(&calls, 1)
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	first := httptest.NewRecorder()
	firstRequest := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/v1/orders",
		bytes.NewReader([]byte(`{"value":9007199254740992}`)),
	)
	firstRequest.Header.Set(idempotencyKeyHeader, "checkout-large-number")
	router.ServeHTTP(first, firstRequest)
	require.Equal(t, http.StatusCreated, first.Code)

	second := httptest.NewRecorder()
	secondRequest := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/v1/orders",
		bytes.NewReader([]byte(`{"value":9007199254740993}`)),
	)
	secondRequest.Header.Set(idempotencyKeyHeader, "checkout-large-number")
	router.ServeHTTP(second, secondRequest)

	require.Equal(t, http.StatusConflict, second.Code)
	require.Equal(t, int32(1), atomic.LoadInt32(&calls))
	require.Contains(t, second.Body.String(), "idempotency_key_conflict")
}

func TestIdempotencyRedisKeysShareRedisClusterHashTag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/orders", nil)

	recordKey := idempotencyRedisKey(context, "checkout-cluster-key")
	ownerKey := idempotencyOwnerRedisKey(recordKey)

	require.NotEmpty(t, redisClusterHashTag(recordKey))
	require.Equal(t, redisClusterHashTag(recordKey), redisClusterHashTag(ownerKey))
}

func TestIdempotencyDoesNotCacheServerErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := newTestIdempotencyRedisClient(t)
	router := gin.New()

	var calls int32
	router.POST("/api/v1/orders", Idempotency(client), func(c *gin.Context) {
		atomic.AddInt32(&calls, 1)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "boom",
		})
	})

	body := []byte(`{"payment_method":"card","shipping_method":"standard"}`)

	first := httptest.NewRecorder()
	req1 := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set(idempotencyKeyHeader, "checkout-server-error")
	router.ServeHTTP(first, req1)

	second := httptest.NewRecorder()
	req2 := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set(idempotencyKeyHeader, "checkout-server-error")
	router.ServeHTTP(second, req2)

	require.Equal(t, http.StatusInternalServerError, first.Code)
	require.Equal(t, http.StatusInternalServerError, second.Code)
	require.Equal(t, int32(2), atomic.LoadInt32(&calls))
	assert.Empty(t, first.Header().Get(idempotencyReplayHeader))
	assert.Empty(t, second.Header().Get(idempotencyReplayHeader))
}

func TestIdempotencyRejectsTooLargeKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := newTestIdempotencyRedisClient(t)
	router := gin.New()

	var calls int32
	router.POST("/api/v1/orders", Idempotency(client), func(c *gin.Context) {
		atomic.AddInt32(&calls, 1)
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/orders", bytes.NewReader([]byte(`{"ok":true}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(idempotencyKeyHeader, string(bytes.Repeat([]byte("a"), idempotencyMaxKeyBytes+1)))
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Equal(t, int32(0), atomic.LoadInt32(&calls))
	require.Contains(t, recorder.Body.String(), "idempotency_key_too_large")
}

func TestIdempotencyRejectsTooLargeBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := newTestIdempotencyRedisClient(t)
	router := gin.New()

	var calls int32
	router.POST("/api/v1/orders", Idempotency(client), func(c *gin.Context) {
		atomic.AddInt32(&calls, 1)
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/orders", bytes.NewReader(bytes.Repeat([]byte("a"), idempotencyMaxRequestBytes+1)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(idempotencyKeyHeader, "checkout-body-too-large")
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
	require.Equal(t, int32(0), atomic.LoadInt32(&calls))
	require.Contains(t, recorder.Body.String(), "idempotency_body_too_large")
}

func TestIdempotencyRenewLeaseExtendsPendingOwnerTTL(t *testing.T) {
	server, client := newTestIdempotencyRedisServerAndClient(t)
	key := "commerce_platform:idempotency:test-renew"
	ownerKey := idempotencyOwnerRedisKey(key)
	ownerToken := "owner-token"
	requestHash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	acquired, err := acquireIdempotency(context.Background(), client, key, ownerKey, ownerToken, requestHash)
	require.NoError(t, err)
	require.True(t, acquired)

	server.FastForward(idempotencyDefaultTTL - time.Second)
	require.LessOrEqual(t, server.TTL(key), time.Second)
	require.LessOrEqual(t, server.TTL(ownerKey), time.Second)

	renewed, err := renewIdempotencyLease(context.Background(), client, key, ownerKey, ownerToken, idempotencyDefaultTTL)
	require.NoError(t, err)
	require.True(t, renewed)
	require.Greater(t, server.TTL(key), idempotencyDefaultTTL-time.Second)
	require.Greater(t, server.TTL(ownerKey), idempotencyDefaultTTL-time.Second)
}

func TestCompleteIdempotencyReturnsFalseWhenOwnerLeaseWasLost(t *testing.T) {
	server, client := newTestIdempotencyRedisServerAndClient(t)
	key := "commerce_platform:idempotency:test-complete-owner-lost"
	ownerKey := idempotencyOwnerRedisKey(key)
	ownerToken := "owner-token"
	requestHash := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	acquired, err := acquireIdempotency(context.Background(), client, key, ownerKey, ownerToken, requestHash)
	require.NoError(t, err)
	require.True(t, acquired)
	require.True(t, server.Del(ownerKey))

	completed, err := completeIdempotency(context.Background(), client, key, ownerKey, ownerToken, idempotencyRecord{
		State:       idempotencyStateCompleted,
		RequestHash: requestHash,
		StatusCode:  http.StatusCreated,
		Body:        `{"ok":true}`,
	})
	require.NoError(t, err)
	require.False(t, completed)

	record, found, err := loadIdempotencyRecord(context.Background(), client, key)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, idempotencyStatePending, record.State)
}

func newTestIdempotencyRedisClient(t *testing.T) redis.UniversalClient {
	t.Helper()

	_, client := newTestIdempotencyRedisServerAndClient(t)
	return client
}

func newTestIdempotencyRedisServerAndClient(t *testing.T) (*miniredis.Miniredis, redis.UniversalClient) {
	t.Helper()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})
	return server, client
}

func redisClusterHashTag(key string) string {
	start := strings.IndexByte(key, '{')
	if start < 0 {
		return ""
	}
	endOffset := strings.IndexByte(key[start+1:], '}')
	if endOffset < 0 {
		return ""
	}
	return key[start+1 : start+1+endOffset]
}
