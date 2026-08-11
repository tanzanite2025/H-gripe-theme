package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/currency"
	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/repository"
)

var ErrPaymentRefundRecommendationNotFound = errors.New("payment refund recommendation not found")

type PaymentRefundRecommendationService struct {
	repo      *repository.PaymentRefundRecommendationRepository
	txManager *repository.TxManager
}

func NewPaymentRefundRecommendationService(repo *repository.PaymentRefundRecommendationRepository, txManager ...*repository.TxManager) *PaymentRefundRecommendationService {
	service := &PaymentRefundRecommendationService{repo: repo}
	if len(txManager) > 0 {
		service.txManager = txManager[0]
	}
	return service
}

func (s *PaymentRefundRecommendationService) Enabled() bool {
	return s != nil && s.repo != nil
}

func (s *PaymentRefundRecommendationService) EnqueueFromRiskEvent(
	input PaymentRiskEventInput,
) (*paymentdomain.PaymentRefundRecommendation, error) {
	if !s.Enabled() {
		return nil, nil
	}

	normalized, err := s.normalizeRiskEventInput(input)
	if err != nil {
		return nil, err
	}
	if !refundRecommendationActionable(normalized) {
		return s.cancelPendingRecommendation(normalized)
	}

	plan := refundRecommendationPlan(normalized)
	metadataJSON, err := json.Marshal(normalized.Metadata)
	if err != nil {
		return nil, fmt.Errorf("encode refund recommendation metadata: %w", err)
	}
	riskEventID := s.riskEventID(normalized)
	recommendation := &paymentdomain.PaymentRefundRecommendation{
		Provider:           normalized.Provider,
		SourceKind:         normalized.Kind,
		ExternalReference:  normalized.ExternalReference,
		WebhookEventID:     normalized.WebhookEventID,
		RiskEventID:        riskEventID,
		OrderID:            normalized.OrderID,
		TransactionID:      normalized.TransactionID,
		ProviderPaymentID:  normalized.ProviderPaymentID,
		PaymentIntentID:    normalized.PaymentIntentID,
		ChargeID:           normalized.ChargeID,
		RecommendedAction:  plan.Action,
		RecommendedAmount:  normalized.Amount,
		Currency:           normalized.Currency,
		Priority:           plan.Priority,
		Status:             paymentdomain.PaymentRefundRecommendationStatusPending,
		Reason:             plan.Reason,
		ProviderReason:     plan.ProviderReason,
		ReviewBy:           plan.ReviewBy,
		SourceMetadataJSON: string(metadataJSON),
	}
	record, _, err := s.repo.UpsertRecommendation(recommendation)
	return record, err
}

func (s *PaymentRefundRecommendationService) ListRecommendations(
	status string,
	provider string,
	page int,
	pageSize int,
) ([]paymentdomain.PaymentRefundRecommendation, int64, error) {
	if !s.Enabled() {
		return nil, 0, nil
	}
	if status = strings.TrimSpace(status); status != "" && !validPaymentRefundRecommendationStatus(status) {
		return nil, 0, errors.New("invalid refund recommendation status")
	}
	return s.repo.ListRecommendations(status, provider, page, pageSize)
}

func (s *PaymentRefundRecommendationService) GetRecommendation(id uint) (*paymentdomain.PaymentRefundRecommendation, error) {
	if !s.Enabled() {
		return nil, ErrPaymentRefundRecommendationNotFound
	}
	record, err := s.repo.FindRecommendationByID(id)
	if repository.IsRecordNotFound(err) {
		return nil, ErrPaymentRefundRecommendationNotFound
	}
	return record, err
}

func (s *PaymentRefundRecommendationService) UpdateRecommendationDecision(
	id uint,
	status string,
	decisionNotes string,
	adminID uint,
) (*paymentdomain.PaymentRefundRecommendation, error) {
	status = strings.TrimSpace(status)
	if !validPaymentRefundRecommendationStatus(status) {
		return nil, errors.New("invalid refund recommendation status")
	}
	record, err := s.GetRecommendation(id)
	if err != nil {
		return nil, err
	}
	if record.Status != paymentdomain.PaymentRefundRecommendationStatusPending && record.Status != status {
		return nil, errors.New("refund recommendation is already finalized")
	}

	record.Status = status
	record.DecisionNotes = strings.TrimSpace(decisionNotes)
	if status != paymentdomain.PaymentRefundRecommendationStatusPending {
		now := time.Now().UTC()
		record.ReviewedAt = &now
		record.ReviewedByID = &adminID
	}
	if err := s.repo.UpdateRecommendation(record); err != nil {
		return nil, err
	}
	return record, nil
}

type CreatePendingRefundFromRecommendationInput struct {
	RecommendationID uint
	Amount           float64
	Reason           string
	DecisionNotes    string
	AdminID          uint
}

