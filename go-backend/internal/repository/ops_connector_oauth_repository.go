package repository

import (
	"time"

	"commerce-platform/internal/domain/ops"

	"gorm.io/gorm"
)

type OpsConnectorOAuthRepository struct {
	db *gorm.DB
}

func NewOpsConnectorOAuthRepository(db *gorm.DB) *OpsConnectorOAuthRepository {
	return &OpsConnectorOAuthRepository{db: db}
}

func (r *OpsConnectorOAuthRepository) Create(session *ops.ConnectorOAuthSession) error {
	return r.db.Create(session).Error
}

func (r *OpsConnectorOAuthRepository) Consume(stateHash string, now time.Time) (*ops.ConnectorOAuthSession, error) {
	var session ops.ConnectorOAuthSession
	if err := r.db.
		Where("state_hash = ? AND consumed_at IS NULL AND expires_at > ?", stateHash, now).
		First(&session).Error; err != nil {
		return nil, err
	}

	consumedAt := now
	result := r.db.Model(&ops.ConnectorOAuthSession{}).
		Where("id = ? AND consumed_at IS NULL", session.ID).
		Updates(map[string]interface{}{
			"consumed_at": consumedAt,
			"status":      ops.ConnectorOAuthStatusConsumed,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected != 1 {
		return nil, gorm.ErrRecordNotFound
	}
	session.ConsumedAt = &consumedAt
	session.Status = ops.ConnectorOAuthStatusConsumed
	return &session, nil
}

func (r *OpsConnectorOAuthRepository) MarkError(id uint, message string) error {
	return r.db.Model(&ops.ConnectorOAuthSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     ops.ConnectorOAuthStatusError,
			"last_error": message,
		}).Error
}

func (r *OpsConnectorOAuthRepository) DeleteExpired(now time.Time) error {
	return r.db.Where("expires_at <= ? OR consumed_at IS NOT NULL", now.Add(-24*time.Hour)).
		Delete(&ops.ConnectorOAuthSession{}).Error
}
