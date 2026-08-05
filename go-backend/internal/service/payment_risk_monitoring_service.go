package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	paymentdomain "tanzanite/internal/domain/payment"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/repository"

	"gorm.io/gorm"
)

type PaymentRiskEventInput struct {
	Provider          string
	Kind              paymentdomain.PaymentRiskEventKind
	ExternalReference string
	WebhookEventID    string
	ProviderPaymentID string
	PaymentIntentID   string
	ChargeID          string
	OrderID           *uint
	TransactionID     *uint
	Amount            float64
	Currency          string
	OccurredAt        time.Time
	Payload           string
	Metadata          map[string]string
}

type PaymentRiskReport struct {
	Snapshot *paymentdomain.PaymentRiskSnapshot `json:"snapshot"`
	Reasons  []string                           `json:"reasons"`
}

type PaymentRiskMonitoringService struct {
	repo            *repository.PaymentRiskRepository
	config          config.PaymentRiskMonitoringConfig
	policy          paymentRiskPolicy
	alertingEnabled bool
}

func NewPaymentRiskMonitoringService(
	repo *repository.PaymentRiskRepository,
	cfg config.PaymentRiskMonitoringConfig,
) *PaymentRiskMonitoringService {
	cfg = normalizePaymentRiskMonitoringConfig(cfg)
	return &PaymentRiskMonitoringService{
		repo:            repo,
		config:          cfg,
		policy:          paymentRiskPolicy{config: cfg},
		alertingEnabled: false,
	}
}

func (s *PaymentRiskMonitoringService) Enabled() bool {
	return s != nil && s.config.Enabled && s.repo != nil
}

// ConfigureAlerting is enabled only after the application has both opted in
// and registered a working Outbox delivery path.
func (s *PaymentRiskMonitoringService) ConfigureAlerting(enabled bool) {
	if s == nil {
		return
	}
	s.alertingEnabled = enabled
}

func (s *PaymentRiskMonitoringService) AlertingEnabled() bool {
	return s != nil && s.alertingEnabled
}

func (s *PaymentRiskMonitoringService) RecordEvent(input PaymentRiskEventInput) error {
	if !s.Enabled() {
		return nil
	}
	provider := strings.ToLower(strings.TrimSpace(input.Provider))
	if provider == "" {
		return errors.New("payment risk provider is required")
	}
	if input.Kind != paymentdomain.PaymentRiskEventEarlyFraudWarning &&
		input.Kind != paymentdomain.PaymentRiskEventDispute {
		return errors.New("unsupported payment risk event kind")
	}
	externalReference := strings.TrimSpace(input.ExternalReference)
	if externalReference == "" {
		externalReference = strings.TrimSpace(input.WebhookEventID)
	}
	if externalReference == "" {
		return errors.New("payment risk external reference is required")
	}
	occurredAt := input.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	metadataJSON, err := json.Marshal(input.Metadata)
	if err != nil {
		return fmt.Errorf("encode payment risk metadata: %w", err)
	}
	providerPaymentID := strings.TrimSpace(input.ProviderPaymentID)
	if input.TransactionID == nil && providerPaymentID != "" {
		transaction, transactionErr := s.repo.FindTransactionByProviderPaymentID(providerPaymentID)
		if transactionErr == nil {
			input.TransactionID = &transaction.ID
			if input.OrderID == nil {
				input.OrderID = &transaction.OrderID
			}
		} else if !repository.IsRecordNotFound(transactionErr) {
			return fmt.Errorf("resolve payment risk transaction: %w", transactionErr)
		}
	}

	return s.repo.UpsertPaymentRiskEvent(&paymentdomain.PaymentRiskEvent{
		Provider:          provider,
		Kind:              input.Kind,
		ExternalReference: externalReference,
		WebhookEventID:    strings.TrimSpace(input.WebhookEventID),
		ProviderPaymentID: providerPaymentID,
		PaymentIntentID:   strings.TrimSpace(input.PaymentIntentID),
		ChargeID:          strings.TrimSpace(input.ChargeID),
		OrderID:           input.OrderID,
		TransactionID:     input.TransactionID,
		Amount:            input.Amount,
		Currency:          strings.ToUpper(strings.TrimSpace(input.Currency)),
		OccurredAt:        occurredAt,
		Payload:           input.Payload,
		MetadataJSON:      string(metadataJSON),
	})
}

