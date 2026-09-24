package middleware

import (
	"commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/logger"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	paymentOperationLeaseDuration = time.Minute

	paymentOperationReconciliationContext = "commerce_payment_operation_reconciliation"
	paymentOperationExternalCallContext   = "commerce_payment_operation_external_call_started"
)

type PaymentOperationRecoveryPolicy uint8

const (
	// PaymentOperationRetryExpiredQuery is for provider reads. An expired
	// owner can be replaced because executing the provider query again has no
	// external financial side effect.
	PaymentOperationRetryExpiredQuery PaymentOperationRecoveryPolicy = iota + 1
	// PaymentOperationReconcileExpiredMutation is for provider writes. An
	// expired owner moves to reconciliation and may only execute provider reads.
	PaymentOperationReconcileExpiredMutation
)

// PaymentOperationIdempotencyStore is the durable database fence used by
// provider-facing payment operations. Redis remains an acceleration layer,
// while this store protects retries after Redis loss or eviction.
type PaymentOperationIdempotencyStore interface {
	TryCreate(*payment.PaymentOperationIdempotency) (bool, error)
	FindByUserScopeKey(userID uint, scope, key string) (*payment.PaymentOperationIdempotency, error)
	ReclaimExpiredQuery(id uint, requestHash, claimToken string, now, leaseExpiresAt time.Time) (bool, error)
	ClaimReconciliation(id uint, requestHash, claimToken string, now, leaseExpiresAt time.Time) (bool, error)
	Complete(id uint, claimToken string, statusCode int, contentType, responseBody string) error
	RequireReconciliation(id uint, claimToken string, now time.Time) error
	ReleaseClaim(id uint, claimToken string, now time.Time) error
}

