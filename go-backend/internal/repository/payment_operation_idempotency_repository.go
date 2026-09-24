package repository

import (
	"commerce-platform/internal/domain/payment"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrPaymentOperationIdempotencyClaimLost = errors.New("payment operation idempotency claim was lost")

type PaymentOperationIdempotencyRepository struct {
	db *gorm.DB
}

func NewPaymentOperationIdempotencyRepository(db *gorm.DB) *PaymentOperationIdempotencyRepository {
	return &PaymentOperationIdempotencyRepository{db: db}
}

// TryCreate claims the operation key using the database unique constraint.
func (r *PaymentOperationIdempotencyRepository) TryCreate(record *payment.PaymentOperationIdempotency) (bool, error) {
	if record == nil {
		return false, fmt.Errorf("payment operation idempotency record is required")
	}
	result := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(record)
	return result.RowsAffected == 1, result.Error
}

func (r *PaymentOperationIdempotencyRepository) FindByUserScopeKey(userID uint, scope, key string) (*payment.PaymentOperationIdempotency, error) {
	var record payment.PaymentOperationIdempotency
	if err := r.db.Where("user_id = ? AND scope = ? AND idempotency_key = ?", userID, scope, key).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// ReclaimExpiredQuery atomically takes an expired pending claim for an
// operation whose provider interaction is read-only.
func (r *PaymentOperationIdempotencyRepository) ReclaimExpiredQuery(
	id uint,
	requestHash string,
	claimToken string,
	now time.Time,
	leaseExpiresAt time.Time,
) (bool, error) {
	result := r.db.Model(&payment.PaymentOperationIdempotency{}).
		Where(
			"id = ? AND request_hash = ? AND status = ? AND (lease_expires_at IS NULL OR lease_expires_at <= ?)",
			id,
			strings.TrimSpace(requestHash),
			payment.PaymentOperationIdempotencyPending,
			now.UTC(),
		).
		Updates(map[string]interface{}{
			"claim_token":      strings.TrimSpace(claimToken),
			"lease_expires_at": leaseExpiresAt.UTC(),
			"updated_at":       now.UTC(),
		})
	return result.RowsAffected == 1, result.Error
}

// ClaimReconciliation atomically turns an expired external mutation into a
// reconciliation lease. Existing reconciliation leases can also be resumed
// after expiry, but the original mutation is never reclaimed.
func (r *PaymentOperationIdempotencyRepository) ClaimReconciliation(
	id uint,
	requestHash string,
	claimToken string,
	now time.Time,
	leaseExpiresAt time.Time,
) (bool, error) {
	now = now.UTC()
	result := r.db.Model(&payment.PaymentOperationIdempotency{}).
		Where(
			"id = ? AND request_hash = ? AND status IN ? AND (lease_expires_at IS NULL OR lease_expires_at <= ?)",
			id,
			strings.TrimSpace(requestHash),
			[]string{
				payment.PaymentOperationIdempotencyPending,
				payment.PaymentOperationIdempotencyReconciling,
			},
			now,
		).
		Updates(map[string]interface{}{
			"status":                    payment.PaymentOperationIdempotencyReconciling,
			"claim_token":               strings.TrimSpace(claimToken),
			"lease_expires_at":          leaseExpiresAt.UTC(),
			"reconciliation_started_at": gorm.Expr("COALESCE(reconciliation_started_at, ?)", now),
			"updated_at":                now,
		})
	return result.RowsAffected == 1, result.Error
}

func (r *PaymentOperationIdempotencyRepository) Complete(id uint, claimToken string, statusCode int, contentType, responseBody string) error {
	result := r.db.Model(&payment.PaymentOperationIdempotency{}).
		Where(
			"id = ? AND claim_token = ? AND status IN ?",
			id,
			strings.TrimSpace(claimToken),
			[]string{
				payment.PaymentOperationIdempotencyPending,
				payment.PaymentOperationIdempotencyReconciling,
			},
		).
		Updates(map[string]interface{}{
			"status":           payment.PaymentOperationIdempotencyCompleted,
			"status_code":      statusCode,
			"content_type":     contentType,
			"response_body":    responseBody,
			"claim_token":      "",
			"lease_expires_at": nil,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("%w: record %d", ErrPaymentOperationIdempotencyClaimLost, id)
	}
	return nil
}

// RequireReconciliation releases the current mutation claim into the
// reconciling state after an external call has an uncertain local outcome.
func (r *PaymentOperationIdempotencyRepository) RequireReconciliation(id uint, claimToken string, now time.Time) error {
	now = now.UTC()
	result := r.db.Model(&payment.PaymentOperationIdempotency{}).
		Where(
			"id = ? AND claim_token = ? AND status = ?",
			id,
			strings.TrimSpace(claimToken),
			payment.PaymentOperationIdempotencyPending,
		).
		Updates(map[string]interface{}{
			"status":                    payment.PaymentOperationIdempotencyReconciling,
			"claim_token":               "",
			"lease_expires_at":          now,
			"reconciliation_started_at": gorm.Expr("COALESCE(reconciliation_started_at, ?)", now),
			"updated_at":                now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("%w: record %d", ErrPaymentOperationIdempotencyClaimLost, id)
	}
	return nil
}

// ReleaseClaim keeps the request hash fence while making the same operation
// immediately claimable again. It is used only after read-only provider work,
// or before an external mutation has started.
func (r *PaymentOperationIdempotencyRepository) ReleaseClaim(id uint, claimToken string, now time.Time) error {
	now = now.UTC()
	result := r.db.Model(&payment.PaymentOperationIdempotency{}).
		Where(
			"id = ? AND claim_token = ? AND status IN ?",
			id,
			strings.TrimSpace(claimToken),
			[]string{
				payment.PaymentOperationIdempotencyPending,
				payment.PaymentOperationIdempotencyReconciling,
			},
		).
		Updates(map[string]interface{}{
			"claim_token":      "",
			"lease_expires_at": now,
			"updated_at":       now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("%w: record %d", ErrPaymentOperationIdempotencyClaimLost, id)
	}
	return nil
}