func (s *PaymentRefundRecommendationService) CreatePendingRefundFromRecommendation(
	input CreatePendingRefundFromRecommendationInput,
) (*paymentdomain.PaymentRefundRecommendation, *paymentdomain.Refund, error) {
	if !s.Enabled() || s.txManager == nil {
		return nil, nil, errors.New("payment refund recommendation workflow is not configured")
	}
	if input.RecommendationID == 0 {
		return nil, nil, errors.New("refund recommendation id is required")
	}
	if input.AdminID == 0 {
		return nil, nil, errors.New("admin user id is required")
	}

	var updatedRecommendation *paymentdomain.PaymentRefundRecommendation
	var createdRefund *paymentdomain.Refund
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.RefundReview == nil {
			return errors.New("payment refund recommendation repository is not configured for transactions")
		}
		recommendation, err := repos.RefundReview.FindRecommendationByIDForUpdate(input.RecommendationID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return ErrPaymentRefundRecommendationNotFound
			}
			return err
		}

		if recommendation.LinkedRefundID != nil {
			refund, err := repos.Payment.FindRefundByID(*recommendation.LinkedRefundID)
			if err != nil {
				return err
			}
			updatedRecommendation = recommendation
			createdRefund = refund
			return nil
		}
		if recommendation.Status != paymentdomain.PaymentRefundRecommendationStatusPending &&
			recommendation.Status != paymentdomain.PaymentRefundRecommendationStatusAccepted {
			return errors.New("refund recommendation is not available for refund draft creation")
		}
		if recommendation.OrderID == nil || *recommendation.OrderID == 0 {
			return errors.New("refund recommendation is missing order linkage")
		}
		if recommendation.TransactionID == nil || *recommendation.TransactionID == 0 {
			return errors.New("refund recommendation is missing transaction linkage")
		}

		amount := roundRefundMoney(input.Amount)
		if amount <= 0 {
			amount = roundRefundMoney(recommendation.RecommendedAmount)
		}
		if amount <= 0 {
			return errors.New("refund amount is required")
		}
		if recommendation.RecommendedAmount > 0 && amount-recommendation.RecommendedAmount > 0.01 {
			return fmt.Errorf("refund amount %.2f exceeds recommended amount %.2f", amount, recommendation.RecommendedAmount)
		}

		refund := &paymentdomain.Refund{
			OrderID:       *recommendation.OrderID,
			TransactionID: *recommendation.TransactionID,
			Amount:        amount,
			Reason:        refundRecommendationDraftReason(recommendation, input.Reason),
		}
		if err := createAdminRefundInTx(repos, refund, input.AdminID); err != nil {
			return err
		}

		now := time.Now().UTC()
		recommendation.Status = paymentdomain.PaymentRefundRecommendationStatusAccepted
		recommendation.LinkedRefundID = &refund.ID
		recommendation.RefundCreatedByID = &input.AdminID
		recommendation.RefundCreatedAt = &now
		recommendation.ReviewedByID = &input.AdminID
		recommendation.ReviewedAt = &now
		recommendation.DecisionNotes = refundRecommendationDraftDecisionNotes(
			recommendation.DecisionNotes,
			input.DecisionNotes,
			refund.ID,
		)
		if err := repos.RefundReview.UpdateRecommendation(recommendation); err != nil {
			return err
		}

		updatedRecommendation = recommendation
		createdRefund = refund
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return updatedRecommendation, createdRefund, nil
}

func refundRecommendationDraftReason(recommendation *paymentdomain.PaymentRefundRecommendation, override string) string {
	if reason := strings.TrimSpace(override); reason != "" {
		return reason
	}
	source := strings.TrimSpace(recommendation.Reason)
	if source == "" {
		source = "payment risk recommendation"
	}
	return fmt.Sprintf("Manual refund draft from risk recommendation #%d: %s", recommendation.ID, source)
}

func refundRecommendationDraftDecisionNotes(existing string, input string, refundID uint) string {
	base := strings.TrimSpace(input)
	if base == "" {
		base = strings.TrimSpace(existing)
	}
	line := fmt.Sprintf("Created local pending refund #%d from this recommendation. No gateway refund was executed.", refundID)
	if base == "" {
		return line
	}
	if strings.Contains(base, line) {
		return base
	}
	return base + "\n" + line
}

func (s *PaymentRefundRecommendationService) normalizeRiskEventInput(input PaymentRiskEventInput) (PaymentRiskEventInput, error) {
	input.Provider = strings.ToLower(strings.TrimSpace(input.Provider))
	if input.Provider == "" {
		return PaymentRiskEventInput{}, errors.New("payment risk provider is required")
	}
	if input.Kind != paymentdomain.PaymentRiskEventEarlyFraudWarning &&
		input.Kind != paymentdomain.PaymentRiskEventDispute {
		return PaymentRiskEventInput{}, errors.New("unsupported payment risk event kind")
	}
	input.ExternalReference = strings.TrimSpace(input.ExternalReference)
	if input.ExternalReference == "" {
		input.ExternalReference = strings.TrimSpace(input.WebhookEventID)
	}
	if input.ExternalReference == "" {
		return PaymentRiskEventInput{}, errors.New("payment risk external reference is required")
	}
	input.WebhookEventID = strings.TrimSpace(input.WebhookEventID)
	input.ProviderPaymentID = strings.TrimSpace(input.ProviderPaymentID)
	input.PaymentIntentID = strings.TrimSpace(input.PaymentIntentID)
	input.ChargeID = strings.TrimSpace(input.ChargeID)
	input.Currency = normalizeRecommendationCurrency(input.Currency)
	if input.Metadata == nil {
		input.Metadata = map[string]string{}
	}
	if input.OccurredAt.IsZero() {
		input.OccurredAt = time.Now().UTC()
	} else {
		input.OccurredAt = input.OccurredAt.UTC()
	}
	if input.TransactionID == nil {
		input = s.resolveRecommendationTransaction(input)
	}
	return input, nil
}

