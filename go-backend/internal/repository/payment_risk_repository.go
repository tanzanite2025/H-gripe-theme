package repository

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	paymentdomain "commerce-platform/internal/domain/payment"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentRiskRepository struct {
	db *gorm.DB
}

func NewPaymentRiskRepository(db *gorm.DB) *PaymentRiskRepository {
	return &PaymentRiskRepository{db: db}
}

func (r *PaymentRiskRepository) WithTx(tx *gorm.DB) *PaymentRiskRepository {
	return &PaymentRiskRepository{db: tx}
}

type PaymentRiskMetricCounts struct {
	SuccessfulPaymentCount  int64
	SuccessfulPaymentAmount float64
	DisputeCount            int64
	DisputeAmount           float64
	EarlyFraudWarningCount  int64
	RefundCount             int64
	RefundAmount            float64
	CheckoutAttemptCount    int64
	ThreeDSUpgradeCount     int64
	ThreeDSChallengeCount   int64
	ThreeDSExemptionCount   int64
}

type PaymentRiskSnapshotPersistResult struct {
	PreviousLevel     paymentdomain.PaymentRiskLevel
	LevelChanged      bool
	AlertEventCreated bool
}

// PaymentRiskLevelChangedPayload deliberately contains operational metrics
// only. Customer, order, payment and raw provider-event identifiers stay out
// of external alert delivery.
type PaymentRiskLevelChangedPayload struct {
	Provider               string                         `json:"provider"`
	PreviousLevel          paymentdomain.PaymentRiskLevel `json:"previous_level"`
	CurrentLevel           paymentdomain.PaymentRiskLevel `json:"current_level"`
	WindowDays             int                            `json:"window_days"`
	SuccessfulPaymentCount int64                          `json:"successful_payment_count"`
	DisputeCount           int64                          `json:"dispute_count"`
	EarlyFraudWarningCount int64                          `json:"early_fraud_warning_count"`
	RefundCount            int64                          `json:"refund_count"`
	CheckoutAttemptCount   int64                          `json:"checkout_attempt_count"`
	ThreeDSUpgradeCount    int64                          `json:"three_ds_upgrade_count"`
	ThreeDSChallengeCount  int64                          `json:"three_ds_challenge_count"`
	ThreeDSExemptionCount  int64                          `json:"three_ds_exemption_count"`
	DisputeActivityRate    float64                        `json:"dispute_activity_rate"`
	EarlyFraudWarningRate  float64                        `json:"early_fraud_warning_rate"`
	RefundRate             float64                        `json:"refund_rate"`
	ThreeDSUpgradeRate     float64                        `json:"three_ds_upgrade_rate"`
	Reasons                []string                       `json:"reasons"`
	RecommendedAction      string                         `json:"recommended_action"`
	ComputedAt             time.Time                      `json:"computed_at"`
}

func (r *PaymentRiskRepository) UpsertPaymentRiskEvent(event *paymentdomain.PaymentRiskEvent) error {
	var existing paymentdomain.PaymentRiskEvent
	err := r.db.Where(
		"provider = ? AND kind = ? AND external_reference = ?",
		event.Provider,
		event.Kind,
		event.ExternalReference,
	).First(&existing).Error
	if err == nil {
		event.ID = existing.ID
		return r.db.Model(&paymentdomain.PaymentRiskEvent{}).
			Where("id = ?", existing.ID).
			Updates(map[string]interface{}{
				"webhook_event_id":    event.WebhookEventID,
				"provider_payment_id": event.ProviderPaymentID,
				"payment_intent_id":   event.PaymentIntentID,
				"charge_id":           event.ChargeID,
				"order_id":            event.OrderID,
				"transaction_id":      event.TransactionID,
				"amount":              event.Amount,
				"currency":            event.Currency,
				"occurred_at":         event.OccurredAt,
				"payload":             event.Payload,
				"metadata_json":       event.MetadataJSON,
				"updated_at":          time.Now().UTC(),
			}).Error
	}
	if !IsRecordNotFound(err) {
		return err
	}
	return r.db.Create(event).Error
}

