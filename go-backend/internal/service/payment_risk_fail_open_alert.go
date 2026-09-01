package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

const (
	paymentRiskFailOpenAlertBucket       = 5 * time.Minute
	maxPaymentRiskFailOpenAlertErrorText = 500
)

type PaymentRiskFailOpenAlertInput struct {
	Provider       string
	Component      string
	Operation      string
	Reason         string
	FallbackAction string
	OccurredAt     time.Time
	Err            error
}

type PaymentRiskFailOpenAlertPublisher interface {
	PublishPaymentRiskFailOpenAlert(ctx context.Context, input PaymentRiskFailOpenAlertInput) error
}

type PaymentRiskFailOpenOutboxPublisher struct {
	repo *repository.OutboxRepository
	now  func() time.Time
}

type paymentRiskFailOpenAlertPayload struct {
	Provider            string    `json:"provider"`
	Component           string    `json:"component"`
	Operation           string    `json:"operation"`
	Severity            string    `json:"severity"`
	Reason              string    `json:"reason"`
	FailOpen            bool      `json:"fail_open"`
	RedisUnavailable    bool      `json:"redis_unavailable"`
	FallbackAction      string    `json:"fallback_action"`
	Error               string    `json:"error,omitempty"`
	OccurredAt          time.Time `json:"occurred_at"`
	DeduplicationBucket time.Time `json:"deduplication_bucket"`
}

func NewPaymentRiskFailOpenOutboxPublisher(repo *repository.OutboxRepository) *PaymentRiskFailOpenOutboxPublisher {
	return &PaymentRiskFailOpenOutboxPublisher{
		repo: repo,
		now:  time.Now,
	}
}

func (p *PaymentRiskFailOpenOutboxPublisher) PublishPaymentRiskFailOpenAlert(ctx context.Context, input PaymentRiskFailOpenAlertInput) error {
	if p == nil || p.repo == nil {
		return nil
	}
	_ = ctx

	provider := strings.ToLower(strings.TrimSpace(input.Provider))
	if provider == "" {
		provider = "unknown"
	}
	component := strings.ToLower(strings.TrimSpace(input.Component))
	if component == "" {
		component = "payment_risk"
	}
	operation := strings.ToLower(strings.TrimSpace(input.Operation))
	if operation == "" {
		operation = "checkout"
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		reason = "redis_unavailable"
	}
	fallbackAction := strings.TrimSpace(input.FallbackAction)
	if fallbackAction == "" {
		fallbackAction = "checkout_allowed_with_degraded_protection"
	}

	occurredAt := input.OccurredAt
	if occurredAt.IsZero() {
		now := p.now
		if now == nil {
			now = time.Now
		}
		occurredAt = now()
	}
	occurredAt = occurredAt.UTC()
	bucket := occurredAt.Truncate(paymentRiskFailOpenAlertBucket)

	payload, err := json.Marshal(paymentRiskFailOpenAlertPayload{
		Provider:            provider,
		Component:           component,
		Operation:           operation,
		Severity:            "critical",
		Reason:              reason,
		FailOpen:            true,
		RedisUnavailable:    true,
		FallbackAction:      fallbackAction,
		Error:               truncatePaymentRiskFailOpenAlertError(input.Err),
		OccurredAt:          occurredAt,
		DeduplicationBucket: bucket,
	})
	if err != nil {
		return fmt.Errorf("encode payment risk fail-open alert payload: %w", err)
	}

	return p.repo.CreateEvent(&outbox.Event{
		EventKey:      paymentRiskFailOpenAlertEventKey(provider, component, operation, bucket),
		EventType:     outbox.EventTypePaymentRiskFailOpen,
		AggregateType: outbox.AggregateTypePaymentRiskProvider,
		AggregateID:   provider,
		Payload:       datatypes.JSON(payload),
		AvailableAt:   occurredAt,
	})
}

func paymentRiskFailOpenAlertEventKey(provider, component, operation string, bucket time.Time) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s",
		outbox.EventTypePaymentRiskFailOpen,
		provider,
		component,
		operation,
		bucket.UTC().Format("200601021504"),
	)
}

func truncatePaymentRiskFailOpenAlertError(err error) string {
	if err == nil {
		return ""
	}
	value := strings.TrimSpace(err.Error())
	if len(value) <= maxPaymentRiskFailOpenAlertErrorText {
		return value
	}
	return value[:maxPaymentRiskFailOpenAlertErrorText]
}
