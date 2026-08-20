package middleware

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"commerce-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	idempotencyKeyHeader = "Idempotency-Key"

	idempotencyNamespace       = "commerce_platform:idempotency"
	idempotencyStatePending    = "pending"
	idempotencyStateCompleted  = "completed"
	idempotencyReplayHeader    = "Idempotency-Replayed"
	idempotencyDefaultTTL      = 15 * time.Minute
	idempotencyLeaseRenewEvery = idempotencyDefaultTTL / 3
	idempotencyRedisOpTimeout  = 2 * time.Second
	idempotencyDefaultWaitTime = 20 * time.Second
	idempotencyDefaultPoll     = 100 * time.Millisecond

	idempotencyMaxKeyBytes      = 255
	idempotencyMaxRequestBytes  = 1 << 20
	idempotencyMaxResponseBytes = 256 << 10

	idempotencyKeyContext         = "commerce_idempotency_key"
	idempotencyRequestHashContext = "commerce_idempotency_request_hash"
)

var (
	errIdempotencyBodyTooLarge = errors.New("idempotency request body too large")

	acquireIdempotencyScript = redis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 1 or redis.call("EXISTS", KEYS[2]) == 1 then
	return 0
end
redis.call("SET", KEYS[2], ARGV[1], "PX", ARGV[3])
redis.call("SET", KEYS[1], ARGV[2], "PX", ARGV[3])
return 1
`)

	completeIdempotencyScript = redis.NewScript(`
if redis.call("GET", KEYS[2]) ~= ARGV[1] then
	return 0
end
redis.call("SET", KEYS[1], ARGV[2], "PX", ARGV[3])
redis.call("DEL", KEYS[2])
return 1
`)

	releaseIdempotencyScript = redis.NewScript(`
if redis.call("GET", KEYS[2]) ~= ARGV[1] then
	return 0
end
redis.call("DEL", KEYS[1])
redis.call("DEL", KEYS[2])
return 1
`)

	renewIdempotencyLeaseScript = redis.NewScript(`
if redis.call("EXISTS", KEYS[1]) ~= 1 then
	return 0
end
if redis.call("GET", KEYS[2]) ~= ARGV[1] then
	return 0
end
redis.call("PEXPIRE", KEYS[1], ARGV[2])
redis.call("PEXPIRE", KEYS[2], ARGV[2])
return 1
`)
)

type idempotencyRecord struct {
	State       string `json:"state"`
	RequestHash string `json:"request_hash"`
	StatusCode  int    `json:"status_code"`
	ContentType string `json:"content_type,omitempty"`
	Body        string `json:"body,omitempty"`
	SavedAt     int64  `json:"saved_at,omitempty"`
}

type idempotencyBodyWriter struct {
	gin.ResponseWriter
	body     bytes.Buffer
	tooLarge bool
}

func (w *idempotencyBodyWriter) Write(data []byte) (int, error) {
	w.capture(data)
	return w.ResponseWriter.Write(data)
}

func (w *idempotencyBodyWriter) WriteString(value string) (int, error) {
	if value != "" {
		w.capture([]byte(value))
	}
	return w.ResponseWriter.WriteString(value)
}

func (w *idempotencyBodyWriter) capture(data []byte) {
	if len(data) == 0 || w.tooLarge {
		return
	}
	remaining := idempotencyMaxResponseBytes - w.body.Len()
	if remaining <= 0 {
		w.tooLarge = true
		return
	}
	if len(data) > remaining {
		_, _ = w.body.Write(data[:remaining])
		w.tooLarge = true
		return
	}
	_, _ = w.body.Write(data)
}

// Idempotency replays the first completed response for the same authenticated
// request key and request payload. It is intended for write endpoints only.
func Idempotency(redisClient redis.UniversalClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		if redisClient == nil {
			respondIdempotencyUnavailable(c)
			c.Abort()
			return
		}

		idempotencyKey := strings.TrimSpace(c.GetHeader(idempotencyKeyHeader))
		if idempotencyKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "idempotency_key_required",
				"message": "Idempotency-Key header is required",
			})
			c.Abort()
			return
		}
		if len([]byte(idempotencyKey)) > idempotencyMaxKeyBytes {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "idempotency_key_too_large",
				"message": "Idempotency-Key header is too large",
			})
			c.Abort()
			return
		}

		body, err := readRequestBody(c)
		if err != nil {
			if errors.Is(err, errIdempotencyBodyTooLarge) {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{
					"error":   "idempotency_body_too_large",
					"message": "Request body is too large for idempotency protection",
				})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "idempotency_body_unreadable",
					"message": "Unable to read request body",
				})
			}
			c.Abort()
			return
		}

		requestPath := idempotencyPath(c)
		requestHash := requestIdempotencyHash(c.Request.Method, requestPath, c.Request.URL.RawQuery, body)
		c.Set(idempotencyKeyContext, idempotencyKey)
		c.Set(idempotencyRequestHashContext, requestHash)

		redisKey := idempotencyRedisKey(c, idempotencyKey)
		ownerKey := idempotencyOwnerRedisKey(redisKey)
		ownerToken := newIdempotencyOwnerToken()
		ctx := c.Request.Context()

		acquired, err := acquireIdempotency(ctx, redisClient, redisKey, ownerKey, ownerToken, requestHash)
		if err != nil {
			respondIdempotencyUnavailable(c)
			c.Abort()
			return
		}
		if !acquired {
			record, found, err := loadIdempotencyRecord(ctx, redisClient, redisKey)
			if err != nil {
				respondIdempotencyUnavailable(c)
				c.Abort()
				return
			}
			if found {
				if record.RequestHash != requestHash {
					c.JSON(http.StatusConflict, gin.H{
						"error":   "idempotency_key_conflict",
						"message": "Idempotency-Key was already used for a different request payload",
					})
					c.Abort()
					return
				}
				if record.State == idempotencyStateCompleted {
					replayIdempotencyRecord(c, record)
					c.Abort()
					return
				}
				waitRecord, ok, err := waitForIdempotencyRecord(ctx, redisClient, redisKey, requestHash)
				if err != nil {
					respondIdempotencyUnavailable(c)
					c.Abort()
					return
				}
				if ok {
					replayIdempotencyRecord(c, waitRecord)
					c.Abort()
					return
				}
			}

			c.Header("Retry-After", "1")
			c.JSON(http.StatusConflict, gin.H{
				"error":   "idempotency_request_in_progress",
				"message": "The same request is already being processed",
			})
			c.Abort()
			return
		}

		writer := &idempotencyBodyWriter{ResponseWriter: c.Writer}
		c.Writer = writer
		handlerCtx, cancelHandler := context.WithCancel(ctx)
		c.Request = c.Request.WithContext(handlerCtx)
		stopLease, leaseDone := startIdempotencyLeaseHeartbeat(
			redisClient,
			redisKey,
			ownerKey,
			ownerToken,
			cancelHandler,
		)
		defer stopIdempotencyLeaseHeartbeat(stopLease, leaseDone)
		defer cancelHandler()

		defer func() {
			if recovered := recover(); recovered != nil {
				releaseCtx, cancel := idempotencyOperationContext()
				_ = releaseIdempotency(releaseCtx, redisClient, redisKey, ownerKey, ownerToken)
				cancel()
				panic(recovered)
			}
		}()

		c.Next()

		statusCode := c.Writer.Status()
		if !shouldCacheIdempotencyResponse(statusCode) || writer.tooLarge {
			stopIdempotencyLeaseHeartbeat(stopLease, leaseDone)
			releaseCtx, cancel := idempotencyOperationContext()
			_ = releaseIdempotency(releaseCtx, redisClient, redisKey, ownerKey, ownerToken)
			cancel()
			return
		}

		completed := idempotencyRecord{
			State:       idempotencyStateCompleted,
			RequestHash: requestHash,
			StatusCode:  statusCode,
			ContentType: strings.TrimSpace(c.Writer.Header().Get("Content-Type")),
			Body:        writer.body.String(),
			SavedAt:     time.Now().UTC().Unix(),
		}
		stopIdempotencyLeaseHeartbeat(stopLease, leaseDone)
		completeCtx, cancel := idempotencyOperationContext()
		cached, err := completeIdempotency(completeCtx, redisClient, redisKey, ownerKey, ownerToken, completed)
		cancel()
		if err != nil || !cached {
			logger.Error(
				"failed to cache completed idempotency response",
				zap.Error(err),
				zap.Bool("owner_matched", cached),
				zap.String("idempotency_redis_key", redisKey),
			)
		}
	}
}

func GetIdempotencyKey(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if value, exists := c.Get(idempotencyKeyContext); exists {
		if text, ok := value.(string); ok {
			return strings.TrimSpace(text)
		}
	}
	return strings.TrimSpace(c.GetHeader(idempotencyKeyHeader))
}

func GetIdempotencyRequestHash(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if value, exists := c.Get(idempotencyRequestHashContext); exists {
		if text, ok := value.(string); ok {
			return strings.TrimSpace(text)
		}
	}
	return ""
}

func readRequestBody(c *gin.Context) ([]byte, error) {
	if c.Request == nil || c.Request.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, idempotencyMaxRequestBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > idempotencyMaxRequestBytes {
		return nil, errIdempotencyBodyTooLarge
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func requestIdempotencyHash(method, requestPath, rawQuery string, body []byte) string {
	h := sha256.New()
	_, _ = h.Write([]byte(strings.ToUpper(strings.TrimSpace(method))))
	_, _ = h.Write([]byte("\n"))
	_, _ = h.Write([]byte(strings.TrimSpace(requestPath)))
	_, _ = h.Write([]byte("\n"))
	_, _ = h.Write([]byte(strings.TrimSpace(rawQuery)))
	_, _ = h.Write([]byte("\n"))
	_, _ = h.Write(normalizeIdempotencyRequestBody(body))
	return hex.EncodeToString(h.Sum(nil))
}

func normalizeIdempotencyRequestBody(body []byte) []byte {
	if len(body) == 0 {
		return body
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return body
	}
	var trailingValue any
	if err := decoder.Decode(&trailingValue); err != io.EOF {
		return body
	}
	normalized, err := json.Marshal(value)
	if err != nil {
		return body
	}
	return normalized
}

func idempotencyRedisKey(c *gin.Context, key string) string {
	requestPath := idempotencyPath(c)
	identity := idempotencyIdentity(c)
	sum := sha256.Sum256([]byte(strings.Join([]string{
		idempotencyNamespace,
		strings.ToUpper(strings.TrimSpace(c.Request.Method)),
		requestPath,
		identity,
		strings.TrimSpace(key),
	}, "\n")))
	// The record and owner keys are used together by Lua scripts. Keep the
	// digest inside a Redis Cluster hash tag so both keys share one slot.
	return idempotencyNamespace + ":{" + hex.EncodeToString(sum[:]) + "}"
}

func idempotencyPath(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	if value := strings.TrimSpace(c.Request.URL.Path); value != "" {
		return value
	}
	return strings.TrimSpace(c.FullPath())
}

func idempotencyOwnerRedisKey(redisKey string) string {
	return redisKey + ":owner"
}

func idempotencyIdentity(c *gin.Context) string {
	if value, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", value)
	}
	if sessionID, err := c.Cookie("session_id"); err == nil && strings.TrimSpace(sessionID) != "" {
		return "session:" + strings.TrimSpace(sessionID)
	}
	if ip := strings.TrimSpace(c.ClientIP()); ip != "" {
		return "ip:" + ip
	}
	return "anonymous"
}

func newIdempotencyOwnerToken() string {
	var token [16]byte
	if _, err := rand.Read(token[:]); err == nil {
		return hex.EncodeToString(token[:])
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(sum[:])
}

func acquireIdempotency(ctx context.Context, redisClient redis.UniversalClient, key, ownerKey, ownerToken, requestHash string) (bool, error) {
	pending := idempotencyRecord{
		State:       idempotencyStatePending,
		RequestHash: requestHash,
		SavedAt:     time.Now().UTC().Unix(),
	}
	result, err := acquireIdempotencyScript.Run(
		ctx,
		redisClient,
		[]string{key, ownerKey},
		ownerToken,
		mustJSON(pending),
		int64(idempotencyDefaultTTL/time.Millisecond),
	).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func completeIdempotency(ctx context.Context, redisClient redis.UniversalClient, key, ownerKey, ownerToken string, record idempotencyRecord) (bool, error) {
	result, err := completeIdempotencyScript.Run(
		ctx,
		redisClient,
		[]string{key, ownerKey},
		ownerToken,
		mustJSON(record),
		int64(idempotencyDefaultTTL/time.Millisecond),
	).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func releaseIdempotency(ctx context.Context, redisClient redis.UniversalClient, key, ownerKey, ownerToken string) error {
	_, err := releaseIdempotencyScript.Run(
		ctx,
		redisClient,
		[]string{key, ownerKey},
		ownerToken,
	).Result()
	return err
}

func renewIdempotencyLease(ctx context.Context, redisClient redis.UniversalClient, key, ownerKey, ownerToken string, ttl time.Duration) (bool, error) {
	result, err := renewIdempotencyLeaseScript.Run(
		ctx,
		redisClient,
		[]string{key, ownerKey},
		ownerToken,
		int64(ttl/time.Millisecond),
	).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func startIdempotencyLeaseHeartbeat(
	redisClient redis.UniversalClient,
	key,
	ownerKey,
	ownerToken string,
	onLeaseLost context.CancelFunc,
) (context.CancelFunc, <-chan struct{}) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)

		ticker := time.NewTicker(idempotencyLeaseRenewEvery)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				renewCtx, renewCancel := idempotencyOperationContextWithParent(ctx)
				renewed, err := renewIdempotencyLease(renewCtx, redisClient, key, ownerKey, ownerToken, idempotencyDefaultTTL)
				renewCancel()
				if err != nil {
					logger.Warn(
						"failed to renew idempotency owner lease",
						zap.Error(err),
						zap.String("idempotency_redis_key", key),
					)
					continue
				}
				if !renewed {
					logger.Warn(
						"idempotency owner lease was lost before request completed",
						zap.String("idempotency_redis_key", key),
					)
					if onLeaseLost != nil {
						onLeaseLost()
					}
					return
				}
			}
		}
	}()
	return cancel, done
}

func stopIdempotencyLeaseHeartbeat(cancel context.CancelFunc, done <-chan struct{}) {
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}

func idempotencyOperationContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), idempotencyRedisOpTimeout)
}

func idempotencyOperationContextWithParent(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, idempotencyRedisOpTimeout)
}

func loadIdempotencyRecord(ctx context.Context, redisClient redis.UniversalClient, key string) (idempotencyRecord, bool, error) {
	value, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return idempotencyRecord{}, false, nil
		}
		return idempotencyRecord{}, false, err
	}

	var record idempotencyRecord
	if err := json.Unmarshal([]byte(value), &record); err != nil {
		return idempotencyRecord{}, false, err
	}
	return record, true, nil
}

func waitForIdempotencyRecord(ctx context.Context, redisClient redis.UniversalClient, key, requestHash string) (idempotencyRecord, bool, error) {
	deadline := time.NewTimer(idempotencyDefaultWaitTime)
	defer deadline.Stop()

	ticker := time.NewTicker(idempotencyDefaultPoll)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return idempotencyRecord{}, false, nil
		case <-deadline.C:
			return idempotencyRecord{}, false, nil
		case <-ticker.C:
			record, found, err := loadIdempotencyRecord(ctx, redisClient, key)
			if err != nil {
				return idempotencyRecord{}, false, err
			}
			if !found {
				return idempotencyRecord{}, false, nil
			}
			if record.RequestHash != requestHash {
				return idempotencyRecord{}, false, nil
			}
			if record.State == idempotencyStateCompleted {
				return record, true, nil
			}
		}
	}
}

func replayIdempotencyRecord(c *gin.Context, record idempotencyRecord) {
	if record.ContentType != "" {
		c.Header("Content-Type", record.ContentType)
	}
	c.Header(idempotencyReplayHeader, "true")
	c.Status(record.StatusCode)
	if record.Body != "" {
		_, _ = c.Writer.Write([]byte(record.Body))
	}
}

func shouldCacheIdempotencyResponse(statusCode int) bool {
	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return true
	}
	switch statusCode {
	case http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusUnprocessableEntity:
		return true
	default:
		return false
	}
}

func respondIdempotencyUnavailable(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"error":   "idempotency_service_unavailable",
		"message": "Idempotency service is temporarily unavailable",
	})
}

func mustJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}