func (r *PaymentRiskRepository) UpsertPaymentRiskCheckoutDecision(decision *paymentdomain.PaymentRiskCheckoutDecision) error {
	if decision == nil {
		return gorm.ErrInvalidData
	}
	decision.Provider = strings.ToLower(strings.TrimSpace(decision.Provider))
	decision.ProviderPaymentID = strings.TrimSpace(decision.ProviderPaymentID)
	if decision.Provider == "" || decision.ProviderPaymentID == "" {
		return gorm.ErrInvalidData
	}

	var existing paymentdomain.PaymentRiskCheckoutDecision
	err := r.db.Where(
		"provider = ? AND provider_payment_id = ?",
		decision.Provider,
		decision.ProviderPaymentID,
	).First(&existing).Error
	if err == nil {
		decision.ID = existing.ID
		return r.db.Model(&paymentdomain.PaymentRiskCheckoutDecision{}).
			Where("id = ?", existing.ID).
			Updates(map[string]interface{}{
				"order_id":             decision.OrderID,
				"mode":                 decision.Mode,
				"strategy":             decision.Strategy,
				"exemption_candidate":  decision.ExemptionCandidate,
				"risk_level":           decision.RiskLevel,
				"risk_score":           decision.RiskScore,
				"portfolio_risk_level": decision.PortfolioRiskLevel,
				"reasons_json":         decision.ReasonsJSON,
				"amount":               decision.Amount,
				"currency":             decision.Currency,
				"occurred_at":          decision.OccurredAt,
			}).Error
	}
	if !IsRecordNotFound(err) {
		return err
	}
	return r.db.Create(decision).Error
}

// CreatePaymentRiskSnapshotWithAlert persists the evaluated snapshot, moves
// provider state forward, and conditionally creates one outbox event inside a
// single transaction. A repeated scheduler run at the same level creates no
// duplicate alert.
func (r *PaymentRiskRepository) CreatePaymentRiskSnapshotWithAlert(
	snapshot *paymentdomain.PaymentRiskSnapshot,
	alertsEnabled bool,
) (PaymentRiskSnapshotPersistResult, error) {
	if snapshot == nil {
		return PaymentRiskSnapshotPersistResult{}, gorm.ErrInvalidData
	}

	snapshot.Provider = strings.ToLower(strings.TrimSpace(snapshot.Provider))
	if snapshot.Provider == "" {
		return PaymentRiskSnapshotPersistResult{}, gorm.ErrInvalidData
	}
	if snapshot.Level == "" {
		snapshot.Level = paymentdomain.PaymentRiskLevelNormal
	}

	var result PaymentRiskSnapshotPersistResult
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(snapshot).Error; err != nil {
			return err
		}

		state, err := lockPaymentRiskAlertState(tx, snapshot.Provider)
		if err != nil {
			return err
		}

		previousLevel := state.CurrentLevel
		if previousLevel == "" {
			previousLevel = paymentdomain.PaymentRiskLevelNormal
		}
		result.PreviousLevel = previousLevel
		result.LevelChanged = previousLevel != snapshot.Level

		if err := tx.Model(&paymentdomain.PaymentRiskAlertState{}).
			Where("provider = ?", snapshot.Provider).
			Updates(map[string]interface{}{
				"current_level":       snapshot.Level,
				"current_snapshot_id": snapshot.ID,
				"updated_at":          snapshot.ComputedAt.UTC(),
			}).Error; err != nil {
			return err
		}

		if !alertsEnabled || !result.LevelChanged {
			return nil
		}

		event, err := newPaymentRiskLevelChangedOutboxEvent(previousLevel, snapshot)
		if err != nil {
			return err
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "event_key"}},
			DoNothing: true,
		}).Create(&event).Error; err != nil {
			return err
		}
		result.AlertEventCreated = event.ID != 0
		return nil
	})
	if err != nil {
		return PaymentRiskSnapshotPersistResult{}, err
	}
	return result, nil
}

