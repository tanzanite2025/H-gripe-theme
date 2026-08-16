package repository

import (
	"commerce-platform/internal/domain/user"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReplaceCustomerServiceAgentAvatar changes the one avatar owned by an
// existing customer-service profile. The callback runs in the same
// transaction after the profile mutation, so callers can persist a durable
// cleanup event together with the new reference.
func (r *UserRepository) ReplaceCustomerServiceAgentAvatar(
	userID uint,
	avatarURL string,
	afterReplace func(tx *gorm.DB, profileID uint, previousAvatarURL string) error,
) (string, error) {
	if r == nil || r.db == nil {
		return "", fmt.Errorf("customer-service avatar repository is not configured")
	}
	if userID == 0 {
		return "", gorm.ErrRecordNotFound
	}

	avatarURL = strings.TrimSpace(avatarURL)
	previousAvatarURL := ""
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var profile user.AgentProfile
		query := tx.Where("user_id = ?", userID)
		switch tx.Dialector.Name() {
		case "postgres", "mysql", "sqlserver":
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := query.First(&profile).Error; err != nil {
			return err
		}

		previousAvatarURL = strings.TrimSpace(profile.Avatar)
		if previousAvatarURL == avatarURL {
			return nil
		}

		result := tx.Model(&user.AgentProfile{}).
			Where("id = ?", profile.ID).
			Update("avatar", avatarURL)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}

		if afterReplace != nil {
			return afterReplace(tx, profile.ID, previousAvatarURL)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return previousAvatarURL, nil
}

// IsCurrentCustomerServiceAgentAvatarURL reports whether the URL is still
// referenced by a live customer-service profile.
func (r *UserRepository) IsCurrentCustomerServiceAgentAvatarURL(avatarURL string) (bool, error) {
	if r == nil || r.db == nil {
		return false, fmt.Errorf("customer-service avatar repository is not configured")
	}
	avatarURL = strings.TrimSpace(avatarURL)
	if avatarURL == "" {
		return false, nil
	}

	var count int64
	if err := r.db.Model(&user.AgentProfile{}).Where("avatar = ?", avatarURL).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
