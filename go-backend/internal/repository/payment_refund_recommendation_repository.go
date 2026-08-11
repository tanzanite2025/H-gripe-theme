package repository

import (
	"strings"
	"time"

	paymentdomain "commerce-platform/internal/domain/payment"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentRefundRecommendationRepository struct {
	db *gorm.DB
}

func NewPaymentRefundRecommendationRepository(db *gorm.DB) *PaymentRefundRecommendationRepository {
	return &PaymentRefundRecommendationRepository{db: db}
}

func (r *PaymentRefundRecommendationRepository) WithTx(tx *gorm.DB) *PaymentRefundRecommendationRepository {
	return &PaymentRefundRecommendationRepository{db: tx}
}

func (r *PaymentRefundRecommendationRepository) lockForUpdate(query *gorm.DB) *gorm.DB {
	switch r.db.Dialector.Name() {
	case "postgres", "mysql", "sqlserver":
		return query.Clauses(clause.Locking{Strength: "UPDATE"})
	default:
		return query
	}
}

func (r *PaymentRefundRecommendationRepository) UpsertRecommendation(
	recommendation *paymentdomain.PaymentRefundRecommendation,
) (*paymentdomain.PaymentRefundRecommendation, bool, error) {
	var existing paymentdomain.PaymentRefundRecommendation
	err := r.db.Where(
		"provider = ? AND source_kind = ? AND external_reference = ?",
		recommendation.Provider,
		recommendation.SourceKind,
		recommendation.ExternalReference,
	).First(&existing).Error
	if err == nil {
		updates := map[string]interface{}{
			"webhook_event_id":     recommendation.WebhookEventID,
			"risk_event_id":        recommendation.RiskEventID,
			"order_id":             recommendation.OrderID,
			"transaction_id":       recommendation.TransactionID,
			"provider_payment_id":  recommendation.ProviderPaymentID,
			"payment_intent_id":    recommendation.PaymentIntentID,
			"charge_id":            recommendation.ChargeID,
			"recommended_action":   recommendation.RecommendedAction,
			"recommended_amount":   recommendation.RecommendedAmount,
			"currency":             recommendation.Currency,
			"priority":             recommendation.Priority,
			"reason":               recommendation.Reason,
			"provider_reason":      recommendation.ProviderReason,
			"review_by":            recommendation.ReviewBy,
			"source_metadata_json": recommendation.SourceMetadataJSON,
			"updated_at":           time.Now().UTC(),
		}
		if err := r.db.Model(&paymentdomain.PaymentRefundRecommendation{}).
			Where("id = ?", existing.ID).
			Updates(updates).Error; err != nil {
			return nil, false, err
		}
		record, err := r.FindRecommendationByID(existing.ID)
		return record, false, err
	}
	if !IsRecordNotFound(err) {
		return nil, false, err
	}
	if err := r.db.Create(recommendation).Error; err != nil {
		return nil, false, err
	}
	return recommendation, true, nil
}

func (r *PaymentRefundRecommendationRepository) FindRecommendationByID(id uint) (*paymentdomain.PaymentRefundRecommendation, error) {
	var recommendation paymentdomain.PaymentRefundRecommendation
	err := r.db.First(&recommendation, id).Error
	if err != nil {
		return nil, err
	}
	return &recommendation, nil
}

func (r *PaymentRefundRecommendationRepository) FindRecommendationByIDForUpdate(id uint) (*paymentdomain.PaymentRefundRecommendation, error) {
	var recommendation paymentdomain.PaymentRefundRecommendation
	err := r.lockForUpdate(r.db).First(&recommendation, id).Error
	if err != nil {
		return nil, err
	}
	return &recommendation, nil
}

func (r *PaymentRefundRecommendationRepository) FindRecommendationBySource(
	provider string,
	sourceKind paymentdomain.PaymentRiskEventKind,
	externalReference string,
) (*paymentdomain.PaymentRefundRecommendation, error) {
	var recommendation paymentdomain.PaymentRefundRecommendation
	err := r.db.Where(
		"provider = ? AND source_kind = ? AND external_reference = ?",
		strings.ToLower(strings.TrimSpace(provider)),
		sourceKind,
		strings.TrimSpace(externalReference),
	).First(&recommendation).Error
	if err != nil {
		return nil, err
	}
	return &recommendation, nil
}

func (r *PaymentRefundRecommendationRepository) FindRiskEventByReference(
	provider string,
	kind paymentdomain.PaymentRiskEventKind,
	externalReference string,
) (*paymentdomain.PaymentRiskEvent, error) {
	var event paymentdomain.PaymentRiskEvent
	err := r.db.Where(
		"provider = ? AND kind = ? AND external_reference = ?",
		strings.ToLower(strings.TrimSpace(provider)),
		kind,
		strings.TrimSpace(externalReference),
	).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *PaymentRefundRecommendationRepository) FindTransactionByProviderPaymentID(providerPaymentID string) (*paymentdomain.Transaction, error) {
	var transaction paymentdomain.Transaction
	err := r.db.Where("transaction_id = ?", strings.TrimSpace(providerPaymentID)).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *PaymentRefundRecommendationRepository) ListRecommendations(
	status string,
	provider string,
	page int,
	pageSize int,
) ([]paymentdomain.PaymentRefundRecommendation, int64, error) {
	var recommendations []paymentdomain.PaymentRefundRecommendation
	var total int64
	query := r.db.Model(&paymentdomain.PaymentRefundRecommendation{})
	if status = strings.TrimSpace(status); status != "" {
		query = query.Where("status = ?", status)
	}
	if provider = strings.ToLower(strings.TrimSpace(provider)); provider != "" {
		query = query.Where("provider = ?", provider)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	err := query.
		Order("CASE priority WHEN 'high' THEN 0 ELSE 1 END").
		Order("review_by IS NULL").
		Order("review_by ASC").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&recommendations).Error
	return recommendations, total, err
}

func (r *PaymentRefundRecommendationRepository) UpdateRecommendation(
	recommendation *paymentdomain.PaymentRefundRecommendation,
) error {
	return r.db.Save(recommendation).Error
}