func (r *PaymentRiskRepository) FindLatestPaymentRiskSnapshot(provider string) (*paymentdomain.PaymentRiskSnapshot, error) {
	var snapshot paymentdomain.PaymentRiskSnapshot
	err := r.db.Where("provider = ?", strings.ToLower(strings.TrimSpace(provider))).
		Order("computed_at DESC, id DESC").
		First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (r *PaymentRiskRepository) CountPaymentRiskMetrics(provider string, start, end time.Time) (PaymentRiskMetricCounts, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return PaymentRiskMetricCounts{}, gorm.ErrInvalidData
	}

	var counts PaymentRiskMetricCounts
	paymentMethod := provider
	successQuery := r.db.Model(&paymentdomain.Transaction{}).
		Where("LOWER(payment_method) = ? AND status = ?", paymentMethod, "completed").
		Where("COALESCE(completed_at, created_at) >= ? AND COALESCE(completed_at, created_at) < ?", start, end)
	if err := successQuery.Count(&counts.SuccessfulPaymentCount).Error; err != nil {
		return PaymentRiskMetricCounts{}, err
	}
	if err := successQuery.Select("COALESCE(SUM(amount), 0)").Scan(&counts.SuccessfulPaymentAmount).Error; err != nil {
		return PaymentRiskMetricCounts{}, err
	}

	if provider == string(paymentdomain.PaymentRiskProviderStripe) {
		disputeQuery := r.db.Model(&paymentdomain.StripeDispute{}).
			Where("created_at >= ? AND created_at < ?", start, end)
		if err := disputeQuery.Count(&counts.DisputeCount).Error; err != nil {
			return PaymentRiskMetricCounts{}, err
		}
		if err := disputeQuery.Select("COALESCE(SUM(amount), 0)").Scan(&counts.DisputeAmount).Error; err != nil {
			return PaymentRiskMetricCounts{}, err
		}
	}

	earlyFraudQuery := r.db.Model(&paymentdomain.PaymentRiskEvent{}).
		Where("provider = ? AND kind = ? AND occurred_at >= ? AND occurred_at < ?",
			provider, paymentdomain.PaymentRiskEventEarlyFraudWarning, start, end)
	if err := earlyFraudQuery.Count(&counts.EarlyFraudWarningCount).Error; err != nil {
		return PaymentRiskMetricCounts{}, err
	}
	var disputeEventCount int64
	disputeEventQuery := r.db.Model(&paymentdomain.PaymentRiskEvent{}).
		Where("provider = ? AND kind = ? AND occurred_at >= ? AND occurred_at < ?",
			provider, paymentdomain.PaymentRiskEventDispute, start, end)
	if err := disputeEventQuery.Count(&disputeEventCount).Error; err != nil {
		return PaymentRiskMetricCounts{}, err
	}
	if disputeEventCount > counts.DisputeCount {
		counts.DisputeCount = disputeEventCount
	}

	refundQuery := r.db.Model(&paymentdomain.Refund{}).
		Joins("JOIN transactions ON transactions.id = refunds.transaction_id").
		Where("LOWER(transactions.payment_method) = ? AND refunds.status = ?", paymentMethod, "completed").
		Where("COALESCE(refunds.completed_at, refunds.created_at) >= ? AND COALESCE(refunds.completed_at, refunds.created_at) < ?", start, end)
	if err := refundQuery.Count(&counts.RefundCount).Error; err != nil {
		return PaymentRiskMetricCounts{}, err
	}
	if err := refundQuery.Select("COALESCE(SUM(refunds.amount), 0)").Scan(&counts.RefundAmount).Error; err != nil {
		return PaymentRiskMetricCounts{}, err
	}

	checkoutDecisionBase := func() *gorm.DB {
		return r.db.Model(&paymentdomain.PaymentRiskCheckoutDecision{}).
			Where("provider = ? AND occurred_at >= ? AND occurred_at < ?", provider, start, end)
	}
	if err := checkoutDecisionBase().Count(&counts.CheckoutAttemptCount).Error; err != nil {
		return PaymentRiskMetricCounts{}, err
	}
	if err := checkoutDecisionBase().
		Where("mode IN ?", []string{"any", "challenge"}).
		Count(&counts.ThreeDSUpgradeCount).Error; err != nil {
		return PaymentRiskMetricCounts{}, err
	}
	if err := checkoutDecisionBase().
		Where("mode = ?", "challenge").
		Count(&counts.ThreeDSChallengeCount).Error; err != nil {
		return PaymentRiskMetricCounts{}, err
	}
	if err := checkoutDecisionBase().
		Where("exemption_candidate = ?", true).
		Count(&counts.ThreeDSExemptionCount).Error; err != nil {
		return PaymentRiskMetricCounts{}, err
	}
	return counts, nil
}

