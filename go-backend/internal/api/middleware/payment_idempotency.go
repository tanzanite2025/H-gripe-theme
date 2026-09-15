package middleware

import (
	"commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/logger"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PaymentOperationIdempotencyStore is the durable database fence used by
// provider-facing payment operations. Redis remains an acceleration layer,
// while this store protects retries after Redis loss or eviction.
type PaymentOperationIdempotencyStore interface {
	TryCreate(*payment.PaymentOperationIdempotency) (bool, error)
	FindByUserScopeKey(userID uint, scope, key string) (*payment.PaymentOperationIdempotency, error)
	Complete(id uint, statusCode int, contentType, responseBody string) error
	Delete(id uint) error
}

// PaymentOperationIdempotency persists the request and response around the
// existing Redis idempotency middleware. Put it before Idempotency in a route
// chain so a database record is claimed even when Redis is unavailable.
func PaymentOperationIdempotency(store PaymentOperationIdempotencyStore, scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if store == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "idempotency_service_unavailable",
				"message": "Durable payment idempotency service is unavailable",
			})
			c.Abort()
			return
		}

		key := strings.TrimSpace(c.GetHeader(idempotencyKeyHeader))
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "idempotency_key_required",
				"message": "Idempotency-Key header is required",
			})
			c.Abort()
			return
		}
		if len([]byte(key)) > idempotencyMaxKeyBytes {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "idempotency_key_too_large",
				"message": "Idempotency-Key header is too large",
			})
			c.Abort()
			return
		}

		userID, ok := paymentIdempotencyUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication_required"})
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

		requestHash := requestIdempotencyHash(c.Request.Method, idempotencyPath(c), c.Request.URL.RawQuery, body)
		c.Set(idempotencyKeyContext, key)
		c.Set(idempotencyRequestHashContext, requestHash)
		c.Set(idempotencyDurableFallbackContext, true)
		record := &payment.PaymentOperationIdempotency{
			UserID:         userID,
			Scope:          strings.TrimSpace(scope),
			IdempotencyKey: key,
			RequestHash:    requestHash,
			Status:         payment.PaymentOperationIdempotencyPending,
		}
		claimed, err := store.TryCreate(record)
		if err != nil {
			logger.Error("failed to claim durable payment idempotency key", zap.Error(err), zap.String("scope", scope))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "idempotency_service_unavailable",
				"message": "Durable payment idempotency service is temporarily unavailable",
			})
			c.Abort()
			return
		}
		if !claimed {
			record, err = store.FindByUserScopeKey(userID, scope, key)
			if err != nil {
				logger.Error("failed to read durable payment idempotency key", zap.Error(err), zap.String("scope", scope))
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error":   "idempotency_service_unavailable",
					"message": "Durable payment idempotency service is temporarily unavailable",
				})
				c.Abort()
				return
			}
			if record.RequestHash != requestHash {
				c.JSON(http.StatusConflict, gin.H{
					"error":   "idempotency_key_conflict",
					"message": "Idempotency-Key was already used for a different request payload",
				})
				c.Abort()
				return
			}
			if record.Status == payment.PaymentOperationIdempotencyCompleted {
				replayDurablePaymentIdempotencyRecord(c, record)
				c.Abort()
				return
			}
			c.Header("Retry-After", "1")
			c.JSON(http.StatusConflict, gin.H{
				"error":   "idempotency_request_in_progress",
				"message": "The same payment request is already being processed",
			})
			c.Abort()
			return
		}

		writer := &idempotencyBodyWriter{ResponseWriter: c.Writer}
		c.Writer = writer
		c.Next()

		statusCode := c.Writer.Status()
		if !shouldCacheIdempotencyResponse(statusCode) || writer.tooLarge {
			// Keep the pending claim for provider-facing operations. A gateway
			// call may already have succeeded even when the HTTP response is a
			// 5xx or the response body could not be cached; releasing the row
			// would allow a retry to perform the provider operation twice.
			logger.Warn("keeping durable payment idempotency key pending after non-replayable response", zap.Int("status_code", statusCode), zap.Uint("record_id", record.ID))
			return
		}
		if err := store.Complete(
			record.ID,
			statusCode,
			strings.TrimSpace(c.Writer.Header().Get("Content-Type")),
			writer.body.String(),
		); err != nil {
			// Keep a pending record when completion cannot be persisted. A retry
			// must stop rather than risk repeating a provider-side operation.
			logger.Error("failed to complete durable payment idempotency key", zap.Error(err), zap.Uint("record_id", record.ID))
		}
	}
}

func replayDurablePaymentIdempotencyRecord(c *gin.Context, record *payment.PaymentOperationIdempotency) {
	if record.ContentType != "" {
		c.Header("Content-Type", record.ContentType)
	}
	c.Header(idempotencyReplayHeader, "true")
	c.Status(record.StatusCode)
	if record.ResponseBody != "" {
		_, _ = c.Writer.Write([]byte(record.ResponseBody))
	}
}

func paymentIdempotencyUserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	switch id := value.(type) {
	case uint:
		return id, id > 0
	case uint64:
		return uint(id), id > 0 && uint64(uint(id)) == id
	case int:
		return uint(id), id > 0
	case int64:
		return uint(id), id > 0 && int64(uint(id)) == id
	default:
		return 0, false
	}
}
