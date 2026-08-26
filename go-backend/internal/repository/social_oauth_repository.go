package repository

import (
	"time"

	"commerce-platform/internal/domain/social"

	"gorm.io/gorm"
)

type SocialOAuthRepository struct {
	db *gorm.DB
}

func NewSocialOAuthRepository(db *gorm.DB) *SocialOAuthRepository {
	return &SocialOAuthRepository{db: db}
}

func (r *SocialOAuthRepository) ListConnections() ([]social.OAuthConnection, error) {
	var connections []social.OAuthConnection
	if err := r.db.Order("provider ASC").Find(&connections).Error; err != nil {
		return nil, err
	}
	return connections, nil
}

func (r *SocialOAuthRepository) FindConnection(provider string) (*social.OAuthConnection, error) {
	var connection social.OAuthConnection
	if err := r.db.Where("provider = ?", provider).First(&connection).Error; err != nil {
		return nil, err
	}
	return &connection, nil
}

func (r *SocialOAuthRepository) SaveConnection(connection *social.OAuthConnection) error {
	return r.db.Save(connection).Error
}

func (r *SocialOAuthRepository) CreateSession(session *social.OAuthSession) error {
	return r.db.Create(session).Error
}

func (r *SocialOAuthRepository) ConsumeSession(stateHash string, now time.Time) (*social.OAuthSession, error) {
	var session social.OAuthSession
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("state_hash = ? AND consumed_at IS NULL AND expires_at > ?", stateHash, now).
			First(&session).Error; err != nil {
			return err
		}

		consumedAt := now
		result := tx.Model(&social.OAuthSession{}).
			Where("id = ? AND consumed_at IS NULL", session.ID).
			Updates(map[string]interface{}{
				"consumed_at": consumedAt,
				"status":      social.OAuthSessionStatusConsumed,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		session.ConsumedAt = &consumedAt
		session.Status = social.OAuthSessionStatusConsumed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SocialOAuthRepository) MarkSessionError(id uint, message string) error {
	return r.db.Model(&social.OAuthSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     social.OAuthSessionStatusError,
			"last_error": message,
		}).Error
}

func (r *SocialOAuthRepository) DeleteExpiredSessions(now time.Time) error {
	return r.db.
		Where("expires_at <= ? OR consumed_at IS NOT NULL", now.Add(-24*time.Hour)).
		Delete(&social.OAuthSession{}).Error
}