func (r *PaymentRiskRepository) FindTransactionByProviderPaymentID(providerPaymentID string) (*paymentdomain.Transaction, error) {
	var transaction paymentdomain.Transaction
	err := r.db.Where("transaction_id = ?", strings.TrimSpace(providerPaymentID)).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func lockPaymentRiskAlertState(tx *gorm.DB, provider string) (*paymentdomain.PaymentRiskAlertState, error) {
	initial := paymentdomain.PaymentRiskAlertState{
		Provider:     provider,
		CurrentLevel: paymentdomain.PaymentRiskLevelNormal,
	}
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "provider"}},
		DoNothing: true,
	}).Create(&initial).Error; err != nil {
		return nil, err
	}

	query := tx.Where("provider = ?", provider)
	switch tx.Dialector.Name() {
	case "postgres", "mysql", "sqlserver":
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var state paymentdomain.PaymentRiskAlertState
	if err := query.First(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

func newPaymentRiskLevelChangedOutboxEvent(
	previousLevel paymentdomain.PaymentRiskLevel,
	snapshot *paymentdomain.PaymentRiskSnapshot,
) (outbox.Event, error) {
	reasons := make([]string, 0)
	if strings.TrimSpace(snapshot.ReasonsJSON) != "" {
		if err := json.Unmarshal([]byte(snapshot.ReasonsJSON), &reasons); err != nil {
			return outbox.Event{}, fmt.Errorf("decode payment risk snapshot reasons: %w", err)
		}
	}
	payload, err := json.Marshal(PaymentRiskLevelChangedPayload{
		Provider:               snapshot.Provider,
		PreviousLevel:          previousLevel,
		CurrentLevel:           snapshot.Level,
		WindowDays:             snapshot.WindowDays,
		SuccessfulPaymentCount: snapshot.SuccessfulPaymentCount,
		DisputeCount:           snapshot.DisputeCount,
		EarlyFraudWarningCount: snapshot.EarlyFraudWarningCount,
		RefundCount:            snapshot.RefundCount,
		CheckoutAttemptCount:   snapshot.CheckoutAttemptCount,
		ThreeDSUpgradeCount:    snapshot.ThreeDSUpgradeCount,
		ThreeDSChallengeCount:  snapshot.ThreeDSChallengeCount,
		ThreeDSExemptionCount:  snapshot.ThreeDSExemptionCount,
		DisputeActivityRate:    snapshot.DisputeActivityRate,
		EarlyFraudWarningRate:  snapshot.EarlyFraudWarningRate,
		RefundRate:             snapshot.RefundRate,
		ThreeDSUpgradeRate:     snapshot.ThreeDSUpgradeRate,
		Reasons:                reasons,
		RecommendedAction:      snapshot.RecommendedAction,
		ComputedAt:             snapshot.ComputedAt.UTC(),
	})
	if err != nil {
		return outbox.Event{}, fmt.Errorf("encode payment risk alert payload: %w", err)
	}
	return outbox.Event{
		EventKey:      fmt.Sprintf("%s:%s:%d", outbox.EventTypePaymentRiskLevelChanged, snapshot.Provider, snapshot.ID),
		EventType:     outbox.EventTypePaymentRiskLevelChanged,
		AggregateType: outbox.AggregateTypePaymentRiskProvider,
		AggregateID:   snapshot.Provider,
		Payload:       datatypes.JSON(payload),
		AvailableAt:   snapshot.ComputedAt.UTC(),
	}, nil
}