func (s *PaymentRiskMonitoringService) RecomputeProvider(ctx context.Context, provider string, now time.Time) (*PaymentRiskReport, error) {
	if !s.Enabled() {
		return nil, nil
	}
	_ = ctx
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return nil, errors.New("payment risk provider is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	windowStart := now.AddDate(0, 0, -s.config.WindowDays)
	counts, err := s.repo.CountPaymentRiskMetrics(provider, windowStart, now)
	if err != nil {
		return nil, fmt.Errorf("count payment risk metrics: %w", err)
	}
	snapshot := s.policy.Evaluate(PaymentRiskMetrics{
		Provider:                provider,
		WindowDays:              s.config.WindowDays,
		WindowStart:             windowStart,
		WindowEnd:               now,
		SuccessfulPaymentCount:  counts.SuccessfulPaymentCount,
		SuccessfulPaymentAmount: counts.SuccessfulPaymentAmount,
		DisputeCount:            counts.DisputeCount,
		DisputeAmount:           counts.DisputeAmount,
		EarlyFraudWarningCount:  counts.EarlyFraudWarningCount,
		RefundCount:             counts.RefundCount,
		RefundAmount:            counts.RefundAmount,
	}, now)
	if _, err := s.repo.CreatePaymentRiskSnapshotWithAlert(snapshot, s.alertingEnabled); err != nil {
		return nil, fmt.Errorf("store payment risk snapshot: %w", err)
	}
	return s.report(snapshot), nil
}

func (s *PaymentRiskMonitoringService) CurrentReport(provider string) (*PaymentRiskReport, error) {
	if !s.Enabled() {
		return nil, nil
	}
	snapshot, err := s.repo.FindLatestPaymentRiskSnapshot(provider)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s.report(snapshot), nil
}

func (s *PaymentRiskMonitoringService) CurrentCheckoutPolicy(provider string) (PaymentRiskCheckoutPolicy, error) {
	report, err := s.CurrentReport(provider)
	if err != nil {
		return PaymentRiskCheckoutPolicy{}, err
	}
	if report == nil || report.Snapshot == nil {
		return PaymentRiskCheckoutPolicy{
			Provider: provider,
			Level:    paymentdomain.PaymentRiskLevelNormal,
		}, nil
	}
	snapshot := report.Snapshot
	return PaymentRiskCheckoutPolicy{
		Provider:   snapshot.Provider,
		Level:      snapshot.Level,
		Force3DS:   s.config.AutoStepUpEnabled && snapshot.Level != paymentdomain.PaymentRiskLevelNormal,
		Reason:     snapshot.RecommendedAction,
		ComputedAt: snapshot.ComputedAt,
	}, nil
}

func (s *PaymentRiskMonitoringService) report(snapshot *paymentdomain.PaymentRiskSnapshot) *PaymentRiskReport {
	return &PaymentRiskReport{
		Snapshot: snapshot,
		Reasons:  decodePaymentRiskReasons(snapshot.ReasonsJSON),
	}
}

func normalizePaymentRiskMonitoringConfig(cfg config.PaymentRiskMonitoringConfig) config.PaymentRiskMonitoringConfig {
	if cfg.WindowDays <= 0 {
		cfg.WindowDays = 30
	}
	if cfg.MinimumSuccessfulPayments <= 0 {
		cfg.MinimumSuccessfulPayments = 20
	}
	if cfg.WarningDisputeActivityRate <= 0 {
		cfg.WarningDisputeActivityRate = 0.005
	}
	if cfg.CriticalDisputeActivityRate <= 0 {
		cfg.CriticalDisputeActivityRate = 0.008
	}
	if cfg.WarningEarlyFraudRate <= 0 {
		cfg.WarningEarlyFraudRate = 0.005
	}
	if cfg.CriticalEarlyFraudRate <= 0 {
		cfg.CriticalEarlyFraudRate = 0.009
	}
	if cfg.WarningRefundRate <= 0 {
		cfg.WarningRefundRate = 0.08
	}
	if cfg.CriticalRefundRate <= 0 {
		cfg.CriticalRefundRate = 0.15
	}
	return cfg
}

func encodePaymentRiskReasons(reasons []string) string {
	payload, err := json.Marshal(reasons)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func decodePaymentRiskReasons(value string) []string {
	var reasons []string
	if err := json.Unmarshal([]byte(value), &reasons); err != nil {
		return []string{}
	}
	return reasons
}