func (s *PaymentRefundRecommendationService) resolveRecommendationTransaction(input PaymentRiskEventInput) PaymentRiskEventInput {
	for _, candidate := range []string{input.ProviderPaymentID, input.PaymentIntentID} {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		transaction, err := s.repo.FindTransactionByProviderPaymentID(candidate)
		if err != nil {
			continue
		}
		input.TransactionID = &transaction.ID
		if input.OrderID == nil {
			input.OrderID = &transaction.OrderID
		}
		return input
	}
	return input
}

func (s *PaymentRefundRecommendationService) riskEventID(input PaymentRiskEventInput) *uint {
	event, err := s.repo.FindRiskEventByReference(input.Provider, input.Kind, input.ExternalReference)
	if err != nil {
		return nil
	}
	return &event.ID
}

func (s *PaymentRefundRecommendationService) cancelPendingRecommendation(
	input PaymentRiskEventInput,
) (*paymentdomain.PaymentRefundRecommendation, error) {
	record, err := s.repo.FindRecommendationBySource(input.Provider, input.Kind, input.ExternalReference)
	if repository.IsRecordNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if record.Status != paymentdomain.PaymentRefundRecommendationStatusPending {
		return record, nil
	}
	record.Status = paymentdomain.PaymentRefundRecommendationStatusCancelled
	record.DecisionNotes = "Provider risk event is no longer actionable."
	now := time.Now().UTC()
	record.ReviewedAt = &now
	if err := s.repo.UpdateRecommendation(record); err != nil {
		return nil, err
	}
	return record, nil
}

type refundRecommendationPlanResult struct {
	Action         string
	Reason         string
	ProviderReason string
	Priority       string
	ReviewBy       *time.Time
}

func refundRecommendationPlan(input PaymentRiskEventInput) refundRecommendationPlanResult {
	providerReason := refundRecommendationProviderReason(input.Metadata)
	switch input.Kind {
	case paymentdomain.PaymentRiskEventEarlyFraudWarning:
		reviewBy := input.OccurredAt.Add(24 * time.Hour)
		return refundRecommendationPlanResult{
			Action:         paymentdomain.PaymentRefundRecommendationActionReviewRefundBeforeDispute,
			Reason:         "Early fraud warning received; review whether a manual refund should be created before dispute escalation.",
			ProviderReason: providerReason,
			Priority:       paymentdomain.PaymentRefundRecommendationPriorityHigh,
			ReviewBy:       &reviewBy,
		}
	case paymentdomain.PaymentRiskEventDispute:
		return refundRecommendationPlanResult{
			Action:         paymentdomain.PaymentRefundRecommendationActionReviewRefundOrEvidence,
			Reason:         "Payment dispute received; review evidence response, customer contact, or manual refund path.",
			ProviderReason: providerReason,
			Priority:       paymentdomain.PaymentRefundRecommendationPriorityHigh,
		}
	default:
		return refundRecommendationPlanResult{
			Action:         paymentdomain.PaymentRefundRecommendationActionReviewRefundOrEvidence,
			Reason:         "Payment risk event received for manual review.",
			ProviderReason: providerReason,
			Priority:       paymentdomain.PaymentRefundRecommendationPriorityNormal,
		}
	}
}

func refundRecommendationActionable(input PaymentRiskEventInput) bool {
	if input.Kind != paymentdomain.PaymentRiskEventDispute {
		return true
	}
	status := strings.ToLower(strings.TrimSpace(input.Metadata["status"]))
	if status == "" {
		return true
	}
	switch status {
	case "closed", "won", "lost", "resolved", "denied", "canceled", "cancelled":
		return false
	default:
		return true
	}
}

func refundRecommendationProviderReason(metadata map[string]string) string {
	for _, key := range []string{"fraud_type", "reason", "status", "event_type"} {
		if value := strings.TrimSpace(metadata[key]); value != "" {
			return value
		}
	}
	return ""
}

func normalizeRecommendationCurrency(value string) string {
	normalized := currency.NormalizeCode(value)
	if normalized == "" {
		return ""
	}
	if !currency.IsValidCode(normalized) || !currency.IsCatalogCode(normalized) {
		return ""
	}
	return normalized
}

func validPaymentRefundRecommendationStatus(value string) bool {
	switch value {
	case paymentdomain.PaymentRefundRecommendationStatusPending,
		paymentdomain.PaymentRefundRecommendationStatusAccepted,
		paymentdomain.PaymentRefundRecommendationStatusDismissed,
		paymentdomain.PaymentRefundRecommendationStatusCancelled:
		return true
	default:
		return false
	}
}
