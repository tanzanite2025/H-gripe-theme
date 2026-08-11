package repository

import (
	"errors"
	"time"

	"commerce-platform/internal/domain/verification"

	"gorm.io/gorm"
)

var ErrEmailChallengeInvalid = errors.New("email challenge is invalid")

type EmailChallengeRepository struct {
	db *gorm.DB
}

func NewEmailChallengeRepository(db *gorm.DB) *EmailChallengeRepository {
	return &EmailChallengeRepository{db: db}
}

func (r *EmailChallengeRepository) Create(challenge *verification.EmailChallenge) error {
	return r.db.Create(challenge).Error
}

func (r *EmailChallengeRepository) Find(tokenHash, purpose string) (*verification.EmailChallenge, error) {
	var challenge verification.EmailChallenge
	if err := r.db.Where("token_hash = ? AND purpose = ?", tokenHash, purpose).First(&challenge).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (r *EmailChallengeRepository) Consume(tokenHash, purpose string, now time.Time) (*verification.EmailChallenge, error) {
	var challenge verification.EmailChallenge
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("token_hash = ? AND purpose = ?", tokenHash, purpose).First(&challenge).Error; err != nil {
			return ErrEmailChallengeInvalid
		}
		if challenge.UsedAt != nil || !challenge.ExpiresAt.After(now) {
			return ErrEmailChallengeInvalid
		}

		usedAt := now
		result := tx.Model(&verification.EmailChallenge{}).
			Where("id = ? AND used_at IS NULL", challenge.ID).
			Update("used_at", usedAt)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrEmailChallengeInvalid
		}
		challenge.UsedAt = &usedAt
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &challenge, nil
}
