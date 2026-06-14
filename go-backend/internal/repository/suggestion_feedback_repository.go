package repository

import (
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/domain/suggestionfeedback"

	"gorm.io/gorm"
)

type SuggestionFeedbackRepository struct {
	db *gorm.DB
}

func NewSuggestionFeedbackRepository(db *gorm.DB) *SuggestionFeedbackRepository {
	return &SuggestionFeedbackRepository{db: db}
}

func (r *SuggestionFeedbackRepository) Create(item *suggestionfeedback.SuggestionFeedback) error {
	return r.db.Create(item).Error
}

func (r *SuggestionFeedbackRepository) GetUserMemberLevelName(userID uint) (string, error) {
	var userLoyalty loyalty.UserLoyalty
	if err := r.db.Where("user_id = ?", userID).First(&userLoyalty).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	if userLoyalty.MemberLevelID == 0 {
		return "", nil
	}

	var level loyalty.MemberLevel
	if err := r.db.First(&level, userLoyalty.MemberLevelID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return level.Name, nil
}