// PaymentOperationIdempotency is the sole correctness boundary for a
// provider-facing payment endpoint. Query operations can reclaim an expired
// lease. Mutation operations instead enter reconciliation, and the handler
// must use IsPaymentOperationReconciliation to perform provider reads only.
func PaymentOperationIdempotency(
	store PaymentOperationIdempotencyStore,
	scope string,
	recoveryPolicy PaymentOperationRecoveryPolicy,
) gin.HandlerFunc {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		panic("payment operation idempotency scope is required")
	}
	if recoveryPolicy != PaymentOperationRetryExpiredQuery && recoveryPolicy != PaymentOperationReconcileExpiredMutation {
		panic(fmt.Sprintf("invalid payment operation recovery policy %d", recoveryPolicy))
	}

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

		now := time.Now().UTC()
		leaseExpiresAt := now.Add(paymentOperationLeaseDuration)
		claimToken := uuid.NewString()
		record := &payment.PaymentOperationIdempotency{
			UserID:         userID,
			Scope:          scope,
			IdempotencyKey: key,
			RequestHash:    requestHash,
			Status:         payment.PaymentOperationIdempotencyPending,
			ClaimToken:     claimToken,
			LeaseExpiresAt: &leaseExpiresAt,
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
		claimStatus := payment.PaymentOperationIdempotencyPending
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

			claimed, err = reclaimPaymentOperation(
				store,
				record,
				recoveryPolicy,
				requestHash,
				claimToken,
				now,
				leaseExpiresAt,
			)
			if err != nil {
				logger.Error("failed to reclaim durable payment idempotency key", zap.Error(err), zap.String("scope", scope))
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error":   "idempotency_service_unavailable",
					"message": "Durable payment idempotency service is temporarily unavailable",
				})
				c.Abort()
				return
			}
			if !claimed {
				latest, findErr := store.FindByUserScopeKey(userID, scope, key)
				if findErr != nil {
					logger.Error("failed to refresh durable payment idempotency key", zap.Error(findErr), zap.String("scope", scope))
					c.JSON(http.StatusServiceUnavailable, gin.H{
						"error":   "idempotency_service_unavailable",
						"message": "Durable payment idempotency service is temporarily unavailable",
					})
					c.Abort()
					return
				}
				if latest.Status == payment.PaymentOperationIdempotencyCompleted {
					replayDurablePaymentIdempotencyRecord(c, latest)
					c.Abort()
					return
				}
				respondPaymentOperationClaimBusy(c, latest.Status)
				return
			}
			if recoveryPolicy == PaymentOperationReconcileExpiredMutation {
				claimStatus = payment.PaymentOperationIdempotencyReconciling
				c.Set(paymentOperationReconciliationContext, true)
			}
		}

		writer := &idempotencyBodyWriter{ResponseWriter: c.Writer}
		c.Writer = writer
		c.Next()

		statusCode := c.Writer.Status()
		responseSucceeded := statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
		if claimStatus == payment.PaymentOperationIdempotencyReconciling && (!responseSucceeded || writer.tooLarge) {
			if err := store.ReleaseClaim(record.ID, claimToken, time.Now().UTC()); err != nil {
				logger.Error("failed to release payment reconciliation claim", zap.Error(err), zap.Uint("record_id", record.ID))
			}
			return
		}
		if recoveryPolicy == PaymentOperationReconcileExpiredMutation &&
			paymentOperationExternalCallStarted(c) &&
			(!responseSucceeded || writer.tooLarge) {
			if err := store.RequireReconciliation(record.ID, claimToken, time.Now().UTC()); err != nil {
				logger.Error("failed to mark payment operation for reconciliation", zap.Error(err), zap.Uint("record_id", record.ID))
			}
			return
		}
		if !shouldCacheIdempotencyResponse(statusCode) || writer.tooLarge {
			if err := store.ReleaseClaim(record.ID, claimToken, time.Now().UTC()); err != nil {
				logger.Error("failed to release durable payment idempotency claim", zap.Error(err), zap.Uint("record_id", record.ID))
			}
			return
		}
		if err := store.Complete(
			record.ID,
			claimToken,
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

func reclaimPaymentOperation(
	store PaymentOperationIdempotencyStore,
	record *payment.PaymentOperationIdempotency,
	recoveryPolicy PaymentOperationRecoveryPolicy,
	requestHash string,
	claimToken string,
	now time.Time,
	leaseExpiresAt time.Time,
) (bool, error) {
	if record == nil {
		return false, errors.New("payment operation idempotency record is required")
	}
	switch recoveryPolicy {
	case PaymentOperationRetryExpiredQuery:
		if record.Status != payment.PaymentOperationIdempotencyPending {
			return false, nil
		}
		return store.ReclaimExpiredQuery(record.ID, requestHash, claimToken, now, leaseExpiresAt)
	case PaymentOperationReconcileExpiredMutation:
		return store.ClaimReconciliation(record.ID, requestHash, claimToken, now, leaseExpiresAt)
	default:
		return false, fmt.Errorf("invalid payment operation recovery policy %d", recoveryPolicy)
	}
}

func respondPaymentOperationClaimBusy(c *gin.Context, status string) {
	c.Header("Retry-After", "1")
	if status == payment.PaymentOperationIdempotencyReconciling {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "payment_operation_reconciliation_in_progress",
			"message": "The payment operation outcome is being reconciled with the provider",
		})
	} else {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "idempotency_request_in_progress",
			"message": "The same payment request is already being processed",
		})
	}
	c.Abort()
}

// IsPaymentOperationReconciliation tells a mutation handler that an earlier
// provider call has an unknown outcome. The handler must perform provider
// reads only and must not issue the mutation again.
func IsPaymentOperationReconciliation(c *gin.Context) bool {
	if c == nil {
		return false
	}
	value, exists := c.Get(paymentOperationReconciliationContext)
	reconciling, _ := value.(bool)
	return exists && reconciling
}

// MarkPaymentOperationExternalCallStarted marks the exact point after which a
// failed mutation response has an uncertain provider outcome.
func MarkPaymentOperationExternalCallStarted(c *gin.Context) {
	if c != nil {
		c.Set(paymentOperationExternalCallContext, true)
	}
}

func paymentOperationExternalCallStarted(c *gin.Context) bool {
	if c == nil {
		return false
	}
	value, exists := c.Get(paymentOperationExternalCallContext)
	started, _ := value.(bool)
	return exists && started
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
