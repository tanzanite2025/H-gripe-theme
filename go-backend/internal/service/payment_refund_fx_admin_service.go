package service

import (
	"errors"
	"strings"
	"time"

	currencydomain "commerce-platform/internal/domain/currency"
	"commerce-platform/internal/repository"
)

var (
	ErrHistoricalRefundFXSnapshotNotApplicable    = errors.New("historical refund FX snapshot backfill is only for non-USD orders")
	ErrHistoricalRefundFXSnapshotAlreadyPresent   = errors.New("historical refund FX snapshot is already present")
	ErrHistoricalRefundFXSnapshotAdminUnavailable = errors.New("historical refund FX snapshot service is unavailable")
)

type BackfillHistoricalRefundFXSnapshotInput struct {
	OrderID       uint
	BaseCurrency  string
	OrderCurrency string
	RateDecimal   string
	Source        string
	CapturedAt    time.Time
	RateFetchedAt *time.Time
}

type BackfillHistoricalRefundFXSnapshotResult struct {
	OrderID     uint                           `json:"order_id"`
	OrderNumber string                         `json:"order_number"`
	Snapshot    currencydomain.OrderFXSnapshot `json:"fx_snapshot"`
}

// BackfillHistoricalRefundFXSnapshot repairs an old order whose immutable
// order-time FX contract was not captured. It never replaces a valid snapshot.
func (s *PaymentService) BackfillHistoricalRefundFXSnapshot(input BackfillHistoricalRefundFXSnapshotInput) (*BackfillHistoricalRefundFXSnapshotResult, error) {
	if s == nil || s.txManager == nil {
		return nil, ErrHistoricalRefundFXSnapshotAdminUnavailable
	}
	if input.OrderID == 0 {
		return nil, errors.New("order id is required")
	}
	if input.CapturedAt.IsZero() {
		return nil, errors.New("captured_at is required")
	}
	baseCode, err := currencydomain.ParseCode(input.BaseCurrency)
	if err != nil {
		return nil, errors.New("base_currency is invalid")
	}
	orderCode, err := currencydomain.ParseCode(input.OrderCurrency)
	if err != nil {
		return nil, errors.New("order_currency is invalid")
	}
	snapshot := currencydomain.OrderFXSnapshot{
		Version:       currencydomain.OrderFXSnapshotVersion,
		BaseCurrency:  baseCode.String(),
		OrderCurrency: orderCode.String(),
		RateDecimal:   strings.TrimSpace(input.RateDecimal),
		Source:        strings.TrimSpace(input.Source),
		CapturedAt:    input.CapturedAt.UTC(),
		RateFetchedAt: normalizeFXTime(input.RateFetchedAt),
	}
	if err := snapshot.Validate(orderCode.String()); err != nil {
		return nil, err
	}
	if snapshot.OrderCurrency == currencydomain.DefaultPrimaryCurrency {
		return nil, ErrHistoricalRefundFXSnapshotNotApplicable
	}

	var result *BackfillHistoricalRefundFXSnapshotResult
	err = s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.Order == nil {
			return ErrHistoricalRefundFXSnapshotAdminUnavailable
		}
		orderRecord, findErr := repos.Order.FindByIDForUpdate(input.OrderID)
		if findErr != nil {
			return findErr
		}
		if strings.EqualFold(strings.TrimSpace(orderRecord.Currency), currencydomain.DefaultPrimaryCurrency) {
			return ErrHistoricalRefundFXSnapshotNotApplicable
		}
		if !strings.EqualFold(strings.TrimSpace(orderRecord.Currency), snapshot.OrderCurrency) {
			return errors.New("order_currency does not match the order currency")
		}
		if existing, parseErr := currencydomain.ParseOrderFXSnapshot(orderRecord.FXSnapshotData); parseErr == nil {
			if validateErr := existing.Validate(orderRecord.Currency); validateErr == nil {
				return ErrHistoricalRefundFXSnapshotAlreadyPresent
			}
		}
		if updateErr := repos.Order.UpdateFXSnapshot(orderRecord.ID, currencydomain.OrderFXSnapshotJSON(snapshot)); updateErr != nil {
			return updateErr
		}
		result = &BackfillHistoricalRefundFXSnapshotResult{
			OrderID: orderRecord.ID, OrderNumber: orderRecord.OrderNumber, Snapshot: snapshot,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func normalizeFXTime(value *time.Time) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}
